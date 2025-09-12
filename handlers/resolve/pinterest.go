package resolve

import (
	"net/http"
	"strings"

	"twitter-down/utils"
)

type resolveResponse struct {
	URL string `json:"url"`
}

func ResolvePinterestUrl() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawUrl := r.URL.Query().Get("url")
		if rawUrl == "" {
			utils.JSONResponse(w, false, "Parameter 'url' dibutuhkan", nil)
			return
		}

		if !strings.Contains(rawUrl, "pin.it") && !strings.Contains(rawUrl, "pinterest.com") {
			utils.JSONResponse(w, false, "Hanya URL Pinterest yang diizinkan", nil)
			return
		}

		finalUrl, err := followRedirect(rawUrl)
		if err != nil {
			utils.JSONResponse(w, false, "Gagal menyelesaikan URL", nil)
			return
		}

		doubleFinalUrl, err := followRedirect(finalUrl)
		if err == nil && doubleFinalUrl != "" {
			finalUrl = doubleFinalUrl
		}

		resp := resolveResponse{URL: finalUrl}
		utils.JSONResponse(w, true, "Url Berhasil di Resolve", resp)
	}
}

func followRedirect(rawUrl string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("HEAD", rawUrl, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if location == "" {
		return rawUrl, nil
	}
	return location, nil
}
