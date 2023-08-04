package storage

import (
	"testing"

	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func Test_Azure_open(t *testing.T) {
	viper := viper.New()
	viper.Set("tenant_id", "tenant-xxx")
	viper.Set("client_id", "client-xxx")
	viper.Set("client_secret", "client-secret-xxx")

	viper.Set("timeout", 20)

	s := Azure{}
	s.viper = viper

	viper.Set("bucket", "hello")

	assert.Nil(t, s.open())
	assert.Equal(t, "https://hello.blob.core.windows.net", s.getBucketURL())
	assert.Equal(t, "gobackup", s.container)
	assert.Equal(t, "hello", s.account)
	assert.Equal(t, 20, s.timeout.Seconds())

	assert.Equal(t, "https://hello.blob.core.windows.net", s.client.URL())

	viper.Set("container", "my-container")
	viper.Set("account", "hello1")
	assert.Nil(t, s.open())
	assert.Equal(t, "hello1", s.account)
	assert.Equal(t, "https://hello1.blob.core.windows.net", s.getBucketURL())
	assert.Equal(t, "my-container", s.container)
}
