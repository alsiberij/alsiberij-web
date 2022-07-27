package main

import (
	"auth/repository"
	"encoding/json"
	"os"
)

type (
	Config struct {
		AuthPG repository.PostgresConfig `json:"authPG"`
	}
)

func ReadConfig(filename string) (Config, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0444)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	var config Config
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
