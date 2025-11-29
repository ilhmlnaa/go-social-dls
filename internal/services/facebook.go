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

// FacebookService handles Facebook API operations
type FacebookService struct {
	client  *http.Client
	cookies map[string]string
}

// FacebookGraphQLResponse represents the response from Facebook GraphQL API
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

// NewFacebookService creates a new Facebook service instance
func NewFacebookService(cookiesDir string) (*FacebookService, error) {
	// Load cookies
	cookies, err := utils.LoadCookies(cookiesDir, "facebook")
	if err != nil {
		return nil, err
	}

	// Validate required cookies (Facebook typically needs c_user and xs)
	required := []string{"c_user", "xs"}
	if err := utils.ValidateCookies(cookies, required); err != nil {
		return nil, err
	}

	return &FacebookService{
		client:  &http.Client{},
		cookies: cookies,
	}, nil
}

// NormalizeFacebookURL converts various Facebook URL formats to standard photo URL
func (s *FacebookService) NormalizeFacebookURL(fbURL string) (string, error) {
	// Parse URL
	u, err := url.Parse(fbURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Case 1: /share/p/{id}/ format - need to fetch and extract real photo URL
	if strings.Contains(u.Path, "/share/p/") {
		return s.resolveShareURL(fbURL)
	}

	// Case 2: /photo/?fbid={id}&set={set} format - already normalized
	if strings.Contains(u.Path, "/photo") && u.Query().Get("fbid") != "" {
		return fbURL, nil
	}

	// Case 3: /share/{id}/ format - try to resolve
	if strings.Contains(u.Path, "/share/") {
		return s.resolveShareURL(fbURL)
	}

	return fbURL, nil
}

// resolveShareURL fetches the share URL and extracts the actual photo URL
func (s *FacebookService) resolveShareURL(shareURL string) (string, error) {
	req, err := http.NewRequest("GET", shareURL, nil)
	if err != nil {
		return "", err
	}

	// Add cookies
	for name, value := range s.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	// Add headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Extract photo URL from HTML using regex
	// Look for /photo?fbid= pattern
	re := regexp.MustCompile(`/photo\?fbid=(\d+)&amp;set=([^"]+)`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) >= 3 {
		fbid := matches[1]
		set := strings.ReplaceAll(matches[2], "&amp;", "&")
		return fmt.Sprintf("https://www.facebook.com/photo?fbid=%s&set=%s", fbid, set), nil
	}

	// If no match, return original URL
	return shareURL, nil
}

// GetPhotoURLs fetches photo URLs from a Facebook post
func (s *FacebookService) GetPhotoURLs(fbURL string) ([]string, error) {
	// Normalize URL first
	normalizedURL, err := s.NormalizeFacebookURL(fbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize URL: %w", err)
	}

	// Parse normalized URL
	u, err := url.Parse(normalizedURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Extract fbid
	fbid := u.Query().Get("fbid")
	if fbid == "" {
		return nil, fmt.Errorf("fbid not found in URL")
	}

	// For now, construct direct image URL
	// Facebook's image URLs follow pattern: https://scontent.xx.fbcdn.net/v/...
	// We'll need to make actual request to get the image URL
	
	// Make request to photo page
	req, err := http.NewRequest("GET", normalizedURL, nil)
	if err != nil {
		return nil, err
	}

	// Add cookies
	for name, value := range s.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	// Add headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Extract image URLs from HTML
	// Facebook stores image URLs in various formats, try multiple patterns
	var photos []string

	// Pattern 1: Direct image URL in img tags
	imgRe := regexp.MustCompile(`<img[^>]+src="(https://scontent[^"]+)"`)
	imgMatches := imgRe.FindAllStringSubmatch(string(body), -1)
	for _, match := range imgMatches {
		if len(match) > 1 {
			// Decode HTML entities
			imgURL := strings.ReplaceAll(match[1], "&amp;", "&")
			// Only add high quality images (check dimensions or size)
			if !contains(photos, imgURL) {
				photos = append(photos, imgURL)
			}
		}
	}

	// Pattern 2: Look for image URLs in og:image meta tags
	ogRe := regexp.MustCompile(`<meta property="og:image" content="([^"]+)"`)
	ogMatches := ogRe.FindAllStringSubmatch(string(body), -1)
	for _, match := range ogMatches {
		if len(match) > 1 {
			imgURL := strings.ReplaceAll(match[1], "&amp;", "&")
			if !contains(photos, imgURL) {
				photos = append(photos, imgURL)
			}
		}
	}

	if len(photos) == 0 {
		return nil, fmt.Errorf("no photos found in Facebook post")
	}

	return photos, nil
}

// ExtractFacebookID extracts photo or post ID from Facebook URL
func ExtractFacebookID(fbURL string) (string, error) {
	u, err := url.Parse(fbURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Try to get fbid from query
	if fbid := u.Query().Get("fbid"); fbid != "" {
		return fbid, nil
	}

	// Try to extract from path
	re := regexp.MustCompile(`/(?:photo|share)/(?:p/)?([^/?]+)`)
	matches := re.FindStringSubmatch(u.Path)
	if len(matches) >= 2 {
		return matches[1], nil
	}

	return "", fmt.Errorf("Facebook ID not found in URL")
}

// contains checks if string slice contains a string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
