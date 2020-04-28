//nolint
package main

import (
	"io"
	"io/ioutil"
	"strings"
)

var (
	ReadEnvFile  = readEnvFile
	ReadJSONFile = readJSONFile
)

type Parameters parameters

func Run(params Parameters, args []string) error {
	return run((parameters)(params), args)
}
func NewParameters() Parameters {
	params := newParameters()
	return (Parameters)(params)
}

func ToReadCloser(s string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(s))
}
