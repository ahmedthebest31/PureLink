package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
)

// Windows API definitions for native interactions (Sound, MessageBox, Mutex)
var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procCreateMutex  = kernel32.NewProc("CreateMutexW")
	procMessageBox   = user32.NewProc("MessageBoxW")
	procMessageBeep  = user32.NewProc("MessageBeep")
)

const (
	ERROR_ALREADY_EXISTS = 183
	MB_ICONERROR         = 0x00000010
	MB_OK                = 0x00000000
)

func main() {
	// 1. Single Instance Check using Named Mutex
	// This prevents the application from running multiple times.
	_, _, err := procCreateMutex.Call(
		0,
		1,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Global\\PureLinkInstanceLock"))),
	)

	// If the mutex already exists, it means an instance is running.
	if err.(syscall.Errno) == ERROR_ALREADY_EXISTS {
		showNativeError("PureLink is already running!", "Error")
		return // Terminate the new instance
	}

	// 2. Initialize the System Tray UI
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("PureLink")
	systray.SetTooltip("PureLink Privacy Guard")

	// --- Menu Layout ---

	// 1. Status Indicator (Disabled item, just for display like AdGuard)
	systray.AddMenuItem("Status: Active", "Protection is enabled").Disable()

	// 2. Statistics Counter
	mCounter := systray.AddMenuItem("Cleaned: 0 Links", "Total links cleaned this session")

	systray.AddSeparator()

	// 3. Sound Control
	mSound := systray.AddMenuItemCheckbox("Play Sound", "Beep when link is cleaned", true)

	// 4. Pause/Resume Control
	mPause := systray.AddMenuItem("Pause Protection", "Temporarily stop cleaning")

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit PureLink")

	// Application State
	isRunning := true
	isSoundEnabled := true
	cleanedCount := 0

	// --- Clipboard Watcher (Background Goroutine) ---
	go func() {
		lastText, _ := clipboard.ReadAll()
		for {
			// If paused, skip processing to save resources
			if !isRunning {
				time.Sleep(1 * time.Second)
				continue
			}

			text, err := clipboard.ReadAll()
			// Process only if text has changed and is valid
			if err == nil && text != "" && text != lastText {

				// Call the cleaning logic from cleaner.go
				cleaned := CleanText(text)

				if cleaned != text {
					// Update clipboard with cleaned URL
					clipboard.WriteAll(cleaned)
					lastText = cleaned

					// Update UI Counter
					cleanedCount++
					mCounter.SetTitle(fmt.Sprintf("Cleaned: %d Links", cleanedCount))

					// Provide Audio Feedback
					if isSoundEnabled {
						nativeBeep()
					}
				} else {
					lastText = text
				}
			}
			// Poll interval (500ms is a good balance for responsiveness/CPU)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// --- Event Loop (Handle Menu Clicks) ---
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
				if isSoundEnabled {
					isSoundEnabled = false
					mSound.Uncheck()
				} else {
					isSoundEnabled = true
					mSound.Check()
					nativeBeep() // Test sound on enable
				}
			}
		}
	}()
}

func onExit() {
	// Cleanup tasks (if any)
}

// --- Native Helper Functions ---

// showNativeError displays a native Windows MessageBox.
// Essential for accessibility as screen readers announce standard dialogs immediately.
func showNativeError(text, title string) {
	procMessageBox.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		MB_ICONERROR|MB_OK,
	)
}

// nativeBeep plays the system default sound using Win32 API.
// Lightweight alternative to external audio libraries.
func nativeBeep() {
	procMessageBeep.Call(0xFFFFFFFF)
}