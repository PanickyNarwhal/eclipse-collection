package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Spawner coordinates game execution and process lifecycle monitoring.
type Spawner struct {
	BaseDir      string
	activeCmd    *exec.Cmd
	activeLock   sync.Mutex
	activeGameID string
}

// NewSpawner creates a new runner spawner anchored to the project root.
func NewSpawner(baseDir string) *Spawner {
	if baseDir == "" {
		baseDir = "."
	}
	return &Spawner{
		BaseDir: baseDir,
	}
}

// CheckGameAvailability examines the filesystem for ROMs, patches, and runner binaries.
func (s *Spawner) CheckGameAvailability(g Game, activeRegion Region) GameStatus {
	status := GameStatus{
		ID:            g.ID,
		ActiveRegion:  activeRegion,
		RegionsStatus: make(map[Region]bool),
	}

	runnerPath, runnerErr := s.resolveRunner(g)
	status.RunnerReady = (runnerErr == nil)
	status.RunnerPath = runnerPath

	for reg, target := range g.Regions {
		fullRomPath := filepath.Join(s.BaseDir, target.RomPath)
		if _, err := os.Stat(fullRomPath); err == nil {
			status.RegionsStatus[reg] = true
		} else {
			status.RegionsStatus[reg] = false
		}
	}

	if activeRegion == "" {
		activeRegion = g.DefaultRegion
	}

	if status.RegionsStatus[activeRegion] {
		status.Available = true
	} else {
		status.Available = false
		target := g.Regions[activeRegion]
		status.MissingFile = target.RomPath
	}

	return status
}

// Launch initiates the game process and calls onLaunch and onExit hooks.
func (s *Spawner) Launch(ctx context.Context, g Game, reg Region, onLaunch func(), onExit func()) error {
	s.activeLock.Lock()
	if s.activeCmd != nil && s.activeCmd.Process != nil {
		s.activeLock.Unlock()
		return fmt.Errorf("another game (%s) is already running", s.activeGameID)
	}
	s.activeLock.Unlock()

	if reg == "" {
		reg = g.DefaultRegion
	}

	target, exists := g.Regions[reg]
	if !exists {
		return fmt.Errorf("region %s not configured for %s", reg, g.Title)
	}

	fullRomPath := filepath.Join(s.BaseDir, target.RomPath)
	if _, err := os.Stat(fullRomPath); os.IsNotExist(err) && g.RunnerType != RunnerNative {
		return fmt.Errorf("ROM file not found: %s\nPlease place your ROM in %s", target.RomPath, fullRomPath)
	}

	runnerPath, err := s.resolveRunner(g)
	if err != nil {
		return fmt.Errorf("failed to locate runner for %s: %w", g.Title, err)
	}

	var cmd *exec.Cmd

	switch g.RunnerType {
	case RunnerRetroArch:
		coreFullPath := filepath.Join(s.BaseDir, s.resolveCorePath(g.CorePath))
		cmd = exec.Command(runnerPath, "-L", coreFullPath, "-f", fullRomPath)

	case RunnerPCSX2:
		cmd = exec.Command(runnerPath, "-batch", "-nogui", "-fullscreen", fullRomPath)

	case RunnerDolphin:
		cmd = exec.Command(runnerPath, "-b", "-e", fullRomPath)

	case RunnerNative:
		if runtime.GOOS == "darwin" && strings.HasSuffix(runnerPath, ".app") {
			cmd = exec.Command("open", "-W", runnerPath)
		} else {
			cmd = exec.Command(runnerPath)
		}

	default:
		return fmt.Errorf("unknown runner type: %s", g.RunnerType)
	}

	cmd.Dir = s.BaseDir

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting game process: %w", err)
	}

	s.activeLock.Lock()
	s.activeCmd = cmd
	s.activeGameID = g.ID
	s.activeLock.Unlock()

	if onLaunch != nil {
		onLaunch()
	}

	go func() {
		_ = cmd.Wait()

		s.activeLock.Lock()
		s.activeCmd = nil
		s.activeGameID = ""
		s.activeLock.Unlock()

		if onExit != nil {
			onExit()
		}
	}()

	return nil
}

// IsGameRunning checks if a game session is currently active.
func (s *Spawner) IsGameRunning() (bool, string) {
	s.activeLock.Lock()
	defer s.activeLock.Unlock()
	if s.activeCmd != nil && s.activeCmd.Process != nil {
		return true, s.activeGameID
	}
	return false, ""
}

// resolveRunner determines the executable path for the requested runner across platforms.
func (s *Spawner) resolveRunner(g Game) (string, error) {
	if g.RunnerType == RunnerNative {
		nativePath := filepath.Join(s.BaseDir, g.ExecutablePath)
		if runtime.GOOS == "windows" && !strings.HasSuffix(nativePath, ".exe") {
			nativePath += ".exe"
		}
		if _, err := os.Stat(nativePath); err == nil {
			return nativePath, nil
		}
		return nativePath, nil
	}

	var candidatePaths []string

	switch g.RunnerType {
	case RunnerRetroArch:
		switch runtime.GOOS {
		case "windows":
			candidatePaths = []string{
				filepath.Join(s.BaseDir, "emulators", "retroarch", "retroarch.exe"),
				"C:\\RetroArch-Win64\\retroarch.exe",
				"retroarch",
			}
		case "darwin":
			candidatePaths = []string{
				filepath.Join(s.BaseDir, "emulators", "retroarch", "RetroArch.app", "Contents", "MacOS", "RetroArch"),
				"/Applications/RetroArch.app/Contents/MacOS/RetroArch",
				"retroarch",
			}
		case "linux":
			candidatePaths = []string{
				filepath.Join(s.BaseDir, "emulators", "retroarch", "retroarch"),
				"/usr/bin/retroarch",
				"/usr/local/bin/retroarch",
				"retroarch",
			}
		}

	case RunnerPCSX2:
		switch runtime.GOOS {
		case "windows":
			candidatePaths = []string{
				filepath.Join(s.BaseDir, "emulators", "pcsx2", "pcsx2-qt.exe"),
				filepath.Join(s.BaseDir, "emulators", "pcsx2", "pcsx2.exe"),
				"C:\\PCSX2\\pcsx2-qt.exe",
				"pcsx2-qt",
			}
		case "darwin":
			candidatePaths = []string{
				filepath.Join(s.BaseDir, "emulators", "pcsx2", "PCSX2.app", "Contents", "MacOS", "PCSX2"),
				"/Applications/PCSX2.app/Contents/MacOS/PCSX2",
				"pcsx2-qt",
			}
		case "linux":
			candidatePaths = []string{
				filepath.Join(s.BaseDir, "emulators", "pcsx2", "pcsx2-qt"),
				"/usr/bin/pcsx2-qt",
				"pcsx2-qt",
			}
		}

	case RunnerDolphin:
		switch runtime.GOOS {
		case "windows":
			candidatePaths = []string{
				filepath.Join(s.BaseDir, "emulators", "dolphin", "Dolphin.exe"),
				"C:\\Dolphin\\Dolphin.exe",
				"dolphin-emu",
			}
		case "darwin":
			candidatePaths = []string{
				filepath.Join(s.BaseDir, "emulators", "dolphin", "Dolphin.app", "Contents", "MacOS", "Dolphin"),
				"/Applications/Dolphin.app/Contents/MacOS/Dolphin",
				"dolphin-emu",
			}
		case "linux":
			candidatePaths = []string{
				filepath.Join(s.BaseDir, "emulators", "dolphin", "dolphin-emu"),
				"/usr/bin/dolphin-emu",
				"dolphin-emu",
			}
		}
	}

	for _, path := range candidatePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		if lp, err := exec.LookPath(path); err == nil {
			return lp, nil
		}
	}

	if len(candidatePaths) > 0 {
		return candidatePaths[0], nil
	}
	return "", fmt.Errorf("no runner candidate configured for %s on %s", g.RunnerType, runtime.GOOS)
}

// resolveCorePath ensures the libretro core has the proper OS extension (.dylib, .dll, .so).
func (s *Spawner) resolveCorePath(coreBase string) string {
	if coreBase == "" {
		return ""
	}
	switch runtime.GOOS {
	case "windows":
		if !strings.HasSuffix(coreBase, ".dll") {
			return coreBase + ".dll"
		}
	case "darwin":
		if !strings.HasSuffix(coreBase, ".dylib") {
			return coreBase + ".dylib"
		}
	case "linux":
		if !strings.HasSuffix(coreBase, ".so") {
			return coreBase + ".so"
		}
	}
	return coreBase
}
