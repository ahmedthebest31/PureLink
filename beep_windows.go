package main

import "syscall"

func NotifyBeep() {
	syscall.NewLazyDLL("user32.dll").NewProc("MessageBeep").Call(0xFFFFFFFF)
}
