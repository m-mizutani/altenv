package main

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

const (
	altenvVersion      = "0.0.1"
	defaultProfileName = "default"
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

	paramConfig := parametersToConfig(params)
	config, err := loadConfigFile(params.ConfigPath, params.Profile, params.OpenFunc)
	if err != nil {
		return err
	}

	if config != nil {
		config.merge(*paramConfig)
	} else {
		config = paramConfig
	}

	var envvars []*envvar

	// Read environment variables
	for _, f := range config.EnvFiles {
		logger.WithField("path", f.Path).Debug("Read EnvFile")
		vars, err := readEnvFile(f.Path, params.OpenFunc)
		if os.IsNotExist(err) && !f.IsRequired() {
			logger.WithField("path", f.Path).Debug("EnvFile is not found, but ignore because not required")
			continue
		} else if err != nil {
			return errors.Wrapf(err, "Fail to read EnvFile %s", f.Path)
		}

		envvars = append(envvars, vars...)
	}

	for _, f := range config.JSONFiles {
		logger.WithField("path", f.Path).Debug("Read JSON file")
		vars, err := readJSONFile(f.Path, params.OpenFunc)
		if os.IsNotExist(err) && !f.IsRequired() {
			logger.WithField("path", f.Path).Debug("JSON File is not found, but ignore because not required")
			continue
		} else if err != nil {
			return errors.Wrapf(err, "Fail to read JSON file %s", f.Path)
		}
		envvars = append(envvars, vars...)
	}

	for _, def := range config.Defines {
		for _, vdef := range def.Vars {
			logger.WithField("define", vdef).Debug("Set temp variables")
			v, err := parseDefine(vdef)
			if err != nil {
				return err
			}
			envvars = append(envvars, v)
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
		Name:    "altenv",
		Usage:   "Powerful CLI Environment Variable Manager",
		Version: altenvVersion,
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
			&cli.StringSliceFlag{
				Name:        "define",
				Aliases:     []string{"d"},
				Usage:       "Set environment variable by FOO=BAR format",
				Destination: &params.Defines,
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

			&cli.StringFlag{
				Name:        "profile",
				Aliases:     []string{"p"},
				Usage:       "Use profile",
				Destination: &params.Profile,
				Value:       defaultProfileName,
			},
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
