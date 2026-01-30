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
    if command -v grep >/dev/null 2>&1; then
        CHECKSUM_LINE=$(grep "$ARCHIVE_NAME" checksums.txt || true)
    fi
    if [ -z "$CHECKSUM_LINE" ]; then
        echo "${YELLOW}Warning: checksum entry for $ARCHIVE_NAME not found; skipping verification${NC}"
    else
        EXPECTED_SUM=$(printf "%s" "$CHECKSUM_LINE" | awk '{print $1}')
        if command -v sha256sum >/dev/null 2>&1; then
            ACTUAL_SUM=$(sha256sum "$ARCHIVE_NAME" | awk '{print $1}')
        elif command -v shasum >/dev/null 2>&1; then
            ACTUAL_SUM=$(shasum -a 256 "$ARCHIVE_NAME" | awk '{print $1}')
        else
            ACTUAL_SUM=""
        fi
        if [ -z "$ACTUAL_SUM" ]; then
            echo "${YELLOW}Warning: sha256sum/shasum not found, skipping checksum verification${NC}"
        elif [ "$EXPECTED_SUM" != "$ACTUAL_SUM" ]; then
            echo "${RED}Checksum verification failed!${NC}"
            echo "The downloaded file may be corrupted or tampered with."
            exit 1
        else
            echo "${GREEN}Checksum verified successfully${NC}"
        fi
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
INSTALL_DIR="$HOME/bin"

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

# Helpers for keeping $INSTALL_DIR in PATH
path_contains_install_dir() {
    case ":$PATH:" in
        *:"$INSTALL_DIR":*)
            return 0
            ;;
    esac
    return 1
}

is_script_sourced() {
    if [ -n "${BASH_SOURCE:-}" ] && [ "${BASH_SOURCE[0]}" != "$0" ]; then
        return 0
    fi
    if [ -n "${ZSH_EVAL_CONTEXT:-}" ] && case $ZSH_EVAL_CONTEXT in *:file) true;; *) false;; esac; then
        return 0
    fi
    return 1
}

is_interactive_shell() {
    if [ -t 0 ] && [ -t 1 ]; then
        return 0
    fi
    case "$-" in
        *i*) return 0 ;;
    esac
    return 1
}

detect_shell_profile() {
    local shell_name profile_file
    shell_name=$(basename "${SHELL:-}" 2>/dev/null)
    case "$shell_name" in
        zsh)
            if [ -f "$HOME/.zprofile" ]; then
                profile_file="$HOME/.zprofile"
            else
                profile_file="$HOME/.zshrc"
            fi
            ;;
        bash)
            if [ -f "$HOME/.bash_profile" ]; then
                profile_file="$HOME/.bash_profile"
            elif [ -f "$HOME/.bashrc" ]; then
                profile_file="$HOME/.bashrc"
            else
                profile_file="$HOME/.profile"
            fi
            ;;
        *)
            profile_file="$HOME/.profile"
            ;;
    esac
    printf "%s" "$profile_file"
}

append_install_dir_to_profile() {
    local profile_file
    profile_file=$(detect_shell_profile)
    [ -z "$profile_file" ] && return 1
    if [ ! -f "$profile_file" ]; then
        mkdir -p "$(dirname "$profile_file")" 2>/dev/null || true
        touch "$profile_file"
    fi
    if grep -Fq "$INSTALL_DIR" "$profile_file" >/dev/null 2>&1; then
        printf "%s" "$profile_file"
        return 0
    fi
    {
        printf "\n# added by amazing-cli installer\n"
        printf 'export PATH="%s:$PATH"\n' "$INSTALL_DIR"
    } >> "$profile_file"
    printf "%s" "$profile_file"
}

if ! path_contains_install_dir; then
    profile_file=$(append_install_dir_to_profile)
    if [ -n "$profile_file" ]; then
        echo "${GREEN}Added $INSTALL_DIR to PATH in $profile_file${NC}"
        if is_script_sourced; then
            . "$profile_file"
            echo "${GREEN}PATH updated in current shell.${NC}"
        elif is_interactive_shell && [ -n "${SHELL:-}" ] && [ -z "${CI:-}" ]; then
            echo "${GREEN}Restarting your shell to apply PATH changes...${NC}"
            exec "$SHELL" -l
        else
            echo "Restart your shell or run ${GREEN}source \"$profile_file\"${NC} to apply the change."
        fi
    else
        echo "${YELLOW}⚠️  Warning: $INSTALL_DIR is not in your PATH${NC}"
        echo "Add it to your PATH by adding this line to your shell profile:"
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
else
    echo "${GREEN}$INSTALL_DIR already in PATH${NC}"
fi

export PATH="$INSTALL_DIR:$PATH"
