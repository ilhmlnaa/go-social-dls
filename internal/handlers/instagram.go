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

	all := c.Query("all") == "true"
	cfg := config.Load()

	log.Printf("[Instagram] Initializing HTTP service for: %s (all=%v)", urlIG, all)
	igSvc, err := services.NewInstagramHTTPService(cfg.CookiesDir)
	if err != nil {
		log.Printf("[Instagram] Failed to initialize HTTP service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to initialize Instagram service. Make sure cookies are configured.",
			err,
		))
	}

	images, err := igSvc.GetPostImages(urlIG, all)
	if err != nil {
		log.Printf("[Instagram] HTTP extraction failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to fetch Instagram images",
			err,
		))
	}

	log.Printf("[Instagram] Successfully fetched %d image(s)", len(images))

	proxiedImages := make([]string, len(images))
	for i, imgURL := range images {
		proxiedImages[i] = fmt.Sprintf("%s/api/v1/proxy/image?imageUrl=%s", cfg.BaseURL, imgURL)
	}

	message := "Images retrieved successfully"
	if all {
		message = "All image variants retrieved successfully"
	}

	return c.JSON(models.NewSuccessResponse(
		message,
		map[string]interface{}{
			"count":  len(images),
			"images": images,
			"source": "instagram",
			"mode":   map[bool]string{true: "all", false: "top3"}[all],
		},
	))
}
