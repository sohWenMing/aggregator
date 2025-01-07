package commands

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	definederrors "github.com/sohWenMing/aggregator/defined_errors"
	errorutils "github.com/sohWenMing/aggregator/error_utils"
	"github.com/sohWenMing/aggregator/internal/database"
)

type handler func(cmd enteredCommand, w io.Writer, state *database.State) (err error)

// ############# command struct, used to house all the configured commands with relevant methods ######### //
type commands struct {
	commandMap map[string]handler
}

func (c *commands) ExecCommand(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	handler, ok := c.commandMap[cmd.name]
	if !ok {
		return definederrors.ErrorHandlerNotExist
	}
	handlerErr := handler(cmd, w, state)
	if handlerErr != nil {
		return handlerErr
	}
	return nil
}

func (c *commands) registerAllHandlers() (err error) {

	for _, nameToHandler := range initAllNameToHandlers() {
		err := c.registerHandler(nameToHandler.name, nameToHandler.handler)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *commands) registerAllHandlersTest(nameToHandlers []nameToHandler) (err error) {
	for _, nameToHandler := range nameToHandlers {
		err := c.registerHandler(nameToHandler.name, nameToHandler.handler)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *commands) registerHandler(name string, handler func(cmd enteredCommand, w io.Writer, state *database.State) (err error)) (err error) {
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
	handler func(cmd enteredCommand, w io.Writer, state *database.State) (err error)
}

func initAllNameToHandlers() []nameToHandler {
	returnedNameToHandlers := []nameToHandler{
		{"login", handlerLogin},
		{"register", handlerRegisterUser},
	}
	return returnedNameToHandlers

}

func handlerLogin(cmd enteredCommand, w io.Writer, state *database.State) (retrieveErr error) {
	if len(cmd.args) != 1 {
		return fmt.Errorf("args passed into handlerLogin %v %w", cmd.args, definederrors.ErrorWrongNumArgs)
	}
	_, err := state.Db.RetrieveUser(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Fprintf(w, "user %s could not be retrieved, user is not logged in\n", cmd.args[0])
		return fmt.Errorf("user %s could not be retrieved, user is not logged in %w", cmd.args[0], definederrors.ErrorUserNotFound)
	}

	state.Cfg.SetUser(cmd.args[0], w)
	fmt.Fprintf(w, "user %s is now logged in\n", cmd.args[0])
	return nil
}

func handlerRegisterUser(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	if len(cmd.args) != 1 {
		return fmt.Errorf("args passed into handlerCreateUser %v %w", cmd.args, definederrors.ErrorWrongNumArgs)
	}
	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}
	_, createErr := state.Db.CreateUser(context.Background(), params)
	if createErr != nil {
		isPQErr, pqErr, rawErr := errorutils.UnwrapPqErr(createErr)
		if isPQErr {
			if pqErr.Code == "23505" {
				fmt.Fprintf(w, "User %s already exists in database\n", cmd.args[0])
				return fmt.Errorf("user %s already exists %w", cmd.args[0], definederrors.ErrorUserAlreadyExists)
			}
		}
		return rawErr
	}
	fmt.Fprintf(w, "user %s has been added\n", cmd.args[0])
	state.Cfg.SetUser(cmd.args[0], w)
	return nil
}

func handlerTest(cmd enteredCommand, w io.Writer, state *database.State) (err error) {
	for _, arg := range cmd.args {
		fmt.Fprintln(w, arg)
	}
	return nil
}

// called at the main program, used to initialise the commandMap so that it can be written to
func InitCommands() (commandsPtr *commands) {
	returnedCommands := commands{}
	commandMap := make(map[string]handler)
	returnedCommands.commandMap = commandMap
	returnedCommands.registerAllHandlers()
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
		returnedCmd.name = strings.ToLower(args[1])
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
