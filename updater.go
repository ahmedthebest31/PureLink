package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sqweek/dialog"
)

const (
	updateURL        = "https://raw.githubusercontent.com/ahmedthebest31/PureLink/main/rules.json"
	gitHubReleaseURL = "https://api.github.com/repos/ahmedthebest31/PureLink/releases/latest"
	releaseNotesURL  = "https://github.com/ahmedthebest31/PureLink/releases/latest"
)

type githubRelease struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func (app *App) UpdateFilters(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, updateURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var newConfig RuleConfig
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&newConfig); err != nil {
		return fmt.Errorf("invalid rule format: %w", err)
	}

	if len(newConfig.Blocklist) == 0 {
		return fmt.Errorf("downloaded rules are empty")
	}

	app.mu.Lock()
	app.ActiveBlocklist = newConfig.Blocklist
	err = app.writeRulesFile(&newConfig)
	app.mu.Unlock()

	if err != nil {
		return fmt.Errorf("failed to persist rules: %w", err)
	}

	return nil
}

func isNewerVersion(current, latest string) bool {
	cur := strings.TrimPrefix(current, "v")
	lat := strings.TrimPrefix(latest, "v")

	curParts := strings.Split(cur, ".")
	latParts := strings.Split(lat, ".")

	maxLen := len(curParts)
	if len(latParts) > maxLen {
		maxLen = len(latParts)
	}

	for i := 0; i < maxLen; i++ {
		var c, l int
		if i < len(curParts) {
			c, _ = strconv.Atoi(curParts[i])
		}
		if i < len(latParts) {
			l, _ = strconv.Atoi(latParts[i])
		}
		if l > c {
			return true
		}
		if c > l {
			return false
		}
	}
	return false
}

func getAssetForPlatform(assets []releaseAsset) *releaseAsset {
	for _, a := range assets {
		if strings.Contains(a.Name, runtime.GOOS) && strings.Contains(a.Name, runtime.GOARCH) {
			return &a
		}
	}
	return nil
}

func (app *App) checkForUpdates(force bool) {
	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Get(gitHubReleaseURL)
	if err != nil {
		if force {
			dialog.Message("Failed to check for updates: %v", err).Title("Update Error").Error()
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if force {
			dialog.Message("Update server returned status %d.", resp.StatusCode).Title("Update Error").Error()
		}
		return
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		if force {
			dialog.Message("Failed to parse update response: %v", err).Title("Update Error").Error()
		}
		return
	}

	if release.TagName == "" {
		if force {
			dialog.Message("No version information found.").Title("Update Error").Error()
		}
		return
	}

	if !force {
		app.mu.RLock()
		skipped := app.Config.SkippedVersion
		app.mu.RUnlock()
		if release.TagName == skipped {
			return
		}
	}

	if !isNewerVersion(AppVersion, release.TagName) {
		if force {
			dialog.Message("You are on the latest version (%s).", AppVersion).Title("Up to Date").Info()
		}
		return
	}

	if !dialog.Message("Update %s is available.\nDo you want to download and install it now?", release.TagName).Title("Update Available").YesNo() {
		app.mu.Lock()
		app.Config.SkippedVersion = release.TagName
		app.writeConfigFile()
		app.mu.Unlock()
		return
	}

	asset := getAssetForPlatform(release.Assets)
	if asset == nil {
		dialog.Message("No compatible binary found for %s/%s in this release.", runtime.GOOS, runtime.GOARCH).Title("Update Error").Error()
		return
	}

	if err := app.downloadAndInstall(asset.BrowserDownloadURL); err != nil {
		dialog.Message("Failed to install update: %v", err).Title("Update Error").Error()
	}
}

func (app *App) downloadAndInstall(downloadURL string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	tmpPath := exe + ".new.tmp"
	oldPath := exe + ".old"

	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("cannot create temp file: %w", err)
	}

	resp, err := http.Get(downloadURL)
	if err != nil {
		out.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("download failed: %w", err)
	}

	_, err = io.Copy(out, resp.Body)
	resp.Body.Close()
	out.Close()

	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("download write failed: %w", err)
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cannot set executable permission: %w", err)
	}

	if err := os.Rename(exe, oldPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cannot rename current executable: %w", err)
	}

	if err := os.Rename(tmpPath, exe); err != nil {
		os.Rename(oldPath, exe)
		return fmt.Errorf("cannot place new executable: %w", err)
	}

	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()

	os.Exit(0)
	return nil
}
