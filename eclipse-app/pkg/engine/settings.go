package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// LauncherSettings stores user-configurable graphics, audio, and container preferences.
// These settings are automatically translated into isolated emulator configuration files
// inside emulators/<runner>/ so the user never has to configure individual emulators.
type LauncherSettings struct {
	DisplayMode        string `json:"display_mode"`        // "fullscreen" | "windowed"
	InternalResolution int    `json:"internal_resolution"` // 1 (Native), 2 (720p), 3 (1080p), 4 (1440p), 6 (4K)
	AspectRatio        string `json:"aspect_ratio"`        // "4:3" | "16:9" | "stretch"
	Renderer           string `json:"renderer"`            // "auto" | "metal" | "vulkan" | "opengl" | "software"
	ShaderFilter       string `json:"shader_filter"`       // "sharp" | "crt" | "smooth"
	VSync              bool   `json:"vsync"`               // true | false
	ShowFPS            bool   `json:"show_fps"`            // true | false
	MenuMusic          bool   `json:"menu_music"`          // true | false
	MenuVolume         int    `json:"menu_volume"`         // 0 - 100
	UISfx              bool   `json:"ui_sfx"`              // true | false
	AudioSync          bool   `json:"audio_sync"`          // true | false
	CustomRomDir       string `json:"custom_rom_dir"`      // Optional extra ROM scan folder
}

// DefaultSettings returns the default console-grade settings profile.
func DefaultSettings() LauncherSettings {
	return LauncherSettings{
		DisplayMode:        "fullscreen",
		InternalResolution: 3, // 1080p Full HD default
		AspectRatio:        "4:3",
		Renderer:           "auto",
		ShaderFilter:       "sharp",
		VSync:              true,
		ShowFPS:            false,
		MenuMusic:          true,
		MenuVolume:         50,
		UISfx:              true,
		AudioSync:          true,
		CustomRomDir:       "",
	}
}

// LoadSettings reads settings.json from disk or returns defaults.
func LoadSettings(baseDir string) LauncherSettings {
	settingsPath := filepath.Join(baseDir, "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		defaults := DefaultSettings()
		_ = SaveSettings(baseDir, defaults)
		return defaults
	}

	var s LauncherSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings()
	}
	if s.InternalResolution < 1 {
		s.InternalResolution = 3
	}
	if s.DisplayMode == "" {
		s.DisplayMode = "fullscreen"
	}
	if s.AspectRatio == "" {
		s.AspectRatio = "4:3"
	}
	if s.Renderer == "" {
		s.Renderer = "auto"
	}
	if s.ShaderFilter == "" {
		s.ShaderFilter = "sharp"
	}
	return s
}

// SaveSettings writes settings.json to disk and immediately updates all container configs.
func SaveSettings(baseDir string, s LauncherSettings) error {
	settingsPath := filepath.Join(baseDir, "settings.json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return err
	}
	return ApplySettingsToContainers(baseDir, s)
}

// ApplySettingsToContainers generates isolated config files inside emulators/
// so RetroArch, PCSX2, and Dolphin run completely containerized and respect launcher graphics settings.
func ApplySettingsToContainers(baseDir string, s LauncherSettings) error {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		absBase = baseDir
	}

	if err := writeRetroArchContainerConfig(absBase, s); err != nil {
		fmt.Printf("Warning writing RetroArch container config: %v\n", err)
	}
	if err := writePCSX2ContainerConfig(absBase, s); err != nil {
		fmt.Printf("Warning writing PCSX2 container config: %v\n", err)
	}
	if err := writeDolphinContainerConfig(absBase, s); err != nil {
		fmt.Printf("Warning writing Dolphin container config: %v\n", err)
	}
	return nil
}

func writeRetroArchContainerConfig(absBase string, s LauncherSettings) error {
	raDir := filepath.Join(absBase, "emulators", "retroarch")
	cfgDir := filepath.Join(raDir, "config")
	coresDir := filepath.Join(raDir, "cores")
	savesDir := filepath.Join(raDir, "saves")
	statesDir := filepath.Join(raDir, "states")
	systemDir := filepath.Join(raDir, "system")

	for _, d := range []string{cfgDir, coresDir, savesDir, statesDir, systemDir} {
		_ = os.MkdirAll(d, 0755)
	}

	fullscreen := "true"
	if s.DisplayMode == "windowed" {
		fullscreen = "false"
	}

	vsync := "true"
	if !s.VSync {
		vsync = "false"
	}

	smooth := "false"
	if s.ShaderFilter == "smooth" {
		smooth = "true"
	}

	fpsShow := "false"
	if s.ShowFPS {
		fpsShow = "true"
	}

	audioSync := "true"
	if !s.AudioSync {
		audioSync = "false"
	}

	// RetroArch aspect_ratio_index: 0 = 4:3, 1 = 16:9, 2 = Config/Stretch, 22 = Core Provided
	aspectIndex := "0"
	if s.AspectRatio == "16:9" {
		aspectIndex = "1"
	} else if s.AspectRatio == "stretch" {
		aspectIndex = "2"
	}

	// IMPORTANT: On macOS RetroArch 1.22.2 Universal, available video drivers are "gl" and "null".
	// Setting "metal" or "glcore" fails driver initialization on Intel macOS.
	videoDriver := "gl"
	if runtime.GOOS != "darwin" {
		switch s.Renderer {
		case "vulkan":
			videoDriver = "vulkan"
		case "opengl":
			videoDriver = "glcore"
		default:
			videoDriver = "gl"
		}
	}

	// pause_nonactive MUST be "false" so minimizing the launcher window on game start does not freeze RetroArch
	cfgContent := fmt.Sprintf(`# Castlevania Eclipse Collection - Isolated RetroArch Container Config
# Automatically generated by Launcher Settings. Do not edit manually.
config_save_on_exit = "false"
video_driver = "%s"
video_fullscreen = "%s"
video_windowed_fullscreen = "true"
video_vsync = "%s"
video_smooth = "%s"
video_aspect_ratio_auto = "false"
aspect_ratio_index = "%s"
video_windowed_position_width = "1600"
video_windowed_position_height = "900"
fps_show = "%s"
audio_sync = "%s"
pause_nonactive = "false"
notification_show_when_menu_is_alive = "false"
notification_show_autoconfig = "false"
notification_show_config_override_load = "false"
notification_show_remap_load = "false"
quit_press_twice = "false"
menu_driver = "ozone"

# Container Isolation Directories (Never touches system RetroArch)
libretro_directory = "%s"
system_directory = "%s"
savefile_directory = "%s"
savestate_directory = "%s"
rgui_config_directory = "%s"
content_history_dir = "%s"
`,
		videoDriver, fullscreen, vsync, smooth, aspectIndex, fpsShow, audioSync,
		filepath.ToSlash(coresDir),
		filepath.ToSlash(systemDir),
		filepath.ToSlash(savesDir),
		filepath.ToSlash(statesDir),
		filepath.ToSlash(cfgDir),
		filepath.ToSlash(cfgDir),
	)

	if err := os.WriteFile(filepath.Join(cfgDir, "retroarch.cfg"), []byte(cfgContent), 0644); err != nil {
		return err
	}

	// Write core options for PS1 & PCE-CD upscaling
	swanOptDir := filepath.Join(cfgDir, "SwanStation")
	beetlePsxDir := filepath.Join(cfgDir, "Beetle PSX HW")
	pcsxRearmedDir := filepath.Join(cfgDir, "PCSX-ReARMed")
	_ = os.MkdirAll(swanOptDir, 0755)
	_ = os.MkdirAll(beetlePsxDir, 0755)
	_ = os.MkdirAll(pcsxRearmedDir, 0755)

	widescreen := "false"
	if s.AspectRatio == "16:9" {
		widescreen = "true"
	}

	swanOpts := fmt.Sprintf(`duckstation_GPU.ResolutionScale = "%d"
duckstation_Display.AspectRatio = "%s"
duckstation_GPU.WidescreenHack = "%s"
duckstation_Main.FastBoot = "true"
`, s.InternalResolution, s.AspectRatio, widescreen)
	_ = os.WriteFile(filepath.Join(swanOptDir, "SwanStation.opt"), []byte(swanOpts), 0644)

	beetleOpts := fmt.Sprintf(`beetle_psx_hw_internal_resolution = "%dx"
beetle_psx_hw_widescreen_hack = "%s"
`, s.InternalResolution, map[string]string{"true": "enabled", "false": "disabled"}[widescreen])
	_ = os.WriteFile(filepath.Join(beetlePsxDir, "Beetle PSX HW.opt"), []byte(beetleOpts), 0644)

	return nil
}

func writePCSX2ContainerConfig(absBase string, s LauncherSettings) error {
	pcsx2Dir := filepath.Join(absBase, "emulators", "pcsx2")
	// Note: On macOS, `PCSX2 -datapath <pcsx2Dir>` uses `<pcsx2Dir>/PCSX2` as DataRoot.
	// We prepare both `<pcsx2Dir>` and `<pcsx2Dir>/PCSX2` so either path mode works seamlessly.
	pcsx2SubDir := filepath.Join(pcsx2Dir, "PCSX2")

	// CRITICAL FIX: Never place portable.ini inside PCSX2.app/Contents/Resources/ on macOS,
	// because adding any file inside *.app/Contents/ breaks macOS _CodeSignature and causes
	// Gatekeeper to report "PCSX2.app is damaged/corrupt". Remove it if present from earlier versions.
	corruptMarker := filepath.Join(pcsx2Dir, "PCSX2.app", "Contents", "Resources", "portable.ini")
	_ = os.Remove(corruptMarker)

	biosDir := filepath.Join(pcsx2Dir, "bios")

	for _, base := range []string{pcsx2Dir, pcsx2SubDir} {
		for _, sub := range []string{"inis", "bios", "memcards", "sstates", "snaps", "logs", "cheats", "patches", "cache"} {
			_ = os.MkdirAll(filepath.Join(base, sub), 0755)
		}
	}

	fullscreen := "true"
	if s.DisplayMode == "windowed" {
		fullscreen = "false"
	}

	vsync := "true"
	if !s.VSync {
		vsync = "false"
	}

	showFps := "false"
	if s.ShowFPS {
		showFps = "true"
	}

	aspectRatio := "4:3"
	if s.AspectRatio == "16:9" {
		aspectRatio = "16:9"
	} else if s.AspectRatio == "stretch" {
		aspectRatio = "Stretch"
	}

	// PCSX2 Renderer: -1 = Auto, 14 = Metal, 12 = OpenGL, 13 = Software, 11 = Vulkan
	rendererCode := "-1"
	switch s.Renderer {
	case "metal":
		rendererCode = "14"
	case "vulkan":
		rendererCode = "11"
	case "opengl":
		rendererCode = "12"
	case "software":
		rendererCode = "13"
	}

	bilinear := "1"
	if s.ShaderFilter == "sharp" {
		bilinear = "0"
	}

	scanlines := "0"
	if s.ShaderFilter == "crt" {
		scanlines = "1"
	}

	// Detect primary PS2 BIOS filename in bios/ so [Filenames] BIOS = <name> is pre-selected
	biosFilename := "SCPH-39001.bin"
	if entries, err := os.ReadDir(biosDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				ext := filepath.Ext(e.Name())
				if ext == ".bin" || ext == ".rom" || ext == ".BIN" || ext == ".ROM" {
					biosFilename = e.Name()
					break
				}
			}
		}
	}

	// Complete PCSX2 v2.x configuration with SetupWizardIncomplete = false, fastboot, BIOS pre-selected, and DualShock2 Pad1 bindings
	iniContent := fmt.Sprintf(`[UI]
SettingsVersion = 1
SetupWizardIncomplete = false
InhibitScreensaver = true
ConfirmShutdown = false
StartPaused = false
PauseOnFocusLoss = false
StartFullscreen = %s
DoubleClickTogglesFullscreen = true
HideMouseCursor = true
RenderToSeparateWindow = false
HideMainWindowWhenRunning = true

[Folders]
Bios = bios
Snapshots = snaps
Savestates = sstates
MemoryCards = memcards
Logs = logs
Cheats = cheats
Patches = patches
Cache = cache

[Filenames]
BIOS = %s

[EmuCore]
EnablePatches = true
EnableWideScreenPatches = %v
EnableNoInterlacingPatches = true
EnableFastBoot = true
EnableGameFixes = true

[EmuCore/GS]
Renderer = %s
upscale_multiplier = %d
AspectRatio = %s
VsyncEnable = %s
OsdShowFPS = %s
linear_present_mode = %s
TVShader = %s
pcrtc_antiblur = true

[SPU2/Output]
StandardVolume = 100
OutputMuted = false
Backend = Cubeb
SyncMode = TimeStretch

[MemoryCards]
Slot1_Enable = true
Slot1_Filename = Mcd001.ps2
Slot2_Enable = true
Slot2_Filename = Mcd002.ps2

[Logging]
EnableFileLogging = true

[InputSources]
Keyboard = true
Mouse = true
SDL = true
SDLControllerEnhancedMode = true

[Hotkeys]
ToggleFullscreen = Keyboard/Alt & Keyboard/Return
OpenPauseMenu = Keyboard/Escape
SaveStateToSlot = Keyboard/F1
LoadStateFromSlot = Keyboard/F3

[Pad1]
Type = DualShock2
Up = Keyboard/Up
Right = Keyboard/Right
Down = Keyboard/Down
Left = Keyboard/Left
Triangle = Keyboard/I
Circle = Keyboard/L
Cross = Keyboard/K
Square = Keyboard/J
Select = Keyboard/Backspace
Start = Keyboard/Return
L1 = Keyboard/Q
L2 = Keyboard/1
R1 = Keyboard/E
R2 = Keyboard/3
L3 = Keyboard/2
R3 = Keyboard/4
LUp = Keyboard/W
LRight = Keyboard/D
LDown = Keyboard/S
LLeft = Keyboard/A
RUp = Keyboard/T
RRight = Keyboard/H
RDown = Keyboard/G
RLeft = Keyboard/F
`,
		fullscreen,
		biosFilename,
		s.AspectRatio == "16:9",
		rendererCode,
		s.InternalResolution,
		aspectRatio,
		vsync,
		showFps,
		bilinear,
		scanlines,
	)

	_ = os.WriteFile(filepath.Join(pcsx2Dir, "inis", "PCSX2.ini"), []byte(iniContent), 0644)
	return os.WriteFile(filepath.Join(pcsx2SubDir, "inis", "PCSX2.ini"), []byte(iniContent), 0644)
}

func writeDolphinContainerConfig(absBase string, s LauncherSettings) error {
	dolphinDir := filepath.Join(absBase, "emulators", "dolphin")
	userDir := filepath.Join(dolphinDir, "user")
	cfgDir := filepath.Join(userDir, "Config")
	wiiDir := filepath.Join(userDir, "Wii")

	for _, d := range []string{cfgDir, wiiDir} {
		_ = os.MkdirAll(d, 0755)
	}

	fullscreen := "True"
	if s.DisplayMode == "windowed" {
		fullscreen = "False"
	}

	vsync := "True"
	if !s.VSync {
		vsync = "False"
	}

	showFps := "False"
	if s.ShowFPS {
		showFps = "True"
	}

	// Dolphin AspectRatio: 0 = Auto, 1 = 16:9, 2 = 4:3, 3 = Stretch
	aspectCode := "2"
	if s.AspectRatio == "16:9" {
		aspectCode = "1"
	} else if s.AspectRatio == "stretch" {
		aspectCode = "3"
	}

	gfxBackend := "Metal"
	if runtime.GOOS != "darwin" {
		gfxBackend = "Vulkan"
	}
	switch s.Renderer {
	case "vulkan":
		gfxBackend = "Vulkan"
	case "opengl":
		gfxBackend = "OGL"
	case "software":
		gfxBackend = "Software Renderer"
	}

	// CRITICAL FIX: Include [Analytics] PermissionAsked = True so Dolphin never blocks
	// on an invisible modal telemetry popup when launched in batch (-b) mode.
	dolphinIni := fmt.Sprintf(`[General]
ShowLag = False
ShowFrameCount = False

[Interface]
ConfirmStop = False
UsePanicHandlers = False
OnScreenDisplayMessages = False

[Display]
Fullscreen = %s
RenderToMain = True

[Analytics]
ID = 00000000000000000000000000000000
Enabled = False
PermissionAsked = True

[AutoUpdate]
UpdateTrack = 

[Core]
GFXBackend = %s
FastDiscSpeed = True
CPUThread = True
`, fullscreen, gfxBackend)

	gfxIni := fmt.Sprintf(`[Settings]
InternalResolution = %d
AspectRatio = %s
ShowFPS = %s

[Hardware]
VSync = %s
`, s.InternalResolution, aspectCode, showFps, vsync)

	// Default horizontal Wii Remote & Classic Controller bindings for Castlevania: The Adventure ReBirth
	wiimoteIni := `[Wiimote1]
Device = Quartz/0/Keyboard & Mouse
Buttons/A = Z
Buttons/B = X
Buttons/1 = X
Buttons/2 = Z
Buttons/- = Backspace
Buttons/+ = Return
Buttons/Home = Escape
D-Pad/Up = Left
D-Pad/Down = Right
D-Pad/Left = Down
D-Pad/Right = Up
Options/Sideways Wiimote = 1.0000000000000000
Source = 1
`
	wiimotePath := filepath.Join(cfgDir, "WiimoteNew.ini")
	if info, err := os.Stat(wiimotePath); err != nil || info.Size() == 0 {
		_ = os.WriteFile(wiimotePath, []byte(wiimoteIni), 0644)
	}

	_ = os.WriteFile(filepath.Join(cfgDir, "Dolphin.ini"), []byte(dolphinIni), 0644)
	return os.WriteFile(filepath.Join(cfgDir, "GFX.ini"), []byte(gfxIni), 0644)
}
