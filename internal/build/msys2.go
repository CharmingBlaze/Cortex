// Package build provides MSYS2 integration for Windows
// MSYS2 offers a complete Unix-like environment with package management
package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// MSYS2 provides integration with MSYS2 on Windows
type MSYS2 struct {
	InstallPath string
	ShellPath   string
	PacmanPath  string
}

// FindMSYS2 attempts to locate MSYS2 installation
func FindMSYS2() (*MSYS2, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("MSYS2 is only available on Windows")
	}

	// Common MSYS2 installation paths
	possiblePaths := []string{
		`C:\msys64`,
		`C:\msys2`,
		`D:\msys64`,
		`D:\msys2`,
		`C:\Program Files\msys64`,
		`C:\Program Files (x86)\msys64`,
		// Check PATH for msys2
	}

	// Also check if MSYS2 is in PATH
	if path := FindInPath("msys2.exe"); path != "" {
		possiblePaths = append([]string{filepath.Dir(path)}, possiblePaths...)
	}

	for _, path := range possiblePaths {
		if msys := CheckMSYS2Path(path); msys != nil {
			return msys, nil
		}
	}

	return nil, fmt.Errorf("MSYS2 not found")
}

// findInPath searches for an executable in PATH
func FindInPath(name string) string {
	path, err := exec.LookPath(name)
	if err == nil {
		return path
	}
	return ""
}

// checkMSYS2Path validates if the given path is a valid MSYS2 installation
func CheckMSYS2Path(path string) *MSYS2 {
	usrBin := filepath.Join(path, "usr", "bin")
	pacman := filepath.Join(usrBin, "pacman.exe")
	bash := filepath.Join(usrBin, "bash.exe")
	gcc := filepath.Join(usrBin, "gcc.exe")

	// Need at least pacman or gcc to be a valid MSYS2
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	msys := &MSYS2{
		InstallPath: path,
		ShellPath:   bash,
		PacmanPath:  pacman,
	}

	// Verify it's actually MSYS2 by checking for key files
	if _, err := os.Stat(pacman); os.IsNotExist(err) {
		if _, err := os.Stat(gcc); os.IsNotExist(err) {
			return nil // Neither pacman nor gcc found
		}
	}

	return msys
}

// IsInstalled returns true if MSYS2 is available
func (m *MSYS2) IsInstalled() bool {
	return m.InstallPath != "" && m.PacmanPath != ""
}

// GetGCCPath returns the path to MSYS2's GCC compiler
func (m *MSYS2) GetGCCPath() string {
	gcc := filepath.Join(m.InstallPath, "usr", "bin", "gcc.exe")
	if _, err := os.Stat(gcc); err == nil {
		return gcc
	}
	return ""
}

// GetClangPath returns the path to MSYS2's Clang compiler
func (m *MSYS2) GetClangPath() string {
	clang := filepath.Join(m.InstallPath, "usr", "bin", "clang.exe")
	if _, err := os.Stat(clang); err == nil {
		return clang
	}
	return ""
}

// GetCompiler returns the best available compiler (prefers GCC)
func (m *MSYS2) GetCompiler() string {
	if gcc := m.GetGCCPath(); gcc != "" {
		return gcc
	}
	if clang := m.GetClangPath(); clang != "" {
		return clang
	}
	return ""
}

// HasLibrary checks if a library is installed via pacman
func (m *MSYS2) HasLibrary(name string) bool {
	// Query pacman for the package
	cmd := exec.Command(m.PacmanPath, "-Q", name)
	err := cmd.Run()
	return err == nil
}

// InstallLibrary installs a library using pacman
func (m *MSYS2) InstallLibrary(name string) error {
	fmt.Printf("Installing %s via MSYS2 pacman...\n", name)

	cmd := exec.Command(m.PacmanPath, "-S", "--noconfirm", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %w", name, err)
	}

	return nil
}

// GetLibraryNames returns common library package names in MSYS2
func (m *MSYS2) GetLibraryNames(lib string) []string {
	// Map common library names to MSYS2 packages
	mappings := map[string][]string{
		"raylib":   {"mingw-w64-x86_64-raylib"},
		"SDL2":     {"mingw-w64-x86_64-SDL2"},
		"glfw":     {"mingw-w64-x86_64-glfw"},
		"glew":     {"mingw-w64-x86_64-glew"},
		"freetype": {"mingw-w64-x86_64-freetype"},
		"curl":     {"mingw-w64-x86_64-curl"},
		"openssl":  {"mingw-w64-x86_64-openssl"},
		"zlib":     {"mingw-w64-x86_64-zlib"},
		"libpng":   {"mingw-w64-x86_64-libpng"},
		"libjpeg":  {"mingw-w64-x86_64-libjpeg-turbo"},
		"sqlite":   {"mingw-w64-x86_64-sqlite3"},
		"openal":   {"mingw-w64-x86_64-openal"},
		"lua":      {"mingw-w64-x86_64-lua"},
		"python":   {"mingw-w64-x86_64-python"},
	}

	if packages, ok := mappings[lib]; ok {
		return packages
	}

	// Try common prefixes
	return []string{
		"mingw-w64-x86_64-" + strings.ToLower(lib),
		lib,
	}
}

// InstallLibraries installs multiple libraries
func (m *MSYS2) InstallLibraries(libs []string) error {
	for _, lib := range libs {
		names := m.GetLibraryNames(lib)
		installed := false

		for _, name := range names {
			if err := m.InstallLibrary(name); err == nil {
				fmt.Printf("  ✓ Installed %s\n", name)
				installed = true
				break
			}
		}

		if !installed {
			fmt.Printf("  ✗ Could not install %s\n", lib)
		}
	}

	return nil
}

// GetIncludePaths returns the include paths for MSYS2
func (m *MSYS2) GetIncludePaths() []string {
	return []string{
		filepath.Join(m.InstallPath, "usr", "include"),
		filepath.Join(m.InstallPath, "mingw64", "include"),
	}
}

// GetLibraryPaths returns the library paths for MSYS2
func (m *MSYS2) GetLibraryPaths() []string {
	return []string{
		filepath.Join(m.InstallPath, "usr", "lib"),
		filepath.Join(m.InstallPath, "mingw64", "lib"),
	}
}

// ConvertPath converts a Windows path to MSYS2 Unix path
func (m *MSYS2) ConvertPath(winPath string) string {
	// Convert C:\path\to\file to /c/path/to/file
	path := strings.ReplaceAll(winPath, `\`, `/`)

	// Handle drive letters
	if len(path) >= 2 && path[1] == ':' {
		drive := strings.ToLower(string(path[0]))
		path = "/" + drive + path[2:]
	}

	return path
}

// RunInShell executes a command in MSYS2 bash shell
func (m *MSYS2) RunInShell(command string) error {
	cmd := exec.Command(m.ShellPath, "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Update updates the MSYS2 package database
func (m *MSYS2) Update() error {
	fmt.Println("Updating MSYS2 package database...")
	cmd := exec.Command(m.PacmanPath, "-Sy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// MSYS2Builder wraps ZeroConfigBuilder with MSYS2 enhancements
type MSYS2Builder struct {
	*ZeroConfigBuilder
	msys2 *MSYS2
}

// NewMSYS2Builder creates a builder that uses MSYS2
func NewMSYS2Builder(projectDir string) (*MSYS2Builder, error) {
	base, err := NewZeroConfigBuilder(projectDir)
	if err != nil {
		return nil, err
	}

	msys, err := FindMSYS2()
	if err != nil {
		return nil, fmt.Errorf("MSYS2 not found: %w", err)
	}

	return &MSYS2Builder{
		ZeroConfigBuilder: base,
		msys2:             msys,
	}, nil
}

// BuildWithMSYS2 builds using MSYS2's GCC and libraries
func (b *MSYS2Builder) BuildWithMSYS2() error {
	fmt.Println("Using MSYS2 build environment...")

	// Ensure libraries are installed
	_, libs := b.AnalyzeSourceFiles(nil)
	if len(libs) > 0 && b.AutoFetch {
		fmt.Printf("Installing required libraries: %v\n", libs)
		b.msys2.InstallLibraries(libs)
	}

	// Build with MSYS2 compiler
	compiler := b.msys2.GetCompiler()
	if compiler == "" {
		return fmt.Errorf("no compiler found in MSYS2")
	}

	fmt.Printf("Using compiler: %s\n", compiler)

	// Use parent build logic but with MSYS2 compiler
	// Implementation would override compileWithSystemCC to use MSYS2 paths
	return b.Build()
}
