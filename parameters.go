package main

import (
	"io"
	"os"

	"github.com/Songmu/prompter"
	cli "github.com/urfave/cli/v2"
)

func wrapOSOpen(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

type parameters struct {
	EnvFiles  cli.StringSlice
	JSONFiles cli.StringSlice
	Defines   cli.StringSlice
	Keychains cli.StringSlice
	Prompt    string

	Profile       string
	ConfigPath    string
	LogLevel      string
	Overwrite     string
	RunMode       string
	WriteKeyChain string

	// For testing
	DryRunOutput io.Writer
	OpenFunc     fileOpen
	InputFunc    promptInput
}

func newParameters() parameters {
	return parameters{
		DryRunOutput: os.Stdout,
		OpenFunc:     wrapOSOpen,
		InputFunc:    prompter.Password,
	}
}
