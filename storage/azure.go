package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// Azure - Microsoft Azure Blob Storage
//
// type: azure
// # Storage Account
// account: gobackup-test
// # Container name
// container: gobackup
// path: backups
// # Authorization https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication
// tenant_id: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// client_id: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// client_secret: xxxxxxxxxxxxxxxxx
// timeout: 300
type Azure struct {
	Base
	account   string
	container string
	path      string
	timeout   time.Duration
	client    *azblob.Client
}

func (s *Azure) open() error {
	s.viper.SetDefault("timeout", "300")
	s.viper.SetDefault("container", "gobackup")
	s.viper.SetDefault("path", "/")

	timeout := s.viper.GetInt("timeout")
	s.timeout = time.Duration(timeout) * time.Second
	s.container = s.viper.GetString("container")
	s.account = s.viper.GetString("account")
	if len(s.account) == 0 {
		s.account = s.viper.GetString("bucket")
	}
	s.path = s.viper.GetString("path")

	tenantId := s.viper.GetString("tenant_id")
	clientId := s.viper.GetString("client_id")
	clientSecret := s.viper.GetString("client_secret")

	credential, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, nil)
	if err != nil {
		return fmt.Errorf("Invalid credentials with error: %w", err)
	}

	s.client, err = azblob.NewClient(s.getBucketURL(), credential, nil)
	if err != nil {
		return err
	}

	return nil
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
		sourcePath := filepath.Join(filepath.Dir(s.archivePath), key)
		remotePath := filepath.Join(s.path, key)

		// Open file
		f, err := os.Open(sourcePath)
		if err != nil {
			return fmt.Errorf("Azure failed to open file %q, %v", sourcePath, err)
		}
		defer f.Close()

		progress := helper.NewProgressBar(logger, f)
		if _, err = s.client.UploadStream(ctx, s.container, remotePath, progress.Reader, nil); err != nil {
			return progress.Errorf("Azure upload error: %v", err)
		}
		progress.Done(remotePath)

	}

	return nil
}

func (s *Azure) delete(fileKey string) (err error) {
	remotePath := filepath.Join(s.path, fileKey)
	var ctx = context.Background()

	if _, err = s.client.DeleteBlob(ctx, s.container, remotePath, nil); err != nil {
		return fmt.Errorf("Azure failed to delete file %q, %v", remotePath, err)
	}

	return nil
}

// List the objects in the bucket with the prefix = parent
// https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/storage/azblob
func (s *Azure) list(parent string) ([]FileItem, error) {
	remotePath := filepath.Join(s.archivePath, parent)
	var ctx = context.Background()

	var fileItems []FileItem

	// Get a result segment starting with the blob indicated by the current Marker.
	pager := s.client.NewListBlobsFlatPager(s.container, &azblob.ListBlobsFlatOptions{
		Prefix: &remotePath,
	})

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, blob := range resp.Segment.BlobItems {
			fileItems = append(fileItems, FileItem{
				Filename:     *blob.Name,
				LastModified: *blob.Properties.LastModified,
				Size:         *blob.Properties.ContentLength,
			})
		}
	}

	return fileItems, nil
}

// Get a client download URL
func (s *Azure) download(fileKey string) (string, error) {
	containerClient := s.client.ServiceClient().NewContainerClient(s.container)
	blobClient := containerClient.NewBlobClient(fileKey)

	return blobClient.GetSASURL(sas.BlobPermissions{Read: true}, time.Now(), time.Now().Add(time.Hour*1))
}
