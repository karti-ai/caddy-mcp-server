#!/bin/bash
set -e

cd "$(dirname "$0")"

# Initialize go module if needed
if [ ! -f go.sum ]; then
    echo "📦 Initializing Go module..."
    go mod tidy
fi

# Build
echo "🔨 Building caddy-mcp..."
go build -o caddy-mcp .

echo "✅ Build complete: ./caddy-mcp"
echo ""
echo "Usage:"
echo "  ./caddy-mcp"
