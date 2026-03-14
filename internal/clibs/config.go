package clibs

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// LibraryConfig holds configuration data for a specific library.
type LibraryConfig struct {
	IncludePaths []string `json:"includePaths"`
	LibraryPaths []string `json:"libraryPaths"`
	LinkerFlags  []string `json:"linkerFlags"`
	HelperFiles  []string `json:"helperFiles"`
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
