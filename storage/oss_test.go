package storage

import (
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func Test_OSS_init(t *testing.T) {
	viper := viper.New()
	viper.Set("bucket", "test-bucket")

	base, err := newBase(
		config.ModelConfig{
			DumpPath: "/data/backups",
		},
		"foo/bar",
		// Creating a new base object.
		config.SubConfig{
			Type:  "mongodb",
			Name:  "mongodb1",
			Viper: viper,
		},
	)
	assert.NoError(t, err)

	storage := &OSS{
		Base: base,
	}

	err = storage.open()
	assert.NoError(t, err)

	assert.Equal(t, oss.StorageClassType("Archive"), storage.storageClass)

	assert.Equal(t, "oss-cn-beijing.aliyuncs.com", storage.endpoint)
	assert.Equal(t, "/", storage.path)
	assert.Equal(t, 3, storage.maxRetries)
	assert.Equal(t, 1, storage.threads)
	assert.Equal(t, 300, storage.timeout)
	assert.Equal(t, 300, storage.client.Client.Config.Timeout)
}
