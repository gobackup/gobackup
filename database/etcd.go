package database

import (
	"fmt"
	"path"
	"strings"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// etcd database
//
// ref:
// https://etcd.io/docs/v3.4/dev-guide/interacting_v3/
//
// # keys
//
//   - type: etcd
//   - endpoints: [localhost]
//   - port: 2379
//   - user:
//   - password:
//   - cacert:
//   - cert:
//   - key:
//   - insecure-skip-tls-verify: false
//   - args:
type Etcd struct {
	Base
	endpoints     []string
	args          string
	_dumpFilePath string
}

func (db *Etcd) init() (err error) {
	viper := db.viper

	db.endpoints = viper.GetStringSlice("endpoints")
	db.args = viper.GetString("args")

	if len(db.endpoints) == 0 {
		return fmt.Errorf("etcd endpoint config is required")
	}

	db._dumpFilePath = path.Join(db.dumpPath, strings.Join(db.endpoints, "-"))

	return nil
}

func (db *Etcd) build() string {
	// etcdctl command
	var etcdctlArgs []string

	etcdctlArgs = append(etcdctlArgs, "snapshot save")
	etcdctlArgs = append(etcdctlArgs, db._dumpFilePath)

	if len(db.endpoints) > 0 {
		etcdctlArgs = append(etcdctlArgs, "--endpoints="+strings.Join(db.endpoints, ","))
	}

	if len(db.args) > 0 {
		etcdctlArgs = append(etcdctlArgs, db.args)
	}

	return "etcdctl " + strings.Join(etcdctlArgs, " ")
}

func (db *Etcd) perform() error {
	logger := logger.Tag("etcd")

	logger.Info("-> Getting snapshot from etcd...")

	_, err := helper.Exec(db.build())
	if err != nil {
		return err
	}
	logger.Info("snapshot path: ", db._dumpFilePath)
	return nil
}
