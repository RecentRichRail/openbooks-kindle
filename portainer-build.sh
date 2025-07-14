#!/bin/bash

# Portainer Build Script for OpenBooks
# This script ensures a clean build environment for Docker

set -e

echo "🧹 Cleaning up potential build artifacts..."

# Remove any existing binaries
find . -name "openbooks" -type f -delete 2>/dev/null || true
find . -name "*.exe" -type f -delete 2>/dev/null || true

# Clean npm cache and node_modules if they exist
if [ -d "server/app/node_modules" ]; then
    echo "🗑️  Removing old node_modules..."
    rm -rf server/app/node_modules
fi

# Clean any dist directories
find . -name "dist" -type d -exec rm -rf {} + 2>/dev/null || true

echo "✅ Cleanup complete!"

# Verify critical files exist
echo "🔍 Verifying build requirements..."

if [ ! -f "go.mod" ]; then
    echo "❌ go.mod not found!"
    exit 1
fi

if [ ! -f "server/app/package.json" ]; then
    echo "❌ server/app/package.json not found!"
    exit 1
fi

echo "✅ All required files present!"

# Check for problematic files
echo "🔍 Checking for problematic binary files..."

if [ -f "cmd/mock_server/great-gatsby.epub" ]; then
    echo "⚠️  Found great-gatsby.epub - this file might cause build issues"
    echo "📝 Make sure .dockerignore excludes this file"
fi

if [ -f "cmd/mock_server/SearchBot_results_for__the_great_gatsby.txt.zip" ]; then
    echo "⚠️  Found SearchBot zip file - this file might cause build issues"
    echo "📝 Make sure .dockerignore excludes this file"
fi

echo "✅ Pre-build checks complete!"
echo "🐳 Ready for Docker build!"

# Optional: Show docker ignore status
if [ -f ".dockerignore" ]; then
    echo "📋 .dockerignore file exists with $(wc -l < .dockerignore) lines"
else
    echo "⚠️  No .dockerignore file found - this might cause issues"
fi

echo ""
echo "🚀 You can now run:"
echo "   docker-compose build"
echo "   or"
echo "   docker build -t openbooks-kindle ."
