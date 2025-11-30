# Instagram Downloader Usage Guide

## 🎯 How It Works

Instagram API endpoint returns **complete image URLs** with all required validation parameters. However, these URLs must be downloaded **from the client-side** (browser) where user has an active Instagram session.

## 🔗 API Endpoint

```
GET /api/v1/instagram?url={instagram_post_url}
```

**Example:**
```bash
curl "http://localhost:3005/api/v1/instagram?url=https://www.instagram.com/p/DRPULnJE7O0/"
```

**Response:**
```json
{
  "success": true,
  "message": "Successfully fetched 1 image(s)",
  "data": [
    "https://scontent.cdninstagram.com/v/t51.82787-15/583246816_17916563028214195_6061616357238707768_n.jpg?_nc_cat=102&_nc_gid=...&oh=...&oe=...&stp=dst-jpg_e35_tt6"
  ]
}
```

## 📥 How to Download (Client-Side)

### ✅ Method 1: Direct Browser Download (Recommended)

User must be logged in to Instagram in their browser, then:

```javascript
// Fetch image URL from API
const response = await fetch('/api/v1/instagram?url=https://www.instagram.com/p/ABC123/');
const data = await response.json();
const imageURL = data.data[0];

// Open in new tab - browser will handle download with user's Instagram session
window.open(imageURL, '_blank');
```

### ✅ Method 2: Fetch with Credentials

```javascript
const response = await fetch('/api/v1/instagram?url=...');
const data = await response.json();
const imageURL = data.data[0];

// Download with credentials
const imageBlob = await fetch(imageURL, {
  credentials: 'include',
  headers: {
    'Referer': 'https://www.instagram.com/'
  }
}).then(r => r.blob());

// Create download link
const url = URL.createObjectURL(imageBlob);
const a = document.createElement('a');
a.href = url;
a.download = 'instagram_image.jpg';
a.click();
URL.revokeObjectURL(url);
```

### ✅ Method 3: Server-Side (dengan Proxy User's Session)

If you need server-side download, you must proxy the user's Instagram session cookies:

```javascript
// Client sends their Instagram cookies to your backend
const cookies = document.cookie; // User's Instagram session

await fetch('/your-backend/download-instagram', {
  method: 'POST',
  body: JSON.stringify({ 
    url: 'https://www.instagram.com/p/...',
    cookies: cookies 
  })
});

// Backend uses these cookies to download
```

## ⚠️ Important Notes

### Why Client-Side Download?

Instagram CDN URLs contain cryptographic parameters (`oh`, `oe`) that are:
- **Session-specific**: Tied to exact browser session
- **IP-validated**: Validated against user's IP address
- **Time-sensitive**: May expire after some time
- **Context-bound**: Validated against cookies, referrer, user-agent

These parameters cannot be used from a different context (like your server).

### URL Parameters Explained

Complete Instagram URL has these parameters:

| Parameter | Purpose |
|-----------|---------|
| `stp` | Image size/transformation (e.g., `dst-jpg_e35_tt6` = full res) |
| `_nc_cat` | CDN category/routing |
| `_nc_gid` | Session/group identifier |
| `_nc_ht` | CDN hostname |
| `_nc_ohc` | Hash/checksum |
| `_nc_oc` | Origin context |
| `_nc_sid` | Session ID |
| `_nc_zt` | Zone/timestamp |
| `ccb` | Cache control behavior |
| `efg` | Encoding flags (JSON-encoded) |
| `oh` | **Cryptographic signature** (session-bound) |
| `oe` | **Expiration timestamp** (session-bound) |

The `oh` and `oe` parameters are the critical ones that prevent server-side download.

## 🛠️ Recommended Implementation

### Frontend (React/Vue/etc)

```javascript
async function downloadInstagramImage(postUrl) {
  try {
    // 1. Get image URL from API
    const response = await fetch(`/api/v1/instagram?url=${encodeURIComponent(postUrl)}`);
    const data = await response.json();
    
    if (!data.success) {
      throw new Error(data.message);
    }
    
    const imageURL = data.data[0];
    
    // 2. Download in browser (user must be logged in to Instagram)
    window.open(imageURL, '_blank');
    
    // Alternative: Fetch and download programmatically
    // const blob = await fetch(imageURL, { credentials: 'include' }).then(r => r.blob());
    // const url = URL.createObjectURL(blob);
    // const a = document.createElement('a');
    // a.href = url;
    // a.download = 'instagram.jpg';
    // a.click();
    // URL.revokeObjectURL(url);
    
  } catch (error) {
    console.error('Failed to download Instagram image:', error);
    alert('Please make sure you are logged in to Instagram');
  }
}
```

### User Instructions

Tell your users:

1. **Login to Instagram** in the same browser
2. **Keep Instagram tab open** (session active)
3. **Click download button** in your app
4. Image will open in new tab or download automatically

## 🔒 Security Considerations

### Don't Store Instagram URLs

Instagram URLs with parameters are:
- ⏱️ **Time-limited**: May expire
- 🔐 **Session-bound**: Won't work for other users
- 🚫 **Non-transferable**: Cannot be shared

Always fetch fresh URLs from API when needed.

### Don't Share User Cookies

Never:
- Send user's Instagram cookies to your server
- Store user's Instagram session
- Share URLs between different users

Each user must use their own Instagram session.

## 📊 Image Quality

API returns **full resolution** URLs with `stp=dst-jpg_e35_tt6`:

- `dst-jpg_e35` = Full resolution encoding
- `tt6` = Transformation tier 6
- No `s640x640` or other size restrictions

This gives you the highest quality image Instagram has available.

## 🐛 Troubleshooting

### "403 Forbidden" Error

**Cause**: User not logged in to Instagram

**Solution**:
1. Open instagram.com in same browser
2. Login to account
3. Try download again

### "URL Expired" Error

**Cause**: URL parameters expired

**Solution**:
1. Fetch fresh URL from API
2. Don't cache URLs for long periods

### "CORS Error"

**Cause**: Browser blocking cross-origin request

**Solution**:
1. Open URL in new tab: `window.open(imageURL, '_blank')`
2. Or use proper CORS headers

## 💡 Alternative: Screenshot Approach

If client-side download is not acceptable, consider:

1. **Browser Extension**: Build extension that can access Instagram directly
2. **Puppeteer/Playwright**: Use headless browser on server (heavy, but works)
3. **User Upload**: Let users screenshot and upload manually

## 📞 Support

For issues or questions:
- Check user is logged in to Instagram
- Verify URL is recent (fetched within last few minutes)
- Test in incognito mode (won't work - proves session dependency)

---

**Last Updated:** 2025-11-29  
**API Version:** 2.0
