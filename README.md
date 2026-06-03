# PureLink 🚀

[![Go Report Card](https://goreportcard.com/badge/github.com/ahmedthebest31/PureLink)](https://goreportcard.com/report/github.com/ahmedthebest31/PureLink)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/ahmedthebest31/PureLink?label=Version&color=blue)
![Platforms](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-blue)

The native privacy guard and clipboard manager for developers. Runs as a background daemon with a system tray interface on Windows, Linux, and macOS.

---

## Features

PureLink is a cross-platform desktop daemon that protects your privacy and accelerates your workflow by intercepting clipboard copies and processing them locally.

### Privacy Guard

- Strips tracking parameters (`utm_*`, `fbclid`, `gclid`, `si`, `ref`, and 15+ more) from links copied to your clipboard
- All processing is local and instant — no data leaves your machine
- Ignore List (whitelist): specify domains like `localhost` or `127.0.0.1` to skip cleaning entirely, configurable via `purelink_config.json`

### Developer Tools

- **WSL Path Mode** — Persistently convert `C:\Projects` to `/mnt/c/Projects` on all copies, or use the one-shot "Convert Clipboard to WSL Path" action
- **Base64 Encode / Decode** — Encode or decode Base64 strings directly in the clipboard
- **Generate UUID** — Create a v4 UUID and copy it instantly
- **URL Encode / Decode** — Percent-encode or decode clipboard content
- **Format JSON** — Pretty-print arbitrary JSON from the clipboard
- **Unix Timestamp** — Generate the current Unix epoch seconds
- **Decode JWT Payload** — Extract and pretty-print the JSON payload from any JWT token

### Link Utilities

- **Unshorten Links** — Resolve shortened URLs (bit.ly, t.co, tinyurl.com, and more) to their original destination
- **Cloud Boost** — Convert Dropbox share links to direct `?dl=1` download links and Google Drive `view` links to direct `uc?export=download` links
- **YouTube Shorts Fix** — Automatically convert `/shorts/` URLs to `/watch` with the video ID preserved

### Social Tools

- **Convert to WhatsApp** — Read a phone number from the clipboard and format it as a `wa.me` link (no browser opens)
- **Convert to Telegram** — Read a Telegram username from the clipboard and format it as a `t.me` link
- **Clipboard Commands** — Type `!wa`, `!tg`, `!b64e`, `!b64d`, or `!uuid` followed by text, copy it, and PureLink processes it automatically (can be toggled off)

### Background Daemon

- **Clipboard Watcher** — Continuously monitors clipboard changes with a 1-second debounce to avoid redundant processing
- **Auto-Update Filters** — Fetches the latest tracking-rule blocklist from GitHub every 7 days in the background
- **Self-Update** — Checks GitHub releases on startup; prompts you when a newer version is available. Supports safe binary replacement (download .tmp, rename, spawn new process)
- **Single-Instance Lock** — Prevents multiple instances via a flock file in the OS-standard config directory
- **Startup Registration** — Auto-start on login (Windows Registry / XDG `.desktop` file on Linux / macOS)

### Architecture

- Thread-safe `App` struct with `sync.RWMutex` protecting `Config` and `ActiveBlocklist`
- Configuration stored as JSON in `os.UserConfigDir()/PureLink/purelink_config.json`
- Rules stored as JSON in `os.UserConfigDir()/PureLink/rules.json`
- No global variables — all state is encapsulated in the `App` struct
- Sentinel-error redirect resolver prevents body downloads during URL unshortening

---

## System Tray Menu Structure

The tray menu is organized into screen-reader-friendly submenus:

1. **Recent History** — Last 5 cleaned links, clickable to re-copy
2. **Social Tools** — Convert to WhatsApp, Convert to Telegram
3. **Updates and Rules** — Update Filters Now, Check for App Updates, View Release Notes
4. **Advanced Settings** — YouTube Shorts Fix, Cloud Boost, Clipboard Commands, Clear Data and History
5. **Developer Tools** — WSL Path Mode, Convert Clipboard to WSL Path, Base64 Encode, Base64 Decode, Generate UUID, URL Encode, URL Decode, Format JSON, Unix Timestamp, Decode JWT Payload
6. **Main Controls** — Pause/Resume Protection, Play Sound, Unshorten Links, Run on Startup
7. **Open Config File** — Opens `purelink_config.json` in your default editor
8. **Quit** — Exit PureLink

---

## Installation

### Option 1: Go Install (Universal)

If you have Go installed:

```bash
go install github.com/ahmedthebest31/PureLink@latest
```

### Option 2: Download Binaries (Recommended)

Pre-compiled binaries for Windows, Linux, and macOS are on the Releases page:

1. Go to the [PureLink Releases page](https://github.com/ahmedthebest31/PureLink/releases)
2. Download the binary for your OS and architecture
3. Make it executable (Linux/macOS: `chmod +x PureLink`) and run it

---

## Usage

Launch PureLink and it appears as an icon in your system tray. Right-click to access the full menu.

Copy any URL to your clipboard. PureLink automatically:

- Strips tracking parameters
- Unshortens links (if enabled)
- Converts Dropbox/Drive links to direct downloads (if enabled)
- Converts YouTube Shorts to watch URLs (if enabled)
- Converts Windows paths to WSL format (if WSL Mode is on)

### Ignore List

Open the config file from the tray menu and add domains to the `ignore_list` array:

```json
"ignore_list": ["localhost", "127.0.0.1", "internal.company.dev"]
```

Links whose host contains any ignored domain are returned completely untouched.

---

## Ecosystem

### Mobile Companion

For on-the-go privacy protection, check out the Android version:

[![PureLink Android](https://img.shields.io/badge/PureLink-Android-green)](https://github.com/ahmedthebest31/PureLink-Android)
[PureLink-Android](https://github.com/ahmedthebest31/PureLink-Android)

---

## License

MIT License

## Contributing

Contributions are welcome! This project uses syscall for Windows native features. Build tags (`//go:build windows` / `//go:build !windows`) keep platform-specific code isolated. Run `go build ./...` and `go vet ./...` before submitting.
