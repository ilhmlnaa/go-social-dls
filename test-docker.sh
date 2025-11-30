#!/bin/bash
# Quick Docker test script

echo "🐳 Testing Docker Setup..."
echo ""

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker not found. Please install Docker first."
    exit 1
fi

echo "✅ Docker found: $(docker --version)"
echo ""

# Test docker-compose availability
if command -v docker-compose &> /dev/null; then
    echo "✅ Docker Compose found: $(docker-compose --version)"
elif docker compose version &> /dev/null; then
    echo "✅ Docker Compose (plugin) found"
else
    echo "⚠️  Docker Compose not found (optional)"
fi

echo ""
echo "📦 Build Options:"
echo "  1. Test build (without pushing): docker build -t social-downloader:test ."
echo "  2. Full build: docker build -t social-downloader:latest ."
echo "  3. Compose up: docker-compose up -d"
echo ""
echo "📊 Estimated build time: 5-10 minutes (first time)"
echo "📊 Estimated image size: ~420MB"
echo ""

read -p "Do you want to test build now? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🚀 Starting test build..."
    docker build -t social-downloader:test .
    
    if [ $? -eq 0 ]; then
        echo ""
        echo "✅ Build successful!"
        echo ""
        echo "📊 Image info:"
        docker images social-downloader:test
        echo ""
        echo "🔍 To inspect:"
        echo "  docker run --rm social-downloader:test google-chrome --version"
    else
        echo "❌ Build failed. Check errors above."
        exit 1
    fi
fi
