package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	definederrors "github.com/sohWenMing/aggregator/defined_errors"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
	CurrentUser     struct {
		ID   uuid.UUID
		Name string
	}
}

func (c *Config) SetUser(username string, w io.Writer) (err error) {
	if len(username) == 0 {
		return definederrors.ErrorInput
	}
	c.CurrentUserName = username
	bytesToWrite, err := json.Marshal(c)
	if err != nil {
		return err
	}
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	writeErr := os.WriteFile(configFilePath, bytesToWrite, 0644)
	if writeErr != nil {
		fmt.Printf("write err happened, %v\n", err)
		return writeErr
	}
	return nil
}

func Read() (config *Config, err error) {
	returnedConfig := Config{}
	file, err := openConfigFile()
	if err != nil {
		return &returnedConfig, err
	}
	defer file.Close()

	fileBytes, fileReadErr := io.ReadAll(file)
	if fileReadErr != nil {
		return &returnedConfig, fileReadErr
	}

	unMarshalErr := json.Unmarshal(fileBytes, &returnedConfig)
	if unMarshalErr != nil {
		return &returnedConfig, unMarshalErr
	}

	return &returnedConfig, nil
}

func getConfigFilePath() (filePath string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configPath := homeDir + "/" + ".gatorconfig.json"
	return configPath, nil
}

func openConfigFile() (file *os.File, err error) {
	configPath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}
	file, fileOpenErr := os.Open(configPath)
	if fileOpenErr != nil {
		return nil, fileOpenErr
	}
	return file, nil
}
