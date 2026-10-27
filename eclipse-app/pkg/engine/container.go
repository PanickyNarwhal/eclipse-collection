package engine

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ContainerStatus describes the readiness and isolation state of a containerized runner.
type ContainerStatus struct {
	Runner         string   `json:"runner"`
	Name           string   `json:"name"`
	Systems        string   `json:"systems"`
	BinaryReady    bool     `json:"binary_ready"`
	BinaryPath     string   `json:"binary_path"`
	Isolated       bool     `json:"isolated"`
	ConfigPath     string   `json:"config_path"`
	CoresReady     bool     `json:"cores_ready"`
	InstalledCores []string `json:"installed_cores,omitempty"`
	BiosReady      bool     `json:"bios_ready"`
	BiosNote       string   `json:"bios_note,omitempty"`
}

// ScanResult summarizes files discovered during a ROM & BIOS directory scan.
type ScanResult struct {
	ScannedDir   string   `json:"scanned_dir"`
	FoundROMs    []string `json:"found_roms"`
	ImportedBios []string `json:"imported_bios"`
	TotalMatched int      `json:"total_matched"`
}

// EnsureContainerDirs initializes the isolated container directories, repairs any damaged macOS bundle seals,
// strips quarantine attributes, signs libretro cores, and provisions required BIOS files.
func (s *Spawner) EnsureContainerDirs() {
	absBase, err := filepath.Abs(s.BaseDir)
	if err != nil {
		absBase = s.BaseDir
	}

	dirs := []string{
		filepath.Join(absBase, "bios"),
		filepath.Join(absBase, "cores"),
		filepath.Join(absBase, "emulators", "retroarch", "config"),
		filepath.Join(absBase, "emulators", "retroarch", "cores"),
		filepath.Join(absBase, "emulators", "retroarch", "saves"),
		filepath.Join(absBase, "emulators", "retroarch", "states"),
		filepath.Join(absBase, "emulators", "retroarch", "system"),
		filepath.Join(absBase, "emulators", "pcsx2", "inis"),
		filepath.Join(absBase, "emulators", "pcsx2", "bios"),
		filepath.Join(absBase, "emulators", "pcsx2", "memcards"),
		filepath.Join(absBase, "emulators", "pcsx2", "sstates"),
		filepath.Join(absBase, "emulators", "pcsx2", "PCSX2", "inis"),
		filepath.Join(absBase, "emulators", "pcsx2", "PCSX2", "bios"),
		filepath.Join(absBase, "emulators", "pcsx2", "PCSX2", "memcards"),
		filepath.Join(absBase, "emulators", "pcsx2", "PCSX2", "sstates"),
		filepath.Join(absBase, "emulators", "dolphin", "user", "Config"),
		filepath.Join(absBase, "emulators", "dolphin", "user", "Wii"),
	}

	for _, d := range dirs {
		_ = os.MkdirAll(d, 0755)
	}

	// Repair any broken macOS bundle signatures & strip quarantine flags
	s.RepairMacOSBundles(absBase)

	// Sync local BIOS files
	s.SyncBiosFiles()

	// Ensure container config files match current settings
	settings := LoadSettings(s.BaseDir)
	_ = ApplySettingsToContainers(s.BaseDir, settings)

	// Provision any missing BIOS files in the background so Rondo, SotN, LoI, and CoD work out of the box
	go s.EnsureBiosFiles(nil)
}

// RepairMacOSBundles removes any files accidentally placed inside *.app/Contents/Resources/ that break
// macOS _CodeSignature seals, strips com.apple.quarantine via cp -R -X, and ad-hoc signs downloaded libretro cores.
func (s *Spawner) RepairMacOSBundles(absBase string) {
	if runtime.GOOS != "darwin" {
		return
	}

	// 1. Remove portable markers that conflict with isolated -datapath / -u arguments or break _CodeSignature
	_ = os.Remove(filepath.Join(absBase, "emulators", "pcsx2", "portable.ini"))
	_ = os.Remove(filepath.Join(absBase, "emulators", "pcsx2", "PCSX2.app", "Contents", "Resources", "portable.ini"))
	_ = os.Remove(filepath.Join(absBase, "emulators", "pcsx2", "PCSX2-App.app", "Contents", "Resources", "portable.ini"))
	_ = os.Remove(filepath.Join(absBase, "emulators", "dolphin", "portable.txt"))

	// 2. Ensure clean, attribute-free PCSX2-App.app and Dolphin-App.app container bundles exist.
	// Using `cp -R -X` preserves macOS framework symlinks while stripping `com.apple.quarantine` and `com.apple.macl`.
	pcsx2ContainerApp := filepath.Join(absBase, "emulators", "pcsx2", "PCSX2-App.app")
	pcsx2ContainerBin := filepath.Join(pcsx2ContainerApp, "Contents", "MacOS", "PCSX2")
	if !isPCSX2BinaryHealthy(pcsx2ContainerBin) {
		if _, err := os.Stat("/Applications/PCSX2.app/Contents/MacOS/PCSX2"); err == nil {
			_ = os.RemoveAll(pcsx2ContainerApp)
			_ = exec.Command("cp", "-R", "-X", "/Applications/PCSX2.app", pcsx2ContainerApp).Run()
		}
	}

	dolphinContainerApp := filepath.Join(absBase, "emulators", "dolphin", "Dolphin-App.app")
	dolphinContainerBin := filepath.Join(dolphinContainerApp, "Contents", "MacOS", "Dolphin")
	if !isDolphinBinaryHealthy(dolphinContainerBin) {
		if _, err := os.Stat("/Applications/Dolphin.app/Contents/MacOS/Dolphin"); err == nil {
			_ = os.RemoveAll(dolphinContainerApp)
			_ = exec.Command("cp", "-R", "-X", "/Applications/Dolphin.app", dolphinContainerApp).Run()
		}
	}

	raApp := filepath.Join(absBase, "emulators", "retroarch", "RetroArch.app")
	if _, err := os.Stat(raApp); err == nil {
		_ = exec.Command("xattr", "-rd", "com.apple.quarantine", raApp).Run()
	}

	// 3. Strip quarantine and ad-hoc sign all libretro .dylib cores
	coresDirs := []string{
		filepath.Join(absBase, "emulators", "retroarch", "cores"),
		filepath.Join(absBase, "cores"),
	}
	for _, cDir := range coresDirs {
		entries, err := os.ReadDir(cDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".dylib") {
				dylibPath := filepath.Join(cDir, e.Name())
				_ = exec.Command("xattr", "-rd", "com.apple.quarantine", dylibPath).Run()
				_ = exec.Command("codesign", "--force", "--sign", "-", dylibPath).Run()
			}
		}
	}
}

func isPCSX2BinaryHealthy(binPath string) bool {
	if _, err := os.Stat(binPath); err != nil {
		return false
	}
	out, _ := exec.Command(binPath, "-version").CombinedOutput()
	return strings.Contains(string(out), "PCSX2")
}

func isDolphinBinaryHealthy(binPath string) bool {
	if _, err := os.Stat(binPath); err != nil {
		return false
	}
	out, _ := exec.Command(binPath, "--version").CombinedOutput()
	return strings.Contains(string(out), "Dolphin")
}

// EnsureBiosFiles guarantees that required system BIOS files (syscard3.pce for Rondo of Blood,
// scph5501.bin for Symphony of the Night, and SCPH-39001.bin for PS2 titles) are present inside the isolated containers.
func (s *Spawner) EnsureBiosFiles(onProgress func(string)) {
	absBase, err := filepath.Abs(s.BaseDir)
	if err != nil {
		absBase = s.BaseDir
	}

	raSystemDir := filepath.Join(absBase, "emulators", "retroarch", "system")
	pcsx2BiosDir := filepath.Join(absBase, "emulators", "pcsx2", "bios")
	pcsx2SubBiosDir := filepath.Join(absBase, "emulators", "pcsx2", "PCSX2", "bios")
	rootBiosDir := filepath.Join(absBase, "bios")

	for _, d := range []string{raSystemDir, pcsx2BiosDir, pcsx2SubBiosDir, rootBiosDir} {
		_ = os.MkdirAll(d, 0755)
	}

	// 1. PC Engine CD System Card 3.0 (syscard3.pce) - Required by mednafen_pce_fast for Rondo of Blood
	pceBiosDst := filepath.Join(raSystemDir, "syscard3.pce")
	if _, err := os.Stat(pceBiosDst); os.IsNotExist(err) {
		if onProgress != nil {
			onProgress("Provisioning PC Engine CD System Card (syscard3.pce)...")
		}
		pceURL := "https://raw.githubusercontent.com/Abdess/retrobios/main/bios/NEC/PC%20Engine/syscard3.pce"
		if downloadFile(pceURL, pceBiosDst) == nil {
			_ = copyFileIfNewer(pceBiosDst, filepath.Join(rootBiosDir, "syscard3.pce"))
		}
	}

	// 2. PlayStation 1 BIOS (scph5501.bin, scph5500.bin, scph5502.bin) - For SwanStation / Beetle PSX / PCSX-ReARMed
	for _, ps1Bios := range []string{"scph5501.bin", "scph5500.bin", "scph5502.bin"} {
		dst := filepath.Join(raSystemDir, ps1Bios)
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			if onProgress != nil {
				onProgress(fmt.Sprintf("Provisioning PlayStation BIOS (%s)...", ps1Bios))
			}
			url := "https://raw.githubusercontent.com/Abdess/retrobios/main/bios/Sony/PlayStation/" + ps1Bios
			if downloadFile(url, dst) == nil {
				_ = copyFileIfNewer(dst, filepath.Join(rootBiosDir, ps1Bios))
			}
		}
	}

	// 3. PlayStation 2 BIOS (SCPH-39001.bin) - Required by PCSX2 for Lament of Innocence & Curse of Darkness
	if !hasAnyFileWithExt(pcsx2BiosDir, []string{".bin", ".rom"}) {
		if onProgress != nil {
			onProgress("Provisioning PlayStation 2 BIOS (SCPH-39001.bin)...")
		}
		ps2Dst := filepath.Join(pcsx2BiosDir, "SCPH-39001.bin")
		ps2URL := "https://raw.githubusercontent.com/Abdess/retrobios/main/bios/Sony/PlayStation%202/SCPH-39001.bin"
		if downloadFile(ps2URL, ps2Dst) == nil {
			_ = copyFileIfNewer(ps2Dst, filepath.Join(pcsx2SubBiosDir, "SCPH-39001.bin"))
			_ = copyFileIfNewer(ps2Dst, filepath.Join(rootBiosDir, "SCPH-39001.bin"))
		}
	} else {
		// Mirror any existing BIOS in pcsx2/bios into pcsx2/PCSX2/bios
		entries, _ := os.ReadDir(pcsx2BiosDir)
		for _, e := range entries {
			if !e.IsDir() {
				src := filepath.Join(pcsx2BiosDir, e.Name())
				_ = copyFileIfNewer(src, filepath.Join(pcsx2SubBiosDir, e.Name()))
			}
		}
	}
}

// SyncBiosFiles copies any BIOS files placed in bios/, games/<id>/, or existing user folders into the isolated containers.
func (s *Spawner) SyncBiosFiles() []string {
	absBase, err := filepath.Abs(s.BaseDir)
	if err != nil {
		absBase = s.BaseDir
	}

	raSystemDir := filepath.Join(absBase, "emulators", "retroarch", "system")
	pcsx2BiosDir := filepath.Join(absBase, "emulators", "pcsx2", "bios")
	pcsx2SubBiosDir := filepath.Join(absBase, "emulators", "pcsx2", "PCSX2", "bios")
	_ = os.MkdirAll(raSystemDir, 0755)
	_ = os.MkdirAll(pcsx2BiosDir, 0755)
	_ = os.MkdirAll(pcsx2SubBiosDir, 0755)

	homeDir, _ := os.UserHomeDir()
	searchDirs := []string{
		filepath.Join(absBase, "bios"),
		filepath.Join(absBase, "games", "rondo"),
		filepath.Join(absBase, "games", "sotn"),
		filepath.Join(absBase, "games", "loi"),
		filepath.Join(absBase, "games", "cod"),
	}
	if homeDir != "" {
		searchDirs = append(searchDirs,
			filepath.Join(homeDir, "Library", "Application Support", "RetroArch", "system"),
			filepath.Join(homeDir, "Documents", "RetroArch", "system"),
			filepath.Join(homeDir, "Library", "Application Support", "PCSX2", "bios"),
		)
	}

	var synced []string
	for _, dir := range searchDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			lower := strings.ToLower(name)
			src := filepath.Join(dir, name)

			// PC Engine CD BIOS (syscard3.pce)
			if strings.Contains(lower, "syscard") && strings.HasSuffix(lower, ".pce") {
				dst := filepath.Join(raSystemDir, "syscard3.pce")
				if copyFileIfNewer(src, dst) == nil {
					synced = append(synced, "syscard3.pce (PCE-CD BIOS)")
				}
			}

			// PS1 / PS2 BIOS files (scph*.bin, ps2*.bin, bios*.bin, rom1.bin)
			if strings.HasPrefix(lower, "scph") || strings.Contains(lower, "bios") || strings.HasPrefix(lower, "ps2") {
				ext := filepath.Ext(lower)
				if ext == ".bin" || ext == ".rom" || ext == ".mec" || ext == ".nvm" {
					info, err := entry.Info()
					if err == nil {
						if info.Size() <= 1024*1024 {
							_ = copyFileIfNewer(src, filepath.Join(raSystemDir, strings.ToLower(name)))
							synced = append(synced, name+" (PS1 BIOS)")
						} else {
							_ = copyFileIfNewer(src, filepath.Join(pcsx2BiosDir, name))
							_ = copyFileIfNewer(src, filepath.Join(pcsx2SubBiosDir, name))
							synced = append(synced, name+" (PS2 BIOS)")
						}
					}
				}
			}
		}
	}
	return synced
}

// GetContainersStatus inspects all 3 containerized emulator environments.
func (s *Spawner) GetContainersStatus() []ContainerStatus {
	absBase, err := filepath.Abs(s.BaseDir)
	if err != nil {
		absBase = s.BaseDir
	}

	// 1. RetroArch Container
	raDummy := Game{RunnerType: RunnerRetroArch}
	raPath, raErr := s.resolveRunner(raDummy)
	var installedCores []string
	for _, coreName := range []string{"mednafen_pce_fast_libretro", "swanstation_libretro", "pcsx_rearmed_libretro", "mednafen_psx_hw_libretro"} {
		if fullCore := s.FindCorePath(coreName); fullCore != "" {
			installedCores = append(installedCores, coreName)
		}
	}
	raSystemDir := filepath.Join(absBase, "emulators", "retroarch", "system")
	_, pceBiosErr := os.Stat(filepath.Join(raSystemDir, "syscard3.pce"))
	raBiosNote := "syscard3.pce & PS1 BIOS ready in emulators/retroarch/system/"
	if pceBiosErr != nil {
		raBiosNote = "Auto-provisioning syscard3.pce & PS1 BIOS..."
	}

	raStatus := ContainerStatus{
		Runner:         string(RunnerRetroArch),
		Name:           "RetroArch Core Container",
		Systems:        "PC Engine CD (Rondo) & PlayStation (SotN)",
		BinaryReady:    raErr == nil,
		BinaryPath:     raPath,
		Isolated:       true,
		ConfigPath:     "emulators/retroarch/config/retroarch.cfg",
		CoresReady:     len(installedCores) >= 2,
		InstalledCores: installedCores,
		BiosReady:      pceBiosErr == nil,
		BiosNote:       raBiosNote,
	}

	// 2. PCSX2 Container
	pcsx2Dummy := Game{RunnerType: RunnerPCSX2}
	pcsx2Path, pcsx2Err := s.resolveRunner(pcsx2Dummy)
	pcsx2BiosDir := filepath.Join(absBase, "emulators", "pcsx2", "bios")
	pcsx2HasBios := hasAnyFileWithExt(pcsx2BiosDir, []string{".bin", ".rom"})
	pcsx2BiosNote := "PS2 BIOS & isolated DataRoot ready in emulators/pcsx2/"
	if !pcsx2HasBios {
		pcsx2BiosNote = "Auto-provisioning PS2 BIOS..."
	}

	pcsx2Status := ContainerStatus{
		Runner:      string(RunnerPCSX2),
		Name:        "PCSX2 Portable Container",
		Systems:     "PlayStation 2 (Lament of Innocence & Curse of Darkness)",
		BinaryReady: pcsx2Err == nil,
		BinaryPath:  pcsx2Path,
		Isolated:    true,
		ConfigPath:  "emulators/pcsx2/PCSX2/inis/PCSX2.ini",
		CoresReady:  true,
		BiosReady:   pcsx2HasBios,
		BiosNote:    pcsx2BiosNote,
	}

	// 3. Dolphin Container
	dolphinDummy := Game{RunnerType: RunnerDolphin}
	dolphinPath, dolphinErr := s.resolveRunner(dolphinDummy)
	dolphinStatus := ContainerStatus{
		Runner:      string(RunnerDolphin),
		Name:        "Dolphin Portable Container",
		Systems:     "WiiWare (The Adventure ReBirth)",
		BinaryReady: dolphinErr == nil,
		BinaryPath:  dolphinPath,
		Isolated:    true,
		ConfigPath:  "emulators/dolphin/user/Config/GFX.ini",
		CoresReady:  true,
		BiosReady:   true,
		BiosNote:    "Isolated Wii NAND & configs in emulators/dolphin/user/",
	}

	return []ContainerStatus{raStatus, pcsx2Status, dolphinStatus}
}

// ProvisionAllContainers downloads/sandboxes RetroArch (+ cores), PCSX2, Dolphin, and BIOS files into emulators/.
func (s *Spawner) ProvisionAllContainers(onProgress func(string)) error {
	absBase, err := filepath.Abs(s.BaseDir)
	if err != nil {
		absBase = s.BaseDir
	}

	s.EnsureContainerDirs()

	runners := []RunnerType{RunnerRetroArch, RunnerPCSX2, RunnerDolphin}
	for _, r := range runners {
		if err := s.ProvisionContainer(r, onProgress); err != nil {
			return fmt.Errorf("%s setup error: %w", r, err)
		}
	}

	s.EnsureBiosFiles(onProgress)
	s.RepairMacOSBundles(absBase)

	settings := LoadSettings(s.BaseDir)
	_ = ApplySettingsToContainers(s.BaseDir, settings)
	return nil
}

// ProvisionContainer downloads and extracts a portable, isolated emulator into emulators/<runner>/.
func (s *Spawner) ProvisionContainer(runner RunnerType, onProgress func(string)) error {
	absBase, err := filepath.Abs(s.BaseDir)
	if err != nil {
		absBase = s.BaseDir
	}

	cacheDir := filepath.Join(absBase, "emulators", ".cache")
	_ = os.MkdirAll(cacheDir, 0755)

	switch runner {
	case RunnerRetroArch:
		if err := s.provisionRetroArch(absBase, cacheDir, onProgress); err != nil {
			return err
		}
		return s.provisionRetroArchCores(absBase, cacheDir, onProgress)

	case RunnerPCSX2:
		return s.provisionPCSX2(absBase, cacheDir, onProgress)

	case RunnerDolphin:
		return s.provisionDolphin(absBase, cacheDir, onProgress)
	}
	return nil
}

func (s *Spawner) provisionRetroArch(absBase, cacheDir string, onProgress func(string)) error {
	targetApp := filepath.Join(absBase, "emulators", "retroarch", "RetroArch.app")
	targetBin := filepath.Join(targetApp, "Contents", "MacOS", "RetroArch")

	if runtime.GOOS == "darwin" {
		if _, err := os.Stat(targetBin); err == nil {
			_ = exec.Command("xattr", "-rd", "com.apple.quarantine", targetApp).Run()
			if onProgress != nil {
				onProgress("RetroArch container binary verified.")
			}
			return nil
		}

		if _, err := os.Stat("/Applications/RetroArch.app/Contents/MacOS/RetroArch"); err == nil {
			if onProgress != nil {
				onProgress("Cloning system RetroArch.app into isolated container...")
			}
			_ = os.RemoveAll(targetApp)
			_ = exec.Command("ditto", "/Applications/RetroArch.app", targetApp).Run()
			_ = exec.Command("xattr", "-rd", "com.apple.quarantine", targetApp).Run()
			return nil
		}

		dlURL := resolveBrewCaskURL("retroarch-metal", "https://buildbot.libretro.com/stable/1.22.2/apple/osx/universal/RetroArch_Metal.dmg")
		dmgPath := filepath.Join(cacheDir, "RetroArch_Metal.dmg")
		if onProgress != nil {
			onProgress("Downloading portable RetroArch container (macOS Universal)...")
		}
		if err := downloadFile(dlURL, dmgPath); err != nil {
			return fmt.Errorf("failed to download RetroArch: %w", err)
		}

		if onProgress != nil {
			onProgress("Extracting RetroArch into isolated emulators/retroarch/...")
		}
		if err := extractAppFromDMG(dmgPath, "RetroArch.app", targetApp); err != nil {
			return err
		}
		_ = exec.Command("xattr", "-rd", "com.apple.quarantine", targetApp).Run()
	} else if runtime.GOOS == "windows" {
		raDir := filepath.Join(absBase, "emulators", "retroarch")
		winBin := filepath.Join(raDir, "retroarch.exe")
		if _, err := os.Stat(winBin); err == nil {
			if onProgress != nil {
				onProgress("RetroArch Windows container binary verified.")
			}
			return nil
		}
		archivePath := filepath.Join(cacheDir, "RetroArch-Win64.7z")
		dlURL := "https://buildbot.libretro.com/stable/1.22.2/windows/x86_64/RetroArch.7z"
		if onProgress != nil {
			onProgress("Downloading portable RetroArch container (Windows x64)...")
		}
		if err := downloadFile(dlURL, archivePath); err != nil {
			return fmt.Errorf("failed to download Windows RetroArch: %w", err)
		}
		if onProgress != nil {
			onProgress("Extracting RetroArch into emulators/retroarch/...")
		}
		extractTemp := filepath.Join(cacheDir, "ra_win_extract")
		_ = os.RemoveAll(extractTemp)
		_ = os.MkdirAll(extractTemp, 0755)
		if err := exec.Command("tar.exe", "-xf", archivePath, "-C", extractTemp).Run(); err != nil {
			return fmt.Errorf("failed to extract RetroArch archive: %w", err)
		}
		// Move extracted contents (handling optional RetroArch-Win64/ subfolder)
		moveExtractedDir(extractTemp, raDir)
	}
	return nil
}

func (s *Spawner) provisionRetroArchCores(absBase, cacheDir string, onProgress func(string)) error {
	coresDir := filepath.Join(absBase, "emulators", "retroarch", "cores")
	rootCoresDir := filepath.Join(absBase, "cores")
	_ = os.MkdirAll(coresDir, 0755)
	_ = os.MkdirAll(rootCoresDir, 0755)

	arch := "x86_64"
	if runtime.GOARCH == "arm64" {
		arch = "arm64"
	}

	ext := ".dylib"
	osPath := "apple/osx/" + arch
	if runtime.GOOS == "windows" {
		ext = ".dll"
		osPath = "windows/x86_64"
	} else if runtime.GOOS == "linux" {
		ext = ".so"
		osPath = "linux/x86_64"
	}

	coreList := []string{
		"mednafen_pce_fast_libretro",
		"swanstation_libretro",
		"pcsx_rearmed_libretro",
		"mednafen_psx_hw_libretro",
	}

	for _, core := range coreList {
		coreFilename := core + ext
		dstCore := filepath.Join(coresDir, coreFilename)
		if _, err := os.Stat(dstCore); err == nil {
			if runtime.GOOS == "darwin" {
				_ = exec.Command("xattr", "-rd", "com.apple.quarantine", dstCore).Run()
				_ = exec.Command("codesign", "--force", "--sign", "-", dstCore).Run()
			}
			_ = copyFileIfNewer(dstCore, filepath.Join(rootCoresDir, coreFilename))
			continue
		}

		if onProgress != nil {
			onProgress(fmt.Sprintf("Downloading core: %s...", core))
		}

		zipURL := fmt.Sprintf("https://buildbot.libretro.com/nightly/%s/latest/%s.zip", osPath, coreFilename)
		zipPath := filepath.Join(cacheDir, coreFilename+".zip")

		if err := downloadFile(zipURL, zipPath); err != nil {
			continue
		}

		if err := unzipSingleFile(zipPath, dstCore); err == nil {
			_ = os.Chmod(dstCore, 0755)
			if runtime.GOOS == "darwin" {
				_ = exec.Command("xattr", "-rd", "com.apple.quarantine", dstCore).Run()
				_ = exec.Command("codesign", "--force", "--sign", "-", dstCore).Run()
			}
			_ = copyFileIfNewer(dstCore, filepath.Join(rootCoresDir, coreFilename))
		}
	}

	return nil
}

func (s *Spawner) provisionPCSX2(absBase, cacheDir string, onProgress func(string)) error {
	pcsx2Dir := filepath.Join(absBase, "emulators", "pcsx2")
	targetApp := filepath.Join(pcsx2Dir, "PCSX2-App.app")
	targetBin := filepath.Join(targetApp, "Contents", "MacOS", "PCSX2")

	if runtime.GOOS == "darwin" {
		_ = os.Remove(filepath.Join(pcsx2Dir, "portable.ini"))
		_ = os.Remove(filepath.Join(targetApp, "Contents", "Resources", "portable.ini"))

		if isPCSX2BinaryHealthy(targetBin) {
			if onProgress != nil {
				onProgress("PCSX2 container binary verified.")
			}
			return nil
		}

		if _, err := os.Stat("/Applications/PCSX2.app/Contents/MacOS/PCSX2"); err == nil {
			if onProgress != nil {
				onProgress("Cloning clean PCSX2.app into isolated container...")
			}
			_ = os.RemoveAll(targetApp)
			_ = exec.Command("cp", "-R", "-X", "/Applications/PCSX2.app", targetApp).Run()
			return nil
		}

		dlURL := resolveBrewCaskURL("pcsx2", "https://github.com/PCSX2/pcsx2/releases/download/v2.8.2/pcsx2-v2.8.2-macos-Qt.tar.xz")
		archivePath := filepath.Join(cacheDir, "pcsx2-macos.tar.xz")
		if onProgress != nil {
			onProgress("Downloading portable PCSX2 container (macOS)...")
		}
		if err := downloadFile(dlURL, archivePath); err != nil {
			return fmt.Errorf("failed to download PCSX2: %w", err)
		}

		if onProgress != nil {
			onProgress("Extracting PCSX2 into isolated emulators/pcsx2/...")
		}
		extractTemp := filepath.Join(cacheDir, "pcsx2_extract")
		_ = os.RemoveAll(extractTemp)
		_ = os.MkdirAll(extractTemp, 0755)

		if err := exec.Command("tar", "-xf", archivePath, "-C", extractTemp).Run(); err != nil {
			return fmt.Errorf("failed to extract PCSX2 archive: %w", err)
		}

		entries, _ := os.ReadDir(extractTemp)
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".app") {
				_ = os.RemoveAll(targetApp)
				_ = exec.Command("cp", "-R", "-X", filepath.Join(extractTemp, e.Name()), targetApp).Run()
				break
			}
		}
	} else if runtime.GOOS == "windows" {
		winBin := filepath.Join(pcsx2Dir, "pcsx2-qt.exe")
		if _, err := os.Stat(winBin); err == nil {
			if onProgress != nil {
				onProgress("PCSX2 Windows container binary verified.")
			}
			return nil
		}
		archivePath := filepath.Join(cacheDir, "pcsx2-windows-x64-Qt.7z")
		dlURL := "https://github.com/PCSX2/pcsx2/releases/download/v2.8.2/pcsx2-v2.8.2-windows-x64-Qt.7z"
		if onProgress != nil {
			onProgress("Downloading portable PCSX2 container (Windows x64)...")
		}
		if err := downloadFile(dlURL, archivePath); err != nil {
			return fmt.Errorf("failed to download Windows PCSX2: %w", err)
		}
		if onProgress != nil {
			onProgress("Extracting PCSX2 into emulators/pcsx2/...")
		}
		extractTemp := filepath.Join(cacheDir, "pcsx2_win_extract")
		_ = os.RemoveAll(extractTemp)
		_ = os.MkdirAll(extractTemp, 0755)
		if err := exec.Command("tar.exe", "-xf", archivePath, "-C", extractTemp).Run(); err != nil {
			return fmt.Errorf("failed to extract PCSX2 archive: %w", err)
		}
		moveExtractedDir(extractTemp, pcsx2Dir)
	}
	return nil
}

func (s *Spawner) provisionDolphin(absBase, cacheDir string, onProgress func(string)) error {
	dolphinDir := filepath.Join(absBase, "emulators", "dolphin")
	targetApp := filepath.Join(dolphinDir, "Dolphin-App.app")
	targetBin := filepath.Join(targetApp, "Contents", "MacOS", "Dolphin")

	if runtime.GOOS == "darwin" {
		_ = os.Remove(filepath.Join(dolphinDir, "portable.txt"))

		if isDolphinBinaryHealthy(targetBin) {
			if onProgress != nil {
				onProgress("Dolphin container binary verified.")
			}
			return nil
		}

		if _, err := os.Stat("/Applications/Dolphin.app/Contents/MacOS/Dolphin"); err == nil {
			if onProgress != nil {
				onProgress("Cloning system Dolphin.app into isolated container...")
			}
			_ = os.RemoveAll(targetApp)
			_ = exec.Command("cp", "-R", "-X", "/Applications/Dolphin.app", targetApp).Run()
			return nil
		}

		dlURL := resolveBrewCaskURL("dolphin", "https://dl.dolphin-emu.org/releases/2609a/dolphin-2609a-universal.dmg")
		dmgPath := filepath.Join(cacheDir, "Dolphin.dmg")
		if onProgress != nil {
			onProgress("Downloading portable Dolphin container (macOS Universal)...")
		}
		if err := downloadFile(dlURL, dmgPath); err != nil {
			return fmt.Errorf("failed to download Dolphin: %w", err)
		}

		if onProgress != nil {
			onProgress("Extracting Dolphin into isolated emulators/dolphin/...")
		}
		if err := extractAppFromDMG(dmgPath, "Dolphin.app", targetApp); err != nil {
			return err
		}
	} else if runtime.GOOS == "windows" {
		winBin := filepath.Join(dolphinDir, "Dolphin.exe")
		if _, err := os.Stat(winBin); err == nil {
			if onProgress != nil {
				onProgress("Dolphin Windows container binary verified.")
			}
			return nil
		}
		archivePath := filepath.Join(cacheDir, "dolphin-x64.7z")
		dlURL := "https://dl.dolphin-emu.org/releases/2609a/dolphin-2609a-x64.7z"
		if onProgress != nil {
			onProgress("Downloading portable Dolphin container (Windows x64)...")
		}
		if err := downloadFile(dlURL, archivePath); err != nil {
			return fmt.Errorf("failed to download Windows Dolphin: %w", err)
		}
		if onProgress != nil {
			onProgress("Extracting Dolphin into emulators/dolphin/...")
		}
		extractTemp := filepath.Join(cacheDir, "dolphin_win_extract")
		_ = os.RemoveAll(extractTemp)
		_ = os.MkdirAll(extractTemp, 0755)
		if err := exec.Command("tar.exe", "-xf", archivePath, "-C", extractTemp).Run(); err != nil {
			return fmt.Errorf("failed to extract Dolphin archive: %w", err)
		}
		moveExtractedDir(extractTemp, dolphinDir)
		_ = os.WriteFile(filepath.Join(dolphinDir, "portable.txt"), []byte(""), 0644)
	}
	return nil
}

func moveExtractedDir(extractTemp, dstDir string) {
	entries, err := os.ReadDir(extractTemp)
	if err != nil {
		return
	}
	srcRoot := extractTemp
	if len(entries) == 1 && entries[0].IsDir() {
		srcRoot = filepath.Join(extractTemp, entries[0].Name())
	}
	syncDirClean(srcRoot, dstDir)
}

// ScanAndImportDirectory scans a folder for ROMs and BIOS files, importing or matching them to the library.
func (s *Spawner) ScanAndImportDirectory(scanDir string) (ScanResult, error) {
	absBase, err := filepath.Abs(s.BaseDir)
	if err != nil {
		absBase = s.BaseDir
	}

	if scanDir == "" {
		scanDir = filepath.Join(absBase, "games")
	}

	res := ScanResult{
		ScannedDir:   scanDir,
		FoundROMs:    []string{},
		ImportedBios: []string{},
	}

	raSystemDir := filepath.Join(absBase, "emulators", "retroarch", "system")
	pcsx2BiosDir := filepath.Join(absBase, "emulators", "pcsx2", "bios")
	pcsx2SubBiosDir := filepath.Join(absBase, "emulators", "pcsx2", "PCSX2", "bios")
	_ = os.MkdirAll(raSystemDir, 0755)
	_ = os.MkdirAll(pcsx2BiosDir, 0755)
	_ = os.MkdirAll(pcsx2SubBiosDir, 0755)

	validExts := map[string]bool{
		".chd": true, ".cue": true, ".iso": true, ".bin": true,
		".img": true, ".wad": true, ".cso": true, ".pce": true, ".wbfs": true,
	}

	_ = filepath.WalkDir(scanDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := d.Name()
		if strings.HasPrefix(name, ".") {
			return nil
		}
		lower := strings.ToLower(name)
		ext := strings.ToLower(filepath.Ext(name))

		if strings.Contains(lower, "syscard") && ext == ".pce" {
			dst := filepath.Join(raSystemDir, "syscard3.pce")
			if copyFileIfNewer(path, dst) == nil {
				res.ImportedBios = append(res.ImportedBios, "syscard3.pce")
			}
			return nil
		}
		if strings.HasPrefix(lower, "scph") || strings.Contains(lower, "bios") {
			if ext == ".bin" || ext == ".rom" {
				if info, err := d.Info(); err == nil {
					if info.Size() <= 1024*1024 {
						_ = copyFileIfNewer(path, filepath.Join(raSystemDir, lower))
						res.ImportedBios = append(res.ImportedBios, name+" (PS1)")
					} else {
						_ = copyFileIfNewer(path, filepath.Join(pcsx2BiosDir, name))
						_ = copyFileIfNewer(path, filepath.Join(pcsx2SubBiosDir, name))
						res.ImportedBios = append(res.ImportedBios, name+" (PS2)")
					}
				}
				return nil
			}
		}

		if !validExts[ext] {
			return nil
		}

		matchedGame := matchGameByFilename(lower, ext)
		if matchedGame != "" {
			targetGameDir := filepath.Join(absBase, "games", matchedGame)
			_ = os.MkdirAll(targetGameDir, 0755)
			dstPath := filepath.Join(targetGameDir, name)
			if filepath.Clean(path) != filepath.Clean(dstPath) {
				if _, err := os.Stat(dstPath); os.IsNotExist(err) {
					if os.Symlink(path, dstPath) != nil {
						_ = copyFileIfNewer(path, dstPath)
					}
				}
			}
			res.FoundROMs = append(res.FoundROMs, fmt.Sprintf("%s -> %s", name, matchedGame))
		} else if strings.Contains(filepath.ToSlash(path), "/games/") {
			res.FoundROMs = append(res.FoundROMs, name)
		}

		return nil
	})

	biosSynced := s.SyncBiosFiles()
	for _, b := range biosSynced {
		res.ImportedBios = append(res.ImportedBios, b)
	}

	res.TotalMatched = len(res.FoundROMs)
	return res, nil
}

func matchGameByFilename(lowerName, ext string) string {
	if ext == ".wad" || strings.Contains(lowerName, "rebirth") {
		return "rebirth"
	}
	if strings.Contains(lowerName, "rondo") || strings.Contains(lowerName, "dracula x") {
		return "rondo"
	}
	if strings.Contains(lowerName, "symphony") || strings.Contains(lowerName, "sotn") || strings.Contains(lowerName, "gekka") {
		return "sotn"
	}
	if strings.Contains(lowerName, "lament") || strings.Contains(lowerName, "innocence") {
		return "loi"
	}
	if strings.Contains(lowerName, "curse") || strings.Contains(lowerName, "darkness") || strings.Contains(lowerName, "yami no juin") {
		return "cod"
	}
	return ""
}

func resolveBrewCaskURL(caskToken, fallback string) string {
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get("https://formulae.brew.sh/api/cask/" + caskToken + ".json")
	if err != nil {
		return fallback
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fallback
	}
	var payload struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil && payload.URL != "" {
		return payload.URL
	}
	return fallback
}

func downloadFile(url, dst string) error {
	if info, err := os.Stat(dst); err == nil && info.Size() > 1024*64 {
		return nil
	}
	tmpDst := dst + ".part"
	out, err := os.Create(tmpDst)
	if err != nil {
		return err
	}
	defer out.Close()

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	out.Close()
	return os.Rename(tmpDst, dst)
}

func extractAppFromDMG(dmgPath, appName, targetAppPath string) error {
	mountPoint := dmgPath + "_mnt"
	_ = os.MkdirAll(mountPoint, 0755)
	defer func() {
		_ = exec.Command("hdiutil", "detach", mountPoint, "-force").Run()
		_ = os.RemoveAll(mountPoint)
	}()

	cmd := exec.Command("hdiutil", "attach", dmgPath, "-nobrowse", "-readonly", "-mountpoint", mountPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mount %s: %w", filepath.Base(dmgPath), err)
	}

	srcApp := filepath.Join(mountPoint, appName)
	if _, err := os.Stat(srcApp); err != nil {
		entries, _ := os.ReadDir(mountPoint)
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".app") {
				srcApp = filepath.Join(mountPoint, e.Name())
				break
			}
		}
	}

	_ = os.RemoveAll(targetAppPath)
	if err := exec.Command("cp", "-R", "-X", srcApp, targetAppPath).Run(); err != nil {
		return fmt.Errorf("failed to copy %s from DMG: %w", appName, err)
	}
	return nil
}

func unzipSingleFile(zipPath, dstFile string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		out, err := os.Create(dstFile)
		if err != nil {
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		return err
	}
	return fmt.Errorf("empty zip archive")
}

func copyFileIfNewer(src, dst string) error {
	if filepath.Clean(src) == filepath.Clean(dst) {
		return nil
	}
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

func hasAnyFileWithExt(dir string, exts []string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		lower := strings.ToLower(e.Name())
		for _, ext := range exts {
			if strings.HasSuffix(lower, ext) {
				return true
			}
		}
	}
	return false
}
