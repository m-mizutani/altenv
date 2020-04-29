package main_test

import (
	"bufio"
	"bytes"
	"io"
	"os"
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

func makeParameters(buf *bytes.Buffer) *Parameters {
	params := &Parameters{
		DryRunOutput: buf,
		OpenFunc: func(fname string) (io.ReadCloser, error) {
			switch fname {
			case "my.env":
				return ToReadCloser("COLOR=BLUE"), nil
			case "my2.env":
				return ToReadCloser("COSMOS=STARS"), nil
			case "my.json":
				return ToReadCloser(`{"MAGIC":"5"}`), nil
			case "my2.json":
				return ToReadCloser(`{"COLOR":"orange"}`), nil
			default:
				return nil, os.ErrNotExist
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

func TestCommandMixMultipleJSONFiles(t *testing.T) {
	buf := bytes.Buffer{}
	app := NewApp(makeParameters(&buf))

	err := app.Run(newArgs("-j", "my.json", "-j", "my2.json"))
	require.NoError(t, err)

	envmap := toEnvVars(&buf)
	assert.Contains(t, envmap, "MAGIC")
	assert.Equal(t, "5", envmap["MAGIC"])
	assert.Contains(t, envmap, "COLOR")
	assert.Equal(t, "orange", envmap["COLOR"])
}

func TestCommandMixMultipleEnvFiles(t *testing.T) {
	buf := bytes.Buffer{}
	app := NewApp(makeParameters(&buf))

	err := app.Run(newArgs("-e", "my.env", "-e", "my2.env"))
	require.NoError(t, err)

	envmap := toEnvVars(&buf)
	assert.Contains(t, envmap, "COLOR")
	assert.Equal(t, "BLUE", envmap["COLOR"])
	assert.Contains(t, envmap, "COSMOS")
	assert.Equal(t, "STARS", envmap["COSMOS"])
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

func TestCommandDefine(t *testing.T) {
	buf := bytes.Buffer{}
	app := NewApp(makeParameters(&buf))

	err := app.Run(newArgs("-d", "COLOR=BLUE", "-d", "NOT=SANE"))
	require.NoError(t, err)

	envmap := toEnvVars(&buf)
	assert.Equal(t, "BLUE", envmap["COLOR"])
	assert.Equal(t, "SANE", envmap["NOT"])
}

func newConfigTestApp(buf *bytes.Buffer) *cli.App {
	configData := `
[global]
	[[global.envfile]]
	path = "envfile_global_1.env"

	[[global.envfile]]
	path = "envfile_global_2.env"

[profile]
	[profile.default]
		[[profile.default.envfile]]
		path = "envfile_default_profile.env"

	[profile.temp]
		[[profile.temp.envfile]]
		path = "envfile_temp_profile.env"
`

	params := &Parameters{
		DryRunOutput: buf,
		OpenFunc: func(fname string) (io.ReadCloser, error) {
			switch fname {
			case "testconfig":
				return ToReadCloser(configData), nil
			case "envfile_global_1.env":
				return ToReadCloser("COLOR=BLUE"), nil
			case "envfile_global_2.env":
				return ToReadCloser("COSMOS=STARS"), nil
			case "envfile_default_profile.env":
				return ToReadCloser("MAGIC=5"), nil
			case "envfile_temp_profile.env":
				return ToReadCloser("WORDS=TIMELESS"), nil
			default:
				return nil, os.ErrNotExist
			}
		},
	}
	app := NewApp(params)
	return app
}

func TestConfigGlobalEnvFile(t *testing.T) {
	buf := bytes.Buffer{}
	app := newConfigTestApp(&buf)

	err := app.Run(newArgs("-c", "testconfig"))
	require.NoError(t, err)

	envmap := toEnvVars(&buf)
	assert.Equal(t, "BLUE", envmap["COLOR"])
	assert.Equal(t, "STARS", envmap["COSMOS"])
	assert.Equal(t, "5", envmap["MAGIC"])
	assert.NotContains(t, envmap, "WORDS")
}

func TestConfigProfile(t *testing.T) {
	buf := bytes.Buffer{}
	app := newConfigTestApp(&buf)

	err := app.Run(newArgs("-c", "testconfig", "-p", "temp"))
	require.NoError(t, err)

	envmap := toEnvVars(&buf)
	assert.Equal(t, "BLUE", envmap["COLOR"])
	assert.Equal(t, "STARS", envmap["COSMOS"])
	assert.Equal(t, "TIMELESS", envmap["WORDS"])
	assert.NotContains(t, envmap, "MAGIC")
}
