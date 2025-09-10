package handlers

import (
	"fmt"
	"net/http"

	"twitter-down/utils"

	"github.com/PuerkitoBio/goquery"
)

func PinterestDownloadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlPin := r.URL.Query().Get("url")
		if urlPin == "" {
			utils.JSONResponse(w, false, "Parameter 'url' dibutuhkan", nil)
			return
		}

		resp, err := http.Get(urlPin)
		if err != nil {
			utils.JSONResponse(w, false, fmt.Sprintf("Gagal request: %v", err), nil)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			utils.JSONResponse(w, false, fmt.Sprintf("Status code %d dari Pinterest", resp.StatusCode), nil)
			return
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			utils.JSONResponse(w, false, fmt.Sprintf("Gagal parsing HTML: %v", err), nil)
			return
		}

		imgURL, exists := doc.Find(`meta[property="og:image"]`).Attr("content")
		if !exists || imgURL == "" {
			utils.JSONResponse(w, false, "Tidak ditemukan gambar di pin ini", nil)
			return
		}

		utils.JSONResponse(w, true, "Berhasil mengambil gambar", []string{imgURL})
	}
}
