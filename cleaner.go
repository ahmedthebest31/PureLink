package main

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CleanText processes the input string to remove tracking parameters.
// It returns the cleaned URL if modifications were made, otherwise returns the original string.
// If unshorten is true, it attempts to resolve short links.
func CleanText(input string, unshorten bool) string {
	// 1. Check if the input is a valid URL
	if !strings.HasPrefix(input, "http") {
		return input
	}

	finalURL := input

	// 2. Unshorten logic (The X-Ray Feature)
	// Only runs if the user enabled it in the menu.
	if unshorten && isShortLink(input) {
		resolved := resolveURL(input)
		// Only update if we successfully got a different URL
		if resolved != "" && resolved != input {
			finalURL = resolved
		}
	}

	// 3. Cleaning logic (UTM and tracking params removal)
	u, err := url.Parse(finalURL)
	if err != nil {
		// Return the unshortened URL even if parsing failed partially
		return finalURL
	}

	// List of query parameters to remove
	dirtyParams := []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "si", "ref", "gclid", "share_id",
		"igshid", "_ga", "yclid", "mc_eid",
	}

	q := u.Query()
	for _, param := range dirtyParams {
		q.Del(param)
	}

	// 4. Special handling for YouTube Shorts -> Watch format
	// Converts /shorts/videoID to /watch?v=videoID for better accessibility.
	if strings.Contains(u.Host, "youtube.com") && strings.Contains(u.Path, "/shorts/") {
		videoID := strings.TrimPrefix(u.Path, "/shorts/")
		u.Path = "/watch"
		q.Set("v", videoID)
	}

	// Reconstruct the URL with cleaned parameters
	u.RawQuery = q.Encode()
	return u.String()
}

// resolveURL follows the redirect chain to find the destination.
func resolveURL(shortURL string) string {
	client := &http.Client{
		// Timeout is set to 3 seconds to prevent the clipboard from hanging
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Stop after 10 redirects to avoid infinite loops
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// Use HEAD request first to fetch headers only (saves bandwidth)
	resp, err := client.Head(shortURL)
	if err != nil {
		// Fallback to GET if HEAD fails (some servers block HEAD requests)
		resp, err = client.Get(shortURL)
		if err != nil {
			return ""
		}
	}
	defer resp.Body.Close()

	// Return the final URL
	return resp.Request.URL.String()
}

// isShortLink checks if the domain is a known shortener or if the URL length is short.
func isShortLink(input string) bool {
	u, err := url.Parse(input)
	if err != nil {
		return false
	}

	// Explicit list of common URL shorteners
	shorteners := []string{
		"bit.ly", "goo.gl", "t.co", "tinyurl.com", "is.gd",
		"buff.ly", "amzn.to", "lnkd.in", "rebrand.ly", "shrtco.de",
	}

	for _, s := range shorteners {
		if u.Host == s || strings.HasSuffix(u.Host, s) {
			return true
		}
	}

	// Aggressive check: If URL is short (less than 30 chars) and not local, treat as potential short link.
	if len(input) < 30 && !strings.Contains(u.Host, "localhost") && !strings.Contains(u.Host, "127.0.0.1") {
		return true
	}

	return false
}