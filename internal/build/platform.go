// Package build provides cross-platform utilities for the build system
package build

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// GetCortexDir returns the directory where the cortex executable is located
// This ensures the build system works from any directory
func GetCortexDir() string {
	exe, err := os.Executable()
	if err != nil {
		// Fallback to current directory
		return "."
	}
	return filepath.Dir(exe)
}

// GetCortexDataDir returns the directory for cortex data (TCC, libs, etc.)
func GetCortexDataDir() string {
	cortexDir := GetCortexDir()
	dataDir := filepath.Join(cortexDir, "data")

	// Create if doesn't exist
	os.MkdirAll(dataDir, 0755)

	return dataDir
}

// Platform represents the current operating system
type Platform struct {
	OS   string // windows, linux, darwin
	Arch string // amd64, arm64, 386
}

// CurrentPlatform returns the current platform info
func CurrentPlatform() Platform {
	return Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// String returns a human-readable platform description
func (p Platform) String() string {
	osName := map[string]string{
		"windows": "Windows",
		"linux":   "Linux",
		"darwin":  "macOS",
	}[p.OS]
	if osName == "" {
		osName = p.OS
	}

	archName := map[string]string{
		"amd64": "x86_64",
		"arm64": "ARM64",
		"386":   "x86",
	}[p.Arch]
	if archName == "" {
		archName = p.Arch
	}

	return fmt.Sprintf("%s %s", osName, archName)
}

// IsWindows returns true if running on Windows
func (p Platform) IsWindows() bool {
	return p.OS == "windows"
}

// IsLinux returns true if running on Linux
func (p Platform) IsLinux() bool {
	return p.OS == "linux"
}

// IsMacOS returns true if running on macOS
func (p Platform) IsMacOS() bool {
	return p.OS == "darwin"
}

// GetExecutableExtension returns the executable extension for the platform
func (p Platform) GetExecutableExtension() string {
	if p.IsWindows() {
		return ".exe"
	}
	return ""
}

// GetLibraryExtension returns the shared library extension for the platform
func (p Platform) GetLibraryExtension() string {
	switch p.OS {
	case "windows":
		return ".dll"
	case "darwin":
		return ".dylib"
	default:
		return ".so"
	}
}

// GetStaticLibraryExtension returns the static library extension for the platform
func (p Platform) GetStaticLibraryExtension() string {
	if p.IsWindows() {
		return ".lib"
	}
	return ".a"
}

// NormalizePath converts a path to the platform's native format
func NormalizePath(path string) string {
	if runtime.GOOS == "windows" {
		// Already using filepath.Join which handles this
		return filepath.Clean(path)
	}
	return filepath.Clean(path)
}

// ToSlash converts a path to use forward slashes (for cross-platform compatibility)
func ToSlash(path string) string {
	return filepath.ToSlash(path)
}

// FromSlash converts a path from forward slashes to platform-specific
func FromSlash(path string) string {
	return filepath.FromSlash(path)
}

// CrossPlatformConfig provides platform-specific configuration
type CrossPlatformConfig struct {
	Platform
}

// NewCrossPlatformConfig creates a config for the current platform
func NewCrossPlatformConfig() *CrossPlatformConfig {
	return &CrossPlatformConfig{
		Platform: CurrentPlatform(),
	}
}

// GetCompilerSearchOrder returns the preferred compiler search order for the platform
func (c *CrossPlatformConfig) GetCompilerSearchOrder() []string {
	switch c.OS {
	case "windows":
		return []string{
			"gcc",   // MinGW/MSYS2 GCC
			"clang", // LLVM/MSYS2 Clang
			"tcc",   // Tiny C Compiler
			"cl",    // MSVC (if available)
		}
	case "darwin":
		return []string{
			"clang", // Xcode/Command Line Tools
			"gcc",   // Homebrew GCC
			"tcc",   // Tiny C Compiler
		}
	default: // linux
		return []string{
			"gcc",   // System GCC
			"clang", // LLVM
			"tcc",   // Tiny C Compiler
		}
	}
}

// GetLibrarySearchPaths returns common library search paths for the platform
func (c *CrossPlatformConfig) GetLibrarySearchPaths() []string {
	switch c.OS {
	case "windows":
		return []string{
			`C:\msys64\mingw64\lib`,
			`C:\msys64\usr\lib`,
		}
	case "darwin":
		return []string{
			"/usr/local/lib",
			"/opt/homebrew/lib",
			"/usr/lib",
		}
	default: // linux
		return []string{
			"/usr/local/lib",
			"/usr/lib",
			"/lib",
		}
	}
}

// GetIncludeSearchPaths returns common include search paths for the platform
func (c *CrossPlatformConfig) GetIncludeSearchPaths() []string {
	switch c.OS {
	case "windows":
		return []string{
			`C:\msys64\mingw64\include`,
			`C:\msys64\usr\include`,
		}
	case "darwin":
		return []string{
			"/usr/local/include",
			"/opt/homebrew/include",
			"/usr/include",
		}
	default: // linux
		return []string{
			"/usr/local/include",
			"/usr/include",
		}
	}
}
