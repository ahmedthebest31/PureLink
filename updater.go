package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const updateURL = "https://raw.githubusercontent.com/ahmedthebest31/PureLink/main/rules.json"

func (app *App) UpdateFilters(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, updateURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var newConfig RuleConfig
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&newConfig); err != nil {
		return fmt.Errorf("invalid rule format: %w", err)
	}

	if len(newConfig.Blocklist) == 0 {
		return fmt.Errorf("downloaded rules are empty")
	}

	app.mu.Lock()
	app.ActiveBlocklist = newConfig.Blocklist
	err = app.writeRulesFile(&newConfig)
	app.mu.Unlock()

	if err != nil {
		return fmt.Errorf("failed to persist rules: %w", err)
	}

	return nil
}
