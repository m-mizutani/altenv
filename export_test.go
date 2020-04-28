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

func ToReadCloser(s string) io.ReadCloser { // nolint
	return ioutil.NopCloser(strings.NewReader(s))
}
