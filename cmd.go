package main

import (
	"fmt"
	"os"
	"path/filepath"

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

	// Setup configuration
	paramConfig := parametersToConfig(params)
	masterConfig, err := loadConfigFile(params.ConfigPath, params.Profile, params.OpenFunc)
	if err != nil {
		return err
	}

	if masterConfig != nil {
		masterConfig.merge(*paramConfig)
	} else {
		masterConfig = paramConfig
	}

	if err := masterConfig.finalize(); err != nil {
		return err
	}

	// Setup environment variables
	envvars, err := loadEnvVars(loadEnvVarsArgs{
		config:    masterConfig,
		openFunc:  params.OpenFunc,
		inputFunc: params.InputFunc,
		queryItem: params.KeychainQueryItem,
	})
	if err != nil {
		return err
	}

	switch params.RunMode {
	case "dryrun":
		if err := dumpEnvVars(params.DryRunOutput, envvars); err != nil {
			return err
		}

	case "update-keychain":
		if masterConfig.WriteKeychainNamespace == "" {
			return fmt.Errorf("--write-keychain-namespace option is required")
		}
		args := putKeyChainValuesArgs{
			envvars:       envvars,
			namespace:     masterConfig.WriteKeychainNamespace,
			servicePrefix: "",
			addItem:       params.KeychainAddItem,
			updateItem:    params.KeychainUpdateItem,
		}
		if err := putKeyChainValues(args); err != nil {
			return err
		}

	case "exec":
		if err := execCommand(envvars, args); err != nil {
			return err
		}

	default:
		return fmt.Errorf("Invalid run mode: `%s`", params.RunMode)
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
			&cli.StringSliceFlag{
				Name:        "keychain",
				Aliases:     []string{"k"},
				Usage:       "Read from Keychain of specified namespace",
				Destination: &params.Keychains,
			},
			&cli.StringFlag{
				Name:        "prompt",
				Usage:       "Set a variable by prompt. Try --prompt FOO -r dryrun",
				Destination: &params.Prompt,
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

			// Running mode
			&cli.StringFlag{
				Name:        "run-mode",
				Aliases:     []string{"r"},
				Usage:       "Run mode [exec|dryrun|update-keychain]",
				Value:       "exec",
				Destination: &params.RunMode,
			},

			&cli.StringFlag{
				Name:        "profile",
				Aliases:     []string{"p"},
				Usage:       "Use profile",
				Destination: &params.Profile,
				Value:       defaultProfileName,
			},

			&cli.StringFlag{
				Name:        "overwrite",
				Usage:       "Overwrite policy [allow|warn|deny] (default: deny)",
				Destination: &params.Overwrite,
			},

			&cli.StringFlag{
				Name:        "write-keychain-namespace",
				Aliases:     []string{"w"},
				Usage:       "keychain namespace to write. Required for run-mode update-keychain",
				Destination: &params.WriteKeyChain,
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
