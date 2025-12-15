# PureLink 🛡️

**PureLink** is a privacy-focused, native Windows utility designed for developers and power users. It sanitizes clipboard links, auto-converts cloud storage links, and bridges the gap between Windows paths and WSL.

![OS](https://img.shields.io/badge/OS-Windows-blue) ![Go](https://img.shields.io/badge/Built%20with-Go-00ADD8) ![License](https://img.shields.io/badge/License-MIT-green)

## ✨ Features

### 🛡️ Smart Cleaning (Automatic)
- **Tracker Remover:** Strips `utm_*`, `fbclid`, `gclid`, and other tracking parameters instantly.
- **Cloud Booster:** Automatically converts **Dropbox** and **Google Drive** links to direct download links (e.g., `dl=1`).
- **X-Ray Mode:** Resolves short links (e.g., `bit.ly`, `t.co`) to their destination *before* you paste.
- **YouTube Fix:** Converts Shorts (`/shorts/`) to standard watch URLs.

### 🛠️ Developer Tools (Manual Menu)
Right-click the tray icon to access the **Tools** menu:
- **Open WhatsApp:** Copies a number and opens the chat directly (Handles International formats).
- **Open Telegram:** Copies a username and opens the profile.
- **Base64 Ops:** Encode or Decode Base64 strings directly in the clipboard.
- **Insert UUID:** Generates a fresh v4 UUID and copies it.

### 🐧 Productivity
- **WSL Bridge:** Toggle "WSL Mode" to convert `C:\Projects` to `/mnt/c/Projects` automatically.
- **Path Normalizer:** Fixes backslashes `\` to universal forward slashes `/`.

## 📦 Installation

### Option 1: Download Binary (Recommended)
1. Go to the [Releases Page](https://github.com/ahmedthebest31/PureLink/releases).
2. Download the latest `PureLink-vX.X.X-Windows.exe`.
3. Run it! (Optional: Add shortcut to Startup folder).

### Option 2: Install via Go
```bash
go install https://github.com/ahmedthebest31/PureLink@latest
```

Note: Requires a Windows environment.


### 🛠️ Usage

1. Run the application. An icon will appear in the System Tray.
2. **Right-Click** the icon to access features:

   **🧰 Tools Menu (Manual Utilities):**
   - **Open WhatsApp:** Detects phone numbers (International format), copies the link, and opens the chat.
   - **Open Telegram:** Detects usernames, copies the link, and opens the profile.
   - **Base64 Ops:** Encode or Decode Base64 strings directly in the clipboard.
   - **Insert UUID:** Generates a fresh v4 UUID and copies it.

   **⚙️ Settings (Toggles):**
   - `[x] Unshorten Links`: Enable to expand short URLs (Requires Internet).
   - `[x] WSL Path Mode`: Enable to convert Windows paths to Linux/WSL format.

3. **Just Copy (Ctrl+C)**. PureLink handles tracking removal, path normalization, and cloud link boosting (Dropbox/Drive) automatically in the background.

### 🤝 Contributing
Contributions are welcome! Please note that this version utilizes `syscall` for Windows native features.

### 📄 License
MIT License
### 🤝 Contributing
Contributions are welcome! Please note that this version utilizes syscall for Windows native features.


### 📄 License

MIT License