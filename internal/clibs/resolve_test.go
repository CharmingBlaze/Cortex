package clibs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDedupeStringsPreserveOrder(t *testing.T) {
	in := []string{"-lraylib", "-lm", "-lraylib", "-lopengl32", "-lm"}
	got := DedupeStringsPreserveOrder(in)
	want := []string{"-lraylib", "-lm", "-lopengl32"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestLinkArgvFromConfig_dedupedByResolve(t *testing.T) {
	cfg := &LibraryConfig{
		LinkerFlags: []string{"-lraylib", "-lopengl32"},
		Libraries:   []string{"raylib", "gdi32"},
	}
	argv := LinkArgvFromConfig(cfg)
	// LinkerFlags + -l for Libraries (may duplicate raylib; ResolveLibraries dedupes)
	if len(argv) < 3 {
		t.Fatalf("expected several flags, got %v", argv)
	}
}

func TestResolveLibraries_withConfigAndFallback(t *testing.T) {
	root := t.TempDir()
	cfgDir := filepath.Join(root, "configs")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}
	jsonPath := filepath.Join(cfgDir, "mylib.json")
	content := `{
  "includePaths": ["include/here"],
  "libraryPaths": ["lib/here"],
  "linkerFlags": ["-lmylib"],
  "cflags": ["-DMYLIB"],
  "libraries": [],
  "helperFiles": []
}`
	if err := os.WriteFile(jsonPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	rb, err := ResolveLibraries([]string{"mylib", "pthread"}, root)
	if err != nil {
		t.Fatal(err)
	}
	if !rb.ConfiguredLibs["mylib"] {
		t.Fatal("expected mylib configured")
	}
	if rb.ConfiguredLibs["pthread"] {
		t.Fatal("pthread should not have a config file")
	}
	foundPthread := false
	for _, a := range rb.LinkArgv {
		if a == "-lpthread" {
			foundPthread = true
		}
	}
	if !foundPthread {
		t.Fatalf("expected -lpthread fallback, got %v", rb.LinkArgv)
	}
	if len(rb.IncludePaths) != 1 || rb.IncludePaths[0] != "include/here" {
		t.Fatalf("include paths: %v", rb.IncludePaths)
	}
	// -lmylib once
	count := 0
	for _, a := range rb.LinkArgv {
		if a == "-lmylib" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one -lmylib, got %d in %v", count, rb.LinkArgv)
	}
}

func TestResolveLibraries_raylibJSON_noDuplicateL(t *testing.T) {
	// Use repo configs/raylib.json if present (run from module root)
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Skip(err)
	}
	rayPath := filepath.Join(repoRoot, "configs", "raylib.json")
	if _, err := os.Stat(rayPath); err != nil {
		t.Skip("no repo configs/raylib.json")
	}
	t.Chdir(repoRoot)
	rb, err := ResolveLibraries([]string{"raylib"}, repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, a := range rb.LinkArgv {
		if a == "-lraylib" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected single -lraylib in %v", rb.LinkArgv)
	}
}
