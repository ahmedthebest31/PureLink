package main

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var errRedirect = errors.New("redirect")

func (app *App) CleanText(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return input
	}

	// 1. Command triggers (!wa, !tg, !b64e, !b64d, !uuid)
	if strings.HasPrefix(trimmed, "!") {
		if result, ok := app.handleCommand(trimmed); ok {
			return result
		}
	}

	// 2. Windows path detection
	if isWindowsPath(trimmed) {
		app.mu.RLock()
		wslMode := app.Config.WSLMode
		app.mu.RUnlock()
		return processPath(trimmed, wslMode)
	}

	// 3. Non-URL handling (auto-format social links)
	if !strings.HasPrefix(trimmed, "http") {
		if result, ok := app.formatSocialLinks(trimmed); ok {
			return result
		}
		return input
	}

	// 4. URL cleaning
	app.mu.RLock()
	unshorten := app.Config.Unshorten
	cloudBoost := app.Config.DirectLink
	app.mu.RUnlock()

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

	params := app.GetRules()

	q := u.Query()
	for _, param := range params {
		q.Del(param)
	}

	if strings.Contains(u.Host, "youtube.com") && strings.Contains(u.Path, "/shorts/") {
		videoID := strings.TrimPrefix(u.Path, "/shorts/")
		u.Path = "/watch"
		q.Set("v", videoID)
	}

	if cloudBoost {
		if strings.Contains(u.Host, "dropbox.com") {
			q.Set("dl", "1")
		} else if strings.Contains(u.Host, "drive.google.com") && strings.Contains(u.Path, "/view") {
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

func (app *App) handleCommand(input string) (string, bool) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", false
	}

	switch strings.ToLower(parts[0]) {
	case "!wa", "!whatsapp":
		if len(parts) < 2 {
			return "", false
		}
		num := strings.Join(parts[1:], "")
		result, err := GetWhatsAppLink(num)
		if err != nil {
			return input, false
		}
		return result, true

	case "!tg", "!telegram":
		if len(parts) < 2 {
			return "", false
		}
		result, err := GetTelegramLink(parts[1])
		if err != nil {
			return input, false
		}
		return result, true

	case "!b64e", "!base64encode":
		if len(parts) < 2 {
			return "", false
		}
		return EncodeBase64(strings.Join(parts[1:], " ")), true

	case "!b64d", "!base64decode":
		if len(parts) < 2 {
			return "", false
		}
		result, err := DecodeBase64(strings.Join(parts[1:], ""))
		if err != nil {
			return input, false
		}
		return result, true

	case "!uuid":
		return GenerateUUID(), true
	}

	return "", false
}

func (app *App) formatSocialLinks(input string) (string, bool) {
	app.mu.RLock()
	wa := app.Config.WhatsAppLinks
	tg := app.Config.TelegramLinks
	app.mu.RUnlock()

	if wa {
		matched, _ := regexp.MatchString(`^\d{7,18}$`, input)
		if matched {
			result, err := GetWhatsAppLink(input)
			if err == nil {
				return result, true
			}
		}
	}

	if tg {
		trimmed := strings.TrimPrefix(input, "@")
		if len(trimmed) >= 4 && len(trimmed) <= 40 && !strings.Contains(trimmed, " ") {
			matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, trimmed)
			if matched {
				return "https://t.me/" + trimmed, true
			}
		}
	}

	return "", false
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
			return errRedirect
		},
	}

	resp, err := client.Head(shortURL)
	if err != nil {
		var urlErr *url.Error
		if !errors.As(err, &urlErr) || !errors.Is(urlErr.Err, errRedirect) {
			resp, err = client.Get(shortURL)
			if err != nil {
				return ""
			}
		}
	}

	if resp == nil {
		return ""
	}
	resp.Body.Close()
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
		if u.Host == s || strings.HasSuffix(u.Host, "."+s) {
			return true
		}
	}
	if len(input) < 30 && !strings.Contains(u.Host, "localhost") && !strings.Contains(u.Host, "127.0.0.1") {
		return true
	}
	return false
}
