package handlers

import (
	"github.com/gofiber/fiber/v2"

	"twitter-down/internal/config"
	"twitter-down/internal/models"
)

// HealthCheck returns the health status of the API
func HealthCheck(c *fiber.Ctx) error {
	cfg := config.Load()

	return c.JSON(models.NewSuccessResponse(
		"API is healthy",
		fiber.Map{
			"status":      "ok",
			"environment": cfg.Environment,
			"version":     "2.0.0",
		},
	))
}
