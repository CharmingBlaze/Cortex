package binder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// PreprocessResult holds the outcome of an optional external preprocessor run.
type PreprocessResult struct {
	Output  []byte
	Tool    string
	Warning string
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func findBundledZig() string {
	exeName := "zig"
	if isWindows() {
		exeName = "zig.exe"
	}
	if exe, err := os.Executable(); err == nil {
		binDir := filepath.Dir(exe)
		releaseDir := filepath.Dir(binDir)
		zigPath := filepath.Join(releaseDir, "zig", exeName)
		if _, err := os.Stat(zigPath); err == nil {
			return zigPath
		}
		toolsZig := filepath.Join(binDir, "tools", exeName)
		if _, err := os.Stat(toolsZig); err == nil {
			return toolsZig
		}
	}
	return ""
}

func findZig() string {
	exeName := "zig"
	if isWindows() {
		exeName = "zig.exe"
	}
	if p, err := exec.LookPath(exeName); err == nil {
		return p
	}
	return ""
}

func absPaths(dirs []string) ([]string, error) {
	out := make([]string, 0, len(dirs))
	for _, d := range dirs {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		a, err := filepath.Abs(d)
		if err != nil {
			return nil, fmt.Errorf("include path %q: %w", d, err)
		}
		out = append(out, a)
	}
	return out, nil
}

// cppArgs builds common -E -P arguments (include dirs, defines, input file).
func cppArgs(headerAbs string, includeDirs, defines []string) ([]string, error) {
	inc, err := absPaths(includeDirs)
	if err != nil {
		return nil, err
	}
	args := []string{"-E", "-P"}
	for _, p := range inc {
		args = append(args, "-I", p)
	}
	for _, d := range defines {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		if strings.HasPrefix(d, "-D") {
			args = append(args, d)
		} else {
			args = append(args, "-D"+d)
		}
	}
	args = append(args, headerAbs)
	return args, nil
}

// RunCPP runs an external C preprocessor (zig cc -E, gcc -E, or clang -E).
// On failure it returns Warning set and empty Output so the caller can fall back.
func RunCPP(headerAbs string, includeDirs, defines []string) PreprocessResult {
	args, err := cppArgs(headerAbs, includeDirs, defines)
	if err != nil {
		return PreprocessResult{Warning: fmt.Sprintf("bind preprocess: %v; using raw header", err)}
	}

	run := func(cmd *exec.Cmd) ([]byte, bool) {
		cmd.Env = os.Environ()
		out, err := cmd.Output()
		if err != nil {
			return nil, false
		}
		return out, true
	}

	if z := findBundledZig(); z != "" {
		cmd := exec.Command(z, append([]string{"cc"}, args...)...)
		if out, ok := run(cmd); ok {
			return PreprocessResult{Output: out, Tool: "zig cc (bundled)"}
		}
	}
	if z := findZig(); z != "" {
		cmd := exec.Command(z, append([]string{"cc"}, args...)...)
		if out, ok := run(cmd); ok {
			return PreprocessResult{Output: out, Tool: "zig cc"}
		}
	}
	if out, ok := run(exec.Command("gcc", args...)); ok {
		return PreprocessResult{Output: out, Tool: "gcc"}
	}
	if out, ok := run(exec.Command("clang", args...)); ok {
		return PreprocessResult{Output: out, Tool: "clang"}
	}

	return PreprocessResult{
		Warning: "bind preprocess: no working zig cc/gcc/clang -E found; parsing raw header (limited macros/includes)",
	}
}
