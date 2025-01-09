package commands

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"testing"

	definederrors "github.com/sohWenMing/aggregator/defined_errors"
	errorutils "github.com/sohWenMing/aggregator/error_utils"
	"github.com/sohWenMing/aggregator/internal/config"
	"github.com/sohWenMing/aggregator/internal/database"
	testutils "github.com/sohWenMing/aggregator/test_utils"
)

func TestParseCommand(t *testing.T) {
	type testStruct struct {
		name          string
		input         []string
		expected      enteredCommand
		isErrExpected bool
		expectedErr   error
	}

	tests := []testStruct{
		{
			"test login success",
			[]string{"gator", "login", "nindgabeet"},
			enteredCommand{"login", []string{"nindgabeet"}},
			false,
			nil,
		},
		{
			"test login fail",
			[]string{"gator", "login"},
			enteredCommand{"login", []string{}},
			false,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseCommand(test.input)
			switch test.isErrExpected {
			case true:
				testutils.AssertHasErr(err, t)
			case false:
				testutils.AssertNoErr(err, t)
				testutils.AssertStrings(got.name, test.input[1], t)

				switch len(test.input) > 2 {
				case true:
					enteredArgs := test.input[2:]
					for i, arg := range enteredArgs {
						testutils.AssertStrings(got.args[i], arg, t)
					}
				case false:
					testutils.AssertInts(len(got.args), 0, t)
				}

			}

		})
	}
}

func TestInitAndExecCommand(t *testing.T) {

	commandsPtr := InitCommands()
	nameToHandlers := []nameToHandler{
		{
			"test",
			handlerTest,
		},
	}
	commandsPtr.registerAllHandlersTest(nameToHandlers)
	args := []string{"test-program", "test", "test_string_1, test_string_2"}
	cmd, err := ParseCommand(args)
	testutils.AssertNoErr(err, t)
	buf := bytes.Buffer{}
	execError := commandsPtr.ExecCommand(cmd, &buf, nil)
	testutils.AssertNoErr(execError, t)
	linesInBuf := []string{}
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		linesInBuf = append(linesInBuf, scanner.Text())
	}
	for i, lineInBuf := range linesInBuf {
		testutils.AssertStrings(lineInBuf, cmd.args[i], t)
	}

}

func TestExecCommands(t *testing.T) {

	commandsPtr, state := initCommandsAndState(t)
	resetErr := state.Db.ResetUsers(context.Background())
	testutils.AssertNoErr(resetErr, t)

	//register nindgabeet

	registerArgs := []string{"test-program", "register", "nindgabeet"}
	registerCmd, err := ParseCommand(registerArgs)
	testutils.AssertNoErr(err, t)
	registerBuf := bytes.Buffer{}
	registerExecError := commandsPtr.ExecCommand(registerCmd, &registerBuf, state)
	testutils.AssertNoErr(registerExecError, t)

	//login nindgabeet
	args := []string{"test-program", "LogIn", "nindgabeet"}
	cmd, err := ParseCommand(args)
	testutils.AssertNoErr(err, t)
	testutils.AssertNoErr(err, t)
	buf := bytes.Buffer{}
	execError := commandsPtr.ExecCommand(cmd, &buf, state)
	testutils.AssertNoErr(execError, t)

	configAfterSetUser, err := config.Read()
	testutils.AssertNoErr(err, t)
	testutils.AssertStrings(configAfterSetUser.CurrentUserName, "nindgabeet", t)

}

func TestRegisterHandler(t *testing.T) {

	type testStruct struct {
		name           string
		args           []string
		isErrExpected  bool
		expectedError  error
		expectedWrites []string
	}

	tests := []testStruct{
		{
			"test initial register nindgabeet",
			[]string{"test-program", "register", "nindgabeet"},
			false,
			nil,
			[]string{"user nindgabeet has been added"},
		},
		{
			"test fail register nindgabeet",
			[]string{"test-program", "register", "nindgabeet"},
			true,
			definederrors.ErrorUserAlreadyExists,
			[]string{},
		},
	}

	commandsPtr, state := initCommandsAndState(t)

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			buf := bytes.Buffer{}
			cmd, err := ParseCommand(test.args)
			testutils.AssertNoErr(err, t)
			execCommandErr := commandsPtr.ExecCommand(cmd, &buf, state)
			switch test.isErrExpected {
			case true:
				testutils.AssertHasErr(execCommandErr, t)
				isErrMatch := errorutils.CheckErrTypeMatch(execCommandErr, test.expectedError)
				if !isErrMatch {
					t.Errorf("got %v\nwant%v", execCommandErr, test.expectedError)
				}
			case false:
				scanner := bufio.NewScanner(&buf)
				got := []string{}
				for scanner.Scan() {
					got = append(got, scanner.Text())
				}
				for i, write := range got {
					testutils.AssertStrings(test.expectedWrites[i], write, t)
				}

			}
		})
	}

}

func TestLoginHandler(t *testing.T) {
	type testStruct struct {
		name             string
		args             []string
		registerUserArgs []string
		expectedOutputs  []string
		isErrExpected    bool
		expectedErr      error
	}

	tests := []testStruct{
		{
			name:             "test user doesn't exist",
			args:             []string{"test-program", "login", "noExist"},
			registerUserArgs: []string{},
			isErrExpected:    true,
			expectedOutputs:  []string{"user noExist could not be retrieved, user is not logged in"},
			expectedErr:      definederrors.ErrorUserNotFound,
		},
		{
			name:             "test user nindgabeet exist",
			args:             []string{"test-program", "login", "nindgabeet"},
			registerUserArgs: []string{"test-program", "register", "nindgabeet"},
			isErrExpected:    false,
			expectedOutputs: []string{
				"user nindgabeet has been added",
				"user nindgabeet is now logged in",
			},
			expectedErr: nil,
		},
	}
	commandsPtr, state := initCommandsAndState(t)
	resetErr := state.Db.ResetUsers(context.Background())
	testutils.AssertNoErr(resetErr, t)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := bytes.Buffer{}
			switch test.isErrExpected {
			case true:
				cmd, err := ParseCommand(test.args)
				testutils.AssertNoErr(err, t)
				execErr := commandsPtr.ExecCommand(cmd, &buf, state)
				isErrMatch := errorutils.CheckUnwrappedError(execErr, test.expectedErr)
				if !isErrMatch {
					t.Errorf("got: %v\nwant: %v", execErr, test.expectedErr)
				}
			case false:
				registerCmd, err := ParseCommand(test.registerUserArgs)
				testutils.AssertNoErr(err, t)
				execErr := commandsPtr.ExecCommand(registerCmd, &buf, state)
				testutils.AssertNoErr(execErr, t)
				loginCmd, err := ParseCommand(test.args)
				testutils.AssertNoErr(err, t)
				loginExecErr := commandsPtr.ExecCommand(loginCmd, &buf, state)
				testutils.AssertNoErr(loginExecErr, t)
			}
			linesInBuf := getLinesInBuf(buf)
			if len(linesInBuf) != len(test.expectedOutputs) {
				t.Fatalf("num lines in buf: %d\n numlines in expectedOutputs: %d", len(linesInBuf), len(test.expectedOutputs))
			}
		})
	}
}

func TestGetUsers(t *testing.T) {

	commandsPtr, state := initCommandsAndState(t)
	resetErr := state.Db.ResetUsers(context.Background())
	testutils.AssertNoErr(resetErr, t)

	usersToRegister := []string{"kahya", "holgith"}
	registerArgs := [][]string{}

	for _, userToRegister := range usersToRegister {
		registerArg := []string{"test-program", "register", userToRegister}
		registerArgs = append(registerArgs, registerArg)
	}
	// loop registers the users to the database
	buf := bytes.Buffer{}
	for _, args := range registerArgs {
		cmd, err := ParseCommand(args)
		testutils.AssertNoErr(err, t)
		execCommandErr := commandsPtr.ExecCommand(cmd, &buf, state)
		testutils.AssertNoErr(execCommandErr, t)
	}

	getUsersCmd, err := ParseCommand([]string{"test-program", "users"})
	testutils.AssertNoErr(err, t)

	gotBuf := bytes.Buffer{}
	getUsersErr := commandsPtr.ExecCommand(getUsersCmd, &gotBuf, state)
	testutils.AssertNoErr(getUsersErr, t)

	linesInBuf := getLinesInBuf(gotBuf)
	expected := []string{
		"* kahya",
		"* holgith (current)",
	}
	for i, lineInBuf := range linesInBuf {
		testutils.AssertStrings(lineInBuf, expected[i], t)
	}
}

func TestHandlerAgg(t *testing.T) {
	commandsPtr, state := initCommandsAndState(t)
	buf := bytes.Buffer{}
	handlerAggCmd, err := ParseCommand([]string{"test-program", "agg"})

	testutils.AssertNoErr(err, t)
	aggErr := commandsPtr.ExecCommand(handlerAggCmd, &buf, state)
	testutils.AssertNoErr(aggErr, t)
	if !strings.Contains(buf.String(), "The Zen of Proverbs") {
		t.Errorf("buffer should have %q written to it\n", "The Zen of Proverbs")
	}

	if !strings.Contains(buf.String(), "Optimize for simplicity") {
		t.Errorf("buffer should have %q written to it\n", "Optimize for simplicity")
	}

}
func initCommandsAndState(t *testing.T) (*commands, *database.State) {
	commandsPtr := InitCommands()
	commandsPtr.registerAllHandlers()
	state, err := database.CreateDBConnection()
	testutils.AssertNoErr(err, t)
	resetErr := state.Db.ResetUsers(context.Background())
	testutils.AssertNoErr(resetErr, t)
	return commandsPtr, state
}

func getLinesInBuf(buf bytes.Buffer) []string {
	linesInBuf := []string{}
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		linesInBuf = append(linesInBuf, scanner.Text())
	}
	return linesInBuf
}
