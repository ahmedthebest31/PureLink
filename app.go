package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type App struct {
	Config         *Config
	ActiveBlocklist []string
	mu             sync.RWMutex
	configPath     string
	rulesPath      string
}

func NewApp() (*App, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("cannot resolve user config directory: %w", err)
	}
	appDir := filepath.Join(configDir, "PureLink")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create config directory %s: %w", appDir, err)
	}
	return &App{
		Config: &Config{
			DirectLink: true,
			Sound:      true,
			History:    []string{},
			IgnoreList: []string{"localhost", "127.0.0.1"},
		},
		ActiveBlocklist: defaultRules(),
		configPath:      filepath.Join(appDir, configFileName),
		rulesPath:       filepath.Join(appDir, rulesFileName),
	}, nil
}

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

func (app *App) Load() error {
	if err := app.loadConfigFile(); err != nil {
		return fmt.Errorf("config load: %w", err)
	}
	if app.Config.IgnoreList == nil {
		app.Config.IgnoreList = []string{"localhost", "127.0.0.1"}
	}
	if err := app.loadRulesFile(); err != nil {
		return fmt.Errorf("rules load: %w", err)
	}
	return nil
}

func (app *App) Save() error {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.writeConfigFile()
}

func (app *App) GetRules() []string {
	app.mu.RLock()
	defer app.mu.RUnlock()
	result := make([]string, len(app.ActiveBlocklist))
	copy(result, app.ActiveBlocklist)
	return result
}
