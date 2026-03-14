# Cortex Compiler

A full-featured C-like programming language without pointers, designed for game and application development with smart dynamic variables, C library interop, feature toggles, and quality of life features.

**New: Self-contained build system** — No external compilers required. Just `cortex build` and go.

## Why Cortex?

- **Smart Intelligence**: Like the brain's cortex, handles complex operations automatically
- **Modern & Memorable**: Easy to remember and sounds professional
- **Zero-Config Build System**: Just `cortex build` — auto-detects everything
- **Cross-Platform**: Windows, Linux, macOS with bundled TCC compiler
- **Optional Manual Control**: Full auto or full manual — your choice

## Quick Start

```bash
# Build and run a Cortex program
cortex -i hello.cx -run

# Build C project with zero configuration
cortex build

# Build and run C project
cortex build -run
```

## Features

### Core Language
- **C-like syntax**: Familiar syntax for C programmers
- **No pointers in Cortex code**: Automatic memory management; use `extern` to call C APIs
- **Smart dynamic variables**: Type inference with `var` keyword
- **Universal type**: `any` type for dynamic typing
- **Full static and dynamic typing**: Use both in the same program

### Modern Features
- **Dict literals**: `{ "key": value }` with compile-time key checking
- **Array literals**: `[1, 2, 3]` with bounds checking
- **Lambdas**: No-capture and by-value capture support
- **String interpolation**: `"Hello ${name}"`
- **Pattern matching**: `match (value) { case int n: ... }`
- **Named & Default Parameters**: `fn greet(name = "World") { ... }`
- **Multiple return values**: `(int, int) f()` and `return (a, b)`
- **defer**: `defer { cleanup(); }`

### GUI System (New!)

**Native cross-platform GUI** powered by Fyne:
- **Simple API**: Create windows, buttons, labels, inputs with ease
- **Event-driven**: Lambda callbacks for clicks, changes, key events
- **Layouts**: VBox, HBox, Grid, Border layouts
- **Widgets**: Buttons, labels, entries, checkboxes, sliders, progress bars
- **Graphics**: Images, rectangles, circles, lines, custom drawing
- **Dialogs**: Info, error, confirm, file open/save dialogs
- **Cross-platform**: Native look on Windows, macOS, Linux

```c
#include <gui_runtime.h>

void main() {
    window w = gui_window("Hello", 800, 600);
    gui_label(w, "Hello World");
    gui_button(w, "Click Me", [](event e) {
        println("Clicked!");
    });
    gui_run();
}
```

See **docs/GUI_SYSTEM.md** for full documentation and **examples/gui/** for examples.
- **Build System**: `cortex build` — auto-detects C files, includes, libraries
- **Library Detection**: Scans `#include` directives, auto-links libraries
- **MSYS2 Support**: Windows users can leverage pacman package manager
- **TCC Bundled**: Tiny C Compiler auto-downloaded if no compiler found
- **Cross-Platform**: Works on Windows, Linux, macOS
- **Optional & Manual Modes**: Full auto or full manual control

### Build System (New!)

**Fully Automatic:**
```bash
# Zero-config build — finds *.c, detects libraries, compiles
cortex build

# Build and run
cortex build -run

# Release build
cortex build -release

# Verbose output
cortex build -v
```

**Manual Control:**
```bash
# Disable auto-detection, specify everything manually
cortex build --manual --sources=main.c,lib.c --includes=./include -o myapp

# Force specific compiler
cortex build --compiler=gcc

# Use bundled TCC
cortex build --tcc

# Windows: Use MSYS2
cortex build --msys2

# Disable auto-download of libraries
cortex build --no-autofetch
```

**Full Control:**
```bash
# Escape hatch — use system compiler directly
cortex build --use-build=false
# Then: gcc main.c -o myapp -lraylib
```

- **C-like syntax**: Familiar syntax for C programmers
- **No pointers in Cortex code**: Automatic memory management; use `extern` to call C APIs that use pointers (`void*`, `char*`, etc.)
- **Smart dynamic variables**: Type inference with `var` keyword
- **Universal type**: `any` type for dynamic typing
- **Game-friendly features**: Built-in vector types, random functions, time functions
- **Modern types**: Built-in support for strings, booleans, and vectors
- **Quality of life features**: String interpolation, pattern matching, list comprehensions, **repeat** loop, **assert**
- **BASIC-like I/O**: `print` for output (no newline); `writeline` or `println` for a line; `printf(format, ...)` and `writeline(format, ...)` like C with format specifiers. `say`/`show` are aliases.
- **Game math**: `sign_float`, `wrap_float`, `round_float`, `floor_float`, `ceil_float` for one-liner logic
- **Simple compilation**: Compiles to C code for maximum portability
- **Full-featured compiler**: Pointer types in `extern` declarations, string/char escape sequences (`\n`, `\t`, `\"`, `\\`), feature gating (async/actors), optional parameter names in extern, and automated tests

## Language Overview

**New to Cortex?** See **[LANGUAGE_GUIDE.md](LANGUAGE_GUIDE.md)** for a full beginner-friendly guide with examples (variables, loops, functions, structs, enums, arrays, error handling, C libraries, and more).

### Full static and dynamic typing

Cortex supports **both** styles in the same program:

- **Static typing** — Declare variables with explicit types (`int`, `float`, `string`, `struct`, `enum`, etc.). The compiler checks that assignments and initializers match the type. Numeric promotion is allowed (`int` → `float` → `double`). Mismatches are reported at compile time.
- **Dynamic typing** — Use `var` (inferred type, can be reassigned to another type) or `any` (explicit “hold anything”) when you need flexibility. No static check on assignments; use `is_type()`, `as_int()`, `as_string()`, etc. at runtime when reading values.

Use static types for clarity and early errors; use `var`/`any` when you need scripting-like behavior or heterogeneous data.

### Types

- **Primitives:** `int`, `float`, `double`, `char`, `bool`, `string`
- **Vectors:** `vec2`, `vec3` (game math)
- **Structs:** `struct Name { ... }` — use **dot notation** `obj.field`; optional **methods** (functions inside the struct) with `obj.method(args)`.
- **Enums:** `enum State { Idle, Running, Done }` — use **dot notation** `State.Idle` or just `Idle` in switch
- **Arrays:** array literals `[1, 2, 3]`; use `for (x in arr)` to iterate
- **var** — Smart dynamic variable (type inference, can change type on reassignment)
- **any** — Universal type for dynamic values; use `is_type()`, `as_int()`, etc.

### Smart dynamic variables

```c
// Type inference
var x = 42;           // Automatically typed as int
var name = "Hello";   // Automatically typed as string
var pos = make_vec2(1.0, 2.0); // Automatically typed as vec2

// Dynamic typing
var dynamic = 10;     // Start as int
dynamic = "string";   // Now string
dynamic = 3.14;       // Now float

// Universal type
any value = 42;
value = "hello";
value = make_vec2(1.0, 2.0);

// Type checking and casting (for any/var at runtime)
if (is_type(value, "int")) {
    int num = as_int(value);
    println("Number: " + num);
}

// Static typing: compiler enforces types
int n = 42;
float f = 3.14;
// n = "hello";   // compile error: type mismatch
// f = "x";       // compile error: type mismatch
```

### Control Structures

**Conditionals:** `if` / `else if` / `else`

**Loops** — use the right one for the job:
- **for** — classic C-style: `for (int i = 0; i < 10; i++) { }`
- **while** — `while (condition) { }`
- **do-while** — `do { } while (condition);`
- **repeat** — run block n times: `repeat (10) { show("Hi"); }`
- **for-in** — iterate over array: `for (x in arr) { }`

**Loop control:** `break;` and `continue;` inside any loop.

**Switch (classic)** — match on constant values (ints, enums, literals):
```c
switch (state) {
    case 1: { show("one"); break; }
    case 2: { show("two"); break; }
    default: { show("other"); }
}
```
Use `match (x) { case type var: ... }` for pattern matching on types.

```c
// If statement
if (condition) {
    // code
} else {
    // code
}

// For loop
for (int i = 0; i < 10; i++) {
    // code
}

// While loop
while (condition) {
    // code
}

// Do-while
do {
    // code
} while (condition);

// Repeat loop — run block n times (great for games and demos)
repeat (10) {
    show("Hello!");
}

// For-in over array
var nums = [10, 20, 30];
for (n in nums) {
    println("" + n);
}
```

### Functions

```c
int add(int a, int b) {
    return a + b;
}

void main() {
    int result = add(5, 3);
    println("Result: " + result);
}
```

### Enums and dot notation

```c
enum State { Idle, Running, Done }

void main() {
    int state = Idle;           // or State.Idle when qualified
    switch (state) {
        case Idle:   { show("idle"); break; }
        case Running: { show("run"); break; }
        default:     { show("done"); }
    }
}
```
Use **dot notation** for struct fields (`player.health`) and enum values (`State.Idle`).

### Struct methods

You can define methods inside a struct. Method bodies can use the struct’s field names directly; they are emitted as `self->field` in C. Call with `obj.method(args)`.

```c
struct Player {
    int x;
    int y;
    void move(int dx) {
        x = x + dx;
        y = y + dx;
    }
}

void main() {
    Player p;
    p.x = 10;
    p.y = 20;
    p.move(5);
    show("x: " + p.x);   // 15
    show("y: " + p.y);   // 25
}
```

### Collections (dynamic array and dictionary)

Cortex provides a **dynamic array** and **dictionary** type with a clean API so you don’t reinvent them in C.

**Dynamic array** — type `array`, grows as needed; elements are stored as `any` (int, float, string, etc.):

```c
array a = array_create();
array_push(a, 42);
array_push(a, "hello");
array_push(a, 3.14);
int n = array_len(a);           // 3
any first = array_get(a, 0);    // use as_int(first), as_string(...) as needed
array_set(a, 1, make_any_int(99));
array_free(a);                  // call when done
```

**Dictionary** — type `dict`, string keys and `any` values:

```c
dict d = dict_create();
dict_set(d, "name", make_any_string("Cortex"));
dict_set(d, "score", make_any_int(100));
bool has = dict_has(d, "name"); // true
any val = dict_get(d, "score"); // use as_int(val)
int size = dict_len(d);
dict_free(d);
```

Values passed to `array_push`, `array_set`, and `dict_set` are automatically boxed from `int`/`float`/`string`/`bool`/`vec2`/`vec3` when you pass literals or typed expressions. Use `as_int`, `as_float`, `as_string`, etc. when reading from `array_get` / `dict_get`. Both types require the QOL feature (default on).

### Error handling with Result

A **Result&lt;T, E&gt;-style** pattern is available so you can return either a value or an error message:

```c
result r = parse_number("42");
if (result_is_ok(r)) {
    any v = result_value(r);
    show("value: " + as_int(v));
} else {
    show("error: " + result_error(r));
}

// Create results
result ok = result_ok(make_any_int(42));
result err = result_err("something went wrong");
```

- `result_ok(value)` — wrap a value (as `any`) in a successful result  
- `result_err("message")` — create an error result  
- `result_is_ok(r)` — true if success  
- `result_value(r)` — the value (only meaningful when `result_is_ok(r)` is true)
- `result_error(r)` — the error message (only meaningful when failed)

**Match on Result** — use `match (r)` with `case Ok(v):` and `case Err(e):` to unwrap:

```c
result r = parse_something();
match (r) {
    case Ok(v): { show("value: " + as_int(v)); }
    case Err(e): { show("error: " + e); }
}
```

`v` is bound as `AnyValue` (use `as_int`, `as_string`, etc.); `e` is bound as `char*` (error message).

### Lambdas (no-capture and by-value capture)

You can pass **lambdas** as callbacks for events, UI, or helpers. **No-capture** lambdas are emitted as static C functions; **by-value capture** lambdas are emitted as a closure (function pointer + env) and are supported in **call-argument position** (pass to C/Cortex functions that take `(fn, env, ...)`).

```c
// No capture: [] (params) [-> returnType] { body }
on_click([](int x, int y) { show("clicked"); });

// By-value capture: [x, y] (params) { body } — pass as (fn, &env); C callee gets (void* fn, void* env, ...)
extern void apply_twice(void* fn, void* env, int x);
int base = 10;
apply_twice([base](int x) { return base + x; }, 1);
```

Use lambdas as **arguments** to C or Cortex functions. No-capture lambdas can be stored in `var` (function pointer). Captured lambdas are only valid as call arguments; the compiler emits two C arguments (function pointer and env pointer).

### Dynamic list API

Beyond `array_create` / `array_push` / `array_get` / `array_set` / `array_len` / `array_free`, you get a full list API:

- `array_pop(a)` — remove and return last element (as `any`)
- `array_insert(a, index, value)` — insert at index
- `array_remove_at(a, index)` — remove element at index
- `array_capacity(a)` — current capacity
- `array_reserve(a, min_cap)` — reserve capacity for fast growth

### Events and callbacks

**Event type** for subscribe/emit patterns (UI, input, game state):

```c
event e = event_create();
event_subscribe(e, my_callback);   // my_callback(AnyValue) in C
event_emit(e, make_any_int(42));
event_unsubscribe(e, my_callback);
event_free(e);
```

Works with lambdas: pass a no-capture lambda as the callback.

### Standard library (string, math, file, debug)

- **String:** `str_split(s, delim)` → array of strings · `str_join(arr, sep)` · `str_replace(s, from, to)` · `str_trim(s)` · `starts_with(s, prefix)` · `ends_with(s, suffix)` · `to_lower(s)` · `to_upper(s)`
- **Math:** `clamp_int(x, lo, hi)` · `pow(base, exp)` · `random_choice(array)` → random element (any)
- **File/path:** `file_exists(path)` · `list_dir(path)` · `path_join(a, b)`
- **Debug:** `debug_log("format", ...)` · `debug_assert(condition, "msg")` · `dump(value)` (introspect `any` to stderr)
- **JSON:** `json_parse(string)` → dict (object with string keys; values = number/string/bool/null) · `json_stringify(value)` / dict → string (minimal implementation in runtime)

### Unit tests

```c
test "addition" {
    assert_eq(1 + 1, 2);
}
test "approx" {
    assert_approx(0.1 + 0.2, 0.3, 0.001);
}
void main() {
    test_run_all();   // runs all test "name" { } blocks
    show("done");
}
```

- `test "name" { block }` — registers a test; compiled as a static function and registered at startup.
- `assert_eq(a, b)` — aborts if `a != b` (int, float, or string).
- `assert_approx(a, b, epsilon)` — for floats.

### Perfect for games and applications

Cortex ships with a rich runtime so you can build games and apps without fighting pointers or C APIs.

**Vectors (2D/3D)** — movement, positions, physics:
- `make_vec2`, `make_vec3` · `vec2_add`, `vec2_sub`, `vec2_scale` (same for vec3)
- `vec2_length`, `vec2_length_sq`, `vec2_distance` · `normalize`, `dot` · `vec3_length`, `vec3_normalize`, `vec3_dot`, `vec3_distance`

**Game math** — one-liners you use everywhere:
- `clamp_float(x, lo, hi)` · `lerp_float(a, b, t)` · `min_float`, `max_float`
- `sign_float(x)` · `wrap_float(x, lo, hi)` (wrap value in range) · `round_float`, `floor_float`, `ceil_float`

**Assertions** — catch bugs early:
- `assert(condition)` · `assert(condition, "message")`

**Print and output (BASIC-like)** — `print(x)` or `print(a, b, c)` (no newline); `writeline(x)` or `writeline("%d bottles", n)` (line with optional format); `printf(format, ...)` exactly like C; `say`/`show` are aliases for `print`/`writeline`.

**Random & time** — gameplay and delta:
- `random_int(min, max)` · `random_float(min, max)` · `get_time()` (seconds since start) · `sleep(seconds)` · `wait(seconds)` (alias for sleep)

**Application I/O** — no manual buffers:
- `print(x)` or `print(a, b, c)` (no newline) · `writeline(x)` or `writeline("%d", n)` (line; C-style format) · `printf(format, ...)` (C printf) · `say`/`show` aliases · `input_line()` (read a line; returns string)
- `read_file(path)` · `write_file(path, content)`

See `examples/guess_game.cx` (simple game) and `examples/app_file.cx` (file I/O) for full examples.

### Built-in game and app functions (reference)

```c
// Vectors
vec2 make_vec2(float x, float y);
vec3 make_vec3(float x, float y, float z);
vec2 vec2_add(vec2 a, vec2 b);
vec2 vec2_sub(vec2 a, vec2 b);
vec2 vec2_scale(vec2 v, float s);
float vec2_length(vec2 v);
float vec2_distance(vec2 a, vec2 b);
vec2 normalize(vec2 v);
float dot(vec2 a, vec2 b);
// ... and vec3_* equivalents

// Math
float clamp_float(float x, float lo, float hi);
float lerp_float(float a, float b, float t);
float min_float(float a, float b);
float max_float(float a, float b);
float sign_float(float x);
float wrap_float(float x, float lo, float hi);
float round_float(float x);
float floor_float(float x);
float ceil_float(float x);

// Assert (aborts on failure): assert(condition) or assert(condition, "message")

// Collections (QOL): dynamic array and dictionary
// array a = array_create(); array_push(a, value); array_get(a, index); array_set(a, index, value); array_len(a); array_free(a);
// dict d = dict_create(); dict_set(d, "key", value); dict_get(d, "key"); dict_has(d, "key"); dict_len(d); dict_free(d);

// Result (error handling): result_ok(value), result_err("msg"), result_is_ok(r), result_value(r), result_error(r)

// Random & time
int random_int(int min, int max);
float random_float(float min, float max);
float get_time();
void sleep(float seconds);
void wait(float seconds);   // alias for sleep

// I/O (print = no newline; writeline/println = with newline; printf = C-style)
void print(...);           // one or more values, no newline
void writeline(...);       // one value, or printf-style: writeline("%d", x)
void println(string s);    // alias: writeline(s)
void say(...);             // same as print
void show(...);            // same as writeline
int printf(const char* format, ...);  // C printf; format specifiers %d %s %f etc.
string input_line();
string read_file(string path);
int write_file(string path, string content);
```

## Building the Compiler

### Prerequisites

- **Go 1.19 or later** to build the Cortex compiler.
- **A C compiler to produce executables** — either:
  - **TinyCC (tcc)** — Single executable, no install; put `tcc.exe` in `tools/` next to `cortex` or in PATH. Then you don't need gcc or CMake. Run `.\scripts\setup_tcc.ps1` for instructions.
  - **GCC (MinGW, etc.)** — Full toolchain; use `-backend gcc` if you have both and prefer gcc.

### Build Steps

1. Clone or download the Cortex compiler source.
2. Navigate to the project directory.
3. Build the compiler:

```bash
go build -o cortex
```

### Standalone: no gcc or CMake

You can compile and run Cortex **without installing gcc or CMake**:

1. **Get TinyCC** — Download a Windows build of [Tiny C Compiler (tcc)](https://repo.or.cz/tinycc.git) (or from [community builds](https://github.com/nickhutchinson/tinycc/releases)), and save it as `tools/tcc.exe` next to your `cortex` executable. Or run `.\scripts\setup_tcc.ps1` for guidance.
2. **Use default backend** — Cortex uses `backend=auto`: it tries tcc first, then gcc. With tcc in `tools/` or PATH, compilation uses tcc only.
3. **Run without leaving an exe** — Use `-run` to compile to a temp file, run, then delete:
   ```bash
   cortex -i main.cx -run
   ```

No raylib or other C libs are needed for plain Cortex programs; the runtime (core.c, game.c) is compiled from source by tcc or gcc. For raylib you still need raylib headers and a prebuilt lib (or build raylib once with gcc/cmake and point config at it).

| Flag / config | Meaning |
|---------------|--------|
| `-backend auto` | (default) Use tcc if found, else gcc |
| `-backend tcc` | Use only tcc (fail if not found) |
| `-backend gcc` | Use only gcc |
| `-run` | Compile, run the program, then remove the temp exe |

## Using the Compiler

### Basic Usage

```bash
# Compile a Cortex file
./cortex -i input.cx -o output.exe

# Or let it choose the output name
./cortex -i input.cx

# Compile and run (no exe left behind)
./cortex -i input.cx -run

# With a feature config (optional)
./cortex -i game.cx -o game.exe -config configs/games_only.json
```

### Feature Configuration

Cortex supports optional feature toggles via a JSON config file. Use `-config <path>` to enable or disable:

| Feature     | Description                          |
|------------|--------------------------------------|
| `async`    | Async/await (future)                  |
| `actors`   | Actors and channels (future)         |
| `blockchain` | Blockchain helpers (sha256, keccak, etc.) |
| `qol`      | Quality-of-life (vec2/vec3, random, time, type_of, etc.) |

Example configs in `configs/`:

- `configs/full_features.json` – all features on (default when no config)
- `configs/games_only.json` – only `qol` (vectors, random, time; no blockchain)
- `configs/minimal.json` – all features off (minimal runtime)
- `configs/blockchain_only.json` – blockchain + qol

If you use a gated built-in (e.g. `sha256_hash`) with that feature disabled, the compiler reports a clear error.

The compiler looks for the `runtime/` directory (containing `core.c` and `core.h`) in: the current working directory, next to the executable, or the path in the `CORTEX_ROOT` environment variable. Run from the project root or set `CORTEX_ROOT` when building from elsewhere.

### Manual C Library Setup (No Auto-Fetch)

**By default, auto-fetch is disabled.** You have full control over library installation. Here are three ways to use C libraries:

#### Option 1: Manual CLI Flags (Simplest)

If you have the library installed on your system:

```bash
# Compile with manual include/library paths
cortex -i game.cx -o game \
  -I C:/raylib/include \
  -L C:/raylib/lib \
  -l raylib
```

#### Option 2: Create a Config File (Recommended)

Create `myproject.json` in your project directory:

```json
{
  "features": { "qol": true },
  "include_paths": ["C:/raylib/include"],
  "library_paths": ["C:/raylib/lib"],
  "libraries": ["raylib"]
}
```

Then build:
```bash
cortex -i game.cx -o game -config myproject.json
```

#### Option 3: Use System Package Manager

**Windows (MSYS2):**
```bash
# Install library manually
pacman -S mingw-w64-x86_64-raylib

# Build with system library
cortex build
```

**Linux (apt/yum/pacman):**
```bash
# Install library
sudo apt install libraylib-dev

# Build - Cortex will find it automatically
cortex build
```

**macOS (Homebrew):**
```bash
# Install library
brew install raylib

# Build
cortex build
```

#### Manual Build System Mode

For full manual control with the build system:

```bash
# Disable auto-detection, specify everything manually
cortex build --manual \
  --sources=main.c \
  --includes=C:/raylib/include \
  --libs=raylib \
  -o mygame

# Or disable build system entirely and use your own compiler
cortex build --use-build=false
# Then compile manually:
gcc main.c -o mygame -I C:/raylib/include -L C:/raylib/lib -lraylib
```

#### Library Installation Guide

**Windows:**
1. Download library headers and .lib/.dll files
2. Place in a folder like `C:/libraries/libname/`
3. Create `libname.json` config with paths
4. Use `cortex -i main.cx -config libname.json`

**Linux/macOS:**
1. Install via package manager (libraries go to /usr/local/ automatically)
2. Cortex auto-detects standard paths
3. Or specify manually with `-I/-L/-l` flags

#### Enabling Auto-Fetch (Optional)

If you want Cortex to automatically download missing libraries:

```bash
# Enable auto-fetch (disabled by default)
cortex build --autofetch

# Or in config file
{
  "auto_fetch": true,
  "libraries": ["raylib"]
}
```

**Note:** Auto-fetch only downloads from trusted sources and respects your system's package manager when available.

### Using C libraries (raylib, SDL, etc.)

Use **C-style `#include`** only. Cortex passes includes to the C compiler and **infers linking** from `#include <name.h>` (e.g. `#include <raylib.h>` links `raylib`). No `#pragma link` needed.

1. **Include the header** at the top of your `.cx` file:
   ```c
   #include <raylib.h>
   ```
2. **Build** with paths from config or CLI:
   ```bash
   cortex -i game.cx -o game -use raylib
   ```
   `-use raylib` loads `configs/raylib.json` if you don’t pass `-config`. For other libs, create `configs/<name>.json` and use `-use <name>` or `-config configs/<name>.json`.

Any library name works: `#use "sdl2"` generates `#include <sdl2.h>` and `-l sdl2`. Add a matching `configs/sdl2.json` with `include_paths` and `library_paths` for your install.

3. **Call C functions** — use `extern` for declarations if the header only defines macros, or just include the header and use the API.

**Example: raylib**

```c
#include <raylib.h>

void main() {
    InitWindow(800, 450, "Cortex + raylib");
    while (!WindowShouldClose()) {
        BeginDrawing();
        ClearBackground(RAYWHITE);
        DrawText("Hello from Cortex!", 190, 200, 20, LIGHTGRAY);
        EndDrawing();
    }
    CloseWindow();
}
```

Build (paths from config or CLI):

```bash
# -use raylib loads configs/raylib.json
cortex -i game.cx -o game -use raylib

# Or explicit config
cortex -i game.cx -o game -config configs/raylib.json

# Or CLI flags
cortex -i game.cx -o game -I /usr/local/include -L /usr/local/lib -l raylib
```

**Shorthand:** `#use "raylib"` is equivalent to `#include <raylib.h>` plus `#pragma link("raylib")`; use either C-style include or `#use`.

### Networking (TCP, UDP, HTTP, RPC, multiplayer)

Cortex includes a **modular networking runtime** (`runtime/network.c` / `runtime/network.h`) so you can build multiplayer games and networked apps with a simple, high-level API. The compiler only links the network module when your code uses it.

**TCP (streams)** — servers and clients:
- `tcp_listen(port)` — start a server; returns server socket (or -1).
- `tcp_accept(server)` — accept one client; returns client socket.
- `tcp_connect(host, port)` — connect to a host; returns socket.
- `tcp_send(sock, data, len)` / `tcp_recv(sock, buf, len)` — send/receive bytes.
- `tcp_recv_string(sock, max_len)` — receive into a new string (caller frees).
- `tcp_close(sock)` — close a socket.

**UDP (datagrams)** — low-latency, peer-to-peer:
- `udp_socket()` — create a UDP socket.
- `udp_send_to(sock, host, port, data, len)` / `udp_recv_from(...)` — send/receive packets.
- `udp_close(sock)`.

**HTTP client** — REST APIs and web services:
- `http_get(url)` — GET request; returns response body (caller frees) or NULL.
- `http_post(url, body)` — POST with body.
- `http_get_with_header(url, user_agent)` — GET with custom User-Agent.

**HTTP server** — minimal server for backends:
- `http_server_listen(port)` — listen (returns server socket).
- `http_server_read_request(client)` — read request line + headers.
- `http_server_send_response(client, status, body)` — send response and close.

**RPC** — JSON over HTTP:
- `rpc_call(url, json_request)` — POST JSON, returns JSON response (caller frees).

**Real-time multiplayer** — length-prefixed messages:
- `net_send_message(sock, data, len)` — send 4-byte length + payload.
- `net_recv_message(sock)` — receive one message; returns string (caller frees).

Use these from Cortex like any other function; socket handles are `int` (-1 = invalid). On Windows the runtime links `ws2_32` automatically when you use networking. See **docs/NETWORKING.md** for examples (client/server, HTTP, simple multiplayer).

**Raylib examples (bundled)** — The repo can use a local copy of [raylib](https://github.com/raysan5/raylib) to build and run Cortex ports of official examples:

1. **Clone and build raylib** (once):
   ```powershell
   git clone --depth 1 https://github.com/raysan5/raylib.git third_party\raylib
   cd third_party\raylib
   cmake -B build -G "MinGW Makefiles" -DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=ON -DCMAKE_C_COMPILER=gcc
   cmake --build build --config Release
   cd ../..
   ```
   Or run `.\scripts\build_raylib_examples.ps1` from the project root to clone, build raylib, and compile all Cortex raylib examples.

2. **Compile an example** (from project root):
   ```bash
   cortex -i examples/raylib/core_basic_window.cx -o examples/raylib/core_basic_window.exe -config configs/raylib.json
   ```

3. **Run** — On Windows you may need `libraylib.dll` next to the exe: copy `third_party\raylib\build\raylib\libraylib.dll` into `examples\raylib\`, or add that directory to `PATH`.

For APIs that need `Vector2` (e.g. `DrawTriangle`, `DrawCircleV`), use `extern Vector2 Vec2(float x, float y);` — the compiler links `runtime/raylib_helper.c` when you link raylib, which provides `Vec2`.

**Config file** — you can put include and library settings in your config JSON so you don’t repeat them:

```json
{
  "features": { "async": false, "actors": false, "blockchain": false, "qol": true },
  "include_paths": ["C:/raylib/include", "/usr/local/include"],
  "library_paths": ["C:/raylib/lib", "/usr/local/lib"],
  "libraries": ["raylib"]
}
```

**Flags**

| Flag | Meaning |
|------|--------|
| `-backend` | C backend: `gcc`, `tcc`, or `auto` (no gcc/cmake if tcc in tools/ or PATH) |
| `-run` | Compile and run (temp exe then delete) |
| `-use <name>` | Use C library by name (loads `configs/<name>.json` if no `-config`) |
| `-I <path>` | Add include path (can repeat) |
| `-L <path>` | Add library search path (can repeat) |
| `-l <lib>` | Link library, e.g. `-l raylib` (can repeat) |

These are passed through to `gcc`. Cortex turns `#pragma link("name")` and `#use "name"` in your source into `-l name` automatically.

### Examples

The `examples/` directory contains several example programs:

- `hello.cx` - Basic hello world program
- `guess_game.cx` - Number-guessing game (input, random, loops)
- `app_file.cx` - File I/O (read_file, write_file)
- `vec2_demo.cx` - 2D vector math and movement
- `raylib_hello.cx` - Minimal raylib window (requires raylib; use -I -L -l or config)
- **`examples/raylib/`** — Cortex ports of [raylib examples](https://github.com/raysan5/raylib/tree/master/examples): `core_basic_window.cx`, `core_input_keys.cx`, `shapes_basic_shapes.cx`. Build raylib first (see below), then use `-config configs/raylib.json`.
- `clibraries.cx` - Using C stdlib/math (printf, malloc, etc.)
- `game.cx` - Simple text-based adventure game
- `calculator.cx` - Mathematical operations and functions
- `dynamic.cx` - Smart dynamic variables and quality of life features demo

#### Compile and Run Examples

```bash
# Compile hello world
./cortex -i examples/hello.cx -o hello

# Run the compiled program
./hello

# Compile the game example
./cortex -i examples/game.cx -o game

# Run the game
./game
```

## Is Cortex “as easy as BASIC but looks like C”? Can it use full C libraries?

### Debugging Cortex programs

Cortex compiles to C, so you can debug with any native debugger (GDB, LLDB, Visual Studio, etc.) using the generated C source.

1. **Emit C and build with debug info** — Generate the C file and compile with debug symbols: `cortex -i main.cx -o out.c` then `gcc -g -O0 -o main out.c runtime/core.c runtime/game.c -I runtime`. Use your compiler's `-g` (e.g. MSVC `/Zi`) and `-O0` to avoid confusing optimizations.
2. **Run under the debugger** — Start the program under GDB/LLDB/VS (e.g. `gdb ./main`, `break main`, `run`). Set breakpoints in the generated `.c` file by line number.
3. **Inspect variables** — Cortex variables become C variables with the same or predictable names (`int x`, `cortex_array* a`). Use the debugger's print/watch to inspect; for `any` values look at the generated `AnyValue` names.
4. **Assertions** — `assert(condition)` or `assert(condition, "message")` compile to runtime checks that abort; the debugger will stop at the failure so you can inspect the call stack.

Keep the generated C file when debugging and build it with `-g` and without aggressive optimization.

**Looks like C:** Yes. Cortex uses C-style syntax: `{}` blocks, `;` statements, `for`/`while`/`if`, type-first declarations, and functions. No line numbers, no `GOTO`.

**Easy like BASIC:** Partly. You get BASIC-style output: `print("Hello")`, `print(a, b)`, `writeline("%d", x)`, `printf` with format specifiers, plus `input_line()`, `read_file`/`write_file`, `vec2`/`vec3`, `var`/`any`, and built-in math/random/time. You still write declarations and braces, so it’s closer to “friendly C” than “classic BASIC.”

**Full C libraries:** Yes, with the usual C interop rules. You can:

- `#include <any_header.h>` (passed to the C compiler)
- Declare C APIs with `extern` (including pointer types: `void*`, `char*`, etc.)
- Use `-I`, `-L`, `-l` (or a config file) so the linker finds headers and libraries
- **Embed raw C** — put `#c <rest of line>` anywhere in a `.cx` file; the rest of the line is emitted verbatim into the generated C, so you can drop in C code (variables, statements, preprocessor) that compiles and runs without restrictions.

So you can call **any C library** (e.g. raylib, SDL, libcurl) from Cortex by including its headers and declaring the functions you need. C macros and types from headers are handled by the C compiler; pointer arguments and return values are supported in `extern` declarations.

---

## Feature Implementation Status

| Milestone | Feature | Status | Notes |
|-----------|---------|--------|-------|
| **2** | Dict Literals | Complete | `{ "key": value }` syntax, compile-time checking |
| **3** | 2D Array Literals | Complete | `[[1,2],[3,4]]` with bounds checking |
| **4** | Lambda Captures | Complete | No-capture and by-value capture `[x,y](params){}` |
| **5** | Nested JSON | Complete | `json_parse`, `json_stringify`, `as_dict`, `as_array` |
| **6** | parse_number/int | Complete | String to number conversion, 0 on failure |
| **7** | Named/Default Params | Complete | `fn f(x = 1)`, call with `f(y: 2)` |
| **8** | Coroutines | Partial | AST has `yield`, needs runtime scheduler |
| **9** | Async/Await | Partial | Keywords gated, compiles synchronously |
| **10** | Hot Reload | Planned | Stretch goal |
| — | Build System | Complete | Zero-config, cross-platform, TCC bundled |
| — | MSYS2 Support | Complete | Windows package manager integration |
| — | Pattern Matching | Complete | `match/case` with type guards |
| — | Multiple Returns | Complete | `(int,int) f()` and `return (a,b)` |
| — | String Interpolation | Complete | `"Hello ${name}"` |
| — | defer | Complete | `defer { cleanup() }` |
| — | Result Type | Complete | `result_ok`, `result_err`, `match` on Result |
| — | Struct Methods | Complete | Methods inside structs, `self->field` |

## What's still missing (vs. LANGUAGE_SPEC.md)

| Feature | Status |
|--------|--------|
| **Range-based for** | `for (x in arr)` — **implemented** |
| **Lambdas** | `[](a,b) -> int { ... }` — **implemented** |
| **Multiple return values** | `(int, int) f()` and `return (a, b)` — **implemented** |
| **String interpolation** | `"Hello ${name}"` — **implemented** |
| **Pattern matching** | `match (value) { case int n: ... default: ... }` — **implemented** |
| **Array literals** | `[1, 2, 3]` — **implemented** |
| **Dict literals** | `{ "key": value }` — **implemented** |
| **Named/Default params** | `fn f(int x = 0)` — **implemented** |
| **defer** | `defer { ... }` — **implemented** |
| **Async/actors** | Keywords gated; runtime is stretch |
| **Hot reloading** | Planned for future |

**The core for "games and apps" is complete:** types, control flow, functions, C interop, vectors, I/O, build system, and config-driven features.

---

## Language documentation

- **[LANGUAGE_GUIDE.md](LANGUAGE_GUIDE.md)** — Beginner-friendly guide: full explanation of the language with examples (variables, types, loops, functions, structs, enums, arrays, dicts, result, vectors, tests, C libraries).
- **LANGUAGE_SPEC.md** — Complete language specification (reference).

## Compiler Architecture

The compiler is split into clear layers and internal packages for maintainability:

- **Lexer** (`lexer.go`) — Tokenizes source code
- **Parser** (`parser.go`) — Builds AST
- **Semantic Analyzer** (`semantic.go`) — Scope and type checking
- **Code Generator** (`generator.go`) — Emits C
- **Compiler** (`compiler.go`, `main.go`) — Pipeline and CLI

**Internal packages (modular):**

- `internal/ast` — AST node definitions (single source of truth)
- `internal/config` — JSON config and feature flags
- `internal/errors` — Structured diagnostics (line, column, code, suggestion) for consistent error reporting
- `internal/optimizer` — AST passes before codegen; **constant folding** (e.g. `2+3` → `5`) runs when enabled

**Runtime (modular):**

- `runtime/core.c` / `core.h` — Core types, I/O, vectors, math, bounds check
- `runtime/game.c` / `game.h` — Game helpers (easing, lerp, rect, angle); linked when present

### AST Nodes

All AST nodes implement the `ASTNode` interface and are defined in `ast.go`:

- `ProgramNode` - Root node containing all declarations
- `FunctionDeclNode` - Function declarations
- `VariableDeclNode` - Variable declarations
- `StructDeclNode` - Struct definitions
- `BlockNode` - Code blocks
- Various expression and statement nodes

### Compilation Process

1. **Lexical Analysis**: Source code → Tokens
2. **Parsing**: Tokens → AST
3. **Semantic Analysis**: AST → Validated AST
4. **Code Generation**: AST → C code
5. **C Compilation**: C code → Executable

## Differences from C

### Removed Features

- **Pointers**: No pointer types, no `*`, `&`, or `->` operators
- **Manual memory management**: No `malloc`, `free`, etc. in Cortex code (use `extern` for C APIs that need them)
- **Preprocessor**: Cortex has `#include`, `#pragma link`, `#use`, `#define`, `#c <raw C line>` (passed through or emitted verbatim); no full C preprocessor

### Added Features

- **Built-in string type**: First-class string support
- **Vector types**: `vec2` and `vec3` for game development
- **Boolean type**: Native `bool` type
- **Enhanced I/O**: `print`, `writeline`, `printf` (C-style format specifiers), and raw C via `#c`
- **Game utilities**: Random numbers, time functions, vector math
- **Smart dynamic variables**: Type inference with `var`
- **Universal type**: `any` type for dynamic typing
- **Quality of life features**: 
  - String interpolation
  - Pattern matching simulation
  - List comprehensions simulation
  - Function overloading simulation
  - Auto-initialization
  - Type-safe operations with automatic conversion
  - Mixed type arrays

## Error Handling

The compiler provides detailed error messages with line and column information:

```
Compilation error: semantic analysis found 1 errors
lexical error: unexpected character '@' at line 5, column 10
syntax error: expected ';' after expression at line 10, column 5
semantic error: undefined variable 'x' at line 15, column 3
```

## Contributing

Feel free to contribute to the SimpleC compiler by:

1. Reporting bugs
2. Suggesting new features
3. Submitting pull requests
4. Improving documentation

## License

This project is open source. See LICENSE file for details.

## Implemented (modular, best practices)

- **Array bounds checking** — Runtime `cortex_bounds_check(len, index, line)` in `runtime/core.c`; codegen wraps identifier array access (e.g. `arr[i]`) with a check when `arr_len` exists. Aborts with a clear message on out-of-bounds.
- **More game builtins** — Modular `runtime/game.c` and `runtime/game.h`: easing (`ease_quad_in`, `ease_quad_out`, `ease_cubic_in`, etc.), `vec2_lerp`/`vec3_lerp`, rect helpers (`rect_contains_point`, `rect_overlaps`), `vec2_angle`/`vec2_from_angle`. Linked automatically when present; declared in semantic builtins.
- **Optimizer** — `internal/optimizer` runs after semantic analysis; constant folding (e.g. `2+3` → `5`) is applied before codegen. Kept modular so more passes can be added without touching codegen.
- **Structured diagnostics** — `internal/errors`: `Diagnostic` (severity, code, line, column, message, suggestion), `Collector`, and formatted output. Ready for use in lexer/parser/semantic for consistent, tooling-friendly errors.
- **Package system & import resolution** — Multi-file: `cortex -i main.cx -i lib.cx -o app`. Use `import "mylib";` in source; the compiler resolves `mylib` to `./mylib.cx` or `./mylib/mod.cx` (relative to the file containing the import) and merges them into one program. No circular imports; each file is merged once.
- **Collections** — Dynamic array (`array` type: `array_create`, `array_push`, `array_get`, `array_set`, `array_len`, `array_free`) and dictionary (`dict` type: `dict_create`, `dict_set`, `dict_get`, `dict_has`, `dict_len`, `dict_free`) with a clean API in `runtime/core.c` (QOL).
- **Result type** — `result_ok(value)`, `result_err("msg")`, `result_is_ok(r)`, `result_value(r)`, `result_error(r)` for error-handling without exceptions.
- **Lambdas** — No-capture lambdas: `[](params) [-> returnType] { body }` emitted as static C functions for callbacks and event handlers.
- **Debugging** — README documents how to debug Cortex programs using the generated C and a native debugger (GDB/LLDB/VS).
- **Dynamic list API** — `array_pop`, `array_insert`, `array_remove_at`, `array_capacity`, `array_reserve` in runtime and compiler.
- **Events** — `event` type: `event_create`, `event_subscribe`, `event_unsubscribe`, `event_emit`, `event_free` (pairs with lambdas).
- **Standard library** — String: `str_split`, `str_join`, `str_replace`, `str_trim`, `starts_with`, `ends_with`, `to_lower`, `to_upper`. Math: `clamp_int`, `pow`, `random_choice`. File: `file_exists`, `list_dir`, `path_join`. Debug: `debug_log`, `debug_assert`, `dump`. JSON stubs: `json_parse`, `json_stringify`.
- **Unit tests** — `test "name" { }` compiles to registered test functions; `assert_eq(a, b)`, `assert_approx(a, b, eps)`; `test_run_all()` runs all tests.
- **Match on Result** — `match (r) { case Ok(v): ... case Err(e): ... }`; `v` as `AnyValue`, `e` as error string.
- **JSON** — Minimal `json_parse(s)` (object with string keys, number/string/bool/null values) and `json_stringify_any` / `json_stringify_dict` in runtime.
- **Structured diagnostics** — Semantic analyzer supports `SetDiagnosticsCollector(*errors.Collector)`; undefined-identifier and feature-gate errors emit line/column/code/suggestion for tooling.
- **Struct methods** — Define methods inside a struct: `struct Player { int x; int y; void move(int dx) { x = x + dx; y = y + dx; } }` and call with `player.move(5)`; field names in the method body refer to the receiver (emitted as `self->field` in C).

## Modular design & best practices

- **Layered architecture:** Lexer → Parser → Semantic → Optimizer → Codegen → C compiler. Each stage is testable and replaceable.
- **Internal packages:** `internal/ast` (single AST definition), `internal/config` (feature flags, paths), `internal/errors` (diagnostics), `internal/optimizer` (optional passes). Keeps the compiler maintainable.
- **Runtime modules:** `runtime/core.c` (types, I/O, math, bounds check); `runtime/game.c` (easing, lerp, rect, angle) when present. Feature flags gate optional builtins.
- **Coding practices:** One responsibility per file; avoid globals; use structured errors; document public APIs. For Cortex code: prefer `var` when the type is obvious; use `struct` and `enum` for clarity; use `break`/`continue` and `switch` for clear control flow; keep functions small and use `defer` for cleanup.

## Future Enhancements

| Feature | Priority | Status |
|---------|----------|--------|
| IDE integration (LSP, diagnostics) | High | 🔄 Planned |
| Coroutines full implementation | Medium | 🔄 AST ready |
| Async/await runtime | Medium | 🔄 Keywords gated |
| Hot reloading | Low | 🔄 Stretch goal |

**Completed Milestones:**
- ~~Dict literals~~ — **Done**
- ~~2D array literals~~ — **Done**
- ~~Lambda captures~~ — **Done**
- ~~Nested JSON~~ — **Done**
- ~~parse_number/int~~ — **Done**
- ~~Named/default parameters~~ — **Done**
- ~~Modules/namespaces~~ — **Done**
- ~~ECS helpers~~ — **Done**
- ~~Build system~~ — **Done**
- ~~MSYS2 integration~~ — **Done**
