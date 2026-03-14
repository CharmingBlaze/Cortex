# Cortex Language Guide for Beginners

This guide explains the **Cortex** language from scratch: what it is, how to run programs, and how to use every major feature with examples.

---

## What is Cortex?

**Cortex** is a programming language that:

- **Looks like C** — curly braces `{}`, semicolons `;`, familiar `if`/`for`/`while`
- **Hides pointers** — no `*` or `&` in your code; memory is handled for you
- **Adds modern features** — strings, dynamic variables, vectors, easy I/O, C library support
- **Compiles to C** — your `.cx` file is turned into C, then into an executable

You write **one or more `.cx` files**, run the **Cortex compiler**, and get an executable. No manual memory management, no pointer bugs—just clear, readable code for games and applications.

---

## Your First Program

Every Cortex program has a **`main`** function. Execution starts there.

```c
void main() {
    println("Hello, World!");
}
```

- `void` = this function doesn’t return a value  
- `main` = the entry point  
- `println("...")` = print a line of text and a newline  

**Run it:**

```bash
cortex -i hello.cx -o hello
./hello
```

Or compile and run in one step:

```bash
cortex -i hello.cx -run
```

---

## Variables and Types

### Static types (you choose the type)

Declare a variable by giving its **type** and **name**, then optionally assign a value:

```c
int age = 25;
float price = 9.99;
string name = "Cortex";
bool ok = true;
```

| Type    | Meaning              | Example values      |
|---------|----------------------|---------------------|
| `int`   | Whole numbers        | `0`, `42`, `-7`     |
| `float` | Decimal numbers      | `3.14`, `-0.5`      |
| `double`| More precise decimal | `3.141592653589`    |
| `string`| Text                 | `"hello"`, `""`     |
| `bool`  | True or false        | `true`, `false`     |
| `char`  | Single character     | `'a'`               |

You can declare first and assign later:

```c
int count;
count = 0;
```

### Smart dynamic variable: `var`

With **`var`**, the compiler infers the type from the value. You can even change the type when you reassign:

```c
var x = 42;        // x is int
var name = "Hi";   // name is string
x = 99;            // still int
x = "oops";        // now x holds a string
```

Use `var` when the type is obvious or when you want one variable to hold different kinds of values.

### Universal type: `any`

**`any`** can hold any value. You need **runtime** checks when you use it:

```c
any value = 42;
value = "hello";
value = make_vec2(1.0, 2.0);

if (is_type(value, "int")) {
    int n = as_int(value);
    println("Number: " + n);
}
```

Common helpers: `is_type(value, "int")`, `as_int(value)`, `as_float(value)`, `as_string(value)`.

---

## Input and Output

### Printing

```c
print("no newline ");
println("with newline");

say("same as print ");
show("same as println");
```

- **`println(s)`** / **`show(s)`** — print string and then a newline  
- **`print(s)`** / **`say(s)`** — print string with no newline  

You can build strings with `+`:

```c
int a = 10;
int b = 20;
println("Sum: " + (a + b));   // Sum: 30
```

### String interpolation

Embed expressions in a string with `${ ... }`:

```c
string name = "Cortex";
int version = 1;
show("Hello, ${name}! Version ${version}");
```

### Reading input

```c
string line = input_line();   // reads one line from the user
println("You typed: " + line);
```

---

## Conditionals: `if` and `else`

```c
int age = 17;

if (age >= 18) {
    show("Adult");
} else if (age >= 13) {
    show("Teen");
} else {
    show("Child");
}
```

Conditions are **boolean**: use `==`, `!=`, `<`, `>`, `<=`, `>=`, and combine with `&&` (and), `||` (or), `!` (not).

```c
if (x > 0 && x < 10) { show("single digit positive"); }
if (!done) { show("not done"); }
```

---

## Loops

### `for` (classic C-style)

```c
for (int i = 0; i < 5; i++) {
    show("i = " + i);
}
```

### `while`

```c
int n = 0;
while (n < 3) {
    show("n = " + n);
    n = n + 1;
}
```

### `do`-`while`

Runs at least once, then checks the condition:

```c
int x = 0;
do {
    show("x = " + x);
    x = x + 1;
} while (x < 3);
```

### `repeat` (run a block N times)

No loop variable—just “do this 10 times”:

```c
repeat (5) {
    show("Hello!");
}
```

### `for`-`in` (over an array)

```c
var nums = [10, 20, 30];
for (n in nums) {
    show("" + n);
}
```

Use **`break;`** to exit a loop and **`continue;`** to skip to the next iteration.

---

## Functions

Define a function with **return type**, **name**, **parameters**, and a **body**:

```c
int add(int a, int b) {
    return a + b;
}

void greet(string name) {
    show("Hello, " + name);
}

void main() {
    int sum = add(5, 3);
    show("5 + 3 = " + sum);
    greet("Cortex");
}
```

- **Return type** — `int`, `float`, `string`, `void`, or any type. Use `void` when the function doesn’t return a value.
- **Parameters** — `type name`; multiple parameters separated by commas.

### Multiple return values

Return several values at once:

```c
(int, int) minmax(int a, int b) {
    if (a < b) {
        return (a, b);
    }
    return (b, a);
}

void main() {
    int lo, hi;
    (lo, hi) = minmax(10, 3);
    show("lo = " + lo);   // 3
    show("hi = " + hi);   // 10
}
```

---

## Structs (custom data)

A **struct** is a type that groups named fields:

```c
struct Player {
    int x;
    int y;
    int health;
}

void main() {
    Player p;
    p.x = 100;
    p.y = 200;
    p.health = 3;
    show("Position: " + p.x + ", " + p.y);
}
```

- **Declare** the struct once (usually at top level).
- **Create** a variable: `Player p;`
- **Use** fields with a dot: `p.x`, `p.health`.

### Struct methods

You can define **methods** inside the struct. They use the struct’s fields directly and are called with a dot:

```c
struct Player {
    int x;
    int y;
    void move(int dx, int dy) {
        x = x + dx;
        y = y + dy;
    }
}

void main() {
    Player p;
    p.x = 10;
    p.y = 20;
    p.move(5, -3);
    show("Now at " + p.x + ", " + p.y);   // 15, 17
}
```

---

## Enums (named constants)

**Enums** give names to a set of integer constants:

```c
enum State { Idle, Running, Done }

void main() {
    int state = Idle;
    if (state == Running) {
        show("Running");
    }
    switch (state) {
        case Idle:   { show("idle"); break; }
        case Running: { show("run"); break; }
        default:     { show("done"); }
    }
}
```

You can use the enum name as a prefix: `State.Idle` or just `Idle` where the type is clear.

---

## Arrays

### Array literals

```c
var nums = [1, 2, 3, 4, 5];
show("" + nums[0]);   // first element: 1
nums[2] = 99;        // change third element
```

**2D arrays** — use a literal of literals and index with `arr[i][j]`:

```c
var m = [[1, 2], [3, 4]];
show("" + m[0][1]);   // 2
m[1][0] = 99;
```

### Dynamic array (growing list)

For a list that grows at runtime, use the **array** API:

```c
array a = array_create();
array_push(a, 42);
array_push(a, "hello");
array_push(a, 3.14);

int n = array_len(a);        // 3
any first = array_get(a, 0);
int val = as_int(first);     // 42

array_set(a, 1, make_any_string("world"));
array_free(a);               // free when done
```

- **`array_create()`** — new empty array  
- **`array_push(a, value)`** — add at end  
- **`array_get(a, index)`** — get element (returns `any`)  
- **`array_set(a, index, value)`** — set element  
- **`array_len(a)`** — number of elements  
- **`array_free(a)`** — release memory  

---

## Dictionaries (key–value storage)

**dict** stores string keys and values (as `any`). You can use a **dict literal** or the API:

```c
dict d = { "name": "Cortex", "score": 100 };
// or: dict d = dict_create();
//     dict_set(d, "name", make_any_string("Cortex"));
//     dict_set(d, "score", make_any_int(100));

bool has = dict_has(d, "name");   // true
any val = dict_get(d, "score");
int score = as_int(val);

int size = dict_len(d);
dict_free(d);
```

- **`dict_create()`** — new empty dictionary  
- **`dict_set(d, "key", value)`** — set or overwrite  
- **`dict_get(d, "key")`** — get value (`any`)  
- **`dict_has(d, "key")`** — true if key exists  
- **`dict_len(d)`** — number of entries  
- **`dict_free(d)`** — release memory  

Use `make_any_int`, `make_any_string`, `make_any_float`, etc. when you need to pass a typed value as `any`.

---

## Error handling with `result`

Instead of crashing, functions can return a **result**: either a value or an error message.

```c
result r = parse_number("42");
if (result_is_ok(r)) {
    any v = result_value(r);
    show("value: " + as_int(v));
} else {
    show("error: " + result_error(r));
}
```

Create results yourself:

```c
result ok = result_ok(make_any_int(42));
result err = result_err("something went wrong");
```

### Match on result

Unpack success or error in one place:

```c
result r = do_something();
match (r) {
    case Ok(v): { show("value: " + as_int(v)); }
    case Err(e): { show("error: " + e); }
}
```

Use `as_int(v)`, `as_string(v)`, etc. on `v` depending on what the function returns.

---

## Vectors (2D and 3D)

For positions, movement, and simple physics:

```c
vec2 pos = make_vec2(100.0, 200.0);
vec2 vel = make_vec2(5.0, -2.0);

pos.x = pos.x + vel.x;
pos.y = pos.y + vel.y;

float len = vec2_length(vel);
vec2 u = normalize(vel);
vec2 sum = vec2_add(pos, vel);
vec2 diff = vec2_sub(pos, vel);
float d = vec2_distance(pos, vel);
```

**vec3** has `.x`, `.y`, `.z` and `make_vec3`, `vec3_add`, `vec3_length`, etc.

---

## Random and time

```c
int n = random_int(1, 10);      // random int from 1 to 10
float f = random_float(0.0, 1.0);

float t = get_time();          // seconds since program start
sleep(1.5);                    // pause 1.5 seconds
wait(0.5);                     // same as sleep
```

---

## Lambdas (anonymous functions)

Use a lambda when you need a small function as a value (e.g. a callback). No captures yet—only parameters and return value:

```c
var add = [](int a, int b) -> int { return a + b; };
// use add(2, 3) if you have a way to call it (e.g. passed to a C API)
```

Typical use: pass to a function that expects a callback (e.g. event handlers, UI).

---

## Unit tests

Define tests and run them from `main`:

```c
test "addition" {
    assert_eq(1 + 1, 2);
}

test "approx" {
    assert_approx(0.1 + 0.2, 0.3, 0.001);
}

void main() {
    test_run_all();
    show("Tests done.");
}
```

- **`test "name" { ... }`** — define a test  
- **`assert_eq(a, b)`** — fail if `a != b`  
- **`assert_approx(a, b, epsilon)`** — for floats  
- **`test_run_all()`** — run all registered tests  

---

## Using C libraries

Cortex can call C code. Use the same **`#include`** as in C; the compiler infers linking from the header name (e.g. `#include <raylib.h>` → link `raylib`).

```c
#include <raylib.h>

void main() {
    InitWindow(800, 450, "My Game");
    SetTargetFPS(60);
    while (!WindowShouldClose()) {
        BeginDrawing();
        ClearBackground(RAYWHITE);
        DrawText("Hello from Cortex!", 190, 200, 20, DARKGRAY);
        EndDrawing();
    }
    CloseWindow();
}
```

Build with include and library paths (e.g. via config):

```bash
cortex -i game.cx -o game -use raylib
```

So: **include the header**, write normal Cortex (and C API) calls, then pass paths/config so the compiler and linker can find the library. No `#pragma` needed for the library name.

---

## Preprocessor and config

- **`#include <file.h>`** or **`#include "file.h"`** — include a C header (e.g. for a library).  
- **`#pragma link("libname")`** — explicitly ask to link a library (optional if you use `#include <libname.h>`).  
- **`#use "libname"`** — shorthand for including `<libname.h>` and linking `libname`.  

The compiler uses **config** (e.g. `-config configs/raylib.json` or `-use raylib`) for include paths, library paths, and feature flags.

---

## Defer (run code when leaving a block)

**`defer`** runs the block when the current block exits (normal exit or return):

```c
void main() {
    show("start");
    defer { show("defer runs"); }
    show("middle");
}   // "defer runs" prints here
```

Use it for cleanup (e.g. closing files, freeing resources).

---

## Quick reference

| Topic           | Syntax / API |
|----------------|--------------|
| Entry point    | `void main() { }` |
| Print line     | `println("text");` or `show("text");` |
| Variables      | `int x = 5;` or `var x = 5;` |
| Conditionals   | `if (cond) { } else if (cond) { } else { }` |
| Loops          | `for (int i=0; i<n; i++)`, `while (cond)`, `repeat (n)`, `for (x in arr)` |
| Functions      | `int f(int a, int b) { return a+b; }` |
| Struct         | `struct T { int x; }` then `T t; t.x = 1;` |
| Struct method  | `void move(int dx) { x = x + dx; }` → `t.move(5);` |
| Enum           | `enum E { A, B }` then `int e = A;` |
| Array literal  | `var a = [1, 2, 3];` then `a[0]` |
| 2D array       | `var m = [[1,2],[3,4]];` then `m[i][j]` |
| Dynamic array  | `array_create`, `array_push`, `array_get`, `array_len`, `array_free` |
| Dict literal   | `dict d = { "k": v };` |
| Dictionary     | `dict_create`, `dict_set`, `dict_get`, `dict_has`, `dict_free` |
| Result         | `result_ok(val)`, `result_err("msg")`, `result_is_ok(r)`, `result_value(r)` |
| Match result   | `match (r) { case Ok(v): ... case Err(e): ... }` |
| Vectors        | `make_vec2(x,y)`, `vec2_add`, `vec2_length`, `normalize` |
| Random         | `random_int(min,max)`, `random_float(min,max)` |
| Time           | `get_time()`, `sleep(sec)`, `wait(sec)` |
| Tests          | `test "name" { assert_eq(a,b); }` then `test_run_all();` |
| C library      | `#include <lib.h>` and build with `-use lib` or config |

---

## Where to go next

- **Examples** — See the `examples/` folder: `hello.cx`, `guess_game.cx`, `struct_methods.cx`, `app_file.cx`, and the raylib examples under `examples/raylib/`.
- **Full reference** — See the main [README](README.md) and [LANGUAGE_SPEC.md](LANGUAGE_SPEC.md) for the complete language and compiler options.
- **Building** — `go build -o cortex .` then `cortex -i file.cx -o program` or `cortex -i file.cx -run`.

You now have a full, beginner-oriented picture of Cortex: variables, control flow, functions, structs, enums, arrays, dicts, results, vectors, tests, and C libraries, all with examples.
