package storage

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hako/durafmt"
	"github.com/huacnlee/gobackup/logger"
)

// S3 - Amazon S3 storage
//
// type: s3
// bucket: gobackup-test
// region: us-east-1
// path: backups
// access_key_id: your-access-key-id
// secret_access_key: your-secret-access-key
// max_retries: 5
// timeout: 300
type S3 struct {
	Base
	bucket string
	path   string
	client *s3manager.Uploader
}

func (s *S3) open() (err error) {
	s.viper.SetDefault("region", "us-east-1")
	s.viper.SetDefault("max_retries", 3)
	s.viper.SetDefault("upload_timeout", "0")

	cfg := aws.NewConfig()
	endpoint := s.viper.GetString("endpoint")

	if len(endpoint) > 0 {
		cfg.Endpoint = aws.String(endpoint)
		cfg.S3ForcePathStyle = aws.Bool(true)
	}

	cfg.Credentials = credentials.NewStaticCredentials(
		s.viper.GetString("access_key_id"),
		s.viper.GetString("secret_access_key"),
		s.viper.GetString("token"),
	)

	cfg.Region = aws.String(s.viper.GetString("region"))
	cfg.MaxRetries = aws.Int(s.viper.GetInt("max_retries"))

	s.bucket = s.viper.GetString("bucket")
	s.path = s.viper.GetString("path")

	timeout := s.viper.GetInt("upload_timeout")
	uploadTimeoutDuration := time.Duration(timeout) * time.Second

	httpClient := &http.Client{Timeout: uploadTimeoutDuration}
	cfg.HTTPClient = httpClient

	sess := session.Must(session.NewSession(cfg))
	s.client = s3manager.NewUploader(sess)

	return
}

func (s *S3) close() {}

func (s *S3) upload(fileKey string) (err error) {
	f, err := os.Open(s.archivePath)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", s.archivePath, err)
	}

	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get size of file %q, %v", s.archivePath, err)
	}

	remotePath := path.Join(s.path, fileKey)

	input := &s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(remotePath),
		Body:   f,
	}

	logger.Info(fmt.Sprintf("-> S3 Uploading (%d MiB)...", info.Size()/(1024*1024)))

	start := time.Now()

	result, err := s.client.Upload(input, func(uploader *s3manager.Uploader) {
		// set the part size as low as possible to avoid timeouts and aborts
		// also set concurrency to 1 for the same reason
		var partSize int64 = 5242880 // 5MiB
		maxParts := math.Ceil(float64(info.Size() / partSize))

		// 10000 parts is the limit for AWS S3. If the resulting number of parts would exceed that limit, increase the
		// part size as much as needed but as little possible
		if maxParts > 10000 {
			partSize = int64(math.Ceil(float64(info.Size()) / 10000))
		}

		uploader.Concurrency = 1
		uploader.LeavePartsOnError = false
		uploader.PartSize = partSize
	})

	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	t := time.Now()
	elapsed := t.Sub(start)

	logger.Info("=>", result.Location)
	logger.Info("=>", fmt.Sprintf("s3://%s/%s", s.bucket, remotePath))
	rate := math.Ceil(float64(info.Size()) / (elapsed.Seconds() * 1024 * 1024))

	logger.Info(fmt.Sprintf("Duration %v, rate %.1f MiB/s", durafmt.Parse(elapsed).LimitFirstN(2).String(), rate))

	return nil
}

func (s *S3) delete(fileKey string) (err error) {
	remotePath := path.Join(s.path, fileKey)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(remotePath),
	}
	_, err = s.client.S3.DeleteObject(input)
	return
}
