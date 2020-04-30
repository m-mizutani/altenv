package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
)

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
