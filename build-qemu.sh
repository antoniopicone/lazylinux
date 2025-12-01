#!/bin/bash

# QEMU Build Script for LazyLinux
# Compiles QEMU with OpenGL (Virgil 3D), Cocoa, HVF, and vmnet support on macOS.

set -e

# Configuration
QEMU_VERSION="10.1.2"  # Latest stable as of late 2024/early 2025
INSTALL_DIR="$HOME/.vm/qemu"
BUILD_DIR="/tmp/qemu-build"
SOURCE_DIR="/tmp/qemu-source"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $*"
}

success() {
    echo -e "${GREEN}✔${NC} $*"
}

error() {
    echo -e "${RED}✗${NC} $*" >&2
    exit 1
}

# Check for Homebrew
if ! command -v brew >/dev/null 2>&1; then
    error "Homebrew is required but not found. Please install it first."
fi

# Install dependencies
log "Installing build dependencies..."
brew install ninja meson glib pixman pkg-config libepoxy usbredir spice-protocol wget rust llvm

# Create directories
mkdir -p "$INSTALL_DIR"
mkdir -p "$BUILD_DIR"
mkdir -p "$SOURCE_DIR"

# Download QEMU
log "Downloading QEMU $QEMU_VERSION..."
cd "$SOURCE_DIR"
if [ ! -f "qemu-$QEMU_VERSION.tar.xz" ]; then
    # Check if version exists, fallback to 9.2.0 if 10.x fails
    if ! wget --spider "https://download.qemu.org/qemu-$QEMU_VERSION.tar.xz" 2>/dev/null; then
        log "Version $QEMU_VERSION not found, falling back to 9.2.0..."
        QEMU_VERSION="9.2.0"
    fi
    wget "https://download.qemu.org/qemu-$QEMU_VERSION.tar.xz"
fi

log "Extracting source..."
tar xf "qemu-$QEMU_VERSION.tar.xz"
cd "qemu-$QEMU_VERSION"

# Configure
log "Configuring QEMU build..."

./configure \
    --prefix="$INSTALL_DIR" \
    --target-list=aarch64-softmmu,x86_64-softmmu \
    --enable-cocoa \
    --enable-hvf \
    --enable-vmnet \
    --disable-vnc \
    --disable-sdl \
    --disable-gtk \
    --enable-slirp \
    --enable-tools

# Compile
log "Compiling QEMU (this may take a while)..."
# Get number of cores
CORES=$(sysctl -n hw.ncpu)
make -j"$CORES"

# Install
log "Installing to $INSTALL_DIR..."
make install

# Cleanup
log "Cleaning up..."
rm -rf "$BUILD_DIR"
# Optional: keep source for future reference or remove it
# rm -rf "$SOURCE_DIR"

success "QEMU build complete!"
echo ""
echo "Custom QEMU installed at: $INSTALL_DIR"
echo "Version: $("$INSTALL_DIR/bin/qemu-system-aarch64" --version | head -1)"
echo ""
echo "To use this build, the 'vm' script will automatically detect it."
echo "Or verify OpenGL support with:"
echo "  $INSTALL_DIR/bin/qemu-system-aarch64 -display help"
echo ""
