package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"twitter-down/utils"

	"github.com/PuerkitoBio/goquery"
)

func FacebookDownloadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlFB := r.URL.Query().Get("url")
		if urlFB == "" {
			utils.JSONResponse(w, false, "Parameter 'url' dibutuhkan", nil)
			return
		}

		client := &http.Client{}
		req, err := http.NewRequest("GET", urlFB, nil)
		if err != nil {
			utils.JSONResponse(w, false, fmt.Sprintf("Gagal membuat request: %v", err), nil)
			return
		}

		// Header biar mirip browser
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		req.Header.Set("Referer", "https://www.facebook.com/")

		// Cookie login Facebook
		cookie := "c_user=100072573614715; xs=24%3AEqJM4kZU5nXkQQ%3A2%3A1757693292%3A-1%3A-1;"
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}

		resp, err := client.Do(req)
		if err != nil {
			utils.JSONResponse(w, false, fmt.Sprintf("Gagal request: %v", err), nil)
			return
		}
		defer resp.Body.Close()

		// 🛠️ Debug: baca isi body (untuk log)
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		fmt.Println("🔍 [DEBUG] Status code:", resp.StatusCode)
		fmt.Println("🔍 [DEBUG] Response body (awal):")
		fmt.Println(bodyStr[:500]) // hanya cetak 500 karakter pertama biar ga banjir log

		// kalau status code != 200 langsung balikan dengan debug info
		if resp.StatusCode != 200 {
			utils.JSONResponse(w, false, fmt.Sprintf("Status code %d dari Facebook. Body: %.200s", resp.StatusCode, bodyStr), nil)
			return
		}

		// karena body sudah dibaca, harus buat ulang reader
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyStr))
		if err != nil {
			utils.JSONResponse(w, false, fmt.Sprintf("Gagal parsing HTML: %v", err), nil)
			return
		}

		var urls []string
		doc.Find("img").Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists && strings.Contains(src, "scontent") {
				urls = append(urls, src)
			}
		})

		if len(urls) > 0 {
			utils.JSONResponse(w, true, "Berhasil menemukan gambar Facebook", urls)
			return
		}

		if ogImg, exists := doc.Find(`meta[property="og:image"]`).Attr("content"); exists && ogImg != "" {
			utils.JSONResponse(w, true, "Berhasil mengambil gambar dari og:image", []string{ogImg})
			return
		}

		utils.JSONResponse(w, false, "Tidak menemukan gambar di Facebook", nil)
	}
}
