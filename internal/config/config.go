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
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if exists (ignore error in production)
	_ = godotenv.Load()

	port := getEnv("PORT", "3005")
	env := getEnv("ENV", "development")
	cookiesDir := getEnv("COOKIES_DIR", "cookies")

	// Parse allowed origins from env (comma-separated)
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	
	if allowedOriginsStr != "" {
		// Split by comma and trim spaces
		origins := strings.Split(allowedOriginsStr, ",")
		for _, origin := range origins {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}
	
	// Default allowed origins for development
	if len(allowedOrigins) == 0 {
		if env == "development" {
			// In development, allow localhost variants
			allowedOrigins = []string{
				"http://localhost:3000",
				"http://localhost:3001",
				"http://localhost:8080",
				"http://127.0.0.1:3000",
			}
		} else {
			// In production, require explicit configuration
			allowedOrigins = []string{}
		}
	}

	return &Config{
		Port:           port,
		Environment:    env,
		CookiesDir:     cookiesDir,
		AllowedOrigins: allowedOrigins,
	}
}

// getEnv gets environment variable with fallback to default value
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
