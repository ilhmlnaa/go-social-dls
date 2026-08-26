package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"twitter-down/internal/utils"
)

// Zerochan mewajibkan User-Agent yang memuat nama proyek + username (lihat /api).
// Bisa dioverride lewat env ZEROCHAN_USER_AGENT.
const defaultZerochanUserAgent = "go-social-dls - ilhmlnaa"

// ZerochanPost memetakan respons /{id}?json.
type ZerochanPost struct {
	ID      int64    `json:"id"`
	Small   string   `json:"small"`
	Medium  string   `json:"medium"`
	Large   string   `json:"large"`
	Full    string   `json:"full"`
	Width   int      `json:"width"`
	Height  int      `json:"height"`
	Size    int64    `json:"size"`
	Hash    string   `json:"hash"`
	Source  string   `json:"source"`
	Primary string   `json:"primary"`
	Tags    []string `json:"tags"`
}

// ZerochanImage adalah bentuk keluaran yang rapi.
type ZerochanImage struct {
	ID       int64    `json:"id"`
	Full     string   `json:"full"`  // original / resolusi penuh
	Large    string   `json:"large"` // versi lebih kecil
	Medium   string   `json:"medium,omitempty"`
	Small    string   `json:"small,omitempty"`
	Ext      string   `json:"ext"`
	FileSize int64    `json:"file_size"`
	Width    int      `json:"width"`
	Height   int      `json:"height"`
	Source   string   `json:"source,omitempty"`
	Primary  string   `json:"primary,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

// ZerochanService mengambil data dari API resmi Zerochan (read-only, tanpa browser).
type ZerochanService struct {
	client    *http.Client
	userAgent string
}

// NewZerochanService membuat service. HTTP client memakai proxy
// PROXY_ZEROCHAN / PROXY_GLOBAL bila di-set.
func NewZerochanService() *ZerochanService {
	ua := os.Getenv("ZEROCHAN_USER_AGENT")
	if ua == "" {
		ua = defaultZerochanUserAgent
	}
	return &ZerochanService{
		client:    utils.NewHTTPClient("zerochan", 20*time.Second),
		userAgent: ua,
	}
}

var zerochanIDRegex = regexp.MustCompile(`zerochan\.net/(\d+)`)

// ExtractZerochanID menerima URL entry ATAU ID mentah.
func ExtractZerochanID(input string) (int64, error) {
	if id, err := strconv.ParseInt(input, 10, 64); err == nil && id > 0 {
		return id, nil
	}
	matches := zerochanIDRegex.FindStringSubmatch(input)
	if len(matches) < 2 {
		return 0, fmt.Errorf("entry id tidak ditemukan pada input")
	}
	return strconv.ParseInt(matches[1], 10, 64)
}

func (s *ZerochanService) fetchJSON(rawURL string, target interface{}) error {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	// User-Agent wajib bernama, kalau anonim proyek bisa dibanned.
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d dari Zerochan", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// extFromURL mengambil ekstensi file dari URL (mis. png/jpg), default "jpg".
var extRegex = regexp.MustCompile(`\.([a-zA-Z0-9]{3,4})(?:\?|$)`)

func extFromURL(rawURL string) string {
	if m := extRegex.FindStringSubmatch(rawURL); len(m) == 2 {
		return m[1]
	}
	return "jpg"
}

// GetEntry mengambil detail satu entry Zerochan berdasarkan URL atau ID.
func (s *ZerochanService) GetEntry(input string) (*ZerochanImage, error) {
	id, err := ExtractZerochanID(input)
	if err != nil {
		return nil, fmt.Errorf("URL/ID Zerochan tidak valid")
	}

	var post ZerochanPost
	apiURL := fmt.Sprintf("https://www.zerochan.net/%d?json", id)
	if err := s.fetchJSON(apiURL, &post); err != nil {
		return nil, fmt.Errorf("gagal mengambil entry: %w", err)
	}

	if post.Full == "" && post.Large == "" {
		return nil, fmt.Errorf("entry ditemukan tetapi tidak punya URL gambar")
	}

	primaryURL := post.Full
	if primaryURL == "" {
		primaryURL = post.Large
	}

	return &ZerochanImage{
		ID:       post.ID,
		Full:     post.Full,
		Large:    post.Large,
		Medium:   post.Medium,
		Small:    post.Small,
		Ext:      extFromURL(primaryURL),
		FileSize: post.Size,
		Width:    post.Width,
		Height:   post.Height,
		Source:   post.Source,
		Primary:  post.Primary,
		Tags:     post.Tags,
	}, nil
}
