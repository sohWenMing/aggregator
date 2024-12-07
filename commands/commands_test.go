package commands

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	definedErrors "github.com/sohWenMing/aggregator/defined_errors"
	"github.com/sohWenMing/aggregator/internal/config"
	testHelpers "github.com/sohWenMing/aggregator/test_helpers"
)

func TestHandlerLogin(t *testing.T) {
	type testStruct struct {
		name             string
		cmd              Command
		isErrExpected    bool
		expectedErr      error
		expectedUserName string
	}
	tests := []testStruct{
		{
			"testing no arguments",
			Command{
				"login",
				[]string{},
			},
			true,
			definedErrors.ErrLoginHandlerZeroArgs,
			"",
		},
		{
			"testing all empty args",
			Command{
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
			Command{
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
			handlerErr := HandlerLogin(&currentConfig, test.cmd, &buf)
			switch test.isErrExpected {
			case true:
				testHelpers.AssertHasError(handlerErr, t)
				testHelpers.AssertErrorType(handlerErr, test.expectedErr, t)
			case false:
				testHelpers.AssertNoError(handlerErr, t)
				configAfterWrite, err := config.Read()
				testHelpers.AssertNoError(err, t)
				testHelpers.AssertStrings(configAfterWrite.Current_user_name, test.expectedUserName, t)
				testHelpers.AssertStrings(buf.String(), fmt.Sprintln("Username has been set to:", test.expectedUserName), t)
			}
		})
	}
}

func TestRegister(t *testing.T) {
	commandStruct := Commands{}
	commandStruct.Register("login", HandlerLogin)
	_, ok := commandStruct.commandMap["login"]
	if !ok {
		t.Errorf("login was not registered")
	}
	testHelpers.AssertStrings(fmt.Sprintf("%p", commandStruct.commandMap["login"]), fmt.Sprintf("%p", HandlerLogin), t)
}

func testHandler(c *config.Config, cmd Command, w io.Writer) error {
	fmt.Fprint(w, cmd.name)
	return nil
}

func TestRun(t *testing.T) {
	type testStruct struct {
		name          string
		command       Command
		commandStruct Commands
		isErrExpected bool
		expectedErr   error
	}

	tests := []testStruct{
		{
			"testing empty Commands struct error",
			Command{},
			Commands{},
			true,
			definedErrors.ErrCommandMapNil,
		},
		{
			"testing non existing command error",
			Command{"wrong command", []string{}},
			Commands{
				commandMap: map[string]func(c *config.Config, cmd Command, w io.Writer) error{
					"login": HandlerLogin,
				},
			},
			true,
			definedErrors.ErrCommandNotExist,
		},
		{
			"testing successful registering of command",
			Command{"testHandler", []string{}},
			Commands{
				commandMap: map[string]func(c *config.Config, cmd Command, w io.Writer) error{
					"testHandler": testHandler,
				},
			},
			false,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			currentConfig, err := config.Read()
			testHelpers.AssertNoError(err, t)
			buf := bytes.Buffer{}
			commandRunErr := test.commandStruct.Run(&currentConfig, test.command, &buf)
			switch test.isErrExpected {
			case true:
				testHelpers.AssertHasError(commandRunErr, t)
				testHelpers.AssertErrorType(commandRunErr, test.expectedErr, t)
			case false:
				testHelpers.AssertStrings(buf.String(), test.command.name, t)
			}

		})
	}
}

func TestGenerateCommand(t *testing.T) {
	type testStruct struct {
		name          string
		args          []string
		isErrExpected bool
		expectedCmd   Command
		expectedErr   error
	}

	tests := []testStruct{
		{
			"testing no args",
			[]string{},
			true,
			Command{},
			definedErrors.ErrNotEnoughArguments,
		},
		{
			"testing 1 arg",
			[]string{"one"},
			true,
			Command{},
			definedErrors.ErrNotEnoughArguments,
		},
		{
			"testing correct args",
			[]string{"one", "two", "three"},
			false,
			Command{
				"one",
				[]string{
					"two", "three",
				},
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GenerateCommand(test.args)
			switch test.isErrExpected {
			case true:
				testHelpers.AssertErrorType(err, test.expectedErr, t)
			case false:
				testHelpers.AssertStrings(test.expectedCmd.name, got.name, t)
				for i, arg := range got.args {
					testHelpers.AssertStrings(test.expectedCmd.args[i], arg, t)
				}
			}
		})
	}
}
