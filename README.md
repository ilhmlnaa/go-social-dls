# 🌐 Media Downloader (Golang)

**Twitter Downloader** is a simple API media downloader built with **Golang** to fetch videos and images directly from Twitter. The project is designed with flexibility in mind, making it easy to extend support to other platforms like **Instagram** in the future.

## Features

- ✅ Download media (video/image) directly from instagram, facebook, x, pinterest and more!.
- ✅ Simple and clean API endpoint.
- ✅ Lightweight and fast — run as a single binary or with Docker.

## Installation

### Prerequisites

- **Go 1.20+** - For the main server
- **Python 3.8+** - For Twitter scraper (uses twscrape)
- **pip** - Python package manager

### Run Locally

Clone the repository and start the server using Go:

```bash
git clone https://github.com/ilhmlnaa/media-downloader-go.git
cd twitter-downloader

# Install Python dependencies for Twitter scraper
pip3 install -r requirements.txt

# Setup Twitter account for scraping (REQUIRED for Twitter endpoint)
# See TWITTER_TWSCRAPE_SETUP.md for detailed instructions
python3 scripts/setup_twscrape.py add <username> <password> <email> <email_password>

# Run Go server
go run main.go
```

### Build as Binary

```bash
go build -o twitter-dl
./twitter-dl
```

### Run with Docker

Build the Docker image:

```bash
docker build -t go-social-dls .
```
or pull the pre-built image from Docker Hub:

```bash
docker pull ghcr.io/ilhmlnaa/go-social-dls:latest
```


Run the Docker container:

```bash
docker run -d -p 3000:3000 --name go-social-dls \
  -e TWITTER_AUTH_TOKEN=your_twitter_auth_token \
  -e TWITTER_CSRF_TOKEN=your_twitter_csrf_token \
  go-social-dls
```

Once running, your API will be available at:

```
http://localhost:3000
```

## Environment Variables

You can optionally create a `.env` file in the project root:

```env
TWITTER_AUTH_TOKEN=your_twitter_auth_token
TWITTER_CSRF_TOKEN=your_twitter_csrf_token
PORT=3000
```

| Variable             | Description                                                             |
| -------------------- | ----------------------------------------------------------------------- |
| `TWITTER_AUTH_TOKEN` | Required for Twitter endpoint. Your Twitter `auth_token` cookie value.  |
| `TWITTER_CSRF_TOKEN` | Required for Twitter endpoint. Your Twitter `ct0` cookie value (CSRF token). |
| `PORT`               | Optional, defaults to 3000. You can change this to any port you prefer. |

### How to Get Twitter Tokens

To use the Twitter downloader endpoint, you need to obtain your authentication tokens from Twitter:

1. **Open Twitter in your browser** (use normal mode, NOT incognito/private)
2. **Login to your Twitter account**
3. **Open Developer Tools** (Press `F12` or `Right-click` → `Inspect`)
4. **Go to Application/Storage tab**:
   - Chrome/Edge: Click `Application` tab → `Storage` → `Cookies` → `https://twitter.com`
   - Firefox: Click `Storage` tab → `Cookies` → `https://twitter.com`
5. **Find and copy these cookies**:
   - Find cookie named `auth_token` → Copy its **Value** → This is your `TWITTER_AUTH_TOKEN`
   - Find cookie named `ct0` → Copy its **Value** → This is your `TWITTER_CSRF_TOKEN`
6. **Update your `.env` file** with these values

**Important Notes:**
- These tokens are tied to your Twitter session and will expire when you logout
- Never share these tokens publicly as they provide full access to your Twitter account
- If you change your Twitter password, you'll need to get new tokens
- If the API returns authentication errors, your tokens may have expired - get new ones

## API Endpoint

Currently, the project provides a single API endpoint:

**GET** `/twitter?url={twitter_url}`
**GET** `/pinterest?url={pinterest_url}`
**GET** `/twitter?url={twitter_url}`


This endpoint allows you to download media from a Twitter link.

### Example usage with curl:

```bash
curl "http://localhost:3000/twitter?url=https://twitter.com/username/status/1234567890"
```

If you prefer to get the direct media URL, the API will return a JSON response:

```json
{
  "status": "success",
  "urls": ["https://pbs.twimg.com/media/Gr35T-DWMAAuVfL.jpg"]
}
```

## Roadmap

- [x] Instagram media downloader support
- [ ] Instagram media downloader support

## License

This project is licensed under the MIT License — feel free to use, modify, and contribute.

---

**💡 Note:** Twitter authentication tokens can expire or change frequently. Make sure to use a valid token from your current Twitter session.

## 🐦 Twitter Endpoint Setup

**TERBARU:** Twitter endpoint uses browser cookies - paling simple dan reliable!

### Quick Setup (3 Steps)

1. **Install Python dependencies:**
   ```bash
   pip3 install -r requirements.txt
   ```

2. **Export cookies dari browser:**
   - Install [Cookie-Editor](https://chrome.google.com/webstore/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm) extension
   - Login ke Twitter/X di browser
   - Click Cookie-Editor icon → Export → JSON
   - Save as `cookie.json` di root folder project

3. **Test & Run:**
   ```bash
   # Test Python script
   python3 scripts/twitter_scraper_cookies.py 1234567890
   
   # Run server
   go build -o twitter-down
   ./twitter-down
   ```

### Full Documentation

For complete setup guide with screenshots and troubleshooting, see: **[TWITTER_COOKIES_SETUP.md](./TWITTER_COOKIES_SETUP.md)**

### Why Browser Cookies Method?

- ✅ **Paling simple** - cukup export cookies sekali
- ✅ **No Cloudflare blocking** - karena pakai cookies dari browser asli  
- ✅ **No login automation** - tidak perlu worry tentang bot detection
- ✅ **Stable dan reliable** - selama cookies valid, akan terus jalan
- ✅ **No rate limit issues** - karena pakai cookies akun Anda sendiri

---

**Important:** Jangan commit file `cookie.json` ke Git! Add ke `.gitignore`.
