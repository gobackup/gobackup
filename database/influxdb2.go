package database

import (
	"fmt"

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
	viper.SetDefault("skip_verify", false)
	viper.SetDefault("http_debug", false)

	db.host = viper.GetString("host")
	db.token = viper.GetString("token")
	db.bucket = viper.GetString("bucket")
	db.bucketId = viper.GetString("bucket_id")
	db.org = viper.GetString("org")
	db.orgId = viper.GetString("org_id")
	db.skipVerify = viper.GetBool("skip_verify")
	db.httpDebug = viper.GetBool("http_debug")

	if db.host == "" {
		return fmt.Errorf("no host specified in influxdb2 configuration '%s'", db.name)
	}
	if db.token == "" {
		return fmt.Errorf("no token specified in influxdb2 configuration '%s'", db.name)
	}

	return nil
}

func (db *InfluxDB2) influxCliArguments() []string {
	args := make([]string, 0, 15)
	args = append(args, "backup")
	args = append(args, "--host="+db.host+"")
	args = append(args, "--token="+db.token+"")
	if db.bucket != "" {
		args = append(args, "--bucket="+db.bucket+"")
	}
	if db.bucketId != "" {
		args = append(args, "--bucket-id="+db.bucketId+"")
	}
	if db.org != "" {
		args = append(args, "--org="+db.org+"")
	}
	if db.orgId != "" {
		args = append(args, "--org-id="+db.orgId+"")
	}
	if db.skipVerify {
		args = append(args, "--skip-verify")
	}
	if db.httpDebug {
		args = append(args, "--http-debug")
	}
	args = append(args, db.dumpPath)
	return args
}

func (db *InfluxDB2) perform() error {
	logger := logger.Tag("InfluxDB2")

	args := db.influxCliArguments()
	out, err := helper.Exec("influx", args...)
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	logger.Info(out)
	logger.Info("dump path:", db.dumpPath)
	return nil
}
