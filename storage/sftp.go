package storage

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// SFTP storage
//
// type: sftp
type SFTP struct {
	Base
	SSH
	path   string
	client *sftp.Client
}

func (s *SFTP) open() error {
	s.viper.SetDefault("port", "22")
	s.viper.SetDefault("timeout", 300)
	s.viper.SetDefault("private_key", "~/.ssh/id_rsa")

	s.host = s.viper.GetString("host")
	s.port = s.viper.GetString("port")
	s.path = s.viper.GetString("path")
	s.username = s.viper.GetString("username")
	s.password = s.viper.GetString("password")
	s.privateKey = helper.ExplandHome(s.viper.GetString("private_key"))
	s.passpharase = s.viper.GetString("passpharase")

	if len(s.host) == 0 {
		return fmt.Errorf("host is required")
	}

	if len(s.username) == 0 {
		user, err := user.Current()
		if err == nil {
			s.username = user.Username
		} else {
			return fmt.Errorf("username is required and it is not able to get current user: %v", err)
		}
	}

	sc := sshConfig{
		username:    s.username,
		password:    s.password,
		privateKey:  s.privateKey,
		passpharase: s.passpharase,
	}

	clientConfig := newSSHClientConfig(sc)
	clientConfig.Timeout = s.viper.GetDuration("timeout") * time.Second

	sshClient, err := ssh.Dial("tcp", s.host+":"+s.port, &clientConfig)
	if err != nil {
		return fmt.Errorf("Failed to ssh %s@%s -p %s: %v", s.username, s.host, s.port, err)
	}

	client, err := sftp.NewClient(sshClient)
	if err != nil {
		return err
	}

	// mkdir
	if err := client.MkdirAll(s.path); err != nil {
		return err
	}

	s.client = client
	return nil
}

func (s *SFTP) close() {
	s.client.Close()
}

func (s *SFTP) upload(fileKey string) error {
	logger := logger.Tag("SFTP")

	var fileKeys []string
	if len(s.fileKeys) != 0 {
		// directory
		// 2022.12.04.07.09.47/2022.12.04.07.09.47.tar.xz-000
		fileKeys = s.fileKeys

		remotePath := filepath.Join(s.path, fileKey)
		remoteDir := filepath.Dir(remotePath)

		// mkdir
		if err := s.client.MkdirAll(remoteDir); err != nil {
			return err
		}
	} else {
		// file
		// 2022.12.04.07.09.25.tar.xz
		fileKeys = append(fileKeys, fileKey)
	}

	//defer s.client.Session.Close()
	for _, key := range fileKeys {
		sourcePath := filepath.Join(filepath.Dir(s.archivePath), key)
		remotePath := filepath.Join(s.path, key)
		if err := s.up(sourcePath, remotePath); err != nil {
			return err
		}
	}

	logger.Info("Store succeeded")
	return nil
}

func (s *SFTP) up(localPath, remotePath string) error {
	logger := logger.Tag("SFTP")

	file, err := os.Open(localPath)
	if err != nil {
		logger.Errorf("Unable to open local file %s: %v", localPath, err)
		return err
	}
	defer file.Close()

	logger.Info("-> upload to", remotePath)
	remoteFile, err := s.client.OpenFile(remotePath, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		logger.Errorf("Unable to open remote file %s: %v", remotePath, err)
		return err
	}
	defer remoteFile.Close()

	if _, err := io.Copy(remoteFile, file); err != nil {
		logger.Errorf("Unable to upload local file %s: %v", localPath, err)
		return err
	}
	logger.Infof("Store %s succeeded", remotePath)

	return nil
}

func (s *SFTP) delete(fileKey string) error {
	logger := logger.Tag("SFTP")

	remotePath := path.Join(s.path, fileKey)
	logger.Info("-> remove", remotePath)
	if err := s.client.Remove(remotePath); err != nil {
		return err
	}

	return nil
}

func (s *SFTP) list(parent string) ([]FileItem, error) {
	remotePath := path.Join(s.path, parent)
	var items []FileItem

	fileInfos, err := s.client.ReadDir(remotePath)
	if err != nil {
		return nil, err
	}
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			items = append(items, FileItem{
				Filename:     fileInfo.Name(),
				Size:         fileInfo.Size(),
				LastModified: fileInfo.ModTime(),
			})
		}
	}

	return items, nil
}

func (s *SFTP) download(fileKey string) (string, error) {
	return "", fmt.Errorf("SFTP not support download")
}
