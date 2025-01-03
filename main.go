package main

import (
	"fmt"
	"os"

	"github.com/sohWenMing/aggregator/internal/config"
)

func main() {
	state, err := config.Read()
	if err != nil {
		os.Exit(1)
	}
	setUserErr := state.SetUser("nindgabeet", os.Stdout)
	if setUserErr != nil {
		os.Exit(1)
	}
	newState, err := config.Read()
	if err != nil {
		os.Exit(1)
	}
	fmt.Printf("%v", newState)

}
