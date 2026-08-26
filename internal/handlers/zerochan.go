package handlers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"twitter-down/internal/models"
	"twitter-down/internal/services"
)

// ZerochanDownload mengambil gambar dari sebuah entry Zerochan lewat API resmi.
//
// Query params:
//
//	url : URL entry (https://www.zerochan.net/4675448) ATAU id mentah (4675448). Wajib.
func ZerochanDownload(c *fiber.Ctx) error {
	input := c.Query("url")
	if input == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Parameter 'url' is required (URL entry atau id)",
			nil,
		))
	}

	svc := services.NewZerochanService()
	img, err := svc.GetEntry(input)
	if err != nil {
		log.Printf("[Zerochan] Failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Failed to fetch Zerochan entry",
			err,
		))
	}

	// Kumpulkan URL gambar terurut dari resolusi tertinggi.
	images := []string{}
	if img.Full != "" {
		images = append(images, img.Full)
	}
	if img.Large != "" && img.Large != img.Full {
		images = append(images, img.Large)
	}

	log.Printf("[Zerochan] entry %d: full=%s", img.ID, img.Full)

	return c.JSON(models.NewSuccessResponse(
		fmt.Sprintf("Berhasil mengambil gambar dari Zerochan (entry %d)", img.ID),
		fiber.Map{
			"id":        img.ID,
			"count":     len(images),
			"images":    images,
			"full":      img.Full,
			"large":     img.Large,
			"medium":    img.Medium,
			"small":     img.Small,
			"ext":       img.Ext,
			"file_size": img.FileSize,
			"width":     img.Width,
			"height":    img.Height,
			"source":    img.Source,
			"primary":   img.Primary,
			"tags":      img.Tags,
			"provider":  "zerochan",
		},
	))
}
