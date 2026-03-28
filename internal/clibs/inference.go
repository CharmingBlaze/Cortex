package clibs

import (
	"path/filepath"
	"strings"
)

// LibraryMapping maps header names to library names.
// Note: Standard C library headers (stdio.h, stdlib.h, etc.) don't need explicit linking
// on any platform - GCC/MSVC automatically link the C runtime.
var LibraryMapping = map[string]string{
	"raylib.h":       "raylib",
	"SDL.h":          "sdl2",
	"SDL2/SDL.h":     "sdl2",
	"curl/curl.h":    "curl",
	"sqlite3.h":      "sqlite3",
	"GLFW/glfw3.h":   "glfw",
	"png.h":          "png",
	"jpeglib.h":      "jpeg",
	"zlib.h":         "z",
	"openssl/ssl.h":  "ssl",
	"openssl/crypto.h": "crypto",
	// Standard C headers - no explicit linking needed
}

// Standard C headers that don't need explicit library linking
var standardCHeaders = map[string]bool{
	"stdio.h":   true,
	"stdlib.h":  true,
	"string.h":  true,
	"math.h":    true,
	"time.h":    true,
	"ctype.h":   true,
	"errno.h":   true,
	"assert.h":  true,
	"stddef.h":  true,
	"stdarg.h":  true,
	"limits.h":  true,
	"float.h":   true,
	"signal.h":  true,
	"setjmp.h":  true,
	"stdbool.h": true,
	// Cortex runtime headers - built into the compiler
	"gui_runtime.h": true,
	"core.h":        true,
	"game.h":        true,
	"network.h":     true,
	"async.h":       true,
	"thread.h":      true,
	"managed.h":     true,
}

// InferLibraryFromHeader infers the library name from a given header file.
func InferLibraryFromHeader(header string) string {
	// Normalize header path
	header = strings.Trim(header, "<>\"\\")
	header = filepath.ToSlash(header)

	// Check if it's a standard C header - no linking needed
	if standardCHeaders[header] {
		return ""
	}

	// Check direct mapping
	if lib, exists := LibraryMapping[header]; exists {
		return lib
	}

	// Check for headers in subdirectories (e.g., curl/curl.h)
	for mappedHeader, lib := range LibraryMapping {
		if strings.HasSuffix(header, mappedHeader) {
			return lib
		}
	}

	// If not found, infer from the header name (remove .h and use basename)
	// But skip standard C headers
	base := filepath.Base(header)
	if standardCHeaders[base] {
		return ""
	}
	if strings.HasSuffix(base, ".h") {
		return strings.TrimSuffix(base, ".h")
	}
	return base
}

// DedupeLibraries removes duplicates from a list of inferred libraries.
func DedupeLibraries(libs []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, lib := range libs {
		if !seen[lib] {
			seen[lib] = true
			result = append(result, lib)
		}
	}
	return result
}
