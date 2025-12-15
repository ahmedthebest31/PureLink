package main

import (
	"net/url"
	"strings"
)

// CleanText processes the input string to remove tracking parameters.
// It returns the cleaned URL if modifications were made, otherwise returns the original string.
func CleanText(input string) string {
	// 1. Check if the input is a valid URL
	if !strings.HasPrefix(input, "http") {
		return input
	}

	u, err := url.Parse(input)
	if err != nil {
		return input
	}

	// 2. List of query parameters to remove (Tracking IDs)
	dirtyParams := []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "si", "ref", "gclid", "share_id",
		"igshid", "_ga",
	}

	q := u.Query()
	for _, param := range dirtyParams {
		q.Del(param)
	}

	// 3. Special handling for YouTube Shorts -> Watch format
	// This improves accessibility for screen readers by using the standard player.
	if strings.Contains(u.Host, "youtube.com") && strings.Contains(u.Path, "/shorts/") {
		videoID := strings.TrimPrefix(u.Path, "/shorts/")
		u.Path = "/watch"
		q.Set("v", videoID)
	}

	// Reconstruct the URL with cleaned parameters
	u.RawQuery = q.Encode()
	return u.String()
}