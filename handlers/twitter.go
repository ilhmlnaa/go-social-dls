package handlers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"twitter-down/utils"

	twitterscraper "github.com/n0madic/twitter-scraper"
)

func TwitterDownloadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validasi environment variables
		authToken := os.Getenv("TWITTER_AUTH_TOKEN")
		csrfToken := os.Getenv("TWITTER_CSRF_TOKEN")
		
		// Optional: guest_id untuk membantu autentikasi
		guestID := os.Getenv("TWITTER_GUEST_ID")

		if authToken == "" || csrfToken == "" {
			log.Println("ERROR: TWITTER_AUTH_TOKEN atau TWITTER_CSRF_TOKEN belum di-set di environment")
			utils.JSONResponse(w, false, "Twitter credentials belum dikonfigurasi. Silakan set TWITTER_AUTH_TOKEN dan TWITTER_CSRF_TOKEN.", nil)
			return
		}

		// Validasi URL parameter
		urlTweet := r.URL.Query().Get("url")
		if urlTweet == "" {
			utils.JSONResponse(w, false, "Parameter 'url' dibutuhkan", nil)
			return
		}

		// Extract tweet ID
		tweetID, err := extractTweetID(urlTweet)
		if err != nil {
			log.Printf("ERROR: Invalid tweet URL: %s - %v", urlTweet, err)
			utils.JSONResponse(w, false, "URL tweet tidak valid. Format yang benar: https://twitter.com/username/status/[tweet_id]", nil)
			return
		}

		// Setup scraper dengan login
		scraper := twitterscraper.New()
		
		// n0madic package menggunakan cookies untuk login
		// Buat cookies array dengan semua cookies yang diperlukan
		cookies := []*http.Cookie{
			{
				Name:   "auth_token",
				Value:  authToken,
				Domain: ".twitter.com",
			},
			{
				Name:   "ct0",
				Value:  csrfToken,
				Domain: ".twitter.com",
			},
		}
		
		// Tambahkan guest_id jika ada
		if guestID != "" {
			cookies = append(cookies, &http.Cookie{
				Name:   "guest_id",
				Value:  guestID,
				Domain: ".twitter.com",
			})
		}
		
		// Set cookies ke scraper
		scraper.SetCookies(cookies)
		
		// Verify login status
		if !scraper.IsLoggedIn() {
			log.Println("WARNING: Scraper reports not logged in, but will try to proceed anyway")
		}

		log.Printf("INFO: Attempting to fetch tweet ID: %s", tweetID)

		// Retry logic dengan exponential backoff
		var tweet *twitterscraper.Tweet
		maxRetries := 3
		var lastErr error

		for i := 0; i < maxRetries; i++ {
			if i > 0 {
				waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
				log.Printf("INFO: Retry attempt %d/%d after %v", i+1, maxRetries, waitTime)
				time.Sleep(waitTime)
			}

			tweet, lastErr = scraper.GetTweet(tweetID)
			if lastErr == nil {
				log.Printf("SUCCESS: Tweet fetched successfully on attempt %d", i+1)
				break
			}

			log.Printf("WARNING: Attempt %d failed: %v", i+1, lastErr)
		}

		// Handle error setelah semua retry
		if lastErr != nil {
			log.Printf("ERROR: Failed to get tweet after %d retries: %v", maxRetries, lastErr)
			errorMsg := fmt.Sprintf("Gagal mengambil tweet setelah %d percobaan. Error: %v. ", maxRetries, lastErr)
			
			// Tambahkan hint berdasarkan jenis error
			if strings.Contains(lastErr.Error(), "401") || strings.Contains(lastErr.Error(), "403") {
				errorMsg += "Kemungkinan token tidak valid atau sudah expired. Silakan perbarui TWITTER_AUTH_TOKEN dan TWITTER_CSRF_TOKEN."
			} else if strings.Contains(lastErr.Error(), "429") {
				errorMsg += "Rate limit tercapai. Silakan coba lagi dalam beberapa menit."
			} else if strings.Contains(lastErr.Error(), "404") {
				errorMsg += "Tweet tidak ditemukan atau sudah dihapus."
			} else {
				errorMsg += "Pastikan token valid dan akun tidak ter-suspend."
			}

			utils.JSONResponse(w, false, errorMsg, nil)
			return
		}

		// Validasi tweet memiliki photos
		if len(tweet.Photos) == 0 {
			log.Printf("INFO: Tweet %s tidak mengandung gambar", tweetID)
			utils.JSONResponse(w, false, "Tweet tidak mengandung gambar", nil)
			return
		}

		// Extract photo URLs dengan kualitas tinggi
		var urls []string
		for _, photo := range tweet.Photos {
			imgURL := strings.Replace(photo.URL, "&name=small", "&name=large", 1)
			imgURL = strings.Replace(imgURL, "&name=medium", "&name=large", 1)
			urls = append(urls, imgURL)
		}

		log.Printf("SUCCESS: Successfully extracted %d images from tweet %s", len(urls), tweetID)
		utils.JSONResponse(w, true, fmt.Sprintf("Berhasil mengambil %d gambar", len(urls)), urls)
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
