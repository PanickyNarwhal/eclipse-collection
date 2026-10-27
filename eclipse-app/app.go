package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"eclipse-app/pkg/engine"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App coordinates Wails events, engine lifecycle, settings, and frontend IPC bindings.
type App struct {
	ctx        context.Context
	spawner    *engine.Spawner
	library    map[string]engine.Game
	configPath string
	baseDir    string
}

// NewApp creates a new App application struct.
func NewApp() *App {
	baseDir := resolveBaseDir()

	// Ensure all active game ROM folders exist on disk (especially for fresh Windows EXE / macOS DMG installs)
	for _, gameID := range []string{"rondo", "sotn", "loi", "cod", "rebirth"} {
		_ = os.MkdirAll(filepath.Join(baseDir, "games", gameID), 0755)
	}

	configPath := filepath.Join(baseDir, "config.json")
	library, err := engine.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Warning loading config: %v, using defaults\n", err)
		library = engine.DefaultLibrary()
	}

	spawner := engine.NewSpawner(baseDir)

	a := &App{
		spawner:    spawner,
		library:    library,
		configPath: configPath,
		baseDir:    baseDir,
	}

	a.syncAssets()

	return a
}

// resolveBaseDir locates the collection root directory across dev mode, Windows .exe, and macOS .app / .dmg launches.
func resolveBaseDir() string {
	// 1. Check current working directory and parent (dev mode / run.sh)
	if _, err := os.Stat("config.json"); err == nil {
		if abs, err := filepath.Abs("."); err == nil {
			return abs
		}
		return "."
	}
	if _, err := os.Stat("../config.json"); err == nil {
		if abs, err := filepath.Abs(".."); err == nil {
			return abs
		}
		return ".."
	}

	// 2. Check relative to executable path (.exe on Windows, or .app/Contents/MacOS/<bin> on macOS)
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)

		// If running inside a macOS .app bundle (<parent>/<App>.app/Contents/MacOS/<bin>)
		slashPath := filepath.ToSlash(exeDir)
		if idx := strings.Index(slashPath, ".app/Contents/MacOS"); idx != -1 {
			appBundleDir := exeDir[:idx+4]
			appParentDir := filepath.Dir(appBundleDir)

			for _, candidate := range []string{appParentDir, filepath.Dir(appParentDir)} {
				if _, err := os.Stat(filepath.Join(candidate, "config.json")); err == nil {
					return candidate
				}
				if _, err := os.Stat(filepath.Join(candidate, "games")); err == nil {
					// Make sure it's not a read-only mounted DMG volume
					if !strings.HasPrefix(filepath.ToSlash(candidate), "/Volumes/") {
						return candidate
					}
				}
			}

			// If launched from a writable user folder (e.g. Desktop or custom folder, not /Volumes/ or /Applications)
			parentSlash := filepath.ToSlash(appParentDir)
			if !strings.HasPrefix(parentSlash, "/Volumes/") && !strings.HasPrefix(parentSlash, "/Applications") {
				return appParentDir
			}

			// If launched directly from inside a mounted read-only DMG (/Volumes/...) or /Applications,
			// use ~/Castlevania Eclipse Collection as the writable collection folder
			if home, err := os.UserHomeDir(); err == nil && home != "" {
				userCollectionDir := filepath.Join(home, "Castlevania Eclipse Collection")
				_ = os.MkdirAll(userCollectionDir, 0755)
				return userCollectionDir
			}
		}

		// Windows / Linux standalone executable: check exeDir and its parent
		for _, candidate := range []string{exeDir, filepath.Dir(exeDir)} {
			if _, err := os.Stat(filepath.Join(candidate, "config.json")); err == nil {
				return candidate
			}
		}
		return exeDir
	}

	return "."
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetLibrary returns the active list of games configured in the collection.
func (a *App) GetLibrary() []engine.Game {
	var list []engine.Game
	// Active games sorted chronologically: Rondo -> SotN -> LoI -> CoD -> ReBirth
	// Note: Legacy of Darkness ("lod") and Castlevania 64 ("cv64") remain in code/config in reserve, but are omitted from active display
	order := []string{"rondo", "sotn", "loi", "cod", "rebirth"}
	for _, id := range order {
		if game, exists := a.library[id]; exists {
			list = append(list, game)
		}
	}
	return list
}

// CheckGameStatus reports whether the ROM and containerized runner are ready for a title.
func (a *App) CheckGameStatus(id string, reg string) engine.GameStatus {
	game, exists := a.library[id]
	if !exists {
		return engine.GameStatus{ID: id, Available: false}
	}
	return a.spawner.CheckGameAvailability(game, engine.Region(reg))
}

// GetSettings returns the current launcher graphics, audio, and container settings.
func (a *App) GetSettings() engine.LauncherSettings {
	return engine.LoadSettings(a.baseDir)
}

// SaveSettings persists the user's settings and immediately writes updated configs into all emulator containers.
func (a *App) SaveSettings(s engine.LauncherSettings) error {
	return engine.SaveSettings(a.baseDir, s)
}

// GetContainersStatus returns the readiness and isolation status of RetroArch, PCSX2, and Dolphin containers.
func (a *App) GetContainersStatus() []engine.ContainerStatus {
	return a.spawner.GetContainersStatus()
}

// ProvisionContainers downloads and initializes all isolated emulator containers inside emulators/.
func (a *App) ProvisionContainers() (string, error) {
	onProgress := func(msg string) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "container:progress", msg)
		}
	}
	if err := a.spawner.ProvisionAllContainers(onProgress); err != nil {
		return "", err
	}
	return "All emulator containers are initialized and isolated in emulators/", nil
}

// ProvisionSingleContainer downloads and initializes a specific emulator container inside emulators/<runner>/.
func (a *App) ProvisionSingleContainer(runner string) (string, error) {
	onProgress := func(msg string) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "container:progress", msg)
		}
	}
	if err := a.spawner.ProvisionContainer(engine.RunnerType(runner), onProgress); err != nil {
		return "", err
	}
	return fmt.Sprintf("Container '%s' initialized in emulators/%s/", runner, runner), nil
}

// ScanRomDirectory opens a folder dialog (or scans games/ and bios/) to import ROMs and BIOS files.
func (a *App) ScanRomDirectory(promptDialog bool) (engine.ScanResult, error) {
	targetDir := ""
	if promptDialog && a.ctx != nil {
		selected, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
			Title: "Select Folder to Scan for Castlevania ROMs & BIOS Files",
		})
		if err == nil && selected != "" {
			targetDir = selected
		} else if err != nil || selected == "" {
			// User cancelled dialog; scan default games/ directory instead
			targetDir = filepath.Join(a.baseDir, "games")
		}
	}
	res, err := a.spawner.ScanAndImportDirectory(targetDir)
	a.syncAssets()
	return res, err
}

// LaunchGame launches the game with the selected region and manages the window lifecycle.
func (a *App) LaunchGame(id string, reg string) (string, error) {
	game, exists := a.library[id]
	if !exists {
		return "", fmt.Errorf("game %s not found in library", id)
	}

	activeRegion := engine.Region(reg)
	if activeRegion == "" {
		activeRegion = game.DefaultRegion
	}

	onLaunch := func() {
		runtime.EventsEmit(a.ctx, "game:started", id)
	}

	onExit := func() {
		runtime.WindowShow(a.ctx)
		runtime.EventsEmit(a.ctx, "game:stopped", id)
	}

	err := a.spawner.Launch(a.ctx, game, activeRegion, onLaunch, onExit)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Launched %s (%s)", game.Title, activeRegion), nil
}

// IsGameRunning checks if an emulation session is active.
func (a *App) IsGameRunning() map[string]interface{} {
	running, gameID := a.spawner.IsGameRunning()
	return map[string]interface{}{
		"running": running,
		"game_id": gameID,
	}
}

// ToggleFullscreen toggles between fullscreen and windowed mode.
func (a *App) ToggleFullscreen() {
	runtime.WindowToggleMaximise(a.ctx)
}

// Quit terminates the launcher application.
func (a *App) Quit() {
	runtime.Quit(a.ctx)
}

// syncAssets automatically scans root games/<id>/ for newly dropped artwork/logos
// and mirrors them into frontend/dist/assets/games/<id>/, creating regional aliases if needed.
func (a *App) syncAssets() {
	gamesDir := filepath.Join(a.baseDir, "games")
	distAssetsDir := filepath.Join(a.baseDir, "eclipse-app", "frontend", "dist", "assets", "games")
	if _, err := os.Stat("frontend/dist/assets/games"); err == nil {
		distAssetsDir = "frontend/dist/assets/games"
	}

	for id := range a.library {
		srcDir := filepath.Join(gamesDir, id)
		dstDir := filepath.Join(distAssetsDir, id)

		entries, err := os.ReadDir(srcDir)
		if err != nil {
			continue
		}

		os.MkdirAll(dstDir, 0755)

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			lower := strings.ToLower(name)
			ext := strings.ToLower(filepath.Ext(name))
			if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" {
				continue
			}

			srcFile := filepath.Join(srcDir, name)
			dstFile := filepath.Join(dstDir, name)

			copyFile(srcFile, dstFile)

			// If user dropped box_en.*, also mirror as box_us.*
			if strings.HasPrefix(lower, "box_en.") {
				dstUs := filepath.Join(dstDir, "box_us"+ext)
				copyFile(srcFile, dstUs)
			}
		}
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
