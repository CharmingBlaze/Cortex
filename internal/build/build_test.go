package build

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestFindMSYS2 tests MSYS2 detection
func TestFindMSYS2(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("MSYS2 tests only run on Windows")
	}

	msys, err := FindMSYS2()
	if err != nil {
		t.Logf("MSYS2 not found (this is OK if not installed): %v", err)
		return
	}

	if msys.InstallPath == "" {
		t.Error("MSYS2 found but InstallPath is empty")
	}

	t.Logf("Found MSYS2 at: %s", msys.InstallPath)

	// Check for compiler (may not be installed)
	compiler := msys.GetCompiler()
	if compiler == "" {
		t.Log("No compiler found in MSYS2 (pacman -S mingw-w64-x86_64-gcc to install)")
	} else {
		t.Logf("Found compiler: %s", compiler)
	}
}

// TestCheckMSYS2Path tests MSYS2 path validation
func TestCheckMSYS2Path(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("MSYS2 tests only run on Windows")
	}

	// Test with non-existent path
	result := CheckMSYS2Path("C:/nonexistent")
	if result != nil {
		t.Error("Should return nil for non-existent path")
	}
}

// TestZeroConfigBuilder tests the zero-config builder
func TestZeroConfigBuilder(t *testing.T) {
	// Create a temporary test directory
	tmpDir := t.TempDir()

	// Create a simple C file
	testFile := filepath.Join(tmpDir, "test.c")
	content := `#include <stdio.h>
int main() { printf("Hello\n"); return 0; }
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create builder
	builder, err := NewZeroConfigBuilder(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	builder.Verbose = true

	// Test source file detection
	sources, err := builder.FindSourceFiles()
	if err != nil {
		t.Fatalf("Failed to find sources: %v", err)
	}

	if len(sources) == 0 {
		t.Error("Should find at least one source file")
	} else {
		t.Logf("Found %d source files", len(sources))
	}

	// Test library detection
	includes, libs := builder.AnalyzeSourceFiles(sources)
	t.Logf("Detected includes: %v", includes)
	t.Logf("Detected libraries: %v", libs)
}

// TestLoadConfig tests configuration loading
func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "build.Json")

	config := `{
		"name": "testapp",
		"version": "1.0.0",
		"sources": ["*.c"],
		"output": "testapp"
	}`

	if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Name != "testapp" {
		t.Errorf("Expected name 'testapp', got '%s'", cfg.Name)
	}

	if cfg.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", cfg.Version)
	}

	// Check defaults
	if cfg.Type != "executable" {
		t.Errorf("Expected type 'executable', got '%s'", cfg.Type)
	}
}

// TestFindSourceFiles tests source file discovery
func TestFindSourceFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test C files
	files := []string{"main.c", "helper.c", "utils.c"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("int x;"), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", f, err)
		}
	}

	builder, _ := NewZeroConfigBuilder(tmpDir)
	sources, err := builder.FindSourceFiles()
	if err != nil {
		t.Fatalf("Failed to find sources: %v", err)
	}

	if len(sources) < len(files) {
		t.Errorf("Expected at least %d sources, found %d", len(files), len(sources))
	}
}

// TestLibraryMapping tests library name mapping
func TestLibraryMapping(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("MSYS2 library mapping tests only on Windows")
	}

	msys := &MSYS2{InstallPath: `C:\msys64`}

	tests := []struct {
		input    string
		expected string
	}{
		{"raylib", "mingw-w64-x86_64-raylib"},
		{"SDL2", "mingw-w64-x86_64-SDL2"},
		{"glfw", "mingw-w64-x86_64-glfw"},
	}

	for _, test := range tests {
		names := msys.GetLibraryNames(test.input)
		found := false
		for _, name := range names {
			if name == test.expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s to map to %s, got %v", test.input, test.expected, names)
		}
	}
}

// TestPathConversion tests MSYS2 path conversion
func TestPathConversion(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Path conversion tests only on Windows")
	}

	msys := &MSYS2{InstallPath: `C:\msys64`}

	tests := []struct {
		windows string
		unix    string
	}{
		{`C:\Users\test`, "/c/Users/test"},
		{`D:\projects\myapp`, "/d/projects/myapp"},
		{`C:\msys64\usr\bin`, "/c/msys64/usr/bin"},
	}

	for _, test := range tests {
		result := msys.ConvertPath(test.windows)
		if result != test.unix {
			t.Errorf("ConvertPath(%s) = %s, want %s", test.windows, result, test.unix)
		}
	}
}
