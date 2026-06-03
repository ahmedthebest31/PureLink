package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ahmedthebest31/PureLink/autostart"
	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
	"github.com/gofrs/flock"
	"github.com/sqweek/dialog"
)

const AppVersion = "v3.0.0"

//go:embed icon.png
var iconData []byte

func lockFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "PureLink")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "purelink.lock"), nil
}

func cleanupOldBinaries() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	dir := filepath.Dir(exe)
	pattern := filepath.Join(dir, "*.old")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	for _, m := range matches {
		os.Remove(m)
	}
}

func main() {
	cleanupOldBinaries()

	lockPath, err := lockFilePath()
	if err != nil {
		dialog.Message("Cannot resolve lock path: %v", err).Title("Error").Error()
		return
	}
	fileLock := flock.New(lockPath)

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
	systray.SetTooltip("PureLink \u2014 Filtering Active")

	// Create App with resolved config/rules paths
	app, err := NewApp()
	if err != nil {
		dialog.Message("Failed to initialize PureLink: %v", err).Title("Fatal Error").Error()
		os.Exit(1)
	}

	if err := app.Load(); err != nil {
		fmt.Println("Warning loading app state:", err)
	}

	// Background version check on startup (silent, respects SkippedVersion)
	go app.checkForUpdates(false)

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

	// --- Social Tools Submenu ---
	mSocialTools := systray.AddMenuItem("Social Tools", "Manual social link utilities")
	tWhatsApp := mSocialTools.AddSubMenuItem("Convert to WhatsApp", "Read clipboard, format as WhatsApp link")
	tTelegram := mSocialTools.AddSubMenuItem("Convert to Telegram", "Read clipboard, format as Telegram link")

	// --- Updates & Rules Submenu ---
	mUpdatesRules := systray.AddMenuItem("Updates && Rules", "Update filters and check for app updates")
	mUpdate := mUpdatesRules.AddSubMenuItem("Update Filters Now", "Download latest tracking rules")
	mCheckUpdates := mUpdatesRules.AddSubMenuItem("Check for App Updates", "Check for new PureLink version")
	mViewReleaseNotes := mUpdatesRules.AddSubMenuItem("View Release Notes", "Open GitHub release page")

	// --- Advanced Settings Submenu ---
	mAdvanced := systray.AddMenuItem("Advanced Settings", "Toggles and data management")
	mYouTubeShorts := mAdvanced.AddSubMenuItemCheckbox("YouTube Shorts Fix", "Convert /shorts/ links to /watch", app.Config.YouTubeShorts)
	mCloudBoost := mAdvanced.AddSubMenuItemCheckbox("Cloud Boost", "Auto-convert Dropbox/Drive links", app.Config.CloudBoost)
	mEnableClipboardCommands := mAdvanced.AddSubMenuItemCheckbox("Clipboard Commands", "Enable !wa !tg !b64e !b64d !uuid prefixes", app.Config.EnableClipboardCommands)
	mClearData := mAdvanced.AddSubMenuItem("Clear Data && History", "Reset cleaned count and history")

	// --- Developer Tools Submenu ---
	mDevTools := systray.AddMenuItem("Developer Tools", "Utilities for developers")
	mWSL := mDevTools.AddSubMenuItemCheckbox("WSL Path Mode", "Persistently convert C:\\ to /mnt/c/", app.Config.WSLMode)
	mConvertWSL := mDevTools.AddSubMenuItem("Convert Clipboard to WSL Path", "One-off WSL path conversion")
	tEncode64 := mDevTools.AddSubMenuItem("Base64 Encode", "Encode clipboard text to Base64")
	tDecode64 := mDevTools.AddSubMenuItem("Base64 Decode", "Decode Base64 string from clipboard")
	tUUID := mDevTools.AddSubMenuItem("Generate UUID", "Generate and copy a new UUID")
	tURLEncode := mDevTools.AddSubMenuItem("URL Encode", "Percent-encode clipboard text")
	tURLDecode := mDevTools.AddSubMenuItem("URL Decode", "Decode percent-encoded clipboard text")
	tFormatJSON := mDevTools.AddSubMenuItem("Format JSON", "Pretty-print JSON from clipboard")
	tUnixTimestamp := mDevTools.AddSubMenuItem("Unix Timestamp", "Generate current epoch seconds")
	tDecodeJWT := mDevTools.AddSubMenuItem("Decode JWT Payload", "Extract and pretty-print JWT payload")

	systray.AddSeparator()
	mPause := systray.AddMenuItem("Pause Protection", "Temporarily stop cleaning")
	mSound := systray.AddMenuItemCheckbox("Play Sound", "Beep when item is cleaned", app.Config.Sound)
	mUnshorten := systray.AddMenuItemCheckbox("Unshorten Links", "Expand short URLs (Requires Internet)", app.Config.Unshorten)
	mStartup := systray.AddMenuItemCheckbox("Run on Startup", "Launch PureLink when system starts", false)

	if startupApp.IsEnabled() {
		mStartup.Check()
	} else {
		mStartup.Uncheck()
	}

	systray.AddSeparator()
	mOpenConfig := systray.AddMenuItem("Open Config File", "Open purelink_config.json in default editor")
	mQuit := systray.AddMenuItem("Quit", "Exit PureLink")

	isRunning := true

	// --- Background Clipboard Watcher ---
	go func() {
		lastText, _ := clipboard.ReadAll()
		lastSeen := time.Now()
		for {
			if !isRunning {
				time.Sleep(1 * time.Second)
				continue
			}

			text, err := clipboard.ReadAll()
			now := time.Now()
			if err == nil && text != "" && (text != lastText || now.Sub(lastSeen) >= time.Second) {
				lastSeen = now
				cleaned := app.CleanText(text)

				if cleaned != text {
					clipboard.WriteAll(cleaned)
					lastText = cleaned
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
			case <-mOpenConfig.ClickedCh:
				switch runtime.GOOS {
				case "windows":
					exec.Command("cmd", "/c", "start", "", app.configPath).Start()
				case "darwin":
					exec.Command("open", app.configPath).Start()
				default:
					exec.Command("xdg-open", app.configPath).Start()
				}

			case <-mQuit.ClickedCh:
				updaterCancel()
				systray.Quit()

			case <-mPause.ClickedCh:
				if isRunning {
					isRunning = false
					mPause.SetTitle("Resume Protection")
					systray.SetTooltip("PureLink \u2014 Filtering Paused")
				} else {
					isRunning = true
					mPause.SetTitle("Pause Protection")
					systray.SetTooltip("PureLink \u2014 Filtering Active")
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

			// --- Advanced Settings ---
			case <-mYouTubeShorts.ClickedCh:
				app.mu.Lock()
				if app.Config.YouTubeShorts {
					app.Config.YouTubeShorts = false
					mYouTubeShorts.Uncheck()
				} else {
					app.Config.YouTubeShorts = true
					mYouTubeShorts.Check()
				}
				app.writeConfigFile()
				app.mu.Unlock()

			case <-mCloudBoost.ClickedCh:
				app.mu.Lock()
				if app.Config.CloudBoost {
					app.Config.CloudBoost = false
					mCloudBoost.Uncheck()
				} else {
					app.Config.CloudBoost = true
					mCloudBoost.Check()
					NotifyBeep()
				}
				app.writeConfigFile()
				app.mu.Unlock()

			case <-mEnableClipboardCommands.ClickedCh:
				app.mu.Lock()
				if app.Config.EnableClipboardCommands {
					app.Config.EnableClipboardCommands = false
					mEnableClipboardCommands.Uncheck()
				} else {
					app.Config.EnableClipboardCommands = true
					mEnableClipboardCommands.Check()
				}
				app.writeConfigFile()
				app.mu.Unlock()

			case <-mClearData.ClickedCh:
				app.mu.Lock()
				app.Config.TotalCleaned = 0
				app.Config.History = nil
				app.writeConfigFile()
				app.mu.Unlock()
				updateHistoryMenu()
				dialog.Message("Cleaned count and history have been reset.").Title("Data Cleared").Info()

			// --- Main Toggles ---
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

			// --- Developer Tools Actions ---
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

			case <-mConvertWSL.ClickedCh:
				text, _ := clipboard.ReadAll()
				if text != "" {
					result := ConvertToWSL(text)
					if result != text {
						clipboard.WriteAll(result)
						NotifyBeep()
					}
				}

			case <-tEncode64.ClickedCh:
				text, _ := clipboard.ReadAll()
				if text != "" {
					clipboard.WriteAll(EncodeBase64(text))
					NotifyBeep()
				}

			case <-tDecode64.ClickedCh:
				text, _ := clipboard.ReadAll()
				decoded, err := DecodeBase64(text)
				if err == nil && decoded != "" {
					clipboard.WriteAll(decoded)
					NotifyBeep()
				} else {
					NotifyBeep()
				}

			case <-tUUID.ClickedCh:
				clipboard.WriteAll(GenerateUUID())
				NotifyBeep()

			case <-tURLEncode.ClickedCh:
				text, _ := clipboard.ReadAll()
				if text != "" {
					clipboard.WriteAll(URLEncode(text))
					NotifyBeep()
				}

			case <-tURLDecode.ClickedCh:
				text, _ := clipboard.ReadAll()
				decoded, err := URLDecode(text)
				if err == nil {
					clipboard.WriteAll(decoded)
					NotifyBeep()
				} else {
					NotifyBeep()
				}

			case <-tFormatJSON.ClickedCh:
				text, _ := clipboard.ReadAll()
				formatted, err := FormatJSON(text)
				if err == nil {
					clipboard.WriteAll(formatted)
					NotifyBeep()
				} else {
					NotifyBeep()
				}

			case <-tUnixTimestamp.ClickedCh:
				clipboard.WriteAll(UnixTimestamp())
				NotifyBeep()

			case <-tDecodeJWT.ClickedCh:
				text, _ := clipboard.ReadAll()
				payload, err := DecodeJWTPayload(text)
				if err == nil {
					clipboard.WriteAll(payload)
					NotifyBeep()
				} else {
					NotifyBeep()
				}

			// --- Updates & Rules Actions ---
			case <-mUpdate.ClickedCh:
				err := app.UpdateFilters(context.Background())
				if err != nil {
					dialog.Message("Update failed: %v", err).Title("Error").Error()
				} else {
					dialog.Message("Filters updated successfully!").Title("Success").Info()
					NotifyBeep()
				}

			case <-mCheckUpdates.ClickedCh:
				app.checkForUpdates(true)

			case <-mViewReleaseNotes.ClickedCh:
				switch runtime.GOOS {
				case "windows":
					exec.Command("rundll32", "url.dll,FileProtocolHandler", releaseNotesURL).Start()
				case "darwin":
					exec.Command("open", releaseNotesURL).Start()
				default:
					exec.Command("xdg-open", releaseNotesURL).Start()
				}

			// --- Social Tools Actions ---
			case <-tWhatsApp.ClickedCh:
				text, _ := clipboard.ReadAll()
				url, err := GetWhatsAppLink(text)
				if err == nil {
					clipboard.WriteAll(url)
					NotifyBeep()
				}

			case <-tTelegram.ClickedCh:
				text, _ := clipboard.ReadAll()
				url, err := GetTelegramLink(text)
				if err == nil {
					clipboard.WriteAll(url)
					NotifyBeep()
				}
			}
		}
	}()
}

func onExit() {}
