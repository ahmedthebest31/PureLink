//go:build !windows

package autostart

import (
	"os"
	"path/filepath"
	"strings"
)

func (a *App) enable() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(configDir, "autostart")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(dir, a.Name+".desktop")

	var sb strings.Builder
	sb.WriteString("[Desktop Entry]\n")
	sb.WriteString("Type=Application\n")
	sb.WriteString("Name=")
	sb.WriteString(a.Name)
	sb.WriteString("\nExec=")
	for i, part := range a.Exec {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(part)
	}
	sb.WriteString("\nHidden=false\nNoDisplay=false\nX-GNOME-Autostart-enabled=true\n")

	return os.WriteFile(filePath, []byte(sb.String()), 0644)
}

func (a *App) disable() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil
	}

	filePath := filepath.Join(configDir, "autostart", a.Name+".desktop")
	if err := os.Remove(filePath); os.IsNotExist(err) {
		return nil
	} else {
		return err
	}
}

func (a *App) isEnabled() bool {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return false
	}

	filePath := filepath.Join(configDir, "autostart", a.Name+".desktop")
	_, err = os.Stat(filePath)
	return err == nil
}
