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
	assert.Equal(t, db.allDatabases, false)
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
		allDatabases:           false,
	}
	assert.Equal(t, db.build(), "sqlpackage /Action:Export /SourceDatabaseName:test /SourceUser:testUser /SourcePassword:xxxYYY2$ /SourceServerName:127.0.0.1,1433 /SourceTrustServerCertificate:True /TargetFile:/tmp/gobackup/test/test.bacpac")
}

// mockMSSQL embeds MSSQL and overrides getAllDatabases for testing
type mockMSSQL struct {
	*MSSQL
	mockDatabases []string
	mockError     error
}

func (m *mockMSSQL) getAllDatabases() ([]string, error) {
	if m.mockError != nil {
		return nil, m.mockError
	}
	return m.mockDatabases, nil
}

func TestMSSQL_build_withAllDatabases(t *testing.T) {
	base := Base{
		dumpPath: "/tmp/gobackup/test",
	}

	mockDBs := []string{"db1", "db2", "db3"}
	mockMSSQL := &mockMSSQL{
		MSSQL: &MSSQL{
			Base:                   base,
			host:                   "127.0.0.1",
			port:                   "1433",
			username:               "testUser",
			password:               "testPass",
			trustServerCertificate: true,
			allDatabases:           true,
		},
		mockDatabases: mockDBs,
	}

	expectedCommands := []string{
		"sqlpackage /Action:Export /SourceDatabaseName:db1 /SourceUser:testUser /SourcePassword:testPass /SourceServerName:127.0.0.1,1433 /SourceTrustServerCertificate:True /TargetFile:/tmp/gobackup/test/db1.bacpac",
		"sqlpackage /Action:Export /SourceDatabaseName:db2 /SourceUser:testUser /SourcePassword:testPass /SourceServerName:127.0.0.1,1433 /SourceTrustServerCertificate:True /TargetFile:/tmp/gobackup/test/db2.bacpac",
		"sqlpackage /Action:Export /SourceDatabaseName:db3 /SourceUser:testUser /SourcePassword:testPass /SourceServerName:127.0.0.1,1433 /SourceTrustServerCertificate:True /TargetFile:/tmp/gobackup/test/db3.bacpac",
	}

	for i, dbName := range mockDBs {
		mockMSSQL.database = dbName
		actualCommand := mockMSSQL.build()
		assert.Equal(t, expectedCommands[i], actualCommand, "Command for database %s should match", dbName)
	}

	databases, err := mockMSSQL.getAllDatabases()
	assert.NoError(t, err)
	assert.Equal(t, mockDBs, databases)
}

func TestMSSQL_shouldSkipDatabase(t *testing.T) {
	db := &MSSQL{
		allDatabases:  false,
		skipDatabases: []string{"db1", "db2"},
	}
	assert.Equal(t, db.shouldSkipDatabase("db1"), false)
	assert.Equal(t, db.shouldSkipDatabase("db3"), false)

	db.allDatabases = true
	assert.Equal(t, db.shouldSkipDatabase("db1"), true)
	assert.Equal(t, db.shouldSkipDatabase("db2"), true)
	assert.Equal(t, db.shouldSkipDatabase("DB1"), true)
	assert.Equal(t, db.shouldSkipDatabase("Db2"), true)
	assert.Equal(t, db.shouldSkipDatabase("db3"), false)

	db.skipDatabases = []string{}
	assert.Equal(t, db.shouldSkipDatabase("db1"), false)
	assert.Equal(t, db.shouldSkipDatabase("anydb"), false)
}

func TestMSSQL_init_withSkipDatabases(t *testing.T) {
	viper := viper.New()
	viper.Set("host", "1.2.3.4")
	viper.Set("port", "1234")
	viper.Set("allDatabases", true)
	viper.Set("skipDatabases", []string{"test_db", "temp_db"})

	base := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
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
	assert.Equal(t, db.allDatabases, true)
	assert.Equal(t, db.skipDatabases, []string{"test_db", "temp_db"})
	assert.Equal(t, db.shouldSkipDatabase("test_db"), true)
	assert.Equal(t, db.shouldSkipDatabase("temp_db"), true)
	assert.Equal(t, db.shouldSkipDatabase("other_db"), false)
}
