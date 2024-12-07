package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sohWenMing/aggregator/commands"
	"github.com/sohWenMing/aggregator/internal/config"
)

func main() {
	initialConfig, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	//note: the first argument is actually the program that is passed in, may need to change this if creating an executable
	args := os.Args

	cmd, err := commands.GenerateCommand(args)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
		return
	}

	commandsRegister := commands.Commands{}
	commandsRegister.Register("login", commands.HandlerLogin)

	commandsRegister.Run(&initialConfig, cmd, os.Stdout)

	newConfig, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(newConfig)
}
