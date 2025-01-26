package database

import (
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestPostgreSQL_init(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("tables", []string{"foo", "bar"})
	viper.Set("exclude_tables", []string{"aa", "bb"})
	viper.Set("args", "--foo --bar --dar")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "postgresql",
			Name:  "postgresql1",
			Viper: viper,
		},
	)

	db := &PostgreSQL{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)

	assert.Equal(t, db.build(), "pg_dump --host=1.2.3.4 --port=1234 --username=user1 --table=foo --table=bar --exclude-table=aa --exclude-table=bb --foo --bar --dar my_db -f /data/backups/postgresql/postgresql1/my_db.sql")
}

func Test_PostgreSQL_prepareForSocket(t *testing.T) {
	db := &PostgreSQL{
		database:      "foo",
		socket:        "/var/run/postgresql/pg.5432",
		args:          "--foo",
		_dumpFilePath: "/tmp/foo.sql",
	}

	assert.Equal(t, db.build(), "pg_dump --host=/var/run/postgresql --port=5432 --foo foo -f /tmp/foo.sql")
}

func Test_PostgreSQL_compressionGzip(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("compress", "gzip:2")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "postgresql",
			Name:  "postgresql1",
			Viper: viper,
		},
	)

	db := &PostgreSQL{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)

	assert.Equal(t, db.build(), "pg_dump --host=1.2.3.4 --port=1234 --username=user1 --compress=gzip:2 --format=custom my_db -f /data/backups/postgresql/postgresql1/my_db.sql.gz")
}

func Test_PostgreSQL_compressionNotSupported(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("database", "my_db")
	viper.Set("username", "user1")
	viper.Set("password", "pass1")
	viper.Set("compress", "arj")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups/",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "postgresql",
			Name:  "postgresql1",
			Viper: viper,
		},
	)

	db := &PostgreSQL{
		Base: base,
	}

	err := db.init()
	assert.Error(t, err)
}
