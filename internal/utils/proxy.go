package utils

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ProxyForProvider mengembalikan string proxy untuk provider tertentu.
//
// Prioritas:
//  1. PROXY_<PROVIDER>  (mis. PROXY_DANBOORU, PROXY_INSTAGRAM) — per-provider
//  2. PROXY_GLOBAL      — dipakai semua provider yang tidak punya proxy khusus
//  3. ""                — tanpa proxy (koneksi langsung)
//
// provider harus lowercase, mis. "danbooru", "instagram", "pixiv".
func ProxyForProvider(provider string) string {
	if provider != "" {
		key := "PROXY_" + strings.ToUpper(provider)
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return strings.TrimSpace(os.Getenv("PROXY_GLOBAL"))
}

// parseProxyURL memvalidasi & mem-parse string proxy.
//
// Format yang didukung (auth opsional):
//
//	http://172.20.20.102:8888
//	http://user:pass@p.webshare.io:80
//	https://... / socks5://...
//	172.20.20.102:8888        (skema http diasumsikan)
func parseProxyURL(proxyStr string) (*url.URL, error) {
	proxyStr = strings.TrimSpace(proxyStr)
	if proxyStr == "" {
		return nil, nil
	}
	// Kalau tanpa skema, asumsikan http.
	if !strings.Contains(proxyStr, "://") {
		proxyStr = "http://" + proxyStr
	}
	return url.Parse(proxyStr)
}

// ProxyTransport membangun *http.Transport dengan proxy untuk provider tertentu.
// Mengembalikan nil bila tidak ada proxy dikonfigurasi (pakai transport default).
func ProxyTransport(provider string) *http.Transport {
	pu, err := parseProxyURL(ProxyForProvider(provider))
	if err != nil || pu == nil {
		return nil
	}
	return &http.Transport{Proxy: http.ProxyURL(pu)}
}

// NewHTTPClient mengembalikan *http.Client yang sudah dikonfigurasi proxy
// untuk provider (jika ada) dan timeout tertentu (0 = tanpa timeout).
func NewHTTPClient(provider string, timeout time.Duration) *http.Client {
	client := &http.Client{Timeout: timeout}
	if tr := ProxyTransport(provider); tr != nil {
		client.Transport = tr
	}
	return client
}

// NewHTTPClientWithRedirect sama seperti NewHTTPClient tetapi dengan
// CheckRedirect kustom (mis. untuk resolve short-URL tanpa mengikuti redirect).
func NewHTTPClientWithRedirect(provider string, timeout time.Duration, checkRedirect func(req *http.Request, via []*http.Request) error) *http.Client {
	client := NewHTTPClient(provider, timeout)
	client.CheckRedirect = checkRedirect
	return client
}
