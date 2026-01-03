package autostart

type App struct {
	Name string
	Exec []string
}

func (a *App) Enable() error {
	return a.enable()
}

func (a *App) Disable() error {
	return a.disable()
}

func (a *App) IsEnabled() bool {
	return a.isEnabled()
}
