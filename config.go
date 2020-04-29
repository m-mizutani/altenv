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
func loadConfigFile(path string, params *parameters) error {
	fd, err := params.OpenFunc(path)
	if os.IsNotExist(err) {
		// Ignore if path is default
		if path != defaultConfigPath {
			return fmt.Errorf("Config file is not found: %s", path)
		}

		logger.WithField("path", path).Debug("Config file does not exist")
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "Fail to open config file: %s", path)
	}
	defer fd.Close()

	return parseConfigFile(fd, params)
}

type envFileConfig struct {
	Path string `toml:"path"`
}

type jsonFileConfig struct {
	Path string `toml:"path"`
}

type profileConfig struct {
	EnvFiles  []envFileConfig  `toml:"envfile"`
	JSONFiles []jsonFileConfig `toml:"jsonfile"`
}

type altenvConfig struct {
	Global   profileConfig            `toml:"global"`
	Profiles map[string]profileConfig `toml:"profile"`
}

func applyProfileToParameters(profile profileConfig, params *parameters) error {
	for _, envConfig := range profile.EnvFiles {
		if err := params.EnvFiles.Set(envConfig.Path); err != nil {
			return errors.Wrap(err, "Fail to set envfile config")
		}
	}
	return nil
}

func parseConfigFile(fd io.Reader, params *parameters) error {
	var config altenvConfig
	raw, err := ioutil.ReadAll(fd)
	if err != nil {
		return errors.Wrap(err, "Fail to read data from config file")
	}

	if err := toml.Unmarshal(raw, &config); err != nil {
		return errors.Wrap(err, "Fail to parse toml config file")
	}

	profile, ok := config.Profiles[params.Profile]
	if !ok {
		if defaultProfileName == params.Profile {
			logger.Debug("profile is default, but no default profile in config")
			return nil
		}
		return fmt.Errorf("profile `%s` is not found in config file", params.Profile)
	}

	if err := applyProfileToParameters(config.Global, params); err != nil {
		return err
	}
	if err := applyProfileToParameters(profile, params); err != nil {
		return err
	}

	return nil
}
