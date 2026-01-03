# PureLink 🚀

[![Go Report Card](https://goreportcard.com/badge/github.com/ahmedthebest31/PureLink)](https://goreportcard.com/report/github.com/ahmedthebest31/PureLink)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/ahmedthebest31/PureLink?label=Version&color=blue)
![Platforms](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-blue)

The native privacy guard and clipboard manager for developers. Now runs everywhere.

---

## ✨ Features

PureLink offers a powerful suite of tools to enhance your privacy and boost your developer productivity across platforms.

*   🌍 **Cross-Platform**: Works seamlessly on **Windows**, **Linux**, and **macOS**.
*   🔄 **Live Updates**: Fetches the latest tracking filter rules from GitHub instantly, keeping your protection up-to-date.
*   📜 **History**: Keeps track of your last 5 cleaned links, easily accessible from the system tray menu.
*   🚀 **Launch on Startup**: Option to automatically launch PureLink when you log in, ensuring continuous protection.
*   🛡️ **Privacy Guard**: Strips common tracking parameters (e.g., `utm_*`, `fbclid`, `gclid`) from links copied to your clipboard, locally and instantly.
*   🔗 **Productivity Boost**:
    *   **Unshorten Links**: Automatically resolves shortened URLs (e.g., `bit.ly`, `t.co`) to their original destination.
    *   **Direct Cloud Links**: Converts Dropbox and Google Drive shareable links into direct download links.
    *   **WSL Bridge**: (Maintain from previous version) Toggle "WSL Mode" to convert `C:\Projects` to `/mnt/c/Projects` automatically.
    *   **Path Normalizer**: (Maintain from previous version) Fixes backslashes `\` to universal forward slashes `/`.

---

## 📦 Installation

PureLink is easy to install on any supported platform.

### Option 1: Go Install (Universal)
If you have Go installed, you can get PureLink directly from the source:

```bash
go install github.com/ahmedthebest31/PureLink@latest
```

This will compile and install the latest version of PureLink to your `$GOPATH/bin` (or `$HOME/go/bin`) directory.

### Option 2: Download Binaries (Recommended)
Pre-compiled binaries for Windows, Linux, and macOS are available on the GitHub Releases page:

1.  Go to the [PureLink Releases page](https://github.com/ahmedthebest31/PureLink/releases).
2.  Download the appropriate binary for your operating system.
3.  Execute the downloaded file.

---

## 🚀 Usage

Once launched, PureLink runs quietly in your system tray (or notification area).

**Right-Click the PureLink Tray Icon to access the menu:**

*   **Settings**: Configure preferences like enabling/disabling features (e.g., Unshorten Links, WSL Path Mode, Launch on Startup, Sound).
*   **Recent History**: View and re-copy your last 5 cleaned links.
*   **Tools Menu (Manual Utilities)**:
    *   **Open WhatsApp**: Copies a number and opens the chat directly (Handles International formats).
    *   **Open Telegram**: Copies a username and opens the profile.
    *   **Base64 Ops**: Encode or Decode Base64 strings directly in the clipboard.
    *   **Insert UUID**: Generates a fresh v4 UUID and copies it.

PureLink continuously monitors your clipboard in the background. Simply **copy a link (Ctrl+C)**, and PureLink will automatically sanitize it, unshorten it, or convert it based on your settings.

---

## 🌍 Ecosystem

### Mobile Companion
For on-the-go privacy protection, check out the Android version of PureLink:

[![PureLink Android](https://img.shields.io/badge/PureLink-Android-green)](https://github.com/ahmedthebest31/PureLink-Android)
[PureLink-Android](https://github.com/ahmedthebest31/PureLink-Android)

---

## 📄 License
MIT License

## 🤝 Contributing
Contributions are welcome! Please note that this version utilizes syscall for Windows native features.
