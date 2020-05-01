// +build linux

package main

import (
	"fmt"
)

type keychainAddItem func() error
type keychainUpdateItem func() error
type keychainQueryItem func() error

func setupKeychainFunc(extIO *ExtIOFunc) {
	extIO.KeychainAddItem = func() error { return nil }
	extIO.KeychainUpdateItem = func() error { return nil }
	extIO.KeychainQueryItem = func() error { return nil }
}

func putKeyChainValues(putKeyChainValuesArgs) error {
	return fmt.Errorf("Keychain is not supported in the OS")
}

func getKeyChainValues(getKeyChainValuesArgs) ([]*envvar, error) {
	return nil, fmt.Errorf("Keychain is not supported in the OS")
}
