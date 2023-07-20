package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/logger"
	"github.com/spf13/viper"
)

// Base storage
// When `archivePath` is a directory, `fileKeys` stores files in the `archivePath` with directory prefix
type Base struct {
	model       config.ModelConfig
	archivePath string
	fileKeys    []string
	viper       *viper.Viper
	keep        int
	cycler      *Cycler
}

type FileItem struct {
	Filename     string    `json:"filename,omitempty"`
	Size         int64     `json:"size,omitempty"`
	LastModified time.Time `json:"last_modified,omitempty"`
}

// Storage interface
type Storage interface {
	open() error
	close()
	upload(fileKey string) error
	delete(fileKey string) error
	list(parent string) ([]FileItem, error)
	download(fileKey string) (string, error)
}

func newBase(model config.ModelConfig, archivePath string, storageConfig config.SubConfig) (base Base, err error) {
	// Backward compatible with `store_with` config
	var cyclerName string
	if storageConfig.Name == "" {
		cyclerName = model.Name
	} else {
		cyclerName = fmt.Sprintf("%s_%s", model.Name, storageConfig.Name)
	}

	var keys []string
	if fi, err := os.Stat(archivePath); err == nil && fi.IsDir() {
		// NOTE: ignore err is not nil scenario here to pass test and should be fine
		// 2022.12.04.07.09.47
		entries, err := os.ReadDir(archivePath)
		if err != nil {
			return base, err
		}
		for _, e := range entries {
			// Assume all entries are file
			// 2022.12.04.07.09.47/2022.12.04.07.09.47.tar.xz-000
			if !e.IsDir() {
				keys = append(keys, filepath.Join(filepath.Base(archivePath), e.Name()))
			}
		}
	}

	base = Base{
		model:       model,
		archivePath: archivePath,
		fileKeys:    keys,
		viper:       storageConfig.Viper,
		cycler:      &Cycler{name: cyclerName},
	}

	if base.viper != nil {
		base.keep = base.viper.GetInt("keep")
	}

	return
}

func new(model config.ModelConfig, archivePath string, storageConfig config.SubConfig) (Base, Storage) {
	base, err := newBase(model, archivePath, storageConfig)
	if err != nil {
		panic(err)
	}

	var s Storage
	switch storageConfig.Type {
	case "local":
		s = &Local{Base: base}
	case "webdav":
		s = &WebDAV{Base: base}
	case "ftp":
		s = &FTP{Base: base}
	case "scp":
		s = &SCP{Base: base}
	case "sftp":
		s = &SFTP{Base: base}
	case "oss":
		s = &S3{Base: base, Service: "oss"}
	case "gcs":
		s = &GCS{Base: base}
	case "s3":
		s = &S3{Base: base, Service: "s3"}
	case "minio":
		s = &S3{Base: base, Service: "minio"}
	case "b2":
		s = &S3{Base: base, Service: "b2"}
	case "us3":
		s = &S3{Base: base, Service: "us3"}
	case "cos":
		s = &S3{Base: base, Service: "cos"}
	case "kodo":
		s = &S3{Base: base, Service: "kodo"}
	case "r2":
		s = &S3{Base: base, Service: "r2"}
	case "spaces":
		s = &S3{Base: base, Service: "spaces"}
	case "bos":
		s = &S3{Base: base, Service: "bos"}
	case "obs":
		s = &S3{Base: base, Service: "obs"}
	case "tos":
		s = &S3{Base: base, Service: "tos"}
	case "upyun":
		s = &S3{Base: base, Service: "upyun"}
	case "azure":
		s = &Azure{Base: base}
	default:
		logger.Errorf("[%s] storage type has not implement.", storageConfig.Type)
	}

	return base, s
}

// run storage
func runModel(model config.ModelConfig, archivePath string, storageConfig config.SubConfig) (err error) {
	logger := logger.Tag("Storage")

	newFileKey := filepath.Base(archivePath)
	base, s := new(model, archivePath, storageConfig)

	logger.Info("=> Storage | " + storageConfig.Type)
	err = s.open()
	if err != nil {
		return err
	}
	defer s.close()

	err = s.upload(newFileKey)
	if err != nil {
		return err
	}

	base.cycler.run(newFileKey, base.fileKeys, base.keep, s.delete)
	return nil
}

// Run storage
func Run(model config.ModelConfig, archivePath string) (err error) {
	var errors []error

	n := len(model.Storages)
	for _, storageConfig := range model.Storages {
		err := runModel(model, archivePath, storageConfig)
		if err != nil {
			if n == 1 {
				return err
			} else {
				errors = append(errors, err)
				continue
			}
		}
	}

	if len(errors) != 0 {
		return fmt.Errorf("Storage errors: %v", errors)
	}

	return nil
}

// List return file list of storage
func List(model config.ModelConfig, parent string) (items []FileItem, err error) {
	if storageConfig, ok := model.Storages[model.DefaultStorage]; ok {
		_, s := new(model, "", storageConfig)
		err = s.open()
		if err != nil {
			return nil, err
		}
		defer s.close()

		if parent == "" {
			parent = "/"
		}

		items, err := s.list(parent)
		if err != nil {
			return []FileItem{}, err
		}

		// Sort items by LastModified, Filename in descending
		sort.Slice(items, func(i, j int) bool {
			if items[i].LastModified == items[j].LastModified {
				return items[i].Filename > items[j].Filename
			}
			return items[i].LastModified.After(items[j].LastModified)
		})

		return items, nil
	}

	return []FileItem{}, fmt.Errorf("Storage %s not found", model.DefaultStorage)
}

func Download(model config.ModelConfig, fileKey string) (string, error) {
	if storageConfig, ok := model.Storages[model.DefaultStorage]; ok {
		_, s := new(model, "", storageConfig)
		err := s.open()
		if err != nil {
			return "", err
		}
		defer s.close()

		return s.download(fileKey)
	}

	return "", fmt.Errorf("Storage %s not found", model.DefaultStorage)
}
