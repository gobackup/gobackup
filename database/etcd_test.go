package database

import (
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestEtcd_init(t *testing.T) {

	viper := viper.New()
	viper.Set("endpoint", "127.0.0.1:2379")
	viper.Set("args", "--foo --bar --baz")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/tmp/backups",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "etcd",
			Name:  "etcd1",
			Viper: viper,
		},
	)

	db := &Etcd{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)

	assert.Equal(t, db.build(), "etcdctl snapshot save "+db._dumpFilePath+" --endpoints 127.0.0.1:2379 --foo --bar --baz")
}
