package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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
		// cfg is already initialized with defaults even on error in LoadConfig
	}

	systray.AddMenuItem("Status: Active", "Protection is enabled").Disable()
	mCounter := systray.AddMenuItem(fmt.Sprintf("Cleaned: %d Links", cfg.TotalCleaned), "Total items processed")
	
systray.AddSeparator()

	// --- Tools Submenu ---
	mTools := systray.AddMenuItem("Tools", "Manual Utilities")
	
	// Added items directly without separators inside the submenu (Library limitation)
	tWhatsApp := mTools.AddSubMenuItem("Open WhatsApp", "Copy link and open WhatsApp")
	tTelegram := mTools.AddSubMenuItem("Open Telegram", "Copy link and open Telegram")
	tDecode64 := mTools.AddSubMenuItem("Decode Base64", "Decode Base64 string from clipboard")
	tEncode64 := mTools.AddSubMenuItem("Encode Base64", "Encode text to Base64")
	tUUID := mTools.AddSubMenuItem("Insert UUID", "Generate and copy a new UUID")

	systray.AddSeparator()

	mUnshorten := systray.AddMenuItemCheckbox("Unshorten Links", "Expand short URLs (Requires Internet)", cfg.Unshorten)
	mWSL := systray.AddMenuItemCheckbox("WSL Path Mode", "Convert C:\\ to /mnt/c/ and fix slashes", cfg.WSLMode)
	mCloudBoost := systray.AddMenuItemCheckbox("Direct Link", "Auto-convert Dropbox/Drive links", cfg.DirectLink)

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
				
				cleaned := CleanText(text, cfg.Unshorten, cfg.WSLMode, cfg.DirectLink)

				if cleaned != text {
					clipboard.WriteAll(cleaned)
					lastText = cleaned
					
					cfg.TotalCleaned++
					mCounter.SetTitle(fmt.Sprintf("Cleaned: %d Items", cfg.TotalCleaned))
					SaveConfig(cfg) // Auto-save on count change

					if cfg.Sound {
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

			case <-mSound.ClickedCh:
				if cfg.Sound {
					cfg.Sound = false
					mSound.Uncheck()
				} else {
					cfg.Sound = true
					mSound.Check()
					NotifyBeep()
				}
				SaveConfig(cfg)

			case <-mUnshorten.ClickedCh:
				if cfg.Unshorten {
					cfg.Unshorten = false
					mUnshorten.Uncheck()
				} else {
					cfg.Unshorten = true
					mUnshorten.Check()
					NotifyBeep()
				}
				SaveConfig(cfg)

			case <-mWSL.ClickedCh:
				if cfg.WSLMode {
					cfg.WSLMode = false
					mWSL.Uncheck()
				} else {
					cfg.WSLMode = true
					mWSL.Check()
					NotifyBeep()
				}
				SaveConfig(cfg)

			case <-mCloudBoost.ClickedCh:
				if cfg.DirectLink {
					cfg.DirectLink = false
					mCloudBoost.Uncheck()
				} else {
					cfg.DirectLink = true
					mCloudBoost.Check()
					NotifyBeep()
				}
				SaveConfig(cfg)
			
			// --- Tools Actions ---
			
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
