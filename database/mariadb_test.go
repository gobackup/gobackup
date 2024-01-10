package database

import (
	"github.com/gobackup/gobackup/config"
	"github.com/spf13/viper"

	// "github.com/spf13/viper"
	"testing"

	"github.com/longbridgeapp/assert"
)

func TestMariaDB_init(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("args", "--a1 --a2 --a3")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "mariadb",
			Name:  "mariadb1",
			Viper: viper,
		},
	)

	db := &MariaDB{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	script := db.build()
	assert.Equal(t, script, "mariadb-backup --backup --host 1.2.3.4 --port 1234 -u user1 -ppass1 --a1 --a2 --a3 --databases=my_db --target-dir=/data/backups/mariadb/mariadb1")
}

func TestMariaDB_dumpArgsWithAdditionalOptions(t *testing.T) {
	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		config.SubConfig{
			Type: "mariadb",
			Name: "mariadb1",
		},
	)
	db := &MariaDB{
		Base:     base,
		host:     "127.0.0.2",
		port:     "6378",
		password: "*&^92'",
		database: "my_db2",
		args:     "--datadir=/var/lib64/mysql",
	}

	assert.Equal(t, db.build(), "mariadb-backup --backup --host 127.0.0.2 --port 6378 -p*&^92' --datadir=/var/lib64/mysql --databases=my_db2 --target-dir=/data/backups/mariadb/mariadb1")
}