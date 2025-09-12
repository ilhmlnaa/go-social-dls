package proxy

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

func ImageProxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		imageUrl := r.URL.Query().Get("imageUrl")
		if imageUrl == "" {
			http.Error(w, "imageUrl parameter is required", http.StatusBadRequest)
			return
		}

		if !strings.HasPrefix(imageUrl, "https://i.pinimg.com/") {
			http.Error(w, "Only pinimg.com is allowed", http.StatusForbidden)
			return
		}

		_, err := url.ParseRequestURI(imageUrl)
		if err != nil {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		resp, err := http.Get(imageUrl)
		if err != nil || resp.StatusCode != 200 {
			http.Error(w, "Failed to fetch image", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.Header().Set("Cache-Control", "public, max-age=86400")

		io.Copy(w, resp.Body)
	}
}
