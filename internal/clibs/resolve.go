package clibs

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolvedBuild aggregates include/lib paths, compile flags, extra C sources, and linker argv
// from per-library JSON configs plus fallback -l for libraries with no config file.
type ResolvedBuild struct {
	IncludePaths   []string
	LibraryPaths   []string
	CFlags         []string
	HelperSources  []string
	LinkArgv       []string
	ConfiguredLibs map[string]bool // lib names that had a configs/<name>.json
}

// ConfigSearchDirs returns ordered directories that contain library JSON files (each path is .../configs).
func ConfigSearchDirs() []string {
	seen := make(map[string]bool)
	var dirs []string
	add := func(dir string) {
		if dir == "" {
			return
		}
		abs, err := filepath.Abs(dir)
		if err != nil {
			abs = dir
		}
		if seen[abs] {
			return
		}
		seen[abs] = true
		dirs = append(dirs, abs)
	}
	if cwd, err := os.Getwd(); err == nil {
		add(filepath.Join(cwd, "configs"))
	}
	if root := os.Getenv("CORTEX_ROOT"); root != "" {
		add(filepath.Join(root, "configs"))
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		add(filepath.Join(dir, "configs"))
		add(filepath.Join(dir, "..", "configs"))
	}
	return dirs
}

// FindLibraryConfig loads the first configs/<lib>.json found in searchDirs.
func FindLibraryConfig(libName string, searchDirs []string) (*LibraryConfig, string, error) {
	for _, d := range searchDirs {
		cfg, err := LoadLibraryConfig(d, libName)
		if err != nil {
			return nil, "", err
		}
		if cfg != nil {
			return cfg, filepath.Join(d, libName+".json"), nil
		}
	}
	return nil, "", nil
}

// ResolveLibraries merges per-library JSON (paths, cflags, helpers, link argv) and appends -l<name>
// for any name in libNames that did not have its own config file.
// Relative paths in JSON are resolved from cwd (project / current working directory).
func ResolveLibraries(libNames []string, cwd string) (ResolvedBuild, error) {
	libNames = DedupeLibraries(libNames)
	searchDirs := ConfigSearchDirs()
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	var rb ResolvedBuild
	rb.ConfiguredLibs = make(map[string]bool)

	for _, lib := range libNames {
		if lib == "" {
			continue
		}
		cfg, configPath, err := FindLibraryConfig(lib, searchDirs)
		if err != nil {
			return ResolvedBuild{}, fmt.Errorf("library config %q: %w", lib, err)
		}
		if cfg == nil {
			continue
		}
		rb.ConfiguredLibs[lib] = true
		configDir := filepath.Dir(configPath)

		rb.IncludePaths = append(rb.IncludePaths, cfg.IncludePaths...)
		rb.LibraryPaths = append(rb.LibraryPaths, cfg.LibraryPaths...)
		rb.CFlags = append(rb.CFlags, CFlagsFromConfig(cfg)...)
		rb.LinkArgv = append(rb.LinkArgv, LinkArgvFromConfig(cfg)...)

		for _, h := range cfg.HelperFiles {
			if h == "" {
				continue
			}
			p := h
			if !filepath.IsAbs(p) {
				if filepath.IsAbs(cwd) {
					p = filepath.Join(cwd, h)
				} else {
					p = h
				}
			}
			// If still not found, try next to the JSON file
			if _, err := os.Stat(p); err != nil && configDir != "" {
				alt := filepath.Join(configDir, h)
				if _, err2 := os.Stat(alt); err2 == nil {
					p = alt
				}
			}
			rb.HelperSources = append(rb.HelperSources, p)
		}
	}

	for _, lib := range libNames {
		if lib == "" {
			continue
		}
		if rb.ConfiguredLibs[lib] {
			continue
		}
		rb.LinkArgv = append(rb.LinkArgv, "-l"+lib)
	}

	rb.IncludePaths = DedupeStringsPreserveOrder(rb.IncludePaths)
	rb.LibraryPaths = DedupeStringsPreserveOrder(rb.LibraryPaths)
	rb.CFlags = DedupeStringsPreserveOrder(rb.CFlags)
	rb.HelperSources = DedupeStringsPreserveOrder(rb.HelperSources)
	rb.LinkArgv = DedupeStringsPreserveOrder(rb.LinkArgv)

	return rb, nil
}

// DedupeStringsPreserveOrder removes duplicate strings while keeping first occurrence order.
func DedupeStringsPreserveOrder(in []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// MissingConfigHint returns a user-facing hint when a header implied a library but no JSON was found.
func MissingConfigHint(libName string) string {
	dirs := ConfigSearchDirs()
	var list string
	for i, d := range dirs {
		if i > 0 {
			list += ", "
		}
		list += filepath.Join(d, libName+".json")
	}
	if list == "" {
		list = "configs/" + libName + ".json"
	}
	return fmt.Sprintf("no library config found for %q (looked for %s). Create one with: cortex -mkconfig %s",
		libName, list, libName)
}
