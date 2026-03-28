package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// ProjectConfig represents a cortex.toml project file
type ProjectConfig struct {
	Project struct {
		Name     string          `toml:"name"`
		Version  string          `toml:"version"`
		Entry    string          `toml:"entry"`
		Backend  string          `toml:"backend"`
		Strict   bool            `toml:"strict"`
		Features map[string]bool `toml:"features"`
	} `toml:"project"`
	Dependencies map[string]Dependency `toml:"dependencies"`
}

// Dependency represents a library dependency
type Dependency struct {
	Path        string   `toml:"path"`
	Include     string   `toml:"include"`
	IncludePath string   `toml:"include_path"`
	LibPath     string   `toml:"lib_path"`
	Libs        []string `toml:"libs"`
	LinkerFlags []string `toml:"linker_flags"`
}

// LoadProject loads a cortex.toml file from the given directory
func LoadProject(dir string) (*ProjectConfig, error) {
	tomlPath := filepath.Join(dir, "cortex.toml")

	data, err := os.ReadFile(tomlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cortex.toml: %w", err)
	}

	var cfg ProjectConfig
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse cortex.toml: %w", err)
	}

	// Set defaults
	if cfg.Project.Entry == "" {
		cfg.Project.Entry = "main.cx"
	}

	return &cfg, nil
}

// FindProjectDir searches upward for a cortex.toml file
func FindProjectDir(startDir string) (string, error) {
	dir := startDir
	for {
		tomlPath := filepath.Join(dir, "cortex.toml")
		if _, err := os.Stat(tomlPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no cortex.toml found (searched from %s)", startDir)
		}
		dir = parent
	}
}

// ToConfig converts a ProjectConfig to a Config
func (p *ProjectConfig) ToConfig() *Config {
	cfg := &Config{}

	// Map features
	if p.Project.Features != nil {
		cfg.Features = FeatureSet{
			Async:      p.Project.Features["async"],
			Actors:     p.Project.Features["actors"],
			Blockchain: p.Project.Features["blockchain"],
			QoL:        p.Project.Features["qol"],
		}
	} else {
		cfg.Features = DefaultFeatures()
	}

	if p.Project.Backend != "" {
		cfg.Backend = p.Project.Backend
	} else {
		cfg.Backend = "auto"
	}
	cfg.Strict = p.Project.Strict

	// Collect all include paths, library paths, and libraries from dependencies
	for name, dep := range p.Dependencies {
		if dep.Include != "" {
			cfg.IncludePaths = append(cfg.IncludePaths, dep.Include)
		}
		if dep.IncludePath != "" {
			cfg.IncludePaths = append(cfg.IncludePaths, dep.IncludePath)
		}
		if dep.LibPath != "" {
			cfg.LibraryPaths = append(cfg.LibraryPaths, dep.LibPath)
		}
		if dep.Path != "" {
			// Auto-detect include/lib from path
			cfg.IncludePaths = append(cfg.IncludePaths,
				filepath.Join(dep.Path, "include"),
				filepath.Join(dep.Path, "src"),
			)
			cfg.LibraryPaths = append(cfg.LibraryPaths,
				filepath.Join(dep.Path, "lib"),
				filepath.Join(dep.Path, "build", name),
			)
		}
		cfg.Libraries = append(cfg.Libraries, dep.Libs...)

		// Add the library name itself if no libs specified
		if len(dep.Libs) == 0 && dep.Path != "" {
			cfg.Libraries = append(cfg.Libraries, name)
		}
	}

	// Add system libraries for common libs
	for _, lib := range cfg.Libraries {
		switch lib {
		case "raylib":
			// Add Windows system libs for raylib
			cfg.Libraries = append(cfg.Libraries, "opengl32", "gdi32", "winmm", "shell32")
		}
	}

	return cfg
}
