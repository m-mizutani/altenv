package main

import (
	"io"
	"os"

	"github.com/Songmu/prompter"
)

type fileOpen func(string) (io.ReadCloser, error) // based on os.Open
type promptInput func(string) string              // based on prompter.Password
type getWD func() (string, error)                 // based on os.Getwd

// ExtIOFunc is external IO function set.
type ExtIOFunc struct {
	DryRunOutput       io.Writer
	Stdin              io.Reader
	OpenFunc           fileOpen
	InputFunc          promptInput
	Getwd              getWD
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
		Getwd:        os.Getwd,
	}
	setupKeychainFunc(extIO)
	return extIO
}
