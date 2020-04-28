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

func NewApp(params *Parameters) *cli.App {
	return newApp((*parameters)(params))
}

func Run(params Parameters, args []string) error {
	return run((parameters)(params), args)
}

func ToReadCloser(s string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(s))
}
