package main_test

import (
	"bytes"
	"strings"
	"testing"

	. "github.com/m-mizutani/altenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvFile(t *testing.T) {
	data := strings.Join([]string{
		"COLOR=BLUE",            // normal
		" W1=FIVE ",             // should be trimed
		"W2 = TIMELESS WORDS ",  // should be trimed around key and value
		"  ",                    // blank line, should be skipped
		" # this is comment",    // sould be skipped
		"W3 = SPELLBOUND=NIGHT", // should be imported including sencond '='
	}, "\n")

	envvars, err := ParseEnvFile(bytes.NewReader([]byte(data)))
	require.NoError(t, err)
	assert.Equal(t, 4, len(envvars))

	assert.Equal(t, "COLOR", envvars[0].Key)
	assert.Equal(t, "BLUE", envvars[0].Value)

	assert.Equal(t, "W1", envvars[1].Key)
	assert.Equal(t, "FIVE", envvars[1].Value)

	assert.Equal(t, "W2", envvars[2].Key)
	assert.Equal(t, "TIMELESS WORDS", envvars[2].Value)

	assert.Equal(t, "W3", envvars[3].Key)
	assert.Equal(t, "SPELLBOUND=NIGHT", envvars[3].Value)
}

func TestEnvFileError(t *testing.T) {
	data := "W1XXXX" // No '='
	ret, err := ParseEnvFile(bytes.NewReader([]byte(data)))
	assert.Error(t, err)
	assert.Nil(t, ret)
}
