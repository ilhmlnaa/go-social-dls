package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	Environment    string
	CookiesDir     string
	AllowedOrigins []string
	BaseURL        string
}

func Load() *Config {
	_ = godotenv.Load()

	port := getEnv("PORT", "3005")
	env := getEnv("ENV", "development")
	cookiesDir := getEnv("COOKIES_DIR", "cookies")
	baseURL := getEnv("BASE_URL", "http://localhost:3005")

	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	
	if allowedOriginsStr != "" {
		origins := strings.Split(allowedOriginsStr, ",")
		for _, origin := range origins {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}
	
	if len(allowedOrigins) == 0 {
		if env == "development" {
			allowedOrigins = []string{
				"http://localhost:3000",
				"http://localhost:3001",
				"http://localhost:8080",
				"http://127.0.0.1:3000",
			}
		} else {
			allowedOrigins = []string{}
		}
	}

	return &Config{
		Port:           port,
		Environment:    env,
		CookiesDir:     cookiesDir,
		AllowedOrigins: allowedOrigins,
		BaseURL:        baseURL,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
