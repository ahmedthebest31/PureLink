# PureLink 🛡️

**PureLink** is a lightweight, background utility written in Go. It automatically monitors your clipboard **to** strip tracking parameters (UTM, Facebook IDs, etc.) and converts YouTube Shorts links to standard watch URLs.

> **Note:** Designed with accessibility in mind. It provides audio/visual feedback via system notifications when a link is cleaned.

## Features
- 🚀 **Zero-Config:** Runs silently in the system tray.
- 🧹 **Privacy First:** Removes `utm_source`, `fbclid`, `gclid`, and more.
- 🔗 **YouTube Fix:** Converts `/shorts/` to `/watch?v=` automatically.
- 🔊 **Feedback:** System notification upon cleaning.
- ⏸️ **Control:** Pause/Resume from the tray menu.

## Installation

### From Source (Go required)
```bash
git clone [https://github.com/ahmedthebest31/PureLink.git](https://github.com/ahmedthebest31/PureLink.git)
cd PureLink
go mod tidy
go build -ldflags -H=windowsgui -o PureLink.exe
```

### Usage
1. Run PureLink.exe.
2. An icon will appear in your system tray.
3. Copy any "dirty" link (e.g., from Facebook or Amazon).
4. Paste it anywhere - it will be clean!



### License
MIT
