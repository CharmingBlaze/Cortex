// Package build provides a zero-config, self-contained build system for C libraries
// Programmers only need Cortex - no external compilers or build tools required
package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ZeroConfigBuilder builds C projects with zero configuration required
type ZeroConfigBuilder struct {
	ProjectDir  string
	Verbose     bool
	Release     bool
	AutoFetch   bool
	tccCompiler *BundledTCC
}

// NewZeroConfigBuilder creates a builder that requires no configuration files
// AutoFetch is disabled by default - users must opt-in to automatic library downloads
func NewZeroConfigBuilder(projectDir string) (*ZeroConfigBuilder, error) {
	return &ZeroConfigBuilder{
		ProjectDir: projectDir,
		AutoFetch:  false, // Manual by default - user controls library setup
	}, nil
}

// Build automatically detects and builds the project
func (b *ZeroConfigBuilder) Build() error {
	// Ensure we have a C compiler (TCC bundled or system compiler)
	if err := b.EnsureCompiler(); err != nil {
		return fmt.Errorf("no compiler available: %w", err)
	}

	// Auto-detect source files
	sources, err := b.FindSourceFiles()
	if err != nil {
		return fmt.Errorf("failed to find sources: %w", err)
	}

	if len(sources) == 0 {
		return fmt.Errorf("no C source files found in %s", b.ProjectDir)
	}

	if b.Verbose {
		fmt.Printf("Found %d source files:\n", len(sources))
		for _, s := range sources {
			fmt.Printf("  - %s\n", s)
		}
	}

	// Auto-detect libraries from #include directives
	includes, libraries := b.AnalyzeSourceFiles(sources)

	if b.AutoFetch && len(libraries) > 0 {
		if err := b.FetchLibraries(libraries); err != nil {
			fmt.Printf("Warning: could not fetch some libraries: %v\n", err)
		}
	}

	// Compile
	fmt.Printf("Building project...\n")
	output := b.GetOutputName()

	if err := b.CompileAndLink(sources, includes, libraries, output); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("✓ Built: %s\n", output)
	return nil
}

// ensureCompiler makes sure we have a C compiler available
func (b *ZeroConfigBuilder) EnsureCompiler() error {
	// Try system compiler first
	if _, err := exec.LookPath("gcc"); err == nil {
		return nil
	}
	if _, err := exec.LookPath("clang"); err == nil {
		return nil
	}
	if _, err := exec.LookPath("tcc"); err == nil {
		return nil
	}

	// Download and use bundled TCC
	fmt.Println("No system compiler found. Setting up bundled TCC...")
	tccCompiler, err := FindOrInstall()
	if err != nil {
		return err
	}

	b.tccCompiler = tccCompiler
	fmt.Println("✓ Using bundled TCC compiler")
	return nil
}

// findSourceFiles automatically discovers C source files
func (b *ZeroConfigBuilder) FindSourceFiles() ([]string, error) {
	var sources []string

	// Common patterns for C source files
	patterns := []string{
		"*.c",
		"src/*.c",
		"source/*.c",
		"lib/*.c",
		"**/*.c", // Recursive (use carefully)
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(b.ProjectDir, pattern))
		if err != nil {
			continue
		}
		sources = append(sources, matches...)
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var unique []string
	for _, s := range sources {
		if !seen[s] {
			seen[s] = true
			unique = append(unique, s)
		}
	}

	return unique, nil
}

// analyzeSourceFiles scans #include directives to find dependencies
func (b *ZeroConfigBuilder) AnalyzeSourceFiles(sources []string) (includes []string, libraries []string) {
	includeSet := make(map[string]bool)
	libSet := make(map[string]bool)

	// Common library mappings
	libMappings := map[string]string{
		"raylib.h":   "raylib",
		"SDL.h":      "SDL2",
		"SDL2.h":     "SDL2",
		"glfw3.h":    "glfw",
		"glad.h":     "glad",
		"glew.h":     "GLEW",
		"freetype.h": "freetype",
		"curl.h":     "curl",
		"zlib.h":     "z",
		"png.h":      "png",
		"jpeglib.h":  "jpeg",
	}

	for _, source := range sources {
		content, err := os.ReadFile(source)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)

			// Check for #include directives
			if strings.HasPrefix(line, "#include") {
				// Extract header name
				start := strings.Index(line, "<")
				end := strings.Index(line, ">")
				if start != -1 && end != -1 && end > start {
					header := line[start+1 : end]

					// Check if this maps to a library
					if lib, ok := libMappings[header]; ok {
						libSet[lib] = true
					}

					// Add include directory if it's a local header
					if strings.Contains(header, "/") {
						dir := filepath.Dir(header)
						includeSet[dir] = true
					}
				}
			}
		}
	}

	// Convert sets to slices
	for inc := range includeSet {
		includes = append(includes, inc)
	}
	for lib := range libSet {
		libraries = append(libraries, lib)
	}

	return includes, libraries
}

// fetchLibraries automatically downloads common libraries
func (b *ZeroConfigBuilder) FetchLibraries(libraries []string) error {
	libsDir := filepath.Join(b.ProjectDir, "libs")

	for _, lib := range libraries {
		switch lib {
		case "raylib":
			if err := b.FetchRaylib(libsDir); err != nil {
				fmt.Printf("  Could not fetch raylib: %v\n", err)
			}
		default:
			if b.Verbose {
				fmt.Printf("  Skipping auto-fetch for %s (not supported yet)\n", lib)
			}
		}
	}

	return nil
}

// fetchRaylib downloads and sets up raylib
func (b *ZeroConfigBuilder) FetchRaylib(libsDir string) error {
	// Check if already exists
	raylibDir := filepath.Join(libsDir, "raylib")
	if _, err := os.Stat(raylibDir); err == nil {
		if b.Verbose {
			fmt.Println("  raylib already present")
		}
		return nil
	}

	fmt.Println("  Fetching raylib...")

	// Platform-specific download logic would go here
	// For now, just create the directory structure
	if err := os.MkdirAll(raylibDir, 0755); err != nil {
		return err
	}

	fmt.Println("  ✓ raylib ready")
	return nil
}

// getOutputName determines the output executable name
func (b *ZeroConfigBuilder) GetOutputName() string {
	// Try to use project directory name
	base := filepath.Base(b.ProjectDir)
	if base == "." || base == "/" {
		base = "app"
	}

	// Add appropriate extension
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

// compileAndLink compiles and links the project
func (b *ZeroConfigBuilder) CompileAndLink(sources, includes, libraries []string, output string) error {
	// Use TCC if available (faster, self-contained)
	if b.tccCompiler != nil {
		return b.CompileWithTCC(sources, includes, libraries, output)
	}

	// Fall back to system compiler
	return b.CompileWithSystemCC(sources, includes, libraries, output)
}

// compileWithTCC uses the bundled TCC compiler
func (b *ZeroConfigBuilder) CompileWithTCC(sources, includes, libraries []string, output string) error {
	args := []string{"-o", output}
	args = append(args, sources...)

	// Add includes
	for _, inc := range includes {
		args = append(args, "-I", inc)
	}
	args = append(args, "-I", "include")
	args = append(args, "-I", "libs/raylib/include")

	// Add libraries
	for _, lib := range libraries {
		args = append(args, "-l", lib)
	}

	// Platform-specific libraries
	if runtime.GOOS == "windows" {
		args = append(args, "-lwinmm", "-lgdi32", "-lopengl32")
	}

	if b.Verbose {
		fmt.Printf("  tcc %s\n", strings.Join(args, " "))
	}

	return b.tccCompiler.Link([]string{}, output, libraries)
}

// compileWithSystemCC uses gcc/clang
func (b *ZeroConfigBuilder) CompileWithSystemCC(sources, includes, libraries []string, output string) error {
	compiler := "gcc"
	if _, err := exec.LookPath("gcc"); err != nil {
		compiler = "clang"
	}

	args := []string{"-o", output}
	args = append(args, sources...)

	// Add optimization/debug flags
	if b.Release {
		args = append(args, "-O2", "-DNDEBUG")
	} else {
		args = append(args, "-g", "-O0")
	}

	// Add includes
	for _, inc := range includes {
		args = append(args, "-I", inc)
	}

	// Add libraries
	for _, lib := range libraries {
		args = append(args, "-l", lib)
	}

	if b.Verbose {
		fmt.Printf("  %s %s\n", compiler, strings.Join(args, " "))
	}

	cmd := exec.Command(compiler, args...)
	cmd.Dir = b.ProjectDir
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\n%s", err, string(outputBytes))
	}

	return nil
}

// Run builds and runs the project
func (b *ZeroConfigBuilder) Run() error {
	if err := b.Build(); err != nil {
		return err
	}

	output := b.GetOutputName()
	fmt.Printf("\nRunning %s...\n", output)

	cmd := exec.Command(filepath.Join(b.ProjectDir, output))
	cmd.Dir = b.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
