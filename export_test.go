//nolint
package main

import (
	"io"
	"io/ioutil"
	"strings"

	cli "github.com/urfave/cli/v2"
)

var (
	ReadEnvFile  = readEnvFile
	ReadJSONFile = readJSONFile
)

type Parameters parameters

// Wrappers
func NewApp(params *Parameters) *cli.App {
	return newApp((*parameters)(params))
}

func Run(params Parameters, args []string) error {
	return run((parameters)(params), args)
}

func LoadConfigFile(path string, params *Parameters) error {
	return loadConfigFile(path, (*parameters)(params))
}

// Utilities
func ToReadCloser(s string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(s))
}
