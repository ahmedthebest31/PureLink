package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ahmedthebest31/PureLink/autostart"
	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
	"github.com/gofrs/flock"
	"github.com/sqweek/dialog"
)

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
	systray.SetTitle("PureLink")
	systray.SetTooltip("PureLink Privacy Guard")

		// Load Config

		cfg, err := LoadConfig()

		if err != nil {

			fmt.Println("Error loading config:", err)

		}

		var cfgMutex sync.Mutex // Protects concurrent access to cfg

	

		// Load Rules

		if err := LoadRules(); err != nil {

			fmt.Println("Error loading rules:", err)

		}

	

		// Autostart Setup

		exe, _ := os.Executable()

		app := &autostart.App{

		Name: "PureLink",

		Exec: []string{exe},

	}

	

		systray.AddMenuItem("Status: Active", "Protection is enabled").Disable()

		mCounter := systray.AddMenuItem(fmt.Sprintf("Cleaned: %d Links", cfg.TotalCleaned), "Total items processed")

	

		systray.AddSeparator()

	

		// --- Recent History ---

		mHistory := systray.AddMenuItem("Recent History", "Last 5 cleaned links")

		var mHistoryItems []*systray.MenuItem

		for i := 0; i < 5; i++ {

			item := mHistory.AddSubMenuItem(fmt.Sprintf("Item %d", i), "")

			item.Hide()

			mHistoryItems = append(mHistoryItems, item)

		}

	

		// Helper to update history menu

		updateHistoryMenu := func() {

			cfgMutex.Lock()

			defer cfgMutex.Unlock()

			for i, item := range mHistoryItems {

				if i < len(cfg.History) {

					title := cfg.History[i]

					if len(title) > 50 {

						title = title[:47] + "..."

					}

					item.SetTitle(title)

					item.SetTooltip(cfg.History[i])

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

	

		// Added items directly without separators inside the submenu (Library limitation)

		mUpdate := mTools.AddSubMenuItem("Check for Filter Updates", "Download latest tracking rules")

	

		tWhatsApp := mTools.AddSubMenuItem("Open WhatsApp", "Copy link and open WhatsApp")

		tTelegram := mTools.AddSubMenuItem("Open Telegram", "Copy link and open Telegram")

		tDecode64 := mTools.AddSubMenuItem("Decode Base64", "Decode Base64 string from clipboard")

		tEncode64 := mTools.AddSubMenuItem("Encode Base64", "Encode text to Base64")

		tUUID := mTools.AddSubMenuItem("Insert UUID", "Generate and copy a new UUID")

	

		systray.AddSeparator()

		mUnshorten := systray.AddMenuItemCheckbox("Unshorten Links", "Expand short URLs (Requires Internet)", cfg.Unshorten)

		mWSL := systray.AddMenuItemCheckbox("WSL Path Mode", "Convert C:\\ to /mnt/c/ and fix slashes", cfg.WSLMode)

		mCloudBoost := systray.AddMenuItemCheckbox("Direct Link", "Auto-convert Dropbox/Drive links", cfg.DirectLink)

		mStartup := systray.AddMenuItemCheckbox("Run on Startup", "Launch PureLink when system starts", false)

	

		if app.IsEnabled() {

			mStartup.Check()

		} else {

			mStartup.Uncheck()

		}

	

		systray.AddSeparator()

	

		mSound := systray.AddMenuItemCheckbox("Play Sound", "Beep when item is cleaned", cfg.Sound)

		mPause := systray.AddMenuItem("Pause Protection", "Temporarily stop cleaning")

	

		systray.AddSeparator()

		mQuit := systray.AddMenuItem("Quit", "Exit PureLink")

	

		// Local runtime state (not persisted)

		isRunning := true

	

		// --- Background Watcher ---

		go func() {

			lastText, _ := clipboard.ReadAll()

			for {

				if !isRunning {

					time.Sleep(1 * time.Second)

					continue

				}

	

				text, err := clipboard.ReadAll()

				if err == nil && text != "" && text != lastText {

	

					cfgMutex.Lock()

					cleaned := CleanText(text, cfg.Unshorten, cfg.WSLMode, cfg.DirectLink)

					cfgMutex.Unlock()

	

					if cleaned != text {

						clipboard.WriteAll(cleaned)

						lastText = cleaned

	

						cfgMutex.Lock()

						cfg.TotalCleaned++

						// Update History

						cfg.History = append([]string{cleaned}, cfg.History...)

						if len(cfg.History) > 5 {

							cfg.History = cfg.History[:5]

						}

						SaveConfig(cfg) // Auto-save on count change

						cfgMutex.Unlock()

	

						mCounter.SetTitle(fmt.Sprintf("Cleaned: %d Items", cfg.TotalCleaned))

						updateHistoryMenu()

	

						cfgMutex.Lock()

						playSound := cfg.Sound

						cfgMutex.Unlock()

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

					cfgMutex.Lock()

					if idx < len(cfg.History) {

						clipboard.WriteAll(cfg.History[idx])

						if cfg.Sound {

							NotifyBeep()

						}

					}

					cfgMutex.Unlock()

	

				case <-mSound.ClickedCh:

					cfgMutex.Lock()

					if cfg.Sound {

						cfg.Sound = false

						mSound.Uncheck()

					} else {

						cfg.Sound = true

						mSound.Check()

						NotifyBeep()

					}

					SaveConfig(cfg)

					cfgMutex.Unlock()

	

				case <-mUnshorten.ClickedCh:

					cfgMutex.Lock()

					if cfg.Unshorten {

						cfg.Unshorten = false

						mUnshorten.Uncheck()

					} else {

						cfg.Unshorten = true

						mUnshorten.Check()

						NotifyBeep()

					}

					SaveConfig(cfg)

					cfgMutex.Unlock()

	

				case <-mWSL.ClickedCh:

					cfgMutex.Lock()

					if cfg.WSLMode {

						cfg.WSLMode = false

						mWSL.Uncheck()

					} else {

						cfg.WSLMode = true

						mWSL.Check()

						NotifyBeep()

					}

					SaveConfig(cfg)

					cfgMutex.Unlock()

	

				case <-mCloudBoost.ClickedCh:

					cfgMutex.Lock()

					if cfg.DirectLink {

						cfg.DirectLink = false

						mCloudBoost.Uncheck()

					} else {

						cfg.DirectLink = true

						mCloudBoost.Check()

						NotifyBeep()

					}

					SaveConfig(cfg)

					cfgMutex.Unlock()

	

				case <-mStartup.ClickedCh:

					if app.IsEnabled() {

						if err := app.Disable(); err != nil {

							dialog.Message("Failed to disable startup: %v", err).Title("Error").Error()

						} else {

							mStartup.Uncheck()

							dialog.Message("PureLink will no longer run on startup.").Title("Startup Disabled").Info()

							NotifyBeep()

						}

					} else {

						if err := app.Enable(); err != nil {

							dialog.Message("Failed to enable startup: %v", err).Title("Error").Error()

						} else {

							mStartup.Check()

							dialog.Message("PureLink will now run automatically when you log in.").Title("Startup Enabled").Info()

							NotifyBeep()

						}

					}

	

				// --- Tools Actions ---

	

				case <-mUpdate.ClickedCh:

					err := UpdateFilters()

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
	} // Close onReady

func onExit() {}
