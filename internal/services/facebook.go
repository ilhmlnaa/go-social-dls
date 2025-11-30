package services

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"twitter-down/internal/utils"
)

type FacebookService struct {
	client  *http.Client
	cookies map[string]string
}

type FacebookGraphQLResponse struct {
	Data struct {
		Node struct {
			CometSections struct{
				Content struct {
					Story struct {
						Message struct {
							Text string `json:"text"`
						} `json:"message"`
						Attachments []struct {
							Styles struct {
								Attachment struct {
									Media struct {
										Image struct {
											URI string `json:"uri"`
										} `json:"image"`
										LargeImage struct {
											URI string `json:"uri"`
										} `json:"largeImage,omitempty"`
									} `json:"media"`
								} `json:"attachment"`
							} `json:"styles"`
						} `json:"attachments"`
					} `json:"story"`
				} `json:"content"`
			} `json:"comet_sections"`
		} `json:"node"`
	} `json:"data"`
}

func NewFacebookService(cookiesDir string) (*FacebookService, error) {
	cookies, err := utils.LoadCookies(cookiesDir, "facebook")
	if err != nil {
		return nil, err
	}

	required := []string{"c_user", "xs"}
	if err := utils.ValidateCookies(cookies, required); err != nil {
		return nil, err
	}

	return &FacebookService{
		client:  &http.Client{},
		cookies: cookies,
	}, nil
}

func (s *FacebookService) NormalizeFacebookURL(fbURL string) (string, error) {
	u, err := url.Parse(fbURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Format 1: /photo?fbid=... (already in correct format)
	if strings.Contains(u.Path, "/photo") && u.Query().Get("fbid") != "" {
		return fbURL, nil
	}

	// Format 2: /share/p/... (modern share format)
	if strings.Contains(u.Path, "/share/p/") {
		return s.resolveShareURL(fbURL)
	}

	// Format 3: /share/... (other share formats)
	if strings.Contains(u.Path, "/share/") {
		return s.resolveShareURL(fbURL)
	}

	// Format 4: /permalink.php?story_fbid=... (old permalink format)
	if strings.Contains(u.Path, "/permalink.php") {
		return s.resolvePermalinkURL(fbURL)
	}

	// Format 5: /posts/... or /photos/...
	if strings.Contains(u.Path, "/posts/") || strings.Contains(u.Path, "/photos/") {
		return s.resolveShareURL(fbURL)
	}

	return fbURL, nil
}

func (s *FacebookService) resolveShareURL(shareURL string) (string, error) {
	req, err := http.NewRequest("GET", shareURL, nil)
	if err != nil {
		return "", err
	}

	for name, value := range s.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	htmlStr := string(body)

	// Try multiple patterns to find fbid
	patterns := []string{
		`/photo\?fbid=(\d+)&amp;set=([^"]+)`,
		`/photo/\?fbid=(\d+)&amp;set=([^"]+)`,
		`"fbid":"(\d+)"`,
		`fbid=(\d+)`,
		`photo_id=(\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlStr)
		if len(matches) >= 2 {
			fbid := matches[1]
			// If we also got the set parameter
			if len(matches) >= 3 {
				set := strings.ReplaceAll(matches[2], "&amp;", "&")
				return fmt.Sprintf("https://www.facebook.com/photo?fbid=%s&set=%s", fbid, set), nil
			}
			// Just fbid
			return fmt.Sprintf("https://www.facebook.com/photo?fbid=%s", fbid), nil
		}
	}

	// If we can't find fbid, return original URL (will be used directly)
	return shareURL, nil
}

func (s *FacebookService) resolvePermalinkURL(permalinkURL string) (string, error) {
	req, err := http.NewRequest("GET", permalinkURL, nil)
	if err != nil {
		return "", err
	}

	for name, value := range s.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	htmlStr := string(body)

	// Find the actual photo link in the HTML
	// Pattern: <a href="https://www.facebook.com/photo/?fbid=...&set=...
	// Use simpler pattern to match any photo URL
	photoLinkRe := regexp.MustCompile(`href="(https://www\.facebook\.com/photo/[^"]+)"`)
	matches := photoLinkRe.FindStringSubmatch(htmlStr)
	if len(matches) >= 2 {
		photoURL := strings.ReplaceAll(matches[1], "&amp;", "&")
		// Clean up tracking parameters
		if strings.Contains(photoURL, "&__cft__") {
			photoURL = strings.Split(photoURL, "&__cft__")[0]
		}
		if strings.Contains(photoURL, "&__tn__") {
			photoURL = strings.Split(photoURL, "&__tn__")[0]
		}
		// Make sure it has fbid parameter
		if strings.Contains(photoURL, "fbid=") {
			return photoURL, nil
		}
	}

	// Fallback: Try to find fbid in HTML
	patterns := []string{
		`"fbid":"(\d+)"`,
		`/photo\?fbid=(\d+)`,
		`fbid=(\d+)`,
		`photo_id=(\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlStr)
		if len(matches) >= 2 {
			fbid := matches[1]
			return fmt.Sprintf("https://www.facebook.com/photo?fbid=%s", fbid), nil
		}
	}

	// If we can't find anything, return original URL
	return permalinkURL, nil
}

func (s *FacebookService) GetPhotoURLs(fbURL string) ([]string, error) {
	normalizedURL, err := s.NormalizeFacebookURL(fbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize URL: %w", err)
	}

	// Try normalized URL first
	photos, err := s.extractPhotosFromURL(normalizedURL)
	if err == nil && len(photos) > 0 {
		return photos, nil
	}

	// If normalized URL fails and it's different from original, try original URL
	if normalizedURL != fbURL {
		photos, err = s.extractPhotosFromURL(fbURL)
		if err == nil && len(photos) > 0 {
			return photos, nil
		}
	}

	return nil, fmt.Errorf("no photos found in Facebook post (tried both normalized and original URLs)")
}

func (s *FacebookService) extractPhotosFromURL(targetURL string) ([]string, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	for name, value := range s.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	htmlStr := string(body)
	var photos []string

	// Pattern 1: og:image meta tag (usually highest quality and always the main content image)
	ogRe := regexp.MustCompile(`<meta property="og:image" content="([^"]+)"`)
	ogMatches := ogRe.FindAllStringSubmatch(htmlStr, -1)
	for _, match := range ogMatches {
		if len(match) > 1 {
			imgURL := match[1]
			// Decode HTML entities
			imgURL = strings.ReplaceAll(imgURL, "&amp;", "&")
			imgURL = strings.ReplaceAll(imgURL, "&#xff08;", "(")
			imgURL = strings.ReplaceAll(imgURL, "&#xff09;", ")")
			imgURL = strings.ReplaceAll(imgURL, "&#x", "")
			
			// og:image is special - it's almost always the main post content, not profile
			// Check if it's a valid photo URL
			if strings.Contains(imgURL, "scontent") && strings.Contains(imgURL, "fbcdn.net") {
				// Exclude only obvious profile picture patterns
				if !strings.Contains(imgURL, "/t1.30497-1/") && 
				   !strings.Contains(imgURL, "/t1.0-1/") &&
				   !strings.Contains(imgURL, "p160x160") && 
				   !strings.Contains(imgURL, "p130x130") {
					if !utils.Contains(photos, imgURL) {
						photos = append(photos, imgURL)
					}
				}
			}
		}
	}

	// Pattern 2: scontent URLs in any attribute
	// More flexible pattern to catch URLs in various attributes
	scontentRe := regexp.MustCompile(`(https://scontent[^"\s]+fbcdn\.net[^"\s]+\.jpg[^"\s]*)`)
	scontentMatches := scontentRe.FindAllStringSubmatch(htmlStr, -1)
	for _, match := range scontentMatches {
		if len(match) > 1 {
			imgURL := match[1]
			// Decode HTML entities
			imgURL = strings.ReplaceAll(imgURL, "&amp;", "&")
			imgURL = strings.ReplaceAll(imgURL, "&#", "")
			
			// Only add if it's a post photo (not profile picture)
			if isPostPhoto(imgURL) {
				if !utils.Contains(photos, imgURL) {
					photos = append(photos, imgURL)
				}
			}
		}
	}

	// Pattern 3: JSON embedded image URLs
	jsonImgRe := regexp.MustCompile(`"image":\{"uri":"(https:[^"]+)"`)
	jsonMatches := jsonImgRe.FindAllStringSubmatch(htmlStr, -1)
	for _, match := range jsonMatches {
		if len(match) > 1 {
			imgURL := strings.ReplaceAll(match[1], `\u0025`, "%")
			imgURL = strings.ReplaceAll(imgURL, `\/`, "/")
			imgURL = strings.ReplaceAll(imgURL, "&amp;", "&")
			if strings.Contains(imgURL, "scontent") && !strings.Contains(imgURL, "p160x160") {
				if !utils.Contains(photos, imgURL) {
					photos = append(photos, imgURL)
				}
			}
		}
	}

	// Pattern 4: High res image URLs in data attributes
	dataImgRe := regexp.MustCompile(`data-[^=]+-src="(https://scontent[^"]+)"`)
	dataMatches := dataImgRe.FindAllStringSubmatch(htmlStr, -1)
	for _, match := range dataMatches {
		if len(match) > 1 {
			imgURL := strings.ReplaceAll(match[1], "&amp;", "&")
			if !utils.Contains(photos, imgURL) {
				photos = append(photos, imgURL)
			}
		}
	}

	if len(photos) == 0 {
		return nil, fmt.Errorf("no photos found in URL: %s", targetURL)
	}

	// Deduplicate by image ID (extract filename from URL)
	uniquePhotos := make(map[string]string) // map[imageID]fullURL
	for _, photo := range photos {
		// Extract image ID from URL (the number before _n.jpg or similar)
		parts := strings.Split(photo, "/")
		if len(parts) > 0 {
			filename := parts[len(parts)-1]
			// Get the ID part (before .jpg)
			imageID := strings.Split(filename, ".jpg")[0]
			// Remove query parameters for comparison
			imageID = strings.Split(imageID, "?")[0]
			
			// Only keep the longest URL for each image ID (has more parameters = better quality)
			if existing, ok := uniquePhotos[imageID]; !ok || len(photo) > len(existing) {
				uniquePhotos[imageID] = photo
			}
		}
	}

	// Convert back to slice
	dedupedPhotos := make([]string, 0, len(uniquePhotos))
	for _, photoURL := range uniquePhotos {
		dedupedPhotos = append(dedupedPhotos, photoURL)
	}

	return dedupedPhotos, nil
}

// isPostPhoto checks if URL is a post photo (not profile picture or other small images)
func isPostPhoto(imgURL string) bool {
	// Must contain scontent
	if !strings.Contains(imgURL, "scontent") {
		return false
	}
	
	// Exclude profile pictures and thumbnails by size
	excludeSizes := []string{
		"p160x160", "p130x130", "s160x160", "s320x320",
		"p50x50", "p75x75", "p100x100", "p120x120",
		"s200x200", "s150x150", "p480x480", "p240x240",
	}
	for _, size := range excludeSizes {
		if strings.Contains(imgURL, size) {
			return false
		}
	}
	
	// Exclude profile pictures by path pattern
	// Profile pictures have these identifiers:
	// /t1.30497-1/ = Facebook profile picture
	// /t1.0-1/ = Old profile picture format
	// /t39.30808-1/ = Could be profile picture (ends with -1)
	excludePatterns := []string{
		"/t1.30497-1/",
		"/t1.0-1/",
		"/t1.6640-1/",
	}
	for _, pattern := range excludePatterns {
		if strings.Contains(imgURL, pattern) {
			return false
		}
	}
	
	// ONLY accept specific post photo patterns
	// Post photos have specific type indicators:
	postPatterns := []string{
		"/t39.30808-6/",  // Standard post photos
		"/t51.2885-",     // High resolution photos
		"/t39.30808-15/", // Large post photos
		"/t39.30808-12/", // Medium post photos
	}
	for _, pattern := range postPatterns {
		if strings.Contains(imgURL, pattern) {
			return true
		}
	}
	
	// If URL has stp parameter (photo processing parameter), check if it's a post
	// Common post indicators: stp=dst-jpg, stp=c* (crop parameters)
	if strings.Contains(imgURL, "stp=") {
		// Check if it's NOT a profile picture path
		if !strings.Contains(imgURL, "/t39.30808-1/") && 
		   !strings.Contains(imgURL, "/t1.") {
			// If it has dst-jpg or crop parameters, it's likely a post
			if strings.Contains(imgURL, "dst-jpg") || strings.Contains(imgURL, "stp=c") {
				return true
			}
		}
	}
	
	// Fallback: If it's t39.30808 (post photo type) and NOT -1 (profile), accept it
	if strings.Contains(imgURL, "/t39.30808-") && !strings.Contains(imgURL, "/t39.30808-1/") {
		return true
	}
	
	// Reject everything else to be safe
	return false
}

func ExtractFacebookID(fbURL string) (string, error) {
	u, err := url.Parse(fbURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if fbid := u.Query().Get("fbid"); fbid != "" {
		return fbid, nil
	}

	re := regexp.MustCompile(`/(?:photo|share)/(?:p/)?([^/?]+)`)
	matches := re.FindStringSubmatch(u.Path)
	if len(matches) >= 2 {
		return matches[1], nil
	}

	return "", fmt.Errorf("Facebook ID not found in URL")
}


