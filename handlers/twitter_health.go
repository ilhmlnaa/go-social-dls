package handlers

import (
	"log"
	"net/http"
	"os"

	"twitter-down/utils"

	twitterscraper "github.com/n0madic/twitter-scraper"
)

// TwitterHealthCheckHandler checks if Twitter credentials are valid
func TwitterHealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authToken := os.Getenv("TWITTER_AUTH_TOKEN")
		csrfToken := os.Getenv("TWITTER_CSRF_TOKEN")
		guestID := os.Getenv("TWITTER_GUEST_ID")

		// Check if credentials are set
		if authToken == "" || csrfToken == "" {
			log.Println("INFO: Twitter credentials not configured")
			utils.JSONResponse(w, false, "Twitter credentials not configured", map[string]interface{}{
				"configured": false,
				"hint":       "Set TWITTER_AUTH_TOKEN and TWITTER_CSRF_TOKEN environment variables",
			})
			return
		}

		// Setup scraper with cookies
		scraper := twitterscraper.New()
		
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
		
		if guestID != "" {
			cookies = append(cookies, &http.Cookie{
				Name:   "guest_id",
				Value:  guestID,
				Domain: ".twitter.com",
			})
		}
		
		scraper.SetCookies(cookies)

		// Test with a simple request - try to get a known public tweet
		// Using Twitter's own tweet as test: https://twitter.com/Twitter/status/20
		testTweetID := "20" // First tweet ever by @Twitter

		_, err := scraper.GetTweet(testTweetID)
		if err != nil {
			log.Printf("WARNING: Twitter credentials validation failed: %v", err)
			utils.JSONResponse(w, false, "Twitter credentials appear to be invalid or expired", map[string]interface{}{
				"configured": true,
				"valid":      false,
				"error":      err.Error(),
				"hint":       "Please update your TWITTER_AUTH_TOKEN and TWITTER_CSRF_TOKEN with fresh values from your browser cookies",
			})
			return
		}

		log.Println("SUCCESS: Twitter credentials are valid")
		utils.JSONResponse(w, true, "Twitter credentials are valid and working", map[string]interface{}{
			"configured": true,
			"valid":      true,
		})
	}
}
