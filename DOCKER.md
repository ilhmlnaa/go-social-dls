# 🐳 Docker Deployment Guide

Complete guide for deploying Social Media Downloader API with Docker.

## 📋 Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- At least 2GB RAM (for Chrome headless)
- Cookie files (twitter.json, facebook.json, instagram.json)

---

## 🚀 Quick Start

### 1. Prepare Cookie Files

```bash
# Create cookies directory if not exists
mkdir -p cookies

# Copy your cookie files
cp /path/to/twitter.json cookies/
cp /path/to/facebook.json cookies/
cp /path/to/instagram.json cookies/

# Set proper permissions
chmod 600 cookies/*.json
```

### 2. Build and Run

```bash
# Build image
docker build -t social-downloader:latest .

# Run container
docker run -d \
  --name social-downloader \
  -p 3005:3005 \
  -v $(pwd)/cookies:/app/cookies:ro \
  --security-opt seccomp:unconfined \
  social-downloader:latest
```

### 3. Or Use Docker Compose (Recommended)

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

---

## 🏗️ Dockerfile Details

### Multi-Stage Build

**Stage 1: Builder (golang:1.21-bookworm)**
- CGO enabled for chromedp support
- Compiles Go binary

**Stage 2: Runner (debian:bookworm-slim)**
- Installs Google Chrome
- Installs Chrome dependencies
- Minimal Debian base (~200MB final image)

### Key Features

✅ **Google Chrome Included**
```dockerfile
RUN wget -q https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
RUN apt-get install -y ./google-chrome-stable_current_amd64.deb
```

✅ **Non-Root User** (Security)
```dockerfile
USER appuser
```

✅ **Health Check**
```dockerfile
HEALTHCHECK CMD wget --spider http://localhost:3005/health
```

---

## 🔧 Configuration

### Environment Variables

Create `.env` file:

```bash
# Server
PORT=3005
ENV=production
COOKIES_DIR=/app/cookies

# CORS
ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com

# Chrome
CHROME_BIN=/usr/bin/google-chrome
CHROME_PATH=/usr/bin/google-chrome
```

### Docker Compose Configuration

Key settings in `docker-compose.yml`:

```yaml
# Required for Chrome sandbox
security_opt:
  - seccomp:unconfined

# Resource limits
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 2G
```

---

## 📊 Resource Requirements

### Minimum Requirements
- **CPU**: 1 core
- **RAM**: 512MB (without Instagram)
- **RAM**: 1GB+ (with Instagram chromedp)
- **Disk**: 500MB (image size)

### Recommended Production
- **CPU**: 2+ cores
- **RAM**: 2GB+
- **Disk**: 1GB+

### Why More RAM for Instagram?
- Chrome headless requires ~200-500MB per instance
- Instagram requests take 3-5 seconds (browser rendering)
- Consider queue system for high traffic

---

## 🐛 Troubleshooting

### Issue: Chrome crashes or hangs

**Solution 1: Increase shared memory**
```yaml
# docker-compose.yml
shm_size: '1gb'
```

**Solution 2: Add more security options**
```yaml
security_opt:
  - seccomp:unconfined
cap_add:
  - SYS_ADMIN
```

### Issue: "Failed to launch browser"

**Check Chrome installation:**
```bash
docker exec -it social-downloader google-chrome --version
```

**Check Chrome dependencies:**
```bash
docker exec -it social-downloader ldd /usr/bin/google-chrome
```

### Issue: Container exits immediately

**Check logs:**
```bash
docker logs social-downloader
```

**Common causes:**
- Missing cookie files
- Port already in use
- Insufficient memory

### Issue: Permission denied on cookies

**Fix permissions:**
```bash
chmod 600 cookies/*.json
chown 1000:1000 cookies/*.json  # Match container user
```

---

## 🔒 Security Best Practices

### 1. Use Non-Root User ✅
Already configured in Dockerfile:
```dockerfile
USER appuser
```

### 2. Mount Cookies as Read-Only
```bash
-v $(pwd)/cookies:/app/cookies:ro
```

### 3. Use Secrets for Production

**Docker Swarm:**
```yaml
secrets:
  twitter_cookies:
    file: ./cookies/twitter.json
```

**Kubernetes:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: social-downloader-cookies
type: Opaque
data:
  twitter.json: <base64-encoded>
```

### 4. Network Isolation
```yaml
networks:
  social-downloader-network:
    internal: true
  web:
    external: true
```

### 5. Limit Resources
```yaml
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 2G
```

---

## 🚦 Production Deployment

### Option 1: Docker Compose (Simple)

```bash
# Production compose file
version: '3.8'

services:
  app:
    image: social-downloader:latest
    restart: always
    ports:
      - "3005:3005"
    environment:
      - ENV=production
    volumes:
      - ./cookies:/app/cookies:ro
    deploy:
      replicas: 2  # Multiple instances
      resources:
        limits:
          memory: 2G
```

### Option 2: Docker Swarm (Orchestration)

```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.yml social-downloader

# Scale service
docker service scale social-downloader_app=3
```

### Option 3: Kubernetes (Advanced)

See `k8s/` directory for:
- Deployment manifest
- Service manifest
- ConfigMap for environment
- Secret for cookies
- Ingress for routing

---

## 📈 Monitoring & Logging

### View Logs

```bash
# Docker
docker logs -f social-downloader

# Docker Compose
docker-compose logs -f

# Last 100 lines
docker logs --tail 100 social-downloader
```

### Health Check

```bash
# Check container health
docker inspect --format='{{.State.Health.Status}}' social-downloader

# Test endpoint
curl http://localhost:3005/health
```

### Resource Usage

```bash
# Real-time stats
docker stats social-downloader

# Detailed info
docker inspect social-downloader
```

---

## 🔄 Updates & Maintenance

### Update Application

```bash
# Pull latest changes
git pull

# Rebuild image
docker-compose build

# Restart with new image
docker-compose up -d
```

### Update Cookies

```bash
# Copy new cookies
cp new-cookies.json cookies/twitter.json

# Restart container (reload cookies)
docker-compose restart
```

### Clean Up

```bash
# Remove stopped containers
docker container prune

# Remove unused images
docker image prune

# Remove all unused resources
docker system prune -a
```

---

## 📝 Example Commands

### Build & Tag

```bash
# Build with tag
docker build -t social-downloader:v1.0.0 .

# Tag for registry
docker tag social-downloader:v1.0.0 registry.example.com/social-downloader:v1.0.0

# Push to registry
docker push registry.example.com/social-downloader:v1.0.0
```

### Run with All Options

```bash
docker run -d \
  --name social-downloader \
  --restart unless-stopped \
  -p 3005:3005 \
  -e PORT=3005 \
  -e ENV=production \
  -e ALLOWED_ORIGINS="https://yourdomain.com" \
  -v $(pwd)/cookies:/app/cookies:ro \
  -v $(pwd)/logs:/app/logs \
  --security-opt seccomp:unconfined \
  --memory="2g" \
  --cpus="2" \
  --health-cmd="wget --spider http://localhost:3005/health || exit 1" \
  --health-interval=30s \
  --health-timeout=10s \
  --health-retries=3 \
  social-downloader:latest
```

### Debug Container

```bash
# Execute bash in container
docker exec -it social-downloader bash

# Check Chrome
google-chrome --version

# Check Go binary
./server --version

# Test API
curl http://localhost:3005/health
```

---

## 🌍 Reverse Proxy (Nginx/Traefik)

### Nginx Configuration

```nginx
upstream social-downloader {
    server 127.0.0.1:3005;
}

server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://social-downloader;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Increase timeout for Instagram requests
        proxy_read_timeout 30s;
        proxy_connect_timeout 10s;
    }
}
```

### Traefik Labels

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.social-downloader.rule=Host(`api.yourdomain.com`)"
  - "traefik.http.services.social-downloader.loadbalancer.server.port=3005"
```

---

## ✅ Checklist

Before deploying to production:

- [ ] Cookie files prepared and secured
- [ ] `.env` file configured
- [ ] CORS origins set correctly
- [ ] Resource limits appropriate
- [ ] Health checks working
- [ ] Logging configured
- [ ] Backup strategy in place
- [ ] Monitoring setup
- [ ] SSL/TLS configured (reverse proxy)
- [ ] Firewall rules set

---

## 📚 Additional Resources

- [Dockerfile Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Chromedp in Docker](https://github.com/chromedp/docker-headless-shell)
- [Go Docker Images](https://hub.docker.com/_/golang)

---

**Need help?** Check the main README.md or open an issue on GitHub.
