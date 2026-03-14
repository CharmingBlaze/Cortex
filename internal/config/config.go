package config

import (
	"encoding/json"
	"errors"
	"os"
)

// FeatureSet lists toggleable language/runtime features.
type FeatureSet struct {
	Async      bool `json:"async"`
	Actors     bool `json:"actors"`
	Blockchain bool `json:"blockchain"`
	QoL        bool `json:"qol"`
}

// Config represents compiler- and runtime-level configuration.
type Config struct {
	Features     FeatureSet `json:"features"`
	Backend      string     `json:"backend,omitempty"`       // "gcc", "tcc", or "auto" (try tcc then gcc)
	IncludePaths []string   `json:"include_paths,omitempty"` // -I for C compiler
	LibraryPaths []string   `json:"library_paths,omitempty"` // -L for linker
	Libraries    []string   `json:"libraries,omitempty"`     // -l for linker (e.g. raylib, m)
	Debug        bool       `json:"debug,omitempty"`         // enable debug output
}

// DefaultFeatures returns the feature set enabled when no config is provided.
func DefaultFeatures() FeatureSet {
	return FeatureSet{
		Async:      true,
		Actors:     true,
		Blockchain: true,
		QoL:        true,
	}
}

// Default returns the default configuration.
func Default() Config {
	return Config{Features: DefaultFeatures()}
}

// Load reads configuration from path. If the file does not exist, defaults are returned.
func Load(path string) (Config, error) {
	if path == "" {
		return Default(), nil
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return Config{}, err
	}

	cfg := Default()
	if len(bytes) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
