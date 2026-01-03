//go:build !windows

package autostart

func (a *App) enable() error {
	return nil
}

func (a *App) disable() error {
	return nil
}

func (a *App) isEnabled() bool {
	return false
}
