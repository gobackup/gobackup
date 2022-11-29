package storage

import (
	"testing"

	"github.com/huacnlee/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

type serviceInfo struct {
	name, endpoint, region string
}

func Test_providerName(t *testing.T) {
	var cases = map[string]serviceInfo{
		"s3":     {"AWS S3", "", "us-east-1"},
		"b2":     {"Backblaze B2", "us-east-001.backblazeb2.com", "us-east-001"},
		"us3":    {"UCloud US3", "s3-cn-bj.ufileos.com", "s3-cn-bj"},
		"cos":    {"QCloud COS", "cos.ap-nanjing.myqcloud.com", "ap-nanjing"},
		"kodo":   {"Qiniu Kodo", "s3-cn-east-1.qiniucs.com", "cn-east-1"},
		"r2":     {"Cloudflare R2", "us-east-1.r2.cloudflarestorage.com", "us-east-1"},
		"spaces": {"DigitalOcean Spaces", "nyc1.digitaloceanspaces.com", "nyc1"},
	}

	base := newBase(config.ModelConfig{}, "test")
	base.viper = viper.New()

	for service, info := range cases {
		s := &S3{Base: base, Service: service}
		s.init()

		assert.Equal(t, info.name, s.providerName(), "providerName for "+service)
		assert.Equal(t, info.endpoint, *s.defaultEndpoint(), "defaultEndpoint for "+service)
		assert.Equal(t, info.region, s.defaultRegion(), "defaultRegion for "+service)

		assert.Equal(t, info.region, s.viper.GetString("region"))
		assert.Equal(t, info.endpoint, s.viper.GetString("endpoint"))
		assert.Equal(t, "3", s.viper.GetString("max_retries"))
		assert.Equal(t, "300", s.viper.GetString("timeout"))
	}

}
