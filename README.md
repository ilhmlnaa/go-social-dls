# 🚀 Social Media Downloader API v2.0

**High-performance API for downloading media from Twitter, Facebook, Instagram, Pinterest, and more!**

Built with **Go** and **Fiber framework** - Fast, reliable, and production-ready.

## ✨ Features

- ✅ **Pure Go** - No Python dependencies, single binary deployment
- ✅ **Fast** - Powered by Fiber framework (Express-like API)
- ✅ **Twitter** - Download images from tweets (pure Go implementation)
- ✅ **Facebook** - Download photos with URL normalization
- ✅ **Instagram** - Full resolution images via Chromedp headless browser
- ✅ **Pinterest** - Get high-quality pin images
- ✅ **Generic** - Download any image from URL
- ✅ **Production Ready** - Proper error handling, logging, CORS support

## 🏗️ Architecture

### Clean Project Structure

```
go-social-dls/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP handlers (Fiber)
│   ├── services/        # Business logic
│   ├── models/          # Data models
│   └── utils/           # Utilities
├── cookies/             # Platform cookies
│   ├── twitter.json
│   └── facebook.json
├── static/              # Static files
└── .env                 # Environment variables
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.20+**
- **Google Chrome** (for Instagram full resolution)
  ```bash
  wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
  sudo apt install ./google-chrome-stable_current_amd64.deb -y
  ```
- **Browser** with Cookie-Editor extension (for Twitter/Facebook/Instagram cookies)

### Installation

```bash
# Clone repository
git clone https://github.com/ilhmlnaa/go-social-dls.git
cd go-social-dls

# Build
go build -o server ./cmd/server/

# Run
./server
```

Server will start on `http://localhost:3005`

## 🍪 Setup Cookies (Required for Twitter/Facebook/Instagram)

### Twitter Setup

1. Install [Cookie-Editor](https://chrome.google.com/webstore/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm) extension
2. Login to Twitter/X in your browser
3. Click Cookie-Editor → Export → JSON
4. Save as `cookies/twitter.json`

### Facebook Setup (Optional)

1. Login to Facebook in your browser
2. Click Cookie-Editor → Export → JSON
3. Save as `cookies/facebook.json`

### Instagram Setup (Required for full resolution)

1. Login to Instagram in your browser
2. Click Cookie-Editor → Export → JSON
3. Save as `cookies/instagram.json`

**Important:** Add `cookies/*.json` to `.gitignore` to keep your cookies secure!

## 📡 API Endpoints

### Root & Health

```bash
# API Information
GET /

# Health Check
GET /health
```

### Download Endpoints

All endpoints return JSON:

```json
{
  "success": true,
  "message": "Successfully fetched 2 photo(s)",
  "data": ["https://image1.jpg", "https://image2.jpg"]
}
```

#### Twitter

```bash
GET /api/v1/twitter?url={tweet_url}

# Example
curl "http://localhost:3005/api/v1/twitter?url=https://twitter.com/user/status/1234567890"
```

#### Facebook

```bash
GET /api/v1/facebook?url={facebook_url}

# Example
curl "http://localhost:3005/api/v1/facebook?url=https://www.facebook.com/photo/?fbid=123456"
```

#### Instagram

```bash
GET /api/v1/instagram?url={instagram_url}

# Example
curl "http://localhost:3005/api/v1/instagram?url=https://www.instagram.com/p/ABC123/"
```

**✨ Full Resolution via Chromedp Headless Browser:**
- Uses **Google Chrome** headless browser to render Instagram page
- Extracts full resolution URLs from JavaScript-rendered content
- Returns images with `ig_cache_key` parameter (full resolution indicator)
- Supports carousel posts (multiple images)
- Response time: ~3-5 seconds (slower due to browser rendering)
- See [INSTAGRAM_HEADLESS_BROWSER_APPROACH.md](./INSTAGRAM_HEADLESS_BROWSER_APPROACH.md) for technical details

**Response Example:**
```json
{
  "success": true,
  "message": "Successfully fetched 6 full resolution image(s)",
  "data": [
    "https://scontent.cdninstagram.com/v/...&ig_cache_key=...",
    "..."
  ]
}
```

#### Pinterest

```bash
GET /api/v1/pinterest?url={pinterest_url}

# Example
curl "http://localhost:3005/api/v1/pinterest?url=https://www.pinterest.com/pin/123456/"
```

#### Generic (Any Image)

```bash
GET /api/v1/generic?url={image_url}

# Example
curl "http://localhost:3005/api/v1/generic?url=https://example.com/image.jpg"
```

## 🔧 Configuration

### Environment Variables

Create a `.env` file:

```env
PORT=3005
ENV=development
COOKIES_DIR=cookies
```

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3005` | Server port |
| `ENV` | `development` | Environment (development/production) |
| `COOKIES_DIR` | `cookies` | Directory for cookie files |

## 🐳 Docker Deployment

### Build Image

```bash
docker build -t social-media-downloader .
```

### Run with Docker

```bash
docker run -d -p 3005:3005 \
  -v $(pwd)/cookies:/app/cookies \
  --name downloader \
  social-media-downloader
```

### Docker Compose

```yaml
version: '3.8'
services:
  api:
    image: social-media-downloader
    ports:
      - "3005:3005"
    volumes:
      - ./cookies:/app/cookies
    environment:
      - PORT=3005
      - ENV=production
```

## 🎯 Key Improvements (v2.0)

### From v1.0 to v2.0:

| Aspect | v1.0 | v2.0 |
|--------|------|------|
| **Framework** | net/http | Fiber (faster) |
| **Twitter** | Python subprocess | Pure Go |
| **Facebook** | Not available | Pure Go |
| **Structure** | Flat | Clean architecture |
| **Dependencies** | Python + Go | Go only |
| **Performance** | Good | Excellent |
| **Deployment** | 2 runtimes | Single binary |
| **Type Safety** | Partial | Full |

## 📊 Performance

- ⚡ **50%+ faster** than v1.0 (no subprocess overhead)
- 🔥 **Single binary** - no Python runtime needed
- 💾 **Lower memory** usage (single process)
- 🚀 **Better concurrency** with Go goroutines

## 🛠️ Development

### Project Structure Explained

```
cmd/server/main.go          # Entry point, Fiber app setup, routing
internal/
  ├── config/               # Config loader from env
  ├── handlers/             # HTTP handlers (convert requests to service calls)
  ├── services/             # Business logic (Twitter, Facebook scraping)
  ├── models/               # Data structures (responses, requests)
  └── utils/                # Helpers (cookie loader, etc)
```

### Adding New Platform

1. Create service in `internal/services/{platform}.go`
2. Create handler in `internal/handlers/{platform}.go`
3. Add route in `cmd/server/main.go`
4. (Optional) Add cookie support in `cookies/{platform}.json`

### Build & Test

```bash
# Build
go build -o server ./cmd/server/

# Run
./server

# Test endpoint
curl "http://localhost:3005/health"
```

## 🔒 Security

### Cookies Safety

- ✅ Cookies stored locally in `cookies/` directory
- ✅ Added to `.gitignore` by default
- ✅ Never commit cookies to git
- ✅ Use environment variables in production

### Best Practices

1. **Don't share cookies** - They provide full account access
2. **Use dedicated accounts** - Don't use personal accounts for scraping
3. **Rotate cookies** - Update periodically for security
4. **Monitor usage** - Check for suspicious activity

## ⚠️ Troubleshooting

### Twitter: "Failed to initialize service"

**Solution:** Make sure `cookies/twitter.json` exists and is valid.

```bash
# Check file exists
ls -la cookies/twitter.json

# Validate JSON
cat cookies/twitter.json | python3 -m json.tool
```

### Facebook: "No photos found"

**Solution:** Facebook requires valid cookies. Some URLs may not be accessible.

### Port Already in Use

```bash
# Find process on port 3005
lsof -i :3005

# Kill process
kill -9 <PID>
```

## 📚 Documentation

- [Architecture Overview](./docs/ARCHITECTURE.md) _(Coming soon)_
- [API Reference](./docs/API.md) _(Coming soon)_
- [Contributing Guide](./CONTRIBUTING.md) _(Coming soon)_

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Fiber](https://gofiber.io/) - Fast HTTP framework
- Inspired by various social media downloaders
- Thanks to all contributors!

---

**Made with ❤️ and Go**

**Star ⭐ this repo if you find it useful!**
