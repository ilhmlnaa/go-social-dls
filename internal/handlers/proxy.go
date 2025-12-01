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
	imageUrl := c.Query("url")
	if imageUrl == "" {
		imageUrl = c.Query("imageUrl") 
	}
	if imageUrl == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Parameter 'url' or 'imageUrl' is required",
			nil,
		))
	}

	referer := c.Query("referer")

	allowedDomains := []string{
		"https://i.pinimg.com/",
		"https://scontent.cdninstagram.com/",
		"https://i.pximg.net/",
	}

	isAllowed := false
	for _, domain := range allowedDomains {
		if strings.HasPrefix(imageUrl, domain) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return c.Status(fiber.StatusForbidden).JSON(models.NewErrorResponse(
			"Image domain not allowed",
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

	req, err := http.NewRequest("GET", imageUrl, nil)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Failed to create request",
			err,
		))
	}

	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
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

