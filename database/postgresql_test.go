package database

import (
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestPostgreSQL_init(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("tables", []string{"foo", "bar"})
	viper.Set("exclude_tables", []string{"aa", "bb"})
	viper.Set("args", "--foo --bar --dar")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "postgresql",
			Name:  "postgresql1",
			Viper: viper,
		},
	)

	db := &PostgreSQL{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)

	assert.Equal(t, db.build(), "pg_dump --host=1.2.3.4 --port=1234 --username=user1 --table=foo --table=bar --exclude-table=aa --exclude-table=bb --foo --bar --dar my_db -f /data/backups/postgresql/postgresql1/my_db.sql")
}

func Test_PostgreSQL_prepareForSocket(t *testing.T) {
	db := &PostgreSQL{
		database:      "foo",
		socket:        "/var/run/postgresql/pg.5432",
		args:          "--foo",
		_dumpFilePath: "/tmp/foo.sql",
	}

	assert.Equal(t, db.build(), "pg_dump --host=/var/run/postgresql --port=5432 --foo foo -f /tmp/foo.sql")
}

func Test_PostgreSQL_compressionGzip(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("compress", "gzip:2")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "postgresql",
			Name:  "postgresql1",
			Viper: viper,
		},
	)

	db := &PostgreSQL{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)

	assert.Equal(t, db.build(), "pg_dump --host=1.2.3.4 --port=1234 --username=user1 --compress=gzip:2 --format=custom my_db -f /data/backups/postgresql/postgresql1/my_db.sql.gz")
}

func Test_PostgreSQL_compressionNotSupported(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("compress", "arj")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "postgresql",
			Name:  "postgresql1",
			Viper: viper,
		},
	)

	db := &PostgreSQL{
		Base: base,
	}

	err := db.init()
	assert.Error(t, err)
}

func TestPostgreSQL_allDatabases(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("all_databases", true)
	viper.Set("args", "--foo --bar")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		config.SubConfig{
			Type:  "postgresql",
			Name:  "postgresql1",
			Viper: viper,
		},
	)

	db := &PostgreSQL{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	script := db.build()
	assert.Equal(t, script, "pg_dumpall --host=1.2.3.4 --port=1234 --username=user1 --foo --bar > /data/backups/postgresql/postgresql1/all_databases.sql")
}

func TestPostgreSQL_allDatabasesWithoutDatabase(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "localhost")
	viper.Set("port", "5432")
	viper.Set("username", "postgres")
	viper.Set("all_databases", true)

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		config.SubConfig{
			Type:  "postgresql",
			Name:  "postgresql1",
			Viper: viper,
		},
	)

	db := &PostgreSQL{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	script := db.build()
	assert.Equal(t, script, "pg_dumpall --host=localhost --port=5432 --username=postgres > /data/backups/postgresql/postgresql1/all_databases.sql")
}

func TestPostgreSQL_allDatabasesWithSocket(t *testing.T) {
	viper := viper.New()
	viper.Set("socket", "/var/run/postgresql/pg.5432")
	viper.Set("username", "user1")
	viper.Set("all_databases", true)

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		config.SubConfig{
			Type:  "postgresql",
			Name:  "postgresql1",
			Viper: viper,
		},
	)

	db := &PostgreSQL{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	script := db.build()
	assert.Equal(t, script, "pg_dumpall --host=/var/run/postgresql --port=5432 --username=user1 > /data/backups/postgresql/postgresql1/all_databases.sql")
}

func TestPostgreSQL_allDatabasesWithCompression(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("username", "user1")
	viper.Set("all_databases", true)
	viper.Set("compress", "gzip")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		config.SubConfig{
			Type:  "postgresql",
			Name:  "postgresql1",
			Viper: viper,
		},
	)

	db := &PostgreSQL{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	script := db.build()
	assert.Equal(t, script, "pg_dumpall --host=1.2.3.4 --port=1234 --username=user1 > /data/backups/postgresql/postgresql1/all_databases.sql.gz")
}

func TestPostgreSQL_allDatabasesSeparateFiles(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("all_databases", true)
	viper.Set("separate_files", true)
	viper.Set("exclude_databases", []string{"staging"})

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		config.SubConfig{
			Type:  "postgresql",
			Name:  "postgresql1",
			Viper: viper,
		},
	)

	db := &PostgreSQL{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	assert.Equal(t, db.separateFiles, true)

	assert.Equal(t, db.listDatabasesArgs(), []string{
		"--host=1.2.3.4", "--port=1234", "--username=user1",
		"--no-align", "--tuples-only", "--dbname=postgres",
		"--command=SELECT datname FROM pg_database WHERE datallowconn AND NOT datistemplate",
	})

	assert.Equal(t, db.buildGlobals(), "pg_dumpall --host=1.2.3.4 --port=1234 --username=user1 --globals-only -f /data/backups/postgresql/postgresql1/globals.sql")
	assert.Equal(t, db.buildDatabase("my_db", db.dumpFilePath("my_db")), "pg_dump --host=1.2.3.4 --port=1234 --username=user1 my_db -f /data/backups/postgresql/postgresql1/my_db.sql")

	assert.Equal(t, db.shouldSkipDatabase("staging"), true)
	assert.Equal(t, db.shouldSkipDatabase("my_db"), false)
}

func TestPostgreSQL_allDatabasesSeparateFilesWithCompression(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("username", "user1")
	viper.Set("all_databases", true)
	viper.Set("separate_files", true)
	viper.Set("compress", "gzip")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		config.SubConfig{
			Type:  "postgresql",
			Name:  "postgresql1",
			Viper: viper,
		},
	)

	db := &PostgreSQL{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	assert.Equal(t, db.buildDatabase("my_db", db.dumpFilePath("my_db")), "pg_dump --host=1.2.3.4 --port=1234 --username=user1 --compress=gzip --format=custom my_db -f /data/backups/postgresql/postgresql1/my_db.sql.gz")
}

func TestPostgreSQL_separateFilesWithoutAllDatabasesError(t *testing.T) {
	viper := viper.New()
	viper.Set("database", "my_db")
	viper.Set("separate_files", true)

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		config.SubConfig{
			Type:  "postgresql",
			Name:  "postgresql1",
			Viper: viper,
		},
	)

	db := &PostgreSQL{
		Base: base,
	}

	err := db.init()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "separate_files option is only supported when using all_databases: true")
}
