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

// Danbooru mewajibkan User-Agent yang jelas (lihat help:api).
const danbooruUserAgent = "go-social-dls/1.0 (+https://github.com/ilhmlnaa/go-social-dls)"

// DanbooruMediaAsset merepresentasikan objek media_asset (berisi variants).
type DanbooruMediaAsset struct {
	Variants []struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"variants"`
}

// DanbooruPost hanya memetakan field yang dibutuhkan dari /posts/{id}.json.
type DanbooruPost struct {
	ID                 int64              `json:"id"`
	ParentID           *int64             `json:"parent_id"`
	HasChildren        bool               `json:"has_children"`
	FileURL            string             `json:"file_url"`       // original
	LargeFileURL       string             `json:"large_file_url"` // sample/large
	PreviewFileURL     string             `json:"preview_file_url"`
	FileExt            string             `json:"file_ext"`
	FileSize           int64              `json:"file_size"`
	ImageWidth         int                `json:"image_width"`
	ImageHeight        int                `json:"image_height"`
	Rating             string             `json:"rating"`
	Source             string             `json:"source"`
	MD5                string             `json:"md5"`
	TagStringArtist    string             `json:"tag_string_artist"`
	TagStringCharacter string             `json:"tag_string_character"`
	TagStringCopyright string             `json:"tag_string_copyright"`
	MediaAsset         DanbooruMediaAsset `json:"media_asset"`
}

// DanbooruImage adalah bentuk keluaran yang rapi per-post.
type DanbooruImage struct {
	ID        int64  `json:"id"`
	ParentID  *int64 `json:"parent_id"`
	Original  string `json:"original"`
	Sample    string `json:"sample,omitempty"`
	Preview   string `json:"preview,omitempty"`
	Ext       string `json:"ext"`
	FileSize  int64  `json:"file_size"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Rating    string `json:"rating"`
	Source    string `json:"source,omitempty"`
	Artist    string `json:"artist,omitempty"`
	Character string `json:"character,omitempty"`
	Copyright string `json:"copyright,omitempty"`
	IsMain    bool   `json:"is_main"`
}

// DanbooruResult adalah payload akhir untuk handler.
type DanbooruResult struct {
	PostID       int64           `json:"post_id"`
	HasChildren  bool            `json:"has_children"`
	Count        int             `json:"count"`
	WithChildren bool            `json:"with_children"`
	Images       []DanbooruImage `json:"images"`
}

// DanbooruService mengambil data post dari API resmi Danbooru (tanpa browser).
type DanbooruService struct {
	client *http.Client
	login  string
	apiKey string
}

// NewDanbooruService membuat service dengan auth opsional dari environment.
// HTTP client memakai proxy PROXY_DANBOORU / PROXY_GLOBAL bila di-set.
func NewDanbooruService() *DanbooruService {
	return &DanbooruService{
		client: utils.NewHTTPClient("danbooru", 20*time.Second),
		login:  os.Getenv("DANBOORU_LOGIN"),
		apiKey: os.Getenv("DANBOORU_API_KEY"),
	}
}

var danbooruIDRegex = regexp.MustCompile(`(?:/posts/|/post/show/)(\d+)`)

// ExtractDanbooruID menerima URL post ATAU ID mentah.
func ExtractDanbooruID(input string) (int64, error) {
	if id, err := strconv.ParseInt(input, 10, 64); err == nil && id > 0 {
		return id, nil
	}
	matches := danbooruIDRegex.FindStringSubmatch(input)
	if len(matches) < 2 {
		return 0, fmt.Errorf("post id tidak ditemukan pada input")
	}
	return strconv.ParseInt(matches[1], 10, 64)
}

// fetchJSON melakukan GET dengan UA + basic auth opsional, lalu decode JSON.
func (s *DanbooruService) fetchJSON(rawURL string, target interface{}) error {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", danbooruUserAgent)

	// Auth opsional: berguna untuk post terbatas / rate limit lebih longgar.
	if s.login != "" && s.apiKey != "" {
		req.SetBasicAuth(s.login, s.apiKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d dari Danbooru", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// pickSample memilih URL sample terbaik dari media_asset bila large_file_url kosong.
func pickSample(p *DanbooruPost) string {
	if p.LargeFileURL != "" {
		return p.LargeFileURL
	}
	for _, v := range p.MediaAsset.Variants {
		if v.Type == "sample" {
			return v.URL
		}
	}
	return ""
}

func toImage(p *DanbooruPost, isMain bool) DanbooruImage {
	return DanbooruImage{
		ID:        p.ID,
		ParentID:  p.ParentID,
		Original:  p.FileURL,
		Sample:    pickSample(p),
		Preview:   p.PreviewFileURL,
		Ext:       p.FileExt,
		FileSize:  p.FileSize,
		Width:     p.ImageWidth,
		Height:    p.ImageHeight,
		Rating:    p.Rating,
		Source:    p.Source,
		Artist:    p.TagStringArtist,
		Character: p.TagStringCharacter,
		Copyright: p.TagStringCopyright,
		IsMain:    isMain,
	}
}

// GetPostImages mengambil satu post; jika includeChildren=true, ikut menyertakan
// induk + seluruh child/sibling dalam grup yang sama.
//
// Catatan: post dengan has_children=true tetap bisa diunduh normal — pembatas
// "child posts" hanyalah peringatan di UI web, API tetap memberi file_url.
func (s *DanbooruService) GetPostImages(input string, includeChildren bool) (*DanbooruResult, error) {
	postID, err := ExtractDanbooruID(input)
	if err != nil {
		return nil, fmt.Errorf("URL/ID Danbooru tidak valid")
	}

	var post DanbooruPost
	apiURL := fmt.Sprintf("https://danbooru.donmai.us/posts/%d.json", postID)
	if err := s.fetchJSON(apiURL, &post); err != nil {
		return nil, fmt.Errorf("gagal mengambil post: %w", err)
	}

	if post.FileURL == "" {
		return nil, fmt.Errorf("post ditemukan tetapi tidak punya file_url (mungkin dihapus/dibatasi; coba set DANBOORU_LOGIN & DANBOORU_API_KEY)")
	}

	images := []DanbooruImage{toImage(&post, true)}

	if includeChildren {
		// Root grup: kalau post punya parent, pakai parent agar dapat semua saudara.
		rootID := post.ID
		if post.ParentID != nil {
			rootID = *post.ParentID
		}

		var related []DanbooruPost
		relURL := fmt.Sprintf("https://danbooru.donmai.us/posts.json?tags=parent:%d&limit=200", rootID)
		if err := s.fetchJSON(relURL, &related); err == nil {
			seen := map[int64]bool{post.ID: true}
			for i := range related {
				rp := &related[i]
				if rp.FileURL == "" || seen[rp.ID] {
					continue
				}
				seen[rp.ID] = true
				images = append(images, toImage(rp, false))
			}
		}
	}

	return &DanbooruResult{
		PostID:       post.ID,
		HasChildren:  post.HasChildren,
		Count:        len(images),
		WithChildren: includeChildren,
		Images:       images,
	}, nil
}
