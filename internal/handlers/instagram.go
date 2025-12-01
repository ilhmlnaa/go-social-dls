package handlers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"twitter-down/internal/config"
	"twitter-down/internal/models"
	"twitter-down/internal/services"
)

func InstagramDownload(c *fiber.Ctx) error {
	urlIG := c.Query("url")
	if urlIG == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Parameter 'url' is required",
			nil,
		))
	}

	cfg := config.Load()

	log.Printf("[Instagram] Initializing browser service for: %s", urlIG)
	igSvc, err := services.NewInstagramBrowserService(cfg.CookiesDir)
	if err != nil {
		log.Printf("[Instagram] Failed to initialize browser service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to initialize Instagram service. Make sure cookies are configured.",
			err,
		))
	}

	images, err := igSvc.GetPostImages(urlIG)
	if err != nil {
		log.Printf("[Instagram] Browser extraction failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to fetch Instagram images using browser",
			err,
		))
	}

	log.Printf("[Instagram] Successfully fetched %d full resolution image(s)", len(images))

	proxiedImages := make([]string, len(images))
	for i, imgURL := range images {
		proxiedImages[i] = fmt.Sprintf("%s/api/v1/proxy/image?imageUrl=%s", cfg.BaseURL, imgURL)
	}

	return c.JSON(models.NewSuccessResponse(
		"Images retrieved successfully",
		map[string]interface{}{
			"count":  len(proxiedImages),
			"images": proxiedImages,
			"source": "instagram",
		},
	))
}
