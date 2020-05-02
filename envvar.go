package main

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type envvar struct {
	Key   string
	Value string
}

func dumpEnvVars(w io.Writer, vars []*envvar) error {
	sort.Slice(vars, func(i, j int) bool {
		return vars[i].Key < vars[j].Key
	})

	for _, v := range vars {
		if _, err := fmt.Fprintf(w, "%s=%s\n", v.Key, v.Value); err != nil {
			return errors.Wrap(err, "Fail to output dryrun results")
		}
	}
	return nil
}

type fileOpen func(string) (io.ReadCloser, error) // based on os.Open
type promptInput func(string) string              // based on prompter.Password

type loadResult struct {
	EnvVars []*envvar
	Error   error
}

func loadEnvVars(config altenvConfig, ext ExtIOFunc) ([]*envvar, error) {
	var envvars []*envvar

	// Read environment variables
	results := []loadResult{
		loadEnvFiles(config.EnvFiles, ext),
		loadJSONFiles(config.JSONFiles, ext),
		loadDefines(config.Defines),
		loadKeychain(config.Keychains, config.KeychainServicePrefix, ext),
		loadStdin(config.Stdin, ext),
		loadPrompt(config.Prompt, ext),
	}

	for _, result := range results {
		if result.Error != nil {
			return nil, result.Error
		}
		envvars = append(envvars, result.EnvVars...)
	}

	// Check overwrite
	varmap := map[string]*envvar{}
	for _, v := range envvars {
		if existValue, ok := varmap[v.Key]; ok {
			logFields := logrus.Fields{
				"key": v.Key,
				"old": existValue,
				"new": v.Value,
			}

			switch config.overwrite {
			case overwriteDeny:
				return nil, fmt.Errorf("Deny to overwrite `%s`, `%s` -> `%s`", v.Key, existValue, v.Value)
			case overwriteWarn:
				logger.WithFields(logFields).Warn("Overwrote environment variable")
			case overwriteAllow:
				logger.WithFields(logFields).Debug("Overwrote environment variable")
			}
		}
		varmap[v.Key] = v
	}

	var newVars []*envvar
	for _, v := range varmap {
		newVars = append(newVars, v)
	}

	return newVars, nil
}

func loadEnvFiles(envFiles []*envFileConfig, ext ExtIOFunc) loadResult {
	var envvars []*envvar

	for _, f := range envFiles {
		logger.WithField("path", f.Path).Debug("Read EnvFile")
		vars, err := readEnvFile(f.Path, ext.OpenFunc)
		if os.IsNotExist(err) && !f.IsRequired() {
			logger.WithField("path", f.Path).Debug("EnvFile is not found, but ignore because not required")
			continue
		} else if err != nil {
			return loadResult{nil, errors.Wrapf(err, "Fail to read EnvFile %s", f.Path)}
		}

		envvars = append(envvars, vars...)
	}

	return loadResult{envvars, nil}
}

func loadJSONFiles(jsonFiles []*jsonFileConfig, ext ExtIOFunc) loadResult {
	var envvars []*envvar

	for _, f := range jsonFiles {
		logger.WithField("path", f.Path).Debug("Read JSON file")
		vars, err := readJSONFile(f.Path, ext.OpenFunc)
		if os.IsNotExist(err) && !f.IsRequired() {
			logger.WithField("path", f.Path).Debug("JSON File is not found, but ignore because not required")
			continue
		} else if err != nil {
			return loadResult{nil, errors.Wrapf(err, "Fail to read JSON file %s", f.Path)}
		}
		envvars = append(envvars, vars...)
	}

	return loadResult{envvars, nil}
}

func loadDefines(defines []*defineConfig) loadResult {
	var envvars []*envvar

	for _, def := range defines {
		for _, vdef := range def.Vars {
			logger.WithField("define", vdef).Debug("Set temp variables")
			v, err := parseDefine(vdef)
			if err != nil {
				return loadResult{nil, err}
			}
			envvars = append(envvars, v)
		}
	}

	return loadResult{envvars, nil}
}

func loadKeychain(keychains []string, servicePrefix string, ext ExtIOFunc) loadResult {
	var envvars []*envvar

	for _, namespace := range keychains {
		args := getKeyChainValuesArgs{
			namespace:     namespace,
			servicePrefix: servicePrefix,
			queryItem:     ext.KeychainQueryItem,
		}
		vars, err := getKeyChainValues(args)
		if err != nil {
			return loadResult{nil, err}
		}
		envvars = append(envvars, vars...)
	}

	return loadResult{envvars, nil}
}

func loadStdin(stdinFmt string, ext ExtIOFunc) loadResult {
	var envvars []*envvar

	switch stdinFmt {
	case "json":
		vars, err := parseJSONFile(ext.Stdin)
		if err != nil {
			return loadResult{nil, errors.Wrap(err, "Fail to parse JSON format from stdin")}
		}
		envvars = append(envvars, vars...)
	case "env":
		vars, err := parseEnvFile(ext.Stdin)
		if err != nil {
			return loadResult{nil, errors.Wrap(err, "Fail to parse EnvFile format from stdin")}
		}
		envvars = append(envvars, vars...)
	case "":
		// nothing to do
	default:
		return loadResult{nil, fmt.Errorf("Invalid input format: `%s`", stdinFmt)}
	}

	return loadResult{envvars, nil}
}

func loadPrompt(prompt string, ext ExtIOFunc) loadResult {
	var envvars []*envvar

	if prompt != "" {
		value := ext.InputFunc(fmt.Sprintf("Enter %s value", prompt))
		envvars = append(envvars, &envvar{
			Key:   prompt,
			Value: value,
		})
	}

	return loadResult{envvars, nil}
}
