# Cortex Language Specification

## Overview
Cortex is a C-like programming language designed for game and application development. It maintains C syntax but removes pointers and adds modern features for easier development.

## Key Differences from C
- **No pointers**: All memory management is automatic
- **Built-in types**: Enhanced primitive types
- **Game-friendly features**: Built-in support for common game operations
- **Modern syntax**: Some quality-of-life improvements

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
- `enum` - Enumeration types
- `var` - Smart dynamic variable with type inference
- `any` - Universal type that can hold any value

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

// Lambda/anonymous function
var multiply = [](int a, int b) -> int {
    return a * b;
};
```

## Game-Specific Features

### Built-in Functions
```c
// Math functions
float sqrt(float x);
float sin(float x);
float cos(float x);
float abs(float x);

// Vector operations
vec2 make_vec2(float x, float y);
vec3 make_vec3(float x, float y, float z);
float dot(vec2 a, vec2 b);
vec2 normalize(vec2 v);

// Random numbers
int random_int(int min, int max);
float random_float(float min, float max);

// Time
float get_time();
void sleep(float seconds);
```

### Input/Output
```c
// Console I/O
void print(string message);
void println(string message);
string input();

// File I/O
string read_file(string path);
void write_file(string path, string content);
```

### Smart Dynamic Variables

SimpleC supports smart dynamic variables with automatic type inference:

### Type Inference with `var`
```c
// Type is automatically inferred from the initializer
var x = 42;           // x becomes int
var y = 3.14;         // y becomes float
var name = "Hello";   // name becomes string
var flag = true;      // flag becomes bool
var pos = make_vec2(1.0, 2.0); // pos becomes vec2

// Variables can change type (dynamic behavior)
var dynamic = 10;     // Initially int
dynamic = "now string"; // Now string
dynamic = 3.14;       // Now float
```

### Universal Type `any`
```c
// Explicitly declare variables that can hold any type
any value = 42;
value = "hello";
value = make_vec2(1.0, 2.0);

// Type checking
if (is_type(value, "int")) {
    println("Value is an integer: " + as_int(value));
} else if (is_type(value, "string")) {
    println("Value is a string: " + as_string(value));
}
```

### Smart Variable Features
```c
// Auto-initialization
var counter;          // Automatically initialized to 0 for numbers, false for bool, "" for string

// Type-safe operations
var a = 10;
var b = 20;
var sum = a + b;      // Works because both are numbers

// Mixed type operations (automatic conversion)
var num = 42;
var str = "The answer is " + num; // Automatic conversion to string

// Array with dynamic elements
var items = [1, "hello", true, 3.14]; // Mixed type array
```

### Built-in Type Functions
```c
string type_of(any value);     // Get type name as string
bool is_type(any value, string type_name); // Type checking
int as_int(any value);         // Cast to int (with validation)
float as_float(any value);     // Cast to float
string as_string(any value);   // Cast to string
bool as_bool(any value);       // Cast to bool
```

### Advanced Smart Features

#### Intelligent Type Inference
```c
// Context-aware type inference
var result = 10 + 3.14;    // result becomes float (int + float)
var text = "Value: " + 42;  // text becomes string (concatenation)
var flag = 5 > 3;         // flag becomes bool (comparison)
var size = strlen("hello"); // size becomes size_t (function return type)
```

#### Automatic Type Promotion
```c
var int_val = 42;
var float_val = int_val + 0.5; // int automatically promoted to float
var result = int_val * 2.0;     // result is float

// Safe numeric operations
var small = 127;
var large = small + 1;       // Still fits in int
var overflow = 2147483647 + 1; // Smart overflow detection
```

#### Smart String Operations
```c
var name = "World";
var greeting = "Hello " + name;           // String concatenation
var message = "Count: " + 42;              // Auto number-to-string
var formatted = "Value: ${name} is ${42}"; // String interpolation
```

#### Type-Aware Collections
```c
// Smart arrays with type inference
var numbers = [1, 2, 3, 4, 5];           // Inferred as int[]
var strings = ["a", "b", "c"];            // Inferred as string[]
var mixed = [1, "hello", true];           // any[] for mixed types

// Type-safe array operations
var nums = [1, 2, 3];
var sum = nums.sum();                     // Type-aware: returns int
var count = nums.length();                // Always returns int

// Smart filtering
var evens = nums.filter(x => x % 2 == 0);  // Returns int[]
var texts = strings.filter(s => s.length > 2); // Returns string[]
```

#### Pattern Matching on Types
```c
any process_data(any data) {
    match data {
        case int n when n > 0:
            return "Positive integer: " + n;
        case int n when n < 0:
            return "Negative integer: " + n;
        case string s:
            return "String: " + s;
        case vec2 v:
            return "Vector: (" + v.x + ", " + v.y + ")";
        case null:
            return "Null value";
        default:
            return "Unknown type: " + type_of(data);
    }
}
```

## Features C Developers Always Wanted

Cortex brings many long-requested improvements over traditional C while keeping the familiar feel:

1. **Modules & Imports** – Replace header juggling with `module math.core;` and `import math.vector;`.
2. **Generics & Templates** – Write `vector<int>` or `optional<string>` natively, no macros needed.
3. **Pattern Matching** – Expressive `match` blocks with guards on enums and structs.
4. **Defer Blocks** – `defer { fclose(file); }` guarantees cleanup on every exit path.
5. **Optional & Result Types** – `optional<T>` and `result<T, E>` eliminate null-pointer guessing.
6. **Async/Await** – Built-in async functions with `await` for ergonomic concurrency.
7. **Immutable by Default** – `let` declares read-only bindings; `var` is used when mutation is intended.
8. **Safe Concurrency** – Actors, channels, and race detection tools are part of the runtime.
9. **First-Class Testing** – `test "vector push" { ... }` blocks integrate with `cortex --test`.
10. **Package & Build Profiles** – `cortex.toml` manages dependencies, versions, and build flags.

All of these features coexist with Cortex’s smart dynamic typing, automatic memory management, and seamless C interop, so C developers gain modern productivity without abandoning their existing ecosystem.

### Additional Quality-of-Life Enhancements

| Feature | What it solves | Cortex syntax |
| --- | --- | --- |
| **Auto Imports** | No more header guard boilerplate | `import net.http;` automatically resolves modules/files |
| **Attributes** | Declarative metadata for functions/types | `@[inline, test_only]` |
| **Pipeline Operator** | Cleaner data transformations | `data |> filter(active) |> map(toDTO)` |
| **Compile-Time Evaluation** | Safer constants and lookup tables | `const table = comptime build_table();` |
| **String Interpolation** | Readable formatting without `printf` gymnastics | `let msg = `"${player.name} HP=${player.hp}"`;` |
| **Named & Default Params** | No more placeholder arguments | `spawn_enemy(hp: 100, speed: 4.5);` |
| **First-Class Tests** | Zero-config unit tests | `test "math clamps" { assert(clamp(5,0,3) == 3); }` |
| **CLI Flags DSL** | Built-in argument parsing | `flags { bool verbose; string output = "out.bin"; }` |
| **Hot Reload Hooks** | Patch code in live apps | `hot_reload { reload_shaders(); }` |
| **Doc Comments** | Markdown docs emitted automatically | `/// Renders a frame` |

These QoL features are lightweight (no runtime tax) and compile down to familiar C constructs, so teams can adopt them incrementally.

#### Smart Function Overloading
```c
// Functions that adapt based on input types
string describe(any value) {
    if (is_type(value, "int")) {
        return "Integer: " + value;
    } else if (is_type(value, "float")) {
        return "Float: " + value;
    } else if (is_type(value, "string")) {
        return "String: " + value;
    }
    return "Unknown type";
}

// Generic-like functions
any add(any a, any b) {
    if (is_type(a, "number") && is_type(b, "number")) {
        return as_float(a) + as_float(b); // Promote to float
    }
    if (is_type(a, "string") || is_type(b, "string")) {
        return as_string(a) + as_string(b); // String concatenation
    }
    return null; // Incompatible types
}
```

#### Smart Error Handling
```c
// Safe type conversion with error handling
int safe_int(any value) {
    if (is_type(value, "int")) {
        return as_int(value);
    } else if (is_type(value, "float")) {
        int result = (int)as_float(value);
        if (as_float(value) != result) {
            println("Warning: Precision loss in float to int conversion");
        }
        return result;
    } else if (is_type(value, "string")) {
        int result = parse_int(as_string(value));
        if (result == null) {
            println("Error: Cannot convert string to int");
        }
        return result;
    }
    return 0; // Default fallback
}
```

#### Intelligent Memory Management
```c
// Smart resource cleanup
var file = fopen("data.txt", "r");
try {
    var content = read_file(file);
    process(content);
} finally {
    if (file != null) {
        fclose(file); // Automatically called
    }
}

// Smart pointer simulation
var smart_ptr = create_smart_ptr(malloc(100));
// Automatically frees when out of scope
```

## External C Library Support

Cortex supports seamless integration with any C library through the `extern` keyword and library directives.

### Including C Headers

```c
// Include standard C libraries
#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <string.h>

// Include custom C headers
#include "my_library.h"
```

### External Function Declarations

```c
// Declare external C functions
extern int printf(char* format, ...);
extern void* malloc(size_t size);
extern void free(void* ptr);
extern double sin(double x);

// Declare functions from custom libraries
extern int my_custom_function(int param);
extern void process_data(char* data);
```

### Library Linking

```c
// Link with specific libraries using #pragma directives
#pragma link("m")        // Link with math library
#pragma link("pthread")  // Link with pthread library
#pragma link("sqlite3")  // Link with SQLite library
#pragma link("mylib")    // Link with custom library

// Or use library declarations
library "m" {
    extern double cos(double x);
    extern double sqrt(double x);
}

library "sqlite3" {
    extern int sqlite3_open(char* filename, void** db);
    extern int sqlite3_exec(void* db, char* sql, void* callback, void* arg, char** errmsg);
}
```

### Using C Libraries

```c
#include <stdio.h>
#include <math.h>
#pragma link("m")

void main() {
    // Use standard C library functions
    printf("Hello from C library!\n");
    
    // Use math library functions
    double result = sin(3.14159 / 2);
    printf("sin(pi/2) = %f\n", result);
    
    // Use memory allocation
    char* buffer = (char*)malloc(100);
    if (buffer != null) {
        strcpy(buffer, "Cortex + C Libraries!");
        printf("Buffer: %s\n", buffer);
        free(buffer);
    }
}
```

### Advanced Library Integration

```c
// OpenGL integration
#include <GL/gl.h>
#pragma link("GL")

void render_triangle() {
    glBegin(GL_TRIANGLES);
    glVertex3f(-1.0f, -1.0f, 0.0f);
    glVertex3f(1.0f, -1.0f, 0.0f);
    glVertex3f(0.0f, 1.0f, 0.0f);
    glEnd();
}

// SDL integration
#include <SDL2/SDL.h>
#pragma link("SDL2")

void init_sdl() {
    if (SDL_Init(SDL_INIT_VIDEO) < 0) {
        printf("SDL initialization failed!\n");
        return;
    }
    
    SDL_Window* window = SDL_CreateWindow(
        "Cortex + SDL", 
        SDL_WINDOWPOS_CENTERED, 
        SDL_WINDOWPOS_CENTERED, 
        800, 600, 
        SDL_WINDOW_SHOWN
    );
    
    // ... use SDL functions
    SDL_DestroyWindow(window);
    SDL_Quit();
}
```

### Library Configuration

```c
// Compiler configuration for libraries
config {
    compiler: "gcc",
    flags: ["-O2", "-Wall"],
    libraries: ["m", "pthread", "sqlite3"],
    include_paths: ["/usr/local/include"],
    library_paths: ["/usr/local/lib"]
}

// Platform-specific library loading
#ifdef WINDOWS
    #pragma link("opengl32")
    #pragma link("ws2_32")
#elif MACOS
    #pragma link("-framework OpenGL")
    #pragma link("-framework CoreFoundation")
#elif LINUX
    #pragma link("GL")
    #pragma link("pthread")
#endif
```

### Custom Library Wrappers

```c
// Create Cortex-friendly wrappers for C libraries
wrapper "sqlite" {
    #include <sqlite3.h>
    #pragma link("sqlite3")
    
    struct Database {
        void* handle;
    }
    
    Database* sqlite_open(string filename) {
        Database* db = (Database*)malloc(sizeof(Database));
        if (sqlite3_open(filename, &db->handle) == 0) {
            return db;
        }
        free(db);
        return null;
    }
    
    void sqlite_exec(Database* db, string sql) {
        sqlite3_exec(db->handle, sql, null, null, null);
    }
    
    void sqlite_close(Database* db) {
        if (db != null) {
            sqlite3_close(db->handle);
            free(db);
        }
    }
}
```

## Memory Management
All memory is automatically managed. No manual allocation/deallocation needed.

## Example Program
```c
struct Player {
    vec2 position;
    int health;
    string name;
}

void main() {
    // Smart dynamic variable with type inference
    var player = Player {
        position: make_vec2(0.0, 0.0),
        health: 100,
        name: "Hero"
    };
    
    println("Player created: " + player.name);
    
    // Dynamic variable that can change types
    var gameState = "menu";
    println("Game state: " + gameState);
    
    gameState = "playing";  // Still a string
    println("Game state: " + gameState);
    
    // Universal type for truly dynamic values
    any score = 0;
    score = 1500;
    score = "highscore";
    
    for (int i = 0; i < 10; i++) {
        player.position.x += random_float(-1.0, 1.0);
        player.position.y += random_float(-1.0, 1.0);
        println("Position: (" + player.position.x + ", " + player.position.y + ")");
    }
}
```
