package commands

import (
	"fmt"
	"io"
	"strings"

	definedErrors "github.com/sohWenMing/aggregator/defined_errors"
	"github.com/sohWenMing/aggregator/internal/config"
)

func GenerateCommand(args []string) (cmd Command, err error) {
	if len(args) < 3 {
		return Command{}, definedErrors.ErrNotEnoughArguments
	}
	cmd.name = args[1]
	cmd.args = args[2:]
	return cmd, nil
}

type Command struct {
	name string
	args []string
}

type Commands struct {
	commandMap map[string]func(c *config.Config, cmd Command, w io.Writer) error
}

func (c *Commands) Register(name string, fn func(c *config.Config, cmd Command, w io.Writer) error) {
	if c.commandMap == nil {
		c.commandMap = make(map[string]func(c *config.Config, cmd Command, w io.Writer) error)
	}
	c.commandMap[name] = fn
}
func (c *Commands) Run(cnf *config.Config, cmd Command, w io.Writer) error {
	if c.commandMap == nil {
		return definedErrors.ErrCommandMapNil
	}
	command, ok := c.commandMap[cmd.name]
	if !ok {
		return fmt.Errorf("%w command name passed in: %s", definedErrors.ErrCommandNotExist, cmd.name)
	}
	err := command(cnf, cmd, w)
	if err != nil {
		return err
	}
	return nil

}

func HandlerLogin(c *config.Config, cmd Command, w io.Writer) error {

	if len(cmd.args) == 0 {
		return definedErrors.ErrLoginHandlerZeroArgs
	}
	var userNameToSet string
	for _, arg := range cmd.args {
		userNameToSet += fmt.Sprintf("%s ", arg)
	}
	trimmedUserNameToSet := strings.TrimSpace(userNameToSet)
	if trimmedUserNameToSet == "" {
		return definedErrors.ErrUserNameNil
	}
	c.SetUser(trimmedUserNameToSet)
	fmt.Fprintln(w, "Username has been set to:", trimmedUserNameToSet)
	return nil
}
