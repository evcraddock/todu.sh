#!/bin/sh
set -e

# Todu CLI Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/evcraddock/todu.sh/main/install.sh | sh
#
# Environment variables:
#   TODU_INSTALL_DIR     Custom install directory (default: ~/.local/bin)

REPO="evcraddock/todu.sh"
BINARY_NAME="todu"
INSTALL_DIR="${TODU_INSTALL_DIR:-$HOME/.local/bin}"

main() {
    echo "Installing todu..."
    detect_platform
    get_latest_version
    download_and_install
    verify_installation
}

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *)
            echo "Error: Unsupported architecture: $ARCH"
            echo "Supported: x86_64/amd64, aarch64/arm64"
            exit 1
            ;;
    esac

    case "$OS" in
        linux) OS="linux" ;;
        darwin) OS="darwin" ;;
        mingw*|msys*|cygwin*)
            echo "Error: Windows is not supported by this installer."
            echo "Download manually from: https://github.com/${REPO}/releases"
            exit 1
            ;;
        *)
            echo "Error: Unsupported OS: $OS"
            exit 1
            ;;
    esac

    PLATFORM="${OS}_${ARCH}"
    echo "Detected platform: $PLATFORM"
}

get_latest_version() {
    echo "Fetching latest version..."
    VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" |
        grep '"tag_name"' |
        sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$VERSION" ]; then
        echo "Error: Could not determine latest version"
        echo "Check: https://github.com/${REPO}/releases"
        exit 1
    fi
    echo "Latest version: $VERSION"
}

download_and_install() {
    # Strip 'v' prefix from version for filename (e.g., v1.1.0 -> 1.1.0)
    VERSION_NUM="${VERSION#v}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}_${VERSION_NUM}_${PLATFORM}.tar.gz"
    echo "Downloading: $DOWNLOAD_URL"

    TEMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TEMP_DIR"' EXIT

    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_DIR/todu.tar.gz"; then
        echo "Error: Download failed"
        echo "Check: https://github.com/${REPO}/releases"
        exit 1
    fi

    tar -xzf "$TEMP_DIR/todu.tar.gz" -C "$TEMP_DIR"

    mkdir -p "$INSTALL_DIR"
    mv "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
    echo "Installed to: $INSTALL_DIR/$BINARY_NAME"
}

verify_installation() {
    echo ""
    if command -v todu >/dev/null 2>&1; then
        echo "Success: $(todu version 2>/dev/null || todu --version 2>/dev/null || echo 'installed')"
    else
        echo "Installation complete."
        # Check if install dir is in PATH
        case ":$PATH:" in
            *":$INSTALL_DIR:"*) ;;
            *)
                echo ""
                echo "$INSTALL_DIR is not in your PATH."
                echo "Add it by running:"
                echo ""
                echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
                echo ""
                echo "Or add this line to your ~/.bashrc or ~/.zshrc"
                ;;
        esac
    fi
}

main
