package database

import (
	"github.com/gobackup/gobackup/config"
	"github.com/spf13/viper"

	// "github.com/spf13/viper"
	"testing"

	"github.com/longbridgeapp/assert"
)

func TestMySQL_init(t *testing.T) {
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
		config.ModelConfig{},
		// Creating a new base object.
		config.SubConfig{
			Type:  "mysql",
			Name:  "mysql1",
			Viper: viper,
		},
	)

	db := &MySQL{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)
	dumpArgs := db.dumpArgs()
	assert.Equal(t, dumpArgs, []string{
		"--host",
		"1.2.3.4",
		"--port",
		"1234",
		"-u", "user1",
		"-ppass1",
		"--ignore-table=my_db.aa", "--ignore-table=my_db.bb",
		"--foo --bar --dar",
		"my_db", "foo", "bar",
		"--result-file=mysql/mysql1/my_db.sql",
	})

	viper.Set("additional_options", "--bar --foo")
	err = db.init()
	assert.NoError(t, err)
	dumpArgs = db.dumpArgs()
	assert.Equal(t, dumpArgs, []string{
		"--host",
		"1.2.3.4",
		"--port",
		"1234",
		"-u", "user1",
		"-ppass1",
		"--ignore-table=my_db.aa", "--ignore-table=my_db.bb",
		"--bar --foo",
		"my_db", "foo", "bar",
		"--result-file=mysql/mysql1/my_db.sql",
	})
}

func TestMySQL_dumpArgsWithAdditionalOptions(t *testing.T) {
	base := newBase(
		config.ModelConfig{
			DumpPath: "/tmp/gobackup/test",
		},
		config.SubConfig{
			Type: "mysql",
			Name: "mysql1",
		},
	)
	db := &MySQL{
		Base:     base,
		database: "dummy_test",
		host:     "127.0.0.2",
		port:     "6378",
		password: "*&^92'",
		args:     "--single-transaction --quick",
	}

	dumpArgs := db.dumpArgs()
	assert.Equal(t, dumpArgs, []string{
		"--host",
		"127.0.0.2",
		"--port",
		"6378",
		"-p*&^92'",
		"--single-transaction --quick",
		"dummy_test",
		"--result-file=/tmp/gobackup/test/mysql/mysql1/dummy_test.sql",
	})
}

func TestMySQLPerform(t *testing.T) {
	model := config.GetModelConfigByName("base_test")
	assert.NotNil(t, model)

	dbConfig := model.GetDatabaseByName("dummy_test")
	assert.NotNil(t, dbConfig)

	base := newBase(*model, *dbConfig)
	db := &MySQL{Base: base}

	db.perform()
	assert.Equal(t, db.database, "dummy_test")
	assert.Equal(t, db.host, "localhost")
	assert.Equal(t, db.port, "3306")
	assert.Equal(t, db.username, "root")
	assert.Equal(t, db.password, "123456")
}
