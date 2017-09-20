package database

import (
	"github.com/huacnlee/gobackup/config"
	// "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMySQLPrepare(t *testing.T) {
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

	assert.Equal(t, mysql.dumpPath, "/foo/bar/mysql/mysql1")
	assert.Equal(t, mysql.dumpCommand, "mysqldump --host 127.0.0.2 --port 6378 -paaaa dummy_test")
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
