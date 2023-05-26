package database

import (
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestMSSQL_init(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("database", "AdventureWorks")
	viper.Set("username", "AdminUser")
	viper.Set("password", "AdminPassword1")
	viper.Set("trustServerCertificate", true)
	viper.Set("args", "/OverwriteFiles:True /MaxParallelism:4")

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		// Creating a new base object.
		config.SubConfig{
			Type:  "mssql",
			Name:  "mssql1",
			Viper: viper,
		},
	)

	db := &MSSQL{
		Base: base,
	}

	err := db.init()
	assert.NoError(t, err)

	assert.Equal(t, db.build(), "sqlpackage /Action:Export /SourceDatabaseName:AdventureWorks /SourceUser:AdminUser /SourcePassword:AdminPassword1 /SourceServerName:1.2.3.4,1234 /SourceTrustServerCertificate:True /OverwriteFiles:True /MaxParallelism:4 /TargetFile:/data/backups/mssql/mssql1/AdventureWorks.bacpac")
}

func TestMSSQL_credentialOptions(t *testing.T) {
	db := &MSSQL{
		username: "AdminUser",
		password: "AdminPassword1",
	}

	assert.Equal(t, db.credentialOptions(), "/SourceUser:AdminUser /SourcePassword:AdminPassword1")
}

func TestMSSQL_connectivityOptions(t *testing.T) {
	db := &MSSQL{
		host: "10.11.12.13",
		port: "12345",
	}
	assert.Equal(t, db.connectivityOptions(), "/SourceServerName:10.11.12.13,12345")

	db = &MSSQL{
		host: "10.11.12.13",
	}
	assert.Equal(t, db.connectivityOptions(), "/SourceServerName:10.11.12.13,1433")

	db = &MSSQL{
		port: "1122",
	}
	assert.Equal(t, db.connectivityOptions(), "/SourceServerName:127.0.0.1,1122")

	db = &MSSQL{}
	assert.Equal(t, db.connectivityOptions(), "/SourceServerName:127.0.0.1,1433")
}

func TestMSSQL_trustServerCertificateOption(t *testing.T) {
	db := &MSSQL{trustServerCertificate: true}
	assert.Equal(t, db.additionOption(), "/SourceTrustServerCertificate:True")
	db.trustServerCertificate = false
	assert.Equal(t, db.additionOption(), "")
	db = &MSSQL{}
	assert.Equal(t, db.additionOption(), "")
}

func TestMSSQL_sqlpackage(t *testing.T) {
	base := Base{
		dumpPath: "/tmp/gobackup/test",
	}
	db := &MSSQL{
		Base:                   base,
		host:                   "127.0.0.1",
		port:                   "1433",
		database:               "test",
		username:               "testUser",
		password:               "xxxYYY2$",
		trustServerCertificate: true,
	}
	assert.Equal(t, db.build(), "sqlpackage /Action:Export /SourceDatabaseName:test /SourceUser:testUser /SourcePassword:xxxYYY2$ /SourceServerName:127.0.0.1,1433 /SourceTrustServerCertificate:True /TargetFile:/tmp/gobackup/test/test.bacpac")
}
