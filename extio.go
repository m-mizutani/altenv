package main

import (
	"io"
	"os"

	"github.com/Songmu/prompter"
)

// ExtIOFunc is external IO function set.
type ExtIOFunc struct {
	DryRunOutput       io.Writer
	Stdin              io.Reader
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
		Stdin:        os.Stdin,
		OpenFunc:     wrapOSOpen,
		InputFunc:    prompter.Password,
	}
	setupKeychainFunc(extIO)
	return extIO
}
