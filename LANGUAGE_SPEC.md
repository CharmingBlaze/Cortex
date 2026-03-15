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

### Optional Types
Optional types represent a value that may or may not exist:
```c
int? maybe_int;       // Either an int or null
string? maybe_name;   // Either a string or null
User? maybe_user;     // Either a User struct or null
```

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
```c
if (condition) {
    // code
} else if (condition) {
    // code
} else {
    // code
}
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

// Do-while loop
do {
    // code
} while (condition);

// Repeat loop (simple counted loop)
repeat (10) {
    // runs 10 times
}
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

// Single expression body
(x) => x * 2

// Block body
(x, y) => {
    int sum = x + y;
    return sum;
}

// Used as callbacks
gui_button("Click", () => println("Clicked!"));
```

### Match Expressions
Pattern matching for type-safe branching:
```c
match value {
    int x => println("int: ${x}");
    string s => println("string: ${s}");
    _ => println("unknown");
}

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
```c
extern int printf(string fmt, ...);
extern double sqrt(double x);
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
