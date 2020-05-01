package main_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/keybase/go-keychain"
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
	base := []string{"altenv", "-r", "dryrun"}
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

func TestConfigNotRequiredFile(t *testing.T) {
	buf := &bytes.Buffer{}

	configData := `
[global]
	[[global.envfile]]
	path = "envfile_global_1.env"
	required = false

	[[global.envfile]]
	path = "envfile_global_2.env"
	required = true
`
	params := &Parameters{
		DryRunOutput: buf,
		OpenFunc: func(fname string) (io.ReadCloser, error) {
			switch fname {
			case "testconfig":
				return ToReadCloser(configData), nil
			case "envfile_global_2.env":
				return ToReadCloser("COSMOS=STARS"), nil
			default:
				return nil, os.ErrNotExist
			}
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("-c", "testconfig"))
	require.NoError(t, err)

	envmap := toEnvVars(buf)
	assert.Equal(t, "STARS", envmap["COSMOS"])
}

func TestConfigRequiredFileNotFound(t *testing.T) {
	buf := &bytes.Buffer{}

	configData := `
[global]
	[[global.envfile]]
	path = "envfile_global_1.env"
	required = false

	[[global.envfile]]
	path = "envfile_global_2.env"
	required = true
`
	params := &Parameters{
		DryRunOutput: buf,
		OpenFunc: func(fname string) (io.ReadCloser, error) {
			switch fname {
			case "testconfig":
				return ToReadCloser(configData), nil
			default:
				return nil, os.ErrNotExist
			}
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("-c", "testconfig"))
	require.Error(t, err)
}

func fileNeverExists(string) (io.ReadCloser, error) {
	return nil, os.ErrNotExist
}

func TestOverwriteDefaultDeny(t *testing.T) {
	buf := &bytes.Buffer{}
	params := &Parameters{
		DryRunOutput: buf,
		OpenFunc:     fileNeverExists,
	}
	app := NewApp(params)

	err := app.Run(newArgs("-d", "COLOR=BLUE", "-d", "COLOR=ORANGE"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Deny to overwrite")
}

func TestOverwriteExplicitDeny(t *testing.T) {
	buf := &bytes.Buffer{}
	params := &Parameters{
		DryRunOutput: buf,
		OpenFunc:     fileNeverExists,
	}
	app := NewApp(params)

	err := app.Run(newArgs("--overwrite", "deny", "-d", "COLOR=BLUE", "-d", "COLOR=ORANGE"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Deny to overwrite")
}

func TestOverwriteExplicitWarn(t *testing.T) {
	buf := &bytes.Buffer{}
	params := &Parameters{
		DryRunOutput: buf,
		OpenFunc:     fileNeverExists,
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
		DryRunOutput: buf,
		OpenFunc:     fileNeverExists,
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
		DryRunOutput: buf,
		OpenFunc:     fileNeverExists,
	}
	app := NewApp(params)

	err := app.Run(newArgs("--overwrite", "xxx", "-d", "COLOR=BLUE"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not valid overwrite option")
}

func TestOverwriteEvnFile(t *testing.T) {
	buf := &bytes.Buffer{}

	params := &Parameters{
		DryRunOutput: buf,
		OpenFunc: func(fname string) (io.ReadCloser, error) {
			switch fname {
			case "my.env":
				return ToReadCloser("COLOR=BLUE"), nil
			default:
				return nil, os.ErrNotExist
			}
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
		DryRunOutput: buf,
		OpenFunc: func(fname string) (io.ReadCloser, error) {
			switch fname {
			case "testconfig":
				return ToReadCloser(configData), nil
			default:
				return nil, os.ErrNotExist
			}
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
		DryRunOutput: buf,
		OpenFunc:     fileNeverExists,
		InputFunc:    func(string) string { return "BLUE" },
	}
	app := NewApp(params)

	err := app.Run(newArgs("--prompt", "COLOR"))
	require.NoError(t, err)
	envmap := toEnvVars(buf)
	assert.Equal(t, "BLUE", envmap["COLOR"])
}

func getKeychainItemString(item keychain.Item, key string) string {
	value := reflect.ValueOf(item)
	vdata := value.FieldByName("attr")
	val := vdata.MapIndex(reflect.ValueOf(key))

	if val.IsValid() {
		return fmt.Sprintf("%v", val)
	} else {
		return ""
	}
}

func TestKeyChainPut(t *testing.T) {
	buf := &bytes.Buffer{}
	callAdd, callUpdate, callQuery := false, false, false
	params := &Parameters{
		DryRunOutput: buf,
		OpenFunc:     fileNeverExists,
		KeychainAddItem: func(item keychain.Item) error {
			callAdd = true
			assert.Equal(t, "altenv.ns1", getKeychainItemString(item, keychain.ServiceKey))
			account := getKeychainItemString(item, keychain.AccountKey)
			switch account {
			case "COLOR":
				return keychain.ErrorDuplicateItem // dup
			case "MAGIC":
				return nil // success
			default:
				require.Failf(t, "Inavlid account key: %s", account)
			}

			return nil
		},
		KeychainUpdateItem: func(query keychain.Item, item keychain.Item) error {
			// Update only COLOR by keychain.ErrorDuplicateItem
			callUpdate = true
			assert.Equal(t, "altenv.ns1", getKeychainItemString(query, keychain.ServiceKey))
			assert.Equal(t, "COLOR", getKeychainItemString(query, keychain.AccountKey))
			assert.Equal(t, "altenv.ns1", getKeychainItemString(item, keychain.ServiceKey))
			assert.Equal(t, "COLOR", getKeychainItemString(item, keychain.AccountKey))

			return nil
		},
		KeychainQueryItem: func(query keychain.Item) ([]keychain.QueryResult, error) {
			callQuery = true
			return nil, nil
		},
	}
	app := NewApp(params)

	err := app.Run([]string{"altenv",
		"-r", "update-keychain",
		"-d", "COLOR=BLUE", "-d", "MAGIC=5",
		"-w", "ns1",
	})
	require.NoError(t, err)
	//	envmap := toEnvVars(buf)
	assert.True(t, callAdd)
	assert.True(t, callUpdate)
	assert.False(t, callQuery)
}

func TestKeyChainGet(t *testing.T) {
	buf := &bytes.Buffer{}
	callAdd, callUpdate, callQueryAll, callQueryOne := 0, 0, 0, 0
	params := &Parameters{
		DryRunOutput: buf,
		OpenFunc:     fileNeverExists,
		KeychainAddItem: func(item keychain.Item) error {
			callAdd++
			return nil
		},
		KeychainUpdateItem: func(query keychain.Item, item keychain.Item) error {
			callUpdate++
			return nil
		},
		KeychainQueryItem: func(query keychain.Item) ([]keychain.QueryResult, error) {
			assert.Equal(t, "altenv.ns1", getKeychainItemString(query, keychain.ServiceKey))
			account := getKeychainItemString(query, keychain.AccountKey)

			switch account {
			case "":
				callQueryAll++
				return []keychain.QueryResult{
					{Account: "COLOR"},
					{Account: "MAGIC"},
				}, nil
			case "COLOR":
				callQueryOne++
				return []keychain.QueryResult{{Data: []byte("BLUE")}}, nil
			case "MAGIC":
				callQueryOne++
				return []keychain.QueryResult{{Data: []byte("5")}}, nil
			}

			return nil, fmt.Errorf("Fail")
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("-k", "ns1"))
	require.NoError(t, err)
	//	envmap := toEnvVars(buf)
	assert.Equal(t, 0, callAdd)
	assert.Equal(t, 0, callUpdate)
	assert.Equal(t, 1, callQueryAll)
	assert.Equal(t, 2, callQueryOne)

	envmap := toEnvVars(buf)
	assert.Equal(t, "BLUE", envmap["COLOR"])
	assert.Equal(t, "5", envmap["MAGIC"])
}

func TestKeyChainGetByConfig(t *testing.T) {
	buf := &bytes.Buffer{}
	callAdd, callUpdate, callQueryAll, callQueryOne := 0, 0, 0, 0
	configData := `
[global]
keychain=["ns1"]
`

	params := &Parameters{
		DryRunOutput: buf,
		OpenFunc: func(fname string) (io.ReadCloser, error) {
			switch fname {
			case "testconfig":
				return ToReadCloser(configData), nil
			default:
				return nil, os.ErrNotExist
			}
		},
		KeychainAddItem: func(item keychain.Item) error {
			callAdd++
			return nil
		},
		KeychainUpdateItem: func(query keychain.Item, item keychain.Item) error {
			callUpdate++
			return nil
		},
		KeychainQueryItem: func(query keychain.Item) ([]keychain.QueryResult, error) {
			assert.Equal(t, "altenv.ns1", getKeychainItemString(query, keychain.ServiceKey))
			account := getKeychainItemString(query, keychain.AccountKey)

			switch account {
			case "":
				callQueryAll++
				return []keychain.QueryResult{
					{Account: "COLOR"},
					{Account: "MAGIC"},
				}, nil
			case "COLOR":
				callQueryOne++
				return []keychain.QueryResult{{Data: []byte("BLUE")}}, nil
			case "MAGIC":
				callQueryOne++
				return []keychain.QueryResult{{Data: []byte("5")}}, nil
			}

			return nil, fmt.Errorf("Fail")
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("-c", "testconfig"))
	require.NoError(t, err)
	//	envmap := toEnvVars(buf)
	assert.Equal(t, 0, callAdd)
	assert.Equal(t, 0, callUpdate)
	assert.Equal(t, 1, callQueryAll)
	assert.Equal(t, 2, callQueryOne)

	envmap := toEnvVars(buf)
	assert.Equal(t, "BLUE", envmap["COLOR"])
	assert.Equal(t, "5", envmap["MAGIC"])
}

func TestKeyChainServicePrefix(t *testing.T) {
	buf := &bytes.Buffer{}
	callAdd, callQuery := 0, 0
	newParam := func() *Parameters {
		return &Parameters{
			DryRunOutput: buf,
			OpenFunc:     fileNeverExists,
			KeychainAddItem: func(item keychain.Item) error {
				callAdd++
				assert.Equal(t, "clocktower-ns1", getKeychainItemString(item, keychain.ServiceKey))
				return nil
			},
			KeychainUpdateItem: func(query keychain.Item, item keychain.Item) error { return nil },
			KeychainQueryItem: func(query keychain.Item) ([]keychain.QueryResult, error) {
				callQuery++
				assert.Equal(t, "clocktower-ns1", getKeychainItemString(query, keychain.ServiceKey))
				switch getKeychainItemString(query, keychain.AccountKey) {
				case "":
					return []keychain.QueryResult{{Account: "COLOR"}}, nil
				case "COLOR":
					return []keychain.QueryResult{{Data: []byte("BLUE")}}, nil
				default:
					return nil, fmt.Errorf("Fail")
				}
			},
		}
	}

	// Put action
	app1 := NewApp(newParam())
	err := app1.Run([]string{"altenv",
		"-r", "update-keychain",
		"--keychain-service-prefix", "clocktower-",
		"-d", "COLOR=BLUE",
		"-w", "ns1",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, callAdd)
	assert.Equal(t, 0, callQuery)

	// Get action
	app2 := NewApp(newParam())
	err = app2.Run(newArgs("-k", "ns1", "--keychain-service-prefix", "clocktower-"))
	require.NoError(t, err)
	assert.Equal(t, 1, callAdd)
	assert.Equal(t, 2, callQuery)
	envmap := toEnvVars(buf)
	assert.Equal(t, "BLUE", envmap["COLOR"])
}

func TestKeyChainServicePrefixByConfig(t *testing.T) {
	buf := &bytes.Buffer{}
	callAdd := 0
	configData := `
[global]
keychainServicePrefix="clocktower-"
`
	newParam := func() *Parameters {
		return &Parameters{
			DryRunOutput: buf,
			OpenFunc: func(fname string) (io.ReadCloser, error) {
				switch fname {
				case "testconfig":
					return ToReadCloser(configData), nil
				default:
					return nil, os.ErrNotExist
				}
			},
			KeychainAddItem: func(item keychain.Item) error {
				callAdd++
				assert.Equal(t, "clocktower-ns1", getKeychainItemString(item, keychain.ServiceKey))
				return nil
			},
			KeychainUpdateItem: func(query keychain.Item, item keychain.Item) error { return nil },
			KeychainQueryItem: func(query keychain.Item) ([]keychain.QueryResult, error) {
				return nil, nil
			},
		}
	}

	// Put action
	app := NewApp(newParam())
	err := app.Run([]string{"altenv",
		"-c", "testconfig",
		"-r", "update-keychain",
		"-d", "COLOR=BLUE",
		"-w", "ns1",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, callAdd)
}
