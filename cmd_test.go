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
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
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
		},
	}

	return params
}

func newArgs(args ...string) []string {
	base := []string{"altenv", "-r", "dryrun"}
	return append(base, args...)
}

func fileNeverExists(string) (io.ReadCloser, error) {
	return nil, os.ErrNotExist
}

func dummyGetwd() (string, error) {
	return "/some/where", nil
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

envfile = [
	"envfile_global_1.env",
	"envfile_global_2.env",
]

[profile]
	[profile.default]
	envfile = ["envfile_default_profile.env"]

	[profile.temp]
	envfile = ["envfile_temp_profile.env"]
`

	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
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

func TestOverwriteDefaultDeny(t *testing.T) {
	buf := &bytes.Buffer{}
	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
			DryRunOutput: buf,
			OpenFunc:     fileNeverExists,
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("-d", "COLOR=BLUE", "-d", "COLOR=ORANGE"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Deny to overwrite")
}

func TestOverwriteExplicitDeny(t *testing.T) {
	buf := &bytes.Buffer{}
	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
			DryRunOutput: buf,
			OpenFunc:     fileNeverExists,
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("--overwrite", "deny", "-d", "COLOR=BLUE", "-d", "COLOR=ORANGE"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Deny to overwrite")
}

func TestOverwriteExplicitWarn(t *testing.T) {
	buf := &bytes.Buffer{}
	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
			DryRunOutput: buf,
			OpenFunc:     fileNeverExists,
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("--overwrite", "warn", "-l", "error", "-d", "COLOR=BLUE", "-d", "COLOR=ORANGE"))
	require.NoError(t, err)
	envmap := toEnvVars(buf)
	assert.Equal(t, "ORANGE", envmap["COLOR"])
}

func TestOverwriteExplicitAllow(t *testing.T) {
	buf := &bytes.Buffer{}
	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
			DryRunOutput: buf,
			OpenFunc:     fileNeverExists,
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("--overwrite", "allow", "-d", "COLOR=BLUE", "-d", "COLOR=ORANGE"))
	require.NoError(t, err)
	envmap := toEnvVars(buf)
	assert.Equal(t, "ORANGE", envmap["COLOR"])
}

func TestOverwriteInvalidPolicy(t *testing.T) {
	buf := &bytes.Buffer{}
	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
			DryRunOutput: buf,
			OpenFunc:     fileNeverExists,
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("--overwrite", "xxx", "-d", "COLOR=BLUE"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not valid overwrite option")
}

func TestOverwriteEvnFile(t *testing.T) {
	buf := &bytes.Buffer{}

	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
			DryRunOutput: buf,
			OpenFunc: func(fname string) (io.ReadCloser, error) {
				switch fname {
				case "my.env":
					return ToReadCloser("COLOR=BLUE"), nil
				default:
					return nil, os.ErrNotExist
				}
			},
		},
	}
	app := NewApp(params)
	err := app.Run(newArgs("--overwrite", "allow", "-e", "my.env", "-d", "COLOR=ORANGE"))
	require.NoError(t, err)
	envmap := toEnvVars(buf)
	assert.Equal(t, "ORANGE", envmap["COLOR"])
}

func TestOverwriteSetInConfig(t *testing.T) {
	buf := &bytes.Buffer{}

	configData := `
[global]
	overwrite = "allow"
`
	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
			DryRunOutput: buf,
			OpenFunc: func(fname string) (io.ReadCloser, error) {
				switch fname {
				case "testconfig":
					return ToReadCloser(configData), nil
				default:
					return nil, os.ErrNotExist
				}
			},
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("-c", "testconfig", "-d", "COLOR=BLUE", "-d", "COLOR=ORANGE"))
	require.NoError(t, err)
	envmap := toEnvVars(buf)
	assert.Equal(t, "ORANGE", envmap["COLOR"])
}

func TestCommandPrompt(t *testing.T) {
	buf := &bytes.Buffer{}
	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
			DryRunOutput: buf,
			OpenFunc:     fileNeverExists,
			InputFunc:    func(string) string { return "BLUE" },
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("--prompt", "COLOR"))
	require.NoError(t, err)
	envmap := toEnvVars(buf)
	assert.Equal(t, "BLUE", envmap["COLOR"])
}

func TestCommandStdinEnvfile(t *testing.T) {
	buf := &bytes.Buffer{}
	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
			DryRunOutput: buf,
			OpenFunc:     fileNeverExists,
			Stdin:        ToReadCloser("COLOR=BLUE"),
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("-i", "env"))
	require.NoError(t, err)
	envmap := toEnvVars(buf)
	assert.Equal(t, "BLUE", envmap["COLOR"])
}

func TestCommandStdinJSONfile(t *testing.T) {
	buf := &bytes.Buffer{}
	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
			DryRunOutput: buf,
			OpenFunc:     fileNeverExists,
			Stdin:        ToReadCloser(`{"COLOR":"BLUE"}`),
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("-i", "json"))
	require.NoError(t, err)
	envmap := toEnvVars(buf)
	assert.Equal(t, "BLUE", envmap["COLOR"])
}

func TestCommandStdinInvalidFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
			DryRunOutput: buf,
			OpenFunc:     fileNeverExists,
			Stdin:        ToReadCloser(`xxx`),
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("-i", "env"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Fail parse envfile")
}

func TestCommandStdinInvalidOption(t *testing.T) {
	buf := &bytes.Buffer{}
	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
			DryRunOutput: buf,
			OpenFunc:     fileNeverExists,
			Stdin:        ToReadCloser(`COLOR=ORANGE`),
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("-i", "xxx"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid input format")
}

func TestConfigDir(t *testing.T) {
	configData := `
[workdir.proj1]
dirpath = "/path/to/proj1/"
define = ["COLOR=BLUE"]

[workdir.proj1srcany]
dirpath = "/path/to/proj1/src"
define = ["FOO=BAA"]

[workdir.proj1src]
dirpath = "/path/to/proj1/src/"
define = ["MAGIC=FIFTH"]

[workdir.proj1srcxxx]
dirpath = "/path/to/proj1/src/xxx"
define = ["NOT=SANE"]

[workdir.proj2]
dirpath = "/path/to/proj2"
define = ["WORDS=TIMELESS"]
`
	buf := &bytes.Buffer{}
	cwd := "/path/to/proj1/src"
	params := &Parameters{
		ExtIO: &ExtIOFunc{
			DryRunOutput: buf,
			OpenFunc: func(fname string) (io.ReadCloser, error) {
				switch fname {
				case "testconfig":
					return ToReadCloser(configData), nil
				default:
					return nil, os.ErrNotExist
				}
			},
			Getwd: func() (string, error) {
				return cwd, nil
			},
		},
	}

	app := NewApp(params)
	err := app.Run(newArgs("-c", "testconfig"))
	require.NoError(t, err)
	envmap := toEnvVars(buf)
	assert.Equal(t, "BLUE", envmap["COLOR"])
	assert.Equal(t, "FIFTH", envmap["MAGIC"])
	assert.Equal(t, "BAA", envmap["FOO"])
	assert.NotContains(t, envmap, "WORDS")
	assert.NotContains(t, envmap, "NOT")
}
