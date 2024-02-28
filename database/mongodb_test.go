package database

import (
	// "github.com/spf13/viper"
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestMongoDB_init(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("authdb", "sssbbb")
	viper.Set("oplog", true)
	viper.Set("exclude_tables", []string{"aa", "bb"})
	viper.Set("args", "--foo --bar --dar")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "mongodb",
			Name:  "mongodb1",
			Viper: viper,
		},
	)

	db := &MongoDB{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)

	assert.Equal(t, db.build(), "mongodump --db=my_db --username=user1 --password=pass1 --authenticationDatabase=sssbbb --host=1.2.3.4 --port=1234 --oplog --excludeCollection=aa --excludeCollection=bb --foo --bar --dar --out=/data/backups/mongodb/mongodb1")

	viper.Set("uri", "mongodb://user1:pass1@1:2:3:4:1234/my_db?authSource=sssbbb")
	viper.Set("oplog", false)
	err = db.init()
	assert.NoError(t, err)
	assert.Equal(t, db.build(), "mongodump --uri=mongodb://user1:pass1@1:2:3:4:1234/my_db?authSource=sssbbb --excludeCollection=aa --excludeCollection=bb --foo --bar --dar --out=/data/backups/mongodb/mongodb1")
}

func TestMongoDB_credentialOptions(t *testing.T) {
	db := &MongoDB{
		username: "foo",
		password: "bar",
		authdb:   "sssbbb",
	}

	assert.Equal(t, db.credentialOptions(), "--username=foo --password=bar --authenticationDatabase=sssbbb")
}

func TestMongoDB_connectivityOptions(t *testing.T) {
	db := &MongoDB{
		host: "10.11.12.13",
		port: "12345",
	}
	assert.Equal(t, db.connectivityOptions(), "--host=10.11.12.13 --port=12345")

	db = &MongoDB{
		host: "10.11.12.13",
	}
	assert.Equal(t, db.connectivityOptions(), "--host=10.11.12.13")

	db = &MongoDB{
		port: "1122",
	}
	assert.Equal(t, db.connectivityOptions(), "--port=1122")
}

func TestMongoDB_oplogOption(t *testing.T) {
	db := &MongoDB{oplog: true}
	assert.Equal(t, db.additionOption(), "--oplog")
	db.oplog = false
	assert.Equal(t, db.additionOption(), "")
}

func TestMongoDB_mongodump(t *testing.T) {
	base := Base{
		dumpPath: "/tmp/gobackup/test",
	}
	db := &MongoDB{
		Base:     base,
		host:     "127.0.0.1",
		port:     "4567",
		database: "hello",
		username: "foo",
		password: "bar",
		authdb:   "sssbbb",
		oplog:    true,
		args:     "--collection foo --gzip",
	}
	assert.Equal(t, db.build(), "mongodump --db=hello --username=foo --password=bar --authenticationDatabase=sssbbb --host=127.0.0.1 --port=4567 --oplog --collection foo --gzip --out=/tmp/gobackup/test")

	db = &MongoDB{
		Base:          base,
		uri:           "mongodb://foo:bar@127:0:0:1:4567/hello?authSource=sssbbb",
		excludeTables: []string{"aa", "bb"},
	}
	assert.Equal(t, db.build(), "mongodump --uri=mongodb://foo:bar@127:0:0:1:4567/hello?authSource=sssbbb --excludeCollection=aa --excludeCollection=bb --out=/tmp/gobackup/test")
}
