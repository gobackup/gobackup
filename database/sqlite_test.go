package database

import (
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestSQLite_init(t *testing.T) {
	viper := viper.New()
	viper.Set("path", "/var/db/my.sqlite")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "sqlite",
			Name:  "sqlite1",
			Viper: viper,
		},
	)

	db := &SQLite{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)

	assert.Equal(t, db._dumpFilePath, "/data/backups/sqlite/sqlite1/my.sql")
	assert.Equal(t, db.buildArgs(), []string{"/var/db/my.sqlite", ".output /data/backups/sqlite/sqlite1/my.sql", ".dump"})
}
