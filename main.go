package main

import (
	"log"
	"os"

	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
	"github.com/huacnlee/gobackup/model"
	"github.com/huacnlee/gobackup/scheduler"
	"github.com/spf13/viper"
	"github.com/takama/daemon"

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

func main() {
	app := cli.NewApp()

	var configFlag = &cli.StringFlag{
		Name:        "config",
		Aliases:     []string{"c"},
		Usage:       "Special a config file",
		Destination: &configFile,
	}

	app.Version = version
	app.Name = "gobackup"
	app.Usage = usage

	app.Commands = []*cli.Command{
		{
			Name: "perform",
			Flags: []cli.Flag{
				configFlag,
				&cli.StringFlag{
					Name:        "model",
					Aliases:     []string{"m"},
					Usage:       "Model name that you want perform",
					Destination: &modelName,
				},
			},
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
			Name: "start",
			Flags: []cli.Flag{
				configFlag,
				&cli.BoolFlag{
					Name:        "daemon",
					Aliases:     []string{"d"},
					Usage:       "Run as daemon",
					Destination: &runAsDaemon,
				},
			},
			Action: func(ctx *cli.Context) error {
				config.Init(configFile)

				scheduler.Start()

				if runAsDaemon {
					service, err := daemon.New("gobackup", "GoBackup daemon", daemon.GlobalDaemon)
					if err != nil {
						log.Fatal("Error: ", err)
					}
					service.Start()
				}

				return nil
			},
		},
	}

	app.Run(os.Args)
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
