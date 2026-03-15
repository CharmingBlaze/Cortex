# Cortex

<img src="assets/cortex%20logo.png" alt="Cortex Logo" width="400">

**The Modern C.** Write C-like code with modern conveniences. No pointers. No manual memory management. No headaches.

---

## What is Cortex?

Cortex is a **systems programming language** that gives you C's power with modern ergonomics. It compiles to C, runs at native speed, and integrates seamlessly with any C library — but feels like a modern language.

```c
// This is Cortex - familiar C syntax, modern features
void main() {
    var message = "Hello, World!";        // Type inference
    var numbers = [1, 2, 3, 4, 5];        // Array literals
    for (var n in numbers) {              // For-each loops
        println("Number: ${n}");          // String interpolation
    }
}
```

## Why Does Cortex Exist?

Because **you shouldn't have to choose between performance and productivity**.

| You Want | C Gives You | Rust Gives You | Cortex Gives You |
|----------|-------------|----------------|------------------|
| C-like syntax | ✓ | ✗ | ✓ |
| Native performance | ✓ | ✓ | ✓ |
| No manual memory management | ✗ | ✓ | ✓ |
| Simple to learn | ✓ | ✗ | ✓ |
| Fast compilation | ✓ | ✗ | ✓ |
| Easy C interop | ✓ | ✗ | ✓ |
| Modern features | ✗ | ✓ | ✓ |

**Cortex is for developers who:**
- Love C's simplicity but hate its footguns
- Want native performance without fighting the borrow checker
- Need to integrate with existing C libraries
- Believe a language can be both powerful *and* pleasant

## What Problems Does Cortex Solve?

### 1. Memory Safety Without Complexity
```c
// C: Manual memory management - easy to leak, double-free, use-after-free
void* buf = malloc(1024);
// ... forgot to free? leak. free twice? crash.

// Cortex: Automatic cleanup with annotations
extern void* my_alloc(int size) cleanup(free);
var buf = my_alloc(1024);  // Automatically freed on scope exit!
```

### 2. Modern Syntax, Zero Learning Curve
```c
// C: Verbose, error-prone
char* s = malloc(100);
sprintf(s, "Hello %s, you have %d messages", name, count);

// Cortex: Clean, intuitive
var s = "Hello ${name}, you have ${count} messages";
```

### 3. Three Concurrency Models, One Language
```c
// Coroutines for game loops
coroutine void animate() { co_yield(); }

// Threads for parallelism  
spawn worker(null);

// Channels for communication
channel_send(ch, &value);
```

### 4. Seamless C Interop
```c
#include <raylib.h>  // That's it - use any C library directly

void main() {
    InitWindow(800, 600, "Game");
    // Full access to C ecosystem
}
```

### 5. Automatic Memory Management
```c
// Annotate extern functions with their cleanup
extern void* my_alloc(int size) cleanup(free);

void main() {
    var buf = my_alloc(1024);  // Automatically freed on scope exit!
    // No free() needed - Cortex handles it
}
```

---

## Quick Start

```bash
# Install (download from releases)
cortex -i hello.cx -run

# Or build from source
git clone https://github.com/CharmingBlaze/Cortex.git
cd Cortex
go build -o cortex.exe ./cmd/cortex
```

**Your first Cortex program:**
```c
void main() {
    println("Hello, World!");
}
```

**With modern features:**
```c
void main() {
    var name = "Cortex";
    var numbers = [1, 2, 3, 4, 5];
    
    for (var n in numbers) {
        println("Number: ${n}");
    }
    
    var result = calculate(10, 20);
    println("Result: ${result}");
}

int calculate(int a, int b) {
    return a + b;
}
```

---

## What You Get

### Modern Syntax, C Performance

```c
// String interpolation
var greeting = "Hello, ${name}!";

// Array literals with bounds checking
var scores = [95, 87, 92, 100];
var first = scores[0];

// Dict literals
var config = { "host": "localhost", "port": 8080 };

// Multiple return values
(int, int) divide(int a, int b) {
    return (a / b, a % b);
}

var (quotient, remainder) = divide(17, 5);
```

### Smart Type System

```c
// Type inference - compiler figures it out
var count = 42;           // int
var price = 19.99;        // double  
var message = "Hello";    // string
var items = [1, 2, 3];    // int[]

// Explicit types when you want them
int count = 42;
string message = "Hello";

// Dynamic typing when you need flexibility
any value = get_value();
```

### Pattern Matching

```c
match (value) {
    case int n:
        println("Got integer: ${n}");
    case string s:
        println("Got string: ${s}");
    default:
        println("Got something else");
}
```

### Lambdas & Closures

```c
var numbers = [1, 2, 3, 4, 5];

// Lambda with capture
var multiplier = 2;
var doubled = map(numbers, [](int x) {
    return x * multiplier;
});

// Event callbacks
gui_button("Click Me", [](event e) {
    println("Button clicked!");
});
```

### Structs with Methods

```c
struct Player {
    string name;
    int health;
    int score;
    
    void take_damage(int amount) {
        health -= amount;  // Implicit self, dot syntax
        if (health < 0) {
            health = 0;
        }
    }
    
    bool is_alive() {
        return health > 0;
    }
}

void main() {
    var player = Player{ name: "Hero", health: 100, score: 0 };
    player.take_damage(20);
    println("Health: ${player.health}");
}
```

### Enums That Work

```c
enum Color {
    Red,
    Green,
    Blue
}

void main() {
    var c = Red;  // No Color:: prefix needed
    
    match (c) {
        Red => println("It's red!"),
        Green => println("It's green!"),
        Blue => println("It's blue!"),
    }
}
```

### Defer for Clean Code

```c
void process_file(string path) {
    var file = open_file(path);
    defer { close_file(file); };  // Runs when function exits
    
    // Do work... if you return early or throw,
    // defer still runs automatically
    var content = read_file(file);
    println(content);
}
```

### Async/Coroutines

```c
void fetch_data(void* arg) {
    println("Fetching...");
    for (int i = 0; i < 3; i++) {
        yield;  // Pause, let other code run
    }
    println("Done!");
}

void main() {
    var task = async_create(fetch_data, null);
    async_await(task);  // Wait for completion
}
```

---

## ⚡ Concurrency That Actually Makes Sense

Cortex gives you **three powerful concurrency models** that work together seamlessly. No more choosing between callbacks, promises, or complex async/await chains.

### 1. Coroutines — Cooperative Multitasking

Perfect for game loops, animations, and state machines:

```c
coroutine void animate_player(Player* p) {
    for (int i = 0; i < 10; i++) {
        p.x += 5;  // Dot syntax everywhere
        yield;     // Simplified yield (no parens needed)
    }
}
```

### 2. Threads — True Parallelism

When you need real CPU parallelism:

```c
void heavy_computation(void* arg) {
    // Runs on separate CPU core
    for (int i = 0; i < 1000000; i++) {
        // ... crunching numbers ...
    }
}

void main() {
    spawn heavy_computation(null);  // Fire and forget
    // Or: spawn t = heavy_computation(null); thread_join(t);
}
```

### 3. Channels — Thread-Safe Communication

Go-style channels for clean thread communication:

```c
void producer(void* arg) {
    cortex_channel ch = (cortex_channel)arg;
    for (int i = 1; i <= 5; i++) {
        channel_send(ch, &i);
    }
    channel_close(ch);
}

void consumer(void* arg) {
    cortex_channel ch = (cortex_channel)arg;
    int value;
    while (channel_recv(ch, &value)) {
        println("Got: ${value}");
    }
}

void main() {
    cortex_channel ch = channel_create(sizeof(int), 10);
    spawn producer(ch);
    spawn consumer(ch);
}
```

### Why This Makes Cortex Special

| Language | Coroutines | Threads | Channels | Simple Syntax |
|----------|------------|---------|----------|---------------|
| C | ❌ | ✓ | ❌ | ✓ |
| C++ | ✓ (complex) | ✓ | ❌ | ❌ |
| Rust | ✓ (async) | ✓ | ✓ | ❌ |
| Go | ❌ | ✓ (goroutines) | ✓ | ✓ |
| **Cortex** | ✓ | ✓ | ✓ | ✓ |

**Cortex is the only language that gives you all three concurrency models with simple, clean syntax.** No callback hell. No complex futures. No lifetime annotations. Just straightforward code that does what you mean.

---

## Native GUI System

Build desktop apps with a clean, simple API. No external dependencies.

```c
void main() {
    gui_window win = gui_window_create("My App", 800, 600);
    gui_window_center(win);
    
    // Create widgets
    gui_widget label = gui_label_create("Hello, Cortex!");
    gui_widget button = gui_button_create("Click Me", [](event e) {
        println("Clicked!");
    });
    
    // Layout
    gui_container vbox = gui_vbox_create();
    gui_container_add(vbox, label);
    gui_container_add(vbox, button);
    
    // Show and run
    gui_window_set_content(win, vbox);
    gui_window_show(win);
    gui_run();
}
```

**Widgets:** Buttons, labels, text entries, checkboxes, sliders, progress bars, images, shapes, and more.

**Layouts:** VBox, HBox, Grid for responsive designs.

**Cross-platform:** Native look on Windows, macOS, Linux.

---

## C Library Integration

Use any C library directly. No wrappers, no bindings.

```c
#include <raylib.h>  // That's it!

void main() {
    InitWindow(800, 600, "Game");
    while (!WindowShouldClose()) {
        BeginDrawing();
        ClearBackground(RAYWHITE);
        DrawText("Hello from Cortex!", 190, 200, 20, LIGHTGRAY);
        EndDrawing();
    }
    CloseWindow();
}
```

**Easy library setup:**
```bash
# Create config for any library
cortex -mkconfig raylib

# Edit configs/raylib.json with your paths, then:
cortex -i game.cx -o game -use raylib
```

---

## Built-in Features

| Feature | Description |
|---------|-------------|
| `println`, `print` | Formatted output |
| `read_file`, `write_file` | File I/O |
| `http_get`, `http_post` | HTTP requests |
| `tcp_connect`, `tcp_listen` | TCP networking |
| `random_int`, `random_float` | Random numbers |
| `time_now`, `time_format` | Time utilities |
| `sha256_hash` | Cryptographic hashing |
| `Vec2`, `Vec3` | 2D/3D vectors |

---

## Feature Flags

Enable only what you need. Smaller binaries, faster compilation.

```json
{
  "features": {
    "qol": true,        // Vectors, random, time
    "blockchain": false, // Crypto features
    "async": true       // Coroutines
  }
}
```

---

## Examples

| Example | Description |
|---------|-------------|
| `hello.cx` | Basic hello world |
| `calculator.cx` | Simple calculator |
| `guess_game.cx` | Number guessing game |
| `drawing_program.cx` | GUI drawing app |
| `async_demo.cx` | Coroutines demo |
| `struct_methods.cx` | Struct methods |
| `match_result.cx` | Pattern matching |

**All 43 examples compile and run.** Check `examples/` directory.

---

## Comparison

| Feature | C | C++ | Rust | Cortex |
|---------|---|-----|------|--------|
| C-like syntax | ✓ | ✓ | ✗ | ✓ |
| No manual memory management | ✗ | ✗ | ✓ | ✓ |
| Simple to learn | ✓ | ✗ | ✗ | ✓ |
| Fast compilation | ✓ | ✗ | ✗ | ✓ |
| C interop | ✓ | ✓ | ✗ | ✓ |
| Modern features | ✗ | ✓ | ✓ | ✓ |
| No complex build system | ✗ | ✗ | ✗ | ✓ |

---

## Documentation

- **[LANGUAGE_GUIDE.md](LANGUAGE_GUIDE.md)** — Learn Cortex from scratch
- **[LANGUAGE_SPEC.md](LANGUAGE_SPEC.md)** — Complete language specification

---

## Build from Source

```bash
# Requirements: Go 1.21+, GCC or TCC
git clone https://github.com/CharmingBlaze/Cortex.git
cd Cortex
go build -o cortex.exe ./cmd/cortex

# Test it works
./cortex -i examples/hello.cx -run
```

---

## Philosophy

**Cortex respects C.** We didn't reinvent the wheel — we made it rounder.

- Same syntax you already know
- Same performance characteristics  
- Same ability to call any C library
- But with modern conveniences that make you productive

C showed us that simplicity and power aren't mutually exclusive. Cortex takes that lesson further.

---

## The Team

Cortex is developed by a team of computer scientists working from an **underground research base in the Himalayas**. Why? Because sometimes you need complete isolation from the noise of the world to build something truly elegant. Plus, the mountain air helps with debugging.

---

## License

MIT License — use it for anything. Commercial projects, open source, education, whatever.

---

## Contributing

Found a bug? Have an idea? Open an issue or PR on [GitHub](https://github.com/CharmingBlaze/Cortex).

---

**Ready to write modern C?**

```bash
cortex -i your_first_program.cx -run
```

*Welcome to Cortex.*
