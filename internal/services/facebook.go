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

	if strings.Contains(u.Path, "/share/p/") {
		return s.resolveShareURL(fbURL)
	}

	if strings.Contains(u.Path, "/photo") && u.Query().Get("fbid") != "" {
		return fbURL, nil
	}

	if strings.Contains(u.Path, "/share/") {
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

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`/photo\?fbid=(\d+)&amp;set=([^"]+)`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) >= 3 {
		fbid := matches[1]
		set := strings.ReplaceAll(matches[2], "&amp;", "&")
		return fmt.Sprintf("https://www.facebook.com/photo?fbid=%s&set=%s", fbid, set), nil
	}

	return shareURL, nil
}

func (s *FacebookService) GetPhotoURLs(fbURL string) ([]string, error) {
	normalizedURL, err := s.NormalizeFacebookURL(fbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize URL: %w", err)
	}

	u, err := url.Parse(normalizedURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

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

	for name, value := range s.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var photos []string

	imgRe := regexp.MustCompile(`<img[^>]+src="(https://scontent[^"]+)"`)
	imgMatches := imgRe.FindAllStringSubmatch(string(body), -1)
	for _, match := range imgMatches {
		if len(match) > 1 {
			imgURL := strings.ReplaceAll(match[1], "&amp;", "&")
			if !utils.Contains(photos, imgURL) {
				photos = append(photos, imgURL)
			}
		}
	}

	ogRe := regexp.MustCompile(`<meta property="og:image" content="([^"]+)"`)
	ogMatches := ogRe.FindAllStringSubmatch(string(body), -1)
	for _, match := range ogMatches {
		if len(match) > 1 {
			imgURL := strings.ReplaceAll(match[1], "&amp;", "&")
			if !utils.Contains(photos, imgURL) {
				photos = append(photos, imgURL)
			}
		}
	}

	if len(photos) == 0 {
		return nil, fmt.Errorf("no photos found in Facebook post")
	}

	return photos, nil
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


