package main

import (
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
)

// CleanText processes input for Privacy, Cloud links, and Path normalization.
func CleanText(input string, unshorten bool, wslMode bool, cloudBoost bool) string {
	trimmed := strings.TrimSpace(input)

	// 1. Path Detection
	if isWindowsPath(trimmed) {
		return processPath(trimmed, wslMode)
	}

	// 2. URL Cleaning
	if !strings.HasPrefix(trimmed, "http") {
		return input
	}

	finalURL := trimmed

	// Unshorten logic
	if unshorten && isShortLink(trimmed) {
		resolved := resolveURL(trimmed)
		if resolved != "" && resolved != trimmed {
			finalURL = resolved
		}
	}

	u, err := url.Parse(finalURL)
	if err != nil {
		return finalURL
	}

	// Remove tracking parameters using dynamic blocklist
	BlocklistLock.RLock()
	currentParams := ActiveBlocklist
	BlocklistLock.RUnlock()

	q := u.Query()
	for _, param := range currentParams {
		q.Del(param)
	}

	// Fix YouTube Shorts
	if strings.Contains(u.Host, "youtube.com") && strings.Contains(u.Path, "/shorts/") {
		videoID := strings.TrimPrefix(u.Path, "/shorts/")
		u.Path = "/watch"
		q.Set("v", videoID)
	}

	// 3. Cloud Booster (Dropbox & Google Drive)
	if cloudBoost {
		// Automatically convert to direct download links
		if strings.Contains(u.Host, "dropbox.com") {
			q.Set("dl", "1")
		} else if strings.Contains(u.Host, "drive.google.com") && strings.Contains(u.Path, "/view") {
			// Convert /file/d/ID/view -> /uc?export=download&id=ID
			parts := strings.Split(u.Path, "/")
			for i, part := range parts {
				if part == "d" && i+1 < len(parts) {
					id := parts[i+1]
					u.Path = "/uc"
					q.Set("export", "download")
					q.Set("id", id)
					break
				}
			}
		}
	}

	u.RawQuery = q.Encode()
	return u.String()
}

func processPath(input string, wslMode bool) string {
	clean := strings.Trim(input, "\"")
	clean = strings.Trim(clean, "'")
	clean = strings.ReplaceAll(clean, "\\", "/")

	if wslMode {
		if len(clean) > 1 && clean[1] == ':' {
			drive := strings.ToLower(string(clean[0]))
			remainder := clean[2:]
			if !strings.HasPrefix(remainder, "/") {
				remainder = "/" + remainder
			}
			clean = "/mnt/" + drive + remainder
		}
	}

	// Smart Quoting
	if strings.Contains(clean, " ") {
		return "\"" + clean + "\""
	}
	return clean
}

func isWindowsPath(s string) bool {
	check := strings.Trim(s, "\"")
	if len(check) < 2 {
		return false
	}
	if unicode.IsLetter(rune(check[0])) && check[1] == ':' {
		return true
	}
	if strings.HasPrefix(check, "\\\\") {
		return true
	}
	return false
}

func resolveURL(shortURL string) string {
	client := &http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	resp, err := client.Head(shortURL)
	if err != nil {
		resp, err = client.Get(shortURL)
		if err != nil {
			return ""
		}
	}
	defer resp.Body.Close()
	return resp.Request.URL.String()
}

func isShortLink(input string) bool {
	u, err := url.Parse(input)
	if err != nil {
		return false
	}
	shorteners := []string{
		"bit.ly", "goo.gl", "t.co", "tinyurl.com", "is.gd",
		"buff.ly", "amzn.to", "lnkd.in", "rebrand.ly", "shrtco.de",
	}
	for _, s := range shorteners {
		if u.Host == s || strings.HasSuffix(u.Host, s) {
			return true
		}
	}
	if len(input) < 30 && !strings.Contains(u.Host, "localhost") && !strings.Contains(u.Host, "127.0.0.1") {
		return true
	}
	return false
}