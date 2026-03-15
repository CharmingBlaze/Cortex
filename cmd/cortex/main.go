package main

import (
	"cortex/internal/binder"
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
	// Check for subcommands (build, run, bind, new)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "build":
			buildCommand(os.Args[2:])
			return
		case "run":
			runCommand(os.Args[2:])
			return
		case "bind":
			bindCommand(os.Args[2:])
			return
		case "new":
			newCommand(os.Args[2:])
			return
		}
	}

	// Legacy flag-based mode
	legacyMode()
}

// buildCommand handles: cortex build [file.cx]
func buildCommand(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	var outputFile string
	var debug bool
	fs.StringVar(&outputFile, "o", "", "Output executable file")
	fs.BoolVar(&debug, "debug", false, "Enable debug output")
	fs.Parse(args)

	inputFile := ""
	if fs.NArg() > 0 {
		inputFile = fs.Arg(0)
	}

	// Find project directory and load cortex.toml
	cwd, _ := os.Getwd()
	projectDir, err := config.FindProjectDir(cwd)
	if err != nil {
		// No cortex.toml - use legacy mode
		if inputFile == "" {
			fmt.Fprintln(os.Stderr, "Error: no input file and no cortex.toml found")
			fmt.Fprintln(os.Stderr, "Usage: cortex build <file.cx>")
			os.Exit(1)
		}
		compileFile(inputFile, outputFile, nil, debug)
		return
	}

	// Load project config
	proj, err := config.LoadProject(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Determine input file
	if inputFile == "" {
		inputFile = proj.Project.Entry
	}

	// Convert project config to compiler config
	cfg := proj.ToConfig()
	cfg.Debug = debug

	compileFile(inputFile, outputFile, cfg, debug)
}

// runCommand handles: cortex run [file.cx]
func runCommand(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	var debug bool
	fs.BoolVar(&debug, "debug", false, "Enable debug output")
	fs.Parse(args)

	inputFile := ""
	if fs.NArg() > 0 {
		inputFile = fs.Arg(0)
	}

	// Find project directory and load cortex.toml
	cwd, _ := os.Getwd()
	projectDir, err := config.FindProjectDir(cwd)
	if err != nil {
		// No cortex.toml - use legacy mode
		if inputFile == "" {
			fmt.Fprintln(os.Stderr, "Error: no input file and no cortex.toml found")
			fmt.Fprintln(os.Stderr, "Usage: cortex run <file.cx>")
			os.Exit(1)
		}
		runFile(inputFile, nil, debug)
		return
	}

	// Load project config
	proj, err := config.LoadProject(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Determine input file
	if inputFile == "" {
		inputFile = proj.Project.Entry
	}

	// Convert project config to compiler config
	cfg := proj.ToConfig()
	cfg.Debug = debug

	runFile(inputFile, cfg, debug)
}

// newCommand handles: cortex new <project_name>
func newCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: cortex new <project_name>")
		os.Exit(1)
	}

	name := args[0]

	// Create project directory
	if err := os.MkdirAll(name, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating project: %v\n", err)
		os.Exit(1)
	}

	// Create cortex.toml
	tomlContent := fmt.Sprintf(`[project]
name = "%s"
version = "0.1.0"
entry = "main.cx"

[dependencies]
# Add libraries here, e.g.:
# raylib = { path = "third_party/raylib" }
`, name)

	if err := os.WriteFile(filepath.Join(name, "cortex.toml"), []byte(tomlContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating cortex.toml: %v\n", err)
		os.Exit(1)
	}

	// Create main.cx
	mainContent := `// Cortex Project
void main() {
    println("Hello from Cortex!");
}
`
	if err := os.WriteFile(filepath.Join(name, "main.cx"), []byte(mainContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating main.cx: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created project '%s'\n", name)
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", name)
	fmt.Println("  cortex run")
}

// bindCommand handles: cortex bind <libname> -i <header.h> [-o output.cx]
func bindCommand(args []string) {
	fs := flag.NewFlagSet("bind", flag.ExitOnError)
	var headerPath string
	var outputPath string
	fs.StringVar(&headerPath, "i", "", "Input C header file")
	fs.StringVar(&outputPath, "o", "", "Output Cortex binding file")

	// Parse flags first (they can appear anywhere with ParseAll)
	err := fs.Parse(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Usage: cortex bind <libname> -i <header.h> [-o output.cx]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  cortex bind raylib -i third_party/raylib/src/raylib.h")
		fmt.Fprintln(os.Stderr, "  cortex bind mylib -i mylib.h -o bindings/mylib.cx")
		os.Exit(1)
	}

	libName := fs.Arg(0)

	// Find header file
	if headerPath == "" {
		// Try common locations
		searchPaths := []string{
			filepath.Join("third_party", libName, "include", libName+".h"),
			filepath.Join("third_party", libName, "src", libName+".h"),
			filepath.Join("include", libName+".h"),
			libName + ".h",
		}
		for _, p := range searchPaths {
			if _, err := os.Stat(p); err == nil {
				headerPath = p
				break
			}
		}
		if headerPath == "" {
			fmt.Fprintf(os.Stderr, "Error: Could not find header for '%s'. Use -i to specify path.\n", libName)
			os.Exit(1)
		}
	}

	if !exists(headerPath) {
		fmt.Fprintf(os.Stderr, "Error: Header file '%s' does not exist\n", headerPath)
		os.Exit(1)
	}

	// Determine output path
	if outputPath == "" {
		outputPath = filepath.Join("bindings", libName+".cx")
	}

	fmt.Printf("Binding %s from %s...\n", libName, headerPath)

	// Create binder
	b := binder.NewBinder(libName)

	// Parse header
	if err := b.ParseHeader(headerPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing header: %v\n", err)
		os.Exit(1)
	}

	// Generate bindings
	if err := b.SaveToFile(outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving bindings: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated bindings: %s\n", outputPath)
	fmt.Println("")
	fn, st, en, dc := b.Stats()
	fmt.Println("Functions found:", fn)
	fmt.Println("Structs found:", st)
	fmt.Println("Enums found:", en)
	fmt.Println("Constants found:", dc)
}

// compileFile compiles a single file to an executable
func compileFile(inputFile, outputFile string, cfg *config.Config, debug bool) {
	if !exists(inputFile) {
		fmt.Fprintf(os.Stderr, "Error: Input file '%s' does not exist\n", inputFile)
		os.Exit(1)
	}

	if outputFile == "" {
		base := filepath.Base(inputFile)
		outputFile = strings.TrimSuffix(base, filepath.Ext(base)) + ".exe"
	}

	if cfg == nil {
		cfg = &config.Config{}
		cfg.Features = config.DefaultFeatures()
		cfg.Backend = "auto"
	}
	cfg.Debug = debug

	compiler := compiler.NewCompiler(*cfg)

	if err := compiler.CompileMulti([]string{inputFile}, outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully compiled to '%s'\n", outputFile)
}

// runFile compiles and runs a file (temp exe, then delete)
func runFile(inputFile string, cfg *config.Config, debug bool) {
	if !exists(inputFile) {
		fmt.Fprintf(os.Stderr, "Error: Input file '%s' does not exist\n", inputFile)
		os.Exit(1)
	}

	// Create temp output
	outputFile := filepath.Join(os.TempDir(), "cortex_run_"+strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(filepath.Base(inputFile)))+".exe")

	if cfg == nil {
		cfg = &config.Config{}
		cfg.Features = config.DefaultFeatures()
		cfg.Backend = "auto"
	}
	cfg.Debug = debug

	compiler := compiler.NewCompiler(*cfg)

	if err := compiler.CompileMulti([]string{inputFile}, outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Run the executable
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
}

// legacyMode handles the old flag-based CLI
func legacyMode() {
	var inputFiles stringList
	var outputFile string
	var runMode bool
	var help bool
	var configPath string
	var useLib string
	var mkConfig string
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
	flag.StringVar(&mkConfig, "mkconfig", "", "Create a config template for a library (e.g. -mkconfig raylib)")
	flag.StringVar(&backend, "backend", "", "C backend: gcc, tcc, or auto (default: auto = try tcc then gcc)")
	flag.Var(&includePaths, "I", "Include path for C headers (can be repeated)")
	flag.Var(&libraryPaths, "L", "Library search path (can be repeated)")
	flag.Var(&libraries, "l", "Library to link, e.g. raylib (can be repeated)")
	flag.BoolVar(&debug, "debug", false, "Enable debug output")
	flag.Parse()

	// Handle -mkconfig: create a config template and exit
	if mkConfig != "" {
		createConfigTemplate(mkConfig)
		return
	}

	if help || len(inputFiles) == 0 {
		fmt.Println("Cortex Compiler - A C-like language for games and applications")
		fmt.Println("Usage:")
		fmt.Println("  cortex -i <input_file> [-o <output_file>] [-config <config_file>]")
		fmt.Println("  cortex -i main.cx -i lib.cx -o app")
		fmt.Println("  cortex -i game.cx -o game -use raylib")
		fmt.Println("  cortex -mkconfig raylib    # Create configs/raylib.json template")
		fmt.Println("  cortex -help")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  -i         Input .cx source file (required; repeat for multi-file / package)")
		fmt.Println("  -o         Output executable (default: <first_input_base>.exe)")
		fmt.Println("  -run       Compile and run (temp exe, then delete); no gcc/cmake if tcc in tools/ or PATH")
		fmt.Println("  -config    Optional JSON config (features, backend, include_paths, library_paths, libraries)")
		fmt.Println("  -use       Use C library by name (e.g. -use raylib loads configs/raylib.json)")
		fmt.Println("  -mkconfig  Create a config template for a library (e.g. -mkconfig mylib)")
		fmt.Println("  -backend   C backend: gcc, tcc, or auto (no gcc/cmake: use tcc; put tcc in tools/ or PATH)")
		fmt.Println("  -I         Add include path (repeat for multiple); for C libraries like raylib")
		fmt.Println("  -L         Add library search path (repeat for multiple)")
		fmt.Println("  -l         Link library (repeat for multiple); e.g. -l raylib -l m")
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

// createConfigTemplate creates a JSON config template for a library
func createConfigTemplate(libName string) {
	// Ensure configs directory exists
	configsDir := "configs"
	if _, err := os.Stat(configsDir); os.IsNotExist(err) {
		os.MkdirAll(configsDir, 0755)
	}

	configFile := filepath.Join(configsDir, libName+".json")

	// Check if config already exists
	if _, err := os.Stat(configFile); err == nil {
		fmt.Printf("Config already exists: %s\n", configFile)
		fmt.Println("Edit it to set your library paths.")
		return
	}

	// Create template
	template := fmt.Sprintf(`{
  "features": { "qol": true },
  "include_paths": [
    "C:/%s/include",
    "/usr/local/include",
    "/usr/include"
  ],
  "library_paths": [
    "C:/%s/lib",
    "/usr/local/lib",
    "/usr/lib"
  ],
  "libraries": ["%s"]
}
`, libName, libName, libName)

	err := os.WriteFile(configFile, []byte(template), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created: %s\n", configFile)
	fmt.Println("")
	fmt.Println("Edit the file to set your library paths, then use:")
	fmt.Printf("  cortex -i game.cx -o game -use %s\n", libName)
}
