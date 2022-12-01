package database

import (
	// "github.com/spf13/viper"
	"testing"

	"github.com/longbridgeapp/assert"
)

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
	assert.Equal(t, db.oplogOption(), "--oplog")
	db.oplog = false
	assert.Equal(t, db.oplogOption(), "")
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
	}
	expect := "mongodump --db=hello --username=foo --password=bar --authenticationDatabase=sssbbb --host=127.0.0.1 --port=4567 --oplog --out=/tmp/gobackup/test"
	assert.Equal(t, db.mongodump(), expect)
}
