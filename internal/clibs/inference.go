package clibs

import (
	"path/filepath"
	"strings"
)

// LibraryMapping maps header names to library names.
var LibraryMapping = map[string]string{
	"raylib.h":    "raylib",
	"SDL.h":       "sdl2",
	"curl/curl.h": "curl",
	"stdio.h":     "c",
	"stdlib.h":    "c",
	"string.h":    "c",
	"math.h":      "c",
}

// InferLibraryFromHeader infers the library name from a given header file.
func InferLibraryFromHeader(header string) string {
	// Normalize header path
	header = strings.Trim(header, "<>\"\\")
	header = filepath.ToSlash(header)

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
	base := filepath.Base(header)
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
