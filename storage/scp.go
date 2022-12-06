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

	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
)

// SCP storage
//
// type: scp
// host: 192.168.1.2
// port: 22
// username: root
// password:
// timeout: 300
// private_key: ~/.ssh/id_rsa
// passpharase:
type SCP struct {
	Base
	path        string
	host        string
	port        string
	privateKey  string
	passpharase string
	username    string
	password    string
	client      *ssh.Client
}

func (s *SCP) open() (err error) {
	logger := logger.Tag("SCP")

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

	var auths []ssh.AuthMethod
	keyCallBack := ssh.InsecureIgnoreHostKey()
	clientConfig := ssh.ClientConfig{
		User:            s.username,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: keyCallBack,
	}

	// PrivateKeyWithPassphrase, PrivateKey
	logger.Debugf("PrivateKey: %s", s.privateKey)
	if len(s.passpharase) != 0 {
		if cc, err := auth.PrivateKeyWithPassphrase(
			s.username,
			[]byte(s.passpharase),
			s.privateKey,
			keyCallBack,
		); err != nil {
			logger.Debugf("PrivateKey with passpharase failed: %v", err)
		} else {
			auths = append(auths, cc.Auth...)
			logger.Debug("Added passpharase private key")
		}
	} else {
		if cc, err := auth.PrivateKey(
			s.username,
			s.privateKey,
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
			s.username,
			keyCallBack,
		); err != nil {
			logger.Debugf("SSH agent failed: %v", err)
		} else {
			auths = append(auths, cc.Auth...)
			logger.Debug("Added SSH agent")
		}
	}

	// PasswordKey
	if len(s.password) != 0 {
		if cc, err := auth.PasswordKey(
			s.username,
			s.password,
			keyCallBack,
		); err != nil {
			logger.Debugf("SSH agent failed: %v", err)
		} else {
			auths = append(auths, cc.Auth...)
			logger.Debug("Added password key")
		}
	}

	logger.Debugf("Auths: %#v", auths)
	clientConfig.Auth = auths
	clientConfig.Timeout = s.viper.GetDuration("timeout") * time.Second

	sshClient, err := ssh.Dial("tcp", s.host+":"+s.port, &clientConfig)
	if err != nil {
		return fmt.Errorf("Failed to ssh %s@%s -p %s: %v", s.username, s.host, s.port, err)
	}

	// mkdir
	if err := sshRun(sshClient, fmt.Sprintf("mkdir -p %s", s.path)); err != nil {
		return err
	}

	s.client = sshClient
	return nil
}

func sshRun(client *ssh.Client, cmd string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("Failed to create session: %v", err)
	}
	defer session.Close()

	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("Failed to run %s: %v", cmd, err)
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
		remoteDir := filepath.Join(s.path, filepath.Base(s.archivePath))
		// mkdir
		if err := sshRun(s.client, fmt.Sprintf("mkdir -p %s", remoteDir)); err != nil {
			return err
		}
	} else {
		// file
		// 2022.12.04.07.09.25.tar.xz
		fileKeys = append(fileKeys, fileKey)
	}

	//defer s.client.Session.Close()
	for _, key := range fileKeys {
		filePath := filepath.Join(filepath.Dir(s.archivePath), key)
		remotePath := filepath.Join(s.path, key)
		if err := upload(s.client, filePath, remotePath); err != nil {
			return err
		}
	}

	logger.Info("Store succeeded")
	return nil
}

func upload(sshClient *ssh.Client, localPath, remotePath string) error {
	logger := logger.Tag("SCP")

	client, err := scp.NewClientBySSH(sshClient)
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

	logger.Info("-> scp to", remotePath)
	if err := client.CopyFromFile(context.Background(), *file, remotePath, "0644"); err != nil {
		return fmt.Errorf("Store %s failed: %v", remotePath, err)
	}
	logger.Infof("Store %s succeeded", remotePath)
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
	if err := sshRun(s.client, fmt.Sprintf("%s %s", rmCmd, remotePath)); err != nil {
		return err
	}

	return
}
