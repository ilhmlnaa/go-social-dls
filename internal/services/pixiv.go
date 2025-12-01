package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"twitter-down/internal/utils"
)

type PixivService struct {
	client  *http.Client
	cookies map[string]string
}

type PixivAjaxResponse struct {
	Error   bool                 `json:"error"`
	Message string               `json:"message"`
	Body    PixivIllustrationData `json:"body"`
}

type PixivIllustrationData struct {
	IllustID    string            `json:"illustId"`
	IllustTitle string            `json:"illustTitle"`
	IllustType  int               `json:"illustType"`
	Urls        PixivUrls         `json:"urls"`
	Tags        PixivTags         `json:"tags"`
	UserId      string            `json:"userId"`
	UserName    string            `json:"userName"`
	PageCount   int               `json:"pageCount"`
}

type PixivUrls struct {
	Mini     string `json:"mini"`
	Thumb    string `json:"thumb"`
	Small    string `json:"small"`
	Regular  string `json:"regular"`
	Original string `json:"original"`
}

type PixivTags struct {
	AuthTags []PixivTag `json:"authorTags"`
	Tags     []PixivTag `json:"tags"`
}

type PixivTag struct {
	Tag         string `json:"tag"`
	Locked      bool   `json:"locked"`
	Deletable   bool   `json:"deletable"`
	UserId      string `json:"userId,omitempty"`
	UserName    string `json:"userName,omitempty"`
	Translation map[string]string `json:"translation,omitempty"`
}

type PixivPagesResponse struct {
	Error bool        `json:"error"`
	Body  []PixivPage `json:"body"`
}

type PixivPage struct {
	Urls   PixivUrls `json:"urls"`
	Width  int       `json:"width"`
	Height int       `json:"height"`
}

func NewPixivService(cookiesDir string) (*PixivService, error) {
	cookies, err := utils.LoadCookies(cookiesDir, "pixiv")
	if err != nil {
		return nil, err
	}

	required := []string{"PHPSESSID"}
	if err := utils.ValidateCookies(cookies, required); err != nil {
		return nil, err
	}

	return &PixivService{
		client:  &http.Client{},
		cookies: cookies,
	}, nil
}

func (s *PixivService) ExtractIllustrationID(url string) (string, error) {
	re := regexp.MustCompile(`(?:www\.)?pixiv\.net/(?:en/)?artworks/(\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid Pixiv URL format")
	}
	return matches[1], nil
}

func (s *PixivService) GetIllustrationData(illustID string) (*PixivIllustrationData, error) {
	url := fmt.Sprintf("https://www.pixiv.net/ajax/illust/%s", illustID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", fmt.Sprintf("https://www.pixiv.net/en/artworks/%s", illustID))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	for name, value := range s.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var ajaxResp PixivAjaxResponse
	if err := json.Unmarshal(body, &ajaxResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if ajaxResp.Error {
		return nil, fmt.Errorf("pixiv API error: %s", ajaxResp.Message)
	}

	return &ajaxResp.Body, nil
}

func (s *PixivService) GetMultiplePages(illustID string) ([]PixivPage, error) {
	url := fmt.Sprintf("https://www.pixiv.net/ajax/illust/%s/pages", illustID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", fmt.Sprintf("https://www.pixiv.net/en/artworks/%s", illustID))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	for name, value := range s.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var pagesResp PixivPagesResponse
	if err := json.Unmarshal(body, &pagesResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if pagesResp.Error {
		return nil, fmt.Errorf("pixiv API error for pages")
	}

	return pagesResp.Body, nil
}

func (s *PixivService) GetIllustrationImages(illustURL string) ([]string, error) {
	illustID, err := s.ExtractIllustrationID(illustURL)
	if err != nil {
		return nil, err
	}

	illustData, err := s.GetIllustrationData(illustID)
	if err != nil {
		return nil, err
	}

	var imageURLs []string

	if illustData.PageCount == 1 {
		if illustData.Urls.Original != "" {
			imageURLs = append(imageURLs, illustData.Urls.Original)
		} else {
			imageURLs = append(imageURLs, illustData.Urls.Regular)
		}
	} else {
		pages, err := s.GetMultiplePages(illustID)
		if err != nil {
			return nil, err
		}

		for _, page := range pages {
			if page.Urls.Original != "" {
				imageURLs = append(imageURLs, page.Urls.Original)
			} else {
				imageURLs = append(imageURLs, page.Urls.Regular)
			}
		}
	}

	return imageURLs, nil
}

func (s *PixivService) ConvertToOriginalURL(url string) string {
	if strings.Contains(url, "img-master") && strings.Contains(url, "_master1200") {
		originalURL := strings.Replace(url, "img-master", "img-original", 1)
		originalURL = strings.Replace(originalURL, "_master1200", "", 1)
		
		if s.testImageURL(originalURL) {
			return originalURL
		}
		
		if strings.HasSuffix(originalURL, ".jpg") {
			pngURL := strings.Replace(originalURL, ".jpg", ".png", 1)
			if s.testImageURL(pngURL) {
				return pngURL
			}
		}
	}
	
	return url
}

func (s *PixivService) testImageURL(url string) bool {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Referer", "https://www.pixiv.net/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	for name, value := range s.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
