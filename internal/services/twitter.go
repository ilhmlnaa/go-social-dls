package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"twitter-down/internal/utils"
)

type TwitterService struct {
	client  *http.Client
	cookies map[string]string
}


type TwitterGraphQLResponse struct {
	Data struct {
		TweetResult struct {
			Result struct {
				Typename string `json:"__typename"`
				Legacy   struct {
					FullText         string `json:"full_text"`
					ExtendedEntities struct {
						Media []struct {
							Type         string `json:"type"`
							MediaURLHTTPS string `json:"media_url_https"`
						} `json:"media"`
					} `json:"extended_entities"`
				} `json:"legacy"`
			} `json:"result"`
		} `json:"tweetResult"`
	} `json:"data"`
}

func NewTwitterService(cookiesDir string) (*TwitterService, error) {
	cookies, err := utils.LoadCookiesOptional(cookiesDir, "twitter")
	if err != nil {
		return nil, err
	}

	return &TwitterService{
		client:  &http.Client{},
		cookies: cookies,
	}, nil
}

func (s *TwitterService) GetTweetPhotos(tweetID string) ([]string, error) {
	baseURL := "https://twitter.com/i/api/graphql/0hWvDhmW8YQ-S_ib3azIrw/TweetResultByRestId"

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("referer", fmt.Sprintf("https://twitter.com/i/status/%s", tweetID))
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	
	if ct0, ok := s.cookies["ct0"]; ok {
		req.Header.Set("x-csrf-token", ct0)
	}
	
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	req.Header.Set("x-twitter-client-language", "en")

	for name, value := range s.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	variables := map[string]interface{}{
		"tweetId":                     tweetID,
		"withCommunity":               false,
		"includePromotedContent":      false,
		"withVoice":                   false,
	}

	features := map[string]interface{}{
		"creator_subscriptions_tweet_preview_api_enabled":                       true,
		"communities_web_enable_tweet_community_results_fetch":                  true,
		"c9s_tweet_anatomy_moderator_badge_enabled":                            true,
		"articles_preview_enabled":                                              true,
		"responsive_web_edit_tweet_api_enabled":                                 true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":            true,
		"view_counts_everywhere_api_enabled":                                    true,
		"longform_notetweets_consumption_enabled":                              true,
		"responsive_web_twitter_article_tweet_consumption_enabled":              true,
		"tweet_awards_web_tipping_enabled":                                      false,
		"creator_subscriptions_quote_tweet_preview_enabled":                     false,
		"freedom_of_speech_not_reach_fetch_enabled":                            true,
		"standardized_nudges_misinfo":                                          true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"rweb_video_timestamps_enabled":                                         true,
		"longform_notetweets_rich_text_read_enabled":                           true,
		"longform_notetweets_inline_media_enabled":                             true,
		"rweb_tipjar_consumption_enabled":                                      true,
		"responsive_web_graphql_exclude_directive_enabled":                      true,
		"verified_phone_label_enabled":                                         false,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":     false,
		"responsive_web_graphql_timeline_navigation_enabled":                    true,
		"responsive_web_enhance_cards_enabled":                                  false,
		"tweetypie_unmention_optimization_enabled":                             true,
		"responsive_web_media_download_video_enabled":                          true,
	}

	variablesJSON, _ := json.Marshal(variables)
	featuresJSON, _ := json.Marshal(features)

	q := req.URL.Query()
	q.Add("variables", string(variablesJSON))
	q.Add("features", string(featuresJSON))
	req.URL.RawQuery = q.Encode()

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		if len(s.cookies) == 0 || s.cookies["auth_token"] == "" || s.cookies["ct0"] == "" {
			return nil, fmt.Errorf("Twitter service requires authentication cookies (auth_token, ct0)")
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result TwitterGraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Data.TweetResult.Result.Typename == "TweetUnavailable" {
		return nil, fmt.Errorf("tweet not found or unavailable")
	}

	media := result.Data.TweetResult.Result.Legacy.ExtendedEntities.Media
	var photos []string
	for _, m := range media {
		if m.Type == "photo" {
			mediaURL := m.MediaURLHTTPS
			if !strings.Contains(mediaURL, "?") {
				mediaURL += "?name=large"
			} else {
				mediaURL = strings.Split(mediaURL, "?")[0] + "?name=large"
			}
			photos = append(photos, mediaURL)
		}
	}

	if len(photos) == 0 {
		return nil, fmt.Errorf("no photos found in tweet")
	}

	return photos, nil
}

func ExtractTweetID(tweetURL string) (string, error) {
	u, err := url.Parse(tweetURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	re := regexp.MustCompile(`status/(\d+)`)
	matches := re.FindStringSubmatch(u.Path)
	if len(matches) < 2 {
		return "", fmt.Errorf("tweet ID not found in URL")
	}

	return matches[1], nil
}
