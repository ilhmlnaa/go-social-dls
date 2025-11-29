#!/usr/bin/env python3
"""
Twitter Scraper using twscrape library
This script fetches tweet details including photos
"""
import asyncio
import json
import sys
import os
from twscrape import API, gather
from twscrape.logger import set_log_level

async def get_tweet_photos(tweet_id):
    """
    Get tweet photos using twscrape
    
    Args:
        tweet_id: Twitter tweet ID
        
    Returns:
        JSON string with success status and photo URLs
    """
    try:
        # Initialize API
        api = API()
        
        # Optional: reduce logging noise
        set_log_level("ERROR")
        
        # Check if we have any accounts configured
        accounts = await api.pool.accounts_info()
        if not accounts:
            return json.dumps({
                "success": False,
                "error": "No Twitter accounts configured. Run setup first.",
                "message": "Please add Twitter account using: twscrape add_accounts"
            })
        
        # Fetch tweet
        tweet = await api.tweet_by_id(int(tweet_id))
        
        if not tweet:
            return json.dumps({
                "success": False,
                "error": "Tweet not found",
                "message": f"Tweet with ID {tweet_id} not found or deleted"
            })
        
        # Extract photo URLs
        photos = []
        if hasattr(tweet, 'media') and tweet.media:
            for media in tweet.media.photos:
                # Get highest quality photo URL
                photo_url = media.url
                # Replace to get large size
                if '&name=' in photo_url:
                    photo_url = photo_url.split('&name=')[0] + '&name=large'
                elif '?format=' in photo_url:
                    photo_url = photo_url + '&name=large'
                photos.append(photo_url)
        
        if not photos:
            return json.dumps({
                "success": False,
                "error": "No photos found",
                "message": "This tweet does not contain any photos"
            })
        
        return json.dumps({
            "success": True,
            "tweet_id": tweet_id,
            "photos": photos,
            "photo_count": len(photos),
            "tweet_text": tweet.rawContent[:100] if hasattr(tweet, 'rawContent') else ""
        })
        
    except Exception as e:
        return json.dumps({
            "success": False,
            "error": str(e),
            "message": f"Failed to fetch tweet: {str(e)}"
        })

def main():
    """Main function"""
    if len(sys.argv) < 2:
        print(json.dumps({
            "success": False,
            "error": "Missing tweet ID",
            "message": "Usage: python twitter_scraper.py <tweet_id>"
        }))
        sys.exit(1)
    
    tweet_id = sys.argv[1]
    
    # Run async function
    result = asyncio.run(get_tweet_photos(tweet_id))
    print(result)

if __name__ == "__main__":
    main()
