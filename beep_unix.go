//go:build !windows

package main

import "os"

func NotifyBeep() {
	os.Stderr.Write([]byte{0x07})
}
