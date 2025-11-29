package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CookieItem represents a single cookie from browser export
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

// LoadCookiesFromFile loads cookies from JSON file exported by Cookie-Editor
func LoadCookiesFromFile(filename string) (map[string]string, error) {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read cookie file: %w", err)
	}

	// Parse JSON array
	var cookieItems []CookieItem
	if err := json.Unmarshal(data, &cookieItems); err != nil {
		return nil, fmt.Errorf("failed to parse cookie file: %w", err)
	}

	// Convert to map
	cookies := make(map[string]string)
	for _, item := range cookieItems {
		cookies[item.Name] = item.Value
	}

	return cookies, nil
}

// LoadCookies loads cookies with fallback strategy:
// 1. Try to load from file in cookies directory
// 2. Try to load from environment variable (base64 encoded)
// 3. Return error if both fail
func LoadCookies(cookiesDir, platform string) (map[string]string, error) {
	// Try file first
	filename := filepath.Join(cookiesDir, platform+".json")
	cookies, err := LoadCookiesFromFile(filename)
	if err == nil {
		return cookies, nil
	}

	// TODO: Try environment variable as fallback
	// For now, just return the file error
	return nil, fmt.Errorf("failed to load cookies for %s: %w", platform, err)
}

// ValidateCookies checks if required cookies are present
func ValidateCookies(cookies map[string]string, required []string) error {
	for _, key := range required {
		if _, ok := cookies[key]; !ok {
			return fmt.Errorf("missing required cookie: %s", key)
		}
	}
	return nil
}
