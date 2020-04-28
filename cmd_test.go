package main_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	. "github.com/m-mizutani/altenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cli "github.com/urfave/cli/v2"
)

func toEnvVars(buf *bytes.Buffer) map[string]string { // nolint
	envmap := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(buf.Bytes()))
	for scanner.Scan() {
		arr := strings.Split(scanner.Text(), "=")
		envmap[arr[0]] = strings.Join(arr[1:], "=")
	}

	return envmap
}

func makeParameters(buf *bytes.Buffer) Parameters {
	params := NewParameters()
	params.DryRun = true
	params.DryRunOutput = buf
	params.LogLevel = "info"
	params.OpenFunc = func(fname string) (io.ReadCloser, error) {
		switch fname {
		case "my.env":
			return ToReadCloser("COLOR=BLUE"), nil
		case "my.json":
			return ToReadCloser(`{"MAGIC":"5"}`), nil
		default:
			return nil, fmt.Errorf("File not found")
		}
	}

	return params
}

func TestCommandEnvFile(t *testing.T) {
	buf := bytes.Buffer{}
	params := makeParameters(&buf)

	params.EnvFiles = *cli.NewStringSlice("my.env")
	err := Run(params, []string{})
	require.NoError(t, err)

	envmap := toEnvVars(&buf)
	assert.Contains(t, envmap, "COLOR")
	assert.NotContains(t, envmap, "MAGIC")
	assert.Equal(t, "BLUE", envmap["COLOR"])
}

func TestCommandJSONFile(t *testing.T) {
	buf := bytes.Buffer{}
	params := makeParameters(&buf)

	params.JSONFiles = *cli.NewStringSlice("my.json")
	err := Run(params, []string{})
	require.NoError(t, err)

	envmap := toEnvVars(&buf)
	assert.NotContains(t, envmap, "COLOR")
	assert.Contains(t, envmap, "MAGIC")
	assert.Equal(t, "5", envmap["MAGIC"])
}

func TestCommandMixFiles(t *testing.T) {
	buf := bytes.Buffer{}
	params := makeParameters(&buf)

	params.JSONFiles = *cli.NewStringSlice("my.json")
	params.EnvFiles = *cli.NewStringSlice("my.env")
	err := Run(params, []string{})
	require.NoError(t, err)

	envmap := toEnvVars(&buf)
	assert.Contains(t, envmap, "COLOR")
	assert.Equal(t, "BLUE", envmap["COLOR"])
	assert.Contains(t, envmap, "MAGIC")
	assert.Equal(t, "5", envmap["MAGIC"])
}
