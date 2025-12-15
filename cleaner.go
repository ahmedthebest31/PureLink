package main

import (
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
)

// CleanText determines if the input is a URL or a File Path and processes it accordingly.
func CleanText(input string, unshorten bool, wslMode bool) string {
	trimmed := strings.TrimSpace(input)

	// 1. Path Detection
	if isWindowsPath(trimmed) {
		return processPath(trimmed, wslMode)
	}

	// 2. URL Cleaning (Standard Logic)
	if !strings.HasPrefix(trimmed, "http") {
		return input
	}

	finalURL := trimmed

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

	dirtyParams := []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "si", "ref", "gclid", "share_id",
		"igshid", "_ga", "yclid", "mc_eid",
	}

	q := u.Query()
	for _, param := range dirtyParams {
		q.Del(param)
	}

	if strings.Contains(u.Host, "youtube.com") && strings.Contains(u.Path, "/shorts/") {
		videoID := strings.TrimPrefix(u.Path, "/shorts/")
		u.Path = "/watch"
		q.Set("v", videoID)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// processPath handles normalization and applies "Smart Quoting".
func processPath(input string, wslMode bool) string {
	// Step 1: Strip existing quotes to start fresh
	clean := strings.Trim(input, "\"")
	clean = strings.Trim(clean, "'")

	// Step 2: Normalize slashes
	clean = strings.ReplaceAll(clean, "\\", "/")

	// Step 3: WSL Conversion
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

	// Step 4: Smart Quoting (The Professional Way)
	// ONLY if the path contains a space, wrap the WHOLE path in quotes.
	if strings.Contains(clean, " ") {
		return "\"" + clean + "\""
	}

	// If no spaces, return raw path
	return clean
}

// isWindowsPath checks for local file paths
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

// resolveURL logic
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

// isShortLink detection
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