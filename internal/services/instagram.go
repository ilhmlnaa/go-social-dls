package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"twitter-down/internal/utils"
)

// InstagramService handles Instagram API operations
type InstagramService struct {
	client  *http.Client
	cookies map[string]string
}

// InstagramGraphQLResponse represents Instagram GraphQL response
type InstagramGraphQLResponse struct {
	Data struct {
		ShortcodeMedia struct {
			Typename      string `json:"__typename"`
			DisplayURL    string `json:"display_url"`
			VideoURL      string `json:"video_url"`
			DisplayResources []struct {
				Src          string `json:"src"`
				ConfigWidth  int    `json:"config_width"`
				ConfigHeight int    `json:"config_height"`
			} `json:"display_resources"`
			EdgeSidecarToChildren struct {
				Edges []struct {
					Node struct {
						Typename   string `json:"__typename"`
						DisplayURL string `json:"display_url"`
						VideoURL   string `json:"video_url"`
						DisplayResources []struct {
							Src          string `json:"src"`
							ConfigWidth  int    `json:"config_width"`
							ConfigHeight int    `json:"config_height"`
						} `json:"display_resources"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"edge_sidecar_to_children"`
		} `json:"shortcode_media"`
	} `json:"data"`
}

// NewInstagramService creates a new Instagram service instance
func NewInstagramService(cookiesDir string) (*InstagramService, error) {
	// Load cookies
	cookies, err := utils.LoadCookies(cookiesDir, "instagram")
	if err != nil {
		return nil, err
	}

	// Validate required cookies (sessionid is the main one for Instagram)
	required := []string{"sessionid"}
	if err := utils.ValidateCookies(cookies, required); err != nil {
		return nil, err
	}

	return &InstagramService{
		client:  &http.Client{},
		cookies: cookies,
	}, nil
}

// GetPostImages fetches image URLs from an Instagram post
func (s *InstagramService) GetPostImages(postURL string) ([]string, error) {
	// Extract shortcode from URL
	shortcode, err := extractInstagramShortcode(postURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract shortcode: %w", err)
	}

	// Build GraphQL query URL
	baseURL := "https://www.instagram.com/graphql/query/"

	// Instagram GraphQL query ID for media
	queryHash := "9f8827793ef34641b2fb195d4d41151c" // This might need to be updated
	
	variables := map[string]interface{}{
		"shortcode":              shortcode,
		"child_comment_count":    3,
		"fetch_comment_count":    40,
		"parent_comment_count":   24,
		"has_threaded_comments": true,
	}

	variablesJSON, _ := json.Marshal(variables)

	// Create request
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("query_hash", queryHash)
	q.Add("variables", string(variablesJSON))
	req.URL.RawQuery = q.Encode()

	// Add headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", postURL)

	// Add cookies
	for name, value := range s.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result InstagramGraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract images with highest resolution
	var images []string

	// Single post - get highest resolution
	if len(result.Data.ShortcodeMedia.DisplayResources) > 0 {
		// Get the last item which is usually the highest resolution
		highestRes := result.Data.ShortcodeMedia.DisplayResources[len(result.Data.ShortcodeMedia.DisplayResources)-1]
		images = append(images, highestRes.Src)
	} else if result.Data.ShortcodeMedia.DisplayURL != "" {
		// Fallback to display_url if display_resources not available
		images = append(images, result.Data.ShortcodeMedia.DisplayURL)
	}

	// Carousel/multiple images - get highest resolution for each
	for _, edge := range result.Data.ShortcodeMedia.EdgeSidecarToChildren.Edges {
		if len(edge.Node.DisplayResources) > 0 {
			// Get the last item which is usually the highest resolution
			highestRes := edge.Node.DisplayResources[len(edge.Node.DisplayResources)-1]
			images = append(images, highestRes.Src)
		} else if edge.Node.DisplayURL != "" {
			// Fallback to display_url
			images = append(images, edge.Node.DisplayURL)
		}
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no images found in Instagram post")
	}

	return images, nil
}

// GetPostImagesHTML fetches images from Instagram post by parsing HTML (fallback method)
func (s *InstagramService) GetPostImagesHTML(postURL string) ([]string, error) {
	// Create request
	req, err := http.NewRequest("GET", postURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	// Add cookies
	for name, value := range s.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	html := string(body)
	var images []string

	// Try to extract from og:image meta tag
	ogImageRe := regexp.MustCompile(`<meta property="og:image" content="([^"]+)"`)
	if matches := ogImageRe.FindStringSubmatch(html); len(matches) > 1 {
		images = append(images, matches[1])
	}

	// Try to extract from JSON-LD
	ldJsonRe := regexp.MustCompile(`<script type="application/ld\+json">(.+?)</script>`)
	ldMatches := ldJsonRe.FindAllStringSubmatch(html, -1)
	
	for _, match := range ldMatches {
		if len(match) > 1 {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(match[1]), &data); err == nil {
				if imgVal, ok := data["image"]; ok {
					switch v := imgVal.(type) {
					case string:
						if !utils.Contains(images, v) {
							images = append(images, v)
						}
					case []interface{}:
						for _, item := range v {
							if str, ok := item.(string); ok && !utils.Contains(images, str) {
								images = append(images, str)
							}
						}
					}
				}
			}
		}
	}

	// Try to extract display_url from shared data
	displayUrlRe := regexp.MustCompile(`"display_url":"([^"]+)"`)
	displayMatches := displayUrlRe.FindAllStringSubmatch(html, -1)
	
	for _, match := range displayMatches {
		if len(match) > 1 {
			// Decode unicode escapes
			imgURL := strings.ReplaceAll(match[1], `\u0026`, "&")
			if !utils.Contains(images, imgURL) {
				images = append(images, imgURL)
			}
		}
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no images found in Instagram post")
	}

	return images, nil
}

// extractInstagramShortcode extracts shortcode from Instagram URL
func extractInstagramShortcode(postURL string) (string, error) {
	// Parse URL
	u, err := url.Parse(postURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Extract shortcode from path
	// Instagram URLs: https://www.instagram.com/p/{shortcode}/ or /reel/{shortcode}/
	re := regexp.MustCompile(`/(p|reel|tv)/([A-Za-z0-9_-]+)`)
	matches := re.FindStringSubmatch(u.Path)
	if len(matches) < 3 {
		return "", fmt.Errorf("shortcode not found in URL")
	}

	return matches[2], nil
}


