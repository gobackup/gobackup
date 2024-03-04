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
//   - endpoint: localhost:2379
//     # depricated
//   - endpoints: [localhost:2379, localhost:22379, localhost:32379]
//   - args:
type Etcd struct {
	Base
	endpoint      string
	endpoints     []string
	args          string
	_dumpFilePath string
}

func (db *Etcd) init() (err error) {
	viper := db.viper

	db.endpoint = viper.GetString("endpoint")
	db.endpoints = viper.GetStringSlice("endpoints")
	db.args = viper.GetString("args")

	if len(db.endpoint) == 0 && len(db.endpoints) == 0 {
		return fmt.Errorf("etcd endpoint config is required")
	}

	if len(db.endpoint) > 0 && len(db.endpoints) > 0 {
		return fmt.Errorf("etcd `endpoint` and `endpoints` config are mutually exclusive")
	}

	if len(db.endpoint) == 0 && len(db.endpoints) > 0 {
		logger.Warn("DEPRECATED: `endpoints` is deprecated, use `endpoint` instead.")
		logger.Warn("The first element of endpoints will be used.")
		db.endpoint = db.endpoints[0]
	}

	db._dumpFilePath = path.Join(db.dumpPath + "-" + db.endpoint)

	return nil
}

func (db *Etcd) build() string {
	// etcdctl command
	var etcdctlArgs []string

	etcdctlArgs = append(etcdctlArgs, "snapshot save")
	etcdctlArgs = append(etcdctlArgs, db._dumpFilePath)

	if len(db.endpoint) > 0 {
		etcdctlArgs = append(etcdctlArgs, "--endpoints "+db.endpoint)
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
