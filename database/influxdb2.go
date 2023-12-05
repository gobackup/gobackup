package database

import (
	"fmt"
	"strings"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// InfluxDB v2 database through `influx` cli
// See https://docs.influxdata.com/influxdb/v2/reference/cli/influx/backup/
type InfluxDB2 struct {
	Base
	host       string
	token      string
	bucket     string
	bucketId   string
	org        string
	orgId      string
	skipVerify bool
	httpDebug  bool
}

func (db *InfluxDB2) init() (err error) {
	viper := db.viper
	viper.SetDefault("skipVerify", false)
	viper.SetDefault("httpDebug", false)

	db.host = viper.GetString("host")
	db.token = viper.GetString("token")
	db.bucket = viper.GetString("bucket")
	db.bucketId = viper.GetString("bucketId")
	db.org = viper.GetString("org")
	db.orgId = viper.GetString("orgId")
	db.skipVerify = viper.GetBool("skipVerify")
	db.httpDebug = viper.GetBool("httpDebug")

	if db.host == "" {
		return fmt.Errorf("no host specified in influxdb2 configuration '%s'", db.name)
	}
	if db.token == "" {
		return fmt.Errorf("no token specified in influxdb2 configuration '%s'", db.name)
	}

	return nil
}

func (db *InfluxDB2) build() string {
	opts := make([]string, 0, 15)
	opts = append(opts, "influx backup")
	opts = append(opts, "--host="+db.host+"")
	opts = append(opts, "--token="+db.token+"")
	if db.bucket != "" {
		opts = append(opts, "--bucket="+db.bucket+"")
	}
	if db.bucketId != "" {
		opts = append(opts, "--bucket-id="+db.bucketId+"")
	}
	if db.org != "" {
		opts = append(opts, "--org="+db.org+"")
	}
	if db.orgId != "" {
		opts = append(opts, "--org-id="+db.orgId+"")
	}
	if db.skipVerify {
		opts = append(opts, "--skip-verify")
	}
	if db.httpDebug {
		opts = append(opts, "--http-debug")
	}
	opts = append(opts, db.dumpPath)
	return strings.Join(opts, " ")
}

func (db *InfluxDB2) perform() error {
	logger := logger.Tag("InfluxDB2")

	command := db.build()
	out, err := helper.Exec(command)
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	logger.Info(out)
	logger.Info("dump path:", db.dumpPath)
	return nil
}
