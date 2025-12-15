package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
)

// Windows API definitions
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
	// Single Instance Lock
	_, _, err := procCreateMutex.Call(
		0,
		1,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Global\\PureLinkInstanceLock"))),
	)

	if err.(syscall.Errno) == ERROR_ALREADY_EXISTS {
		showNativeError("PureLink is already running!", "Error")
		return
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("PureLink")
	systray.SetTooltip("PureLink Privacy Guard")

	// --- Menu Layout ---
	systray.AddMenuItem("Status: Active", "Protection is enabled").Disable()
	mCounter := systray.AddMenuItem("Cleaned: 0 Links", "Total links cleaned this session")
	
	systray.AddSeparator()

	// New Feature: Unshorten Checkbox
	// Default is false (OFF) to save internet, user can enable it.
	mUnshorten := systray.AddMenuItemCheckbox("Unshorten Links", "Expand bit.ly and t.co links (Requires Internet)", false)

	mSound := systray.AddMenuItemCheckbox("Play Sound", "Beep when link is cleaned", true)
	mPause := systray.AddMenuItem("Pause Protection", "Temporarily stop cleaning")

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit PureLink")

	// App State
	isRunning := true
	isSoundEnabled := true
	isUnshortenEnabled := false
	cleanedCount := 0

	// --- Background Worker ---
	go func() {
		lastText, _ := clipboard.ReadAll()
		for {
			if !isRunning {
				time.Sleep(1 * time.Second)
				continue
			}

			text, err := clipboard.ReadAll()
			if err == nil && text != "" && text != lastText {
				
				// Pass the unshorten flag to the cleaner
				cleaned := CleanText(text, isUnshortenEnabled)

				if cleaned != text {
					clipboard.WriteAll(cleaned)
					lastText = cleaned
					
					cleanedCount++
					mCounter.SetTitle(fmt.Sprintf("Cleaned: %d Links", cleanedCount))

					if isSoundEnabled {
						nativeBeep()
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
				if isSoundEnabled {
					isSoundEnabled = false
					mSound.Uncheck()
				} else {
					isSoundEnabled = true
					mSound.Check()
					nativeBeep()
				}

			case <-mUnshorten.ClickedCh:
				if isUnshortenEnabled {
					isUnshortenEnabled = false
					mUnshorten.Uncheck()
				} else {
					isUnshortenEnabled = true
					mUnshorten.Check()
					// Feedback beep to confirm mode change
					nativeBeep() 
				}
			}
		}
	}()
}

func onExit() {}

func showNativeError(text, title string) {
	procMessageBox.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		MB_ICONERROR|MB_OK,
	)
}

func nativeBeep() {
	procMessageBeep.Call(0xFFFFFFFF)
}