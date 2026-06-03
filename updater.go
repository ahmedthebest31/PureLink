package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const updateURL = "https://raw.githubusercontent.com/ahmedthebest31/PureLink/main/rules.json"

func (app *App) UpdateFilters() error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(updateURL)
	if err != nil {
		return fmt.Errorf("network error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var newConfig RuleConfig
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&newConfig); err != nil {
		return fmt.Errorf("invalid rule format: %v", err)
	}

	if len(newConfig.Blocklist) == 0 {
		return fmt.Errorf("downloaded rules are empty")
	}

	if err := app.writeRulesFile(&newConfig); err != nil {
		return fmt.Errorf("failed to save rules: %v", err)
	}

	if err := app.loadRulesFile(); err != nil {
		return fmt.Errorf("failed to apply new rules: %v", err)
	}

	return nil
}
