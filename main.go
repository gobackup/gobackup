package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/sevlyar/go-daemon"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"

	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/logger"
	"github.com/gobackup/gobackup/model"
	"github.com/gobackup/gobackup/scheduler"
	"github.com/gobackup/gobackup/web"
)

const (
	usage = "Backup your databases, files to FTP / SCP / S3 / GCS and other cloud storages."
)

var (
	configFile string
	version    = "master"
	signal     = flag.String("s", "", `Send signal to the daemon:
  quit — graceful shutdown
  stop — fast shutdown
  reload — reloading the configuration file`)
)

func buildFlags(flags []cli.Flag) []cli.Flag {
	return append(flags, &cli.StringFlag{
		Name:        "config",
		Aliases:     []string{"c"},
		Usage:       "Special a config file",
		Destination: &configFile,
	})
}

func termHandler(sig os.Signal) error {
	logger.Info("Received QUIT signal, exiting...")
	scheduler.Stop()
	os.Exit(0)
	return nil
}

func reloadHandler(sig os.Signal) error {
	logger.Info("Reloading config...")
	err := config.Init(configFile)
	if err != nil {
		logger.Error(err)
	}

	return nil
}

func main() {
	app := cli.NewApp()

	app.Version = version
	app.Name = "gobackup"
	app.Usage = usage

	daemon.AddCommand(daemon.StringFlag(signal, "quit"), syscall.SIGQUIT, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "reload"), syscall.SIGHUP, reloadHandler)

	app.Commands = []*cli.Command{
		{
			Name: "perform",
			Flags: buildFlags([]cli.Flag{
				&cli.StringSliceFlag{
					Name:    "model",
					Aliases: []string{"m"},
					Usage:   "Model name that you want perform",
				},
			}),
			Action: func(ctx *cli.Context) error {
				var modelNames []string
				err := initApplication()
				if err != nil {
					return err
				}
				modelNames = append(ctx.StringSlice("model"), ctx.Args().Slice()...)
				perform(modelNames)

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
					PidFileName: config.PidFilePath,
					PidFilePerm: 0644,
					WorkDir:     "./",
					Args:        args,
				}

				d, err := dm.Reborn()
				if err != nil {
					logger.Fatal("Start failed, please check is there another instance running.")
				}
				if d != nil {
					return nil
				}
				defer dm.Release()

				logger.SetLogger(config.LogFilePath)

				err = initApplication()
				if err != nil {
					return err
				}

				scheduler.Start()

				return nil
			},
		},
		{
			Name:  "run",
			Usage: "Run GoBackup",
			Flags: buildFlags([]cli.Flag{}),
			Action: func(ctx *cli.Context) error {
				logger.SetLogger(config.LogFilePath)

				err := initApplication()
				if err != nil {
					return err
				}

				scheduler.Start()

				web.StartHTTP(version)

				return nil
			},
		},
	}

	app.Run(os.Args)
}

func initApplication() error {
	err := config.Init(configFile)
	if err != nil {
		return err
	}

	return nil
}

func perform(modelNames []string) {
	var models []*model.Model
	if len(modelNames) == 0 {
		// perform all
		models = model.GetModels()
	} else {
		for _, name := range modelNames {
			if m := model.GetModelByName(name); m == nil {
				logger.Fatalf("Model %s not found in %s", name, viper.ConfigFileUsed())
			} else {
				models = append(models, m)
			}
		}
	}

	for _, m := range models {
		m.Perform()
	}
}
