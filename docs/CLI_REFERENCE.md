# Cortex CLI Reference

Complete reference for all Cortex command-line interface commands.

## Quick Reference

| Command | Purpose | Example |
|---------|---------|---------|
| `cortex new` | Create new project | `cortex new mygame` |
| `cortex run` | Compile and run | `cortex run main.cx` |
| `cortex build` | Build executable | `cortex build main.cx -o app` |
| `cortex bind` | Generate bindings | `cortex bind raylib -i raylib.h` |
| `cortex -help` | Show help | `cortex -help` |

---

## `cortex new` - Create New Project

Creates a new Cortex project with proper structure.

### Usage

```bash
cortex new <project_name> [raylib]
```

With the optional **`raylib`** argument, Cortex also creates **`configs/raylib.json`** (paths defaulting to **`third_party/raylib`**) and a minimal **`#include <raylib.h>`** window example in **`main.cx`**.

### What It Creates

```
myproject/
├── cortex.toml      # Project configuration
├── main.cx          # Entry point
├── src/             # Source directory
│   └── .gitkeep
└── .gitignore       # Git ignore rules
```

### Example

```bash
$ cortex new mygame
Creating project 'mygame'...
  ✓ Created cortex.toml
  ✓ Created main.cx
  ✓ Created src/
  
Project created! Next steps:
  cd mygame
  cortex run
```

### cortex.toml Template

```toml
[project]
name = "mygame"
version = "0.1.0"
entry = "main.cx"
backend = "auto"

[project.features]
async = true
actors = true
qol = true
```

---

## `cortex run` - Compile and Run

Compiles a Cortex program and runs it immediately. Perfect for development.

### Usage

```bash
# Run a single file
cortex run <file.cx>

# Run a project (uses cortex.toml)
cortex run

# Pass arguments to the program
cortex run <file.cx> -- <args...>

# Keep generated C code for inspection
cortex run <file.cx> --keep-c
```

### Options

| Option | Description |
|--------|-------------|
| `-debug` | Enable debug output |

### Examples

```bash
# Run a single file
cortex run hello.cx
# Output: Hello, World!

# Run with debug output
cortex run main.cx -debug

# Run project
cd myproject
cortex run
```

### What Happens

1. Cortex parses your `.cx` file
2. Generates C code in a temporary file
3. Invokes C compiler (Zig CC bundled, or system GCC/Clang)
4. Runs the resulting executable
5. Cleans up temp files

---

## `cortex build` - Build Executable

Compiles a Cortex program into a standalone executable.

### Usage

```bash
# Build with default name (removes .cx extension)
cortex build <file.cx>

# Specify output name
cortex build <file.cx> -o <output_name>

# Build project
cortex build [-o <output_name>]

# Release build (optimized)
cortex build <file.cx> -o <output> --release
```

### Options

| Option | Description |
|--------|-------------|
| `-o <name>` | Output executable name |
| `-debug` | Enable debug output |

### Examples

```bash
# Simple build
cortex build hello.cx
# Creates: hello.exe (Windows) or hello (Linux/macOS)

# Named output
cortex build main.cx -o mygame
# Creates: mygame.exe or mygame

# With debug output
cortex build main.cx -debug
```

### Build Process

```
main.cx → Cortex Compiler → main.c → C Compiler → main.exe
```

---

## `cortex bind` - Generate C Library Bindings

Generates Cortex bindings from C header files for better IDE support and type safety.

### Usage

```bash
cortex bind <libname> -i <header.h>
```

### What It Does

1. Parses C header file
2. Extracts functions, structs, enums, constants
3. Generates Cortex-compatible `.cx` file

### Output Location

Creates `bindings/<libname>.cx`

### Examples

```bash
# Generate raylib bindings
cortex bind raylib -i third_party/raylib/src/raylib.h
# Creates: bindings/raylib.cx

# Generate SDL2 bindings
cortex bind sdl2 -i /usr/include/SDL2/SDL.h
# Creates: bindings/sdl2.cx

# Generate from multiple headers
cortex bind mylib -i include/mylib.h
```

### What Gets Converted

| C Construct | Cortex Output |
|-------------|---------------|
| `void func(int x);` | `extern void func(int x);` |
| `typedef struct {...} S;` | `struct S {...}` |
| `enum { A, B };` | `const int A = 0; const int B = 1;` |
| `#define VALUE 100` | `const int VALUE = 100;` |
| `#define FUNC(x) ((x)*2)` | `// Macro (commented, needs manual binding)` |

### Using Generated Bindings

```c
// main.cx
#include "bindings/raylib.cx"

void main() {
    InitWindow(800, 450, "Game");
    // Cortex now knows InitWindow signature
}
```

---

## Global Options

### `-help`

Show help:

```bash
cortex -help
```

### `-debug`

Show detailed output:

```bash
cortex run main.cx -debug
```

---

## Legacy Commands (Flag-based)

Old-style commands still work for backward compatibility:

```bash
# Run a file
cortex -i main.cx -run

# Build with output
cortex -i main.cx -o output.exe

# Build with library
cortex -i main.cx -o output.exe -use raylib

# Add include paths
cortex -i main.cx -I ./include -I /usr/local/include

# Add library paths
cortex -i main.cx -L ./lib -L /usr/local/lib

# Link libraries
cortex -i main.cx -l raylib -l opengl32

# Stricter semantic checks (e.g. no shadowing outer declarations)
cortex -strict -i main.cx -o app.exe
```

### `cortex.toml` — `strict`

```toml
[project]
name = "my_game"
entry = "main.cx"
strict = true
```

When `strict = true`, the compiler rejects bindings that **shadow** an outer variable, function, or parameter (same spirit as stricter C-adjacent style guides).

---

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `CORTEX_PATH` | Additional search path for libraries |
| `ZIG_PATH` | Path to Zig compiler (if not bundled) |
| `CC` | C compiler to use (default: auto-detect) |

### Examples

```bash
# Use specific C compiler
export CC=clang
cortex build main.cx

# Set Zig path (if using system Zig)
export ZIG_PATH=/usr/local/bin/zig
cortex run main.cx
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Compilation error |
| 2 | Runtime error |
| 3 | File not found |
| 4 | Invalid arguments |

---

## Configuration File (cortex.toml)

Project configuration file in TOML format:

```toml
[project]
name = "mygame"
version = "0.1.0"
entry = "main.cx"           # Entry point file
backend = "auto"            # gcc, zig, or auto

[project.features]
async = true                # Enable async/await
actors = true               # Enable actor model
qol = true                  # Quality-of-life features

[dependencies.raylib]
path = "third_party/raylib" # Library path
# Or explicit paths:
# include_path = "third_party/raylib/src"
# lib_path = "third_party/raylib/build/raylib"
libs = ["raylib", "opengl32", "gdi32", "winmm", "shell32"]

[build]
flags = ["-O2"]             # Additional C compiler flags
defines = ["DEBUG"]         # Preprocessor defines
```

### Using cortex.toml

```bash
# In project directory with cortex.toml
cortex run        # Uses entry from config
cortex build      # Uses all dependencies
```

---

## Tips and Tricks

### Faster Development Iterations

```bash
# Use run for development (faster, no output file)
cortex run main.cx

# Use build only for final executable
cortex build main.cx -o mygame
```

### Debug Generated C Code

```bash
# Use -debug flag to see what Cortex generates
cortex run main.cx -debug
```

### Build for Different Platforms

Cortex generates C code, so you can cross-compile:

```bash
# Generate C code manually by inspecting temp files
# Then cross-compile with GCC
x86_64-w64-mingw32-gcc main.c -o main.exe
arm-linux-gnueabihf-gcc main.c -o main.arm
```

### Using Libraries

```bash
# Use -use flag for library configs
cortex -i main.cx -o game.exe -use raylib

# Or add include/library paths manually
cortex -i main.cx -I ./include -L ./lib -l mylib
```

---

## Common Workflows

### Starting a New Project

```bash
cortex new myproject
cd myproject
cortex run
```

### Adding a C Library

```bash
# 1. Download library
mkdir -p third_party/mylib
# ... copy library files ...

# 2. Generate bindings
cortex bind mylib -i third_party/mylib/mylib.h

# 3. Add to cortex.toml
[dependencies.mylib]
path = "third_party/mylib"
libs = ["mylib"]

# 4. Use in code
# main.cx:
#include "bindings/mylib.cx"
```

### Building a Game

```bash
# Development
cortex run game.cx

# Build final executable
cortex build game.cx -o game
./game

# Distribute
# Upload game.exe + any DLLs needed
```

---

## See Also

- [Beginner's Guide](BEGINNERS_GUIDE.md) - Getting started tutorial
- [Language Spec](LANGUAGE_SPEC.md) - Language reference
- [Examples](../examples/) - Sample programs
