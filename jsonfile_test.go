package main_test

import (
	"bytes"
	"sort"
	"testing"

	. "github.com/m-mizutani/envctl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONFile(t *testing.T) {
	data := `{
		"COLOR": "BLUE",
		"WORD": "FIVE"
	}`

	envvars, err := ParseJSONFile(bytes.NewReader([]byte(data)))
	require.NoError(t, err)
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

	envvars, err := ParseJSONFile(bytes.NewReader([]byte(data)))
	require.Error(t, err)
	assert.Nil(t, envvars)
}
