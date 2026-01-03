package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Unshorten    bool `json:"unshorten"`
	WSLMode      bool `json:"wsl_mode"`
	DirectLink   bool `json:"direct_link"`
	Sound        bool `json:"sound"`
	TotalCleaned int  `json:"total_cleaned"`
}

const configFileName = "purelink_config.json"

func LoadConfig() (*Config, error) {
	// Default config
	cfg := &Config{
		Unshorten:    false,
		WSLMode:      false,
		DirectLink:   true,
		Sound:        true,
		TotalCleaned: 0,
	}

	file, err := os.Open(configFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return default if file doesn't exist
		}
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(cfg)
	if err != nil {
		return cfg, nil // Return default/partial on error, or handle differently
	}

	return cfg, nil
}

func SaveConfig(cfg *Config) error {
	file, err := os.Create(configFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}
