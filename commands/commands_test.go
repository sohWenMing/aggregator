package commands

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/sohWenMing/aggregator/internal/config"
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
	execError := commandsPtr.execCommand(cmd, &buf, "", nil)
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

	commandsPtr := InitCommands()
	commandsPtr.registerAllHandlers()
	args := []string{"test-program", "login", "nindgabeet"}
	cmd, err := ParseCommand(args)
	testutils.AssertNoErr(err, t)

	initialConfig, err := config.Read()
	testutils.AssertNoErr(err, t)
	buf := bytes.Buffer{}
	execError := commandsPtr.execCommand(cmd, &buf, "", initialConfig)
	testutils.AssertNoErr(execError, t)

	configAfterSetUser, err := config.Read()
	testutils.AssertNoErr(err, t)
	testutils.AssertStrings(configAfterSetUser.CurrentUserName, "nindgabeet", t)

}
