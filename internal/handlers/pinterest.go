package handlers

import (
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"

	"twitter-down/internal/models"
)

func PinterestDownload(c *fiber.Ctx) error {
	urlPin := c.Query("url")
	if urlPin == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Parameter 'url' is required",
			nil,
		))
	}

	resp, err := http.Get(urlPin)
	if err != nil {
		log.Printf("[Pinterest] Failed to fetch URL: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to fetch Pinterest URL",
			err,
		))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.Status(resp.StatusCode).JSON(models.NewErrorResponse(
			"Failed to fetch Pinterest page",
			nil,
		))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("[Pinterest] Failed to parse HTML: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to parse Pinterest page",
			err,
		))
	}

	imgURL, exists := doc.Find(`meta[property="og:image"]`).Attr("content")
	if !exists || imgURL == "" {
		return c.Status(fiber.StatusNotFound).JSON(models.NewErrorResponse(
			"No image found in Pinterest pin",
			nil,
		))
	}

	log.Printf("[Pinterest] Successfully fetched image from pin")

	return c.JSON(models.NewSuccessResponse(
		"Images retrieved successfully",
		map[string]interface{}{
			"count":  1,
			"images": []string{imgURL},
			"source": "pinterest",
		},
	))
}
