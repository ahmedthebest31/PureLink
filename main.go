package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ahmedthebest31/PureLink/autostart"
	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
	"github.com/gofrs/flock"
	"github.com/sqweek/dialog"
)

//go:embed icon.png
var iconData []byte

func main() {
	lockFile := filepath.Join(os.TempDir(), "purelink.lock")
	fileLock := flock.New(lockFile)

	locked, err := fileLock.TryLock()
	if err != nil {
		dialog.Message("Error checking instance lock: %v", err).Title("Error").Error()
		return
	}

	if !locked {
		dialog.Message("PureLink is already running!").Title("Error").Error()
		return
	}
	defer fileLock.Unlock()

	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("PureLink")
	systray.SetTooltip("PureLink Privacy Guard")

	// Create App with resolved config/rules paths
	app, err := NewApp()
	if err != nil {
		dialog.Message("Failed to initialize PureLink: %v", err).Title("Fatal Error").Error()
		os.Exit(1)
	}

	if err := app.Load(); err != nil {
		fmt.Println("Warning loading app state:", err)
	}

	// Background auto-updater (every 7 days, respects shutdown)
	updaterCtx, updaterCancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(7 * 24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = app.UpdateFilters(updaterCtx)
			case <-updaterCtx.Done():
				return
			}
		}
	}()

	// Autostart Setup
	exe, _ := os.Executable()
	startupApp := &autostart.App{
		Name: "PureLink",
		Exec: []string{exe},
	}

	systray.AddMenuItem("Status: Active", "Protection is enabled").Disable()
	mCounter := systray.AddMenuItem(fmt.Sprintf("Cleaned: %d Links", app.Config.TotalCleaned), "Total items processed")
	systray.AddSeparator()

	// --- Recent History ---
	mHistory := systray.AddMenuItem("Recent History", "Last 5 cleaned links")
	var mHistoryItems []*systray.MenuItem

	for i := 0; i < 5; i++ {
		item := mHistory.AddSubMenuItem(fmt.Sprintf("Item %d", i), "")
		item.Hide()
		mHistoryItems = append(mHistoryItems, item)
	}

	updateHistoryMenu := func() {
		app.mu.RLock()
		defer app.mu.RUnlock()
		for i, item := range mHistoryItems {
			if i < len(app.Config.History) {
				title := app.Config.History[i]
				if len(title) > 50 {
					title = title[:47] + "..."
				}
				item.SetTitle(title)
				item.SetTooltip(app.Config.History[i])
				item.Show()
			} else {
				item.Hide()
			}
		}
	}

	updateHistoryMenu() // Initial load

	// Channel to aggregate history clicks
	historyClicked := make(chan int)
	for i, item := range mHistoryItems {
		go func(idx int, m *systray.MenuItem) {
			for range m.ClickedCh {
				historyClicked <- idx
			}
		}(i, item)
	}

	systray.AddSeparator()

	// --- Tools Submenu ---
	mTools := systray.AddMenuItem("Tools", "Manual Utilities")
	mUpdate := mTools.AddSubMenuItem("Update Filters Now", "Download latest tracking rules")
	tWhatsApp := mTools.AddSubMenuItem("Open WhatsApp", "Copy link and open WhatsApp")
	tTelegram := mTools.AddSubMenuItem("Open Telegram", "Copy link and open Telegram")
	tDecode64 := mTools.AddSubMenuItem("Decode Base64", "Decode Base64 string from clipboard")
	tEncode64 := mTools.AddSubMenuItem("Encode Base64", "Encode text to Base64")
	tUUID := mTools.AddSubMenuItem("Insert UUID", "Generate and copy a new UUID")

	systray.AddSeparator()
	mUnshorten := systray.AddMenuItemCheckbox("Unshorten Links", "Expand short URLs (Requires Internet)", app.Config.Unshorten)
	mWSL := systray.AddMenuItemCheckbox("WSL Path Mode", "Convert C:\\ to /mnt/c/ and fix slashes", app.Config.WSLMode)
	mCloudBoost := systray.AddMenuItemCheckbox("Direct Link", "Auto-convert Dropbox/Drive links", app.Config.DirectLink)
	mWhatsApp := systray.AddMenuItemCheckbox("Auto WhatsApp", "Auto-convert phone numbers to WhatsApp links", app.Config.WhatsAppLinks)
	mTelegram := systray.AddMenuItemCheckbox("Auto Telegram", "Auto-convert usernames to Telegram links", app.Config.TelegramLinks)
	mStartup := systray.AddMenuItemCheckbox("Run on Startup", "Launch PureLink when system starts", false)

	if startupApp.IsEnabled() {
		mStartup.Check()
	} else {
		mStartup.Uncheck()
	}

	systray.AddSeparator()
	mSound := systray.AddMenuItemCheckbox("Play Sound", "Beep when item is cleaned", app.Config.Sound)
	mPause := systray.AddMenuItem("Pause Protection", "Temporarily stop cleaning")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit PureLink")

	isRunning := true

	// --- Background Clipboard Watcher ---
	go func() {
		lastText, _ := clipboard.ReadAll()
		for {
			if !isRunning {
				time.Sleep(1 * time.Second)
				continue
			}

			text, err := clipboard.ReadAll()
			if err == nil && text != "" && text != lastText {
				cleaned := app.CleanText(text)

				if cleaned != text {
					clipboard.WriteAll(cleaned)
					lastText = cleaned

					app.mu.Lock()
					app.Config.TotalCleaned++

					var newHistory []string
					for _, item := range app.Config.History {
						if item != cleaned {
							newHistory = append(newHistory, item)
						}
					}
					app.Config.History = append([]string{cleaned}, newHistory...)
					if len(app.Config.History) > 5 {
						app.Config.History = app.Config.History[:5]
					}
					app.writeConfigFile()
					app.mu.Unlock()

					mCounter.SetTitle(fmt.Sprintf("Cleaned: %d Items", app.Config.TotalCleaned))
					updateHistoryMenu()

					app.mu.RLock()
					playSound := app.Config.Sound
					app.mu.RUnlock()
					if playSound {
						NotifyBeep()
					}
				} else {
					lastText = text
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// --- Event Handler ---
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				updaterCancel()
				systray.Quit()

			case <-mPause.ClickedCh:
				if isRunning {
					isRunning = false
					mPause.SetTitle("Resume Protection")
					systray.SetTooltip("PureLink (Paused)")
				} else {
					isRunning = true
					mPause.SetTitle("Pause Protection")
					systray.SetTooltip("PureLink (Active)")
				}

			case idx := <-historyClicked:
				app.mu.RLock()
				if idx < len(app.Config.History) {
					clipboard.WriteAll(app.Config.History[idx])
					if app.Config.Sound {
						NotifyBeep()
					}
				}
				app.mu.RUnlock()

			case <-mSound.ClickedCh:
				app.mu.Lock()
				if app.Config.Sound {
					app.Config.Sound = false
					mSound.Uncheck()
				} else {
					app.Config.Sound = true
					mSound.Check()
					NotifyBeep()
				}
				app.writeConfigFile()
				app.mu.Unlock()

			case <-mUnshorten.ClickedCh:
				app.mu.Lock()
				if app.Config.Unshorten {
					app.Config.Unshorten = false
					mUnshorten.Uncheck()
				} else {
					app.Config.Unshorten = true
					mUnshorten.Check()
					NotifyBeep()
				}
				app.writeConfigFile()
				app.mu.Unlock()

			case <-mWSL.ClickedCh:
				app.mu.Lock()
				if app.Config.WSLMode {
					app.Config.WSLMode = false
					mWSL.Uncheck()
				} else {
					app.Config.WSLMode = true
					mWSL.Check()
					NotifyBeep()
				}
				app.writeConfigFile()
				app.mu.Unlock()

			case <-mCloudBoost.ClickedCh:
				app.mu.Lock()
				if app.Config.DirectLink {
					app.Config.DirectLink = false
					mCloudBoost.Uncheck()
				} else {
					app.Config.DirectLink = true
					mCloudBoost.Check()
					NotifyBeep()
				}
				app.writeConfigFile()
				app.mu.Unlock()

			case <-mWhatsApp.ClickedCh:
				app.mu.Lock()
				if app.Config.WhatsAppLinks {
					app.Config.WhatsAppLinks = false
					mWhatsApp.Uncheck()
				} else {
					app.Config.WhatsAppLinks = true
					mWhatsApp.Check()
					NotifyBeep()
				}
				app.writeConfigFile()
				app.mu.Unlock()

			case <-mTelegram.ClickedCh:
				app.mu.Lock()
				if app.Config.TelegramLinks {
					app.Config.TelegramLinks = false
					mTelegram.Uncheck()
				} else {
					app.Config.TelegramLinks = true
					mTelegram.Check()
					NotifyBeep()
				}
				app.writeConfigFile()
				app.mu.Unlock()

			case <-mStartup.ClickedCh:
				if startupApp.IsEnabled() {
					if err := startupApp.Disable(); err != nil {
						dialog.Message("Failed to disable startup: %v", err).Title("Error").Error()
					} else {
						mStartup.Uncheck()
						dialog.Message("PureLink will no longer run on startup.").Title("Startup Disabled").Info()
						NotifyBeep()
					}
				} else {
					if err := startupApp.Enable(); err != nil {
						dialog.Message("Failed to enable startup: %v", err).Title("Error").Error()
					} else {
						mStartup.Check()
						dialog.Message("PureLink will now run automatically when you log in.").Title("Startup Enabled").Info()
						NotifyBeep()
					}
				}

			// --- Tools Actions ---
			case <-mUpdate.ClickedCh:
				err := app.UpdateFilters(context.Background())
				if err != nil {
					dialog.Message("Update failed: %v", err).Title("Error").Error()
				} else {
					dialog.Message("Filters updated successfully!").Title("Success").Info()
					NotifyBeep()
				}

			case <-tWhatsApp.ClickedCh:
				text, _ := clipboard.ReadAll()
				url, err := GetWhatsAppLink(text)
				if err == nil {
					clipboard.WriteAll(url)
					OpenBrowser(url)
					NotifyBeep()
				}

			case <-tTelegram.ClickedCh:
				text, _ := clipboard.ReadAll()
				url, err := GetTelegramLink(text)
				if err == nil {
					clipboard.WriteAll(url)
					OpenBrowser(url)
					NotifyBeep()
				}

			case <-tDecode64.ClickedCh:
				text, _ := clipboard.ReadAll()
				decoded, err := DecodeBase64(text)
				if err == nil && decoded != "" {
					clipboard.WriteAll(decoded)
					NotifyBeep()
				}

			case <-tEncode64.ClickedCh:
				text, _ := clipboard.ReadAll()
				if text != "" {
					encoded := EncodeBase64(text)
					clipboard.WriteAll(encoded)
					NotifyBeep()
				}

			case <-tUUID.ClickedCh:
				id := GenerateUUID()
				clipboard.WriteAll(id)
				NotifyBeep()
			}
		}
	}()
}

func onExit() {}
