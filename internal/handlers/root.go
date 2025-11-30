package handlers

import (
	"github.com/gofiber/fiber/v2"

	"twitter-down/internal/models"
)

func Root(c *fiber.Ctx) error {
	endpoints := []string{
		"GET /health - Health check",
		"GET /api/v1/twitter?url={tweet_url} - Download Twitter images",
		"GET /api/v1/facebook?url={fb_url} - Download Facebook images",
		"GET /api/v1/pinterest?url={pin_url} - Download Pinterest images",
		"GET /api/v1/instagram?url={ig_url} - Download Instagram images",
		"GET /api/v1/generic?url={image_url} - Download any image",
		"GET /api/v1/proxy/image?imageUrl={image_url} - Proxy Pinterest image to avoid CORS",
		"GET /api/v1/resolve/pinterest?url={pinterest_short_url} - Resolve Pinterest short URL",
	}

	return c.JSON(models.NewSuccessResponse(
		"Social Media Downloader API v2.0",
		fiber.Map{
			"endpoints": endpoints,
			"docs":      "https://github.com/ilhmlnaa/go-social-dls",
		},
	))
}
