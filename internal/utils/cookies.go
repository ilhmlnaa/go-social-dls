package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type CookieItem struct {
	Domain         string  `json:"domain"`
	ExpirationDate float64 `json:"expirationDate,omitempty"`
	HostOnly       bool    `json:"hostOnly"`
	HTTPOnly       bool    `json:"httpOnly"`
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	SameSite       string  `json:"sameSite,omitempty"`
	Secure         bool    `json:"secure"`
	Session        bool    `json:"session"`
	StoreID        *string `json:"storeId"`
	Value          string  `json:"value"`
}

func LoadCookiesFromFile(filename string) (map[string]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read cookie file: %w", err)
	}

	var cookieItems []CookieItem
	if err := json.Unmarshal(data, &cookieItems); err != nil {
		return nil, fmt.Errorf("failed to parse cookie file: %w", err)
	}

	cookies := make(map[string]string)
	for _, item := range cookieItems {
		// Clean cookie value - remove surrounding quotes and escape characters
		// This fixes "net/http: invalid byte" errors
		value := item.Value
		
		// Remove surrounding quotes
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		
		// Remove backslashes (escape characters)
		value = strings.ReplaceAll(value, `\`, "")
		
		cookies[item.Name] = value
	}

	return cookies, nil
}

func LoadCookies(cookiesDir, platform string) (map[string]string, error) {
	filename := filepath.Join(cookiesDir, platform+".json")
	cookies, err := LoadCookiesFromFile(filename)
	if err == nil {
		return cookies, nil
	}

	// TODO: Try environment variable as fallback
	return nil, fmt.Errorf("failed to load cookies for %s: %w", platform, err)
}

func ValidateCookies(cookies map[string]string, required []string) error {
	for _, key := range required {
		if _, ok := cookies[key]; !ok {
			return fmt.Errorf("missing required cookie: %s", key)
		}
	}
	return nil
}
