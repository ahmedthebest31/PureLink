package main

import (
	"encoding/json"
	"os"
)

type RuleConfig struct {
	Blocklist []string `json:"blocklist"`
}

const rulesFileName = "rules.json"

func defaultRules() []string {
	return []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "si", "ref", "gclid", "gclsrc", "dclid",
		"msclkid", "mc_eid", "_ga", "yclid", "vero_conv", "vero_id", "wickedid",
		"share_id", "igshid",
	}
}

func (app *App) loadRulesFile() error {
	if _, err := os.Stat(app.rulesPath); os.IsNotExist(err) {
		initialConfig := RuleConfig{Blocklist: defaultRules()}
		if err := app.writeRulesFile(&initialConfig); err != nil {
			app.mu.Lock()
			app.ActiveBlocklist = defaultRules()
			app.mu.Unlock()
			return err
		}
		app.mu.Lock()
		app.ActiveBlocklist = defaultRules()
		app.mu.Unlock()
		return nil
	}

	file, err := os.Open(app.rulesPath)
	if err != nil {
		app.mu.Lock()
		app.ActiveBlocklist = defaultRules()
		app.mu.Unlock()
		return err
	}
	defer file.Close()

	var config RuleConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		app.mu.Lock()
		app.ActiveBlocklist = defaultRules()
		app.mu.Unlock()
		return err
	}

	app.mu.Lock()
	app.ActiveBlocklist = config.Blocklist
	app.mu.Unlock()
	return nil
}

func (app *App) writeRulesFile(config *RuleConfig) error {
	file, err := os.Create(app.rulesPath)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}
