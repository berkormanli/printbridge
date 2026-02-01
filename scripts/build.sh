#!/bin/bash
# Build script for PrintBridge
# Run on the target platform for native builds, or use CGO_ENABLED=0 for cross-compile

set -e

VERSION="${1:-1.0.0}"
OUTPUT_DIR="build"

echo "Building PrintBridge v${VERSION}..."

mkdir -p "${OUTPUT_DIR}/windows"

# Windows AMD64 - CGO disabled for cross-compile
echo "Building Windows AMD64..."
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "${OUTPUT_DIR}/windows/printbridge.exe" .
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "${OUTPUT_DIR}/windows/printbridge-tray.exe" ./cmd/tray/

# Copy supporting files
cp config.json "${OUTPUT_DIR}/windows/" 2>/dev/null || true
cp README.md "${OUTPUT_DIR}/windows/" 2>/dev/null || true
[ -f LICENSE ] && cp LICENSE "${OUTPUT_DIR}/windows/"

echo ""
echo "✅ Windows build complete!"
ls -la "${OUTPUT_DIR}/windows/"
echo ""
echo "Note: USB adapter disabled in cross-compiled builds."
echo "Use 'network' adapter in config.json for Windows."
