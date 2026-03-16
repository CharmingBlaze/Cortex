# Cortex Language Specification

## CORTEX LANGUAGE VISION

### Overall Goal
Cortex is designed to feel like the perfect blend of four worlds:

**C — structure and performance**
- Blocks, functions, types
- Predictable execution
- Low level control without the footguns

**TypeScript — ergonomics and expressiveness**
- Clean syntax
- Implicit self
- Optional types
- Union types
- Async/await

**Go — simplicity and concurrency**
- defer
- channels
- coroutines
- simple thread spawning

**Swift/Kotlin — readability and modern design**
- Dot syntax
- Constructors
- Enums
- Pattern matching

This combination makes Cortex easy to learn, pleasant to write, and powerful enough for real systems programming.

---

## Build System

### CLI Commands

Cortex provides a modern, simple CLI inspired by Go and Rust:

| Command | Description |
|---------|-------------|
| `cortex new <name>` | Create a new project with cortex.toml |
| `cortex run [file.cx]` | Compile and run (uses cortex.toml if found) |
| `cortex build [file.cx] [-o output]` | Compile to executable |
| `cortex bind <lib> -i <header.h>` | Generate Cortex bindings from C header |
| `cortex -i file.cx -run` | Invokes C compiler (Zig CC bundled, or system GCC/Clang) |
| `cortex -i file.cx -o output -use raylib` | Legacy: compile with library |

### C Library Binding Generator

Cortex can automatically generate bindings from C headers:

```bash
# Generate bindings from a C library
cortex bind raylib -i third_party/raylib/src/raylib.h

# Output: bindings/raylib.cx with functions, structs, enums, constants
```

The binder:
- Parses C function declarations → Cortex `extern` functions
- Parses C structs → Cortex `struct` definitions
- Parses C enums → Cortex constants
- Parses `#define` constants → Cortex `const` declarations
- Auto-detects cleanup functions for memory management

### Project Configuration (cortex.toml)

Cortex uses TOML for project configuration. Create a `cortex.toml` in your project root:

```toml
[project]
name = "my_game"
version = "0.1.0"
entry = "main.cx"           # Entry point (default: main.cx)
backend = "auto"            # C backend: zig, or auto

[project.features]
async = true                # Enable async/await
actors = true               # Enable actor model
qol = true                  # Enable quality-of-life features

[dependencies.raylib]
path = "third_party/raylib" # Auto-detect include/lib paths
# Or specify explicitly:
# include_path = "third_party/raylib/src"
# lib_path = "third_party/raylib/build/raylib"
libs = ["raylib", "opengl32", "gdi32", "winmm", "shell32"]
```

### Automatic Library Detection

When you specify `path` in a dependency, Cortex automatically:
- Adds `<path>/include` and `<path>/src` to include paths
- Adds `<path>/lib` and `<path>/build/<libname>` to library paths
- Links the library and required system libraries

### Example: Raylib Game

```bash
# Create project
cortex new my_game
cd my_game

# Add raylib dependency to cortex.toml
# Then simply:
cortex run
```

No flags. No paths. No pain.

---

## Overview
Cortex is a modern systems programming language that combines C's performance with modern ergonomics. It removes pointers and manual memory management while adding TypeScript-style type features, Go-style concurrency, and Swift-style readability.

## Key Differences from C
- **No pointers**: All memory management is automatic via managed handles
- **Dot syntax everywhere**: Use `.` for all member access (no `->`)
- **Modern type system**: Optional types (`int?`), union types (`int | float | string`)
- **Implicit self**: No need to write `self.field` in struct methods
- **Defer**: Go-style cleanup with `defer expr;`
- **Async/await**: TypeScript-style async functions
- **Channels**: Built-in concurrency with method syntax

## Types

### Primitive Types
- `int` - 32-bit integer
- `float` - 32-bit floating point
- `double` - 64-bit floating point
- `char` - 8-bit character
- `bool` - boolean (true/false)
- `string` - immutable string type
- `vec2` - 2D vector (float x, float y)
- `vec3` - 3D vector (float x, float y, float z)

### Composite Types
- `array[T]` - Dynamic array of type T
- `struct` - User-defined structures
- `enum` - Enumeration types (including string enums)
- `var` - Smart dynamic variable with type inference
- `any` - Universal type that can hold any value

#### Array Type Syntax
Arrays can be declared using C-style syntax with square brackets:
```c
int[] nums = {1, 2, 3, 4, 5};      // Dynamic array of ints
string[] names = {"Alice", "Bob"};  // Dynamic array of strings
int[][] matrix = {{1,2}, {3,4}};    // 2D array (array of arrays)

// Fixed-size arrays
int arr[5];                         // Array with 5 elements
arr[0] = 10;
```

#### Array Literals
Arrays can be initialized using curly brace literals:
```c
int[] nums = {1, 2, 3, 4, 5};
string[] names = {"Alice", "Bob", "Charlie"};
```

### Optional Types
Optional types represent a value that may or may not exist:
```c
int? maybe_int;       // Either an int or null
string? maybe_name;   // Either a string or null
User? maybe_user;     // Either a User struct or null
```

#### Optional Operators
Cortex provides postfix operators for working with optionals:

**Postfix `?` - Optional Check**
Returns `true` if the optional has a value, `false` otherwise:
```c
int? maybe = 42;
if (maybe?) {
    println("Has a value!");
}
if (!(maybe?)) {
    println("No value!");
}
```

**Postfix `!` - Force Unwrap**
Returns the value inside the optional (use only when you know it has a value):
```c
int? maybe = 42;
if (maybe?) {
    int value = maybe!;  // Unwraps to 42
    println(value);
}
```

These operators work with the underlying `cortex_optional_T` struct representation that has `has_value` and `value` fields.

### Union Types
Union types allow a variable to hold one of several types:
```c
fn parse(string input) -> int | float | string;

int | float | string result;
result = 42;
result = 3.14;
result = "hello";
```

## Control Structures

### Conditionals

#### Optional Parentheses in if Statements
Cortex allows you to write if statements with or without parentheses around the condition. Both of these are valid:

```c
if (x > 5) { ... }
if x > 5 { ... }
```

**When parentheses are useful:**
- Complex expressions
- Mixing operators
- When you want C-like clarity

**When parentheses can be omitted:**
- Simple comparisons
- Clean, readable conditions
- When writing modern Cortex-style code

#### Single-Statement Bodies
Cortex allows single-statement bodies without braces:

```c
if x > 0 println("positive");
```

This keeps simple logic compact.

#### elif - Sugar for else if
Cortex provides `elif` as cleaner syntax for `else if` chains:

```c
if x == 5 {
    println("five");
} elif x == 10 {
    println("ten");
} elif x == 15 {
    println("fifteen");
} else {
    println("other");
}
```

This matches Python, Swift, and Kotlin style.

#### unless - Inverse of if
`unless` executes its body when the condition is **false**:

```c
unless x > 5 {
    println("x is NOT greater than 5");
}

// Equivalent to:
if !(x > 5) {
    println("x is NOT greater than 5");
}
```

`unless` supports single-statement bodies and `else`:

```c
unless x < 0 println("x is non-negative");

unless x == 10 {
    println("x is NOT 10");
} else {
    println("x IS 10");
}
```

#### if let - Pattern Matching for Optionals
`if let` binds a value only if it exists (not null):

```c
string maybe_name = get_name();
if let name = maybe_name {
    println("Hello, ${name}!");
} else {
    println("No name provided");
}
```

This is sugar for:
```c
if maybe_name != null {
    var name = maybe_name;
    println("Hello, ${name}!");
}
```

#### Else and Else-If
All combinations are allowed:

```c
if x < 0 println("neg");
else println("not neg");

if x == 5 println("five");
else if x == 10 println("ten");
else println("other");
```

Or with braces:

```c
if x == 5 {
    println("five");
} else if x == 10 {
    println("ten");
} else {
    println("other");
}
```

#### Full Example
```c
void main() {
    int x = 10;

    // Traditional (C-style)
    if (x > 5) {
        println("x > 5");
    }

    // Modern Cortex style — no parentheses
    if x > 5 {
        println("x > 5");
    }

    // Single-statement body
    if x > 0 println("positive");

    // Mixed style is allowed
    if (x < 20) println("x < 20");

    // elif chain
    if x == 5 {
        println("five");
    } elif x == 10 {
        println("ten");
    } else {
        println("other");
    }

    // unless
    unless x < 0 println("non-negative");
}

### Constants
```c
// Const with type inference
const greeting = "Hello";
const max_items = 100;

// Const with explicit type
const int max_score = 100;
const float pi = 3.14159;

// Const in game context
const paddle_speed = 5;
const paddle_height = 80;
```

### Loops
```c
// Traditional for loop
for (int i = 0; i < 10; i++) {
    // code
}

// Range-based for loop
for (value in array) {
    // code
}

// While loop
while (condition) {
    // code
}

// Infinite loop sugar
loop {
    // runs forever until break
    if (should_exit) break;
}

// Do-while loop
do {
    // code
} while (condition);

// Repeat loop (simple counted loop)
repeat (10) {
    // runs 10 times
}
```

### Operators

#### Compound Assignment Operators
Cortex supports shorthand assignment operators that modify a variable in place:

| Long form | Short form | Meaning |
|-----------|------------|---------|
| `x = x + y` | `x += y` | Add to x |
| `x = x - y` | `x -= y` | Subtract from x |
| `x = x * y` | `x *= y` | Multiply x |
| `x = x / y` | `x /= y` | Divide x |

```c
int score = 0;
score += 10;      // score is now 10
score -= 5;       // score is now 5
score *= 2;       // score is now 10
score /= 2;       // score is now 5
```

#### Increment and Decrement
Cortex supports the familiar increment/decrement operators:

```c
int i = 0;
i++;              // Postfix increment: i is now 1
i--;              // Postfix decrement: i is now 0
++i;              // Prefix increment: i is now 1
--i;              // Prefix decrement: i is now 0
```

These are especially useful in loops:

```c
for (int i = 0; i < 10; i++) {
    println(i);
}

// Or in game logic
player.score++;
ball.speed *= 1.1;
```

### Functions
```c
// Function declaration
int add(int a, int b) {
    return a + b;
}

// Function with multiple return values
(int, int) get_position() {
    return (10, 20);
}

// Named and default parameters
void greet(string name = "World", int count = 1) {
    for (int i = 0; i < count; i++) {
        println("Hello, " + name + "!");
    }
}

greet();                    // Hello, World! (once)
greet("Alice");             // Hello, Alice! (once)
greet(count: 3);            // Hello, World! (3 times)
greet("Bob", 2);            // Hello, Bob! (2 times)
```

### Arrow Functions
Arrow functions provide concise syntax for callbacks and lambdas:
```c
// No parameters
() => println("Clicked!")

// Single expression body with untyped parameters
(x) => x * 2

// Typed parameters
(int a, int b) => a + b

// Block body
(x, y) => {
    int sum = x + y;
    return sum;
}

// Used as callbacks
gui_button("Click", () => println("Clicked!"));

// Assigned to variables with fn type inference
fn add = (int a, int b) => a + b;
fn mul = (int x, int y) => x * y;
printf("add(3,4)=%d\n", add(3, 4));
```

### Match Expressions
Pattern matching for type-safe branching:
```c
match value {
    int x => println("int: ${x}"),
    string s => println("string: ${s}"),
    _ => println("unknown"),
}

// Match as an expression
int n = 2;
string result = match n {
    1 => "one",
    2 => "two",
    _ => "other"
};
printf("result: %s\n", result);  // prints "two"

// Nested match expressions
int x = 1, y = 2;
string nested = match x {
    1 => match y {
        1 => "1,1",
        2 => "1,2",
        _ => "1,_"
    },
    _ => "other"
};

// With guards
match result {
    case Ok(value) => println("Success: ${value}");
    case Err(msg) => println("Error: ${msg}");
}

// On enums
enum Color { Red, Green, Blue }

match color {
    Red => println("It's red!");
    Green => println("It's green!");
    Blue => println("It's blue!");
}
```

## Structs

### Struct Declaration
```c
struct Player {
    string name;
    int health;
    int score;

    // Method with implicit self
    void damage(int amount) {
        health -= amount;  // No self. needed
        if (health < 0) {
            health = 0;
        }
    }

    int get_health() {
        return health;  // Implicit self
    }
}
```

### Implicit Self
Inside struct methods, `self` is implicitly available. Field access does not require `self.` unless needed for disambiguation:
```c
struct Counter {
    int count;

    void increment() {
        count++;  // Same as self.count++
    }

    void set_count(int count) {
        self.count = count;  // Disambiguation needed
    }
}
```

### Dot Syntax
Use `.` for all member access - no `->` needed:
```c
Player p;
p.name = "Hero";
p.health = 100;
p.damage(10);  // Method call
```

### Struct Constructors
Every struct automatically gets two constructor styles:

**Positional constructor:**
```c
let p = Player("Hero", 100, 0);
```

**Named field constructor:**
```c
let p = Player { name: "Hero", health: 100, score: 0 };
```

**Shorthand field syntax:**
```c
string name = "Hero";
int health = 100;
let p = Player { name, health };  // Same as { name: name, health: health }
```

## Enums

### Simple Enums
```c
enum Color { Red, Green, Blue }

int c = Red;  // No Color:: prefix needed

switch (c) {
    case Red: println("red"); break;
    case Green: println("green"); break;
    case Blue: println("blue"); break;
}
```

### String Enums
```c
enum Status {
    Ok = "ok",
    Error = "error",
    Pending = "pending"
}

Status s = Ok;
println(s);  // prints "ok"
```

## Defer

Go-style defer for guaranteed cleanup:
```c
// Single expression defer
defer close_file(file);

// Deferred expressions run in LIFO order
void process() {
    defer println("Cleanup 1");
    defer println("Cleanup 2");
    println("Processing");
}
// Output: Processing, Cleanup 2, Cleanup 1
```

## Async/Await

TypeScript-style async functions:
```c
async fn fetch() {
    println("Fetching...");
    await sleep(1000);
    println("Done!");
}

fn main() {
    await fetch();
}
```

## Coroutines

Coroutines with simplified yield:
```c
coroutine generator() {
    int i = 0;
    while (i < 10) {
        yield;  // No co_yield() needed
        i++;
    }
}
```

## Thread Spawning

Simple thread spawning without null arguments:
```c
// Old style (deprecated)
spawn heavy_computation(null);

// New style
let t = spawn heavy_computation();
await t;

// With assignment
spawn worker(ch);
```

## Channels

Channels with method syntax:
```c
let ch = channel<int>(10);

// Send
ch.send(42);

// Receive
let value = ch.recv();

// In a spawn
spawn producer(ch);
spawn consumer(ch);
```

## Cleanup Annotations

Automatic memory safety for C interop:
```c
extern void* malloc(int size) cleanup(free);
extern void free(void* ptr);

extern Texture* load_texture(string path) cleanup(destroy_texture);
extern void destroy_texture(Texture* t);

fn main() {
    let buf = malloc(1024);  // Auto-cleaned on scope exit
    let tex = load_texture("hero.png");  // Auto-cleaned
}  // free(buf) and destroy_texture(tex) called automatically
```

## C Interop

Cortex works seamlessly with C libraries:

### Including C Headers
```c
#include <stdio.h>
#include <math.h>
#pragma link("m")
```

### External Function Declarations

**Auto-Extern**: Cortex automatically generates extern declarations for C functions when you include headers:

```c
#include <stdio.h>
#include <stdlib.h>

fn main() {
    // Just call C functions - extern is auto-generated!
    var buf = malloc(1024);
    printf("Buffer allocated\n");
    free(buf);
}
```

**Manual extern** is optional but needed for:
- Cleanup annotations (`cleanup(free)`)
- Specific return types (default is `int`)
- Pointer return types

```c
extern int printf(string fmt, ...);
extern double sqrt(double x);

// With cleanup for automatic memory management
extern void* malloc(int size) cleanup(free);
```

### Managed C Handles
When a C function returns a pointer, Cortex wraps it in a managed handle that automatically frees itself:
```c
extern void* malloc(int size) cleanup(free);

fn main() {
    let buf = malloc(1024);  // Managed handle
    // No manual free needed - automatic cleanup
}
```

### Borrowed Pointers
If a C function returns a pointer that should not be freed, omit the cleanup annotation:
```c
extern const char* getenv(string name);  // Borrowed, not freed
```

### Callbacks
Cortex functions can be passed to C as callbacks:
```c
extern void register_callback(fn(int) cb);

fn on_event(int code) {
    println("Event: ${code}");
}

register_callback(on_event);
```

## Built-in Functions

### Math Functions
```c
float sqrt(float x);
float sin(float x);
float cos(float x);
float abs(float x);
```

### Vector Operations
```c
vec2 make_vec2(float x, float y);
vec3 make_vec3(float x, float y, float z);
float dot(vec2 a, vec2 b);
vec2 normalize(vec2 v);
```

### Random Numbers
```c
int random_int(int min, int max);
float random_float(float min, float max);
```

### Time
```c
float get_time();
void sleep(float seconds);
```

### Input/Output
```c
void print(string message);
void println(string message);
string input();
string read_file(string path);
void write_file(string path, string content);
```

### Type Functions
```c
string type_of(any value);
bool is_type(any value, string type_name);
int as_int(any value);
float as_float(any value);
string as_string(any value);
bool as_bool(any value);
```

## Smart Dynamic Variables

### Type Inference with `var`
```c
var x = 42;           // x becomes int
var y = 3.14;         // y becomes float
var name = "Hello";   // name becomes string
var flag = true;      // flag becomes bool
var pos = make_vec2(1.0, 2.0); // pos becomes vec2
```

### String Interpolation
```c
var name = "World";
var msg = "Hello, ${name}!";  // String interpolation
var count = 42;
var info = "Count: ${count}";
```

## Example Program
```c
struct Player {
    string name;
    int health;
    int score;

    void damage(int amount) {
        health -= amount;
        if (health < 0) health = 0;
    }
}

enum Status { Ok, Error, Pending }

async fn load_game() {
    println("Loading...");
    await sleep(1000);
    println("Ready!");
}

fn main() {
    // Struct with named constructor
    let p = Player { name: "Hero", health: 100, score: 0 };
    
    // Or positional
    let p2 = Player("Hero2", 100, 0);
    
    // Method call with implicit self
    p.damage(10);
    println("${p.name} has ${p.health} HP");
    
    // Enum access
    Status s = Ok;
    match s {
        Ok => println("OK!"),
        Error => println("Error!"),
        Pending => println("Pending..."),
    }
    
    // Async
    await load_game();
    
    // Channel
    let ch = channel<int>(10);
    ch.send(42);
    let value = ch.recv();
    println("Received: ${value}");
}
```
