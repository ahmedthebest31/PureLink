package main

import "syscall"

var (
	user32          = syscall.NewLazyDLL("user32.dll")
	procMessageBeep = user32.NewProc("MessageBeep")
)

func NotifyBeep() {
	procMessageBeep.Call(0xFFFFFFFF)
}
