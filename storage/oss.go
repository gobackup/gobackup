package storage

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
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
	Base
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

func (ctx *OSS) perform() error {
	ctx.viper.SetDefault("endpoint", "oss-cn-beijing.aliyuncs.com")
	ctx.viper.SetDefault("max_retries", 3)
	ctx.viper.SetDefault("path", "/")
	ctx.viper.SetDefault("timeout", 300)

	ctx.endpoint = ctx.viper.GetString("endpoint")
	ctx.bucket = ctx.viper.GetString("bucket")
	ctx.accessKeyID = ctx.viper.GetString("access_key_id")
	ctx.accessKeySecret = ctx.viper.GetString("access_key_secret")
	ctx.path = ctx.viper.GetString("path")
	ctx.maxRetries = ctx.viper.GetInt("max_retries")
	ctx.timeout = ctx.viper.GetInt("timeout")

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

	remotePath := path.Join(ctx.path, ctx.fileKey)

	logger.Info("-> Uploading OSS...")
	err = bucket.UploadFile(remotePath, ctx.archivePath, ossPartSize, oss.Routines(4))
	if err != nil {
		return err
	}
	logger.Info("Success")

	return nil
}
