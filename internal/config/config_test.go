package config

import (
	"bytes"
	"os"
	"testing"

	definederrors "github.com/sohWenMing/aggregator/defined_errors"
	errorutils "github.com/sohWenMing/aggregator/error_utils"
	testUtils "github.com/sohWenMing/aggregator/test_utils"
)

func TestReadFunc(t *testing.T) {
	config, err := Read()
	testUtils.AssertNoErr(err, t)
	testUtils.AssertStrings(config.DbUrl, os.Getenv("DB_STRING"), t)
}

func TestSetUser(t *testing.T) {
	type testStruct struct {
		name          string
		input         string
		isErrExpected bool
		expectedErr   error
	}

	tests := []testStruct{
		{
			"test success",
			"nindgabeet",
			false,
			nil,
		},
		{
			"test nil input",
			"",
			true,
			definederrors.ErrorInput,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config, _ := Read()
			buf := bytes.Buffer{}
			err := config.SetUser(test.input, &buf)
			switch test.isErrExpected {
			case true:
				testUtils.AssertHasErr(err, t)
				isErrMatch := errorutils.CheckErrTypeMatch(err, test.expectedErr)
				if !isErrMatch {
					t.Errorf("got error: %v\nexpected error: %v", err, test.expectedErr)
				}
			case false:
				testUtils.AssertNoErr(err, t)
				newConfig, _ := Read()
				testUtils.AssertStrings(newConfig.CurrentUserName, test.input, t)
			}

		})
	}
}
