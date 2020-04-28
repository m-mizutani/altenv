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

func makeParameters(buf *bytes.Buffer) *Parameters {
	params := &Parameters{
		DryRunOutput: buf,
		OpenFunc: func(fname string) (io.ReadCloser, error) {
			switch fname {
			case "my.env":
				return ToReadCloser("COLOR=BLUE"), nil
			case "my.json":
				return ToReadCloser(`{"MAGIC":"5"}`), nil
			default:
				return nil, fmt.Errorf("File not found (testing)")
			}
		},
	}

	return params
}

func newArgs(args ...string) []string {
	base := []string{"altenv", "--dryrun"}
	return append(base, args...)
}

func TestCommandEnvFile(t *testing.T) {
	buf := bytes.Buffer{}
	app := NewApp(makeParameters(&buf))

	err := app.Run(newArgs("-e", "my.env"))
	require.NoError(t, err)

	envmap := toEnvVars(&buf)
	assert.Contains(t, envmap, "COLOR")
	assert.NotContains(t, envmap, "MAGIC")
	assert.Equal(t, "BLUE", envmap["COLOR"])
}

func TestCommandJSONFile(t *testing.T) {
	buf := bytes.Buffer{}
	app := NewApp(makeParameters(&buf))

	err := app.Run(newArgs("-j", "my.json"))
	require.NoError(t, err)

	envmap := toEnvVars(&buf)
	assert.NotContains(t, envmap, "COLOR")
	assert.Contains(t, envmap, "MAGIC")
	assert.Equal(t, "5", envmap["MAGIC"])
}

func TestCommandMixFiles(t *testing.T) {
	buf := bytes.Buffer{}
	app := NewApp(makeParameters(&buf))

	err := app.Run(newArgs("-j", "my.json", "-e", "my.env"))
	require.NoError(t, err)

	envmap := toEnvVars(&buf)
	assert.Contains(t, envmap, "COLOR")
	assert.Equal(t, "BLUE", envmap["COLOR"])
	assert.Contains(t, envmap, "MAGIC")
	assert.Equal(t, "5", envmap["MAGIC"])
}

func TestCommandEnvFileNotFound(t *testing.T) {
	buf := bytes.Buffer{}
	app := NewApp(makeParameters(&buf))

	err := app.Run(newArgs("-j", "my.json", "-e", "bad.env"))
	require.Error(t, err)
}

func TestCommandJSONFileNotFound(t *testing.T) {
	buf := bytes.Buffer{}
	app := NewApp(makeParameters(&buf))

	err := app.Run(newArgs("-j", "bad.json", "-e", "my.env"))
	require.Error(t, err)
}
