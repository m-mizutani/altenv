package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"syscall"

	"github.com/pkg/errors"
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

func execCommand(vars []*envvar, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("No arguments")
	}

	binary, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}

	envvars := os.Environ()
	for _, v := range vars {
		envvars = append(envvars, fmt.Sprintf("%s=%s", v.Key, v.Value))
	}

	if err := syscall.Exec(binary, args, envvars); err != nil {
		return errors.Wrapf(err, "Fail to exec: %v", args)
	}

	return nil
}
