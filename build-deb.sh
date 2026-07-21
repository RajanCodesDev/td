#!/usr/bin/env bash
set -e

VERSION=1.5.0
ARCH=amd64
PKG="td_${VERSION}_${ARCH}"

echo "Building td..."
go build -o td .

echo "Preparing package..."
rm -rf "$PKG"

mkdir -p "$PKG/DEBIAN"
mkdir -p "$PKG/usr/bin"

install -m755 td "$PKG/usr/bin/td"

cat > "$PKG/DEBIAN/control" <<EOF
Package: td
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: Shivendra Rajan <you@example.com>
Description: Simple terminal todo manager written in Go.
 A lightweight CLI task manager with SQLite persistence
 and editor integration.
EOF

echo "Building deb..."
dpkg-deb --build "$PKG"

echo
echo "Package created:"
echo "  ${PKG}.deb"
echo

dpkg -c "${PKG}.deb"