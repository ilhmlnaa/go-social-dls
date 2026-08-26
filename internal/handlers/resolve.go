package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"twitter-down/internal/models"
	"twitter-down/internal/utils"
)

func ResolvePinterestUrl(c *fiber.Ctx) error {
	rawUrl := c.Query("url")
	if rawUrl == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Parameter 'url' is required",
			nil,
		))
	}

	if !strings.Contains(rawUrl, "pin.it") && !strings.Contains(rawUrl, "pinterest.com") {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Only Pinterest URLs are allowed",
			nil,
		))
	}

	finalUrl, err := followRedirect(rawUrl)
	if err != nil {
		log.Printf("[Resolve] Failed to resolve URL: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to resolve URL",
			err,
		))
	}

	doubleFinalUrl, err := followRedirect(finalUrl)
	if err == nil && doubleFinalUrl != "" {
		finalUrl = doubleFinalUrl
	}

	log.Printf("[Resolve] Successfully resolved Pinterest URL")

	return c.JSON(models.NewSuccessResponse(
		"URL successfully resolved",
		map[string]interface{}{
			"url": finalUrl,
		},
	))
}

func followRedirect(rawUrl string) (string, error) {
	client := utils.NewHTTPClientWithRedirect("pinterest", 20*time.Second, func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})

	req, err := http.NewRequest("HEAD", rawUrl, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if location == "" {
		return rawUrl, nil
	}
	return location, nil
}
