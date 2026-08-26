package handlers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"twitter-down/internal/models"
	"twitter-down/internal/services"
)

// DanbooruDownload mengambil gambar dari sebuah post Danbooru lewat API resmi.
//
// Query params:
//
//	url      : URL post (https://danbooru.donmai.us/posts/12039889) ATAU id mentah. Wajib.
//	children : "true" untuk menyertakan induk + semua child posts dalam satu grup.
func DanbooruDownload(c *fiber.Ctx) error {
	input := c.Query("url")
	if input == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Parameter 'url' is required (URL post atau id)",
			nil,
		))
	}

	includeChildren := c.Query("children") == "true"

	svc := services.NewDanbooruService()
	result, err := svc.GetPostImages(input, includeChildren)
	if err != nil {
		log.Printf("[Danbooru] Failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Failed to fetch Danbooru post",
			err,
		))
	}

	msg := fmt.Sprintf("Berhasil mengambil %d gambar dari Danbooru", result.Count)
	if result.HasChildren && !includeChildren {
		msg += " (post ini punya child posts — tambahkan &children=true untuk mengambil semuanya)"
	}

	log.Printf("[Danbooru] post %d: %d image(s), children=%v", result.PostID, result.Count, includeChildren)

	return c.JSON(models.NewSuccessResponse(msg, fiber.Map{
		"post_id":       result.PostID,
		"has_children":  result.HasChildren,
		"count":         result.Count,
		"with_children": result.WithChildren,
		"images":        result.Images,
		"source":        "danbooru",
	}))
}
