package main

import (
	"io"
	"os"

	cli "github.com/urfave/cli/v2"
)

type fileOpen func(string) (io.ReadCloser, error)

func wrapOSOpen(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

type parameters struct {
	EnvFiles  cli.StringSlice
	JSONFiles cli.StringSlice
	Defines   cli.StringSlice

	Profile    string
	ConfigPath string
	LogLevel   string

	// For testing
	DryRun       bool
	DryRunOutput io.Writer
	OpenFunc     fileOpen
}

func newParameters() parameters {
	return parameters{
		DryRunOutput: os.Stdout,
		OpenFunc:     wrapOSOpen,
	}
}
