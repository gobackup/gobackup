package database

import (
	"github.com/huacnlee/gobackup/config"
	// "github.com/spf13/viper"
	"testing"

	"github.com/longbridgeapp/assert"
)

func TestMySQL_dumpArgs(t *testing.T) {
	base := newBase(
		config.ModelConfig{
			DumpPath: "/tmp/gobackup/test",
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
		password: "aaaa",
	}

	dumpArgs := db.dumpArgs()
	assert.Equal(t, dumpArgs, []string{
		"--host",
		"127.0.0.2",
		"--port",
		"6378",
		"-paaaa",
		"dummy_test",
		"--result-file=/tmp/gobackup/test/mysql/mysql1/dummy_test.sql",
	})
}

func TestMySQL_dumpArgsWithAdditionalOptions(t *testing.T) {
	base := newBase(
		config.ModelConfig{
			DumpPath: "/tmp/gobackup/test",
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
		additionalOptions: []string{
			"--single-transaction",
			"--quick",
		},
	}

	dumpArgs := db.dumpArgs()
	assert.Equal(t, dumpArgs, []string{
		"--host",
		"127.0.0.2",
		"--port",
		"6378",
		"-p*&^92'",
		"--single-transaction",
		"--quick",
		"dummy_test",
		"--result-file=/tmp/gobackup/test/mysql/mysql1/dummy_test.sql",
	})
}

func TestMySQLPerform(t *testing.T) {
	model := config.GetModelConfigByName("base_test")
	assert.NotNil(t, model)

	dbConfig := model.GetDatabaseByName("dummy_test")
	assert.NotNil(t, dbConfig)

	base := newBase(*model, *dbConfig)
	db := &MySQL{Base: base}

	db.perform()
	assert.Equal(t, db.database, "dummy_test")
	assert.Equal(t, db.host, "localhost")
	assert.Equal(t, db.port, "3306")
	assert.Equal(t, db.username, "root")
	assert.Equal(t, db.password, "123456")
}
