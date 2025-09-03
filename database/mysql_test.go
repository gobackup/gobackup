package database

import (
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/spf13/viper"

	"github.com/longbridgeapp/assert"
)

func TestMySQL_init(t *testing.T) {
	vpr := viper.New()
	vpr.Set("host", "1.2.3.4")
	vpr.Set("port", "1234")
	vpr.Set("database", "my_db")
	vpr.Set("username", "user1")
	vpr.Set("password", "pass1")
	vpr.Set("tables", []string{"foo", "bar"})
	vpr.Set("exclude_tables", []string{"aa", "bb"})
	vpr.Set("args", "--a1 --a2 --a3")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "mysql",
			Name:  "mysql1",
			Viper: vpr,
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

func TestMySQL_dumpArgsWithDollarInPassword(t *testing.T) {
	base := newBase(
		config.ModelConfig{
			DumpPath: "/tmp/backups/",
		},
		config.SubConfig{
			Type: "mysql",
			Name: "mysql1",
		},
	)
	db := &MySQL{
		Base:     base,
		database: "dummy_test",
		host:     "127.0.0.1",
		port:     "3306",
		username: "test",
		password: "$ecure_pa$$word",
		args:     "--skip-ssl-verify-server-cert",
	}

	err := db.perform()
	if err != nil {
		println(err.Error())
	}

	assert.Equal(t, db.build(), "mysqldump --host 127.0.0.1 --port 3306 -u test -p'$ecure_pa$$word' --skip-ssl-verify-server-cert dummy_test --result-file=/tmp/backups/mysql/mysql1/dummy_test.sql")
}

func TestMySQL_dumpWithSocket(t *testing.T) {
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
		socket:   "/var/run/mysqld/mysqld.sock",
		args:     "--single-transaction --quick",
	}

	assert.Equal(t, db.build(), "mysqldump --socket /var/run/mysqld/mysqld.sock --single-transaction --quick dummy_test --result-file=/data/backups/mysql/mysql1/dummy_test.sql")
}
