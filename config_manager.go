package main

import (
	"encoding/json"
	"os"
	"sync"
)

// RuleConfig defines the structure of the rules.json file
type RuleConfig struct {
	Blocklist []string `json:"blocklist"`
}

var (
	// ActiveBlocklist holds the currently loaded tracking parameters
	ActiveBlocklist []string
	// BlocklistLock ensures safe concurrent access to ActiveBlocklist
	BlocklistLock sync.RWMutex
)

const rulesFileName = "rules.json"

// LoadRules reads the rules.json file. If it doesn't exist, it creates it with defaults.
func LoadRules() error {
	BlocklistLock.Lock()
	defer BlocklistLock.Unlock()

	// Default rules
	defaultRules := []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "si", "ref", "gclid", "gclsrc", "dclid",
		"msclkid", "mc_eid", "_ga", "yclid", "vero_conv", "vero_id", "wickedid",
		"share_id", "igshid",
	}

	// Check if file exists
	if _, err := os.Stat(rulesFileName); os.IsNotExist(err) {
		// Create default file
		initialConfig := RuleConfig{Blocklist: defaultRules}
		if err := saveRulesToFile(&initialConfig); err != nil {
			// If we can't save, just load defaults into memory
			ActiveBlocklist = defaultRules
			return err
		}
	}

	// Read file
	file, err := os.Open(rulesFileName)
	if err != nil {
		// Fallback to defaults if read fails
		ActiveBlocklist = defaultRules
		return err
	}
	defer file.Close()

	var config RuleConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		ActiveBlocklist = defaultRules
		return err
	}

	ActiveBlocklist = config.Blocklist
	return nil
}

func saveRulesToFile(config *RuleConfig) error {
	file, err := os.Create(rulesFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}
