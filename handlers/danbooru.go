package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"twitter-down/utils"
)

// Danbooru mewajibkan User-Agent yang jelas (lihat help:api).
const danbooruUserAgent = "go-social-dls/1.0 (+https://github.com/ilhmlnaa/go-social-dls)"

// danbooruMediaAsset merepresentasikan objek media_asset (berisi variants).
type danbooruMediaAsset struct {
	Variants []struct {
		Type   string `json:"type"`
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Ext    string `json:"file_ext"`
	} `json:"variants"`
}

// danbooruPost hanya memetakan field yang kita butuhkan dari /posts/{id}.json.
type danbooruPost struct {
	ID               int64              `json:"id"`
	ParentID         *int64             `json:"parent_id"`
	HasChildren      bool               `json:"has_children"`
	FileURL          string             `json:"file_url"`       // original
	LargeFileURL     string             `json:"large_file_url"` // sample/large
	PreviewFileURL   string             `json:"preview_file_url"`
	FileExt          string             `json:"file_ext"`
	FileSize         int64              `json:"file_size"`
	ImageWidth       int                `json:"image_width"`
	ImageHeight      int                `json:"image_height"`
	Rating           string             `json:"rating"`
	Source           string             `json:"source"`
	MD5              string             `json:"md5"`
	TagStringArtist  string             `json:"tag_string_artist"`
	TagStringCharacter string           `json:"tag_string_character"`
	TagStringCopyright string           `json:"tag_string_copyright"`
	MediaAsset       danbooruMediaAsset `json:"media_asset"`
}

// danbooruImage adalah bentuk keluaran yang rapi per-post.
type danbooruImage struct {
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

var danbooruIDRegex = regexp.MustCompile(`(?:/posts/|/post/show/)(\d+)`)

// extractDanbooruID menerima URL post ATAU ID mentah.
func extractDanbooruID(input string) (int64, error) {
	// ID mentah (mis. "12039889")
	if id, err := strconv.ParseInt(input, 10, 64); err == nil && id > 0 {
		return id, nil
	}
	matches := danbooruIDRegex.FindStringSubmatch(input)
	if len(matches) < 2 {
		return 0, fmt.Errorf("tidak ditemukan post id pada input")
	}
	return strconv.ParseInt(matches[1], 10, 64)
}

// danbooruClient adalah http.Client dengan timeout wajar.
var danbooruClient = &http.Client{Timeout: 20 * time.Second}

// fetchDanbooruJSON melakukan GET dengan UA + auth opsional, lalu decode JSON.
func fetchDanbooruJSON(rawURL string, target interface{}) error {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", danbooruUserAgent)

	// Auth opsional: hanya dipakai kalau di-set (berguna untuk rating tertentu / rate limit lebih longgar).
	if login := os.Getenv("DANBOORU_LOGIN"); login != "" {
		if apiKey := os.Getenv("DANBOORU_API_KEY"); apiKey != "" {
			req.SetBasicAuth(login, apiKey)
		}
	}

	resp, err := danbooruClient.Do(req)
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
func pickSample(p *danbooruPost) string {
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

// toImage mengubah danbooruPost menjadi danbooruImage yang rapi.
func toImage(p *danbooruPost, isMain bool) danbooruImage {
	return danbooruImage{
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

// DanbooruDownloadHandler mengambil gambar dari sebuah post Danbooru lewat API resmi.
//
// Query params:
//   url      : URL post (https://danbooru.donmai.us/posts/12039889) ATAU id mentah (12039889). Wajib.
//   children : "true" untuk ikut menyertakan semua post terkait (parent + children) dalam satu grup.
//
// Catatan: post dengan has_children=true tetap bisa diunduh — pembatas "child posts"
// hanyalah peringatan di UI web, API tetap mengembalikan file_url secara normal.
func DanbooruDownloadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input := r.URL.Query().Get("url")
		if input == "" {
			utils.JSONResponse(w, false, "Parameter 'url' dibutuhkan (URL post atau id)", nil)
			return
		}

		postID, err := extractDanbooruID(input)
		if err != nil {
			utils.JSONResponse(w, false, "URL/ID Danbooru tidak valid", nil)
			return
		}

		var post danbooruPost
		apiURL := fmt.Sprintf("https://danbooru.donmai.us/posts/%d.json", postID)
		if err := fetchDanbooruJSON(apiURL, &post); err != nil {
			utils.JSONResponse(w, false, fmt.Sprintf("Gagal mengambil post: %v", err), nil)
			return
		}

		if post.FileURL == "" {
			utils.JSONResponse(w, false, "Post ditemukan tetapi tidak punya file_url (mungkin dihapus/dibatasi; coba set DANBOORU_LOGIN & DANBOORU_API_KEY)", nil)
			return
		}

		includeChildren := r.URL.Query().Get("children") == "true"

		images := []danbooruImage{toImage(&post, true)}

		if includeChildren {
			// Tentukan id "akar" grup: kalau post ini punya parent, pakai parent-nya
			// agar dapat seluruh saudara; kalau tidak, pakai post ini sendiri.
			rootID := post.ID
			if post.ParentID != nil {
				rootID = *post.ParentID
			}

			// tags=parent:ROOT mengembalikan induk + seluruh anaknya dalam grup.
			var related []danbooruPost
			relURL := fmt.Sprintf("https://danbooru.donmai.us/posts.json?tags=parent:%d&limit=200", rootID)
			if err := fetchDanbooruJSON(relURL, &related); err == nil {
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

		msg := fmt.Sprintf("Berhasil mengambil %d gambar dari Danbooru", len(images))
		if post.HasChildren && !includeChildren {
			msg += " (post ini punya child posts — tambahkan &children=true untuk mengambil semuanya)"
		}

		utils.JSONResponse(w, true, msg, map[string]interface{}{
			"post_id":       post.ID,
			"has_children":  post.HasChildren,
			"count":         len(images),
			"with_children": includeChildren,
			"images":        images,
		})
	}
}
