# PureLink 🛡️

**PureLink** is a privacy-focused, native Windows utility designed for developers and power users. It sanitizes clipboard links, unshortens URLs, and bridges the gap between Windows paths and WSL (Linux) terminals.

![OS](https://img.shields.io/badge/OS-Windows-blue) ![Go](https://img.shields.io/badge/Built%20with-Go-00ADD8) ![License](https://img.shields.io/badge/License-MIT-green)

## ✨ Features

### 🛡️ Privacy & Security
- **Smart Cleaner:** Instantly strips tracking parameters (`utm_*`, `fbclid`, `gclid`, etc.).
- **X-Ray Mode:** Resolves short links (e.g., `bit.ly`, `t.co`) to their destination *before* you paste.
- **YouTube Fix:** Converts Shorts (`/shorts/`) to standard watch URLs for better accessibility.

### 🐧 Developer Productivity
- **WSL Bridge:** Toggle "WSL Mode" to convert `C:\Projects` to `/mnt/c/Projects` automatically.
- **Smart Quoting:** Intelligent handling of file paths—only adds quotes if the path contains spaces.
- **Path Normalizer:** Fixes backslashes `\` to universal forward slashes `/`.

### 🚀 Native Performance
- **Zero-Config:** Runs silently in the System Tray.
- **Tiny Footprint:** Native Win32 API implementation (< 5MB RAM).
- **Single Instance:** Prevents multiple copies from running.

## 📦 Installation

### Option 1: Download Binary (Recommended)
1. Go to the [Releases Page](https://github.com/ahmedthebest31/PureLink/releases).
2. Download the latest `PureLink-vX.X.X-Windows.exe`.
3. Run it! (Optional: Add shortcut to Startup folder).

### Option 2: Install via Go (For Developers)
If you have Go installed, you can compile and install directly:

```bash
go install [github.com/ahmedthebest31/PureLink@latest](https://github.com/ahmedthebest31/PureLink@latest)
```

Note: Requires a Windows environment.


### 🛠️ Usage
1. Run the application. An icon will appear in the System Tray.
2. Right-Click the icon to configure:

- [x] Unshorten Links: Enable to expand short URLs (Requires Internet).
- [x] WSL Path Mode: Enable to convert Windows paths to Linux/WSL format.
- Just Copy (Ctrl+C). PureLink handles the rest.

### 🤝 Contributing
Contributions are welcome! Please note that this version utilizes syscall for Windows native features.


### 📄 License

MIT License