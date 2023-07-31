package storage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// SFTP storage
//
// type: sftp
type Supabase struct {
	Base
	projectId string
	api_key   string
	path      string
	bucket    string

	client *http.Client
}

func (s *Supabase) open() error {

	s.api_key = s.viper.GetString("api_key")
	s.projectId = s.viper.GetString("project_id")
	s.path = s.viper.GetString("path")
	s.bucket = s.viper.GetString("bucket")

	if len(s.api_key) == 0 {
		return fmt.Errorf("api_key is required")
	}

	if len(s.projectId) == 0 {
		return fmt.Errorf("project_id is required")
	}

	if len(s.bucket) == 0 {
		return fmt.Errorf("bucket is required")
	}

	timeout := s.viper.GetInt("timeout")
	uploadTimeoutDuration := time.Duration(timeout) * time.Second

	s.client = &http.Client{
		Timeout: uploadTimeoutDuration,
	}

	return nil
}

func (s *Supabase) close() {
}

func (s *Supabase) upload(fileKey string) (err error) {
	logger := logger.Tag("Supabase")

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

		reqURL := fmt.Sprintf("https://%s.supabase.co/storage/v1/object/%s/%s", s.projectId, s.bucket, remotePath)

		client, err := http.NewRequest("POST", reqURL, f)
		if err != nil {
			return fmt.Errorf("failed to create request, %v", err)
		}

		client.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.api_key))
		client.Header.Set("Content-Type", "application/octet-stream")

		resp, err := s.client.Do(client)
		if err != nil {
			return fmt.Errorf("failed to send request, %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("failed to upload file %q, %v", sourcePath, err)
		}

		fileUrl := fmt.Sprintf("https://%s.supabase.co/storage/v1/object/%s/%s", s.projectId, s.bucket, remotePath)

		progress.Done(fileUrl)
	}

	return nil
}
