#!/usr/bin/env python3
"""
Instagram Full Resolution Extractor (Python + Pyppeteer version)
Lebih ringan daripada Playwright.
"""

import json
import re
import asyncio
from pyppeteer import launch


def load_cookies(cookie_file):
    with open(cookie_file, "r") as f:
        cookies_data = json.load(f)

    pyppeteer_cookies = []

    for c in cookies_data:
        cookie = {
            "name": c["name"],
            "value": c["value"],
            "domain": c.get("domain", ".instagram.com"),
            "path": c.get("path", "/"),
        }

        if "expirationDate" in c:
            cookie["expires"] = c["expirationDate"]
        if "httpOnly" in c:
            cookie["httpOnly"] = c["httpOnly"]
        if "secure" in c:
            cookie["secure"] = c["secure"]

        # ---- FIX: handle sameSite None/null ----
        raw_same = c.get("sameSite")
        if raw_same is None:
            raw = "none"
        else:
            raw = str(raw_same).lower()

        if raw in ["none", "no_restriction", "unspecified"]:
            cookie["sameSite"] = "None"
        elif raw == "lax":
            cookie["sameSite"] = "Lax"
        elif raw == "strict":
            cookie["sameSite"] = "Strict"
        else:
            cookie["sameSite"] = "None"
        # ----------------------------------------

        pyppeteer_cookies.append(cookie)

    return pyppeteer_cookies

    with open(cookie_file, "r") as f:
        cookies_data = json.load(f)

    pyppeteer_cookies = []

    for c in cookies_data:
        cookie = {
            "name": c["name"],
            "value": c["value"],
            "domain": c.get("domain", ".instagram.com"),
            "path": c.get("path", "/"),
        }

        if "expirationDate" in c:
            cookie["expires"] = c["expirationDate"]
        if "httpOnly" in c:
            cookie["httpOnly"] = c["httpOnly"]
        if "secure" in c:
            cookie["secure"] = c["secure"]

        # Normalize sameSite for Pyppeteer
        raw = c.get("sameSite", "none").lower()
        if raw in ["none", "no_restriction", "unspecified"]:
            cookie["sameSite"] = "None"
        elif raw == "lax":
            cookie["sameSite"] = "Lax"
        elif raw == "strict":
            cookie["sameSite"] = "Strict"

        pyppeteer_cookies.append(cookie)

    return pyppeteer_cookies


async def extract_full_res(instagram_url, cookie_file):
    print(f"[*] Loading cookies from {cookie_file}...")
    cookies = load_cookies(cookie_file)
    print(f"[*] Loaded {len(cookies)} cookies")

    print("[*] Launching browser (pyppeteer)...")
    browser = await launch(
        headless=True,
        args=[
            "--no-sandbox",
            "--disable-setuid-sandbox",
            "--disable-blink-features=AutomationControlled",
        ],
    )

    page = await browser.newPage()

    # Apply cookies
    for c in cookies:
        try:
            await page.setCookie(c)
        except Exception:
            pass

    print("[*] Opening Instagram page...")
    await page.goto(instagram_url, {"waitUntil": "domcontentloaded"})

    print("[*] Waiting for Instagram scripts...")
    await page.waitFor(2500)

    # METHOD 1 — Extract CDN URLs from HTML
    print("\n[Method 1] HTML CDN extraction...")

    html = await page.content()

    cdn_pattern = r'https://scontent[^"\s]+cdninstagram\.com[^"\s]+\.jpg\?[^"\s]+'
    found = re.findall(cdn_pattern, html)

    print(f"[✓] Found {len(found)} URLs from HTML")

    cleaned = [
        u.replace("\\u0026", "&").replace("&amp;", "&")
        for u in found
    ]

    full_res = [
        u for u in cleaned
        if "oh=" in u and "oe=" in u and "s640x640" not in u
    ]

    if full_res:
        await browser.close()
        return full_res

    # METHOD 2 — img tags
    print("\n[Method 2] Extracting <img> elements...")

    img_srcs = await page.evaluate(
        """() => {
            return Array.from(document.querySelectorAll('img'))
                        .map(img => img.src)
                        .filter(src => src.includes('cdninstagram.com'));
        }"""
    )

    await browser.close()

    if img_srcs:
        print(f"[✓] Found {len(img_srcs)} image sources from DOM")
        return img_srcs

    print("[!] No images found.")
    return []


async def main():
    INSTAGRAM_URL = "https://www.instagram.com/p/DRPULnJE7O0/"
    COOKIES_FILE = "cookies/instagram.json"

    print("=" * 80)
    print("Instagram Full Resolution Extractor (Pyppeteer Version)")
    print("=" * 80)
    print(f"URL: {INSTAGRAM_URL}")
    print(f"Cookies: {COOKIES_FILE}")
    print("=" * 80)

    urls = await extract_full_res(INSTAGRAM_URL, COOKIES_FILE)

    print("\nRESULTS")
    print("=" * 80)

    if urls:
        for i, url in enumerate(urls, 1):
            print(f"\n[{i}] {url}")

            if "s640x640" in url:
                print("    ⚠️  size-limited (640px)")
            elif "s1080x1080" in url:
                print("    ⚠️  size-limited (1080px)")
            else:
                print("    ✓ Full resolution")

            if "oh=" in url and "oe=" in url:
                print("    ✓ Valid IG URL token")
    else:
        print("[!] No URLs extracted")


if __name__ == "__main__":
    asyncio.get_event_loop().run_until_complete(main())
