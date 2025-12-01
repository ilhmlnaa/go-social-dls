package handlers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"twitter-down/internal/config"
	"twitter-down/internal/models"
	"twitter-down/internal/services"
)

func PixivDownload(c *fiber.Ctx) error {
	url := c.Query("url")
	if url == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"URL parameter is required",
			nil,
		))
	}

	cfg := config.Load()
	pixivService, err := services.NewPixivService(cfg.CookiesDir)
	if err != nil {
		log.Printf("Failed to initialize Pixiv service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to initialize Pixiv service. Please check cookies configuration.",
			err,
		))
	}

	images, err := pixivService.GetIllustrationImages(url)
	if err != nil {
		log.Printf("Failed to get Pixiv images: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Failed to fetch images",
			err,
		))
	}

	if len(images) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(models.NewErrorResponse(
			"No images found in the provided Pixiv URL",
			nil,
		))
	}

	var proxyImages []string
	for _, imageURL := range images {
		proxyURL := fmt.Sprintf("%s/api/v1/proxy/image?url=%s&referer=%s", 
			cfg.BaseURL,
			imageURL, 
			"https://www.pixiv.net/")
		proxyImages = append(proxyImages, proxyURL)
	}

	return c.JSON(models.NewSuccessResponse(
		"Images retrieved successfully",
		map[string]interface{}{
			"count":  len(proxyImages),
			"images": proxyImages,
			"source": "pixiv",
		},
	))
}
