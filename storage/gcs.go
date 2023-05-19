package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"golang.org/x/oauth2/google"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// GCS - Google Clound storage
//
// type: gcs
// bucket: gobackup-test
// path: backups
// credentials: { ... }
// credentials_file:
// timeout: 300
type GCS struct {
	Base
	bucket  string
	path    string
	timeout time.Duration
	client  *storage.Client
}

func (s *GCS) open() (err error) {
	// https://cloud.google.com/storage/docs/locations
	s.viper.SetDefault("timeout", "300")

	timeout := s.viper.GetInt("timeout")
	s.timeout = time.Duration(timeout) * time.Second
	s.path = s.viper.GetString("path")
	s.bucket = s.viper.GetString("bucket")
	ctx := context.Background()

	credentials := s.viper.GetString("credentials")
	credentialsFile := s.viper.GetString("credentials_file")

	var opt option.ClientOption
	if len(credentials) != 0 {
		opt = option.WithCredentialsJSON([]byte(credentials))
	} else if len(credentialsFile) != 0 {
		opt = option.WithCredentialsFile(credentialsFile)
	} else {
		// Defaults to search for credentials in several locations: https://pkg.go.dev/golang.org/x/oauth2/google#FindDefaultCredentials
		// of which of interest to us are:
		// 1. A JSON file whose path is specified by the GOOGLE_APPLICATION_CREDENTIALS environment variable, similar to how credentials_file works
		// 4. Fetches credentials from the metadata server which allows us to assign a GCP Service Account to an instance where gobackup runs,
		//    thus avoiding the need to add use a static secret
		creds, err := google.FindDefaultCredentials(ctx, storage.ScopeReadWrite)
		if err != nil {
			return fmt.Errorf("Cannot find default application credentials: %v", err)
		}
		opt = option.WithCredentials(creds)
	}

	s.client, err = storage.NewClient(ctx, opt)
	if err != nil {
		return err
	}

	return
}

func (s *GCS) close() {
	s.client.Close()
}

func (s *GCS) upload(fileKey string) (err error) {
	logger := logger.Tag("GCS")

	var ctx = context.Background()
	var cancel context.CancelFunc

	if s.timeout.Seconds() > 0 {
		logger.Info(fmt.Sprintf("timeout: %s", s.timeout))
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}

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

		// Open file
		f, err := os.Open(sourcePath)
		if err != nil {
			return fmt.Errorf("GCS failed to open file %q, %v", sourcePath, err)
		}
		defer f.Close()

		progress := helper.NewProgressBar(logger, f)
		object := s.client.Bucket(s.bucket).Object(remotePath).If(storage.Conditions{DoesNotExist: true})
		writer := object.NewWriter(ctx)

		if _, err = io.Copy(writer, progress.Reader); err != nil {
			return progress.Errorf("GCS upload error: %v", err)
		}
		if err := writer.Close(); err != nil {
			return progress.Errorf("GCS upload Writer.Close: %v", err)
		}
		progress.Done(remotePath)
	}

	return nil
}

func (s *GCS) delete(fileKey string) (err error) {
	// No need to remove empty directory
	if !strings.HasSuffix(fileKey, "/") {
		remotePath := filepath.Join(s.path, fileKey)
		object := s.client.Bucket(s.bucket).Object(remotePath)
		if err = object.Delete(context.Background()); err != nil {
			return fmt.Errorf("GCS failed to delete file %q, %v", remotePath, err)
		}
	}

	return nil
}

// List all files in the bucket
func (s *GCS) list(parent string) ([]FileItem, error) {
	var files []FileItem
	remotePath := filepath.Join(s.path, parent)

	it := s.client.Bucket(s.bucket).Objects(context.Background(), &storage.Query{Prefix: remotePath})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		file := FileItem{
			Filename:     attrs.Name,
			Size:         attrs.Size,
			LastModified: attrs.Created,
		}

		files = append(files, file)
	}

	return files, nil
}

// Generate a sign URL for download
func (s *GCS) download(fileKey string) (string, error) {
	return s.client.Bucket(s.bucket).SignedURL(fileKey, &storage.SignedURLOptions{
		Expires: time.Now().Add(time.Hour * 1),
	})
}
