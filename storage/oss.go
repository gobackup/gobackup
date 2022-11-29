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

func (s *OSS) open() (err error) {
	s.viper.SetDefault("endpoint", "oss-cn-beijing.aliyuncs.com")
	s.viper.SetDefault("max_retries", 3)
	s.viper.SetDefault("path", "/")
	s.viper.SetDefault("timeout", 300)
	s.viper.SetDefault("threads", 1)

	s.endpoint = s.viper.GetString("endpoint")
	s.bucket = s.viper.GetString("bucket")
	s.accessKeyID = s.viper.GetString("access_key_id")
	s.accessKeySecret = s.viper.GetString("access_key_secret")
	s.path = s.viper.GetString("path")
	s.maxRetries = s.viper.GetInt("max_retries")
	s.timeout = s.viper.GetInt("timeout")
	s.threads = s.viper.GetInt("threads")

	// limit thread in 1..100
	if s.threads < 1 {
		s.threads = 1
	}
	if s.threads > 100 {
		s.threads = 100
	}

	logger.Info("endpoint:", s.endpoint)
	logger.Info("bucket:", s.bucket)

	ossClient, err := oss.New(s.endpoint, s.accessKeyID, s.accessKeySecret)
	if err != nil {
		return err
	}
	ossClient.Config.Timeout = uint(s.timeout)
	ossClient.Config.RetryTimes = uint(s.maxRetries)

	s.client, err = ossClient.Bucket(s.bucket)
	if err != nil {
		return err
	}

	return
}

func (s *OSS) close() {
}

func (s *OSS) upload(fileKey string) (err error) {
	remotePath := path.Join(s.path, fileKey)

	logger.Info("-> Uploading OSS...")
	err = s.client.UploadFile(remotePath, s.archivePath, ossPartSize, oss.Routines(s.threads))

	if err != nil {
		return err
	}
	logger.Info("Success")

	return nil
}

func (s *OSS) delete(fileKey string) (err error) {
	remotePath := path.Join(s.path, fileKey)
	err = s.client.DeleteObject(remotePath)
	return
}
