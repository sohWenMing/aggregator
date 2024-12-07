package commands

import (
	"bytes"
	"testing"

	definedErrors "github.com/sohWenMing/aggregator/defined_errors"
	"github.com/sohWenMing/aggregator/internal/config"
	testHelpers "github.com/sohWenMing/aggregator/test_helpers"
)

func TestHandlerLogin(t *testing.T) {
	type testStruct struct {
		name             string
		cmd              command
		isErrExpected    bool
		expectedErr      error
		expectedUserName string
	}
	tests := []testStruct{
		{
			"testing no arguments",
			command{
				"login",
				[]string{},
			},
			true,
			definedErrors.ErrLoginHandlerZeroArgs,
			"",
		},
		{
			"testing all empty args",
			command{
				"login",
				[]string{
					"           ",
					"     ",
				},
			},
			true,
			definedErrors.ErrUserNameNil,
			"",
		},
		{
			"testing proper arguments",
			command{
				"login",
				[]string{
					"Soh",
					"Wen",
					"Ming",
				},
			},
			false,
			nil,
			"Soh Wen Ming",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			currentConfig, err := config.Read()
			testHelpers.AssertNoError(err, t)
			buf := bytes.Buffer{}
			handlerErr := handlerLogin(&currentConfig, test.cmd, &buf)
			switch test.isErrExpected {
			case true:
				testHelpers.AssertHasError(handlerErr, t)
				testHelpers.AssertErrorType(handlerErr, test.expectedErr, t)
			case false:
				testHelpers.AssertNoError(handlerErr, t)
				configAfterWrite, err := config.Read()
				testHelpers.AssertNoError(err, t)
				testHelpers.AssertStrings(configAfterWrite.Current_user_name, test.expectedUserName, t)
			}
		})
	}
}
