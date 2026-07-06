package config

import (
	"os"
	"testing"
	"time"

	"github.com/longbridgeapp/assert"
)

var (
	testConfigFile = "../gobackup_test.yml"
)

func init() {
	os.Setenv("S3_ACCESS_KEY_ID", "xxxxxxxxxxxxxxxxxxxx")
	os.Setenv("S3_SECRET_ACCESS_KEY", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	if err := Init(testConfigFile); err != nil {
		panic(err.Error())
	}
}

func TestModelsLength(t *testing.T) {
	assert.Equal(t, Exist, true)
	assert.Equal(t, len(Models), 5)
}

func TestModel(t *testing.T) {
	model := GetModelConfigByName("base_test")

	assert.Equal(t, model.Name, "base_test")
	assert.Equal(t, model.Description, "This is base test.")

	// compress_with
	assert.Equal(t, model.CompressWith.Type, "tgz")
	assert.NotNil(t, model.CompressWith.Viper)

	// encrypt_with
	assert.Equal(t, model.EncryptWith.Type, "openssl")
	assert.NotNil(t, model.EncryptWith.Viper)

	assert.Equal(t, model.DefaultStorage, "local")
	assert.Equal(t, model.Storages["local"].Type, "local")
	assert.Equal(t, model.Storages["local"].Viper.GetString("path"), "/Users/jason/Downloads/backup1")

	assert.Equal(t, model.Storages["scp"].Type, "scp")
	assert.Equal(t, model.Storages["scp"].Viper.GetString("host"), "your-host.com")

	// databases
	assert.Len(t, model.Databases, 3)

	// mysql
	db := model.GetDatabaseByName("dummy_test")
	assert.Equal(t, db.Name, "dummy_test")
	assert.Equal(t, db.Type, "mysql")
	assert.Equal(t, db.Viper.GetString("host"), "localhost")
	assert.Equal(t, db.Viper.GetString("port"), "3306")
	assert.Equal(t, db.Viper.GetString("database"), "dummy_test")
	assert.Equal(t, db.Viper.GetString("username"), "root")
	assert.Equal(t, db.Viper.GetString("password"), "123456")

	// redis
	db = model.GetDatabaseByName("redis1")
	assert.Equal(t, db.Name, "redis1")
	assert.Equal(t, db.Type, "redis")
	assert.Equal(t, db.Viper.GetString("mode"), "sync")
	assert.Equal(t, db.Viper.GetString("rdb_path"), "/var/db/redis/dump.rdb")
	assert.Equal(t, db.Viper.GetBool("invoke_save"), true)
	assert.Equal(t, db.Viper.GetString("password"), "456123")

	// redis
	db = model.GetDatabaseByName("postgresql")
	assert.Equal(t, db.Name, "postgresql")
	assert.Equal(t, db.Type, "postgresql")
	assert.Equal(t, db.Viper.GetString("host"), "localhost")

	// archive
	includes := model.Archive.GetStringSlice("includes")
	assert.Len(t, includes, 4)
	assert.Contains(t, includes, "/home/ubuntu/.ssh/")
	assert.Contains(t, includes, "/etc/nginx/nginx.conf")

	excludes := model.Archive.GetStringSlice("excludes")
	assert.Len(t, excludes, 2)
	assert.Contains(t, excludes, "/home/ubuntu/.ssh/known_hosts")

	// schedule
	schedule := model.Schedule
	assert.Equal(t, true, schedule.Enabled)
	assert.Equal(t, "5 4 * * sun", schedule.Cron)
}

func Test_otherModels(t *testing.T) {
	model := GetModelConfigByName("normal_files")

	// default_storage
	assert.Equal(t, model.DefaultStorage, "scp")

	// schedule
	schedule := model.Schedule
	assert.Equal(t, true, schedule.Enabled)
	assert.Equal(t, "", schedule.Cron)
	assert.Equal(t, "1day", schedule.Every)
	assert.Equal(t, "0:30", schedule.At)

	model = GetModelConfigByName("test_model")
	assert.Equal(t, false, model.Schedule.Enabled)
}

func Test_ScheduleConfig_String(t *testing.T) {
	schedule := ScheduleConfig{
		Enabled: true,
		Every:   "1day",
		At:      "0:30",
	}
	assert.Equal(t, schedule.String(), "every 1day at 0:30")

	schedule = ScheduleConfig{
		Enabled: true,
		Every:   "1day",
	}
	assert.Equal(t, schedule.String(), "every 1day")

	schedule = ScheduleConfig{
		Enabled: true,
		Cron:    "5 4 * * sun",
	}

	assert.Equal(t, schedule.String(), "cron 5 4 * * sun")

	schedule = ScheduleConfig{
		Enabled: false,
	}
	assert.Equal(t, schedule.String(), "disabled")
}

func TestExpandEnv(t *testing.T) {
	model := GetModelConfigByName("expand_env")

	assert.Equal(t, model.Storages["s3"].Type, "s3")
	assert.Equal(t, model.Storages["s3"].Viper.GetString("access_key_id"), "xxxxxxxxxxxxxxxxxxxx")
	assert.Equal(t, model.Storages["s3"].Viper.GetString("secret_access_key"), "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
}

func TestWebConfig(t *testing.T) {
	assert.Equal(t, Web.Host, "0.0.0.0")
	assert.Equal(t, Web.Port, "2703")
	assert.Equal(t, Web.Username, "gobackup")
	assert.Equal(t, Web.Password, "123456")
}

func TestInitWithNotExistsConfigFile(t *testing.T) {
	err := Init("config/path/not-exist.yml")
	assert.NotNil(t, err)
}

func TestInitWithEmptyDatabasesOrStorages(t *testing.T) {
	testCases := []struct {
		name               string
		configContent      string
		expectErr          bool
		expectedErrContain string
	}{
		{
			name: "storages_null",
			configContent: `models:
  myjob:
    databases:
      postgres_all:
        host: localhost
        port: 5432
    storages: null
`,
			expectErr:          true,
			expectedErrContain: "no storage found in model myjob",
		},
		{
			name: "storages_empty",
			configContent: `models:
  myjob:
    databases:
      postgres_all:
        host: localhost
        port: 5432
    storages:
`,
			expectErr:          true,
			expectedErrContain: "no storage found in model myjob",
		},
		{
			name: "storages_empty_map",
			configContent: `models:
  myjob:
    databases:
      postgres_all:
        host: localhost
        port: 5432
    storages: {}
`,
			expectErr:          true,
			expectedErrContain: "no storage found in model myjob",
		},
		{
			name: "storages_missing",
			configContent: `models:
  myjob:
    databases:
      postgres_all:
        host: localhost
        port: 5432
`,
			expectErr:          true,
			expectedErrContain: "no storage found in model myjob",
		},
		{
			name: "databases_null",
			configContent: `models:
  myjob:
    databases: null
    storages:
      local:
        type: local
        keep: 2
`,
			expectErr:          true,
			expectedErrContain: "model myjob must configure databases or archive",
		},
		{
			name: "databases_empty",
			configContent: `models:
  myjob:
    databases:
    storages:
      local:
        type: local
        keep: 2
`,
			expectErr:          true,
			expectedErrContain: "model myjob must configure databases or archive",
		},
		{
			name: "databases_empty_map",
			configContent: `models:
  myjob:
    databases: {}
    storages:
      local:
        type: local
        keep: 2
`,
			expectErr:          true,
			expectedErrContain: "model myjob must configure databases or archive",
		},
		{
			name: "databases_missing",
			configContent: `models:
  myjob:
    storages:
      local:
        type: local
        keep: 2
`,
			expectErr:          true,
			expectedErrContain: "model myjob must configure databases or archive",
		},
		{
			name: "archive_only_with_storage",
			configContent: `models:
  myjob:
    storages:
      local:
        type: local
        keep: 2
    archive:
      includes:
        - /etc/hosts
`,
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "gobackup-test-*.yml")
			assert.Nil(t, err)
			t.Cleanup(func() {
				_ = os.Remove(file.Name())
			})

			err = os.WriteFile(file.Name(), []byte(tc.configContent), 0644)
			assert.Nil(t, err)

			err = Init(file.Name())
			if tc.expectErr {
				assert.NotNil(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedErrContain)
				}
			} else {
				assert.Nil(t, err)
			}

			err = Init(testConfigFile)
			assert.Nil(t, err)
		})
	}
}

func TestWatchConfigToReload(t *testing.T) {
	err := Init(testConfigFile)
	assert.Nil(t, err)

	lastUpdatedAt := UpdatedAt.UnixNano()
	time.Sleep(1 * time.Millisecond)

	// Touch `testConfigFile` to trigger file changes event
	err = updateFile(testConfigFile)
	assert.Nil(t, err)

	// Wait for reload
	time.Sleep(10 * time.Millisecond)

	// check config reload updated_at
	assert.NotEqual(t, lastUpdatedAt, UpdatedAt.UnixNano())
}

func updateFile(path string) error {
	// Open file and write it again without any changes
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
