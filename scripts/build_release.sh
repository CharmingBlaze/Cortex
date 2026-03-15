#!/bin/bash
# Cortex Release Build Script
# Builds Cortex for Windows, Linux, and macOS

set -e

VERSION=${1:-"0.1.0"}
DIST_DIR="dist"
BUILD_DIR="build"

echo "=== Cortex Release Build v$VERSION ==="

# Clean previous builds
rm -rf $DIST_DIR
mkdir -p $DIST_DIR

# Build function
build_cortex() {
    OS=$1
    ARCH=$2
    SUFFIX=$3
    
    echo "Building for $OS/$ARCH..."
    
    OUTPUT="cortex$suffix"
    if [ "$OS" = "windows" ]; then
        OUTPUT="cortex.exe"
    fi
    
    GOOS=$OS GOARCH=$ARCH go build -ldflags="-s -w -X main.Version=$VERSION" -o $BUILD_DIR/$OUTPUT ./cmd/cortex
    
    # Create package directory
    PKG_DIR="$DIST_DIR/cortex-$VERSION-$OS-$ARCH"
    mkdir -p $PKG_DIR/bin
    mkdir -p $PKG_DIR/runtime
    mkdir -p $PKG_DIR/examples
    mkdir -p $PKG_DIR/docs
    
    # Copy binary
    cp $BUILD_DIR/$OUTPUT $PKG_DIR/bin/
    
    # Copy runtime files
    cp -r runtime/*.c $PKG_DIR/runtime/
    cp -r runtime/*.h $PKG_DIR/runtime/
    
    # Copy examples
    cp -r examples/*.cx $PKG_DIR/examples/ 2>/dev/null || true
    cp -r examples/raylib $PKG_DIR/examples/ 2>/dev/null || true
    
    # Copy documentation
    cp README.md $PKG_DIR/
    cp LANGUAGE_SPEC.md $PKG_DIR/docs/
    cp LANGUAGE_GUIDE.md $PKG_DIR/docs/
    cp LICENSE $PKG_DIR/
    
    # Create archive
    cd $DIST_DIR
    if [ "$OS" = "windows" ]; then
        zip -r cortex-$VERSION-$OS-$ARCH.zip cortex-$VERSION-$OS-$ARCH
    else
        tar -czvf cortex-$VERSION-$OS-$ARCH.tar.gz cortex-$VERSION-$OS-$ARCH
    fi
    cd ..
    
    echo "Created package for $OS/$ARCH"
}

# Build for all platforms
mkdir -p $BUILD_DIR

# Windows (amd64)
build_cortex windows amd64 .exe

# Windows (386) - for older systems
build_cortex windows 386 .exe

# Linux (amd64)
build_cortex linux amd64 ""

# Linux (arm64) - for Apple Silicon Linux, Raspberry Pi
build_cortex linux arm64 ""

# macOS (amd64) - Intel Macs
build_cortex darwin amd64 ""

# macOS (arm64) - Apple Silicon
build_cortex darwin arm64 ""

# Cleanup build dir
rm -rf $BUILD_DIR

echo ""
echo "=== Build Complete ==="
echo "Release packages created in $DIST_DIR/"
ls -la $DIST_DIR/
