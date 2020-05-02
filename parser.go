package main

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
)

func parseAwsAssumeRole(fd io.Reader) ([]*envvar, error) {
	var assumeRole sts.AssumeRoleOutput
	raw, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(raw, &assumeRole); err != nil {
		return nil, errors.Wrapf(err, "Fail to parse AWS assume role response")
	}

	return []*envvar{
		{Key: "AWS_ACCESS_KEY_ID", Value: aws.StringValue(assumeRole.Credentials.AccessKeyId)},
		{Key: "AWS_SECRET_ACCESS_KEY", Value: aws.StringValue(assumeRole.Credentials.SecretAccessKey)},
		{Key: "AWS_SESSION_TOKEN", Value: aws.StringValue(assumeRole.Credentials.SessionToken)},
	}, nil
}
