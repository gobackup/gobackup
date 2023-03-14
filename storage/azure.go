package storage

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/dustin/go-humanize"
	"github.com/hako/durafmt"

	"github.com/gobackup/gobackup/logger"
)

// Azure - Microsoft Azure Blob Storage
//
// type: azure
// # Storage Account
// account: gobackup-test
// # Container name
// container: gobackup
// # Authorization https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication
// tenant_id: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// client_id: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// client_secret: xxxxxxxxxxxxxxxxx
// timeout: 300
type Azure struct {
	Base
	account   string
	container string
	timeout   time.Duration
	client    *azblob.Client
}

func (s *Azure) open() (err error) {
	s.viper.SetDefault("timeout", "300")
	s.viper.SetDefault("container", "gobackup")

	timeout := s.viper.GetInt("timeout")
	s.timeout = time.Duration(timeout) * time.Second
	s.container = s.viper.GetString("container")
	s.account = s.viper.GetString("account")
	if len(s.account) == 0 {
		s.account = s.viper.GetString("bucket")
	}

	tenantId := s.viper.GetString("tenant_id")
	clientId := s.viper.GetString("client_id")
	clientSecret := s.viper.GetString("client_secret")

	credential, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, nil)
	if err != nil {
		logger.Fatal("Invalid credentials with error: " + err.Error())
	}

	s.client, err = azblob.NewClient(s.getBucketURL(), credential, nil)
	if err != nil {
		return err
	}

	return
}

func (s *Azure) close() {
}

func (s Azure) getBucketURL() string {
	return fmt.Sprintf("https://%s.blob.core.windows.net", s.account)
}

func (s *Azure) upload(fileKey string) (err error) {
	logger := logger.Tag("Azure")

	var ctx = context.Background()
	var cancel context.CancelFunc

	if s.timeout.Seconds() > 0 {
		logger.Info(fmt.Sprintf("timeout: %s", s.timeout))
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}

	// Check to create Azure Storage Container, And ignore error
	_, err = s.client.CreateContainer(ctx, s.container, nil)

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
		filePath := filepath.Join(filepath.Dir(s.archivePath), key)
		// Open file
		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("Azure failed to open file %q, %v", filePath, err)
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return fmt.Errorf("Azure failed to get size of file %q, %v", filePath, err)
		}

		remotePath := key

		logger.Infof("-> Uploading (%s)...", humanize.Bytes(uint64(info.Size())))

		start := time.Now()
		if _, err = s.client.UploadFile(ctx, s.container, remotePath, f, nil); err != nil {
			return fmt.Errorf("Azure upload error: %v", err)
		}

		t := time.Now()
		elapsed := t.Sub(start)

		rate := math.Ceil(float64(info.Size()) / (elapsed.Seconds() * 1024 * 1024))

		logger.Info(fmt.Sprintf("Duration %v, rate %.1f MiB/s", durafmt.Parse(elapsed).LimitFirstN(2).String(), rate))
	}

	return nil
}

func (s *Azure) delete(fileKey string) (err error) {
	var ctx = context.Background()

	// No need to remove empty directory
	if !strings.HasSuffix(fileKey, "/") {
		remotePath := fileKey
		if _, err = s.client.DeleteBlob(ctx, s.container, remotePath, nil); err != nil {
			return fmt.Errorf("Azure failed to delete file %q, %v", remotePath, err)
		}
	}

	return nil
}

func (s *Azure) list(parent string) ([]FileItem, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Azure) download(fileKey string) (string, error) {
	return "", fmt.Errorf("not implemented")
}
