#!/usr/bin/env python3
"""
Test Instagram full resolution download using Playwright (headless browser)
This simulates what web-based Instagram downloaders do
"""

import json
import sys
import re
from playwright.sync_api import sync_playwright

def load_cookies(cookie_file):
    with open(cookie_file, 'r') as f:
        cookies_data = json.load(f)

    playwright_cookies = []

    for cookie in cookies_data:
        playwright_cookie = {
            "name": cookie["name"],
            "value": cookie["value"],
            "domain": cookie.get("domain", ".instagram.com"),
            "path": cookie.get("path", "/")
        }

        if "expirationDate" in cookie:
            playwright_cookie["expires"] = int(cookie["expirationDate"])

        if "httpOnly" in cookie:
            playwright_cookie["httpOnly"] = cookie["httpOnly"]

        if "secure" in cookie:
            playwright_cookie["secure"] = cookie["secure"]

        # --- FIX sameSite ---
        raw_same_site = cookie.get("sameSite")
        if raw_same_site:
            s = raw_same_site.lower()
            if s in ["no_restriction", "unspecified", "none"]:
                playwright_cookie["sameSite"] = "None"
            elif s == "lax":
                playwright_cookie["sameSite"] = "Lax"
            elif s == "strict":
                playwright_cookie["sameSite"] = "Strict"
        else:
            playwright_cookie["sameSite"] = "None"
        # ---------------------

        playwright_cookies.append(playwright_cookie)

    return playwright_cookies

    """Load cookies from Cookie-Editor JSON format and convert to Playwright format"""
    with open(cookie_file, 'r') as f:
        cookies_data = json.load(f)
    
    # Convert to Playwright cookie format
    playwright_cookies = []
    for cookie in cookies_data:
        playwright_cookie = {
            'name': cookie['name'],
            'value': cookie['value'],
            'domain': cookie.get('domain', '.instagram.com'),
            'path': cookie.get('path', '/'),
        }
        
        # Add optional fields
        if 'expirationDate' in cookie:
            playwright_cookie['expires'] = int(cookie['expirationDate'])
        if 'httpOnly' in cookie:
            playwright_cookie['httpOnly'] = cookie['httpOnly']
        if 'secure' in cookie:
            playwright_cookie['secure'] = cookie['secure']
        if 'sameSite' in cookie and cookie['sameSite']:
            playwright_cookie['sameSite'] = cookie['sameSite']
            
        playwright_cookies.append(playwright_cookie)
    
    return playwright_cookies

def extract_full_res_urls(instagram_url, cookies_file):
    """Extract full resolution image URLs using Playwright"""
    
    print(f"[*] Loading cookies from {cookies_file}...")
    cookies = load_cookies(cookies_file)
    print(f"[*] Loaded {len(cookies)} cookies")
    
    with sync_playwright() as p:
        print("[*] Launching browser...")
        # Launch browser in headless mode
        browser = p.chromium.launch(headless=True)
        
        # Create context with cookies
        context = browser.new_context()
        context.add_cookies(cookies)
        
        print("[*] Opening Instagram page...")
        page = context.new_page()
        
        # Navigate to Instagram post
        page.goto(instagram_url, wait_until="domcontentloaded", timeout=30000)
        
        print("[*] Page loaded, waiting for content...")
        page.wait_for_timeout(3000)  # Wait 3 seconds for JavaScript to execute
        
        # Method 1: Extract from window._sharedData
        print("\n[Method 1] Extracting from window._sharedData...")
        shared_data = page.evaluate("""() => {
            return window._sharedData || null;
        }""")
        
        if shared_data:
            print(f"[✓] Found _sharedData")
            # Navigate through the structure
            try:
                entry_data = shared_data.get('entry_data', {})
                post_page = entry_data.get('PostPage', [])
                if post_page:
                    graphql = post_page[0].get('graphql', {})
                    shortcode_media = graphql.get('shortcode_media', {})
                    
                    # Get display_resources (array of different resolutions)
                    display_resources = shortcode_media.get('display_resources', [])
                    if display_resources:
                        print(f"[✓] Found {len(display_resources)} display_resources:")
                        for i, res in enumerate(display_resources):
                            print(f"    [{i}] {res.get('config_width')}x{res.get('config_height')}: {res.get('src', '')[:100]}...")
                        
                        # Get highest resolution (usually last one)
                        highest = display_resources[-1]
                        print(f"\n[✓] Highest resolution: {highest.get('config_width')}x{highest.get('config_height')}")
                        return [highest.get('src')]
                    
                    # Fallback to display_url
                    display_url = shortcode_media.get('display_url', '')
                    if display_url:
                        print(f"[✓] Found display_url (fallback): {display_url[:100]}...")
                        return [display_url]
            except Exception as e:
                print(f"[!] Error parsing _sharedData: {e}")
        
        # Method 2: Extract from page content / network requests
        print("\n[Method 2] Extracting from page HTML...")
        html = page.content()
        
        # Find all Instagram CDN URLs
        url_pattern = r'https://scontent[^"\s]+cdninstagram\.com[^"\s]+\.jpg\?[^"\s]+'
        urls = re.findall(url_pattern, html)
        
        print(f"[✓] Found {len(urls)} CDN URLs in HTML")
        
        # Analyze URLs
        best_urls = []
        for i, url in enumerate(urls):
            # Clean up
            url = url.replace('\\u0026', '&').replace('&amp;', '&')
            
            # Check for size parameters
            has_size = any(size in url for size in ['s150x150', 's320x320', 's480x480', 's640x640', 's1080x1080'])
            param_count = url.count('&')
            
            if i < 3:  # Log first 3
                print(f"    [{i+1}] len={len(url)}, params={param_count}, hasSize={has_size}")
                print(f"         {url[:150]}...")
            
            # Prefer URLs without size restrictions
            if not has_size and param_count > 5:
                best_urls.append(url)
        
        if best_urls:
            print(f"\n[✓] Found {len(best_urls)} URLs without size restrictions!")
            return best_urls
        
        # Method 3: Try to get from img tags directly
        print("\n[Method 3] Extracting from img tags...")
        img_srcs = page.evaluate("""() => {
            const imgs = document.querySelectorAll('img');
            return Array.from(imgs).map(img => img.src).filter(src => src.includes('cdninstagram.com'));
        }""")
        
        if img_srcs:
            print(f"[✓] Found {len(img_srcs)} images from DOM:")
            for i, src in enumerate(img_srcs[:3]):
                print(f"    [{i+1}] {src[:150]}...")
            return img_srcs
        
        print("[!] No URLs found")
        return []

if __name__ == '__main__':
    # Configuration
    INSTAGRAM_URL = "https://www.instagram.com/p/DRPULnJE7O0/"
    COOKIES_FILE = "cookies/instagram.json"
    
    if len(sys.argv) > 1:
        INSTAGRAM_URL = sys.argv[1]
    
    print("=" * 80)
    print("Instagram Full Resolution URL Extractor (Playwright Test)")
    print("=" * 80)
    print(f"URL: {INSTAGRAM_URL}")
    print(f"Cookies: {COOKIES_FILE}")
    print("=" * 80)
    
    try:
        urls = extract_full_res_urls(INSTAGRAM_URL, COOKIES_FILE)
        
        print("\n" + "=" * 80)
        print("RESULTS:")
        print("=" * 80)
        
        if urls:
            print(f"[✓] Successfully extracted {len(urls)} URL(s):")
            for i, url in enumerate(urls):
                print(f"\n[{i+1}] {url}")
                
                # Check resolution indicators
                if 's640x640' in url:
                    print("    ⚠️  Contains s640x640 (640px limit)")
                elif 's1080x1080' in url:
                    print("    ⚠️  Contains s1080x1080 (1080px limit)")
                else:
                    print("    ✓ No size restriction found!")
                
                if 'oh=' in url and 'oe=' in url:
                    print("    ✓ Has validation parameters (oh, oe)")
        else:
            print("[!] No URLs extracted")
            
    except Exception as e:
        print(f"\n[ERROR] {e}")
        import traceback
        traceback.print_exc()
