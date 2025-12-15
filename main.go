package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
)

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

	systray.AddMenuItem("Status: Active", "Protection is enabled").Disable()
	mCounter := systray.AddMenuItem("Cleaned: 0 Links", "Total items processed")
	
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

	mUnshorten := systray.AddMenuItemCheckbox("Unshorten Links", "Expand short URLs (Requires Internet)", false)
	mWSL := systray.AddMenuItemCheckbox("WSL Path Mode", "Convert C:\\ to /mnt/c/ and fix slashes", false)

	systray.AddSeparator()

	mSound := systray.AddMenuItemCheckbox("Play Sound", "Beep when item is cleaned", true)
	mPause := systray.AddMenuItem("Pause Protection", "Temporarily stop cleaning")

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit PureLink")

	// State
	isRunning := true
	isSoundEnabled := true
	isUnshortenEnabled := false
	isWSLMode := false
	cleanedCount := 0

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
				
				cleaned := CleanText(text, isUnshortenEnabled, isWSLMode)

				if cleaned != text {
					clipboard.WriteAll(cleaned)
					lastText = cleaned
					
					cleanedCount++
					mCounter.SetTitle(fmt.Sprintf("Cleaned: %d Items", cleanedCount))

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
					nativeBeep()
				}

			case <-mWSL.ClickedCh:
				if isWSLMode {
					isWSLMode = false
					mWSL.Uncheck()
				} else {
					isWSLMode = true
					mWSL.Check()
					nativeBeep()
				}
			
			// --- Tools Actions ---
			
			case <-tWhatsApp.ClickedCh:
				text, _ := clipboard.ReadAll()
				url, err := GetWhatsAppLink(text)
				if err == nil {
					clipboard.WriteAll(url)
					OpenBrowser(url)
					nativeBeep()
				}

			case <-tTelegram.ClickedCh:
				text, _ := clipboard.ReadAll()
				url, err := GetTelegramLink(text)
				if err == nil {
					clipboard.WriteAll(url)
					OpenBrowser(url)
					nativeBeep()
				}

			case <-tDecode64.ClickedCh:
				text, _ := clipboard.ReadAll()
				decoded, err := DecodeBase64(text)
				if err == nil && decoded != "" {
					clipboard.WriteAll(decoded)
					nativeBeep()
				}
			
			case <-tEncode64.ClickedCh:
				text, _ := clipboard.ReadAll()
				if text != "" {
					encoded := EncodeBase64(text)
					clipboard.WriteAll(encoded)
					nativeBeep()
				}

			case <-tUUID.ClickedCh:
				id := GenerateUUID()
				clipboard.WriteAll(id)
				nativeBeep()
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