package handlers

import (
	"bytes"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/gofiber/fiber/v2"

	"twitter-down/internal/models"
	"twitter-down/utils"
)

func getFilenameFromResponse(resp *http.Response, imageURL string) string {
	contentDisposition := resp.Header.Get("Content-Disposition")

	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			if filename, ok := params["filename"]; ok && filename != "" {
				return filename
			}
		}

		if idx := strings.Index(strings.ToLower(contentDisposition), "filename="); idx != -1 {
			filename := contentDisposition[idx+len("filename="):]
			filename = strings.Trim(filename, "\"'; ")
			filename = strings.Split(filename, ";")[0]
			filename = strings.Trim(filename, "\"'; ")
			if filename != "" {
				return filename
			}
		}
	}

	filename := path.Base(strings.Split(imageURL, "?")[0])
	if filename == "" || filename == "." || filename == "/" {
		filename = "downloaded_" + utils.GenerateRandomString(8)
	}

	if !strings.Contains(filename, ".") {
		contentType := strings.ToLower(resp.Header.Get("Content-Type"))
		switch {
		case strings.Contains(contentType, "jpeg"):
			filename += ".jpg"
		case strings.Contains(contentType, "png"):
			filename += ".png"
		case strings.Contains(contentType, "gif"):
			filename += ".gif"
		case strings.Contains(contentType, "webp"):
			filename += ".webp"
		default:
			filename += ".jpg"
		}
	}

	return filename
}

// GenericDownload handles direct image download from URL
func GenericDownload(c *fiber.Ctx) error {
	imageURL := c.Query("url")
	if imageURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Parameter 'url' is required",
			nil,
		))
	}

	if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			"Invalid URL",
			nil,
		))
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		log.Printf("[Generic] Failed to create request: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to create request",
			err,
		))
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Generic] Failed to download image: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			"Failed to download image",
			err,
		))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.Status(resp.StatusCode).JSON(models.NewErrorResponse(
			"Failed to download image",
			nil,
		))
	}

	buf := make([]byte, 512)
	n, _ := io.ReadFull(resp.Body, buf)
	reader := io.MultiReader(bytes.NewReader(buf[:n]), resp.Body)

	filename := getFilenameFromResponse(resp, imageURL)

	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Set("Content-Type", resp.Header.Get("Content-Type"))

	return c.SendStream(reader)
}
