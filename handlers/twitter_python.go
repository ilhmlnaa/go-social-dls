package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"twitter-down/utils"
)

// TwitterScraperResponse represents the response from Python scraper
type TwitterScraperResponse struct {
	Success    bool     `json:"success"`
	TweetID    string   `json:"tweet_id"`
	Photos     []string `json:"photos"`
	PhotoCount int      `json:"photo_count"`
	TweetText  string   `json:"tweet_text"`
	Error      string   `json:"error"`
	Message    string   `json:"message"`
}

// TwitterDownloadHandlerPython uses Python twscrape to download Twitter media
func TwitterDownloadHandlerPython() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validasi URL parameter
		urlTweet := r.URL.Query().Get("url")
		if urlTweet == "" {
			utils.JSONResponse(w, false, "Parameter 'url' dibutuhkan", nil)
			return
		}

		// Extract tweet ID
		tweetID, err := extractTweetIDPython(urlTweet)
		if err != nil {
			log.Printf("ERROR: Invalid tweet URL: %s - %v", urlTweet, err)
			utils.JSONResponse(w, false, "URL tweet tidak valid. Format yang benar: https://twitter.com/username/status/[tweet_id]", nil)
			return
		}

		log.Printf("INFO: Attempting to fetch tweet ID via Python: %s", tweetID)

		// Retry logic dengan exponential backoff
		var result *TwitterScraperResponse
		maxRetries := 3
		var lastErr error

		for i := 0; i < maxRetries; i++ {
			if i > 0 {
				waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
				log.Printf("INFO: Retry attempt %d/%d after %v", i+1, maxRetries, waitTime)
				time.Sleep(waitTime)
			}

			result, lastErr = callPythonScraper(tweetID)
			if lastErr == nil && result.Success {
				log.Printf("SUCCESS: Tweet fetched successfully via Python on attempt %d", i+1)
				break
			}

			if lastErr != nil {
				log.Printf("WARNING: Attempt %d failed: %v", i+1, lastErr)
			} else if !result.Success {
				log.Printf("WARNING: Attempt %d failed: %s", i+1, result.Message)
				lastErr = fmt.Errorf("%s", result.Message)
			}
		}

		// Handle error setelah semua retry
		if lastErr != nil || !result.Success {
			log.Printf("ERROR: Failed to get tweet after %d retries", maxRetries)
			
			errorMsg := "Gagal mengambil tweet. "
			if result != nil && result.Message != "" {
				errorMsg += result.Message
			} else if lastErr != nil {
				errorMsg += lastErr.Error()
			}
			
			// Tambahkan hint untuk setup jika belum dikonfigurasi
			if result != nil && strings.Contains(result.Message, "No Twitter accounts configured") {
				errorMsg += " Silakan setup Twitter account terlebih dahulu. Lihat TWITTER_TWSCRAPE_SETUP.md"
			}

			utils.JSONResponse(w, false, errorMsg, nil)
			return
		}

		// Validasi ada photos
		if result.PhotoCount == 0 || len(result.Photos) == 0 {
			log.Printf("INFO: Tweet %s tidak mengandung gambar", tweetID)
			utils.JSONResponse(w, false, "Tweet tidak mengandung gambar", nil)
			return
		}

		log.Printf("SUCCESS: Successfully extracted %d images from tweet %s via Python", result.PhotoCount, tweetID)
		utils.JSONResponse(w, true, fmt.Sprintf("Berhasil mengambil %d gambar", result.PhotoCount), result.Photos)
	}
}

// callPythonScraper calls the Python scraper script
func callPythonScraper(tweetID string) (*TwitterScraperResponse, error) {
	// Path to Python script
	scriptPath := "scripts/twitter_scraper.py"
	
	// Execute Python script
	cmd := exec.Command("python3", scriptPath, tweetID)
	
	// Set timeout
	timeout := 30 * time.Second
	timer := time.AfterFunc(timeout, func() {
		cmd.Process.Kill()
	})
	defer timer.Stop()
	
	// Run command and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute Python script: %v, output: %s", err, string(output))
	}
	
	// Parse JSON response
	var result TwitterScraperResponse
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Python script output: %v, output: %s", err, string(output))
	}
	
	return &result, nil
}

// extractTweetIDPython extracts tweet ID from URL
func extractTweetIDPython(url string) (string, error) {
	re := regexp.MustCompile(`status/(\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("tidak ditemukan tweet id")
	}
	return matches[1], nil
}
