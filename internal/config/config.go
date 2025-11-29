package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	Environment  string
	CookiesDir   string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if exists (ignore error in production)
	_ = godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "3005"),
		Environment: getEnv("ENV", "development"),
		CookiesDir:  getEnv("COOKIES_DIR", "cookies"),
	}
}

// getEnv gets environment variable with fallback to default value
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
