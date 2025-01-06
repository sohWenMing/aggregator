package main

import (
	"os"

	_ "github.com/lib/pq"
	//this needs to be done to import the postgres driver
	"github.com/sohWenMing/aggregator/commands"
	"github.com/sohWenMing/aggregator/internal/config"
)

func main() {

	cmd, err := commands.ParseCommand(os.Args)
	if err != nil {
		os.Exit(1)
	}

	state, err := config.Read()
	if err != nil {
		os.Exit(1)
	}
	commandsPtr := commands.InitCommands()
	execErr := commandsPtr.ExecCommand(cmd, os.Stdout, state)
	if execErr != nil {
		os.Exit(1)
	}
}
