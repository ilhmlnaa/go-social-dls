package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"twitter-down/utils"

	"github.com/PuerkitoBio/goquery"
)

func InstagramDownloadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlIG := r.URL.Query().Get("url")
		if urlIG == "" {
			utils.JSONResponse(w, false, "Parameter 'url' dibutuhkan", nil)
			return
		}

		resp, err := http.Get(urlIG)
		if err != nil {
			utils.JSONResponse(w, false, fmt.Sprintf("Gagal request: %v", err), nil)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			utils.JSONResponse(w, false, fmt.Sprintf("Status code %d dari Instagram", resp.StatusCode), nil)
			return
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			utils.JSONResponse(w, false, fmt.Sprintf("Gagal parsing HTML: %v", err), nil)
			return
		}

		// --- 1. Cek savefrom-helper ---
		if href, exists := doc.Find(`a.savefrom-helper--btn`).Attr("href"); exists && href != "" {
			utils.JSONResponse(w, true, "Berhasil mengambil gambar resolusi penuh (savefrom)", []string{href})
			return
		}

		// --- 2. Fallback og:image ---
		if ogImg, exists := doc.Find(`meta[property="og:image"]`).Attr("content"); exists && ogImg != "" {
			utils.JSONResponse(w, true, "Berhasil mengambil gambar (og:image)", []string{ogImg})
			return
		}

		// --- 3. Fallback ld+json ---
		doc.Find(`script[type="application/ld+json"]`).Each(func(i int, s *goquery.Selection) {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(s.Text()), &data); err == nil {
				if imgVal, ok := data["image"]; ok {
					switch v := imgVal.(type) {
					case string:
						utils.JSONResponse(w, true, "Berhasil mengambil gambar (ld+json)", []string{v})
						return
					case []interface{}:
						var urls []string
						for _, item := range v {
							if str, ok := item.(string); ok {
								urls = append(urls, str)
							}
						}
						if len(urls) > 0 {
							utils.JSONResponse(w, true, "Berhasil mengambil gambar (ld+json)", urls)
							return
						}
					}
				}
			}
		})

		// --- Kalau semua gagal ---
		utils.JSONResponse(w, false, "Tidak menemukan gambar di Instagram", nil)
	}
}
