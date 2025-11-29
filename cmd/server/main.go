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
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Social Media Downloader API v2.0",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, OPTIONS",
	}))
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))

	// Setup routes
	setupRoutes(app)

	// Start server
	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Printf("📁 Cookies directory: %s", cfg.CookiesDir)
	log.Printf("🌍 Environment: %s", cfg.Environment)
	log.Fatal(app.Listen(":" + cfg.Port))
}

func setupRoutes(app *fiber.App) {
	// Root endpoint
	app.Get("/", handlers.Root)

	// Health check
	app.Get("/health", handlers.HealthCheck)

	// API v1 group
	api := app.Group("/api/v1")

	// Download endpoints
	api.Get("/twitter", handlers.TwitterDownload)
	api.Get("/facebook", handlers.FacebookDownload)
	api.Get("/pinterest", handlers.PinterestDownload)
	api.Get("/instagram", handlers.InstagramDownload)
	api.Get("/generic", handlers.GenericDownload)

	// Serve static files (if needed)
	app.Static("/static", "./static")
}
