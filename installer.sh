#!/bin/bash

# Function to get latest release tag from GitHub API
get_latest_tag() {
    local repo=$1
    curl -s "https://api.github.com/repos/$repo/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
}

# Function to detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        *)
            echo "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    case "$OS" in
        linux|darwin)
            ;;
        *)
            echo "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
}

# Main installation function
install_dockermon() {
    # Replace with your GitHub username/repo
    REPO="malletgaetan/dockermon"
    INSTALL_DIR="/usr/local/bin"

    echo "Detecting platform..."
    detect_platform

    echo "Getting latest release..."
    TAG=$(get_latest_tag "$REPO")
    if [ -z "$TAG" ]; then
        echo "Failed to get latest release tag"
        exit 1
    fi

    BINARY_NAME="dockermon-$OS-$ARCH"
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$TAG/$BINARY_NAME"

    echo "Downloading $BINARY_NAME..."
    if ! curl -L -o "$INSTALL_DIR/dockermon" "$DOWNLOAD_URL"; then
        echo "Failed to download binary"
        exit 1
    fi

    echo "Setting executable permissions..."
    chmod +x "$INSTALL_DIR/dockermon"

    echo "Installation complete! dockermon has been installed to $INSTALL_DIR/dockermon"
    echo "You can now run 'dockermon' from anywhere"
}

if [ "$(id -u)" -ne 0 ]; then
    echo "This script must be run as root (sudo)"
    exit 1
fi

install_dockermon