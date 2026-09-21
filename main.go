package main

import (
	"eclipse/engine"
	"fmt"
)

func main() {
	fmt.Println("====================================================")
	fmt.Println(" Castlevania Eclipse Collection Engine (alpha) v0.1 ")
	fmt.Println("====================================================")
	library := map[string]engine.Game{
		"sotn": {
			ID:             "sotn",
			Title:          "Castlevania: Symphony of the Night",
			Type:           "native",
			ExecutablePath: "games/sotn/symphony_recomp",
		},
		"dracula_x": {
			ID:             "dracula_x",
			Title:          "Castlevania: Rondo of Blood",
			Type:           "emulated",
			ExecutablePath: "C:\\RetroArch-Win64\\retroarch",
			CorePath:       "C:\\RetroArch-Win64\\cores\\geargrafx_libretro.dll",
			RomPath:        "C:\\rondo\\Akumajou Dracula X - Chi no Rondo (JP).cue",
		},
	}

	selectedGame := "dracula_x"

	gameToPlay, exists := library[selectedGame]
	if !exists {
		fmt.Println("Game not found in library!")
		return
	}

	engine.Play(gameToPlay)
}
