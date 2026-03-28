# C Library Binding Guide

This guide explains how to use C libraries in Cortex and generate bindings for better type safety and IDE support.

## Table of Contents

1. [Why Use C Libraries?](#why-use-c-libraries)
2. [Quick Start: Using a C Library](#quick-start-using-a-c-library)
3. [The Two Approaches](#the-two-approaches)
4. [Generating Bindings](#generating-bindings)
5. [What Gets Converted](#what-gets-converted)
6. [Manual Binding](#manual-binding)
7. [Common Libraries](#common-libraries)
8. [Troubleshooting](#troubleshooting)

---

## Why Use C Libraries?

Cortex compiles to C, which means you can use **any C library** directly. This gives you:

- **Thousands of libraries** - Graphics, networking, databases, etc.
- **Native performance** - No wrappers or FFI overhead
- **Battle-tested code** - Use mature, well-tested libraries

Popular libraries you can use:
- **raylib** - Game development
- **SDL2** - Cross-platform multimedia
- **OpenGL** - 3D graphics
- **libcurl** - HTTP/networking
- **SQLite** - Embedded database
- **FFmpeg** - Audio/video processing

---

## Quick Start: Using a C Library

### Step 1: Get the Library

```bash
# Windows - download pre-built binaries
# Linux - use package manager
sudo apt install libraylib-dev

# macOS - use homebrew
brew install raylib
```

### Step 2: Include the Header

```c
// main.cx
#include <raylib.h>

void main() {
    InitWindow(800, 450, "My Window");
    // Use raylib functions...
    CloseWindow();
}
```

### Step 3: Library paths (one small JSON per library)

Cortex infers the library name from your `#include` (for example `raylib.h` → `raylib`) and loads **`configs/raylib.json`** automatically when it exists. That file lists `includePaths`, `libraryPaths`, and link flags — same idea as `-I` / `-L` / `-l`, but you set them once.

**Default workflow (recommended):**

```bash
# From a project/repo that has configs/raylib.json (or create one — see below)
cortex run main.cx
```

**Optional overrides** (same as before): `-use raylib` loads `configs/raylib.json` into the legacy top-level config, or pass `-I` / `-L` / `-l` by hand.

**No config yet?** Create a template and edit the paths for your machine:

```bash
cortex -mkconfig mylib
```

This repository ships **`configs/raylib.json`** and **`configs/sdl2.json`** at the project root — copy or adjust them for your layout (for example `third_party/raylib`).

**Starter project with raylib:**

```bash
cortex new my_game raylib
cd my_game
# Point configs/raylib.json at your raylib build, then:
cortex run
```

That's it! You're using a C library in Cortex.

### Header-to-library inference

Cortex maps common `#include` names to link names (for example `raylib.h` → `raylib`, `SDL.h` / `SDL2/SDL.h` → `sdl2`, `curl/curl.h` → `curl`, `sqlite3.h` → `sqlite3`, `GLFW/glfw3.h` → `glfw`, and more). Unknown headers use the **basename** (`foo.h` → `foo`). The table lives in the compiler source [`internal/clibs/inference.go`](../internal/clibs/inference.go) — extend it when you add first-class support for another library.

**Pointer-free interop** (what `.cx` code should look like vs generated C) is described in [POINTER_FREE_AND_FFI.md](POINTER_FREE_AND_FFI.md).

---

## The Two Approaches

### Approach 1: Direct Include (Simple)

Just include the C header and use the functions:

```c
#include <raylib.h>

void main() {
    InitWindow(800, 450, "Game");
    DrawCircle(400, 225, 50, RED);
    CloseWindow();
}
```

**Pros:**
- Simple, no extra steps
- Works with any C header

**Cons:**
- No type checking by Cortex
- No IDE autocomplete
- Errors may be cryptic

### Approach 2: Generated Bindings (Recommended)

Generate Cortex bindings first:

```bash
cortex bind raylib -i /usr/include/raylib.h
```

Then use them:

```c
#include "bindings/raylib.cx"

void main() {
    InitWindow(800, 450, "Game");
    // Now Cortex knows the function signatures
}
```

**Pros:**
- Type checking at compile time
- Better error messages
- IDE support works better
- Documentation preserved

**Cons:**
- Extra step to generate bindings
- May need manual fixes for complex macros

---

## Generating Bindings

### Basic Usage

```bash
cortex bind <library_name> -i <header_file>
```

### Examples

```bash
# Raylib
cortex bind raylib -i third_party/raylib/src/raylib.h

# SDL2 (include path + generated #include line)
cortex bind sdl2 -i /usr/include/SDL2/SDL.h -I /usr/include/SDL2 -include SDL2/SDL.h

# Regex-only (no host gcc/clang required)
cortex bind mylib -i include/mylib.h -legacy-bind

# Your own library
cortex bind mylib -i include/mylib.h
```

### Output

Creates `bindings/<library_name>.cx`:

```
your_project/
├── bindings/
│   └── raylib.cx    # Generated bindings
├── main.cx
└── cortex.toml
```

### What the Binder Does

By default `cortex bind` uses a **preprocessor + C AST** pipeline ([modernc.org/cc/v3](https://pkg.go.dev/modernc.org/cc/v3)):

1. **Optional external preprocess** — if `zig cc`, `gcc`, or `clang` is available, the header is run through `cc -E -P` with your `-I` / `-D` flags so macros and includes match your real toolchain. If none is found, a warning is printed and the **raw header** is parsed (macros and includes are then handled only by the internal parser, which is weaker for heavy system headers).
2. **Parse & type-check** — the translation unit is built with host `cpp`-style predefined macros (`CC`, `gcc`, `clang`, or `cpp` via `HostConfig`) plus a small built-in prelude (derived from the cc test suite).
3. **Lower to the binder model** — file-scope function prototypes, typedefs, structs, and enums are mapped into Cortex-facing structs.
4. **Codegen** — `GenerateCortex()` emits a `.cx` file; object-like `#define` lines are still scraped from the **original** header text (best effort).

#### Flags

| Flag | Meaning |
|------|---------|
| `-i path` | Input header (required unless a default path is found). |
| `-o path` | Output `.cx` (default `bindings/<lib>.cx`). |
| `-I dir` | Include directory for preprocess / `#include` resolution (repeatable). |
| `-D NAME` / `-D NAME=value` | Preprocessor define (repeatable). |
| `-include hdr` | Override the generated `#include` line (e.g. `SDL2/SDL.h`). |
| `-no-preprocess` | Skip external `zig cc` / `gcc` / `clang -E` (internal CPP only). |
| `-legacy-bind` | **Regex-only** parser (previous behavior). Does not need a host C toolchain; accuracy is lower on real headers. |

**Requirements (AST mode):** a POSIX-style C preprocessor must be discoverable for predefined macros and system include paths (typically `gcc` or `clang` on Linux/macOS; on Windows install a toolchain or set `CC` to a working compiler). If that fails, use `-legacy-bind` or install `gcc`/`clang`.

---

## What Gets Converted

AST mode understands typical **function prototypes**, **typedef** names, **struct** definitions with named fields, and **enum** definitions with enumerators. It does **not** fully model C++ templates, macro-generated declarations that are not visible as C declarations after preprocessing, or function-like macros as Cortex functions.

**Limitations:**

- **Function pointers** are emitted as `void*` with `// TODO` comments; you may need `@c` or hand-written prototypes.
- **Multi-level pointers** are collapsed to a single opaque `void*` where the binder does not preserve arity.
- **`const char*`** is mapped to `string`; other pointers stay `void*` with comments pointing to [POINTER_FREE_AND_FFI.md](POINTER_FREE_AND_FFI.md) where relevant.

### Functions

**C Header:**
```c
void InitWindow(int width, int height, const char* title);
int GetScreenWidth(void);
Color GetColor(int hexValue);
```

**Cortex Binding:**
```c
extern void InitWindow(int width, int height, string title);
extern int GetScreenWidth();
extern Color GetColor(int hexValue);
```

Note: `const char*` becomes `string` in Cortex.

### Structs

**C Header:**
```c
typedef struct Vector2 {
    float x;
    float y;
} Vector2;

typedef struct Color {
    unsigned char r;
    unsigned char g;
    unsigned char b;
    unsigned char a;
} Color;
```

**Cortex Binding:**
```c
struct Vector2 {
    float x;
    float y;
}

struct Color {
    unsigned char r;
    unsigned char g;
    unsigned char b;
    unsigned char a;
}
```

### Enums

**C Header:**
```c
enum {
    KEY_NULL = 0,
    KEY_A = 65,
    KEY_B = 66,
    KEY_C = 67
};

typedef enum {
    FLAG_NONE = 0,
    FLAG_DEBUG = 1,
    FLAG_RELEASE = 2
} ConfigFlags;
```

**Cortex Binding:**
```c
const int KEY_NULL = 0;
const int KEY_A = 65;
const int KEY_B = 66;
const int KEY_C = 67;

const int FLAG_NONE = 0;
const int FLAG_DEBUG = 1;
const int FLAG_RELEASE = 2;
```

### Defines

**C Header:**
```c
#define MAX_TOUCH_POINTS 10
#define PI 3.14159265358979323846
#define SCREEN_WIDTH 800
```

**Cortex Binding:**
```c
const int MAX_TOUCH_POINTS = 10;
const float PI = 3.14159265358979323846;
const int SCREEN_WIDTH = 800;
```

### Function Pointers

**C Header:**
```c
typedef void (*Callback)(int event, void* data);
```

**Cortex Binding:**
```c
// Function pointers need manual handling
// Comment generated, implement manually if needed
```

### Macros

**C Header:**
```c
#define CLITERAL(type) (type)
#define MAX(a, b) ((a) > (b) ? (a) : (b))
```

**Cortex Binding:**
```c
// Complex macros are commented out
// #define MAX(a, b) ((a) > (b) ? (a) : (b))
// Implement manually if needed
```

---

## Manual Binding

Sometimes you need to write bindings manually:

### When to Write Manually

- Complex macros that don't convert automatically
- Function pointers/callbacks
- Generic types
- Platform-specific code

### How to Write Manual Bindings

Create a file `bindings/mylib.cx`:

```c
// bindings/mylib.cx

// External function declarations
extern void mylib_init();
extern void mylib_cleanup();
extern int mylib_process(string input);

// Struct definitions
struct MyLibConfig {
    int timeout;
    bool debug;
    string log_path;
}

// Constants
const int MYLIB_VERSION = 100;
const int MYLIB_MAX_INPUT = 1024;

// Manual macro conversion
int MAX(int a, int b) {
    if (a > b) return a;
    return b;
}
```

Then use it:

```c
// main.cx
#include "bindings/mylib.cx"

void main() {
    mylib_init();
    int result = mylib_process("hello");
    mylib_cleanup();
}
```

---

## Common Libraries

### raylib (Game Development)

```bash
# Install
# Windows: Download from raylib.com
# Linux: sudo apt install libraylib-dev
# macOS: brew install raylib

# Generate bindings
cortex bind raylib -i raylib.h

# Build / run (ensure configs/raylib.json paths match your install)
cortex run game.cx
# Legacy single-shot: cortex -i game.cx -o game.exe -use raylib
```

```c
#include <raylib.h>

void main() {
    InitWindow(800, 450, "Game");
    SetTargetFPS(60);
    
    while (!WindowShouldClose()) {
        BeginDrawing();
        ClearBackground(BLACK);
        DrawText("Hello!", 400, 225, 20, WHITE);
        EndDrawing();
    }
    
    CloseWindow();
}
```

### SDL2 (Multimedia)

```bash
# Install
# Windows: Download from libsdl.org
# Linux: sudo apt install libsdl2-dev
# macOS: brew install sdl2

# Generate bindings
cortex bind sdl2 -i SDL2/SDL.h

# Build / run (configs/sdl2.json)
cortex run app.cx
# Legacy: cortex -i app.cx -o app.exe -use sdl2
```

```c
#include <SDL2/SDL.h>

void main() {
    SDL_Init(SDL_INIT_VIDEO);
    SDL_Window* window = SDL_CreateWindow(
        "Window",
        SDL_WINDOWPOS_CENTERED,
        SDL_WINDOWPOS_CENTERED,
        800, 450,
        SDL_WINDOW_SHOWN
    );
    
    // Event loop...
    
    SDL_DestroyWindow(window);
    SDL_Quit();
}
```

### libcurl (HTTP/Networking)

```bash
# Install
# Linux: sudo apt install libcurl4-openssl-dev
# macOS: brew install curl

# Generate bindings
cortex bind curl -i curl/curl.h

# Build
cortex build app.cx -o app -l curl
```

### SQLite (Database)

```bash
# Install
# Linux: sudo apt install libsqlite3-dev
# macOS: brew install sqlite

# Generate bindings
cortex bind sqlite3 -i sqlite3.h

# Build
cortex build app.cx -o app -l sqlite3
```

---

## Project Setup with Libraries

### cortex.toml Configuration

```toml
[project]
name = "mygame"
version = "0.1.0"
entry = "main.cx"

[dependencies.raylib]
path = "third_party/raylib"
include_path = "third_party/raylib/src"
lib_path = "third_party/raylib/build"
libs = ["raylib", "opengl32", "gdi32", "winmm"]
```

### Directory Structure

```
mygame/
├── cortex.toml
├── main.cx
├── bindings/
│   └── raylib.cx        # Generated
├── third_party/
│   └── raylib/          # Library source/binaries
│       ├── src/
│       │   └── raylib.h
│       └── build/
│           └── raylib.lib
└── src/
    ├── player.cx
    └── enemy.cx
```

---

## Troubleshooting

### "undefined reference to..."

**Problem:** Library not linked.

**Solution:**
```bash
# Add library linkage
cortex build main.cx -l mylib

# Or in cortex.toml
[dependencies.mylib]
libs = ["mylib"]
```

### "cannot find header file"

**Problem:** Include path not set.

**Solution:**
```bash
# Add include path
cortex build main.cx -I /path/to/include

# Or in cortex.toml
[dependencies.mylib]
include_path = "/path/to/include"
```

### "conflicting types for..."

**Problem:** Binding doesn't match actual function signature.

**Solution:** Check the C header and fix the binding manually:

```c
// If generated binding is wrong:
extern void myfunc(int x);

// Fix it manually:
extern void myfunc(int x, int y);  // Correct signature
```

### "macro not converted"

**Problem:** Complex macro wasn't auto-converted.

**Solution:** Write it manually as a function:

```c
// C macro:
// #define MAX(a, b) ((a) > (b) ? (a) : (b))

// Cortex function:
int MAX(int a, int b) {
    return a > b ? a : b;
}
```

### "struct has incomplete type"

**Problem:** Struct not defined before use.

**Solution:** Include the binding before using the struct:

```c
#include "bindings/raylib.cx"  // Must come first

void main() {
    Vector2 pos;  // Now defined
}
```

---

## Best Practices

1. **Generate bindings once** - Run `cortex bind` when you first add a library
2. **Commit bindings** - Add `bindings/*.cx` to version control
3. **Document manual fixes** - Comment any manual changes you make
4. **Keep headers updated** - Re-run binding generation when library updates
5. **Use project config** - Put library paths in `cortex.toml`

---

## Summary

| Task | Command |
|------|---------|
| Generate bindings | `cortex bind libname -i header.h` |
| Use bindings | `#include "bindings/libname.cx"` |
| Build with library | `cortex build main.cx -l libname` |
| Configure in project | Add to `cortex.toml` |

**Next Steps:**
- Try binding a library you want to use
- Check `examples/` for real-world examples
- See [CLI Reference](CLI_REFERENCE.md) for more build options
