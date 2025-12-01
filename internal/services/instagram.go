package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"twitter-down/internal/utils"
)

type InstagramHTTPService struct {
	cookiesDir string
	cookies    map[string]string
	client     *http.Client
}

func NewInstagramHTTPService(cookiesDir string) (*InstagramHTTPService, error) {
	cookies, err := utils.LoadCookies(cookiesDir, "instagram")
	if err != nil {
		return nil, fmt.Errorf("failed to load Instagram cookies: %w", err)
	}

	cl := &http.Client{
		Timeout: 20 * time.Second,
	}

	return &InstagramHTTPService{
		cookiesDir: cookiesDir,
		cookies:    cookies,
		client:     cl,
	}, nil
}

func contentKey(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	cacheKey := u.Query().Get("ig_cache_key")
	stp := u.Query().Get("stp")
	
	if cacheKey != "" {
		if stp != "" {
			return cacheKey + "|" + stp
		}
		return cacheKey
	}

	parts := strings.Split(u.Path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			if stp != "" {
				return parts[i] + "|" + stp
			}
			return parts[i]
		}
	}
	return raw
}

func sizeHint(raw string) int {
	u, err := url.Parse(raw)
	if err != nil {
		return 0
	}
	stp := u.Query().Get("stp")
	
	if stp == "" {
		return 500_000
	}

	isCropped := regexp.MustCompile(`c\d+\.\d+\.\d+\.\d+`).MatchString(stp)

	if regexp.MustCompile(`dst-jpg[^ps]*_e35[^ps]*_tt6$`).MatchString(stp) {
		return 100_000_000
	}

	reP := regexp.MustCompile(`p(\d+)x(\d+)`)
	if m := reP.FindStringSubmatch(stp); len(m) == 3 {
		w, _ := strconv.Atoi(m[1])
		h, _ := strconv.Atoi(m[2])
		area := w * h
		
		if (w == 1080 && h == 1080) || (w == 720 && h == 720) {
			return area * 1000
		}
		
		return area * 100
	}

	reS := regexp.MustCompile(`s(\d+)x(\d+)`)
	if m := reS.FindStringSubmatch(stp); len(m) == 3 {
		w, _ := strconv.Atoi(m[1])
		h, _ := strconv.Atoi(m[2])
		area := w * h
		
		if (w == 1080 && h == 1080) || (w == 750 && h == 750) {
			if isCropped {
				return area * 50 
			}
			return area * 500 
		}
		
		if isCropped {
			return area / 10 
		}
		return area * 10
	}
	
	if isCropped {
		return 1 
	}
	
	return 100_000
}

func decodeEscapes(s string) string {
	s = strings.ReplaceAll(s, `\u0026`, "&")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, `\/`, "/")
	s = strings.ReplaceAll(s, `\u00253D`, "=")
	s = strings.ReplaceAll(s, `\u003D`, "=")
	return s
}


func (s *InstagramHTTPService) httpGetWithCookies(ctx context.Context, rawURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 14; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0 Mobile Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,id;q=0.8")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Referer", "https://www.instagram.com/")

	for name, value := range s.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("instagram returned status %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *InstagramHTTPService) extractImageCandidates(html string) []string {
    var out []string
    uniq := map[string]bool{}

    reDisplayURL := regexp.MustCompile(`"display_url"\s*:\s*"([^"]+)"`)
    for _, m := range reDisplayURL.FindAllStringSubmatch(html, -1) {
        u := decodeEscapes(m[1])
        if strings.Contains(u, "scontent") {
            uniq[u] = true
        }
    }

    reCandidates := regexp.MustCompile(`"candidates"\s*:\s*\[([^\]]+)\]`)
    reURL := regexp.MustCompile(`"url"\s*:\s*"([^"]+)"`)

    for _, block := range reCandidates.FindAllStringSubmatch(html, -1) {
        inner := block[1]
        for _, m := range reURL.FindAllStringSubmatch(inner, -1) {
            u := decodeEscapes(m[1])
            if strings.Contains(u, "scontent") {
                uniq[u] = true
            }
        }
    }

    reCarousel := regexp.MustCompile(`"carousel_media"\s*:\s*\[([^\]]+(?:\[[^\]]*\][^\]]*)*)\]`)
    for _, block := range reCarousel.FindAllStringSubmatch(html, -1) {
        inner := block[1]
        for _, m := range reDisplayURL.FindAllStringSubmatch(inner, -1) {
            u := decodeEscapes(m[1])
            if strings.Contains(u, "scontent") {
                uniq[u] = true
            }
        }
        for _, m := range reURL.FindAllStringSubmatch(inner, -1) {
            u := decodeEscapes(m[1])
            if strings.Contains(u, "scontent") {
                uniq[u] = true
            }
        }
    }

    reDirect := regexp.MustCompile(`https://scontent[^"\\<>\s]+`)
    for _, u := range reDirect.FindAllString(html, -1) {
        u = decodeEscapes(u)
        if strings.Contains(u, "scontent") {
            uniq[u] = true
        }
    }

    for u := range uniq {
        out = append(out, u)
    }

    log.Printf("[Instagram] extracted %d unique image URLs", len(out))
    return out
}


func (s *InstagramHTTPService) pickBestPerContentKey(urls []string, returnAll bool) []string {
	bestPerKey := map[string]string{}
	bestScore := map[string]int{}

	log.Printf("[Instagram] Processing %d candidate URLs...", len(urls))
	
	for i, u := range urls {
		key := contentKey(u)
		score := sizeHint(u)
		
		keyPreview := key
		if len(keyPreview) > 30 {
			keyPreview = keyPreview[:30] + "..."
		}
		
		if _, ok := bestPerKey[key]; !ok {
			bestPerKey[key] = u
			bestScore[key] = score
			log.Printf("[Instagram] [%d] NEW key=%s score=%d", i+1, keyPreview, score)
		} else if score > bestScore[key] {
			log.Printf("[Instagram] [%d] UPGRADE key=%s score=%d > %d", i+1, keyPreview, score, bestScore[key])
			bestPerKey[key] = u
			bestScore[key] = score
		} else {
			log.Printf("[Instagram] [%d] SKIP key=%s score=%d <= %d", i+1, keyPreview, score, bestScore[key])
		}
	}

	type scoredURL struct {
		url   string
		score int
	}
	var scored []scoredURL
	for key, u := range bestPerKey {
		scored = append(scored, scoredURL{url: u, score: bestScore[key]})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	maxResults := len(scored)
	if !returnAll {
		maxResults = 3
		if len(scored) < maxResults {
			maxResults = len(scored)
		}
	}

	result := make([]string, 0, maxResults)
	for i := 0; i < maxResults; i++ {
		result = append(result, scored[i].url)
		log.Printf("[Instagram] FINAL SELECTED [%d]: score=%d url=%s", i+1, scored[i].score, scored[i].url)
	}

	return result
}

func (s *InstagramHTTPService) GetPostImages(postURL string, returnAll bool) ([]string, error) {
	log.Printf("[Instagram HTTP] fetching: %s (returnAll=%v)", postURL, returnAll)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	html, err := s.httpGetWithCookies(ctx, postURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch instagram page: %w", err)
	}

	cands := s.extractImageCandidates(html)
	if len(cands) == 0 {
		if strings.Contains(postURL, "?") {
			postURL = postURL + "&hl=en"
		} else {
			postURL = postURL + "?hl=en"
		}
		html2, err2 := s.httpGetWithCookies(ctx, postURL)
		if err2 == nil {
			cands = s.extractImageCandidates(html2)
		}
	}

	if len(cands) == 0 {
		return nil, errors.New("no images found in Instagram post")
	}

	log.Printf("[Instagram HTTP] extracted %d unique image URLs", len(cands))

	best := s.pickBestPerContentKey(cands, returnAll)
	if len(best) == 0 {
		return nil, errors.New("images detected but failed to select best quality")
	}

	log.Printf("[Instagram HTTP] selected %d image(s) from %d candidates", len(best), len(cands))
	return best, nil
}
