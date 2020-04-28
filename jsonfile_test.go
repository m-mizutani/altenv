package main_test

import (
	"io"
	"sort"
	"testing"

	. "github.com/m-mizutani/altenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONFile(t *testing.T) {
	data := `{
		"COLOR": "BLUE",
		"WORD": "FIVE"
	}`

	var openFileName string
	dummyOpen := func(fname string) (io.ReadCloser, error) {
		openFileName = fname
		return ToReadCloser(data), nil
	}

	envvars, err := ReadJSONFile("mytest.json", dummyOpen)
	require.NoError(t, err)
	assert.Equal(t, "mytest.json", openFileName)
	assert.Equal(t, 2, len(envvars))

	sort.Slice(envvars, func(i, j int) bool {
		return envvars[i].Key < envvars[j].Key
	})
	assert.Equal(t, "COLOR", envvars[0].Key)
	assert.Equal(t, "BLUE", envvars[0].Value)
	assert.Equal(t, "WORD", envvars[1].Key)
	assert.Equal(t, "FIVE", envvars[1].Value)
}

func TestJSONFileError(t *testing.T) {
	data := `{
		"COLOR": "BLUE",
		"WORD": "FIVE",
	}` // invalid comma in third line

	envvars, err := ReadJSONFile("mytest.json", func(string) (io.ReadCloser, error) {
		return ToReadCloser(data), nil
	})
	require.Error(t, err)
	assert.Nil(t, envvars)
}
