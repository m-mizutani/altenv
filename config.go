package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

// loadConfigFile reads config from path and update params. It's destructive function.
func loadConfigFile(path string, profile string, ext ExtIOFunc) (*altenvConfig, error) {
	fd, err := ext.OpenFunc(path)
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

	cwd, err := ext.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "Fail to get CWD")
	}

	return parseConfigFile(fd, profile, cwd)
}

type configFile struct {
	Global   altenvConfig            `toml:"global"`
	Profiles map[string]altenvConfig `toml:"profile"`
	Workdirs map[string]altenvConfig `toml:"workdir"`
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
	EnvFiles  []string `toml:"envfile"`
	JSONFiles []string `toml:"jsonfile"`
	Defines   []string `toml:"define"`
	Keychains []string `toml:"keychain"`
	Overwrite *string  `toml:"overwrite"`

	KeychainServicePrefix string `toml:"keychainServicePrefix"`

	// Config identifiers
	// DirPath is read in all section, but available only in WorkDir
	DirPath string `toml:"dirpath"`

	// Only available by CLI option
	Prompt                 string `toml:"-"`
	Stdin                  string `toml:"-"`
	WriteKeychainNamespace string `toml:"-"`

	overwrite overwritePolicy
}

func (x *altenvConfig) merge(src altenvConfig) {
	x.EnvFiles = append(x.EnvFiles, src.EnvFiles...)
	x.JSONFiles = append(x.JSONFiles, src.JSONFiles...)
	x.Defines = append(x.Defines, src.Defines...)
	x.Keychains = append(x.Keychains, src.Keychains...)
	if src.Overwrite != nil {
		x.Overwrite = src.Overwrite
	}
	if src.KeychainServicePrefix != "" {
		x.KeychainServicePrefix = src.KeychainServicePrefix
	}
	if src.WriteKeychainNamespace != "" {
		x.WriteKeychainNamespace = src.WriteKeychainNamespace
	}
	if src.Stdin != "" {
		x.Stdin = src.Stdin
	}
	if src.Prompt != "" {
		x.Prompt = src.Prompt
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

	config.EnvFiles = append(config.EnvFiles, params.EnvFiles.Value()...)
	config.JSONFiles = append(config.JSONFiles, params.JSONFiles.Value()...)
	config.Defines = append(config.Defines, params.Defines.Value()...)

	config.Keychains = append(config.Keychains, params.Keychains.Value()...)

	config.Prompt = params.Prompt
	config.Stdin = params.Stdin
	config.WriteKeychainNamespace = params.WriteKeyChain
	config.KeychainServicePrefix = params.KeychainServicePrefix

	if params.Overwrite != "" {
		config.Overwrite = &params.Overwrite
	}

	return config
}

func parseConfigFile(fd io.Reader, profile, cwd string) (*altenvConfig, error) {
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

	var dirCfgs []altenvConfig
	for k, dir := range fileCfg.Workdirs {
		if dir.DirPath == "" {
			return nil, fmt.Errorf("workdir config `%s` has no `dirpath` field", k)
		}
		if strings.HasPrefix(cwd, dir.DirPath) {
			dirCfgs = append(dirCfgs, dir)
		}
	}

	config.merge(fileCfg.Global)
	for _, dirCfg := range dirCfgs {
		config.merge(dirCfg)
	}
	config.merge(profileCfg)

	return &config, nil
}
