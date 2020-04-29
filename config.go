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
	Path     string `toml:"path"`
	Required *bool  `toml:"required"`
}

type jsonFileConfig struct {
	Path     string `toml:"path"`
	Required *bool  `toml:"required"`
}

type defineConfig struct {
	// Expected FOO=BAR format
	Vars []string `toml:"vars"`
}

type configFile struct {
	Global   altenvConfig            `toml:"global"`
	Profiles map[string]altenvConfig `toml:"profile"`
}

type altenvConfig struct {
	EnvFiles  []*envFileConfig  `toml:"envfile"`
	JSONFiles []*jsonFileConfig `toml:"jsonfile"`
	Defines   []*defineConfig   `toml:"define"`
}

func (x *altenvConfig) merge(src altenvConfig) {
	x.EnvFiles = append(x.EnvFiles, src.EnvFiles...)
	x.JSONFiles = append(x.JSONFiles, src.JSONFiles...)
	x.Defines = append(x.Defines, src.Defines...)
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
		if defaultProfileName == profile {
			logger.Debug("profile is default, but no default profile in config")
			return nil, nil
		}
		return nil, fmt.Errorf("profile `%s` is not found in config file", profile)
	}

	config.merge(fileCfg.Global)
	config.merge(profileCfg)

	return &config, nil
}
