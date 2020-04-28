package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

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

		rows := strings.Split(line, "=")
		if len(rows) < 2 {
			return nil, fmt.Errorf("Invalid format: '%s' at line %d", line, lineNo)
		}

		key := strings.TrimSpace(rows[0])
		value := strings.TrimSpace(strings.Join(rows[1:], "="))
		envvars = append(envvars, &envvar{Key: key, Value: value})

		logger.WithFields(logrus.Fields{
			"type":  "envfile",
			"key":   key,
			"value": value,
		}).Debug("Add a new variable")
	}

	return envvars, nil
}
