package storage

import (
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

type serviceInfo struct {
	name, endpoint, region, storageClass string
	forcePathStyle                       bool
}

func Test_S3_open(t *testing.T) {
	viper := viper.New()
	viper.Set("bucket", "test-bucket")
	viper.Set("region", "us-east-2")

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

	storage := &S3{
		Base: base,
	}

	err = storage.open()
	assert.NoError(t, err)

	assert.Equal(t, "STANDARD_IA", storage.storageClass)
	assert.Equal(t, "test-bucket", storage.bucket)
	assert.Equal(t, "", storage.path)

	assert.Equal(t, 3, *storage.awsCfg.MaxRetries)
	assert.Equal(t, "us-east-2", *storage.awsCfg.Region)
	assert.Equal(t, 300, storage.awsCfg.HTTPClient.Timeout.Seconds())
}

func Test_providerName(t *testing.T) {
	var cases = map[string]serviceInfo{
		"s3":     {"AWS S3", "", "us-east-1", "STANDARD_IA", true},
		"b2":     {"Backblaze B2", "us-east-001.backblazeb2.com", "us-east-001", "STANDARD", true},
		"us3":    {"UCloud US3", "s3-cn-bj.ufileos.com", "s3-cn-bj", "ARCHIVE", true},
		"cos":    {"QCloud COS", "cos.ap-nanjing.myqcloud.com", "ap-nanjing", "STANDARD_IA", true},
		"kodo":   {"Qiniu Kodo", "s3-cn-east-1.qiniucs.com", "cn-east-1", "LINE", true},
		"r2":     {"Cloudflare R2", ".r2.cloudflarestorage.com", "us-east-1", "", true},
		"spaces": {"DigitalOcean Spaces", "nyc1.digitaloceanspaces.com", "nyc1", "STANDARD", true},
		"bos":    {"Baidu BOS", "s3.bj.bcebos.com", "bj", "STANDARD_IA", true},
		"oss":    {"Aliyun OSS", "oss-cn-hangzhou.aliyuncs.com", "cn-hangzhou", "STANDARD_IA", false},
		"obs":    {"Huawei OBS", "obs.cn-north-1.myhuaweicloud.com", "cn-north-1", "STANDARD_IA", true},
		"tos":    {"Volcengine TOS", "tos-s3-cn-beijing.volces.com", "cn-beijing", "STANDARD_IA", false},
		"upyun":  {"UpYun", "s3.api.upyun.com", "none", "", true},
		"minio":  {"MinIO", "", "us-east-1", "", true},
	}

	base, _ := newBase(config.ModelConfig{}, "test", config.SubConfig{})
	base.viper = viper.New()
	base.viper.SetDefault("bucket", "test-bucket")

	for service, info := range cases {
		s := &S3{Base: base, Service: service}
		s.init()

		assert.Equal(t, info.name, s.providerName(), "providerName for "+service)
		assert.Equal(t, info.endpoint, *s.defaultEndpoint(), "defaultEndpoint for "+service)
		assert.Equal(t, info.region, s.defaultRegion(), "defaultRegion for "+service)
		assert.Equal(t, info.storageClass, s.defaultStorageClass(), "defaultStorageClass for "+service)
		assert.Equal(t, info.forcePathStyle, s.forcePathStyle(), "forcePathStyle for "+service)

		assert.Equal(t, info.region, s.viper.GetString("region"))
		assert.Equal(t, info.endpoint, s.viper.GetString("endpoint"))
		assert.Equal(t, "3", s.viper.GetString("max_retries"))
		assert.Equal(t, "300", s.viper.GetString("timeout"))
	}

}
