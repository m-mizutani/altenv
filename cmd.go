package main

import (
	"log"
	"os"

	"github.com/pkg/errors"
	cli "github.com/urfave/cli/v2"
)

type parameters struct {
	EnvFile cli.StringSlice

	Profile    string
	ConfigPath string
	LogLevel   string
}

func run(params parameters, args []string) error {
	if err := setLogLevel(params.LogLevel); err != nil {
		return err
	}

	var envvars []*envvar

	// Read environment variables
	if len(params.EnvFile.Value()) > 0 {
		for _, fpath := range params.EnvFile.Value() {
			logger.WithField("path", fpath).Debug("Read EnvFile")
			vars, err := readEnvFile(fpath)
			if err != nil {
				return errors.Wrapf(err, "Fail to read EnvFile %s", fpath)
			}
			envvars = append(envvars, vars...)
		}
	}

	// Execute command
	if err := execCommand(envvars, args); err != nil {
		return err
	}

	return nil
}

func main() {
	var params parameters

	app := &cli.App{
		Name:  "envctl",
		Usage: "CLI Environment Variable Controller",
		Action: func(c *cli.Context) error {
			var args []string
			for i := 0; i < c.Args().Len(); i++ {
				args = append(args, c.Args().Get(i))
			}

			if err := run(params, args); err != nil {
				logger.WithError(err).Fatal("Failed")
			}
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "env",
				Aliases:     []string{"e"},
				Usage:       "Read from EnvVar file",
				Destination: &params.EnvFile,
			},
			&cli.StringFlag{
				Name:        "log-level",
				Aliases:     []string{"l"},
				Usage:       "Set log level [trace|debug|info|warn|error|fatal]",
				Destination: &params.LogLevel,
				Value:       "info",
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

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
