package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"twitter-down/internal/utils"
)

type InstagramService struct {
	client  *http.Client
	cookies map[string]string
}

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

func NewInstagramService(cookiesDir string) (*InstagramService, error) {
	cookies, err := utils.LoadCookies(cookiesDir, "instagram")
	if err != nil {
		return nil, err
	}

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

	// Strategy: Find URLs that Instagram provides WITHOUT size limitations
	// Don't modify any URLs - they are signed and will break if modified
	
	// First, decode common HTML entities
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, `\u0026`, "&")
	
	// DEBUG: Save HTML to file to inspect
	// os.WriteFile("/tmp/instagram_debug.html", []byte(html), 0644)
	// log.Printf("[Instagram] HTML saved to /tmp/instagram_debug.html for debugging")
	
	// Try to extract data from embedded JSON in <script> tags
	// Instagram embeds data in window._sharedData or similar
	scriptRe := regexp.MustCompile(`<script[^>]*>window\._sharedData\s*=\s*(\{.+?\});</script>`)
	scriptMatches := scriptRe.FindStringSubmatch(html)
	
	if len(scriptMatches) > 1 {
		log.Printf("[Instagram] Found _sharedData JSON, parsing...")
		var sharedData map[string]interface{}
		if err := json.Unmarshal([]byte(scriptMatches[1]), &sharedData); err == nil {
			// Navigate through the JSON structure to find image URLs
			if entryData, ok := sharedData["entry_data"].(map[string]interface{}); ok {
				if postPage, ok := entryData["PostPage"].([]interface{}); ok && len(postPage) > 0 {
					if firstPost, ok := postPage[0].(map[string]interface{}); ok {
						if graphql, ok := firstPost["graphql"].(map[string]interface{}); ok {
							if shortcodeMedia, ok := graphql["shortcode_media"].(map[string]interface{}); ok {
								// Try to get display_url
								if displayURL, ok := shortcodeMedia["display_url"].(string); ok {
									log.Printf("[Instagram] ✓ Found display_url from _sharedData!")
									images = append(images, displayURL)
									return images, nil
								}
							}
						}
					}
				}
			}
		}
	}
	
	// Alternative: Try to extract from __additionalDataLoaded or other script tags
	scriptRe2 := regexp.MustCompile(`<script[^>]*type="application/ld\+json"[^>]*>(.+?)</script>`)
	ldJsonMatches := scriptRe2.FindAllStringSubmatch(html, -1)
	
	for _, match := range ldJsonMatches {
		if len(match) > 1 {
			var ldData map[string]interface{}
			if err := json.Unmarshal([]byte(match[1]), &ldData); err == nil {
				if imageVal, ok := ldData["image"].(string); ok {
					log.Printf("[Instagram] ✓ Found image from LD+JSON!")
					images = append(images, imageVal)
					return images, nil
				}
			}
		}
	}
	
	log.Printf("[Instagram] No JSON data found, trying URL extraction...")
	
	// Look for URLs with ig_cache_key FIRST (these are the full resolution ones)
	// The key is to match URLs that may be in different contexts (attributes, JSON, etc)
	
	// Try to find ig_cache_key URLs with more flexible pattern
	igCacheKeyPattern := `https://scontent[^"'\s]*\.cdninstagram\.com[^"'\s]*\.jpg\?[^"'\s]*ig_cache_key=[^"'\s&]*[^"'\s]*`
	igRe := regexp.MustCompile(igCacheKeyPattern)
	igURLs := igRe.FindAllString(html, -1)
	
	log.Printf("[Instagram] Found %d URLs with ig_cache_key", len(igURLs))
	
	if len(igURLs) > 0 {
		for i, u := range igURLs {
			cleanURL := strings.ReplaceAll(u, `\"`, "")
			cleanURL = strings.ReplaceAll(cleanURL, `\\`, "")
			
			if i < 3 {
				urlPreview := cleanURL
				if len(urlPreview) > 150 {
					urlPreview = urlPreview[:150] + "..."
				}
				log.Printf("[Instagram] ig_cache_key URL #%d: %s", i+1, urlPreview)
			}
			
			// Return first one that doesn't have size limitation
			if !strings.Contains(cleanURL, "s150x150") &&
			   !strings.Contains(cleanURL, "s320x320") &&
			   !strings.Contains(cleanURL, "s640x640") &&
			   !strings.Contains(cleanURL, "s1080x1080") {
				log.Printf("[Instagram] ✓ Found full resolution URL with ig_cache_key!")
				images = append(images, cleanURL)
				return images, nil
			}
		}
	}
	
	// Look for ALL Instagram CDN URLs (not just ig_cache_key)
	urlPattern := `https://scontent[^"'\s<>\\]*\.cdninstagram\.com[^"'\s<>\\]*\.jpg\?[^"'\s<>\\]+`
	urlRe := regexp.MustCompile(urlPattern)
	allURLs := urlRe.FindAllString(html, -1)
	
	log.Printf("[Instagram] Found %d total Instagram CDN URLs", len(allURLs))
	
	if len(allURLs) > 0 {
		// Analyze all URLs
		type urlInfo struct {
			url        string
			paramCount int
			length     int
			hasSize    bool
			sizeTag    string
		}
		
		var urlList []urlInfo
		
		for i, url := range allURLs {
			// Clean up escape characters
			cleanURL := strings.ReplaceAll(url, `\"`, "")
			cleanURL = strings.ReplaceAll(cleanURL, `\\`, "")
			
			// Count parameters
			paramCount := strings.Count(cleanURL, "&") + strings.Count(cleanURL, "=")
			
			// Check for size limitations
			sizeTag := ""
			hasSize := false
			if strings.Contains(cleanURL, "s150x150") {
				hasSize = true
				sizeTag = "s150x150"
			} else if strings.Contains(cleanURL, "s320x320") {
				hasSize = true
				sizeTag = "s320x320"
			} else if strings.Contains(cleanURL, "s640x640") {
				hasSize = true
				sizeTag = "s640x640"
			} else if strings.Contains(cleanURL, "s1080x1080") {
				hasSize = true
				sizeTag = "s1080x1080"
			}
			
			info := urlInfo{
				url:        cleanURL,
				paramCount: paramCount,
				length:     len(cleanURL),
				hasSize:    hasSize,
				sizeTag:    sizeTag,
			}
			urlList = append(urlList, info)
			
			// Log first 5 URLs for debugging
			if i < 5 {
				urlPreview := cleanURL
				if len(urlPreview) > 150 {
					urlPreview = urlPreview[:150] + "..."
				}
				log.Printf("[Instagram] URL #%d: len=%d, params=%d, hasSize=%v(%s)", 
					i+1, info.length, info.paramCount, info.hasSize, info.sizeTag)
				log.Printf("[Instagram]   -> %s", urlPreview)
			}
		}
		
		// Find best URL: prioritize non-size-limited URLs
		var bestURL string
		bestScore := -1
		bestIdx := -1
		
		for idx, info := range urlList {
			// Score: strongly prefer URLs without size limitation
			score := 0
			if !info.hasSize {
				score += 10000 // Strongly prefer non-size-limited URLs
			}
			score += info.paramCount * 10
			score += info.length / 10
			
			if score > bestScore {
				bestScore = score
				bestURL = info.url
				bestIdx = idx
			}
		}
		
		if bestURL != "" {
			log.Printf("[Instagram] ✓ Selected URL #%d (score=%d, len=%d, hasSize=%v)", 
				bestIdx+1, bestScore, len(bestURL), urlList[bestIdx].hasSize)
			urlPreview := bestURL
			if len(urlPreview) > 150 {
				urlPreview = urlPreview[:150] + "..."
			}
			log.Printf("[Instagram] ✓ %s", urlPreview)
			images = append(images, bestURL)
			return images, nil
		}
	}
	
	log.Printf("[Instagram] No URLs found, trying other fallbacks...")
	
	// Try to extract high-res URLs from img tags
	// Prefer URLs with dst-jpg_e35 or dst-jpg_e15 (full resolution)
	highResRe := regexp.MustCompile(`src="(https://scontent[^"]+cdninstagram\.com[^"]+dst-jpg[^"]+)"`)
	highResMatches := highResRe.FindAllStringSubmatch(html, -1)
	for _, match := range highResMatches {
		if len(match) > 1 {
			imgURL := strings.ReplaceAll(match[1], "&amp;", "&")
			// Skip if it's a thumbnail (has s150x150, s320x320, s640x640)
			if !strings.Contains(imgURL, "s150x150") && 
			   !strings.Contains(imgURL, "s320x320") && 
			   !strings.Contains(imgURL, "s640x640") &&
			   !utils.Contains(images, imgURL) {
				images = append(images, imgURL)
			}
		}
	}

	// Fallback: extract from og:image meta tag (might be medium quality)
	if len(images) == 0 {
		ogImageRe := regexp.MustCompile(`<meta property="og:image" content="([^"]+)"`)
		if matches := ogImageRe.FindStringSubmatch(html); len(matches) > 1 {
			images = append(images, matches[1])
		}
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

	// If no high-res images found, try display_url as-is
	if len(images) == 0 {
		// Try display_url from JSON
		displayUrlRe := regexp.MustCompile(`"display_url":"([^"]+)"`)
		if matches := displayUrlRe.FindStringSubmatch(html); len(matches) > 1 {
			imgURL := strings.ReplaceAll(matches[1], `\u0026`, "&")
			images = append(images, imgURL)
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


