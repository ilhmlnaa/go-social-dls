package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"

	"twitter-down/internal/models"
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

	resp, err := http.Get(urlIG)
	if err != nil {
		log.Printf("[Instagram] Failed to fetch URL: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to fetch Instagram URL",
			err,
		))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.Status(resp.StatusCode).JSON(models.NewErrorResponse(
			"Failed to fetch Instagram page",
			nil,
		))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("[Instagram] Failed to parse HTML: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to parse Instagram page",
			err,
		))
	}

	// Try savefrom-helper first
	if href, exists := doc.Find(`a.savefrom-helper--btn`).Attr("href"); exists && href != "" {
		log.Printf("[Instagram] Found image via savefrom-helper")
		return c.JSON(models.NewSuccessResponse(
			"Successfully fetched full resolution image (savefrom)",
			[]string{href},
		))
	}

	// Fallback to og:image
	if ogImg, exists := doc.Find(`meta[property="og:image"]`).Attr("content"); exists && ogImg != "" {
		log.Printf("[Instagram] Found image via og:image")
		return c.JSON(models.NewSuccessResponse(
			"Successfully fetched image (og:image)",
			[]string{ogImg},
		))
	}

	// Fallback to ld+json
	var foundImages []string
	doc.Find(`script[type="application/ld+json"]`).Each(func(i int, s *goquery.Selection) {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(s.Text()), &data); err == nil {
			if imgVal, ok := data["image"]; ok {
				switch v := imgVal.(type) {
				case string:
					foundImages = append(foundImages, v)
				case []interface{}:
					for _, item := range v {
						if str, ok := item.(string); ok {
							foundImages = append(foundImages, str)
						}
					}
				}
			}
		}
	})

	if len(foundImages) > 0 {
		log.Printf("[Instagram] Found %d image(s) via ld+json", len(foundImages))
		return c.JSON(models.NewSuccessResponse(
			"Successfully fetched image (ld+json)",
			foundImages,
		))
	}

	// No images found
	log.Printf("[Instagram] No images found for URL: %s", urlIG)
	return c.Status(fiber.StatusNotFound).JSON(models.NewErrorResponse(
		"No images found in Instagram post",
		nil,
	))
}
