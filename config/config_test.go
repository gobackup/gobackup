package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	Init("../gobackup_test.yml")
}

func TestModelsLength(t *testing.T) {
	assert.Equal(t, Exist, true)
	assert.Len(t, Models, 5)
}

func TestModel(t *testing.T) {
	model := GetModelByName("base_test")

	assert.Equal(t, model.Name, "base_test")

	// compress_with
	assert.Equal(t, model.CompressWith.Type, "tgz")
	assert.NotNil(t, model.CompressWith.Viper)

	// encrypt_with
	assert.Equal(t, model.EncryptWith.Type, "openssl")
	assert.NotNil(t, model.EncryptWith.Viper)

	// store_with
	assert.Equal(t, model.StoreWith.Type, "local")
	assert.Equal(t, model.StoreWith.Viper.GetString("path"), "/Users/jason/Downloads/backup1")

	// databases
	assert.Len(t, model.Databases, 3)

	// mysql
	db := model.GetDatabaseByName("dummy_test")
	assert.Equal(t, db.Name, "dummy_test")
	assert.Equal(t, db.Type, "mysql")
	assert.Equal(t, db.Viper.GetString("host"), "localhost")
	assert.Equal(t, db.Viper.GetString("port"), "3306")
	assert.Equal(t, db.Viper.GetString("database"), "dummy_test")
	assert.Equal(t, db.Viper.GetString("username"), "root")
	assert.Equal(t, db.Viper.GetString("password"), "123456")

	// redis
	db = model.GetDatabaseByName("redis1")
	assert.Equal(t, db.Name, "redis1")
	assert.Equal(t, db.Type, "redis")
	assert.Equal(t, db.Viper.GetString("mode"), "sync")
	assert.Equal(t, db.Viper.GetString("rdb_path"), "/var/db/redis/dump.rdb")
	assert.Equal(t, db.Viper.GetBool("invoke_save"), true)
	assert.Equal(t, db.Viper.GetString("password"), "456123")

	// redis
	db = model.GetDatabaseByName("postgresql")
	assert.Equal(t, db.Name, "postgresql")
	assert.Equal(t, db.Type, "postgresql")
	assert.Equal(t, db.Viper.GetString("host"), "localhost")

	// archive
	includes := model.Archive.GetStringSlice("includes")
	assert.Len(t, includes, 4)
	assert.Contains(t, includes, "/home/ubuntu/.ssh/")
	assert.Contains(t, includes, "/etc/nginx/nginx.conf")

	excludes := model.Archive.GetStringSlice("excludes")
	assert.Len(t, excludes, 2)
	assert.Contains(t, excludes, "/home/ubuntu/.ssh/known_hosts")
}
