package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultLibrary creates the official 7-game lineup configuration.
func DefaultLibrary() map[string]Game {
	return map[string]Game{
		"sotn": {
			ID:             "sotn",
			Title:          "Castlevania: Symphony of the Night",
			OriginalSystem: "PlayStation / PC Recomp",
			ReleaseYear:    "1997",
			Developer:      "Konami Computer Entertainment Tokyo",
			Synopsis:       "Dracula's castle rises once more, but Richter Belmont has vanished. Awoken from his eternal slumber, Alucard enters the inverted realm to end his father's curse.",
			RunnerType:     RunnerNative,
			ExecutablePath: "games/sotn/symphony_recomp",
			CorePath:       "cores/mednafen_psx_hw_libretro",
			SupportedRegions: []Region{RegionUS, RegionJP, RegionEU},
			DefaultRegion:    RegionUS,
			ThemeTrack:       "assets/audio/sotn_theme.mp3",
			Regions: map[Region]RegionTarget{
				RegionUS: {
					Title:   "Castlevania: Symphony of the Night (US)",
					RomPath: "games/sotn/sotn_us.cue",
				},
				RegionJP: {
					Title:   "Akumajou Dracula X: Gekka no Yasoukyoku (JP)",
					RomPath: "games/sotn/sotn_jp.cue",
				},
				RegionEU: {
					Title:   "Castlevania: Symphony of the Night (EU)",
					RomPath: "games/sotn/sotn_eu.cue",
				},
			},
		},
		"rondo": {
			ID:             "rondo",
			Title:          "Castlevania: Rondo of Blood",
			OriginalSystem: "PC Engine Super CD-ROM²",
			ReleaseYear:    "1993",
			Developer:      "Konami",
			Synopsis:       "Richter Belmont sets out with the legendary whip Vampire Killer to rescue his beloved Annette and young Maria Renard from the resurrected Lord of Darkness.",
			RunnerType:     RunnerRetroArch,
			CorePath:       "cores/mednafen_pce_fast_libretro",
			SupportedRegions: []Region{RegionUS, RegionJP},
			DefaultRegion:    RegionUS,
			ThemeTrack:       "assets/audio/rondo_theme.mp3",
			Regions: map[Region]RegionTarget{
				RegionUS: {
					Title:     "Castlevania: Rondo of Blood (English Patched)",
					RomPath:   "games/rondo/rondo_en.chd",
					PatchPath: "patches/rondo_english.bps",
				},
				RegionJP: {
					Title:   "Akumajou Dracula X: Chi no Rondo (Original JP)",
					RomPath: "games/rondo/rondo_jp.chd",
				},
			},
		},
		"loi": {
			ID:             "loi",
			Title:          "Castlevania: Lament of Innocence",
			OriginalSystem: "PlayStation 2",
			ReleaseYear:    "2003",
			Developer:      "Konami Computer Entertainment Tokyo",
			Synopsis:       "In the 11th century, knight Leon Belmont renounces his title to enter the forest of eternal night and save his betrothed Sara from the vampire Walter Bernhard—unveiling the origin of the whip and the Belmont clan's eternal crusade.",
			RunnerType:     RunnerPCSX2,
			SupportedRegions: []Region{RegionUS, RegionJP, RegionEU},
			DefaultRegion:    RegionUS,
			ThemeTrack:       "assets/audio/loi_theme.mp3",
			Regions: map[Region]RegionTarget{
				RegionUS: {
					Title:   "Castlevania: Lament of Innocence (US)",
					RomPath: "games/loi/loi_us.iso",
				},
				RegionJP: {
					Title:   "Castlevania (JP)",
					RomPath: "games/loi/loi_jp.iso",
				},
				RegionEU: {
					Title:   "Castlevania (EU)",
					RomPath: "games/loi/loi_eu.iso",
				},
			},
		},
		"cod": {
			ID:             "cod",
			Title:          "Castlevania: Curse of Darkness",
			OriginalSystem: "PlayStation 2",
			ReleaseYear:    "2005",
			Developer:      "Konami Digital Entertainment",
			Synopsis:       "Three years after Trevor Belmont defeated Dracula, Hector, a former Devil Forgemaster seeking redemption, hunts his former comrade Isaac across a curse-ravaged Europe.",
			RunnerType:     RunnerPCSX2,
			SupportedRegions: []Region{RegionUS, RegionJP, RegionEU},
			DefaultRegion:    RegionUS,
			ThemeTrack:       "assets/audio/cod_theme.mp3",
			Regions: map[Region]RegionTarget{
				RegionUS: {
					Title:   "Castlevania: Curse of Darkness (US)",
					RomPath: "games/cod/cod_us.iso",
				},
				RegionJP: {
					Title:   "Akumajou Dracula: Yami no Juin (JP)",
					RomPath: "games/cod/cod_jp.iso",
				},
				RegionEU: {
					Title:   "Castlevania: Curse of Darkness (EU)",
					RomPath: "games/cod/cod_eu.iso",
				},
			},
		},
		"lod": {
			ID:             "lod",
			Title:          "Castlevania: Legacy of Darkness",
			OriginalSystem: "Nintendo 64",
			ReleaseYear:    "1999",
			Developer:      "Konami Computer Entertainment Kobe",
			Synopsis:       "The definitive 3D gothic quest on Nintendo 64. Play as Cornell the Man-Beast, Henry the Knight, Reinhardt Schneider, or Carrie Fernandez to rescue young Ada from the dark ritual.",
			RunnerType:     RunnerRetroArch,
			CorePath:       "cores/mupen64plus_next_libretro",
			SupportedRegions: []Region{RegionUS, RegionJP, RegionEU},
			DefaultRegion:    RegionUS,
			ThemeTrack:       "assets/audio/lod_theme.mp3",
			Regions: map[Region]RegionTarget{
				RegionUS: {
					Title:   "Castlevania: Legacy of Darkness (US)",
					RomPath: "games/lod/lod_us.z64",
				},
				RegionJP: {
					Title:   "Akumajou Dracula Mokushiroku Gaiden: Legend of Cornell (JP)",
					RomPath: "games/lod/lod_jp.z64",
				},
				RegionEU: {
					Title:   "Castlevania: Legacy of Darkness (EU)",
					RomPath: "games/lod/lod_eu.z64",
				},
			},
		},
		"cv64": {
			ID:             "cv64",
			Title:          "Castlevania 64",
			OriginalSystem: "Nintendo 64",
			ReleaseYear:    "1999",
			Developer:      "Konami Computer Entertainment Kobe",
			Synopsis:       "Castlevania's first journey into three dimensions. Take control of whip-wielding Reinhardt Schneider or young magical prodigy Carrie Fernandez as they ascend Dracula's castle.",
			RunnerType:     RunnerRetroArch,
			CorePath:       "cores/mupen64plus_next_libretro",
			SupportedRegions: []Region{RegionUS, RegionJP, RegionEU},
			DefaultRegion:    RegionUS,
			ThemeTrack:       "assets/audio/cv64_theme.mp3",
			Regions: map[Region]RegionTarget{
				RegionUS: {
					Title:   "Castlevania (US)",
					RomPath: "games/cv64/cv64_us.z64",
				},
				RegionJP: {
					Title:   "Akumajou Dracula Mokushiroku: Real Action Adventure (JP)",
					RomPath: "games/cv64/cv64_jp.z64",
				},
				RegionEU: {
					Title:   "Castlevania (EU)",
					RomPath: "games/cv64/cv64_eu.z64",
				},
			},
		},
		"rebirth": {
			ID:             "rebirth",
			Title:          "Castlevania: The Adventure ReBirth",
			OriginalSystem: "WiiWare",
			ReleaseYear:    "2009",
			Developer:      "Konami",
			Synopsis:       "M2's acclaimed remake of the classic adventure. Guide Christopher Belmont through punishing traps, branched routes, and classic gothic action with timeless arranged tracks.",
			RunnerType:     RunnerDolphin,
			SupportedRegions: []Region{RegionUS, RegionJP, RegionEU},
			DefaultRegion:    RegionUS,
			ThemeTrack:       "assets/audio/rebirth_theme.mp3",
			Regions: map[Region]RegionTarget{
				RegionUS: {
					Title:   "Castlevania: The Adventure ReBirth (US)",
					RomPath: "games/rebirth/rebirth_us.wad",
				},
				RegionJP: {
					Title:   "Dracula Densetsu ReBirth (JP)",
					RomPath: "games/rebirth/rebirth_jp.wad",
				},
				RegionEU: {
					Title:   "Castlevania: The Adventure ReBirth (EU)",
					RomPath: "games/rebirth/rebirth_eu.wad",
				},
			},
		},
	}
}

// LoadConfig loads the library configuration from config.json or returns default library.
func LoadConfig(configPath string) (map[string]Game, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		library := DefaultLibrary()
		_ = SaveConfig(configPath, library)
		return library, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultLibrary(), fmt.Errorf("failed to read config file: %w", err)
	}

	var library map[string]Game
	if err := json.Unmarshal(data, &library); err != nil {
		return DefaultLibrary(), fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return library, nil
}

// SaveConfig persists the current game library to config.json.
func SaveConfig(configPath string, library map[string]Game) error {
	dir := filepath.Dir(configPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(library, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
