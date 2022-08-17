package main

import (
	"auth/database"
	"encoding/json"
	"os"
)

type (
	Config struct {
		Pgs database.PostgresConfig `json:"pgs-1"`
		Rds database.RedisConfig    `json:"rds-1"`
	}
)

func ReadConfig(filename string) (Config, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0444)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = json.NewDecoder(f).Decode(&config)
	_ = f.Close()
	return config, err
}
