package commands

import (
	"fmt"
	"io"

	definederrors "github.com/sohWenMing/aggregator/defined_errors"
	"github.com/sohWenMing/aggregator/internal/config"
)

// ############# command struct, used to house all the configured commands with relevant methods ######### //
type commands struct {
	commandMap map[string]func(cmd enteredCommand, w io.Writer, filePath string, c *config.Config) (err error)
}

func (c *commands) execCommand(cmd enteredCommand, w io.Writer, filePath string, config *config.Config) (err error) {
	handler, ok := c.commandMap[cmd.name]
	if !ok {
		return definederrors.ErrorHandlerNotExist
	}
	handlerErr := handler(cmd, w, filePath, config)
	if handlerErr != nil {
		return handlerErr
	}
	return nil
}

func (c *commands) registerAllHandlers(nameToHandlers []nameToHandler) (err error) {
	for _, nameToHandler := range nameToHandlers {
		err := c.registerHandler(nameToHandler.name, nameToHandler.handler)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *commands) registerHandler(name string, handler func(cmd enteredCommand, w io.Writer, filePath string, c *config.Config) (err error)) (err error) {
	if c.commandMap == nil {
		return fmt.Errorf("pointer to commandMap is nil pointer %w", definederrors.ErrorNilPointer)
	}
	_, found := c.commandMap[name]
	if found {
		return fmt.Errorf("handler %s already exists in commandMap", name)
	}
	c.commandMap[name] = handler
	return nil
}

type nameToHandler struct {
	name    string
	handler func(cmd enteredCommand, w io.Writer, filePath string, c *config.Config) (err error)
}

func handlerTest(cmd enteredCommand, w io.Writer, filePath string, config *config.Config) (err error) {
	for _, arg := range cmd.args {
		fmt.Fprintln(w, arg)
	}
	return nil
}

// called at the main program, used to initialise the commandMap so that it can be written to
func InitCommands() (commandsPtr *commands) {
	returnedCommands := commands{}
	commandMap := make(map[string]func(cmd enteredCommand, w io.Writer, filePath string, c *config.Config) (err error))
	returnedCommands.commandMap = commandMap
	return &returnedCommands
}

// function to parse the input from os.Args, if no error should return a parsed enteredCommand
func ParseCommand(args []string) (cmd enteredCommand, err error) {
	returnedCmd := enteredCommand{}
	switch len(args) {
	case 0:
		return returnedCmd,
			fmt.Errorf("no arguments passed into ParseCommand %w",
				definederrors.ErrorNoArgs)
	case 1:
		return returnedCmd,
			fmt.Errorf("only one arguement passed into ParseCommand arg:%s %w",
				args[0], definederrors.ErrorWrongNumArgs)
	default:
		returnedCmd.name = args[1]
		returnedCmd.args = args[2:]
		return returnedCmd, nil
	}
}

type enteredCommand struct {
	name string
	args []string
}

/*
what needs to be achieved

when something is entered as a command, it needs to be parsed into an entered command
the entered command's name has to be used to correlate to the map of strings to handlers


*/
