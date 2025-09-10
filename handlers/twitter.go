package handlers

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"twitter-down/utils"

	twitterscraper "github.com/imperatrona/twitter-scraper"
)

func TwitterDownloadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		authToken := os.Getenv("TWITTER_AUTH_TOKEN")
		csrfToken := os.Getenv("TWITTER_CSRF_TOKEN")

		if authToken == "" || csrfToken == "" {
			panic("TWITTER_AUTH_TOKEN dan TWITTER_CSRF_TOKEN harus di-set di environment")
		}

		scraper := twitterscraper.New()
		scraper.SetAuthToken(twitterscraper.AuthToken{
			Token:     authToken,
			CSRFToken: csrfToken,
		})

		if !scraper.IsLoggedIn() {
			panic("AuthToken tidak valid")
		}

		urlTweet := r.URL.Query().Get("url")
		if urlTweet == "" {
			utils.JSONResponse(w, false, "Parameter 'url' dibutuhkan", nil)
			return
		}

		tweetID, err := extractTweetID(urlTweet)
		if err != nil {
			utils.JSONResponse(w, false, "URL tweet tidak valid", nil)
			return
		}

		tweet, err := scraper.GetTweet(tweetID)
		if err != nil {
			utils.JSONResponse(w, false, fmt.Sprintf("Gagal mengambil tweet: %v", err), nil)
			return
		}

		if len(tweet.Photos) == 0 {
			utils.JSONResponse(w, false, "Tweet tidak mengandung gambar", nil)
			return
		}

		var urls []string
		for _, photo := range tweet.Photos {
			imgURL := strings.Replace(photo.URL, "&name=small", "&name=large", 1)
			urls = append(urls, imgURL)
		}

		utils.JSONResponse(w, true, "Berhasil mengambil gambar", urls)
	}
}

func extractTweetID(url string) (string, error) {
	re := regexp.MustCompile(`status/(\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("tidak ditemukan tweet id")
	}
	return matches[1], nil
}
