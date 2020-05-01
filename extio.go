package main

import (
	"io"
	"os"

	"github.com/Songmu/prompter"
)

// ExtIOFunc is external IO function set.
type ExtIOFunc struct {
	DryRunOutput       io.Writer
	OpenFunc           fileOpen
	InputFunc          promptInput
	KeychainAddItem    keychainAddItem
	KeychainUpdateItem keychainUpdateItem
	KeychainQueryItem  keychainQueryItem
}

// NewExtIOFunc is constructor to set default IO functions
func NewExtIOFunc() *ExtIOFunc {
	extIO := &ExtIOFunc{
		DryRunOutput: os.Stdout,
		OpenFunc:     wrapOSOpen,
		InputFunc:    prompter.Password,
	}
	setupKeychainFunc(extIO)
	return extIO
}
