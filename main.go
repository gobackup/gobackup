package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
	"github.com/huacnlee/gobackup/model"
	"github.com/huacnlee/gobackup/scheduler"
	"github.com/sevlyar/go-daemon"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

const (
	usage = "Backup your databases, files to FTP / SCP / S3 / GCS and other cloud storages."
)

var (
	modelName   = ""
	configFile  = ""
	version     = "master"
	runAsDaemon = false
)

func buildFlags(flags []cli.Flag) []cli.Flag {
	return append(flags, &cli.StringFlag{
		Name:        "config",
		Aliases:     []string{"c"},
		Usage:       "Special a config file",
		Destination: &configFile,
	})
}

func main() {
	app := cli.NewApp()

	app.Version = version
	app.Name = "gobackup"
	app.Usage = usage

	app.Commands = []*cli.Command{
		{
			Name: "perform",
			Flags: buildFlags([]cli.Flag{
				&cli.StringFlag{
					Name:        "model",
					Aliases:     []string{"m"},
					Usage:       "Model name that you want perform",
					Destination: &modelName,
				},
			}),
			Action: func(ctx *cli.Context) error {
				config.Init(configFile)

				if len(modelName) == 0 {
					performAll()
				} else {
					performOne(modelName)
				}

				return nil
			},
		},
		{
			Name:  "start",
			Usage: "Start as daemon",
			Flags: buildFlags([]cli.Flag{}),
			Action: func(ctx *cli.Context) error {
				fmt.Println("GoBackup starting...")

				args := []string{"gobackup", "run"}
				if len(configFile) != 0 {
					args = append(args, "--config", configFile)
				}

				dm := &daemon.Context{
					PidFileName: filepath.Join(config.GoBackupDir, "gobackup.pid"),
					LogFileName: filepath.Join(config.GoBackupDir, "gobackup.log"),
					WorkDir:     "./",
					Args:        args,
				}
				d, err := dm.Reborn()
				if err != nil {
					log.Fatal("Unable to run: ", err)
				}
				if d != nil {
					return nil
				}
				defer dm.Release()

				initApplication()
				scheduler.Start()

				return nil
			},
		},
		{
			Name:  "run",
			Usage: "Run GoBackup",
			Flags: buildFlags([]cli.Flag{}),
			Action: func(ctx *cli.Context) error {
				initApplication()
				scheduler.Start()

				err := daemon.ServeSignals()
				if err != nil {
					log.Printf("Error: %s", err.Error())
				}

				log.Println("daemon terminated")

				return nil
			},
		},
	}

	app.Run(os.Args)
}

func initApplication() {
	config.Init(configFile)

}

func performAll() {
	for _, modelConfig := range config.Models {
		m := model.Model{
			Config: modelConfig,
		}
		m.Perform()
	}
}

func performOne(modelName string) {
	modelConfig := config.GetModelByName(modelName)
	if modelConfig == nil {
		return
	}
	logger.Fatalf("Model %s not found in %s", modelName, viper.ConfigFileUsed())

	m := model.Model{
		Config: *modelConfig,
	}
	m.Perform()
}
