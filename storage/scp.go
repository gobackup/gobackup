package storage

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"golang.org/x/crypto/ssh"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// SSH
//
// host:
// port: 22
// username:
// password:
// timeout: 300
// private_key: ~/.ssh/id_rsa
// passpharase:
type SSH struct {
	host        string
	port        string
	privateKey  string
	passpharase string
	username    string
	password    string
}

// SCP storage
//
// type: scp
type SCP struct {
	Base
	SSH
	path   string
	client *ssh.Client
}

func (s *SCP) open() (err error) {
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
		return fmt.Errorf("failed to ssh %s@%s -p %s: %v", s.username, s.host, s.port, err)
	}

	s.client = sshClient

	// mkdir
	if err := s.run(fmt.Sprintf("mkdir -p %s", s.path)); err != nil {
		return err
	}

	return nil
}

func (s *SCP) run(cmd string) error {
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to run %s: %v", cmd, err)
	}

	return nil
}

func (s *SCP) close() {
	s.client.Close()
}

func (s *SCP) upload(fileKey string) error {
	logger := logger.Tag("SCP")

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

		// mkdir
		if err := s.run(fmt.Sprintf("mkdir -p %s", filepath.Dir(remotePath))); err != nil {
			return err
		}

		// upload file
		if err := s.up(sourcePath, remotePath); err != nil {
			return err
		}
	}

	logger.Info("Store succeeded")
	return nil
}

func (s *SCP) up(localPath, remotePath string) error {
	logger := logger.Tag("SCP")

	client, err := scp.NewClientBySSH(s.client)
	if err != nil {
		return err
	}
	if err := client.Connect(); err != nil {
		return err
	}
	defer client.Close()

	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	progress := helper.NewProgressBar(logger, file)
	if err := client.CopyFile(context.Background(), progress.Reader, remotePath, "0644"); err != nil {
		return progress.Errorf("store %s failed: %v", remotePath, err)
	}
	progress.Done(remotePath)

	return nil
}

func (s *SCP) delete(fileKey string) (err error) {
	logger := logger.Tag("SCP")

	remotePath := path.Join(s.path, fileKey)
	logger.Info("-> remove", remotePath)
	rmCmd := "rm"
	if strings.HasSuffix(fileKey, "/") {
		rmCmd = "rmdir"
	}
	if err := s.run(fmt.Sprintf("%s %s", rmCmd, remotePath)); err != nil {
		return err
	}

	return
}

type sshConfig struct {
	username    string
	password    string
	privateKey  string
	passpharase string
}

func newSSHClientConfig(c sshConfig) ssh.ClientConfig {
	logger := logger.Tag("SSH")

	var auths []ssh.AuthMethod
	keyCallBack := ssh.InsecureIgnoreHostKey()

	// PrivateKeyWithPassphrase, PrivateKey
	logger.Debugf("PrivateKey: %s", c.privateKey)
	if len(c.passpharase) != 0 {
		if cc, err := auth.PrivateKeyWithPassphrase(
			c.username,
			[]byte(c.passpharase),
			c.privateKey,
			keyCallBack,
		); err != nil {
			logger.Debugf("PrivateKey with passpharase failed: %v", err)
		} else {
			auths = append(auths, cc.Auth...)
			logger.Debug("Added passpharase private key")
		}
	} else {
		if cc, err := auth.PrivateKey(
			c.username,
			c.privateKey,
			keyCallBack,
		); err != nil {
			logger.Debugf("PrivateKey failed: %v", err)
		} else {
			auths = append(auths, cc.Auth...)
			logger.Debug("Added private key")
		}
	}

	// private key has higher priority than SSH agent here since crypto/ssh will only try the first instance of a particular RFC 4252 method.
	// https://pkg.go.dev/golang.org/x/crypto/ssh#ClientConfig
	if len(auths) == 0 {
		// SshAgent
		if cc, err := auth.SshAgent(
			c.username,
			keyCallBack,
		); err != nil {
			logger.Debugf("SSH agent failed: %v", err)
		} else {
			auths = append(auths, cc.Auth...)
			logger.Debug("Added SSH agent")
		}
	}

	// PasswordKey
	if len(c.password) != 0 {
		if cc, err := auth.PasswordKey(
			c.username,
			c.password,
			keyCallBack,
		); err != nil {
			logger.Debugf("SSH agent failed: %v", err)
		} else {
			auths = append(auths, cc.Auth...)
			logger.Debug("Added password key")
		}
	}

	logger.Debugf("Auths: %#v", auths)

	return ssh.ClientConfig{
		User:            c.username,
		Auth:            auths,
		HostKeyCallback: keyCallBack,
	}
}

func (s *SCP) list(parent string) ([]FileItem, error) {
	return nil, fmt.Errorf("SCP not support list")
}

func (s *SCP) download(fileKey string) (string, error) {
	return "", fmt.Errorf("SCP not support download")
}
