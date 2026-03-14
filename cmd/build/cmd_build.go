package main

import (
	"cortex/internal/build"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	handleBuildCommand(os.Args[1:])
}

// handleBuildCommand processes the 'cortex build' command with optional/manual modes
func handleBuildCommand(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)

	// Mode selection
	useBuild := fs.Bool("use-build", true, "Use integrated build system")
	manual := fs.Bool("manual", false, "Manual mode - disable auto-detection")
	noAutoFetch := fs.Bool("no-autofetch", false, "Don't auto-download libraries")

	// Compiler selection
	compiler := fs.String("compiler", "", "Specific compiler (gcc, clang, tcc)")
	useTCC := fs.Bool("tcc", false, "Force use bundled TCC")
	useMSYS2 := fs.Bool("msys2", false, "Force use MSYS2 (Windows only)")

	// Build options
	verbose := fs.Bool("v", false, "Verbose output")
	clean := fs.Bool("clean", false, "Clean build artifacts")
	release := fs.Bool("release", false, "Release build (optimized)")
	debug := fs.Bool("debug", false, "Debug build (with symbols)")
	run := fs.Bool("run", false, "Run after building")

	// Manual overrides
	sources := fs.String("sources", "", "Manual source files (comma-separated)")
	includes := fs.String("includes", "", "Manual include paths (comma-separated)")
	libraries := fs.String("libs", "", "Manual libraries (comma-separated)")
	output := fs.String("o", "", "Output executable name")
	cflags := fs.String("cflags", "", "Custom compiler flags")
	ldflags := fs.String("ldflags", "", "Custom linker flags")

	// Config file (fallback for complex builds)
	configFile := fs.String("f", "", "Build config file (optional)")

	fs.Parse(args)

	// Check if build system is disabled
	if !*useBuild {
		fmt.Println("Build system disabled (--use-build=false)")
		fmt.Println("Use your system compiler manually:")
		fmt.Println("  gcc main.c -o myapp")
		return
	}

	// Print mode info
	if *manual {
		fmt.Println("=== Manual Mode ===")
		fmt.Println("Auto-detection disabled. Using manual settings only.")
	}

	// Create builder based on options
	var cfg build.Config

	// Load config file if specified
	if *configFile != "" {
		loadedCfg, err := build.LoadConfig(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		cfg = loadedCfg
	} else {
		// Auto-configure
		cfg = build.Config{
			Name:     "app",
			Output:   *output,
			Compiler: *compiler,
		}

		// Apply manual overrides
		if *sources != "" {
			cfg.Sources = strings.Split(*sources, ",")
		}
		if *includes != "" {
			cfg.Includes = strings.Split(*includes, ",")
		}
		if *libraries != "" {
			cfg.Libraries = strings.Split(*libraries, ",")
		}
		if *cflags != "" {
			cfg.CFlags = strings.Split(*cflags, ",")
		}
		if *ldflags != "" {
			cfg.LDFlags = strings.Split(*ldflags, ",")
		}
	}

	// Apply release/debug flags
	if *release {
		cfg.CFlags = append(cfg.CFlags, "-O2", "-DNDEBUG")
	}
	if *debug {
		cfg.CFlags = append(cfg.CFlags, "-g", "-O0")
	}

	// Force specific compiler mode
	if *useTCC {
		cfg.Compiler = "tcc"
	}
	if *useMSYS2 {
		// Will use MSYS2 GCC
		cfg.Compiler = "msys2-gcc"
	}

	// Create and configure builder
	builder := build.NewBuilder(cfg)
	builder.Verbose = *verbose
	builder.Clean = *clean
	builder.Release = *release
	builder.Manual = *manual
	builder.NoAutoFetch = *noAutoFetch

	// Execute build
	fmt.Printf("Building %s...\n", cfg.Name)
	if err := builder.Build(); err != nil {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", err)
		os.Exit(1)
	}

	// Run if requested
	if *run {
		fmt.Printf("Running %s...\n", cfg.Output)
		if err := runExecutable(cfg.Output); err != nil {
			fmt.Fprintf(os.Stderr, "Run failed: %v\n", err)
			os.Exit(1)
		}
	}
}

func runExecutable(path string) error {
	// Implementation would run the built executable
	return nil
}
