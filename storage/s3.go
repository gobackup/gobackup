package storage

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
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
// storage_class:
// timeout: 300
// force_path_style:
type S3 struct {
	Base
	Service      string
	bucket       string
	path         string
	client       *s3manager.Uploader
	storageClass string
	awsCfg       *aws.Config
}

func (s S3) providerName() string {
	switch s.Service {
	case "s3":
		return "AWS S3"
	case "b2":
		return "Backblaze B2"
	case "us3":
		return "UCloud US3"
	case "cos":
		return "QCloud COS"
	case "kodo":
		return "Qiniu Kodo"
	case "r2":
		return "Cloudflare R2"
	case "spaces":
		return "DigitalOcean Spaces"
	case "bos":
		return "Baidu BOS"
	case "oss":
		return "Aliyun OSS"
	case "minio":
		return "MinIO"
	case "obs":
		return "Huawei OBS"
	case "tos":
		return "Volcengine TOS"
	case "upyun":
		return "UpYun"
	}

	return "AWS S3"
}

func (s S3) defaultRegion() string {
	switch s.Service {
	case "s3":
		return "us-east-1"
	case "b2":
		return "us-east-001"
	case "us3":
		return "s3-cn-bj"
	case "cos":
		return "ap-nanjing"
	case "kodo":
		return "cn-east-1"
	case "r2":
		return "us-east-1"
	case "spaces":
		return "nyc1"
	case "bos":
		return "bj"
	case "oss":
		return "cn-hangzhou"
	case "minio":
		return "us-east-1"
	case "obs":
		return "cn-north-1"
	case "tos":
		// https://www.volcengine.com/docs/6349/107356
		return "cn-beijing"
	case "upyun":
		// UpYun does not support region
		return "none"
	}

	return "us-east-1"
}

func (s S3) defaultEndpoint() *string {
	switch s.Service {
	case "b2":
		return aws.String(fmt.Sprintf("%s.backblazeb2.com", s.viper.GetString("region")))
	case "us3":
		return aws.String(fmt.Sprintf("%s.ufileos.com", s.viper.GetString("region")))
	case "cos":
		return aws.String(fmt.Sprintf("cos.%s.myqcloud.com", s.viper.GetString("region")))
	case "kodo":
		return aws.String(fmt.Sprintf("s3-%s.qiniucs.com", s.viper.GetString("region")))
	case "r2":
		return aws.String(fmt.Sprintf("%s.r2.cloudflarestorage.com", s.viper.GetString("account_id")))
	case "spaces":
		return aws.String(fmt.Sprintf("%s.digitaloceanspaces.com", s.viper.GetString("region")))
	case "bos":
		return aws.String(fmt.Sprintf("s3.%s.bcebos.com", s.viper.GetString("region")))
	case "oss":
		return aws.String(fmt.Sprintf("oss-%s.aliyuncs.com", s.viper.GetString("region")))
	case "obs":
		return aws.String(fmt.Sprintf("obs.%s.myhuaweicloud.com", s.viper.GetString("region")))
	case "tos":
		return aws.String(fmt.Sprintf("tos-s3-%s.volces.com", s.viper.GetString("region")))
	case "upyun":
		return aws.String("s3.api.upyun.com")
	}

	return aws.String("")
}

func (s *S3) defaultStorageClass() string {
	switch s.Service {
	case "s3":
		return "STANDARD_IA"
	case "b2":
		return "STANDARD"
	case "us3":
		return "ARCHIVE"
	case "cos":
		return "STANDARD_IA"
	case "kodo":
		return "LINE"
	case "r2":
		// https://developers.cloudflare.com/r2/api/s3/api/
		return ""
	case "spaces":
		// Allowed for compatibility purposes. Spaces only accepts the default value, STANDARD,
		// and will reject other, unsupported storage class values.
		// https://docs.digitalocean.com/reference/api/spaces-api/#upload-an-object-put
		return "STANDARD"
	case "bos":
		return "STANDARD_IA"
	case "oss":
		// https://help.aliyun.com/document_detail/389025.html
		// By test, Aliyun OSS only support "Standard" via S3 SDK, even we set "STANDARD_IA" or "ARCHIVE"
		return "STANDARD_IA"
	case "minio":
		return ""
	case "obs":
		// https://support.huaweicloud.com/api-obs/obs_04_0044.html#obs_04_0044__table63485364
		// STANDARD, STANDARD_IA, GLACIER, DEEP_ARCHIVE
		return "STANDARD_IA"
	case "tos":
		// https://www.volcengine.com/docs/6349/147050
		// STANDARD, STANDARD_IA, GLACIER_IR
		return "STANDARD_IA"
	case "upyun":
		// https://help.upyun.com/knowledge-base/s3-api
		// UpYun API only support STANDARD, so keep this in empty.
		// And they S3 API only support upload to STANDARD (普通) bucket, it will return 403 when the bucket type is STANDARD_IA (低频).
		return ""
	}

	return ""
}

func (s *S3) forcePathStyle() bool {
	switch s.Service {
	case "tos", "oss":
		return false
	default:
		return true
	}
}

func (s *S3) init() {
	if len(s.Service) == 0 {
		s.Service = "s3"
	}

	s.viper.SetDefault("region", s.defaultRegion())
	s.viper.SetDefault("endpoint", s.defaultEndpoint())
	s.viper.SetDefault("max_retries", 3)
	s.viper.SetDefault("timeout", "300")
	s.viper.SetDefault("storage_class", s.defaultStorageClass())
}

func (s *S3) open() (err error) {
	s.init()

	logger := logger.Tag(s.providerName())

	cfg := aws.NewConfig()
	endpoint := s.viper.GetString("endpoint")

	if len(endpoint) > 0 {
		cfg.Endpoint = aws.String(endpoint)
	}

	cfg.S3ForcePathStyle = aws.Bool(s.forcePathStyle())
	if s.viper.IsSet("force_path_style") {
		cfg.S3ForcePathStyle = aws.Bool(s.viper.GetBool("force_path_style"))
	}

	accessKeyId := s.viper.GetString("access_key_id")
	secretAccessKey := s.viper.GetString("secret_access_key")
	if len(secretAccessKey) == 0 {
		secretAccessKey = s.viper.GetString("access_key_secret")
	}

	if len(accessKeyId) == 0 || len(secretAccessKey) == 0 {
		logger.Warn("`access_key_id` or `secret_access_key` is empty.")
	}

	cfg.Credentials = credentials.NewStaticCredentials(
		accessKeyId,
		secretAccessKey,
		s.viper.GetString("token"),
	)

	cfg.Region = aws.String(s.viper.GetString("region"))
	cfg.MaxRetries = aws.Int(s.viper.GetInt("max_retries"))

	s.bucket = s.viper.GetString("bucket")
	s.path = s.viper.GetString("path")
	s.storageClass = s.viper.GetString("storage_class")

	timeout := s.viper.GetInt("timeout")
	uploadTimeoutDuration := time.Duration(timeout) * time.Second

	httpClient := &http.Client{Timeout: uploadTimeoutDuration}
	cfg.HTTPClient = httpClient
	s.awsCfg = cfg

	sess := session.Must(session.NewSession(s.awsCfg))
	s.client = s3manager.NewUploader(sess)

	return
}

func (s *S3) close() {
}

func (s *S3) upload(fileKey string) (err error) {
	logger := logger.Tag(s.providerName())

	var fileKeys []string
	if len(s.fileKeys) != 0 {
		// directory
		// 2022.12.04.07.09.47/2022.12.04.07.09.47.tar.xz-000
		fileKeys = s.fileKeys
	} else {
		// file
		// 2022.12.04.07.09.25.tar.xz
		fileKeys = append(fileKeys, fileKey)
	}

	for _, key := range fileKeys {
		sourcePath := filepath.Join(filepath.Dir(s.archivePath), key)
		remotePath := filepath.Join(s.path, key)

		f, err := os.Open(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to open file %q, %v", sourcePath, err)
		}
		defer f.Close()

		progress := helper.NewProgressBar(logger, f)

		input := &s3manager.UploadInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(remotePath),
			Body:   progress.Reader,
		}

		// Only present storage_class when it is set.
		// Some storage backend may not support storage_class.
		// https://github.com/gobackup/gobackup/issues/183
		if len(s.storageClass) > 0 {
			input.StorageClass = aws.String(s.storageClass)
		}

		result, err := s.client.Upload(input, func(uploader *s3manager.Uploader) {
			// set the part size as low as possible to avoid timeouts and aborts
			// also set concurrency to 1 for the same reason
			var partSize int64 = 64 * 1024 * 1024 // 64MiB
			maxParts := progress.FileLength / partSize

			// 10000 parts is the limit for AWS S3. If the resulting number of parts would exceed that limit, increase the
			// part size as much as needed but as little possible
			if maxParts > 10000 {
				partSize = int64(math.Ceil(float64(progress.FileLength) / 10000))
			}

			uploader.Concurrency = 1
			uploader.LeavePartsOnError = false
			uploader.PartSize = partSize
		})

		if err != nil {
			return progress.Errorf("%v", err)
		}

		progress.Done(result.Location)

		if s.Service == "s3" {
			logger.Info("=>", fmt.Sprintf("s3://%s/%s", s.bucket, remotePath))
		}
	}

	return nil
}

func (s *S3) delete(fileKey string) (err error) {
	remotePath := filepath.Join(s.path, fileKey)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(remotePath),
	}
	_, err = s.client.S3.DeleteObject(input)
	return
}

// List the objects in the bucket with the prefix = parent
func (s *S3) list(parent string) ([]FileItem, error) {
	remotePath := filepath.Join(s.path, parent)
	continueToken := ""
	var items []FileItem

	for {
		input := &s3.ListObjectsV2Input{
			Bucket: aws.String(s.bucket),
			Prefix: aws.String(remotePath),
		}

		// Only present ContinuationToken when it is set.
		// Some S3 compatible storage like MinIO will raise error when ContinuationToken is empty.
		// https://github.com/gobackup/gobackup/issues/179
		if len(continueToken) > 0 {
			input.ContinuationToken = aws.String(continueToken)
		}

		result, err := s.client.S3.ListObjectsV2(input)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects, %v", err)
		}

		for _, object := range result.Contents {
			items = append(items, FileItem{
				Filename:     *object.Key,
				Size:         *object.Size,
				LastModified: *object.LastModified,
			})
		}

		if *result.IsTruncated {
			continueToken = *result.NextContinuationToken
		} else {
			break
		}
	}

	return items, nil
}

// Get the object download URL by fileKey (include remote_path)
func (s *S3) download(fileKey string) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileKey),
	}

	req, _ := s.client.S3.GetObjectRequest(input)
	url, err := req.Presign(1 * time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to sign request, %v", err)
	}

	return url, nil
}
