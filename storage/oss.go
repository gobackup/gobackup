package storage

import (
	"path"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/huacnlee/gobackup/logger"
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
// threads: 1 (1 .. 100)
type OSS struct {
	Base
	endpoint        string
	bucket          string
	accessKeyID     string
	accessKeySecret string
	path            string
	maxRetries      int
	timeout         int
	client          *oss.Bucket
	threads         int
}

var (
	// 1 Mb one part
	ossPartSize int64 = 1024 * 1024
)

func (ctx *OSS) open() (err error) {
	ctx.viper.SetDefault("endpoint", "oss-cn-beijing.aliyuncs.com")
	ctx.viper.SetDefault("max_retries", 3)
	ctx.viper.SetDefault("path", "/")
	ctx.viper.SetDefault("timeout", 300)
	ctx.viper.SetDefault("threads", 1)

	ctx.endpoint = ctx.viper.GetString("endpoint")
	ctx.bucket = ctx.viper.GetString("bucket")
	ctx.accessKeyID = ctx.viper.GetString("access_key_id")
	ctx.accessKeySecret = ctx.viper.GetString("access_key_secret")
	ctx.path = ctx.viper.GetString("path")
	ctx.maxRetries = ctx.viper.GetInt("max_retries")
	ctx.timeout = ctx.viper.GetInt("timeout")
	ctx.threads = ctx.viper.GetInt("threads")

	// limit thread in 1..100
	if ctx.threads < 1 {
		ctx.threads = 1
	}
	if ctx.threads > 100 {
		ctx.threads = 100
	}

	logger.Info("endpoint:", ctx.endpoint)
	logger.Info("bucket:", ctx.bucket)

	ossClient, err := oss.New(ctx.endpoint, ctx.accessKeyID, ctx.accessKeySecret)
	if err != nil {
		return err
	}
	ossClient.Config.Timeout = uint(ctx.timeout)
	ossClient.Config.RetryTimes = uint(ctx.maxRetries)

	ctx.client, err = ossClient.Bucket(ctx.bucket)
	if err != nil {
		return err
	}

	return
}

func (ctx *OSS) close() {
}

func (ctx *OSS) upload(fileKey string) (err error) {
	remotePath := path.Join(ctx.path, fileKey)

	logger.Info("-> Uploading OSS...")
	err = ctx.client.UploadFile(remotePath, ctx.archivePath, ossPartSize, oss.Routines(ctx.threads))

	if err != nil {
		return err
	}
	logger.Info("Success")

	return nil
}

func (ctx *OSS) delete(fileKey string) (err error) {
	remotePath := path.Join(ctx.path, fileKey)
	err = ctx.client.DeleteObject(remotePath)
	return
}
