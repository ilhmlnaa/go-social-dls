package handlers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"twitter-down/internal/config"
	"twitter-down/internal/models"
	"twitter-down/internal/services"
)

// FacebookDownload handles Facebook photo download requests
func FacebookDownload(c *fiber.Ctx) error {
	// Get URL parameter
	urlParam := c.Query("url")
	if urlParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Parameter 'url' is required",
			nil,
		))
	}

	// Load config
	cfg := config.Load()

	// Initialize Facebook service
	fbSvc, err := services.NewFacebookService(cfg.CookiesDir)
	if err != nil {
		log.Printf("[Facebook] Failed to initialize service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to initialize Facebook service. Make sure cookies are configured.",
			err,
		))
	}

	// Fetch photos
	photos, err := fbSvc.GetPhotoURLs(urlParam)
	if err != nil {
		log.Printf("[Facebook] Failed to fetch photos from %s: %v", urlParam, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to fetch Facebook photos",
			err,
		))
	}

	log.Printf("[Facebook] Successfully fetched %d photos", len(photos))

	// Return success response
	return c.JSON(models.NewSuccessResponse(
		fmt.Sprintf("Successfully fetched %d photo(s)", len(photos)),
		photos,
	))
}
