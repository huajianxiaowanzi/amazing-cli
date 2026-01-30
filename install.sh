#!/bin/sh
# Installation script for amazing-cli
# This script downloads and installs the latest version of amazing-cli

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    linux*)
        OS="Linux"
        ;;
    darwin*)
        OS="Darwin"
        ;;
    msys*|mingw*|cygwin*)
        OS="Windows"
        ;;
    *)
        echo "${RED}Unsupported operating system: $OS${NC}"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64|amd64)
        ARCH="x86_64"
        ;;
    i386|i686)
        ARCH="i386"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

# GitHub repository
REPO="huajianxiaowanzi/amazing-cli"
BINARY_NAME="amazing"

echo "${GREEN}🚀 Installing amazing-cli...${NC}"
echo "Detected OS: $OS"
echo "Detected Architecture: $ARCH"

# Get latest release
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo "${RED}Failed to get latest release information${NC}"
    echo ""
    echo "${YELLOW}It appears this repository doesn't have any releases yet.${NC}"
    echo ""
    echo "${GREEN}Alternative installation methods:${NC}"
    echo "1. Install from source (requires Go):"
    echo "   ${YELLOW}go install github.com/$REPO@latest${NC}"
    echo ""
    echo "2. Build from source:"
    echo "   ${YELLOW}git clone https://github.com/$REPO.git${NC}"
    echo "   ${YELLOW}cd ${REPO##*/}${NC}"
    echo "   ${YELLOW}go build -o $BINARY_NAME${NC}"
    echo ""
    echo "For more information, visit: ${GREEN}https://github.com/$REPO${NC}"
    exit 1
fi

echo "Latest version: $LATEST_RELEASE"

# Construct download URL
ARCHIVE_NAME="amazing-cli_${OS}_${ARCH}"
if [ "$OS" = "Windows" ]; then
    ARCHIVE_NAME="${ARCHIVE_NAME}.zip"
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/$ARCHIVE_NAME"
else
    ARCHIVE_NAME="${ARCHIVE_NAME}.tar.gz"
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/$ARCHIVE_NAME"
fi

echo "Downloading from: $DOWNLOAD_URL"

# Download
TMPDIR=$(mktemp -d)
cd "$TMPDIR"

if ! curl -fsSL -o "$ARCHIVE_NAME" "$DOWNLOAD_URL"; then
    echo "${RED}Failed to download binary${NC}"
    echo "Please check your internet connection and that the release exists"
    exit 1
fi

# Download and verify checksum
echo "Downloading checksums..."
CHECKSUM_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/checksums.txt"
if ! curl -fsSL -o "checksums.txt" "$CHECKSUM_URL"; then
    echo "${YELLOW}Warning: Could not download checksums file${NC}"
else
    echo "Verifying checksum..."
    if command -v sha256sum >/dev/null 2>&1; then
        if ! grep "$ARCHIVE_NAME" checksums.txt | sha256sum -c --status; then
            echo "${RED}Checksum verification failed!${NC}"
            echo "The downloaded file may be corrupted or tampered with."
            exit 1
        fi
        echo "${GREEN}Checksum verified successfully${NC}"
    elif command -v shasum >/dev/null 2>&1; then
        if ! grep "$ARCHIVE_NAME" checksums.txt | shasum -a 256 -c --status; then
            echo "${RED}Checksum verification failed!${NC}"
            echo "The downloaded file may be corrupted or tampered with."
            exit 1
        fi
        echo "${GREEN}Checksum verified successfully${NC}"
    else
        echo "${YELLOW}Warning: sha256sum/shasum not found, skipping checksum verification${NC}"
    fi
fi

# Extract
echo "Extracting..."
if [ "$OS" = "Windows" ]; then
    unzip -q "$ARCHIVE_NAME"
    BINARY_NAME="${BINARY_NAME}.exe"
else
    tar -xzf "$ARCHIVE_NAME"
fi

# Install
if [ "$OS" = "Windows" ]; then
    INSTALL_DIR="$HOME/bin"
else
    INSTALL_DIR="/usr/local/bin"
fi

echo "Installing to $INSTALL_DIR..."

if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR"
fi

if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    echo "${YELLOW}Permission required for installation...${NC}"
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

# Cleanup
cd - > /dev/null 2>&1
rm -rf "$TMPDIR"

echo "${GREEN}✅ Installation complete!${NC}"
echo ""
echo "Run ${GREEN}amazing${NC} to start using the CLI"
echo ""

# Check if binary is in PATH
if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
    echo "${YELLOW}⚠️  Warning: $INSTALL_DIR is not in your PATH${NC}"
    echo "Add it to your PATH by adding this line to your shell profile:"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
fi
