#!/bin/bash
# VSM Protocol - Cross-Compilation Build Script
#
# Builds the shared library for all supported platforms.
# CGO + libpcap means we need native toolchains per platform.
#
# What you can build locally on macOS:
#   - darwin/arm64  (Apple Silicon - your machine)
#   - darwin/amd64  (Intel Mac)
#
# Linux and Windows require either:
#   - GitHub Actions CI (see .github/workflows/build.yml)
#   - Docker with cross-compilation toolchains
#
# Usage:
#   ./build.sh              # Build for current platform
#   ./build.sh all          # Build all macOS targets
#   ./build.sh ci-linux     # (Run inside Linux CI only)
#   ./build.sh ci-windows   # (Run inside Windows CI only)

set -e

GO="${GO:-go}"
SRC="./bindings/c_api.go"
DIST="./dist"

build() {
    local os=$1 arch=$2 ext=$3
    local outdir="${DIST}/${os}-${arch}"
    mkdir -p "$outdir"
    
    echo " [BUILD] ${os}/${arch} → ${outdir}/vsmprotocol${ext}"
    
    CGO_ENABLED=1 GOOS=$os GOARCH=$arch \
        $GO build -o "${outdir}/vsmprotocol${ext}" -buildmode=c-shared $SRC
    
    # The .h file is the same for all platforms, copy once to dist root
    if [ -f "vsmprotocol.h" ]; then
        cp vsmprotocol.h "${DIST}/vsmprotocol.h"
    fi
    
    echo " [BUILD] ✓ Done"
}

case "${1:-current}" in
    current)
        # Detect current platform
        OS=$(uname -s | tr '[:upper:]' '[:lower:]')
        ARCH=$(uname -m)
        [ "$ARCH" = "x86_64" ] && ARCH="amd64"
        [ "$ARCH" = "aarch64" ] && ARCH="arm64"
        
        if [ "$OS" = "darwin" ]; then
            build darwin $ARCH .dylib
        elif [ "$OS" = "linux" ]; then
            build linux $ARCH .so
        fi
        ;;
    all)
        # All macOS targets (can be done locally)
        build darwin arm64 .dylib
        build darwin amd64 .dylib
        echo ""
        echo " [BUILD] All macOS targets complete."
        echo " [BUILD] For Linux/Windows, use GitHub Actions CI."
        ;;
    ci-linux)
        # Called from GitHub Actions on a Linux runner
        build linux amd64 .so
        ;;
    ci-windows)
        # Called from GitHub Actions on a Windows runner
        build windows amd64 .dll
        ;;
    *)
        echo "Usage: ./build.sh [current|all|ci-linux|ci-windows]"
        exit 1
        ;;
esac
