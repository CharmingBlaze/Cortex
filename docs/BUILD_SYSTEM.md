# Cortex Build System

A simple, declarative build system for C libraries that follows Cortex's philosophy of simplicity.

## Quick Start

Create a `build.cx` file in your project root:

```c
// build.cx - Cortex build configuration
build {
    name: "myproject"
    version: "1.0.0"
    
    // C source files to compile
    sources: ["src/*.c", "src/**/*.c"]
    
    // Include directories
    includes: ["include", "third_party/include"]
    
    // Libraries to link
    libraries: ["raylib", "m"]
    
    // Library search paths
    lib_paths: ["/usr/local/lib", "C:/raylib/lib"]
    
    // Output executable
    output: "myapp"
    
    // C compiler (optional - auto-detected: gcc, tcc, clang)
    compiler: "gcc"
    
    // Compiler flags
    cflags: ["-O2", "-Wall"]
}

// External dependencies
dependency {
    name: "raylib"
    version: ">=4.0"
    
    // Auto-download or use system library
    source: "github:raysan5/raylib"
    
    // Or specify local path
    // path: "../raylib"
    
    // Include/library paths for this dependency
    includes: ["include"]
    lib_paths: ["lib"]
    libraries: ["raylib"]
}

// Library target (build as static library)
library {
    name: "mylib"
    type: "static"  // or "shared"
    sources: ["lib/*.c"]
    includes: ["include"]
}
```

## Build Commands

```bash
# Build the project
cortex build

# Build and run
cortex build -run

# Clean build artifacts
cortex build -clean

# Build specific target
cortex build -target mylib

# Verbose output
cortex build -v

# Release build (optimized)
cortex build -release

# Debug build (with symbols)
cortex build -debug
```

## Simple Example

**build.cx:**
```c
build {
    name: "game"
    sources: ["main.c", "game.c"]
    includes: ["include"]
    libraries: ["raylib"]
    output: "game.exe"
}
```

**Build:**
```bash
cortex build -run
```

## Features

- **Simple syntax**: Familiar C-like braces and colons
- **Glob patterns**: `src/*.c` automatically finds all .c files
- **Auto-detection**: Finds compiler (gcc, tcc, clang) automatically
- **Cross-platform**: Works on Windows, macOS, Linux
- **Incremental builds**: Only rebuilds changed files
- **Dependency resolution**: Can fetch common libraries automatically
- **Integration**: Uses existing Cortex C-interop system

## How It Works

1. Parse `build.cx` using Cortex parser
2. Resolve dependencies (download if needed)
3. Generate C compilation commands
4. Compile incrementally (cache object files)
5. Link final executable or library

## Integration with Cortex Compiler

The build system integrates with the existing Cortex compiler:
- Uses same `#include` / `#use` syntax for C libraries
- Reuses the library inference system from `internal/clibs`
- Supports `#pragma link` for automatic linking
