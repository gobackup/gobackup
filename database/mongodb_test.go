package database

import (
	// "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMongoDB_credentialOptions(t *testing.T) {
	ctx := &MongoDB{
		username: "foo",
		password: "bar",
		authdb:   "sssbbb",
	}

	assert.Equal(t, ctx.credentialOptions(), "--username=foo --password=bar --authenticationDatabase=sssbbb")
}

func TestMongoDB_connectivityOptions(t *testing.T) {
	ctx := &MongoDB{
		host: "10.11.12.13",
		port: "12345",
	}
	assert.Equal(t, ctx.connectivityOptions(), "--host=10.11.12.13 --port=12345")

	ctx = &MongoDB{
		host: "10.11.12.13",
	}
	assert.Equal(t, ctx.connectivityOptions(), "--host=10.11.12.13")

	ctx = &MongoDB{
		port: "1122",
	}
	assert.Equal(t, ctx.connectivityOptions(), "--port=1122")
}

func TestMongoDB_oplogOption(t *testing.T) {
	ctx := &MongoDB{oplog: true}
	assert.Equal(t, ctx.oplogOption(), "--oplog")
	ctx.oplog = false
	assert.Equal(t, ctx.oplogOption(), "")
}

func TestMongoDB_mongodump(t *testing.T) {
	base := Base{
		dumpPath: "/tmp/gobackup/test",
	}
	ctx := &MongoDB{
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
	assert.Equal(t, ctx.mongodump(), expect)
}
