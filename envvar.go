package main

import (
	"fmt"
	"io"
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

func loadEnvFiles(envFiles []string, ext ExtIOFunc) loadResult {
	var envvars []*envvar

	for _, path := range envFiles {
		logger.WithField("path", path).Debug("Read EnvFile")
		vars, err := readEnvFile(path, ext.OpenFunc)
		if err != nil {
			return loadResult{nil, errors.Wrapf(err, "Fail to read EnvFile %s", path)}
		}

		envvars = append(envvars, vars...)
	}

	return loadResult{envvars, nil}
}

func loadJSONFiles(jsonFiles []string, ext ExtIOFunc) loadResult {
	var envvars []*envvar

	for _, path := range jsonFiles {
		logger.WithField("path", path).Debug("Read JSON file")
		vars, err := readJSONFile(path, ext.OpenFunc)
		if err != nil {
			return loadResult{nil, errors.Wrapf(err, "Fail to read JSON file %s", path)}
		}
		envvars = append(envvars, vars...)
	}

	return loadResult{envvars, nil}
}

func loadDefines(defines []string) loadResult {
	var envvars []*envvar

	for _, def := range defines {
		logger.WithField("define", def).Debug("Set temp variables")
		v, err := parseDefine(def)
		if err != nil {
			return loadResult{nil, err}
		}
		envvars = append(envvars, v)
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

	parser := func(io.Reader) ([]*envvar, error) { return nil, nil }

	switch stdinFmt {
	case "json":
		parser = parseJSONFile
	case "env":
		parser = parseEnvFile
	case "aws-assume-role":
		parser = parseAwsAssumeRole
	case "":
		// nothing to do
	default:
		return loadResult{nil, fmt.Errorf("Invalid input format: `%s`", stdinFmt)}
	}

	vars, err := parser(ext.Stdin)
	if err != nil {
		return loadResult{nil, errors.Wrap(err, "Fail to parse data from stdin")}
	}
	envvars = append(envvars, vars...)

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
