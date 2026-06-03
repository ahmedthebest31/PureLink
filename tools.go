package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GenerateUUID creates a random version 4 UUID.
func GenerateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[8] = b[8]&^0xc0 | 0x80
	b[6] = b[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// DecodeBase64 attempts to decode a base64 string.
func DecodeBase64(input string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// EncodeBase64 converts string to base64.
func EncodeBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// GetWhatsAppLink cleans the input (removes + and spaces) and creates link.
// It relies on the user to provide the correct Country Code.
func GetWhatsAppLink(input string) (string, error) {
	// 1. Clean: Remove everything except digits (removes +, -, spaces, etc.)
	re := regexp.MustCompile(`\D`)
	number := re.ReplaceAllString(input, "")

	// 2. Validation:
	// We allow a flexible range (e.g., 7 to 18 digits) to cover all countries.
	// If it's too long (like a UUID), we reject it.
	if len(number) < 7 || len(number) > 18 {
		return "", fmt.Errorf("invalid phone number length")
	}

	return "https://wa.me/" + number, nil
}

// GetTelegramLink validates username format.
func GetTelegramLink(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	trimmed = strings.TrimPrefix(trimmed, "@")
	
	if len(trimmed) < 4 || len(trimmed) > 40 || strings.Contains(trimmed, " ") {
		return "", fmt.Errorf("invalid username")
	}

	match, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", trimmed)
	if !match {
		return "", fmt.Errorf("invalid characters in username")
	}

	return "https://t.me/" + trimmed, nil
}

// OpenBrowser opens a URL using the default Windows handler.
func OpenBrowser(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

// ConvertToWSL converts a Windows path to WSL /mnt/ format.
func ConvertToWSL(input string) string {
	clean := strings.Trim(input, "\"'")
	clean = strings.ReplaceAll(clean, "\\", "/")
	if len(clean) > 1 && clean[1] == ':' {
		drive := strings.ToLower(string(clean[0]))
		remainder := clean[2:]
		if !strings.HasPrefix(remainder, "/") {
			remainder = "/" + remainder
		}
		clean = "/mnt/" + drive + remainder
	}
	return clean
}

// URLEncode percent-encodes a string.
func URLEncode(input string) string {
	return url.QueryEscape(input)
}

// URLDecode decodes a percent-encoded string.
func URLDecode(input string) (string, error) {
	return url.QueryUnescape(input)
}

// FormatJSON pretty-prints a JSON string.
func FormatJSON(input string) (string, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(input), &v); err != nil {
		return "", err
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// UnixTimestamp returns the current Unix epoch seconds as a string.
func UnixTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

// DecodeJWTPayload extracts and pretty-prints the JSON payload from a JWT token.
func DecodeJWTPayload(input string) (string, error) {
	parts := strings.Split(input, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid JWT format")
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("cannot decode JWT payload: %w", err)
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return "", fmt.Errorf("JWT payload is not valid JSON: %w", err)
	}
	out, _ := json.MarshalIndent(v, "", "  ")
	return string(out), nil
}