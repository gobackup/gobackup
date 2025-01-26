package database

import (
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestRedis_init_for_copy(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("mode", "copy")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("invoke_save", "true")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "redis",
			Name:  "redis1",
			Viper: viper,
		},
	)

	db := &Redis{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)

	assert.Equal(t, db.invokeSave, false)
	assert.Equal(t, db.mode, redisModeCopy)
	assert.Equal(t, db._dumpFilePath, "/data/backups/redis/redis1/dump.rdb")
	assert.Equal(t, db.build(), "cp /var/db/redis/dump.rdb /data/backups/redis/redis1/dump.rdb")
}

func TestRedis_init_for_sync(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("mode", "sync")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("invoke_save", "true")
	viper.Set("args", "--tls --cacert redis_ca.pem")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "redis",
			Name:  "redis1",
			Viper: viper,
		},
	)

	db := &Redis{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)

	assert.Equal(t, db.build(), "redis-cli -h 1.2.3.4 -p 1234 -a pass1 --tls --cacert redis_ca.pem --rdb /data/backups/redis/redis1/dump.rdb")
}
