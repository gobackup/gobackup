package storage

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/huacnlee/gobackup/logger"
	"os"
	"path"
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

	remotePath := path.Join(ctx.path, fileKey)

	input := &s3manager.UploadInput{
		Bucket: aws.String(ctx.bucket),
		Key:    aws.String(remotePath),
		Body:   f,
	}

	logger.Info("-> S3 Uploading...")
	result, err := ctx.client.Upload(input)
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	logger.Info("=>", result.Location)
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
