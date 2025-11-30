package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"twitter-down/utils"

	"github.com/PuerkitoBio/goquery"
)

func FacebookDownloadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		originalURL := r.URL.Query().Get("url")
		if originalURL == "" {
			utils.JSONResponse(w, false, "Parameter 'url' dibutuhkan", nil)
			return
		}

		// Get Cookie from various sources
		cookieStr := r.URL.Query().Get("cookie")
		if cookieStr == "" {
			cookieStr = r.Header.Get("X-Facebook-Cookie")
		}
		if cookieStr == "" {
			// Try loading from export-cookie.json
			// Use absolute path to be safe
			cookieStr = loadCookieFromFile(`e:\Data Kuliah\Tingkat 3\nganggur\EndGame\media-downloader-api\export-cookie.json`)
		}

		// 🛠️ Fix: Handle unencoded URLs where params like &id=... are treated as API params
		// Reconstruct the full URL by appending other query parameters
		q := r.URL.Query()
		for k, v := range q {
			if k == "url" {
				continue
			}
			separator := "&"
			if !strings.Contains(originalURL, "?") {
				separator = "?"
			}
			originalURL += fmt.Sprintf("%s%s=%s", separator, k, v[0])
		}

		fmt.Println("🔍 [DEBUG] Processing URL:", originalURL)

		// Strategy 1: Try mbasic.facebook.com
		// mbasic prefers /photo.php?fbid=... over /photo/?fbid=...
		mbasicURL := originalURL
		if strings.Contains(mbasicURL, "www.facebook.com") {
			mbasicURL = strings.Replace(mbasicURL, "www.facebook.com", "mbasic.facebook.com", 1)
		} else if strings.Contains(mbasicURL, "facebook.com") && !strings.Contains(mbasicURL, "mbasic.facebook.com") {
			mbasicURL = strings.Replace(mbasicURL, "facebook.com", "mbasic.facebook.com", 1)
		}
		
		// Fix path for mbasic: /photo/ -> /photo.php
		if strings.Contains(mbasicURL, "/photo/") {
			mbasicURL = strings.Replace(mbasicURL, "/photo/", "/photo.php", 1)
		}

		urls, err := scrapeFacebook(mbasicURL, true, cookieStr)
		if err == nil && len(urls) > 0 {
			utils.JSONResponse(w, true, "Berhasil menemukan gambar Facebook (mbasic)", urls)
			return
		}
		
		fmt.Printf("⚠️ [DEBUG] mbasic failed: %v. Retrying with www.facebook.com...\n", err)

		// Strategy 2: Fallback to www.facebook.com (scrape og:image)
		// Use the original URL (ensure it's www)
		wwwURL := originalURL
		if strings.Contains(wwwURL, "mbasic.facebook.com") {
			wwwURL = strings.Replace(wwwURL, "mbasic.facebook.com", "www.facebook.com", 1)
		}
		
		urls, err = scrapeFacebook(wwwURL, false, cookieStr)
		if err == nil && len(urls) > 0 {
			utils.JSONResponse(w, true, "Berhasil menemukan gambar Facebook (www)", urls)
			return
		}

		utils.JSONResponse(w, false, "Tidak menemukan gambar. Pastikan postingan bersifat publik, atau sediakan cookie yang valid via parameter 'cookie' atau file 'export-cookie.json'.", nil)
	}
}

func scrapeFacebook(targetURL string, isMobile bool, cookieStr string) ([]string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	if isMobile {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Mobile Safari/537.36")
		req.Header.Set("Referer", "https://mbasic.facebook.com/")
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
		req.Header.Set("Referer", "https://www.facebook.com/")
	}
	
	if cookieStr != "" {
		req.Header.Set("Cookie", cookieStr)
	}
	
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyStr := string(bodyBytes)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyStr))
	if err != nil {
		return nil, err
	}

	var urls []string
	
	// 1. Try finding direct images (common in mbasic)
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists {
			if strings.Contains(src, "fbcdn.net") || strings.Contains(src, "scontent") {
				if !strings.Contains(src, "emoji.php") && !strings.Contains(src, "static.xx") && !strings.Contains(src, "_nc_cat=1") { 
					urls = append(urls, src)
				}
			}
		}
	})

	// 2. Try og:image (common in www)
	if len(urls) == 0 {
		if ogImg, exists := doc.Find(`meta[property="og:image"]`).Attr("content"); exists && ogImg != "" {
			urls = append(urls, ogImg)
		}
	}

	// 3. Try twitter:image
	if len(urls) == 0 {
		if twImg, exists := doc.Find(`meta[name="twitter:image"]`).Attr("content"); exists && twImg != "" {
			urls = append(urls, twImg)
		}
	}

	// 4. Regex Fallback (for images inside scripts/JSON)
	if len(urls) == 0 {
		// Pattern to find fbcdn urls. 
		// Matches https://...fbcdn.net/... .jpg or similar
		// We look for strings starting with https and containing fbcdn.net
		// Be careful with escaped slashes in JSON
		
		// Regex to find URLs that contain fbcdn.net and end with .jpg (roughly)
		// We look for "https?://[^"]*fbcdn\.net[^"'\s]*\.jpg[^"'\s]*"
		// Also handle escaped slashes \/
		
		re := regexp.MustCompile(`https?:\\?\/\\?\/[^"'\s]*fbcdn\.net[^"'\s]*\.jpg[^"'\s]*`)
		matches := re.FindAllString(bodyStr, -1)
		
		for _, match := range matches {
			// Clean up escaped slashes
			cleanURL := strings.ReplaceAll(match, `\/`, `/`)
			// Remove any trailing backslashes or quotes if regex caught them (it shouldn't with [^"'\s])
			
			// Filter out small icons if possible
			if !strings.Contains(cleanURL, "emoji.php") && !strings.Contains(cleanURL, "static.xx") && !strings.Contains(cleanURL, "_nc_cat=1") {
				urls = append(urls, cleanURL)
			}
		}
	}

	if len(urls) > 0 {
		return urls, nil
	}

	fmt.Printf("⚠️ [DEBUG] No images found for %s. Body start: %.500s\n", targetURL, bodyStr)
	return nil, fmt.Errorf("no images found")
}

func loadCookieFromFile(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("⚠️ [DEBUG] Failed to open cookie file %s: %v\n", filename, err)
		return ""
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return ""
	}

	// Try parsing as JSON array (EditThisCookie format)
	var cookies []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(bytes, &cookies); err == nil {
		var cookieParts []string
		for _, c := range cookies {
			cookieParts = append(cookieParts, fmt.Sprintf("%s=%s", c.Name, c.Value))
		}
		fmt.Printf("🔍 [DEBUG] Loaded %d cookies from %s\n", len(cookies), filename)
		return strings.Join(cookieParts, "; ")
	}

	// If not JSON, maybe it's just a raw string
	return string(bytes)
}
