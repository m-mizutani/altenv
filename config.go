package main

import (
	"fmt"
	"io"
	"os"

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

func parseConfigFile(fd io.Reader, params *parameters) error {
	return nil
}
