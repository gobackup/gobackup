package database

import (
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestFirebirdSQL_init(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "3051")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("role", "role1")
	viper.Set("password", "pass1")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "firebird",
			Name:  "firebird1",
			Viper: viper,
		},
	)

	db := &Firebird{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)

	assert.Equal(t, db.build(), "gbak -b -user user1 -pass pass1 -role role1 1.2.3.4/3051:my_db /data/backups/firebird/firebird1/my_db.fbk")
}

func TestFirebirdSQL_withDatabasePath(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "3050")
	viper.Set("database", "/var/databases/test.fdb")
	viper.Set("username", "user1")
	viper.Set("role", "role1")
	viper.Set("password", "pass1")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "firebird",
			Name:  "firebird1",
			Viper: viper,
		},
	)

	db := &Firebird{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)

	assert.Equal(t, db.build(), "gbak -b -user user1 -pass pass1 -role role1 1.2.3.4/3050:/var/databases/test.fdb /data/backups/firebird/firebird1/test.fbk")
}

func TestFirebirdSQL_withoutHost(t *testing.T) {
	viper := viper.New()
	viper.Set("database", "DB_ALIAS")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "firebird",
			Name:  "firebird1",
			Viper: viper,
		},
	)

	db := &Firebird{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)

	assert.Equal(t, db.build(), "gbak -b -user user1 -pass pass1 127.0.0.1/3050:DB_ALIAS /data/backups/firebird/firebird1/DB_ALIAS.fbk")
}
