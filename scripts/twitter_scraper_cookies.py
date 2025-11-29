#!/usr/bin/env python3
"""
Twitter Scraper using cookies from browser
This script uses exported cookies to fetch tweet details
"""
import json
import sys
import requests
from http.cookies import SimpleCookie

def load_cookies_from_file(cookie_file="cookie.json"):
    """Load cookies from JSON file exported by Cookie Editor"""
    try:
        with open(cookie_file, 'r') as f:
            cookies_json = json.load(f)
        
        # Convert to requests-compatible format
        cookies = {}
        for cookie in cookies_json:
            cookies[cookie['name']] = cookie['value']
        
        return cookies
    except Exception as e:
        return None

def get_tweet_photos(tweet_id, cookies):
    """
    Fetch tweet photos using Twitter GraphQL API with cookies
    
    Args:
        tweet_id: Twitter tweet ID
        cookies: Dict of cookies from browser
        
    Returns:
        JSON response with success status and photo URLs
    """
    try:
        # Twitter GraphQL endpoint for tweet details
        url = "https://twitter.com/i/api/graphql/0hWvDhmW8YQ-S_ib3azIrw/TweetResultByRestId"
        
        # Get CSRF token from cookies
        csrf_token = cookies.get('ct0', '')
        auth_token = cookies.get('auth_token', '')
        
        if not csrf_token or not auth_token:
            return {
                "success": False,
                "error": "Missing required cookies",
                "message": "auth_token or ct0 (CSRF token) not found in cookies"
            }
        
        # Request headers - mimic real browser
        headers = {
            'authority': 'twitter.com',
            'accept': '*/*',
            'accept-language': 'en-US,en;q=0.9',
            'authorization': 'Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA',
            'content-type': 'application/json',
            'referer': f'https://twitter.com/i/status/{tweet_id}',
            'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
            'x-csrf-token': csrf_token,
            'x-twitter-active-user': 'yes',
            'x-twitter-auth-type': 'OAuth2Session',
            'x-twitter-client-language': 'en',
        }
        
        # GraphQL variables
        variables = {
            "tweetId": str(tweet_id),
            "withCommunity": False,
            "includePromotedContent": False,
            "withVoice": False
        }
        
        features = {
            "creator_subscriptions_tweet_preview_api_enabled": True,
            "communities_web_enable_tweet_community_results_fetch": True,
            "c9s_tweet_anatomy_moderator_badge_enabled": True,
            "articles_preview_enabled": True,
            "responsive_web_edit_tweet_api_enabled": True,
            "graphql_is_translatable_rweb_tweet_is_translatable_enabled": True,
            "view_counts_everywhere_api_enabled": True,
            "longform_notetweets_consumption_enabled": True,
            "responsive_web_twitter_article_tweet_consumption_enabled": True,
            "tweet_awards_web_tipping_enabled": False,
            "creator_subscriptions_quote_tweet_preview_enabled": False,
            "freedom_of_speech_not_reach_fetch_enabled": True,
            "standardized_nudges_misinfo": True,
            "tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": True,
            "rweb_video_timestamps_enabled": True,
            "longform_notetweets_rich_text_read_enabled": True,
            "longform_notetweets_inline_media_enabled": True,
            "rweb_tipjar_consumption_enabled": True,
            "responsive_web_graphql_exclude_directive_enabled": True,
            "verified_phone_label_enabled": False,
            "responsive_web_graphql_skip_user_profile_image_extensions_enabled": False,
            "responsive_web_graphql_timeline_navigation_enabled": True,
            "responsive_web_enhance_cards_enabled": False,
            "tweetypie_unmention_optimization_enabled": True,
            "responsive_web_media_download_video_enabled": True
        }
        
        # Build request params
        params = {
            'variables': json.dumps(variables),
            'features': json.dumps(features)
        }
        
        # Make request
        response = requests.get(url, headers=headers, cookies=cookies, params=params)
        
        if response.status_code != 200:
            return {
                "success": False,
                "error": f"HTTP {response.status_code}",
                "message": f"Failed to fetch tweet: {response.text[:200]}"
            }
        
        # Parse response
        data = response.json()
        
        # Navigate to tweet data
        tweet_result = data.get('data', {}).get('tweetResult', {}).get('result', {})
        
        if not tweet_result or tweet_result.get('__typename') == 'TweetUnavailable':
            return {
                "success": False,
                "error": "Tweet not found",
                "message": "Tweet not found, deleted, or private"
            }
        
        # Extract legacy tweet data
        legacy = tweet_result.get('legacy', {})
        
        # Get media (photos)
        extended_entities = legacy.get('extended_entities', {})
        media_list = extended_entities.get('media', [])
        
        photos = []
        for media in media_list:
            if media.get('type') == 'photo':
                # Get highest quality image
                media_url = media.get('media_url_https', '')
                if media_url:
                    # Replace to get large size
                    if '?' in media_url:
                        media_url = media_url.split('?')[0]
                    media_url = media_url + '?name=large'
                    photos.append(media_url)
        
        if not photos:
            return {
                "success": False,
                "error": "No photos found",
                "message": "This tweet does not contain any photos"
            }
        
        # Get tweet text
        tweet_text = legacy.get('full_text', '')
        
        return {
            "success": True,
            "tweet_id": str(tweet_id),
            "photos": photos,
            "photo_count": len(photos),
            "tweet_text": tweet_text[:100] if tweet_text else ""
        }
        
    except Exception as e:
        return {
            "success": False,
            "error": str(e),
            "message": f"Failed to fetch tweet: {str(e)}"
        }

def main():
    """Main function"""
    if len(sys.argv) < 2:
        print(json.dumps({
            "success": False,
            "error": "Missing tweet ID",
            "message": "Usage: python twitter_scraper_cookies.py <tweet_id>"
        }))
        sys.exit(1)
    
    tweet_id = sys.argv[1]
    
    # Load cookies from file
    cookies = load_cookies_from_file("cookie.json")
    
    if not cookies:
        print(json.dumps({
            "success": False,
            "error": "Failed to load cookies",
            "message": "Could not load cookie.json file. Make sure it exists and is valid JSON."
        }))
        sys.exit(1)
    
    # Fetch tweet photos
    result = get_tweet_photos(tweet_id, cookies)
    print(json.dumps(result))

if __name__ == "__main__":
    main()
