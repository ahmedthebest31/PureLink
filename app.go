package main

import (
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
		},
		ActiveBlocklist: defaultRules(),
		configPath:      filepath.Join(appDir, configFileName),
		rulesPath:       filepath.Join(appDir, rulesFileName),
	}, nil
}

func (app *App) Load() error {
	if err := app.loadConfigFile(); err != nil {
		return fmt.Errorf("config load: %w", err)
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
