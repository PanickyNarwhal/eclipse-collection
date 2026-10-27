#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
RELEASE_DIR="$ROOT_DIR/releases"
VERSION="v1.0.7"
APP_NAME="Castlevania Eclipse Collection"

mkdir -p "$RELEASE_DIR"
export GOCACHE="/tmp/eclipse-go-cache"

echo "==> [1/4] Syncing latest game artwork & logos into frontend/dist/assets/games/..."
for game in rondo sotn loi cod rebirth; do
  mkdir -p "$SCRIPT_DIR/frontend/dist/assets/games/$game"
  if [ -d "$ROOT_DIR/games/$game" ]; then
    for img in "$ROOT_DIR/games/$game"/*.{png,jpg,jpeg,webp,PNG,JPG,JPEG,WEBP}; do
      [ -e "$img" ] || continue
      cp -f "$img" "$SCRIPT_DIR/frontend/dist/assets/games/$game/"
      base="$(basename "$img")"
      lower="$(echo "$base" | tr '[:upper:]' '[:lower:]')"
      ext=".${lower##*.}"
      if [[ "$lower" == box_en.* ]]; then
        cp -f "$img" "$SCRIPT_DIR/frontend/dist/assets/games/$game/box_us${ext}"
      fi
    done
  fi
done

echo "==> [2/4] Building Windows x64 Standalone EXE ($VERSION)..."
cd "$SCRIPT_DIR"
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
  -tags "desktop,production" \
  -ldflags "-w -s -H windowsgui" \
  -o "$RELEASE_DIR/Castlevania-Eclipse-Collection-${VERSION}-Windows-x64.exe" .

# Also build a portable Windows folder + ZIP with config.json and pre-created games/<id>/ directories
WIN_STAGE="/tmp/eclipse-win-stage"
rm -rf "$WIN_STAGE"
mkdir -p "$WIN_STAGE/Castlevania-Eclipse-Collection/games/"{rondo,sotn,loi,cod,rebirth}
cp -f "$RELEASE_DIR/Castlevania-Eclipse-Collection-${VERSION}-Windows-x64.exe" \
      "$WIN_STAGE/Castlevania-Eclipse-Collection/Castlevania-Eclipse-Collection.exe"
cp -f "$ROOT_DIR/config.json" "$WIN_STAGE/Castlevania-Eclipse-Collection/config.json"

cat > "$WIN_STAGE/Castlevania-Eclipse-Collection/README-WINDOWS.txt" <<EOF
Castlevania: Eclipse Collection ($VERSION) — Windows x64 Edition
================================================================

1. Place your ROM files into the corresponding subfolders inside "games/":
   - games/rondo/   (.chd, .cue/.bin, .iso)
   - games/sotn/    (.chd, .cue/.bin, .img)
   - games/loi/     (.iso, .chd, .cso)
   - games/cod/     (.iso, .chd, .cso)
   - games/rebirth/ (.wad)

2. Double-click "Castlevania-Eclipse-Collection.exe" to launch.
3. Open SETTINGS (top-right or press 'S') -> CONTAINERS & ROMS -> click "INITIALIZE CONTAINERS"
   to automatically download and configure isolated RetroArch, PCSX2, and Dolphin containers.
EOF

(cd "$WIN_STAGE" && zip -rq "$RELEASE_DIR/Castlevania-Eclipse-Collection-${VERSION}-Windows-x64-Portable.zip" "Castlevania-Eclipse-Collection")
rm -rf "$WIN_STAGE"

echo "==> [3/4] Building macOS Universal Binary (x86_64 + arm64 Apple Silicon for MacBook Neo)..."
MAC_STAGE="/tmp/eclipse-mac-stage"
rm -rf "$MAC_STAGE"
APP_BUNDLE="$MAC_STAGE/${APP_NAME}.app"
mkdir -p "$APP_BUNDLE/Contents/MacOS"
mkdir -p "$APP_BUNDLE/Contents/Resources"

export CGO_LDFLAGS="-framework UniformTypeIdentifiers"

echo "    -> Compiling darwin/amd64 (Intel Mac)..."
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build \
  -tags "desktop,production" \
  -ldflags "-w -s" \
  -o "/tmp/eclipse-darwin-amd64" .

echo "    -> Compiling darwin/arm64 (Apple Silicon / MacBook Neo)..."
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build \
  -tags "desktop,production" \
  -ldflags "-w -s" \
  -o "/tmp/eclipse-darwin-arm64" .

echo "    -> Merging into Universal Mach-O binary via lipo..."
lipo -create -output "$APP_BUNDLE/Contents/MacOS/eclipse-app" \
  "/tmp/eclipse-darwin-amd64" \
  "/tmp/eclipse-darwin-arm64"
chmod +x "$APP_BUNDLE/Contents/MacOS/eclipse-app"
rm -f "/tmp/eclipse-darwin-amd64" "/tmp/eclipse-darwin-arm64"

cat > "$APP_BUNDLE/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>eclipse-app</string>
    <key>CFBundleIdentifier</key>
    <string>com.eclipsecollection.castlevania</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundleDisplayName</key>
    <string>${APP_NAME}</string>
    <key>CFBundleVersion</key>
    <string>1.0.7</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.7</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>LSMinimumSystemVersion</key>
    <string>11.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

codesign --force --deep --sign - "$APP_BUNDLE"

# Prepare DMG contents: App bundle + pre-structured portable folder layout + README
mkdir -p "$MAC_STAGE/games/"{rondo,sotn,loi,cod,rebirth}
cp -f "$ROOT_DIR/config.json" "$MAC_STAGE/config.json"
ln -s /Applications "$MAC_STAGE/Applications"

cat > "$MAC_STAGE/README-MACOS.txt" <<EOF
Castlevania: Eclipse Collection ($VERSION) — macOS Universal Edition (Intel & Apple Silicon / MacBook Neo)
=========================================================================================================

HOW TO INSTALL ON MACBOOK NEO / MACOS:
1. Create a folder on your Mac (for example on your Desktop or in Documents) and drag the contents
   of this disk image ("Castlevania Eclipse Collection.app", "config.json", and the "games" folder) into it,
   OR drag "Castlevania Eclipse Collection.app" to Applications (it will automatically create a
   "~/Castlevania Eclipse Collection" folder in your Home directory for your ROMs and isolated emulators).
2. Drop your ROMs into "games/rondo", "games/sotn", "games/loi", "games/cod", and "games/rebirth".
3. Open "Castlevania Eclipse Collection.app", click SETTINGS -> CONTAINERS & ROMS -> "INITIALIZE CONTAINERS"
   to automatically download and sandbox RetroArch, PCSX2, Dolphin, and required BIOS files.
EOF

echo "==> [4/4] Packaging macOS Universal DMG..."
DMG_PATH="$RELEASE_DIR/Castlevania-Eclipse-Collection-${VERSION}-macOS-Universal.dmg"
rm -f "$DMG_PATH"
hdiutil create -volname "${APP_NAME} ${VERSION}" \
  -srcfolder "$MAC_STAGE" \
  -ov -format UDZO \
  "$DMG_PATH"

rm -rf "$MAC_STAGE"

echo ""
echo "==> Build Complete! Artifacts written to: $RELEASE_DIR"
ls -lh "$RELEASE_DIR"
