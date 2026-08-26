package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"twitter-down/internal/config"
	"twitter-down/internal/handlers"
	"twitter-down/internal/middleware"
	"twitter-down/internal/models"
)

func main() {
	cfg := config.Load()

	app := fiber.New(fiber.Config{
		AppName:      "Social Media Downloader API v2.0",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	app.Use(recover.New())
	corsConfig := middleware.SetupCORS(cfg)
	app.Use(cors.New(corsConfig))
	
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))

	setupRoutes(app)
	
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(
			models.NewErrorResponse(
				"Endpoint not found",
				fiber.NewError(fiber.StatusNotFound, "The requested endpoint does not exist"),
			),
		)
	})

	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Printf("📁 Cookies directory: %s", cfg.CookiesDir)
	log.Printf("🌍 Environment: %s", cfg.Environment)
	
	if len(cfg.AllowedOrigins) > 0 {
		log.Printf("🔒 CORS allowed origins: %v", cfg.AllowedOrigins)
	} else {
		log.Printf("⚠️  CORS: No allowed origins configured (all blocked)")
	}
	
	log.Fatal(app.Listen(":" + cfg.Port))
}

func setupRoutes(app *fiber.App) {
	app.Get("/", handlers.Root)

	app.Get("/health", handlers.HealthCheck)

	api := app.Group("/api/v1")

	api.Get("/twitter", handlers.TwitterDownload)
	api.Get("/facebook", handlers.FacebookDownload)
	api.Get("/pinterest", handlers.PinterestDownload)
	api.Get("/instagram", handlers.InstagramDownload)
	api.Get("/pixiv", handlers.PixivDownload)
	api.Get("/danbooru", handlers.DanbooruDownload)
	api.Get("/generic", handlers.GenericDownload)
	api.Get("/proxy/image", handlers.ImageProxy)
	api.Get("/resolve/pinterest", handlers.ResolvePinterestUrl)

	app.Static("/static", "./static")
}

// isKnownPath checks if the given path is a registered endpoint
func isKnownPath(path string) bool {
	knownPaths := []string{
		"/",
		"/health",
		"/api/v1/twitter",
		"/api/v1/facebook",
		"/api/v1/pinterest",
		"/api/v1/instagram",
		"/api/v1/pixiv",
		"/api/v1/danbooru",
		"/api/v1/generic",
		"/api/v1/proxy/image",
		"/api/v1/resolve/pinterest",
	}
	
	for _, knownPath := range knownPaths {
		if path == knownPath {
			return true
		}
	}
	return false
}
