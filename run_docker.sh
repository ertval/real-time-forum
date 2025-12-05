#!/usr/bin/env bash
set -e

echo "🐳 Building Docker image..."
make docker-build

echo "🚀 Starting Docker container..."
make docker-up &

# Delay to allow container to fully start
sleep 2

# Open browser (optional)
URL="http://localhost:8080"
echo "🌍 Opening browser at $URL"

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    xdg-open "$URL" >/dev/null 2>&1
elif [[ "$OSTYPE" == "darwin"* ]]; then
    open "$URL"
elif [[ "$OSTYPE" == "msys"* || "$OSTYPE" == "cygwin"* ]]; then
    start "$URL"
fi

echo "📜 Streaming container logs (Ctrl + C to stop)..."
docker compose logs -f
