package engine

import (
	"fmt"
	"os/exec"
	"runtime"
)

func LaunchNative(eclipse string) {
	fmt.Printf("Attempting to launch %s on %s...\n", eclipse, runtime.GOOS)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command(eclipse + ".exe")
	case "darwin":
		cmd = exec.Command("open", eclipse+".app")
	case "linux":
		cmd = exec.Command(eclipse)
	default:
		fmt.Printf("Unsupported OS: %s\n", runtime.GOOS)
		return
	}
	err := cmd.Start()

	if err != nil {
		fmt.Printf("Error occurred while launching %s: %v\n", eclipse, err)
	}

	fmt.Printf("Game launched successfully.")
}

type Game struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Type           string `json:"type"`            // "native" or "emulated"
	ExecutablePath string `json:"executable_path"` // e.g., "symphony_recomp" or "retroarch"
	CorePath       string `json:"core_path"`       // e.g., "mednafen_pce_fast_libretro.dll"
	RomPath        string `json:"rom_path"`        // e.g., "rondo_of_blood.cue"
}

func Play(g Game) {
	fmt.Printf("Starting %s...\n", g.Title)

	switch g.Type {
	case "native":
		launchNative(g.ExecutablePath)
	case "emulated":
		launchEmulator(g.ExecutablePath, g.CorePath, g.RomPath)
	default:
		fmt.Printf("Error: Unknown game type '%s'\n", g.Type)
	}
}

func launchNative(executableName string) {
	fmt.Printf("➔ Routing native binary on %s...\n", runtime.GOOS)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command(executableName + ".exe")
	case "darwin":
		cmd = exec.Command("open", "-a", executableName)
	case "linux":
		cmd = exec.Command("./" + executableName)
	default:
		fmt.Println("Unsupported operating system!")
		return
	}

	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error launching native game: %v\n", err)
		return
	}
	fmt.Println("Game is running!")
}

func launchEmulator(emulatorExe string, corePath string, romPath string) {
	fmt.Printf("➔ Routing to RetroArch on %s...\n", runtime.GOOS)

	var cmd *exec.Cmd

	// RetroArch command line format: retroarch -L "core_path" "rom_path"
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command(emulatorExe+".exe", "-L", corePath, romPath)
	case "darwin":
		// macOS usually requires the full path to the binary inside the .app bundle
		cmd = exec.Command(emulatorExe, "-L", corePath, romPath)
	case "linux":
		cmd = exec.Command(emulatorExe, "-L", corePath, romPath)
	default:
		fmt.Println("Unsupported operating system!")
		return
	}

	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error launching emulator: %v\n", err)
		return
	}
	fmt.Println("Emulator is running!")
}
