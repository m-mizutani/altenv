package main_test

import (
	"bytes"
	"testing"

	. "github.com/m-mizutani/altenv"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandStdinAwsAssumeRole(t *testing.T) {
	inputData := `
	{
		"Credentials": {
			"AccessKeyId": "BLUE_KEY",
			"SecretAccessKey": "ORANGE_KEY",
			"SessionToken": "RED_TOKEN",
			"Expiration": "2020-05-02T13:29:23+00:00"
		},
		"AssumedRoleUser": {
			"AssumedRoleId": "ASSUMEROLEID:dev",
			"Arn": "arn:aws:sts::1234567890xx:assumed-role/Developper/yes"
		}
	}`

	buf := &bytes.Buffer{}
	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
			DryRunOutput: buf,
			OpenFunc:     fileNeverExists,
			Stdin:        ToReadCloser(inputData),
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("-i", "aws-assume-role"))
	require.NoError(t, err)
	envmap := toEnvVars(buf)
	assert.Equal(t, "BLUE_KEY", envmap["AWS_ACCESS_KEY_ID"])
	assert.Equal(t, "ORANGE_KEY", envmap["AWS_SECRET_ACCESS_KEY"])
	assert.Equal(t, "RED_TOKEN", envmap["AWS_SESSION_TOKEN"])
}

func TestCommandStdinAwsAssumeRoleInvalidFormat(t *testing.T) {
	inputData := `
	{
		"Credentials": {
			"AccessKeyId": "BLUE_KEY",
			"SecretAccessKey": "ORANGE_KEY",
			"SessionToken": "RED_TOKEN",
			"Expiration": "2020-05-02T13:29:23+00:00"
		},
		"AssumedRoleUser": {
			"AssumedRoleId": "ASSUMEROLEID:dev",
			"Arn": "arn:aws:sts::1234567890xx:assumed-role/Developper/yes"
		}
	` // missing last bracket

	buf := &bytes.Buffer{}
	params := &Parameters{
		ExtIO: &ExtIOFunc{
			Getwd:        dummyGetwd,
			DryRunOutput: buf,
			OpenFunc:     fileNeverExists,
			Stdin:        ToReadCloser(inputData),
		},
	}
	app := NewApp(params)

	err := app.Run(newArgs("-i", "aws-assume-role"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Fail to parse AWS assume role response")
}
