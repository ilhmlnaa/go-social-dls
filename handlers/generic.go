package handlers

import (
	"io"
	"net/http"
	"path"
	"strings"

	"twitter-down/utils"
)

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

		resp, err := http.Get(imageURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			utils.JSONResponse(w, false, "Gagal mengunduh gambar", nil)
			return
		}
		defer resp.Body.Close()

		// Get filename from URL
		filename := path.Base(imageURL)
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))

		io.Copy(w, resp.Body)
	}
}
