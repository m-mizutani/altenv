package main

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

var defaultConfigPath = filepath.Join(os.Getenv("HOME"), ".altenv")

func run(params parameters, args []string) error {
	if err := setLogLevel(params.LogLevel); err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"params": params,
		"args":   args,
	}).Debug("Run altenv")

	if err := loadConfigFile(params.ConfigPath, &params); err != nil {
		return err
	}

	var envvars []*envvar

	// Read environment variables
	if len(params.EnvFiles.Value()) > 0 {
		for _, fpath := range params.EnvFiles.Value() {
			logger.WithField("path", fpath).Debug("Read EnvFile")
			vars, err := readEnvFile(fpath, params.OpenFunc)
			if err != nil {
				return errors.Wrapf(err, "Fail to read EnvFile %s", fpath)
			}
			envvars = append(envvars, vars...)
		}
	}

	if len(params.JSONFiles.Value()) > 0 {
		for _, fpath := range params.JSONFiles.Value() {
			logger.WithField("path", fpath).Debug("Read JSON file")
			vars, err := readJSONFile(fpath, params.OpenFunc)
			if err != nil {
				return errors.Wrapf(err, "Fail to read JSON file %s", fpath)
			}
			envvars = append(envvars, vars...)
		}
	}

	if params.DryRun {
		// Dryrun
		if err := dumpEnvVars(params.DryRunOutput, envvars); err != nil {
			return err
		}
	} else {
		// Execute command
		if err := execCommand(envvars, args); err != nil {
			return err
		}
	}

	return nil
}

func newApp(params *parameters) *cli.App {
	app := &cli.App{
		Name:  "altenv",
		Usage: "Powerful CLI Environment Variable Manager",
		Action: func(c *cli.Context) error {
			var args []string
			for i := 0; i < c.Args().Len(); i++ {
				args = append(args, c.Args().Get(i))
			}

			if err := run(*params, args); err != nil {
				return err
			}
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "env",
				Aliases:     []string{"e"},
				Usage:       "Read from EnvVar file",
				Destination: &params.EnvFiles,
			},
			&cli.StringSliceFlag{
				Name:        "json",
				Aliases:     []string{"j"},
				Usage:       "Read from JSON file",
				Destination: &params.JSONFiles,
			},

			&cli.StringFlag{
				Name:        "log-level",
				Aliases:     []string{"l"},
				Usage:       "Set log level [trace|debug|info|warn|error|fatal]",
				Destination: &params.LogLevel,
				Value:       "info",
			},
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Usage:       "Config file",
				Destination: &params.ConfigPath,
				Value:       defaultConfigPath,
			},
			&cli.BoolFlag{
				Name:        "dryrun",
				Usage:       "Enable dryrun mode",
				Destination: &params.DryRun,
			},
			/*
				&cli.StringFlag{
					Name:        "profile",
					Aliases:     []string{"p"},
					Usage:       "Use profile",
					Destination: &args.EnvFile,
				},
			*/
		},
	}

	return app
}

func main() {
	params := newParameters()
	app := newApp(&params)

	err := app.Run(os.Args)
	if err != nil {
		logger.WithError(err).Fatal("altenv failed")
	}
}
