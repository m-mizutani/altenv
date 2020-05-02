package main

import (
	"encoding/json"
	"io"

	"github.com/sirupsen/logrus"
)

func readJSONFile(fpath string, open fileOpen) ([]*envvar, error) {
	fd, err := open(fpath)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	return parseJSONFile(fd)
}

func parseJSONFile(fd io.Reader) ([]*envvar, error) {
	var jdata map[string]string

	if err := json.NewDecoder(fd).Decode(&jdata); err != nil {
		return nil, err
	}
	var envvars []*envvar

	for key, value := range jdata {
		envvars = append(envvars, &envvar{Key: key, Value: value})

		logger.WithFields(logrus.Fields{
			"type":  "jsonfile",
			"key":   key,
			"value": value,
		}).Debug("Add a new variable")
	}

	return envvars, nil
}
