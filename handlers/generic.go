package handlers

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"

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

		if strings.Contains(strings.ToLower(contentDisposition), "filename") {
			parts := strings.Split(contentDisposition, "filename")
			if len(parts) > 1 {
				part := parts[1]
				part = strings.TrimLeft(part, "=*")
				part = strings.Trim(part, "\"'; ")
				part = strings.Split(part, ";")[0]
				part = strings.Trim(part, "\"'; ")
				if part != "" && !strings.Contains(part, "=") {
					return part
				}
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

func GenericDownloadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		imageURL := r.URL.Query().Get("url")
		if imageURL == "" {
			utils.JSONResponse(w, false, "Parameter 'url' dibutuhkan", nil)
			return
		}

		if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
			utils.JSONResponse(w, false, "URL tidak valid", nil)
			return
		}

		client := &http.Client{}
		req, err := http.NewRequest("GET", imageURL, nil)
		if err != nil {
			utils.JSONResponse(w, false, "Gagal membuat request: "+err.Error(), nil)
			return
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		
		resp, err := client.Do(req)
		if err != nil {
			utils.JSONResponse(w, false, "Gagal mengunduh gambar: "+err.Error(), nil)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			utils.JSONResponse(w, false, "Gagal mengunduh gambar, status code: "+http.StatusText(resp.StatusCode), nil)
			return
		}

		buf := make([]byte, 512)
		n, _ := io.ReadFull(resp.Body, buf)
		reader := io.MultiReader(bytes.NewReader(buf[:n]), resp.Body)

		filename := getFilenameFromResponse(resp, imageURL)

		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		io.Copy(w, reader)
	}
}
