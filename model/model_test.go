package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/longbridgeapp/assert"

	"github.com/gobackup/gobackup/config"
)

// setupModel writes a config with an archive-only model and a local storage,
// using the given hooks yaml (already indented under `models.demo`).
func setupModel(t *testing.T, modelExtra string) (string, string) {
	t.Helper()
	dir := t.TempDir()

	// compressor changes cwd into the temp dump path which is removed after
	// Perform, so make sure every test starts from an existing directory
	cwd, _ := os.Getwd()
	assert.Nil(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	src := filepath.Join(dir, "src.txt")
	assert.Nil(t, os.WriteFile(src, []byte("hello"), 0644))
	marker := filepath.Join(dir, "marker")

	content := `workdir: ` + filepath.Join(dir, "work") + `
models:
  demo:
    archive:
      includes:
        - ` + src + `
    compress_with:
      type: tgz
    storages:
      local:
        type: local
        keep: 2
        path: ` + filepath.Join(dir, "backups") + `
` + strings.NewReplacer("MARKER", marker, "WORKDIR", filepath.Join(dir, "work")).Replace(modelExtra)

	cfg := filepath.Join(dir, "gobackup.yml")
	assert.Nil(t, os.WriteFile(cfg, []byte(content), 0644))
	assert.Nil(t, config.Init(cfg))

	return marker, dir
}

func readMarker(t *testing.T, file string) string {
	data, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return ""
	}
	assert.Nil(t, err)
	return string(data)
}

func TestPerformWithHooks(t *testing.T) {
	marker, dir := setupModel(t, `
    before_script: echo legacy_before >> MARKER
    after_script: echo legacy_after >> MARKER
    hooks:
      - name: lock
        when: before
        script: echo before >> MARKER
      - when: after
        on_exit: success
        script: echo after_success >> MARKER
      - when: after
        on_exit: failure
        script: echo after_failure >> MARKER
      - when: after
        script: test -d "WORKDIR" && echo temp_exists >> MARKER
`)

	m := GetModelByName("demo")
	assert.NotNil(t, m)
	err := m.Perform()
	assert.Nil(t, err)

	// legacy first, after hooks run before temp cleanup
	assert.Equal(t, "legacy_before\nbefore\nlegacy_after\nafter_success\ntemp_exists\n", readMarker(t, marker))

	// temp cleaned up after hooks
	_, statErr := os.Stat(m.Config.TempPath)
	assert.True(t, os.IsNotExist(statErr))

	// backup stored
	entries, err := os.ReadDir(filepath.Join(dir, "backups"))
	assert.Nil(t, err)
	assert.True(t, len(entries) > 0)
}

func TestPerformBeforeHookFails(t *testing.T) {
	marker, dir := setupModel(t, `
    hooks:
      - name: guard
        when: before
        script: echo guard >&2; exit 1
      - when: after
        on_exit: failure
        script: echo after_failure >> MARKER
      - when: after
        on_exit: success
        script: echo after_success >> MARKER
`)

	m := GetModelByName("demo")
	err := m.Perform()
	assert.NotNil(t, err)
	assert.Equal(t, `before hook "guard" (demo/run) failed: guard`, err.Error())

	assert.Equal(t, "after_failure\n", readMarker(t, marker))

	// nothing stored
	_, statErr := os.Stat(filepath.Join(dir, "backups"))
	assert.True(t, os.IsNotExist(statErr))
}

func TestPerformLegacyBeforeScriptFailureIsIgnored(t *testing.T) {
	marker, _ := setupModel(t, `
    before_script: exit 1
    after_script: echo legacy_after >> MARKER
`)

	m := GetModelByName("demo")
	err := m.Perform()
	assert.Nil(t, err)
	assert.Equal(t, "legacy_after\n", readMarker(t, marker))
}

// setupFullModel writes a config with sqlite database, archive, compress,
// encrypt, split and local storage, each level carrying hooks that append
// to the marker file.
func setupFullModel(t *testing.T, override map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	cwd, _ := os.Getwd()
	assert.Nil(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	src := filepath.Join(dir, "src.txt")
	assert.Nil(t, os.WriteFile(src, []byte("hello"), 0644))
	dbFile := filepath.Join(dir, "app.sqlite3")
	assert.Nil(t, os.WriteFile(dbFile, []byte(""), 0644))
	marker := filepath.Join(dir, "marker")

	hooks := func(level string, indent int) string {
		pad := strings.Repeat(" ", indent)
		block := `
hooks:
  - name: ` + level + `_before
    when: before
    script: ` + override[level+"_before"] + `echo ` + level + `_before >> MARKER
  - name: ` + level + `_after
    when: after
    script: echo ` + level + `_after >> MARKER
  - when: after
    on_exit: failure
    script: echo ` + level + `_failure >> MARKER`
		return strings.ReplaceAll(block, "\n", "\n"+pad) + "\n"
	}

	content := `workdir: ` + filepath.Join(dir, "work") + `
models:
  full:` + hooks("run", 4) + `
    databases:
      app:
        type: sqlite
        path: ` + dbFile + hooks("database", 8) + `
    archive:
      includes:
        - ` + src + hooks("archive", 6) + `
    compress_with:
      type: tgz` + hooks("compress", 6) + `
    encrypt_with:
      type: openssl
      password: 123456` + hooks("encrypt", 6) + `
    split_with:
      chunk_size: 1k` + hooks("split", 6) + `
    storages:
      local:
        type: local
        keep: 2
        path: ` + filepath.Join(dir, "backups") + hooks("storage", 8) + `
`
	content = strings.ReplaceAll(content, "MARKER", marker)

	cfg := filepath.Join(dir, "gobackup.yml")
	assert.Nil(t, os.WriteFile(cfg, []byte(content), 0644))
	assert.Nil(t, config.Init(cfg))

	return marker
}

func TestPerformHooksOnEveryLevel(t *testing.T) {
	marker := setupFullModel(t, nil)

	m := GetModelByName("full")
	assert.NotNil(t, m)
	err := m.Perform()
	assert.Nil(t, err)

	assert.Equal(t, strings.Join([]string{
		"run_before",
		"database_before", "database_after",
		"archive_before", "archive_after",
		"compress_before", "compress_after",
		"encrypt_before", "encrypt_after",
		"split_before", "split_after",
		"storage_before", "storage_after",
		"run_after",
	}, "\n")+"\n", readMarker(t, marker))
}

func TestPerformDatabaseBeforeHookFails(t *testing.T) {
	marker := setupFullModel(t, map[string]string{"database_before": "exit 1; "})

	m := GetModelByName("full")
	err := m.Perform()
	assert.NotNil(t, err)
	assert.Equal(t, `before hook "database_before" (full/database:app) failed: exit status 1`, err.Error())

	// database after hooks see failure, later stages are skipped, run after hooks see failure
	assert.Equal(t, strings.Join([]string{
		"run_before",
		"database_after", "database_failure",
		"run_after", "run_failure",
	}, "\n")+"\n", readMarker(t, marker))
}

func TestPerformCompressBeforeHookFails(t *testing.T) {
	marker := setupFullModel(t, map[string]string{"compress_before": "echo nope >&2; exit 2; "})

	m := GetModelByName("full")
	err := m.Perform()
	assert.NotNil(t, err)
	assert.Equal(t, `before hook "compress_before" (full/compress) failed: nope`, err.Error())

	assert.Equal(t, strings.Join([]string{
		"run_before",
		"database_before", "database_after",
		"archive_before", "archive_after",
		"compress_after", "compress_failure",
		"run_after", "run_failure",
	}, "\n")+"\n", readMarker(t, marker))
}
