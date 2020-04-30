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

type loadEnvVarsArgs struct {
	config    *altenvConfig
	openFunc  fileOpen
	inputFunc promptInput
}

func loadEnvVars(args loadEnvVarsArgs) ([]*envvar, error) {
	var envvars []*envvar

	// Read environment variables
	for _, f := range args.config.EnvFiles {
		logger.WithField("path", f.Path).Debug("Read EnvFile")
		vars, err := readEnvFile(f.Path, args.openFunc)
		if os.IsNotExist(err) && !f.IsRequired() {
			logger.WithField("path", f.Path).Debug("EnvFile is not found, but ignore because not required")
			continue
		} else if err != nil {
			return nil, errors.Wrapf(err, "Fail to read EnvFile %s", f.Path)
		}

		envvars = append(envvars, vars...)
	}

	for _, f := range args.config.JSONFiles {
		logger.WithField("path", f.Path).Debug("Read JSON file")
		vars, err := readJSONFile(f.Path, args.openFunc)
		if os.IsNotExist(err) && !f.IsRequired() {
			logger.WithField("path", f.Path).Debug("JSON File is not found, but ignore because not required")
			continue
		} else if err != nil {
			return nil, errors.Wrapf(err, "Fail to read JSON file %s", f.Path)
		}
		envvars = append(envvars, vars...)
	}

	for _, def := range args.config.Defines {
		for _, vdef := range def.Vars {
			logger.WithField("define", vdef).Debug("Set temp variables")
			v, err := parseDefine(vdef)
			if err != nil {
				return nil, err
			}
			envvars = append(envvars, v)
		}
	}

	for _, namespace := range args.config.Keychains {
		vars, err := getKeyChainValues(namespace)
		if err != nil {
			return nil, err
		}
		envvars = append(envvars, vars...)
	}

	if args.config.Prompt != "" {
		value := args.inputFunc(fmt.Sprintf("Enter %s value", args.config.Prompt))
		envvars = append(envvars, &envvar{
			Key:   args.config.Prompt,
			Value: value,
		})
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

			switch args.config.overwrite {
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
