# 🍪 Cookies Setup Guide

Complete guide for setting up cookies for Twitter and Facebook downloaders.

## 📋 Overview

This API uses cookies from your browser to access Twitter and Facebook. This method is:
- ✅ Simple - Just export and save
- ✅ Reliable - Uses your actual browser session
- ✅ No Cloudflare blocking
- ✅ No bot detection issues

## 🐦 Twitter Setup

### Step 1: Install Cookie-Editor Extension

Choose your browser:

- **Chrome/Edge:** [Cookie-Editor](https://chrome.google.com/webstore/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm)
- **Firefox:** [Cookie-Editor](https://addons.mozilla.org/en-US/firefox/addon/cookie-editor/)

### Step 2: Export Twitter Cookies

1. **Login to Twitter/X** in your browser (https://twitter.com or https://x.com)
2. **Make sure you're logged in** - you should see your timeline
3. **Click Cookie-Editor icon** in browser toolbar
4. **Click "Export"** button
5. **Select "JSON"** format
6. **Copy** the JSON output

### Step 3: Save Cookies

Create file: `cookies/twitter.json`

```json
[
  {
    "domain": ".x.com",
    "name": "auth_token",
    "value": "your_auth_token_here",
    ...
  },
  {
    "domain": ".x.com",
    "name": "ct0",
    "value": "your_csrf_token_here",
    ...
  }
  // ... more cookies
]
```

**Important:** File must be named exactly `twitter.json` in the `cookies/` directory.

### Step 4: Verify

```bash
# Check file exists
ls -la cookies/twitter.json

# Validate JSON format
cat cookies/twitter.json | python3 -m json.tool
```

---

## 📘 Facebook Setup (Optional)

### Same Process as Twitter

1. **Login to Facebook** (https://facebook.com)
2. **Export cookies** with Cookie-Editor
3. **Save as** `cookies/facebook.json`

### Required Cookies

Facebook needs these cookies:
- `c_user` - User ID
- `xs` - Session token
- `datr` - Device token

---

## 🔒 Security Best Practices

### ⚠️ Important Warnings

1. **Don't commit cookies to Git**
   ```bash
   # Already in .gitignore
   cookies/*.json
   ```

2. **Don't share cookies**
   - Cookies = Full account access
   - Treat them like passwords

3. **Use dedicated accounts**
   - Don't use your personal account
   - Create a separate account for scraping

### 🔄 Cookie Expiration

Cookies typically last **2-4 weeks**. When expired:

1. **Login again** in browser
2. **Export new cookies**
3. **Replace** old cookie files
4. **Restart server** (no rebuild needed)

### ⚡ Quick Cookie Refresh

```bash
# Export cookies from browser
# Replace the files
cp ~/Downloads/twitter-cookies.json cookies/twitter.json

# Restart server
pkill server && ./server
```

---

## 🐳 Docker Setup

### Method 1: Volume Mount (Development)

```bash
docker run -d \
  -v $(pwd)/cookies:/app/cookies \
  -p 3005:3005 \
  social-media-downloader
```

### Method 2: Copy Files (Production)

```dockerfile
# In Dockerfile
COPY cookies/twitter.json /app/cookies/twitter.json
COPY cookies/facebook.json /app/cookies/facebook.json
```

### Method 3: Environment Variables (Best for Production)

```bash
# Base64 encode cookies
cat cookies/twitter.json | base64 > twitter_cookies_b64.txt

# Set as environment variable
docker run -d \
  -e TWITTER_COOKIES_B64="$(cat twitter_cookies_b64.txt)" \
  -p 3005:3005 \
  social-media-downloader
```

Then decode in code:

```go
// In internal/utils/cookies.go
func LoadCookiesFromEnv(platform string) (map[string]string, error) {
    envKey := strings.ToUpper(platform) + "_COOKIES_B64"
    encoded := os.Getenv(envKey)
    
    if encoded == "" {
        return nil, fmt.Errorf("env variable %s not set", envKey)
    }
    
    decoded, err := base64.StdEncoding.DecodeString(encoded)
    if err != nil {
        return nil, err
    }
    
    // Parse JSON...
}
```

---

## ✅ Verification

### Test Twitter Endpoint

```bash
curl "http://localhost:3005/api/v1/twitter?url=https://twitter.com/user/status/1234567890"
```

**Success:**
```json
{
  "success": true,
  "message": "Successfully fetched 2 photo(s)",
  "data": ["https://image1.jpg", "https://image2.jpg"]
}
```

**Error (Invalid Cookies):**
```json
{
  "success": false,
  "message": "Failed to initialize Twitter service",
  "error": "missing required cookie: ct0"
}
```

---

## 🔍 Troubleshooting

### "Failed to load cookies"

**Problem:** File not found or invalid JSON

**Solution:**
```bash
# Check file exists
ls -la cookies/twitter.json

# Validate JSON
cat cookies/twitter.json | python3 -m json.tool
```

### "Missing required cookie"

**Problem:** Exported cookies incomplete

**Solution:**
1. Make sure you're **logged in** to Twitter before exporting
2. Export cookies **while on Twitter website** (not other sites)
3. Try **logging out and back in**, then export again

### "HTTP 401 Unauthorized"

**Problem:** Cookies expired

**Solution:**
1. Check if you're still logged in to Twitter
2. Export fresh cookies
3. Replace `cookies/twitter.json`
4. Restart server

### "Tweet not found" (but tweet exists)

**Problem:** Cookies might be expired or tweet is private

**Solution:**
1. Try accessing the tweet in browser (logged in)
2. If you can see it, export new cookies
3. If tweet is private, cookies from that account needed

---

## 📝 Cookie File Format

### Valid Format

```json
[
  {
    "domain": ".x.com",
    "expirationDate": 1798985426.264508,
    "hostOnly": false,
    "httpOnly": true,
    "name": "auth_token",
    "path": "/",
    "sameSite": "no_restriction",
    "secure": true,
    "session": false,
    "storeId": null,
    "value": "4c4caaab62a0b32b72908a98c74c88581f5d560b"
  }
  // More cookies...
]
```

### Invalid Formats

❌ **Not an array:**
```json
{
  "auth_token": "...",
  "ct0": "..."
}
```

❌ **With quotes around value:**
```env
TWITTER_AUTH_TOKEN="abc123"
```

❌ **Partial export:**
```json
[
  {
    "name": "auth_token"
    // Missing value and other fields
  }
]
```

---

## 🆘 Still Having Issues?

1. **Check logs:**
   ```bash
   ./server 2>&1 | tee server.log
   ```

2. **Test with known working tweet:**
   ```bash
   curl "http://localhost:3005/api/v1/twitter?url=https://twitter.com/Twitter/status/20"
   ```

3. **Verify cookies are fresh:**
   - Login to Twitter in browser
   - Can you see tweets normally?
   - If yes, export new cookies
   - If no, account might be suspended

4. **Check cookie file permissions:**
   ```bash
   chmod 644 cookies/twitter.json
   ```

---

**Need more help?** Open an issue on GitHub with:
- Error messages from logs
- Steps you've tried
- Platform (Twitter/Facebook)
- Browser used for cookie export

---

**Last Updated:** 2025-11-29
