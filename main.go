package main

import (
	"os"

	"github.com/sohWenMing/aggregator/commands"
	"github.com/sohWenMing/aggregator/internal/database"
)

func main() {
	args := os.Args
	//gets the arguments entered when user fires off the program
	state, err := database.CreateDBConnection()
	if err != nil {
		os.Exit(1)
	}
	commandsPtr := commands.InitCommands()
	writer := os.Stdout
	cmd, err := commands.ParseCommand(args)
	if err != nil {
		os.Exit(1)
	}
	execCommandErr := commandsPtr.ExecCommand(cmd, writer, state)
	if execCommandErr != nil {
		os.Exit(1)
	}

}
