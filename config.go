package main

import (
	"io"
	"os"

	"github.com/pkg/errors"
)

// loadConfigFile reads config from path and update params. It's destructive function.
func loadConfigFile(path string, params *parameters) error {
	fd, err := os.Open(path)
	if os.IsNotExist(err) {
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
