package main

import (
	"io"
	"io/ioutil"
	"strings"
)

var (
	ReadEnvFile  = readEnvFile  // nolint
	ReadJSONFile = readJSONFile // nolint
)

type Parameters parameters

func Run(params Parameters, args []string) error {
	return run((parameters)(params), args)
}
func NewParameters() Parameters {
	params := newParameters()
	return (Parameters)(params)
}

func ToReadCloser(s string) io.ReadCloser { // nolint
	return ioutil.NopCloser(strings.NewReader(s))
}
