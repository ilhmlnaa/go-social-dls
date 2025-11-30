package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Cookie struct {
	Name     string      `json:"name"`
	Value    string      `json:"value"`
	Domain   string      `json:"domain"`
	Path     string      `json:"path"`
	Expires  interface{} `json:"expirationDate"`
	SameSite string      `json:"sameSite"`
	Secure   bool        `json:"secure"`
	HTTPOnly bool        `json:"httpOnly"`
}

// --- Load cookies properly ---
func loadCookies(path string) ([]Cookie, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cookies []Cookie
	err = json.Unmarshal(raw, &cookies)
	if err != nil {
		return nil, err
	}

	return cookies, nil
}

// --- Convert string/float expiration to *cdp.TimeSinceEpoch ---
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
		// Attempt string → int
		sec, err := time.Parse(time.RFC3339, t)
		if err != nil {
			return nil
		}
		exp := cdp.TimeSinceEpoch(sec)
		return &exp
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ig-chromedp.go <instagram URL>")
		return
	}

	instaURL := os.Args[1]
	cookiePath := "cookies/instagram.json"

	cookies, err := loadCookies(cookiePath)
	if err != nil {
		log.Fatal("Error reading cookies: ", err)
	}
	fmt.Printf("[*] Loaded %d cookies\n", len(cookies))

	// ChromeDP context
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// Enable network
	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		log.Fatal(err)
	}

	// --- Inject cookies using network.SetCookie ---
	for _, c := range cookies {
		exp := parseExpires(c.Expires)

		sc := network.SetCookie(c.Name, c.Value).
			WithDomain(c.Domain).
			WithPath(c.Path).
			WithHTTPOnly(c.HTTPOnly).
			WithSecure(c.Secure)

		if exp != nil {
			sc = sc.WithExpires(exp)
		}

		// SameSite mapping
		switch c.SameSite {
		case "Lax", "lax":
			sc = sc.WithSameSite(network.CookieSameSiteLax)
		case "Strict", "strict":
			sc = sc.WithSameSite(network.CookieSameSiteStrict)
		default:
			sc = sc.WithSameSite(network.CookieSameSiteNone)
		}

		// 🔥 FIX: sc.Do returns ONLY error now
		err := sc.Do(ctx)
		if err != nil {
			log.Printf("Failed to set cookie %s: %v", c.Name, err)
		} else {
			log.Printf("Cookie set: %s", c.Name)
		}
	}


	fmt.Println("[*] Cookies injected successfully!")

	// Visit the Instagram URL
	var html string
	err = chromedp.Run(ctx,
		chromedp.Navigate(instaURL),
		chromedp.Sleep(3*time.Second),
		chromedp.InnerHTML("html", &html, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("[*] Page loaded, scanning HTML...")

	// Regex for CDN images
	reImg := regexp.MustCompile(`https://scontent[^"]+\.jpg[^"]*`)
	matches := reImg.FindAllString(html, -1)

	fmt.Printf("[*] Found %d image URLs\n", len(matches))

	unique := map[string]bool{}
	final := []string{}
	for _, u := range matches {
		u = regexp.MustCompile(`\\u0026`).ReplaceAllString(u, "&")
		u = regexp.MustCompile(`&amp;`).ReplaceAllString(u, "&")

		if !unique[u] {
			unique[u] = true
			final = append(final, u)
		}
	}

	fmt.Println("\n========== RESULT ==========")
	for i, url := range final {
		fmt.Printf("[%d] %s\n", i+1, url)
	}
}
