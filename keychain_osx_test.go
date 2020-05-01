// +build darwin

package main_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	. "github.com/m-mizutani/altenv"

	keychain "github.com/keybase/go-keychain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		ExtIO: &ExtIOFunc{
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
		ExtIO: &ExtIOFunc{
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
			ExtIO: &ExtIOFunc{
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
			KeychainAddItem: func(item keychain.Item) error {
				callAdd++
				assert.Equal(t, "clocktower-ns1", getKeychainItemString(item, keychain.ServiceKey))
				return nil
			},
			KeychainUpdateItem: func(query keychain.Item, item keychain.Item) error { return nil },
			KeychainQueryItem: func(query keychain.Item) ([]keychain.QueryResult, error) {
				return nil, nil
			},
		},
	}

	// Put action
	app := NewApp(params)
	err := app.Run([]string{"altenv",
		"-c", "testconfig",
		"-r", "update-keychain",
		"-d", "COLOR=BLUE",
		"-w", "ns1",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, callAdd)
}
