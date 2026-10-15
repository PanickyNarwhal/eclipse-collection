#!/usr/bin/env bash
# Castlevania Eclipse Collection - Automated Cross-Platform Runner Setup
# Configures contained, portable emulator runners for macOS, Linux, and Windows.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
EMU_DIR="$ROOT_DIR/emulators"
CORES_DIR="$ROOT_DIR/cores"

echo "=========================================================="
echo " ☩ Castlevania Eclipse Collection - Runner Setup ☩"
echo "=========================================================="

mkdir -p "$EMU_DIR/retroarch"
mkdir -p "$EMU_DIR/pcsx2"
mkdir -p "$EMU_DIR/dolphin"
mkdir -p "$CORES_DIR"

OS="$(uname -s)"
ARCH="$(uname -m)"

echo "Detected Platform: $OS ($ARCH)"

case "$OS" in
  Darwin*)
    echo "Configuring for macOS..."
    # If RetroArch is installed in /Applications, link it portably
    if [ -d "/Applications/RetroArch.app" ]; then
      echo "✓ Found system RetroArch.app, linking to local emulators/..."
      ln -sf "/Applications/RetroArch.app" "$EMU_DIR/retroarch/RetroArch.app"
    else
      echo "ℹ RetroArch not found in /Applications."
      echo "  You can download the macOS RetroArch.app and place it into: $EMU_DIR/retroarch/RetroArch.app"
    fi

    if [ -d "/Applications/PCSX2.app" ]; then
      echo "✓ Found system PCSX2.app, linking to local emulators/..."
      ln -sf "/Applications/PCSX2.app" "$EMU_DIR/pcsx2/PCSX2.app"
    fi

    if [ -d "/Applications/Dolphin.app" ]; then
      echo "✓ Found system Dolphin.app, linking to local emulators/..."
      ln -sf "/Applications/Dolphin.app" "$EMU_DIR/dolphin/Dolphin.app"
    fi
    ;;

  Linux*)
    echo "Configuring for Linux..."
    echo "Checking for system emulators..."
    which retroarch 2>/dev/null && echo "✓ Found retroarch" || echo "ℹ Run: sudo apt install retroarch (or use Flatpak/AppImage)"
    which pcsx2-qt 2>/dev/null && echo "✓ Found pcsx2-qt" || echo "ℹ Run: sudo apt install pcsx2-qt"
    which dolphin-emu 2>/dev/null && echo "✓ Found dolphin-emu" || echo "ℹ Run: sudo apt install dolphin-emu"
    ;;

  *)
    echo "For Windows, please run: scripts\\setup_runners.bat"
    ;;
esac

# Create seamless, borderless configuration for RetroArch
CFG_FILE="$EMU_DIR/retroarch/retroarch.cfg"
if [ ! -f "$CFG_FILE" ]; then
  cat << 'EOF' > "$CFG_FILE"
video_fullscreen = "true"
video_windowed_fullscreen = "true"
pause_nonactive = "true"
notification_show_when_menu_is_alive = "false"
video_msg_pos_x = "0.05"
video_msg_pos_y = "0.05"
quit_press_twice = "false"
input_enable_hotkey_btn = "8"
input_exit_emulator_btn = "9"
EOF
  echo "✓ Generated silent console profile: $CFG_FILE"
fi

echo ""
echo "Setup complete! Ready to launch Castlevania: Eclipse Collection."
