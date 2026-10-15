package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"eclipse-app/pkg/engine"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App coordinates Wails events, engine lifecycle, and frontend IPC bindings.
type App struct {
	ctx        context.Context
	spawner    *engine.Spawner
	library    map[string]engine.Game
	configPath string
	baseDir    string
}

// NewApp creates a new App application struct.
func NewApp() *App {
	baseDir := "."
	// If running from eclipse-app, check if parent directory is root
	if _, err := os.Stat("../config.json"); err == nil {
		baseDir = ".."
	}

	configPath := filepath.Join(baseDir, "config.json")
	library, err := engine.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Warning loading config: %v, using defaults\n", err)
		library = engine.DefaultLibrary()
	}

	spawner := engine.NewSpawner(baseDir)

	return &App{
		spawner:    spawner,
		library:    library,
		configPath: configPath,
		baseDir:    baseDir,
	}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetLibrary returns the full list of games configured in the collection.
func (a *App) GetLibrary() []engine.Game {
	var list []engine.Game
	// Games sorted by release year, with Legacy of Darkness, 64, and ReBirth at the end
	order := []string{"rondo", "sotn", "loi", "cod", "lod", "cv64", "rebirth"}
	for _, id := range order {
		if game, exists := a.library[id]; exists {
			list = append(list, game)
		}
	}
	// Append any custom added games
	for id, game := range a.library {
		found := false
		for _, o := range order {
			if o == id {
				found = true
				break
			}
		}
		if !found {
			list = append(list, game)
		}
	}
	return list
}

// CheckGameStatus reports whether the ROM and runner are ready for a title.
func (a *App) CheckGameStatus(id string, reg string) engine.GameStatus {
	game, exists := a.library[id]
	if !exists {
		return engine.GameStatus{ID: id, Available: false}
	}
	return a.spawner.CheckGameAvailability(game, engine.Region(reg))
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
		runtime.WindowMinimise(a.ctx)
	}

	onExit := func() {
		runtime.WindowUnminimise(a.ctx)
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
