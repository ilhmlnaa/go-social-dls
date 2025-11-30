package handlers

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"twitter-down/internal/models"
)

func ImageProxy(c *fiber.Ctx) error {
	imageUrl := c.Query("imageUrl")
	if imageUrl == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Parameter 'imageUrl' is required",
			nil,
		))
	}

	if !strings.HasPrefix(imageUrl, "https://i.pinimg.com/") {
		return c.Status(fiber.StatusForbidden).JSON(models.NewErrorResponse(
			"Only i.pinimg.com images are allowed",
			nil,
		))
	}

	_, err := url.ParseRequestURI(imageUrl)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Invalid URL format",
			err,
		))
	}

	resp, err := http.Get(imageUrl)
	if err != nil {
		log.Printf("[Proxy] Failed to fetch image: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(models.NewErrorResponse(
			"Failed to fetch image",
			err,
		))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.Status(resp.StatusCode).JSON(models.NewErrorResponse(
			"Failed to fetch image from source",
			nil,
		))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Proxy] Failed to read image body: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(models.NewErrorResponse(
			"Failed to read image",
			err,
		))
	}

	c.Set("Content-Type", resp.Header.Get("Content-Type"))
	c.Set("Cache-Control", "public, max-age=86400")
	c.Set("Content-Length", strconv.Itoa(len(body)))

	return c.Send(body)
}
