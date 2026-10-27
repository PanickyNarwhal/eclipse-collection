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

// Spawner coordinates game execution and isolated container lifecycle monitoring.
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
	s := &Spawner{
		BaseDir: baseDir,
	}
	s.EnsureContainerDirs()
	return s
}

// effectiveRunnerType returns the actual runner to use.
// For SotN, if a native recomp binary is not present on disk, it automatically falls back to the containerized RetroArch PS1 runner.
func (s *Spawner) effectiveRunnerType(g Game) RunnerType {
	if g.RunnerType == RunnerNative {
		nativePath := filepath.Join(s.BaseDir, g.ExecutablePath)
		if runtime.GOOS == "windows" && !strings.HasSuffix(nativePath, ".exe") {
			nativePath += ".exe"
		}
		if _, err := os.Stat(nativePath); err == nil {
			return RunnerNative
		}
		if g.CorePath != "" || g.ID == "sotn" {
			return RunnerRetroArch
		}
	}
	return g.RunnerType
}

// ResolveRomPath scans the game's folder for exact or variant ROM names.
func (s *Spawner) ResolveRomPath(g Game, reg Region) (string, bool) {
	if g.RunnerType == RunnerNative {
		nativePath := filepath.Join(s.BaseDir, g.ExecutablePath)
		if runtime.GOOS == "windows" && !strings.HasSuffix(nativePath, ".exe") {
			nativePath += ".exe"
		}
		if _, err := os.Stat(nativePath); err == nil {
			return nativePath, true
		}
	}

	// 1. Check exact configured path first
	if target, exists := g.Regions[reg]; exists && target.RomPath != "" {
		exactPath := filepath.Join(s.BaseDir, target.RomPath)
		if _, err := os.Stat(exactPath); err == nil {
			return exactPath, true
		}
	}

	// 2. Scan the game's directory for any compatible ROM file
	gameDir := filepath.Join(s.BaseDir, "games", g.ID)
	entries, err := os.ReadDir(gameDir)
	if err != nil {
		return "", false
	}

	// Prefer primary cue/chd/iso/wad files over raw .img/.bin tracks when both exist
	primaryExts := map[string]bool{
		".chd": true, ".cue": true, ".iso": true,
		".z64": true, ".n64": true, ".v64": true, ".wad": true,
		".cso": true, ".pce": true, ".wbfs": true, ".dol": true,
	}
	secondaryExts := map[string]bool{
		".bin": true, ".img": true,
	}

	var candidates []string
	var secondaryCandidates []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		lower := strings.ToLower(name)
		if strings.Contains(lower, "syscard") || strings.HasPrefix(lower, "scph") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		if primaryExts[ext] {
			candidates = append(candidates, filepath.Join(gameDir, name))
		} else if secondaryExts[ext] {
			secondaryCandidates = append(secondaryCandidates, filepath.Join(gameDir, name))
		}
	}

	if len(candidates) == 0 {
		candidates = secondaryCandidates
	}

	if len(candidates) == 0 {
		return "", false
	}

	if len(candidates) == 1 {
		return candidates[0], true
	}

	regLower := strings.ToLower(string(reg))
	var regKeywords []string
	switch regLower {
	case "jp":
		regKeywords = []string{"(jp)", "japan", "japanese", "_jp", "chi no rondo", "gekka"}
	case "us":
		regKeywords = []string{"(us)", "(usa)", "usa", "english", "en", "_us", "_en"}
	case "eu":
		regKeywords = []string{"(eu)", "(eur)", "europe", "european", "_eu"}
	}

	for _, cand := range candidates {
		candLower := strings.ToLower(filepath.Base(cand))
		for _, kw := range regKeywords {
			if strings.Contains(candLower, kw) {
				return cand, true
			}
		}
	}

	return candidates[0], true
}

// CheckGameAvailability examines the filesystem for ROMs and containerized runner binaries.
func (s *Spawner) CheckGameAvailability(g Game, activeRegion Region) GameStatus {
	status := GameStatus{
		ID:            g.ID,
		ActiveRegion:  activeRegion,
		RegionsStatus: make(map[Region]bool),
	}

	runnerPath, runnerErr := s.resolveRunner(g)
	status.RunnerReady = (runnerErr == nil)
	status.RunnerPath = runnerPath
	if runnerErr != nil {
		effRunner := s.effectiveRunnerType(g)
		status.MissingRunner = fmt.Sprintf("Container '%s' not initialized (click SETTINGS -> INITIALIZE CONTAINERS)", effRunner)
	}

	for reg := range g.Regions {
		_, found := s.ResolveRomPath(g, reg)
		status.RegionsStatus[reg] = found
	}

	if activeRegion == "" {
		activeRegion = g.DefaultRegion
	}

	resolvedPath, romFound := s.ResolveRomPath(g, activeRegion)
	status.Available = romFound && status.RunnerReady
	if !romFound {
		target := g.Regions[activeRegion]
		if target.RomPath != "" {
			status.MissingFile = target.RomPath
		} else {
			status.MissingFile = fmt.Sprintf("games/%s/", g.ID)
		}
	} else {
		status.ResolvedRom = resolvedPath
	}

	return status
}

// Launch initiates the game process inside its isolated container configuration and calls onLaunch/onExit hooks.
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

	absBase, err := filepath.Abs(s.BaseDir)
	if err != nil {
		absBase = s.BaseDir
	}

	// Repair any damaged macOS bundles, sync/provision BIOS files, and inject current Settings into all container configs
	s.RepairMacOSBundles(absBase)
	s.SyncBiosFiles()
	s.EnsureBiosFiles(nil)
	settings := LoadSettings(s.BaseDir)
	_ = ApplySettingsToContainers(s.BaseDir, settings)

	effRunner := s.effectiveRunnerType(g)

	fullRomPath, found := s.ResolveRomPath(g, reg)
	if !found && effRunner != RunnerNative {
		return fmt.Errorf("no ROM file found in games/%s/\nPlease drop any valid ROM (.chd, .iso, .cue, .wad) into that folder", g.ID)
	}
	if fullRomPath != "" && !filepath.IsAbs(fullRomPath) {
		fullRomPath = filepath.Join(absBase, fullRomPath)
	}

	runnerPath, err := s.resolveRunner(g)
	if err != nil {
		return fmt.Errorf("containerized runner '%s' is not yet initialized.\nOpen SETTINGS -> CONTAINERS & ROMS and click 'INITIALIZE CONTAINERS' to set it up automatically.", effRunner)
	}

	var cmd *exec.Cmd
	var postExitSync func()

	switch effRunner {
	case RunnerRetroArch:
		coreFullPath := s.resolveBestCoreForGame(g)
		if coreFullPath == "" {
			return fmt.Errorf("libretro core for %s is not installed in emulators/retroarch/cores/.\nOpen SETTINGS -> CONTAINERS & ROMS and click 'INITIALIZE CONTAINERS'.", g.Title)
		}
		raCfgPath := filepath.Join(absBase, "emulators", "retroarch", "config", "retroarch.cfg")
		args := []string{"--config", raCfgPath, "-L", coreFullPath}
		if settings.DisplayMode == "fullscreen" {
			args = append(args, "-f")
		}
		args = append(args, fullRomPath)
		cmd = exec.Command(runnerPath, args...)

	case RunnerPCSX2:
		pcsx2DataDir := filepath.Join(absBase, "emulators", "pcsx2")
		if runtime.GOOS == "darwin" {
			pcsx2DataDir, postExitSync = s.preparePCSX2RuntimeContainer(absBase)
		}
		args := []string{"-datapath", pcsx2DataDir, "-batch", "-nogui", "-fastboot"}
		if settings.DisplayMode == "fullscreen" {
			args = append(args, "-fullscreen")
		} else {
			args = append(args, "-nofullscreen")
		}
		args = append(args, "--", fullRomPath)
		cmd = exec.Command(runnerPath, args...)

	case RunnerDolphin:
		dolphinUserDir := filepath.Join(absBase, "emulators", "dolphin", "user")
		launchRomPath := fullRomPath
		if runtime.GOOS == "darwin" {
			dolphinUserDir, launchRomPath, postExitSync = s.prepareDolphinRuntimeContainer(absBase, fullRomPath)
		}
		fullscreenVal := "True"
		if settings.DisplayMode == "windowed" {
			fullscreenVal = "False"
		}
		vsyncVal := "True"
		if !settings.VSync {
			vsyncVal = "False"
		}
		fpsVal := "False"
		if settings.ShowFPS {
			fpsVal = "True"
		}
		aspectVal := "2"
		if settings.AspectRatio == "16:9" {
			aspectVal = "1"
		} else if settings.AspectRatio == "stretch" {
			aspectVal = "3"
		}

		args := []string{
			"-u", dolphinUserDir,
			"-C", "Dolphin.Analytics.PermissionAsked=True",
			"-C", "Dolphin.Analytics.Enabled=False",
			"-C", "Dolphin.Interface.ConfirmStop=False",
			"-C", "Dolphin.Interface.UsePanicHandlers=False",
			"-C", "Dolphin.Display.RenderToMain=True",
			"-C", "Dolphin.Display.Fullscreen=" + fullscreenVal,
			"-C", fmt.Sprintf("GFX.Settings.InternalResolution=%d", settings.InternalResolution),
			"-C", "GFX.Settings.AspectRatio=" + aspectVal,
			"-C", "GFX.Settings.ShowFPS=" + fpsVal,
			"-C", "GFX.Hardware.VSync=" + vsyncVal,
			"-b",
			"-e", launchRomPath,
		}
		cmd = exec.Command(runnerPath, args...)

	case RunnerNative:
		if runtime.GOOS == "darwin" && strings.HasSuffix(runnerPath, ".app") {
			cmd = exec.Command("open", "-W", runnerPath)
		} else {
			cmd = exec.Command(runnerPath)
		}

	default:
		return fmt.Errorf("unknown runner type: %s", effRunner)
	}

	cmd.Dir = absBase

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting containerized game process: %w", err)
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

		if postExitSync != nil {
			postExitSync()
		}

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

// prepareDolphinRuntimeContainer mirrors the isolated Dolphin user folder and .wad ROM into a clean
// attribute-free runtime directory outside ~/Desktop so macOS com.apple.macl never blocks NAND writes or WAD boot.
func (s *Spawner) prepareDolphinRuntimeContainer(absBase, fullRomPath string) (string, string, func()) {
	srcUserDir := filepath.Join(absBase, "emulators", "dolphin", "user")
	runtimeRoot := filepath.Join(os.TempDir(), "eclipse-collection-dolphin")
	runtimeUserDir := filepath.Join(runtimeRoot, "user")
	runtimeRomDir := filepath.Join(runtimeRoot, "roms")

	for _, sub := range []string{"Config", "Wii", "StateSaves", "GC", "Cache", "Logs"} {
		_ = os.MkdirAll(filepath.Join(runtimeUserDir, sub), 0755)
		syncDirClean(filepath.Join(srcUserDir, sub), filepath.Join(runtimeUserDir, sub))
	}

	_ = os.MkdirAll(runtimeRomDir, 0755)
	stagedRom := filepath.Join(runtimeRomDir, filepath.Base(fullRomPath))
	if srcInfo, err := os.Stat(fullRomPath); err == nil {
		dstInfo, dstErr := os.Stat(stagedRom)
		if dstErr != nil || dstInfo.Size() != srcInfo.Size() {
			if copyFileIfNewer(fullRomPath, stagedRom) != nil {
				stagedRom = fullRomPath
			}
		}
	} else {
		stagedRom = fullRomPath
	}

	syncBack := func() {
		for _, sub := range []string{"Wii", "StateSaves", "GC"} {
			syncDirClean(filepath.Join(runtimeUserDir, sub), filepath.Join(srcUserDir, sub))
		}
	}

	return runtimeUserDir, stagedRom, syncBack
}

// preparePCSX2RuntimeContainer mirrors the isolated PCSX2 DataRoot into a clean attribute-free runtime
// directory outside ~/Desktop so macOS com.apple.macl never blocks BIOS, INI, or memory card access.
func (s *Spawner) preparePCSX2RuntimeContainer(absBase string) (string, func()) {
	srcPcsx2Dir := filepath.Join(absBase, "emulators", "pcsx2")
	runtimeRoot := filepath.Join(os.TempDir(), "eclipse-collection-pcsx2")
	runtimeSub := filepath.Join(runtimeRoot, "PCSX2")

	for _, targetBase := range []string{runtimeRoot, runtimeSub} {
		for _, sub := range []string{"inis", "bios", "memcards", "sstates", "snaps", "logs", "cheats", "patches", "cache"} {
			_ = os.MkdirAll(filepath.Join(targetBase, sub), 0755)
		}
		for _, sub := range []string{"inis", "bios", "memcards", "sstates"} {
			syncDirClean(filepath.Join(srcPcsx2Dir, sub), filepath.Join(targetBase, sub))
			syncDirClean(filepath.Join(srcPcsx2Dir, "PCSX2", sub), filepath.Join(targetBase, sub))
		}
	}

	syncBack := func() {
		for _, sub := range []string{"memcards", "sstates"} {
			syncDirClean(filepath.Join(runtimeSub, sub), filepath.Join(srcPcsx2Dir, sub))
			syncDirClean(filepath.Join(runtimeSub, sub), filepath.Join(srcPcsx2Dir, "PCSX2", sub))
		}
	}

	return runtimeRoot, syncBack
}

func syncDirClean(srcDir, dstDir string) {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return
	}
	_ = os.MkdirAll(dstDir, 0755)
	for _, e := range entries {
		srcPath := filepath.Join(srcDir, e.Name())
		dstPath := filepath.Join(dstDir, e.Name())
		if e.IsDir() {
			syncDirClean(srcPath, dstPath)
			continue
		}
		_ = copyFileIfNewer(srcPath, dstPath)
	}
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

// resolveRunner prioritizes the isolated container inside emulators/<runner>/
// and validates macOS bundle code signature integrity before returning the binary path.
func (s *Spawner) resolveRunner(g Game) (string, error) {
	effRunner := s.effectiveRunnerType(g)

	if effRunner == RunnerNative {
		nativePath := filepath.Join(s.BaseDir, g.ExecutablePath)
		if runtime.GOOS == "windows" && !strings.HasSuffix(nativePath, ".exe") {
			nativePath += ".exe"
		}
		if _, err := os.Stat(nativePath); err == nil {
			return nativePath, nil
		}
		return "", fmt.Errorf("native executable %s not found", nativePath)
	}

	absBase, err := filepath.Abs(s.BaseDir)
	if err != nil {
		absBase = s.BaseDir
	}

	var candidatePaths []string

	switch effRunner {
	case RunnerRetroArch:
		switch runtime.GOOS {
		case "windows":
			candidatePaths = []string{
				filepath.Join(absBase, "emulators", "retroarch", "retroarch.exe"),
				"C:\\RetroArch-Win64\\retroarch.exe",
			}
		case "darwin":
			candidatePaths = []string{
				filepath.Join(absBase, "emulators", "retroarch", "RetroArch.app", "Contents", "MacOS", "RetroArch"),
				"/Applications/RetroArch.app/Contents/MacOS/RetroArch",
			}
		case "linux":
			candidatePaths = []string{
				filepath.Join(absBase, "emulators", "retroarch", "retroarch"),
				"/usr/bin/retroarch",
				"/usr/local/bin/retroarch",
			}
		}

	case RunnerPCSX2:
		switch runtime.GOOS {
		case "windows":
			candidatePaths = []string{
				filepath.Join(absBase, "emulators", "pcsx2", "pcsx2-qt.exe"),
				filepath.Join(absBase, "emulators", "pcsx2", "pcsx2.exe"),
				"C:\\PCSX2\\pcsx2-qt.exe",
			}
		case "darwin":
			candidatePaths = []string{
				filepath.Join(absBase, "emulators", "pcsx2", "PCSX2-App.app", "Contents", "MacOS", "PCSX2"),
				filepath.Join(absBase, "emulators", "pcsx2", "PCSX2.app", "Contents", "MacOS", "PCSX2"),
				"/Applications/PCSX2.app/Contents/MacOS/PCSX2",
			}
		case "linux":
			candidatePaths = []string{
				filepath.Join(absBase, "emulators", "pcsx2", "pcsx2-qt"),
				"/usr/bin/pcsx2-qt",
			}
		}

	case RunnerDolphin:
		switch runtime.GOOS {
		case "windows":
			candidatePaths = []string{
				filepath.Join(absBase, "emulators", "dolphin", "Dolphin.exe"),
				"C:\\Dolphin\\Dolphin.exe",
			}
		case "darwin":
			candidatePaths = []string{
				filepath.Join(absBase, "emulators", "dolphin", "Dolphin-App.app", "Contents", "MacOS", "Dolphin"),
				filepath.Join(absBase, "emulators", "dolphin", "Dolphin.app", "Contents", "MacOS", "Dolphin"),
				"/Applications/Dolphin.app/Contents/MacOS/Dolphin",
			}
		case "linux":
			candidatePaths = []string{
				filepath.Join(absBase, "emulators", "dolphin", "dolphin-emu"),
				"/usr/bin/dolphin-emu",
			}
		}
	}

	for _, path := range candidatePaths {
		if _, err := os.Stat(path); err == nil {
			// On macOS, verify the binary executes cleanly and is not blocked by AMFI/Gatekeeper
			if runtime.GOOS == "darwin" {
				if effRunner == RunnerPCSX2 && !isPCSX2BinaryHealthy(path) {
					continue
				}
				if effRunner == RunnerDolphin && !isDolphinBinaryHealthy(path) {
					continue
				}
			}
			return path, nil
		}
	}

	return "", fmt.Errorf("containerized runner '%s' not found in emulators/%s/", effRunner, effRunner)
}

// FindCorePath locates a libretro core inside emulators/retroarch/cores/, cores/, or RetroArch.app bundle.
func (s *Spawner) FindCorePath(coreBase string) string {
	if coreBase == "" {
		return ""
	}
	absBase, err := filepath.Abs(s.BaseDir)
	if err != nil {
		absBase = s.BaseDir
	}

	cleanName := filepath.Base(coreBase)
	withExt := s.resolveCoreExt(cleanName)

	candidates := []string{
		filepath.Join(absBase, "emulators", "retroarch", "cores", withExt),
		filepath.Join(absBase, "cores", withExt),
		filepath.Join(absBase, "emulators", "retroarch", "RetroArch.app", "Contents", "Resources", "cores", withExt),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func (s *Spawner) resolveBestCoreForGame(g Game) string {
	absBase, err := filepath.Abs(s.BaseDir)
	if err != nil {
		absBase = s.BaseDir
	}

	if g.ID == "sotn" {
		raSystemDir := filepath.Join(absBase, "emulators", "retroarch", "system")
		_, hasPs1Bios := os.Stat(filepath.Join(raSystemDir, "scph5501.bin"))
		if hasPs1Bios == nil {
			for _, candidate := range []string{"pcsx_rearmed_libretro", "swanstation_libretro", "mednafen_psx_hw_libretro"} {
				if p := s.FindCorePath(candidate); p != "" {
					return p
				}
			}
		} else {
			// PCSX-ReARMed includes a built-in HLE BIOS so it boots SotN even without scph5501.bin
			for _, candidate := range []string{"pcsx_rearmed_libretro", "swanstation_libretro", "mednafen_psx_hw_libretro"} {
				if p := s.FindCorePath(candidate); p != "" {
					return p
				}
			}
		}
	}
	if p := s.FindCorePath(g.CorePath); p != "" {
		return p
	}
	return ""
}

// resolveCoreExt ensures the libretro core has the proper OS extension (.dylib, .dll, .so).
func (s *Spawner) resolveCoreExt(coreBase string) string {
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
