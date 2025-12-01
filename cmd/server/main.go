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
	api.Get("/generic", handlers.GenericDownload)
	api.Get("/proxy/image", handlers.ImageProxy)
	api.Get("/resolve/pinterest", handlers.ResolvePinterestUrl)

	app.Static("/static", "./static")
}
