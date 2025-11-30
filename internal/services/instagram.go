package services

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"twitter-down/internal/utils"
)

type InstagramBrowserService struct {
	cookiesDir string
	cookies    map[string]string
}

func NewInstagramBrowserService(cookiesDir string) (*InstagramBrowserService, error) {
	cookies, err := utils.LoadCookies(cookiesDir, "instagram")
	if err != nil {
		return nil, fmt.Errorf("failed to load Instagram cookies: %w", err)
	}

	return &InstagramBrowserService{
		cookiesDir: cookiesDir,
		cookies:    cookies,
	}, nil
}

type cookieJSON struct {
	Name     string      `json:"name"`
	Value    string      `json:"value"`
	Domain   string      `json:"domain"`
	Path     string      `json:"path"`
	Expires  interface{} `json:"expirationDate"`
	SameSite string      `json:"sameSite"`
	Secure   bool        `json:"secure"`
	HTTPOnly bool        `json:"httpOnly"`
}

func parseExpires(v interface{}) *cdp.TimeSinceEpoch {
	if v == nil {
		return nil
	}

	switch t := v.(type) {
	case float64:
		sec := int64(t)
		exp := cdp.TimeSinceEpoch(time.Unix(sec, 0))
		return &exp

	case string:
		if secInt, err := strconv.ParseInt(t, 10, 64); err == nil {
			exp := cdp.TimeSinceEpoch(time.Unix(secInt, 0))
			return &exp
		}
		if tm, err := time.Parse(time.RFC3339, t); err == nil {
			exp := cdp.TimeSinceEpoch(tm)
			return &exp
		}
	}

	return nil
}

func contentKey(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	if q := u.Query().Get("ig_cache_key"); q != "" {
		return q
	}

	parts := strings.Split(u.Path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
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
		return 1_000_000
	}
	re := regexp.MustCompile(`s(\d+)x(\d+)`)
	m := re.FindStringSubmatch(stp)
	if len(m) == 3 {
		w, _ := strconv.Atoi(m[1])
		h, _ := strconv.Atoi(m[2])
		if w > h {
			return w
		}
		return h
	}
	return 0
}

func (s *InstagramBrowserService) GetPostImages(postURL string) ([]string, error) {
	log.Printf("[Instagram Browser] Starting extraction for: %s", postURL)

	var cookieList []cookieJSON
	for name, value := range s.cookies {
		cookieList = append(cookieList, cookieJSON{
			Name:   name,
			Value:  value,
			Domain: ".instagram.com",
			Path:   "/",
		})
	}

	log.Printf("[Instagram Browser] Loaded %d cookies", len(cookieList))

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath("/usr/bin/google-chrome"),
		chromedp.NoSandbox,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	if err := chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		network.Enable(),
	); err != nil {
		return nil, fmt.Errorf("failed to initialize browser: %w", err)
	}

	var cookieActions []chromedp.Action
	for _, c := range cookieList {
		exp := parseExpires(c.Expires)

		sc := network.SetCookie(c.Name, c.Value).
			WithDomain(c.Domain).
			WithPath(c.Path).
			WithHTTPOnly(c.HTTPOnly).
			WithSecure(c.Secure).
			WithURL("https://www.instagram.com/")

		if exp != nil {
			sc = sc.WithExpires(exp)
		}

		switch strings.ToLower(c.SameSite) {
		case "lax":
			sc = sc.WithSameSite(network.CookieSameSiteLax)
		case "strict":
			sc = sc.WithSameSite(network.CookieSameSiteStrict)
		default:
			sc = sc.WithSameSite(network.CookieSameSiteNone)
		}

		cookieActions = append(cookieActions, sc)
	}

	if err := chromedp.Run(ctx, cookieActions...); err != nil {
		log.Printf("[Instagram Browser] Warning: Some cookies may have failed to set: %v", err)
	} else {
		log.Printf("[Instagram Browser] Cookies injected successfully!")
	}

	var html string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(postURL),
		chromedp.Sleep(3*time.Second), 
		chromedp.InnerHTML("html", &html, chromedp.ByQuery),
	); err != nil {
		return nil, fmt.Errorf("failed to load Instagram page: %w", err)
	}

	log.Printf("[Instagram Browser] Page loaded, scanning HTML...")

	reImg := regexp.MustCompile(`https://scontent[^"]+\.jpg[^"]*`)
	matches := reImg.FindAllString(html, -1)
	log.Printf("[Instagram Browser] Found %d image URLs", len(matches))

	unique := map[string]bool{}
	var final []string
	u0026 := regexp.MustCompile(`\\u0026`)
	amp := regexp.MustCompile(`&amp;`)

	for _, u := range matches {
		u = u0026.ReplaceAllString(u, "&")
		u = amp.ReplaceAllString(u, "&")

		if !strings.Contains(u, "scontent") || !strings.Contains(u, "instagram.com") {
			continue
		}

		if !unique[u] {
			unique[u] = true
			final = append(final, u)
		}
	}

	bestPerKey := map[string]string{}
	bestScore := map[string]int{}

	for _, u := range final {
		key := contentKey(u)
		score := sizeHint(u)

		if _, ok := bestPerKey[key]; !ok || score > bestScore[key] {
			bestPerKey[key] = u
			bestScore[key] = score
		}
	}

	var result []string
	for _, u := range bestPerKey {
		result = append(result, u)
	}

	log.Printf("[Instagram Browser] Selected %d best quality images from %d total URLs", len(result), len(final))

	if len(result) == 0 {
		return nil, fmt.Errorf("no images found in Instagram post")
	}

	return result, nil
}
