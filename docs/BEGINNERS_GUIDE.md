# Cortex Beginner's Guide

Welcome to Cortex! This guide will teach you everything you need to know to get started, from installation to building your first programs.

## Table of Contents

1. [What is Cortex?](#what-is-cortex)
2. [Installation](#installation)
3. [Your First Program](#your-first-program)
4. [Running Programs](#running-programs)
5. [Building Executables](#building-executables)
6. [Project Structure](#project-structure)
7. [All CLI Commands](#all-cli-commands)
8. [Using C Libraries](#using-c-libraries)
9. [Binding C Headers to Cortex](#binding-c-headers-to-cortex)
10. [Language Features](#language-features)
11. [Common Patterns](#common-patterns)
12. [Troubleshooting](#troubleshooting)

---

## What is Cortex?

Cortex is a **systems programming language** that compiles to C. Think of it as "modern C" - it has:

- **C-like syntax** that feels familiar
- **Modern features** like type inference, defer, and channels
- **No manual memory management** - automatic cleanup
- **Easy C interop** - use any C library directly


### Why Cortex Instead of C?

| Problem in C | Solution in Cortex |
|--------------|-------------------|
| Manual `malloc`/`free` - easy to leak memory | Automatic cleanup with `defer` |
| Verbose type declarations | Type inference with `var` |
| No string interpolation | `"Hello ${name}!"` |
| Complex build systems | Simple `cortex run` and `cortex build` |
| Hard to use libraries | `#include` and `cortex bind` |


---

## Installation

I havent tested MAC or Linux yet. Please let me know if you encounter any issues.

### Windows

1. **Download** the latest `cortex-x.x.x-windows-amd64.zip` from [Releases](https://github.com/CharmingBlaze/Cortex/releases)

2. **Extract** to a folder like `C:\Cortex`

3. **Install** by running `install.bat` (adds Cortex and Zig to PATH)

4. **Verify** installation:
   ```cmd
   cortex --version
   cortex --help
   ```

### Linux

```bash
# Download and extract
wget https://github.com/CharmingBlaze/Cortex/releases/download/v0.1.0/cortex-0.1.0-linux-amd64.tar.gz
tar -xzf cortex-*.tar.gz

# Add to PATH (includes bundled Zig)
export PATH="$PWD/cortex-0.1.0-linux-amd64/bin:$PWD/cortex-0.1.0-linux-amd64/zig:$PATH"

# Add to your shell profile for persistence
echo 'export PATH="$HOME/cortex-0.1.0-linux-amd64/bin:$HOME/cortex-0.1.0-linux-amd64/zig:$PATH"' >> ~/.bashrc

# Verify
cortex --version
```

### macOS

```bash
# Download and extract (choose arm64 for M1/M2, amd64 for Intel)
curl -LO https://github.com/CharmingBlaze/Cortex/releases/download/v0.1.0/cortex-0.1.0-darwin-arm64.tar.gz
tar -xzf cortex-*.tar.gz

# Add to PATH (includes bundled Zig)
export PATH="$PWD/cortex-0.1.0-darwin-arm64/bin:$PWD/cortex-0.1.0-darwin-arm64/zig:$PATH"

# Add to your shell profile
echo 'export PATH="$HOME/cortex-0.1.0-darwin-arm64/bin:$HOME/cortex-0.1.0-darwin-arm64/zig:$PATH"' >> ~/.zshrc

# Verify
cortex --version
```

### Requirements

**None!** Cortex releases include Zig CC, a complete C compiler. Just download and run.

## Your First Program

### Option 1: Quick Start (Single File)

Create a file called `hello.cx`:

```c
// hello.cx - My first Cortex program
void main() {
    println("Hello, World!");
}
```

Run it:
```bash
cortex run hello.cx
```

Output:
```
Hello, World!
```

### Option 2: Create a Project

```bash
# Create a new project
cortex new myapp

# Navigate to the project
cd myapp

# Run it
cortex run
```

This creates:
```
myapp/
├── cortex.toml    # Project configuration
├── main.cx        # Entry point
└── src/           # Source files
```

---

## Running Programs

### Run a Single File

```bash
# Basic run
cortex run hello.cx

# With arguments
cortex run hello.cx -- arg1 arg2
```

### Run a Project

If you have a `cortex.toml` file:

```bash
# cortex.toml specifies the entry point
cortex run
```

### What Happens When You Run?

1. Cortex reads your `.cx` file
2. Compiles it to C code
3. Invokes the C compiler (Zig CC bundled, or system GCC/Clang)
4. Runs the resulting executable
5. Cleans up temporary files

---

## Building Executables

### Build a Single File

```bash
# Creates hello.exe (Windows) or hello (Linux/macOS)
cortex build hello.cx

# Specify output name
cortex build hello.cx -o myprogram
```

### Build a Project

```bash
# Uses cortex.toml configuration
cortex build

# Specify output
cortex build -o myapp
```

### Build with Libraries

Use the legacy flag-based mode for library linking:

```bash
# Link with raylib
cortex -i game.cx -o game.exe -use raylib

# Or in cortex.toml:
[dependencies.raylib]
path = "third_party/raylib"
libs = ["raylib", "opengl32", "gdi32"]
```

---

## Project Structure

### cortex.toml Configuration

```toml
[project]
name = "myapp"
version = "0.1.0"
entry = "main.cx"        # Entry point file
backend = "auto"         # C compiler: gcc, zig, or auto

[project.features]
async = true             # Enable async/await
actors = true            # Enable actor model
qol = true               # Quality-of-life features

[dependencies.raylib]
path = "third_party/raylib"
libs = ["raylib", "opengl32", "gdi32", "winmm"]
```

### Typical Project Layout

```
myapp/
├── cortex.toml          # Project config
├── main.cx              # Entry point
├── src/
│   ├── player.cx        # Modules
│   ├── enemy.cx
│   └── utils.cx
├── assets/
│   └── sprites.png
├── third_party/         # External libraries
│   └── raylib/
└── bindings/            # Generated bindings
    └── raylib.cx
```

---

## All CLI Commands

### `cortex new` - Create a Project

```bash
cortex new mygame
```

Creates a new project with:
- `cortex.toml` configuration
- `main.cx` entry point
- `src/` directory for modules

### `cortex run` - Compile and Run

```bash
# Run a file
cortex run main.cx

# Run a project (uses cortex.toml)
cortex run

# Pass arguments to program
cortex run main.cx -- --debug --level=1

# Keep generated C code
cortex run main.cx --keep-c
```

### `cortex build` - Compile to Executable

```bash
# Build with default name
cortex build main.cx

# Specify output name
cortex build main.cx -o mygame

# Release build (optimized)
cortex build main.cx -o mygame --release

# Build project
cortex build
```

### `cortex bind` - Generate C Library Bindings

```bash
# Generate bindings from a C header
cortex bind raylib -i third_party/raylib/src/raylib.h

# Creates bindings/raylib.cx
```

### Legacy Commands

```bash
# Old style (still works)
cortex -i main.cx -run
cortex -i main.cx -o output.exe -use raylib
```

---

## Using C Libraries

### Step 1: Include the Header

```c
// main.cx
#include <raylib.h>    // Include like in C
```

### Step 2: Use the Functions

```c
#include <raylib.h>

void main() {
    InitWindow(800, 450, "My Game");
    SetTargetFPS(60);
    
    while (!WindowShouldClose()) {
        BeginDrawing();
        ClearBackground(BLACK);
        DrawText("Hello!", 10, 10, 20, WHITE);
        EndDrawing();
    }
    
    CloseWindow();
}
```

### Step 3: Build with the Library

```bash
cortex build main.cx -o game --use raylib
```

### How It Works

Cortex passes `#include` directives directly to the C compiler. Any function declared in the header becomes available in Cortex automatically.

---

## Binding C Headers to Cortex

For a better experience, generate Cortex bindings from C headers.

### Why Generate Bindings?

- **Type safety** - Cortex knows the function signatures
- **Better errors** - Catch mistakes at compile time
- **Documentation** - Comments are preserved
- **IDE support** - Autocomplete works better

### How to Generate Bindings

```bash
# Basic binding generation
cortex bind raylib -i raylib.h

# This creates bindings/raylib.cx
```

### What Gets Generated

The binder extracts:

| C Construct | Cortex Output |
|-------------|---------------|
| Functions | `extern` declarations |
| Structs | `struct` definitions |
| Enums | Constants |
| `#define` constants | `const` declarations |

### Example Output

From C header:
```c
// raylib.h
typedef struct Vector2 {
    float x;
    float y;
} Vector2;

void DrawCircle(int x, int y, float radius, Color color);
#define LIGHTGRAY  CLITERAL(Color){ 200, 200, 200, 255 }
```

To Cortex binding:
```c
// bindings/raylib.cx
struct Vector2 {
    float x;
    float y;
}

extern void DrawCircle(int x, int y, float radius, Color color);
const Color LIGHTGRAY = { 200, 200, 200, 255 };
```

### Using Bindings

```c
// main.cx
#include "bindings/raylib.cx"

void main() {
    InitWindow(800, 450, "Game");
    // Now Cortex knows about Vector2, DrawCircle, etc.
}
```

---

## Language Features

### Variables and Types

```c
// Explicit types
int count = 10;
float price = 19.99;
string name = "Alice";

// Type inference with var
var count = 10;          // int
var price = 19.99;       // float
var name = "Alice";      // string

// Constants
const max_players = 4;
const float pi = 3.14159;
```

### Operators

```c
// Arithmetic
int a = 10 + 5;    // 15
int b = 10 - 3;    // 7
int c = 10 * 2;    // 20
int d = 10 / 2;    // 5

// Compound assignment
int score = 0;
score += 10;       // score = 15
score -= 5;        // score = 10
score *= 2;        // score = 20
score /= 4;        // score = 5

// Increment/decrement
int i = 0;
i++;               // i = 1
i--;               // i = 0
++i;               // i = 1 (prefix)
--i;               // i = 0 (prefix)
```

### Control Flow

```c
// If statements (parentheses optional)
if x > 10 {
    println("big");
} else if x > 5 {
    println("medium");
} else {
    println("small");
}

// Single-line if
if x > 0 println("positive");

// elif - cleaner than else if
if x == 5 {
    println("five");
} elif x == 10 {
    println("ten");
} elif x == 15 {
    println("fifteen");
} else {
    println("other");
}

// unless - inverse of if (runs when condition is false)
unless x < 0 {
    println("x is non-negative");
}

// unless with single statement
unless x > 100 println("x is not huge");

// if let - pattern matching for optionals
string maybe_name = get_name();
if let name = maybe_name {
    println("Hello, ${name}!");
} else {
    println("No name");
}

// While loop
while (x > 0) {
    x--;
}

// Loop forever (sugar for while(true))
loop {
    if (should_exit) break;
}

// For loop
for (int i = 0; i < 10; i++) {
    println(i);
}

// For-each loop
var numbers = [1, 2, 3, 4, 5];
for (var n in numbers) {
    println(n);
}
```

### Functions

```c
// Basic function
int add(int a, int b) {
    return a + b;
}

// Multiple return values
(int, int) divide(int a, int b) {
    return (a / b, a % b);
}

var (quotient, remainder) = divide(17, 5);

// Default parameters
void greet(string name = "World", int count = 1) {
    for (int i = 0; i < count; i++) {
        println("Hello, " + name + "!");
    }
}

greet();                // Hello, World!
greet("Alice");         // Hello, Alice!
greet(count: 3);        // Hello, World! (3 times)
```

### Structs

```c
struct Player {
    string name;
    int health;
    int score;
    
    // Method with implicit self
    void damage(int amount) {
        health -= amount;    // No self. needed
        if (health < 0) {
            health = 0;
        }
    }
    
    int get_health() {
        return health;       // Implicit self
    }
}

// Usage
Player p;
p.name = "Hero";
p.health = 100;
p.damage(20);
println(p.get_health());  // 80
```

### Arrays and Dicts

```c
// Arrays
var numbers = [1, 2, 3, 4, 5];
numbers[0] = 10;
println(numbers[0]);      // 10

// Dicts
var config = {
    "host": "localhost",
    "port": 8080
};
println(config["host"]);   // localhost
```

### String Interpolation

```c
string name = "Alice";
int age = 30;

// Interpolation
var message = "Hello, ${name}! You are ${age} years old.";
println(message);
// Output: Hello, Alice! You are 30 years old.
```

### Defer for Cleanup

```c
void process_file() {
    var file = fopen("data.txt", "r");
    defer fclose(file);    // Runs when function exits
    
    // Use file...
    // fclose is called automatically
}
```

---

## Common Patterns

### Game Loop Pattern

```c
#include <raylib.h>

void main() {
    InitWindow(800, 450, "Game");
    SetTargetFPS(60);
    
    const paddle_speed = 5;
    int player_y = 200;
    
    loop {
        // Input
        if IsKeyDown(KEY_W) player_y -= paddle_speed;
        if IsKeyDown(KEY_S) player_y += paddle_speed;
        
        // Update
        if player_y < 0 player_y = 0;
        if player_y > 400 player_y = 400;
        
        // Draw
        BeginDrawing();
        ClearBackground(BLACK);
        DrawRectangle(20, player_y, 10, 50, WHITE);
        EndDrawing();
        
        // Exit
        if WindowShouldClose() break;
    }
    
    CloseWindow();
}
```

### Configuration Pattern

```c
// config.cx
const WINDOW_WIDTH = 800;
const WINDOW_HEIGHT = 450;
const GAME_TITLE = "My Game";
const FPS = 60;

// main.cx
#include "config.cx"

void main() {
    InitWindow(WINDOW_WIDTH, WINDOW_HEIGHT, GAME_TITLE);
    SetTargetFPS(FPS);
    // ...
}
```

### Module Pattern

```c
// src/player.cx
struct Player {
    int x, y;
    int health;
    
    void init(int start_x, int start_y) {
        x = start_x;
        y = start_y;
        health = 100;
    }
    
    void move(int dx, int dy) {
        x += dx;
        y += dy;
    }
}

// main.cx
#include "src/player.cx"

void main() {
    Player p;
    p.init(100, 100);
    p.move(10, 0);
}
```

---

## Troubleshooting

### "cortex: command not found"

**Problem:** Cortex isn't in your PATH.

**Solution:**
- Windows: Run `install.bat` and restart terminal
- Linux/macOS: Add to PATH in your shell profile

### "gcc: command not found"

**Problem:** No C compiler installed.

**Solution:**
- Windows: Install MinGW or use bundled Zig CC
- Linux: `sudo apt install gcc`
- macOS: `xcode-select --install`

### "undefined reference to..."

**Problem:** Missing library linkage.

**Solution:**
```bash
cortex build main.cx -o myapp --use mylib
```

Or add to `cortex.toml`:
```toml
[dependencies.mylib]
libs = ["mylib"]
```

### "cannot find header file"

**Problem:** Include path not set.

**Solution:**
```bash
# Add include path
cortex build main.cx -I /path/to/headers
```

Or in `cortex.toml`:
```toml
[dependencies.mylib]
include_path = "/path/to/headers"
```

### "syntax error"

**Problem:** Cortex syntax issue.

**Solution:** Check for:
- Missing semicolons after statements
- Mismatched braces
- Wrong operator syntax

### View Generated C Code

When debugging, see what Cortex generates:

```bash
cortex run main.cx --keep-c
# Creates main.c you can inspect
```

---

## Getting Help

- **Documentation:** `docs/LANGUAGE_SPEC.md`, `docs/LANGUAGE_GUIDE.md`
- **Examples:** `examples/` directory
- **GitHub:** https://github.com/CharmingBlaze/Cortex
- **Issues:** https://github.com/CharmingBlaze/Cortex/issues

---

## Next Steps

1. Try the examples in `examples/`
2. Build a simple game with raylib
3. Create your own project with `cortex new`
4. Bind a C library you want to use


