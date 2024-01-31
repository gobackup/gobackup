package database

import (
	"fmt"
	"path"
	"strconv"
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
	endpoints             []string
	user                  string
	password              string
	caCert                string
	cert                  string
	key                   string
	insecureSkipTlsVerify string
	args                  string
	_dumpFilePath         string
}

func (db *Etcd) init() (err error) {
	viper := db.viper

	db.endpoints = viper.GetStringSlice("endpoints")
	db.user = viper.GetString("user")
	db.password = viper.GetString("password")
	db.caCert = viper.GetString("cacert")
	db.cert = viper.GetString("cert")
	db.key = viper.GetString("key")
	db.insecureSkipTlsVerify = viper.GetString("insecure-skip-tls-verify")
	db.args = viper.GetString("args")

	if len(db.endpoints) == 0 {
		return fmt.Errorf("etcd endpoint config is required")
	}

	db._dumpFilePath = path.Join(db.dumpPath, strings.Join(db.endpoints, "-"))

	if len(db.caCert) > 0 {
		if !helper.IsExistsPath(db.caCert) {
			return fmt.Errorf("ca-cert: " + db.caCert + " does not exist.")
		}
	}

	if len(db.cert) > 0 {
		if !helper.IsExistsPath(db.cert) {
			return fmt.Errorf("cert: " + db.caCert + " does not exist.")
		}
	}

	if len(db.key) > 0 {
		if !helper.IsExistsPath(db.key) {
			return fmt.Errorf("key: " + db.key + " does not exist.")
		}
	}

	if len(db.insecureSkipTlsVerify) > 0 {
		_, err = strconv.ParseBool(db.insecureSkipTlsVerify)
		if err != nil {
			return fmt.Errorf("insecure-skip-tls-verify should be true or false")
		}
	}

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

	if len(db.user) > 0 {
		etcdctlArgs = append(etcdctlArgs, "--user=\""+db.user+"\"")
	}

	if len(db.password) > 0 {
		etcdctlArgs = append(etcdctlArgs, "--password=\""+db.password+"\"")
	}

	if len(db.caCert) > 0 {
		etcdctlArgs = append(etcdctlArgs, "--cacert=\""+db.caCert+"\"")
	}

	if len(db.cert) > 0 {
		etcdctlArgs = append(etcdctlArgs, "--cert=\""+db.cert+"\"")
	}

	if len(db.key) > 0 {
		etcdctlArgs = append(etcdctlArgs, "--key=\""+db.key+"\"")
	}

	if len(db.insecureSkipTlsVerify) > 0 {
		etcdctlArgs = append(etcdctlArgs, "--insecure-skip-tls-verify="+db.insecureSkipTlsVerify)
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
