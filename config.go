package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Unshorten      bool     `json:"unshorten"`
	WSLMode        bool     `json:"wsl_mode"`
	CloudBoost     bool     `json:"cloud_boost"`
	YouTubeShorts  bool     `json:"youtube_shorts"`
	Sound          bool     `json:"sound"`
	TotalCleaned   int      `json:"total_cleaned"`
	History        []string `json:"history"`
	IgnoreList            []string `json:"ignore_list"`
	SkippedVersion        string   `json:"skipped_version"`
	EnableClipboardCommands bool   `json:"enable_clipboard_commands"`
}

const configFileName = "purelink_config.json"

func (app *App) loadConfigFile() error {
	file, err := os.Open(app.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(app.Config); err != nil {
		return nil
	}
	return nil
}

func (app *App) writeConfigFile() error {
	file, err := os.Create(app.configPath)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(app.Config)
}
