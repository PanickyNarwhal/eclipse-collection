#!/usr/bin/env bash
# Castlevania Eclipse Collection - Containerized Portable Runner Setup (v1.0.5)
# Downloads and sandboxes isolated portable emulators, libretro cores, and BIOS files into emulators/
# without touching or modifying any existing system installations or breaking macOS bundle seals.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
EMU_DIR="$ROOT_DIR/emulators"
CACHE_DIR="$EMU_DIR/.cache"
CORES_DIR="$EMU_DIR/retroarch/cores"
ROOT_CORES_DIR="$ROOT_DIR/cores"
RA_SYS_DIR="$EMU_DIR/retroarch/system"
PCSX2_BIOS_DIR="$EMU_DIR/pcsx2/bios"
PCSX2_SUB_BIOS_DIR="$EMU_DIR/pcsx2/PCSX2/bios"

echo "=========================================================="
echo " ☩ Castlevania Eclipse Collection - Container Setup ☩"
echo "=========================================================="

mkdir -p "$CACHE_DIR"
mkdir -p "$EMU_DIR/retroarch/config" "$EMU_DIR/retroarch/cores" "$EMU_DIR/retroarch/saves" "$EMU_DIR/retroarch/states" "$RA_SYS_DIR"
mkdir -p "$EMU_DIR/pcsx2/inis" "$PCSX2_BIOS_DIR" "$EMU_DIR/pcsx2/memcards" "$EMU_DIR/pcsx2/sstates"
mkdir -p "$EMU_DIR/pcsx2/PCSX2/inis" "$PCSX2_SUB_BIOS_DIR" "$EMU_DIR/pcsx2/PCSX2/memcards" "$EMU_DIR/pcsx2/PCSX2/sstates"
mkdir -p "$EMU_DIR/dolphin/user/Config" "$EMU_DIR/dolphin/user/Wii"
mkdir -p "$ROOT_CORES_DIR" "$ROOT_DIR/bios"

OS="$(uname -s)"
ARCH="$(uname -m)"

echo "Detected Platform: $OS ($ARCH)"

case "$OS" in
  Darwin*)
    echo "Provisioning isolated macOS containers in $EMU_DIR..."

    # Remove any portable.ini inside PCSX2.app/Contents/Resources that breaks _CodeSignature
    rm -f "$EMU_DIR/pcsx2/PCSX2.app/Contents/Resources/portable.ini" 2>/dev/null || true

    # 1. RetroArch Container
    if [ ! -f "$EMU_DIR/retroarch/RetroArch.app/Contents/MacOS/RetroArch" ] || ! codesign -vv "$EMU_DIR/retroarch/RetroArch.app" >/dev/null 2>&1; then
      rm -rf "$EMU_DIR/retroarch/RetroArch.app"
      if [ -d "/Applications/RetroArch.app" ]; then
        echo "-> Cloning RetroArch.app into isolated container..."
        ditto "/Applications/RetroArch.app" "$EMU_DIR/retroarch/RetroArch.app"
      else
        echo "-> Downloading portable RetroArch (macOS Universal)..."
        RA_URL=$(curl -sL "https://formulae.brew.sh/api/cask/retroarch-metal.json" | sed -n 's/.*"url":"\([^"]*\)".*/\1/p' | head -n 1)
        [ -z "$RA_URL" ] && RA_URL="https://buildbot.libretro.com/stable/1.22.2/apple/osx/universal/RetroArch_Metal.dmg"
        curl -L --progress-bar "$RA_URL" -o "$CACHE_DIR/RetroArch.dmg"
        MNT="$CACHE_DIR/ra_mnt"
        mkdir -p "$MNT"
        hdiutil attach "$CACHE_DIR/RetroArch.dmg" -nobrowse -readonly -mountpoint "$MNT" >/dev/null
        ditto "$MNT/RetroArch.app" "$EMU_DIR/retroarch/RetroArch.app"
        hdiutil detach "$MNT" -force >/dev/null
        rmdir "$MNT"
      fi
    fi
    xattr -rd com.apple.quarantine "$EMU_DIR/retroarch/RetroArch.app" 2>/dev/null || true
    echo "✓ RetroArch container ready ($EMU_DIR/retroarch/RetroArch.app)"

    # 2. Libretro Cores (PCE-CD & PS1)
    CORE_ARCH="x86_64"
    [ "$ARCH" = "arm64" ] && CORE_ARCH="arm64"
    for CORE in mednafen_pce_fast_libretro swanstation_libretro pcsx_rearmed_libretro mednafen_psx_hw_libretro; do
      if [ ! -f "$CORES_DIR/${CORE}.dylib" ]; then
        echo "-> Downloading core: ${CORE}.dylib ($CORE_ARCH)..."
        curl -sL "https://buildbot.libretro.com/nightly/apple/osx/${CORE_ARCH}/latest/${CORE}.dylib.zip" -o "$CACHE_DIR/${CORE}.zip" && \
        unzip -qo "$CACHE_DIR/${CORE}.zip" -d "$CORES_DIR"
      fi
      if [ -f "$CORES_DIR/${CORE}.dylib" ]; then
        xattr -rd com.apple.quarantine "$CORES_DIR/${CORE}.dylib" 2>/dev/null || true
        codesign --force --sign - "$CORES_DIR/${CORE}.dylib" 2>/dev/null || true
        cp -f "$CORES_DIR/${CORE}.dylib" "$ROOT_CORES_DIR/${CORE}.dylib"
      fi
    done
    echo "✓ Libretro cores signed and ready in $CORES_DIR"

    # 3. PCSX2 Container
    if [ ! -f "$EMU_DIR/pcsx2/PCSX2.app/Contents/MacOS/PCSX2" ] || ! codesign -vv "$EMU_DIR/pcsx2/PCSX2.app" >/dev/null 2>&1; then
      rm -rf "$EMU_DIR/pcsx2/PCSX2.app"
      if [ -d "/Applications/PCSX2.app" ]; then
        echo "-> Cloning clean PCSX2.app into isolated container..."
        ditto "/Applications/PCSX2.app" "$EMU_DIR/pcsx2/PCSX2.app"
      else
        echo "-> Downloading portable PCSX2 (macOS)..."
        PCSX2_URL=$(curl -sL "https://formulae.brew.sh/api/cask/pcsx2.json" | sed -n 's/.*"url":"\([^"]*\)".*/\1/p' | head -n 1)
        [ -z "$PCSX2_URL" ] && PCSX2_URL="https://github.com/PCSX2/pcsx2/releases/download/v2.8.2/pcsx2-v2.8.2-macos-Qt.tar.xz"
        curl -L --progress-bar "$PCSX2_URL" -o "$CACHE_DIR/pcsx2.tar.xz"
        mkdir -p "$CACHE_DIR/pcsx2_ext"
        tar -xf "$CACHE_DIR/pcsx2.tar.xz" -C "$CACHE_DIR/pcsx2_ext"
        APP_FOUND=$(find "$CACHE_DIR/pcsx2_ext" -maxdepth 1 -name "*.app" | head -n 1)
        if [ -n "$APP_FOUND" ]; then
          ditto "$APP_FOUND" "$EMU_DIR/pcsx2/PCSX2.app"
        fi
        rm -rf "$CACHE_DIR/pcsx2_ext"
      fi
    fi
    xattr -rd com.apple.quarantine "$EMU_DIR/pcsx2/PCSX2.app" 2>/dev/null || true
    echo "✓ PCSX2 container ready ($EMU_DIR/pcsx2/PCSX2.app)"

    # 4. Dolphin Container
    if [ ! -f "$EMU_DIR/dolphin/Dolphin.app/Contents/MacOS/Dolphin" ] || ! codesign -vv "$EMU_DIR/dolphin/Dolphin.app" >/dev/null 2>&1; then
      rm -rf "$EMU_DIR/dolphin/Dolphin.app"
      if [ -d "/Applications/Dolphin.app" ]; then
        echo "-> Cloning Dolphin.app into isolated container..."
        ditto "/Applications/Dolphin.app" "$EMU_DIR/dolphin/Dolphin.app"
      else
        echo "-> Downloading portable Dolphin (macOS Universal)..."
        DOLPHIN_URL=$(curl -sL "https://formulae.brew.sh/api/cask/dolphin.json" | sed -n 's/.*"url":"\([^"]*\)".*/\1/p' | head -n 1)
        [ -z "$DOLPHIN_URL" ] && DOLPHIN_URL="https://dl.dolphin-emu.org/releases/2609a/dolphin-2609a-universal.dmg"
        curl -L --progress-bar "$DOLPHIN_URL" -o "$CACHE_DIR/Dolphin.dmg"
        MNT="$CACHE_DIR/dolphin_mnt"
        mkdir -p "$MNT"
        hdiutil attach "$CACHE_DIR/Dolphin.dmg" -nobrowse -readonly -mountpoint "$MNT" >/dev/null
        ditto "$MNT/Dolphin.app" "$EMU_DIR/dolphin/Dolphin.app"
        hdiutil detach "$MNT" -force >/dev/null
        rmdir "$MNT"
      fi
    fi
    xattr -rd com.apple.quarantine "$EMU_DIR/dolphin/Dolphin.app" 2>/dev/null || true
    echo "✓ Dolphin container ready ($EMU_DIR/dolphin/Dolphin.app)"

    # 5. Provision Required System BIOS Files (PCE-CD, PS1, PS2)
    if [ ! -f "$RA_SYS_DIR/syscard3.pce" ]; then
      echo "-> Provisioning PC Engine CD BIOS (syscard3.pce)..."
      curl -sL "https://raw.githubusercontent.com/Abdess/retrobios/main/bios/NEC/PC%20Engine/syscard3.pce" -o "$RA_SYS_DIR/syscard3.pce" || true
    fi
    for PS1B in scph5501.bin scph5500.bin scph5502.bin; do
      if [ ! -f "$RA_SYS_DIR/$PS1B" ]; then
        echo "-> Provisioning PlayStation BIOS ($PS1B)..."
        curl -sL "https://raw.githubusercontent.com/Abdess/retrobios/main/bios/Sony/PlayStation/$PS1B" -o "$RA_SYS_DIR/$PS1B" || true
      fi
    done
    if [ ! -f "$PCSX2_BIOS_DIR/SCPH-39001.bin" ]; then
      echo "-> Provisioning PlayStation 2 BIOS (SCPH-39001.bin)..."
      curl -sL "https://raw.githubusercontent.com/Abdess/retrobios/main/bios/Sony/PlayStation%202/SCPH-39001.bin" -o "$PCSX2_BIOS_DIR/SCPH-39001.bin" || true
      cp -f "$PCSX2_BIOS_DIR/SCPH-39001.bin" "$PCSX2_SUB_BIOS_DIR/SCPH-39001.bin" 2>/dev/null || true
    fi
    echo "✓ System BIOS files ready"
    ;;
esac

echo ""
echo "All isolated emulator containers are configured and verified in: $EMU_DIR"
