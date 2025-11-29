package handlers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"twitter-down/internal/config"
	"twitter-down/internal/models"
	"twitter-down/internal/services"
)

func TwitterDownload(c *fiber.Ctx) error {
	urlParam := c.Query("url")
	if urlParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Parameter 'url' is required",
			nil,
		))
	}

	tweetID, err := services.ExtractTweetID(urlParam)
	if err != nil {
		log.Printf("[Twitter] Failed to extract tweet ID from %s: %v", urlParam, err)
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Invalid Twitter URL format",
			err,
		))
	}


	cfg := config.Load()

	twitterSvc, err := services.NewTwitterService(cfg.CookiesDir)
	if err != nil {
		log.Printf("[Twitter] Failed to initialize service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to initialize Twitter service. Make sure cookies are configured.",
			err,
		))
	}

	photos, err := twitterSvc.GetTweetPhotos(tweetID)
	if err != nil {
		log.Printf("[Twitter] Failed to fetch photos for tweet %s: %v", tweetID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to fetch tweet photos",
			err,
		))
	}

	log.Printf("[Twitter] Successfully fetched %d photos for tweet %s", len(photos), tweetID)

	return c.JSON(models.NewSuccessResponse(
		fmt.Sprintf("Successfully fetched %d photo(s)", len(photos)),
		photos,
	))
}
