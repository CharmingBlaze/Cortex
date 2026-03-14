package main

import (
	"cortex/internal/compiler"
	"cortex/internal/config"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// stringList allows multiple -I, -L, -l on the command line.
type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, " ")
}

func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func main() {
	var inputFiles stringList
	var outputFile string
	var runMode bool
	var help bool
	var configPath string
	var useLib string
	var backend string
	var includePaths stringList
	var libraryPaths stringList
	var libraries stringList
	var debug bool

	flag.Var(&inputFiles, "i", "Input .cx source file (repeat for multi-file)")
	flag.StringVar(&outputFile, "o", "", "Output executable file")
	flag.BoolVar(&runMode, "run", false, "Compile and run (no exe left behind; uses temp file)")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.StringVar(&configPath, "config", "", "Path to cortex config (optional)")
	flag.StringVar(&useLib, "use", "", "Use a C library by name (loads configs/<name>.json if no -config); e.g. -use raylib")
	flag.StringVar(&backend, "backend", "", "C backend: gcc, tcc, or auto (default: auto = try tcc then gcc)")
	flag.Var(&includePaths, "I", "Include path for C headers (can be repeated)")
	flag.Var(&libraryPaths, "L", "Library search path (can be repeated)")
	flag.Var(&libraries, "l", "Library to link, e.g. raylib (can be repeated)")
	flag.BoolVar(&debug, "debug", false, "Enable debug output")
	flag.Parse()

	if help || len(inputFiles) == 0 {
		fmt.Println("Cortex Compiler - A C-like language for games and applications")
		fmt.Println("Usage:")
		fmt.Println("  cortex -i <input_file> [-o <output_file>] [-config <config_file>]")
		fmt.Println("  cortex -i main.cx -i lib.cx -o app")
		fmt.Println("  cortex -i game.cx -o game -I /path/to/raylib/include -L /path/to/raylib/lib -l raylib")
		fmt.Println("  cortex -help")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  -i        Input .cx source file (required; repeat for multi-file / package)")
		fmt.Println("  -o        Output executable (default: <first_input_base>.exe)")
		fmt.Println("  -run      Compile and run (temp exe, then delete); no gcc/cmake if tcc in tools/ or PATH")
		fmt.Println("  -config   Optional JSON config (features, backend, include_paths, library_paths, libraries)")
		fmt.Println("  -use      Use C library by name (e.g. -use raylib loads configs/raylib.json); combine with #use \"name\" in source")
		fmt.Println("  -backend  C backend: gcc, tcc, or auto (no gcc/cmake: use tcc; put tcc in tools/ or PATH)")
		fmt.Println("  -I        Add include path (repeat for multiple); for C libraries like raylib")
		fmt.Println("  -L        Add library search path (repeat for multiple)")
		fmt.Println("  -l        Link library (repeat for multiple); e.g. -l raylib -l m")
		return
	}

	for _, f := range inputFiles {
		if !exists(f) {
			fmt.Fprintf(os.Stderr, "Error: Input file '%s' does not exist\n", f)
			os.Exit(1)
		}
	}

	if outputFile == "" {
		base := filepath.Base(inputFiles[0])
		outputFile = strings.TrimSuffix(base, filepath.Ext(base)) + ".exe"
	}
	if runMode {
		outputFile = filepath.Join(os.TempDir(), "cortex_run_"+strings.TrimSuffix(filepath.Base(inputFiles[0]), filepath.Ext(filepath.Base(inputFiles[0])))+".exe")
	}

	if useLib != "" && configPath == "" {
		configPath = filepath.Join("configs", useLib+".json")
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	cfg.IncludePaths = append(cfg.IncludePaths, includePaths...)
	cfg.LibraryPaths = append(cfg.LibraryPaths, libraryPaths...)
	cfg.Libraries = append(cfg.Libraries, libraries...)
	if backend != "" {
		cfg.Backend = backend
	}
	cfg.Debug = debug
	if cfg.Backend == "" {
		cfg.Backend = "auto"
	}

	compiler := compiler.NewCompiler(cfg)

	if err := compiler.CompileMulti(inputFiles, outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if runMode {
		cmd := exec.Command(outputFile)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			if exit, ok := err.(*exec.ExitError); ok {
				os.Exit(exit.ExitCode())
			}
			fmt.Fprintf(os.Stderr, "run failed: %v\n", err)
			os.Exit(1)
		}
		os.Remove(outputFile)
		return
	}

	fmt.Printf("Successfully compiled %d file(s) to '%s'\n", len(inputFiles), outputFile)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
