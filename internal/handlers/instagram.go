package handlers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"twitter-down/internal/config"
	"twitter-down/internal/models"
	"twitter-down/internal/services"
)

// InstagramDownload handles Instagram image download requests
func InstagramDownload(c *fiber.Ctx) error {
	urlIG := c.Query("url")
	if urlIG == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Parameter 'url' is required",
			nil,
		))
	}

	// Load config
	cfg := config.Load()

	// Initialize Instagram service
	igSvc, err := services.NewInstagramService(cfg.CookiesDir)
	if err != nil {
		log.Printf("[Instagram] Failed to initialize service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to initialize Instagram service. Make sure cookies are configured.",
			err,
		))
	}

	// Try GraphQL API first
	images, err := igSvc.GetPostImages(urlIG)
	if err != nil {
		log.Printf("[Instagram] GraphQL method failed, trying HTML fallback: %v", err)
		
		// Fallback to HTML parsing
		images, err = igSvc.GetPostImagesHTML(urlIG)
		if err != nil {
			log.Printf("[Instagram] HTML fallback also failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
				"Failed to fetch Instagram images",
				err,
			))
		}
	}

	log.Printf("[Instagram] Successfully fetched %d image(s)", len(images))

	return c.JSON(models.NewSuccessResponse(
		fmt.Sprintf("Successfully fetched %d image(s)", len(images)),
		images,
	))
}
