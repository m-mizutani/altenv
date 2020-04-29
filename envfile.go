package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func parseDefine(s string) (*envvar, error) {
	rows := strings.Split(s, "=")
	if len(rows) < 2 {
		return nil, fmt.Errorf("Invalid format: '%s'", s)
	}

	key := strings.TrimSpace(rows[0])
	value := strings.TrimSpace(strings.Join(rows[1:], "="))
	return &envvar{Key: key, Value: value}, nil
}

func readEnvFile(fpath string, open fileOpen) ([]*envvar, error) {
	fd, err := open(fpath)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	var envvars []*envvar
	scanner := bufio.NewScanner(fd)
	lineNo := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNo++

		if len(line) == 0 {
			continue // blank line
		} else if line[0:1] == "#" {
			continue // comment out
		}

		v, err := parseDefine(line)
		if err != nil {
			return nil, errors.Wrapf(err, "Fail parse envfile at line %d", lineNo)
		}
		envvars = append(envvars, v)

		logger.WithFields(logrus.Fields{
			"type":  "envfile",
			"key":   v.Key,
			"value": v.Value,
		}).Debug("Add a new variable")
	}

	return envvars, nil
}
