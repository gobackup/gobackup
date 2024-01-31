package database

import (
	"os"
	"path"
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestEtcd_init(t *testing.T) {
	dest := path.Join(os.TempDir(), "etcd")
	cacertPath := path.Join(dest, "ca.crt")
	certPath := path.Join(dest, "server.crt")
	keyPath := path.Join(dest, "server.key")

	// create the certificate files for the test
	_ = os.MkdirAll(dest, 0750)
	_ = os.MkdirAll(cacertPath, 0750)
	_ = os.MkdirAll(certPath, 0750)
	_ = os.MkdirAll(keyPath, 0750)
	defer os.RemoveAll(dest)
	defer os.RemoveAll(cacertPath)
	defer os.RemoveAll(certPath)
	defer os.RemoveAll(keyPath)

	viper := viper.New()
	viper.Set("endpoints", "127.0.0.1:2379")
	viper.Set("user", "user1")
	viper.Set("password", "pass1")
	viper.Set("cacert", cacertPath)
	viper.Set("cert", certPath)
	viper.Set("key", keyPath)
	viper.Set("insecure-skip-tls-verify", "false")
	viper.Set("args", "--foo --bar --dar")

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

	assert.Equal(t, db.build(), "etcdctl snapshot save "+db._dumpFilePath+" --endpoints=127.0.0.1:2379 --user=\"user1\" --password=\"pass1\" --cacert=\""+cacertPath+"\" --cert=\""+certPath+"\" --key=\""+keyPath+"\" --insecure-skip-tls-verify=false --foo --bar --dar")
}
