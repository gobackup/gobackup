package storage

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hako/durafmt"
	"github.com/huacnlee/gobackup/logger"
	"math"
	"net/http"
	"os"
	"path"
	"time"
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

func (ctx *S3) open() (err error) {
	ctx.viper.SetDefault("region", "us-east-1")
	ctx.viper.SetDefault("max_retries", 3)
	ctx.viper.SetDefault("upload_timeout", "0")

	cfg := aws.NewConfig()
	endpoint := ctx.viper.GetString("endpoint")

	if len(endpoint) > 0 {
		cfg.Endpoint = aws.String(endpoint)
		cfg.S3ForcePathStyle = aws.Bool(true)
	}

	cfg.Credentials = credentials.NewStaticCredentials(
		ctx.viper.GetString("access_key_id"),
		ctx.viper.GetString("secret_access_key"),
		ctx.viper.GetString("token"),
	)

	cfg.Region = aws.String(ctx.viper.GetString("region"))
	cfg.MaxRetries = aws.Int(ctx.viper.GetInt("max_retries"))

	ctx.bucket = ctx.viper.GetString("bucket")
	ctx.path = ctx.viper.GetString("path")

	timeout := ctx.viper.GetString("upload_timeout")
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("failed to parse timeout duration %v", err)
	}

	httpClient := &http.Client{Timeout: duration}
	cfg.HTTPClient = httpClient

	sess := session.Must(session.NewSession(cfg))
	ctx.client = s3manager.NewUploader(sess)

	return
}

func (ctx *S3) close() {}

func (ctx *S3) upload(fileKey string) (err error) {
	f, err := os.Open(ctx.archivePath)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", ctx.archivePath, err)
	}

	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get size of file %q, %v", ctx.archivePath, err)
	}

	remotePath := path.Join(ctx.path, fileKey)

	input := &s3manager.UploadInput{
		Bucket: aws.String(ctx.bucket),
		Key:    aws.String(remotePath),
		Body:   f,
	}

	logger.Info(fmt.Sprintf("-> S3 Uploading (%d MiB)...", info.Size()/(1024*1024)))

	start := time.Now()

	result, err := ctx.client.Upload(input, func(uploader *s3manager.Uploader) {
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
	rate := math.Ceil(float64(info.Size()) / (elapsed.Seconds() * 1024 * 1024))

	logger.Info(fmt.Sprintf("Duration %v, rate %.1f MiB/s", durafmt.Parse(elapsed).LimitFirstN(2).String(), rate))

	return nil
}

func (ctx *S3) delete(fileKey string) (err error) {
	remotePath := path.Join(ctx.path, fileKey)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(ctx.bucket),
		Key:    aws.String(remotePath),
	}
	_, err = ctx.client.S3.DeleteObject(input)
	return
}
