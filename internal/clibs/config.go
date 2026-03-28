package clibs

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// LibraryConfig holds configuration data for a specific library.
// JSON matches shipped configs (e.g. configs/raylib.json): camelCase keys.
type LibraryConfig struct {
	IncludePaths []string `json:"includePaths"`
	LibraryPaths []string `json:"libraryPaths"`
	LinkerFlags  []string `json:"linkerFlags"`
	HelperFiles  []string `json:"helperFiles"`
	Libraries    []string `json:"libraries"` // bare names → -lname when linking
	CFlags       []string `json:"cflags"`    // extra compile flags (e.g. -DRAYGUI_IMPLEMENTATION)
}

// LoadLibraryConfig loads a library configuration from a JSON file in the configs directory.
func LoadLibraryConfig(configDir, libName string) (*LibraryConfig, error) {
	configPath := filepath.Join(configDir, libName+".json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil // No config file, return nil without error
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config LibraryConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// ApplyConfig applies the library configuration to the compiler options.
func ApplyConfig(config *LibraryConfig, includePaths, libraryPaths, linkerFlags *[]string) {
	if config == nil {
		return
	}
	*includePaths = append(*includePaths, config.IncludePaths...)
	*libraryPaths = append(*libraryPaths, config.LibraryPaths...)
	*linkerFlags = append(*linkerFlags, config.LinkerFlags...)
}

// LinkArgvFromConfig returns linker argv pieces from a library JSON: linkerFlags as-is, plus -l for each Libraries entry.
func LinkArgvFromConfig(config *LibraryConfig) []string {
	if config == nil {
		return nil
	}
	var out []string
	out = append(out, config.LinkerFlags...)
	for _, name := range config.Libraries {
		if name == "" {
			continue
		}
		out = append(out, "-l"+name)
	}
	return out
}

// CFlagsFromConfig returns extra C compiler flags from the library JSON.
func CFlagsFromConfig(config *LibraryConfig) []string {
	if config == nil {
		return nil
	}
	return append([]string(nil), config.CFlags...)
}
