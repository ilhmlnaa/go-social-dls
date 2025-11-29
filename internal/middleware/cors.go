package middleware

import (
	"github.com/gofiber/fiber/v2/middleware/cors"

	"twitter-down/internal/config"
)

func SetupCORS(cfg *config.Config) cors.Config {
	if len(cfg.AllowedOrigins) == 0 {
		return cors.Config{
			AllowOrigins:     "",
			AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
			AllowMethods:     "GET, POST, OPTIONS",
			AllowCredentials: false,
		}
	}

	originsStr := ""
	for i, origin := range cfg.AllowedOrigins {
		if i > 0 {
			originsStr += ", "
		}
		originsStr += origin
	}
	return cors.Config{
		AllowOrigins:     originsStr,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, OPTIONS",
		AllowCredentials: true,
		MaxAge:           3600,
	}
}
