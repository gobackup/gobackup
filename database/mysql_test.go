package database

import (
	"github.com/gobackup/gobackup/config"
	"github.com/spf13/viper"

	// "github.com/spf13/viper"
	"testing"

	"github.com/longbridgeapp/assert"
)

func TestMySQL_init(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("tables", []string{"foo", "bar"})
	viper.Set("exclude_tables", []string{"aa", "bb"})
	viper.Set("args", "--a1 --a2 --a3")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "mysql",
			Name:  "mysql1",
			Viper: viper,
		},
	)

	db := &MySQL{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	script := db.build()
	assert.Equal(t, script, "mysqldump --host 1.2.3.4 --port 1234 -u user1 -ppass1 --ignore-table=my_db.aa --ignore-table=my_db.bb --a1 --a2 --a3 my_db foo bar --result-file=/data/backups/mysql/mysql1/my_db.sql")
}

func TestMySQL_dumpArgsWithAdditionalOptions(t *testing.T) {
	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		config.SubConfig{
			Type: "mysql",
			Name: "mysql1",
		},
	)
	db := &MySQL{
		Base:     base,
		database: "dummy_test",
		host:     "127.0.0.2",
		port:     "6378",
		password: "*&^92'",
		args:     "--single-transaction --quick",
	}

	assert.Equal(t, db.build(), "mysqldump --host 127.0.0.2 --port 6378 -p*&^92' --single-transaction --quick dummy_test --result-file=/data/backups/mysql/mysql1/dummy_test.sql")
}

func TestMySQL_allDatabases(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "127.0.0.1")
	viper.Set("port", "3306")
	viper.Set("username", "root")
	viper.Set("password", "secret")
	viper.Set("all_databases", true)

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		config.SubConfig{
			Type:  "mysql",
			Name:  "mysql1",
			Viper: viper,
		},
	)

	db := &MySQL{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	assert.Equal(t, db.allDatabases, true)

	script := db.build()
	assert.Contains(t, script, "--all-databases")
	assert.Contains(t, script, "--result-file=/data/backups/mysql/mysql1/all-databases.sql")
	assert.Equal(t, script, "mysqldump --host 127.0.0.1 --port 3306 -u root -psecret --all-databases --result-file=/data/backups/mysql/mysql1/all-databases.sql")
}

func TestMySQL_allDatabasesWithTablesError(t *testing.T) {
	viper := viper.New()
	viper.Set("all_databases", true)
	viper.Set("tables", []string{"foo", "bar"})

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		config.SubConfig{
			Type:  "mysql",
			Name:  "mysql1",
			Viper: viper,
		},
	)

	db := &MySQL{
		Base: base,
	}

	err := db.init()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tables and exclude_tables options are not supported when using all_databases: true")
}

func TestMySQL_allDatabasesWithExcludeTablesError(t *testing.T) {
	viper := viper.New()
	viper.Set("all_databases", true)
	viper.Set("exclude_tables", []string{"test_table"})

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		config.SubConfig{
			Type:  "mysql",
			Name:  "mysql1",
			Viper: viper,
		},
	)

	db := &MySQL{
		Base: base,
	}

	err := db.init()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tables and exclude_tables options are not supported when using all_databases: true")
}
