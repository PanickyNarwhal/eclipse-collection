@echo off
REM Castlevania Eclipse Collection - Automated Windows Runner Setup
title Castlevania Eclipse Collection Setup

echo ==========================================================
echo  Castlevania Eclipse Collection - Windows Runner Setup
echo ==========================================================

set ROOT_DIR=%~dp0..
set EMU_DIR=%ROOT_DIR%\emulators
set CORES_DIR=%ROOT_DIR%\cores

if not exist "%EMU_DIR%\retroarch" mkdir "%EMU_DIR%\retroarch"
if not exist "%EMU_DIR%\pcsx2" mkdir "%EMU_DIR%\pcsx2"
if not exist "%EMU_DIR%\dolphin" mkdir "%EMU_DIR%\dolphin"
if not exist "%CORES_DIR%" mkdir "%CORES_DIR%"

echo Setting up portable directories in %EMU_DIR%...

REM Silent RetroArch config
set CFG_FILE=%EMU_DIR%\retroarch\retroarch.cfg
if not exist "%CFG_FILE%" (
  (
    echo video_fullscreen = "true"
    echo video_windowed_fullscreen = "true"
    echo pause_nonactive = "true"
    echo notification_show_when_menu_is_alive = "false"
    echo quit_press_twice = "false"
    echo input_enable_hotkey_btn = "8"
    echo input_exit_emulator_btn = "9"
  ) > "%CFG_FILE%"
  echo Generated console profile: %CFG_FILE%
)

echo.
echo Windows setup complete!
pause
