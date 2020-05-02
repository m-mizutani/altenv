package main

import (
	"io"
	"os"

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
	Stdin     string

	Profile               string
	ConfigPath            string
	LogLevel              string
	Overwrite             string
	RunMode               string
	WriteKeyChain         string
	KeychainServicePrefix string

	// For testing
	ExtIO *ExtIOFunc
}

func newParameters() parameters {
	return parameters{
		ExtIO: NewExtIOFunc(),
	}
}
