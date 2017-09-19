package storage

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
	"path"
)

// OSS - Aliyun OSS storage
//
// type: oss
// bucket: gobackup-test
// endpoint: oss-cn-beijing.aliyuncs.com
// path: /
// access_key_id: your-access-key-id
// access_key_secret: your-access-key-secret
// max_retries: 5
// timeout: 300
type OSS struct {
	endpoint        string
	bucket          string
	accessKeyID     string
	accessKeySecret string
	path            string
	maxRetries      int
	timeout         int
}

var (
	// 4 Mb
	ossPartSize int64 = 4 * 1024 * 1024
)

func (ctx *OSS) perform(model config.ModelConfig, fileKey, archivePath string) error {
	ossViper := model.StoreWith.Viper
	ossViper.SetDefault("endpoint", "oss-cn-beijing.aliyuncs.com")
	ossViper.SetDefault("max_retries", 3)
	ossViper.SetDefault("path", "/")
	ossViper.SetDefault("timeout", 300)

	ctx.endpoint = ossViper.GetString("endpoint")
	ctx.bucket = ossViper.GetString("bucket")
	ctx.accessKeyID = ossViper.GetString("access_key_id")
	ctx.accessKeySecret = ossViper.GetString("access_key_secret")
	ctx.path = ossViper.GetString("path")
	ctx.maxRetries = ossViper.GetInt("max_retries")
	ctx.timeout = ossViper.GetInt("timeout")

	logger.Info("endpoint:", ctx.endpoint)
	logger.Info("bucket:", ctx.bucket)

	client, err := oss.New(ctx.endpoint, ctx.accessKeyID, ctx.accessKeySecret)
	if err != nil {
		return err
	}
	client.Config.Timeout = uint(ctx.timeout)
	client.Config.RetryTimes = uint(ctx.maxRetries)

	bucket, err := client.Bucket(ctx.bucket)
	if err != nil {
		return err
	}

	remotePath := path.Join(ctx.path, fileKey)

	logger.Info("-> Uploading OSS...")
	err = bucket.UploadFile(remotePath, archivePath, ossPartSize, oss.Routines(4))
	if err != nil {
		return err
	}
	logger.Info("Success")

	return nil
}
