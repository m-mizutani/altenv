package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

// loadConfigFile reads config from path and update params. It's destructive function.
func loadConfigFile(path string, profile string, open fileOpen) (*altenvConfig, error) {
	fd, err := open(path)
	if os.IsNotExist(err) {
		// Ignore if path is default
		if path != defaultConfigPath {
			return nil, fmt.Errorf("Config file is not found: %s", path)
		}

		logger.WithField("path", path).Debug("Config file does not exist")
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "Fail to open config file: %s", path)
	}
	defer fd.Close()

	return parseConfigFile(fd, profile)
}

type envFileConfig struct {
	Path string `toml:"path"`
	// If true, fail when file not found. Default (nil) is true
	Required *bool `toml:"required"`
}

func (x *envFileConfig) IsRequired() bool {
	return (x.Required == nil || *x.Required)
}

type jsonFileConfig struct {
	Path string `toml:"path"`
	// If true, fail when file not found. Default (nil) is true
	Required *bool `toml:"required"`
}

func (x *jsonFileConfig) IsRequired() bool {
	return (x.Required == nil || *x.Required)
}

type defineConfig struct {
	// Expected FOO=BAR format
	Vars []string `toml:"vars"`
}

type configFile struct {
	Global   altenvConfig            `toml:"global"`
	Profiles map[string]altenvConfig `toml:"profile"`
}

type overwritePolicy int

const (
	overwriteDeny = iota
	overwriteWarn
	overwriteAllow
)

var overwritePolicyMap = map[string]overwritePolicy{
	"deny":  overwriteDeny,
	"warn":  overwriteWarn,
	"allow": overwriteAllow,
}

type altenvConfig struct {
	EnvFiles  []*envFileConfig  `toml:"envfile"`
	JSONFiles []*jsonFileConfig `toml:"jsonfile"`
	Defines   []*defineConfig   `toml:"define"`
	Prompt    string            `toml:"-"` // not available in toml
	Overwrite *string           `toml:"overwrite"`

	overwrite overwritePolicy
}

func (x *altenvConfig) merge(src altenvConfig) {
	x.EnvFiles = append(x.EnvFiles, src.EnvFiles...)
	x.JSONFiles = append(x.JSONFiles, src.JSONFiles...)
	x.Defines = append(x.Defines, src.Defines...)
	if src.Overwrite != nil {
		x.Overwrite = src.Overwrite
	}
}

func (x *altenvConfig) finalize() error {
	if x.Overwrite == nil {
		deny := "deny"
		x.Overwrite = &deny
	}

	policy, ok := overwritePolicyMap[*x.Overwrite]
	if !ok {
		return fmt.Errorf("`%s` is not valid overwrite option, must be [deny|warn|allow]", *x.Overwrite)
	}
	x.overwrite = policy

	return nil
}

func parametersToConfig(params parameters) *altenvConfig {
	config := &altenvConfig{}
	for _, fpath := range params.EnvFiles.Value() {
		config.EnvFiles = append(config.EnvFiles, &envFileConfig{
			Path: fpath,
		})
	}

	for _, fpath := range params.JSONFiles.Value() {
		config.JSONFiles = append(config.JSONFiles, &jsonFileConfig{
			Path: fpath,
		})
	}

	config.Defines = append(config.Defines, &defineConfig{
		Vars: params.Defines.Value(),
	})

	config.Prompt = params.Prompt

	if params.Overwrite != "" {
		config.Overwrite = &params.Overwrite
	}

	return config
}

func parseConfigFile(fd io.Reader, profile string) (*altenvConfig, error) {
	var fileCfg configFile
	var config altenvConfig

	raw, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to read data from config file")
	}

	if err := toml.Unmarshal(raw, &fileCfg); err != nil {
		return nil, errors.Wrap(err, "Fail to parse toml config file")
	}

	profileCfg, ok := fileCfg.Profiles[profile]
	if !ok {
		if defaultProfileName != profile {
			return nil, fmt.Errorf("profile `%s` is not found in config file", profile)
		}
		logger.Debug("profile is default, but no default profile in config")
	}

	config.merge(fileCfg.Global)
	config.merge(profileCfg)

	return &config, nil
}
