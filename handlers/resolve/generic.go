package resolve

import (
	"net/http"

	"twitter-down/utils"
)

func GenericResolveUrl() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortUrl := r.URL.Query().Get("url")
		if shortUrl == "" {
			utils.JSONResponse(w, false, "Parameter 'url' dibutuhkan", nil)
			return
		}

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		req, err := http.NewRequest("HEAD", shortUrl, nil)
		if err != nil {
			utils.JSONResponse(w, false, "Gagal menyelesaikan URL", nil)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			utils.JSONResponse(w, false, "Gagal Resolve URL", nil)
			return
		}
		defer resp.Body.Close()

		finalUrl := resp.Header.Get("Location")
		if finalUrl == "" {
			utils.JSONResponse(w, false, "Tidak ada lokasi redirect ditemukan", nil)
			return
		}

		utils.JSONResponse(w, true, "Url Berhasil di Resolve", map[string]string{"url": finalUrl})
	}
}
