package database

import (
	// "github.com/spf13/viper"
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestAtlas_init(t *testing.T) {
	viper := viper.New()
	viper.Set("uri", "mongodb+srv://user1:pass1@example.abcdefg.mongodb.net/database?authSource=admin")
	viper.Set("exclude_tables", []string{"aa", "bb"})
	viper.Set("args", "--foo --bar --dar")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "atlas",
			Name:  "atlas1",
			Viper: viper,
		},
	)

	db := &Atlas{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)

	assert.Equal(t, db.build(), "mongodump --uri=mongodb+srv://user1:pass1@example.abcdefg.mongodb.net/database?authSource=admin --excludeCollection=aa --excludeCollection=bb --foo --bar --dar --out=/data/backups/atlas/atlas1")

}

func TestAtlas_mongodump(t *testing.T) {
	base := Base{
		dumpPath: "/tmp/gobackup/test",
	}
	db := &Atlas{
		Base: base,
		uri:  "mongodb+srv://user1:pass1@example.abcdefg.mongodb.net/database?authSource=admin",
		args: "--gzip",
	}
	assert.Equal(t, db.build(), "mongodump --uri=mongodb+srv://user1:pass1@example.abcdefg.mongodb.net/database?authSource=admin --gzip --out=/tmp/gobackup/test")
}
