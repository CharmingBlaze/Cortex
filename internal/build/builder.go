// Package build provides a simple build system for C libraries
package build

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Config represents a build configuration
type Config struct {
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Sources   []string `json:"sources"`
	Includes  []string `json:"includes"`
	Libraries []string `json:"libraries"`
	LibPaths  []string `json:"lib_paths"`
	Output    string   `json:"output"`
	Compiler  string   `json:"compiler"`
	CFlags    []string `json:"cflags"`
	LDFlags   []string `json:"ldflags"`
	Type      string   `json:"type"` // "executable" or "library"
}

// Dependency represents an external dependency
type Dependency struct {
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Source    string   `json:"source"`
	Path      string   `json:"path"`
	Includes  []string `json:"includes"`
	LibPaths  []string `json:"lib_paths"`
	Libraries []string `json:"libraries"`
}

// Builder handles the build process
type Builder struct {
	Config       Config
	Dependencies []Dependency
	Verbose      bool
	Clean        bool
	Release      bool
	Manual       bool // Disable auto-detection
	NoAutoFetch  bool // Don't auto-download libraries
}

// NewBuilder creates a new builder instance
func NewBuilder(config Config) *Builder {
	return &Builder{
		Config:       config,
		Dependencies: []Dependency{},
		Verbose:      false,
		Clean:        false,
		Release:      false,
	}
}

// LoadConfig loads build configuration from a JSON file
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if config.Compiler == "" {
		config.Compiler = AutoDetectCompiler()
	}
	if config.Type == "" {
		config.Type = "executable"
	}
	if config.Output == "" {
		config.Output = config.Name
	}

	return config, nil
}

// autoDetectCompiler finds an available C compiler
func AutoDetectCompiler() string {
	candidates := []string{"gcc", "tcc", "clang", "cc"}
	for _, compiler := range candidates {
		if _, err := exec.LookPath(compiler); err == nil {
			return compiler
		}
	}
	return "gcc" // fallback
}

// Build executes the build process
func (b *Builder) Build() error {
	if b.Clean {
		if err := b.CleanBuild(); err != nil {
			return fmt.Errorf("clean failed: %w", err)
		}
	}

	// Resolve dependencies
	if err := b.ResolveDependencies(); err != nil {
		return fmt.Errorf("dependency resolution failed: %w", err)
	}

	// Find source files
	sources, err := b.FindSources()
	if err != nil {
		return fmt.Errorf("source discovery failed: %w", err)
	}

	if len(sources) == 0 {
		return fmt.Errorf("no source files found")
	}

	// Compile object files
	objects := make([]string, 0, len(sources))
	for _, source := range sources {
		obj, err := b.CompileSource(source)
		if err != nil {
			return fmt.Errorf("compilation failed for %s: %w", source, err)
		}
		objects = append(objects, obj)
	}

	// Link final output
	if err := b.Link(objects); err != nil {
		return fmt.Errorf("linking failed: %w", err)
	}

	fmt.Printf("Build successful: %s\n", b.Config.Output)
	return nil
}

// findSources expands glob patterns to find all source files
func (b *Builder) FindSources() ([]string, error) {
	var sources []string

	for _, pattern := range b.Config.Sources {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %s: %w", pattern, err)
		}
		sources = append(sources, matches...)
	}

	return sources, nil
}

// compileSource compiles a single C source file to an object file
func (b *Builder) CompileSource(source string) (string, error) {
	// Create build directory if it doesn't exist
	buildDir := "cortex/internal/build"
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return "", err
	}

	// Generate object file path
	base := filepath.Base(source)
	obj := filepath.Join(buildDir, strings.TrimSuffix(base, filepath.Ext(base))+".o")

	// Check if we need to recompile
	if !b.NeedsRecompile(source, obj) {
		if b.Verbose {
			fmt.Printf("  [cached] %s\n", source)
		}
		return obj, nil
	}

	// Build compile command
	args := []string{"-c", source, "-o", obj}

	// Add includes
	for _, inc := range b.Config.Includes {
		args = append(args, "-I", inc)
	}

	// Add dependency includes
	for _, dep := range b.Dependencies {
		for _, inc := range dep.Includes {
			args = append(args, "-I", inc)
		}
	}

	// Add compiler flags
	if b.Release {
		args = append(args, "-O2", "-DNDEBUG")
	} else {
		args = append(args, "-g", "-O0")
	}
	args = append(args, b.Config.CFlags...)

	// Run compiler
	cmd := exec.Command(b.Config.Compiler, args...)
	if b.Verbose {
		fmt.Printf("  %s\n", strings.Join(cmd.Args, " "))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, string(output))
	}

	return obj, nil
}

// needsRecompile checks if source file is newer than object file
func (b *Builder) NeedsRecompile(source, obj string) bool {
	srcInfo, err := os.Stat(source)
	if err != nil {
		return true
	}

	objInfo, err := os.Stat(obj)
	if err != nil {
		return true // Object doesn't exist
	}

	return srcInfo.ModTime().After(objInfo.ModTime())
}

// link links object files into final executable or library
func (b *Builder) Link(objects []string) error {
	args := []string{}

	if b.Config.Type == "library" {
		// Create static library using ar
		args = append(args, "rcs", b.Config.Output)
		args = append(args, objects...)
		cmd := exec.Command("ar", args...)
		if b.Verbose {
			fmt.Printf("  %s\n", strings.Join(cmd.Args, " "))
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %s", err, string(output))
		}
		return nil
	}

	// Link executable
	args = append(args, objects...)
	args = append(args, "-o", b.Config.Output)

	// Add library paths
	for _, path := range b.Config.LibPaths {
		args = append(args, "-L", path)
	}

	// Add dependency library paths
	for _, dep := range b.Dependencies {
		for _, path := range dep.LibPaths {
			args = append(args, "-L", path)
		}
	}

	// Add libraries
	for _, lib := range b.Config.Libraries {
		args = append(args, "-l", lib)
	}

	// Add dependency libraries
	for _, dep := range b.Dependencies {
		for _, lib := range dep.Libraries {
			args = append(args, "-l", lib)
		}
	}

	// Add linker flags
	args = append(args, b.Config.LDFlags...)

	// Run linker
	cmd := exec.Command(b.Config.Compiler, args...)
	if b.Verbose {
		fmt.Printf("  %s\n", strings.Join(cmd.Args, " "))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}

	return nil
}

// resolveDependencies resolves external dependencies
func (b *Builder) ResolveDependencies() error {
	// For now, just verify dependency paths exist
	for _, dep := range b.Dependencies {
		if dep.Path != "" {
			if _, err := os.Stat(dep.Path); err != nil {
				return fmt.Errorf("dependency %s not found at %s", dep.Name, dep.Path)
			}
		}
	}
	return nil
}

// CleanBuild removes all build artifacts
func (b *Builder) CleanBuild() error {
	buildDir := "cortex/internal/build"
	if err := os.RemoveAll(buildDir); err != nil {
		return err
	}
	return os.RemoveAll(b.Config.Output)
}
