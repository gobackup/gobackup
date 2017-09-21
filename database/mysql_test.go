package database

import (
	"github.com/huacnlee/gobackup/config"
	// "github.com/spf13/viper"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMySQL_dumpArgs(t *testing.T) {
	mysql := &MySQL{
		Name:     "mysql1",
		database: "dummy_test",
		host:     "127.0.0.2",
		port:     "6378",
		password: "aaaa",
		model: config.ModelConfig{
			DumpPath: "/foo/bar",
		},
	}
	err := mysql.prepare()
	assert.NoError(t, err)

	dumpArgs := mysql.dumpArgs()
	assert.Equal(t, dumpArgs, []string{
		"--host",
		"127.0.0.2",
		"--port",
		"6378",
		"-p'aaaa'",
		"dummy_test",
		"--result-file=/foo/bar/mysql/mysql1/dummy_test.sql",
	})
}

func TestMySQL_dumpArgsWithAdditionalOptions(t *testing.T) {
	mysql := &MySQL{
		Name:              "mysql1",
		database:          "dummy_test",
		host:              "127.0.0.2",
		port:              "6378",
		password:          "aaaa",
		additionalOptions: "--single-transaction --quick",
		model: config.ModelConfig{
			DumpPath: "/foo/bar",
		},
	}
	err := mysql.prepare()
	assert.NoError(t, err)

	assert.Equal(t, mysql.dumpPath, "/foo/bar/mysql/mysql1")

	dumpArgs := mysql.dumpArgs()
	assert.Equal(t, dumpArgs, []string{
		"--host",
		"127.0.0.2",
		"--port",
		"6378",
		"-p'aaaa'",
		"--single-transaction --quick",
		"dummy_test",
		"--result-file=/foo/bar/mysql/mysql1/dummy_test.sql",
	})
}

func TestMySQLPerform(t *testing.T) {
	mysql := &MySQL{}

	model := config.GetModelByName("base_test")
	assert.NotNil(t, model)

	dbConfig := model.GetDatabaseByName("dummy_test")
	assert.NotNil(t, dbConfig)

	mysql.perform(*model, *dbConfig)
	assert.Equal(t, mysql.database, "dummy_test")
	assert.Equal(t, mysql.host, "localhost")
	assert.Equal(t, mysql.port, "3306")
	assert.Equal(t, mysql.username, "root")
	assert.Equal(t, mysql.password, "123456")
}
