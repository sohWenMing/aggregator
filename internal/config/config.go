package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Config struct {
	Db_url            string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func (c *Config) SetUser(username string) error {

	c.Current_user_name = username
	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}
	writeErr := os.WriteFile(getConfigFilePath(), bytes, 0644)
	if writeErr != nil {
		return writeErr
	}
	return nil

}

func Read() (config Config, err error) {
	var returnedConfig Config

	file, err := os.ReadFile(getConfigFilePath())
	if err != nil {
		return returnedConfig, err
	}
	jsonErr := json.Unmarshal(file, &returnedConfig)
	if jsonErr != nil {
		return returnedConfig, jsonErr
	}
	return returnedConfig, nil
}

func getConfigFilePath() (filepath string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("error occured when acccessing homedir")
	}
	return fmt.Sprintf("%s/.gatorconfig.json", homeDir)
}
