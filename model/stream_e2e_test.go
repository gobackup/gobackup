package model

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/internal/s3mock"
	"github.com/longbridgeapp/assert"
)

// fakeMysqldump puts a fake `mysqldump` first in PATH.
func fakeMysqldump(t *testing.T, script string) {
	dir := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "mysqldump"), []byte("#!/bin/sh\n"+script+"\n"), 0755))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func loadStreamModel(t *testing.T, name, endpoint, compress string) Model {
	return loadStreamModelDB(t, name, endpoint, compress, "mysql")
}

func loadStreamModelDB(t *testing.T, name, endpoint, compress, dbType string) Model {
	cfg := fmt.Sprintf(`models:
  %s:
    stream: true
    %s
    databases:
      app:
        type: %s
        database: app
        username: root
    storages:
      s3:
        type: s3
        bucket: bk
        endpoint: %s
        path: backups
        access_key_id: a
        secret_access_key: b
`, name, compress, dbType, endpoint)
	f := filepath.Join(t.TempDir(), "gobackup.yml")
	assert.NoError(t, os.WriteFile(f, []byte(cfg), 0644))
	assert.NoError(t, config.Init(f))
	t.Cleanup(func() { os.Remove(filepath.Join(config.GoBackupDir, "cycler", name+"_s3.json")) })
	m := GetModelByName(name)
	assert.NotNil(t, m)
	return *m
}

func Test_Perform_stream_raw(t *testing.T) {
	fakeMysqldump(t, `echo "-- dump of app"; echo "CREATE TABLE t(id int);"`)
	srv := s3mock.New()
	defer srv.Close()

	m := loadStreamModel(t, "stream_raw_test", srv.URL, "")
	assert.NoError(t, m.Perform())

	var found string
	for _, k := range srv.Keys() {
		if strings.HasPrefix(k, "bk/backups/") && strings.HasSuffix(k, ".sql") {
			found = k
		}
	}
	assert.NotEqual(t, "", found)
	got, _ := srv.Object(found)
	assert.Equal(t, "-- dump of app\nCREATE TABLE t(id int);\n", string(got))
}

func Test_Perform_stream_gzip(t *testing.T) {
	fakeMysqldump(t, `echo "CREATE TABLE t(id int);"`)
	srv := s3mock.New()
	defer srv.Close()

	m := loadStreamModel(t, "stream_gz_test", srv.URL, "compress_with:\n      type: gz")
	assert.NoError(t, m.Perform())

	for _, k := range srv.Keys() {
		if strings.HasSuffix(k, ".sql.gz") {
			b, _ := srv.Object(k)
			zr, err := gzip.NewReader(bytes.NewReader(b))
			assert.NoError(t, err)
			out, _ := io.ReadAll(zr)
			assert.Equal(t, "CREATE TABLE t(id int);\n", string(out))
			return
		}
	}
	t.Fatal("no .sql.gz object uploaded")
}

func Test_Perform_stream_dumpFailureLeavesNoObject(t *testing.T) {
	fakeMysqldump(t, `echo "partial"; echo "Access denied" >&2; exit 2`)
	srv := s3mock.New()
	defer srv.Close()

	m := loadStreamModel(t, "stream_fail_test", srv.URL, "")
	err := m.Perform()
	assert.NotNil(t, err)
	assert.Equal(t, true, strings.Contains(err.Error(), "Access denied"))
	for _, k := range srv.Keys() {
		assert.Equal(t, false, strings.HasSuffix(k, ".sql"))
	}
}

func Test_Perform_stream_mongodb(t *testing.T) {
	dir := t.TempDir()
	// fake mongodump: writes the archive to stdout only when --archive is given
	script := "#!/bin/sh\ncase \"$*\" in *--archive*) ;; *) echo 'no --archive' >&2; exit 1;; esac\nprintf 'MONGOARCHIVE'\n"
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "mongodump"), []byte(script), 0755))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	srv := s3mock.New()
	defer srv.Close()

	m := loadStreamModelDB(t, "stream_mongo_test", srv.URL, "compress_with:\n      type: gz", "mongodb")
	assert.NoError(t, m.Perform())

	for _, k := range srv.Keys() {
		if strings.HasSuffix(k, ".archive.gz") {
			b, _ := srv.Object(k)
			zr, err := gzip.NewReader(bytes.NewReader(b))
			assert.NoError(t, err)
			out, _ := io.ReadAll(zr)
			assert.Equal(t, "MONGOARCHIVE", string(out))
			return
		}
	}
	t.Fatalf("no .archive.gz object, got %v", srv.Keys())
}

func fakeBin(t *testing.T, name, script string) {
	dir := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+script+"\n"), 0755))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func Test_Perform_stream_postgresql(t *testing.T) {
	// the fake checks that no output file option is passed and prints to stdout
	fakeBin(t, "pg_dump", `case "$*" in *" -f "*) echo "unexpected -f" >&2; exit 1;; esac; echo "PGDUMP $PGPASSWORD"`)
	srv := s3mock.New()
	defer srv.Close()

	m := loadStreamModelDB(t, "stream_pg_test", srv.URL, "compress_with:\n      type: gz", "postgresql")
	m.Config.Databases["app"].Viper.Set("password", "secret")
	assert.NoError(t, m.Perform())

	for _, k := range srv.Keys() {
		if strings.HasSuffix(k, ".sql.gz") {
			b, _ := srv.Object(k)
			zr, err := gzip.NewReader(bytes.NewReader(b))
			assert.NoError(t, err)
			out, _ := io.ReadAll(zr)
			assert.Equal(t, "PGDUMP secret\n", string(out))
			return
		}
	}
	t.Fatalf("no .sql.gz object, got %v", srv.Keys())
}

func Test_Perform_stream_postgresql_all_databases(t *testing.T) {
	fakeBin(t, "pg_dumpall", `echo "PGDUMPALL"`)
	srv := s3mock.New()
	defer srv.Close()

	m := loadStreamModelDB(t, "stream_pgall_test", srv.URL, "", "postgresql")
	m.Config.Databases["app"].Viper.Set("all_databases", true)
	assert.NoError(t, m.Perform())

	for _, k := range srv.Keys() {
		if strings.HasSuffix(k, ".sql") {
			b, _ := srv.Object(k)
			assert.Equal(t, "PGDUMPALL\n", string(b))
			return
		}
	}
	t.Fatalf("no .sql object, got %v", srv.Keys())
}
