package engine

// Region represents a game release territory.
type Region string

const (
	RegionUS Region = "US"
	RegionJP Region = "JP"
	RegionEU Region = "EU"
)

// RunnerType specifies which execution engine launches the title.
type RunnerType string

const (
	RunnerNative    RunnerType = "native"
	RunnerRetroArch RunnerType = "retroarch"
	RunnerPCSX2     RunnerType = "pcsx2"
	RunnerDolphin   RunnerType = "dolphin"
)

// RegionTarget holds file target paths for a specific region.
type RegionTarget struct {
	Title     string `json:"title"`
	RomPath   string `json:"rom_path"`
	PatchPath string `json:"patch_path,omitempty"`
	Available bool   `json:"available"`
}

// Game defines full metadata and launch configuration for a title.
type Game struct {
	ID               string                  `json:"id"`
	Title            string                  `json:"title"`
	OriginalSystem   string                  `json:"original_system"`
	ReleaseYear      string                  `json:"release_year"`
	Developer        string                  `json:"developer"`
	Synopsis         string                  `json:"synopsis"`
	RunnerType       RunnerType              `json:"runner_type"`
	ExecutablePath   string                  `json:"executable_path,omitempty"`
	CorePath         string                  `json:"core_path,omitempty"`
	SupportedRegions []Region                `json:"supported_regions"`
	DefaultRegion    Region                  `json:"default_region"`
	Regions          map[Region]RegionTarget `json:"regions"`
	ThemeTrack       string                  `json:"theme_track,omitempty"`
}

// GameStatus informs the frontend whether a title is ready to play.
type GameStatus struct {
	ID             string          `json:"id"`
	Available      bool            `json:"available"`
	ActiveRegion   Region          `json:"active_region"`
	MissingFile    string          `json:"missing_file,omitempty"`
	RegionsStatus  map[Region]bool `json:"regions_status"`
	RunnerReady    bool            `json:"runner_ready"`
	RunnerPath     string          `json:"runner_path"`
}
