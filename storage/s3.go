package storage

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
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
}

func (ctx *S3) perform() error {
	ctx.viper.SetDefault("region", "us-east-1")

	cfg := aws.NewConfig()
	endpoint := ctx.viper.GetString("endpoint")
	if len(endpoint) > 0 {
		cfg.Endpoint = aws.String(endpoint)
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
	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(ctx.archivePath)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", ctx.archivePath, err)
	}

	remotePath := path.Join(ctx.path, ctx.fileKey)

	input := &s3manager.UploadInput{
		Bucket: aws.String(ctx.bucket),
		Key:    aws.String(remotePath),
		Body:   f,
	}

	logger.Info("-> S3 Uploading...")
	result, err := uploader.Upload(input)
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	logger.Info("=>", result.Location)
	return nil
}
