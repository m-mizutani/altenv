// +build linux

package main

import (
	"fmt"
)

func putKeyChainValues(envvars []*envvar, namespace string) error {
	return fmt.Errorf("Keychain is not supported in the OS")
}

func getKeyChainValues(namespace string) ([]*envvar, error) {
	return nil, fmt.Errorf("Keychain is not supported in the OS")
}
