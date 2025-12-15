package main

import (
	"time"

	"github.com/atotto/clipboard"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
)

func main() {
	// Initialize the system tray application
	systray.Run(onReady, onExit)
}

func onReady() {
	// Setup tray icon and tooltip
	systray.SetTitle("PureLink")
	systray.SetTooltip("PureLink: Privacy & Clipboard Guard")

	// Menu items
	mStatus := systray.AddMenuItem("Status: Active", "The cleaner is running in the background")
	mPause := systray.AddMenuItem("Pause Cleaning", "Toggle the cleaning process")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit the application")

	// App state
	isRunning := true

	// Goroutine: Watch clipboard for changes
	go func() {
		lastText, _ := clipboard.ReadAll()
		for {
			// Check if cleaning is paused
			if !isRunning {
				time.Sleep(1 * time.Second)
				continue
			}

			text, err := clipboard.ReadAll()
			// Proceed only if there is new text and no read error
			if err == nil && text != "" && text != lastText {
				
				// Call the cleaning logic from cleaner.go
				cleaned := CleanText(text)

				if cleaned != text {
					// Update clipboard with cleaned text
					clipboard.WriteAll(cleaned)
					lastText = cleaned

					// Send desktop notification (Audio/Visual)
					// This is crucial for accessibility feedback.
					beeep.Notify("PureLink 🛡️", "Tracking parameters removed!", "")
				} else {
					lastText = text
				}
			}
			// Poll every 500ms to save CPU resources
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Goroutine: Handle menu clicks
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
			case <-mPause.ClickedCh:
				if isRunning {
					isRunning = false
					mPause.SetTitle("Resume Cleaning")
					mStatus.SetTitle("Status: Paused")
				} else {
					isRunning = true
					mPause.SetTitle("Pause Cleaning")
					mStatus.SetTitle("Status: Active")
				}
			}
		}
	}()
}

func onExit() {
	// Cleanup tasks can go here
}