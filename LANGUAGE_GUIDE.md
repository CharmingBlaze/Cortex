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

## Running Cortex Programs

### Simple Commands (Recommended)

```bash
# Create a new project
cortex new my_game
cd my_game
cortex run

# Run a single file
cortex run hello.cx

# Build to executable
cortex build game.cx -o game.exe

# Generate bindings from C library
cortex bind raylib -i third_party/raylib/src/raylib.h
```

### Project Configuration (cortex.toml)

For projects with dependencies, create a `cortex.toml`:

```toml
[project]
name = "my_game"
version = "0.1.0"
entry = "main.cx"

[dependencies.raylib]
include_path = "third_party/raylib/src"
lib_path = "third_party/raylib/build/raylib"
libs = ["raylib", "opengl32", "gdi32", "winmm", "shell32"]
```

Then just run:
```bash
cortex run
```

### Legacy Commands

```bash
# Compile and run in one step
cortex -i hello.cx -run

# Compile to executable
cortex -i hello.cx -o hello
./hello
```

---

## Your First Program

Every Cortex program has a **`main`** function. Execution starts there.

```c
void main() {
    println("Hello, World!");
}
```

- `void` = this function doesn't return a value  
- `main` = the entry point  
- `println("...")` = print a line of text and a newline  

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

### Constants with `const`

**`const`** declares immutable values that cannot be changed:

```c
const int MAX_SIZE = 100;
const string APP_NAME = "Cortex";
const float PI = 3.14159;

MAX_SIZE = 200;  // Error: cannot reassign const
```

Use `const` for:
- Configuration values
- Magic numbers with meaningful names
- Values that should never change

### C Library Functions (Auto-Extern)

**Cortex automatically generates extern declarations** for C functions when you include headers. No manual `extern` needed!

```c
#include <stdio.h>
#include <stdlib.h>

void main() {
    // Just call C functions directly - extern is auto-generated!
    var buf = malloc(1024);
    printf("Buffer allocated\n");
    free(buf);
}
```

When Cortex sees `#include`, it automatically:
1. Detects undefined function calls
2. Infers parameter types from your arguments
3. Generates the extern declaration in the output C code

### Manual Extern Declarations

Use manual `extern` when you need:
- **Cleanup annotations** for automatic memory management
- **Specific return types** (default is `int`)
- **Pointer return types** for type safety

```c
// Manual extern with cleanup annotation
extern void* my_alloc(int size) cleanup(free);
var buf = my_alloc(1024);  // Auto-freed on scope exit!

// Manual extern for pointer return type
extern char* strdup(string s);  // Returns char*, not int
```

### Extern declarations

**`extern`** declares functions from C libraries (optional with auto-extern):

```c
// Declare C functions you want to call (optional if header included)
extern void* malloc(int size);
extern void free(void* ptr);
extern int printf(string format, ...);

void main() {
    var buf = malloc(1024);
    printf("Buffer allocated at: %p\n", buf);
    free(buf);
}
```

Combine with `cleanup` for automatic memory management:

```c
extern void* my_alloc(int size) cleanup(free);
var buf = my_alloc(1024);  // Auto-freed on scope exit!
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

### Binding immutability: `const`

**`const`** declares a variable that cannot be reassigned after initialization. This is **binding immutability** (like TypeScript/Swift), not deep immutability:

```c
const x = 10;           // Type inferred as int
const int y = 20;       // Explicit type
const var name = "Cortex";  // var type, still const binding

x = 30;  // ERROR: cannot assign to const 'x'
```

**What const prevents:**
- Reassigning the variable to a new value

**What const allows:**
- Modifying contents of arrays or dicts (the binding is const, not the data)

```c
const arr = [1, 2, 3];
arr[0] = 99;        // Allowed - modifying array contents
arr.push(4);        // Allowed - mutating the array
arr = [4, 5, 6];    // ERROR - cannot reassign const binding

const config = { "host": "localhost" };
config["port"] = 8080;  // Allowed - adding to dict
config = {};            // ERROR - cannot reassign
```

**When to use const:**
- Configuration values that shouldn't change
- Constants like math values, limits, thresholds
- Preventing accidental reassignment in long functions
- Making intent clear to other developers

---

## Type System

Cortex uses a **hybrid type system**: static types for safety, dynamic types for flexibility.

### Static typing

When you declare a type explicitly, the compiler checks types at compile time:

```c
int x = 10;
x = "hello";  // ERROR: type mismatch
```

Benefits:
- Catch errors before running
- Better IDE support and autocomplete
- Clearer code intent

### Dynamic typing with `var`

`var` infers the type from the initializer:

```c
var x = 10;       // inferred as int
var y = "hello";  // inferred as string
var z = [1, 2];   // inferred as int[]
```

`var` variables can change type at runtime:

```c
var x = 10;
x = "now a string";  // OK at runtime
```

### The `any` type

`any` is explicitly dynamic. Use it when:
- Storing values of different types in a collection
- Working with external data (JSON, user input)
- Gradual typing (migrating from dynamic to static)

```c
array a = array_create();
array_push(a, 42);       // int
array_push(a, "hello");  // string
array_push(a, 3.14);     // float

any val = array_get(a, 0);
if (is_type(val, "int")) {
    int n = as_int(val);
}
```

### Type inference rules

| Expression | Inferred type |
|------------|---------------|
| `var x = 42` | `int` |
| `var x = 3.14` | `double` |
| `var x = "hi"` | `string` |
| `var x = true` | `bool` |
| `var x = [1, 2, 3]` | `int[]` |
| `var x = {"a": 1}` | `dict` |
| `var x = make_vec2(1, 2)` | `vec2` |

### Type checking

The compiler checks:
- Assignments: `int x = "string"` → error
- Function arguments: `void f(int x); f("hi")` → error
- Return values: `int f() { return "hi"; }` → error

But `var` and `any` bypass compile-time checks:

```c
var x = 10;
x = "string";  // OK (runtime)

any y = 10;
y = "string";  // OK (runtime)
```

### When to use each

| Use | When |
|-----|------|
| Explicit type (`int`, `string`) | Known, fixed type; want compile-time safety |
| `var` | Type obvious from context; want brevity |
| `any` | Multiple possible types; external data |
| `const` | Value shouldn't change |

---

## Memory Model

Cortex handles memory automatically so you don't need to manage it manually.

### How memory works

**Stack allocation:**
- Local variables are allocated on the stack
- Automatically freed when function returns
- Fast and deterministic

```c
void example() {
    int x = 10;      // Stack allocated
    string s = "hi"; // Stack allocated
}  // x and s freed automatically
```

**Heap allocation:**
- Dynamic collections (`array`, `dict`) are heap-allocated
- Managed by Cortex runtime
- Freed automatically when no longer reachable

```c
void example() {
    array a = array_create();  // Heap allocated
    array_push(a, 42);
    // No need to free - handled by runtime
}
```

### No pointers

Cortex hides pointers from your code:

| C | Cortex |
|---|--------|
| `int* p = &x;` | Not needed |
| `malloc(sizeof(int))` | Automatic |
| `free(p)` | Automatic |
| `p->field` | `obj.field` |

You can still work with C pointers via `extern` declarations when needed.

### When cleanup happens

1. **Scope exit**: Local variables freed when function returns
2. **Defer**: Run cleanup code explicitly with `defer`
3. **End of program**: All memory freed on exit

```c
void process() {
    var file = open("data.txt");
    defer { close(file); };  // Explicit cleanup
    
    var data = load(file);   // Automatic cleanup when done
}
```

### Working with C memory

When calling C functions via `extern`:

- C memory is **not** managed by Cortex
- Use `defer` to call C cleanup functions
- Be careful with ownership: who frees?

```c
extern void* malloc(int size);
extern void free(void* ptr);

void example() {
    void* buf = malloc(1024);
    defer { free(buf); };  // Ensure cleanup
    // Use buf...
}
```

### Memory safety

Cortex prevents common memory bugs:

| Bug | C | Cortex |
|-----|---|--------|
| Use after free | Possible | Prevented |
| Double free | Possible | Prevented |
| Buffer overflow | Possible | Bounds checked |
| Null pointer deref | Possible | Prevented |
| Dangling pointer | Possible | Prevented |

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

### TYPE...ENDTYPE (BASIC-style user-defined types)

Cortex also supports a more readable TYPE block syntax for defining custom types:

```c
TYPE AccountEntry
    Number AS int
    Name AS string
    Amount AS float
ENDTYPE

void main() {
    AccountEntry account;
    account.Number = 12345;
    account.Name = "Lee";
    account.Amount = 0.42;
    show("Account: " + account.Name + " has $" + account.Amount);
}
```

**Rules:**
- Fields without `AS` default to `int`
- `AS` specifies the type: `FieldName AS Type`
- Types can be built-in (`int`, `float`, `string`) or other user-defined types

### Nested Types

You can nest types within types for complex data structures:

```c
TYPE Amounts
    CurrentBalance AS float
    SavingsBalance AS float
    CreditCardBalance AS float
ENDTYPE

TYPE AccountEntry
    Number AS int
    Name AS string
    Amount AS Amounts
ENDTYPE

void main() {
    AccountEntry account;
    account.Number = 12345;
    account.Name = "Lee";
    account.Amount.CurrentBalance = 0.42;
    account.Amount.SavingsBalance = 100.0;
    account.Amount.CreditCardBalance = -5000.0;
}
```

### Variable Declaration with AS

You can also declare variables using the `AS` syntax:

```c
void main() {
    account AS AccountEntry;
    numbers AS int[100];  // Array of 100 integers
    account.Name = "Test";
}
```

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

For a list that grows at runtime, use the **array** type with method syntax:

```c
array a = array_create();
a.push(42);
a.push("hello");
a.push(3.14);

int n = a.len();           // 3
any first = a.get(0);
int val = as_int(first);   // 42

a.set(1, make_any_string("world"));
a.pop();                   // remove last element
a.free();                  // free when done
```

**Array Methods:**
- **`a.push(value)`** — add element at end
- **`a.pop()`** — remove and return last element
- **`a.len()`** — number of elements
- **`a.get(index)`** — get element (returns `any`)
- **`a.set(index, value)`** — set element at index
- **`a.insert(index, value)`** — insert at index
- **`a.remove(index)`** — remove element at index
- **`a.capacity()`** — current capacity
- **`a.reserve(size)`** — pre-allocate capacity
- **`a.free()`** — release memory

**Legacy API (still supported):**
- `array_push(a, value)`, `array_get(a, index)`, `array_set(a, index, value)`, etc.  

---

## Dictionaries (key–value storage)

**dict** stores string keys and values (as `any`). You can use a **dict literal** or method syntax:

```c
dict d = { "name": "Cortex", "score": 100 };
// or: dict d = dict_create();
//     d.set("name", make_any_string("Cortex"));
//     d.set("score", make_any_int(100));

bool has = d.has("name");   // true
any val = d.get("score");
int score = as_int(val);

int size = d.len();
d.remove("score");
d.free();
```

**Dict Methods:**
- **`d.set("key", value)`** — set or overwrite a key
- **`d.get("key")`** — get value (returns `any`)
- **`d.has("key")`** — true if key exists
- **`d.remove("key")`** — remove a key
- **`d.len()`** — number of entries
- **`d.keys()`** — get array of keys
- **`d.values()`** — get array of values
- **`d.free()`** — release memory

**Legacy API (still supported):**
- `dict_set(d, "key", value)`, `dict_get(d, "key")`, `dict_has(d, "key")`, etc.

Use `make_any_int`, `make_any_string`, `make_any_float`, etc. when you need to pass a typed value as `any`.

---

## String Methods

Strings have built-in methods for common operations:

```c
string s = "  Hello World  ";

// Basic operations
int length = s.len();              // 14
string trimmed = s.trim();         // "Hello World"
string upper = s.upper();          // "  HELLO WORLD  "
string lower = s.lower();          // "  hello world  "

// Search and check
bool found = s.contains("World");  // true
bool starts = s.starts_with("  H"); // true
bool ends = s.ends_with("  ");     // true

// Transform
string replaced = s.replace("World", "Cortex");  // "  Hello Cortex  "
array parts = s.split(" ");         // Split by space
```

**String Methods:**
- **`s.len()`** — string length
- **`s.trim()`** — remove leading/trailing whitespace
- **`s.upper()`** — convert to uppercase
- **`s.lower()`** — convert to lowercase
- **`s.contains(sub)`** — check if substring exists
- **`s.starts_with(prefix)`** — check prefix
- **`s.ends_with(suffix)`** — check suffix
- **`s.replace(old, new)`** — replace all occurrences
- **`s.split(delim)`** — split into array

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

## Null Coalescing Operator (`??`)

The **`??`** operator provides a default value when a nullable expression is null:

```c
string? name = null;
string result = name ?? "default";  // result = "default"

string? value = "hello";
string result2 = value ?? "default";  // result2 = "hello"
```

This is useful for providing fallback values without explicit null checks.

---

## Optional Chaining (`?.`)

The **`?.`** operator safely accesses members and methods on nullable objects:

```c
struct Person {
    string name;
    int age;
}

Person? person = get_person();
string? name = person?.name;  // Safe access - returns null if person is null
int? age = person?.age;       // Also safe
```

If `person` is null, the expression returns null instead of crashing. This chains safely:

```c
string? city = person?.address?.city;  // Safe nested access
```

---

## Range Operators

Cortex supports range expressions for creating numeric ranges:

### Inclusive Range (`..`)

Creates a range from start to end (inclusive):

```c
// For loop with range
for (int i in 0..5) {
    print(i);  // Prints 0, 1, 2, 3, 4, 5
}

// Range expression
var r = 1..10;  // Range from 1 to 10 inclusive
```

### Exclusive Range (`..<`)

Creates a range from start to end (exclusive):

```c
for (int i in 0..<5) {
    print(i);  // Prints 0, 1, 2, 3, 4
}

var r = 1..<10;  // Range from 1 to 9
```

Ranges are useful in for loops and array slicing operations.

---

## Exception Handling with `try`/`catch`/`throw`

Cortex supports structured exception handling:

### Basic try/catch

```c
try {
    risky_operation();
} catch (Exception e) {
    show("Caught: " + e.message);
}
```

### Multiple catch clauses

```c
try {
    process_file(path);
} catch (FileNotFound e) {
    show("File not found: " + path);
} catch (PermissionDenied e) {
    show("Permission denied");
} catch (Exception e) {
    show("Unknown error: " + e.message);
}
```

### Catch-all

```c
try {
    do_something();
} catch {
    show("An error occurred");
}
```

### Finally block

```c
try {
    var file = open("data.txt");
    process(file);
} catch (Exception e) {
    show("Error: " + e.message);
} finally {
    close(file);  // Always runs
}
```

### Throwing exceptions

```c
fn validate(int age) {
    if (age < 0) {
        throw Exception("Age cannot be negative");
    }
    if (age > 150) {
        throw Exception("Age seems unrealistic");
    }
}
```

---

## Generic Types

Cortex supports generic types for type-safe containers:

### `vector<T>` - Typed Dynamic Arrays

A `vector<T>` is a dynamically growing array with compile-time type safety:

```c
vector<int> nums = vector_create_int();
nums.push(10);
nums.push(20);
nums.push(30);

int first = nums.get(0);    // 10
int len = nums.len();       // 3
int last = nums.pop();      // 30

nums.free();                // Release memory
```

**Vector Methods:**
- **`v.push(value)`** — add element at end
- **`v.pop()`** — remove and return last element
- **`v.get(index)`** — get element at index
- **`v.set(index, value)`** — set element at index
- **`v.len()`** — number of elements
- **`v.free()`** — release memory

**Supported Types:** `vector<int>`, `vector<float>`, `vector<double>`, `vector<string>`, `vector<bool>`, `vector<vec2>`, `vector<vec3>`

### `optional<T>` - Explicit Optional Values

An `optional<T>` explicitly represents a value that may or may not exist:

```c
optional<int> find_user(int id) {
    if (id > 0) {
        return optional_some_int(id * 100);
    }
    return optional_none_int();
}

optional<int> result = find_user(5);
if (result.has_value) {
    show("Found: " + result.value);
}
```

**Note:** The shorthand `T?` syntax (e.g., `int?`) is preferred for most cases.

---

## Modules and Imports

Cortex supports a minimal module system for organizing code:

### Declaring a Module

```c
module "math";

fn add(int a, int b) -> int {
    return a + b;
}

fn multiply(int a, int b) -> int {
    return a * b;
}
```

### Using Module Functions

```c
// In another file
import "math.cx";

void main() {
    int sum = math.add(5, 3);        // 8
    int product = math.multiply(4, 2);  // 8
}
```

Functions declared in a module are prefixed with the module name when called from other files.

---

## Async/Await

Cortex supports asynchronous programming with coroutines:

### Async Functions

```c
async fn fetch_data(string url) -> string {
    // Simulate async operation
    yield;
    return "data from " + url;
}

void main() {
    async task = fetch_data("https://example.com");
    // Do other work...
    string result = await task;
    show(result);
}
```

### Yielding

The `yield` keyword pauses an async function, allowing other tasks to run:

```c
async fn process_items(array items) {
    for (int i = 0; i < items.len(); i++) {
        process(items.get(i));
        yield;  // Allow other tasks to run
    }
}
```

### Running Multiple Tasks

```c
void main() {
    async t1 = task1();
    async t2 = task2();
    async t3 = task3();
    
    // Run all tasks to completion
    async_run_all();
}
```

---

## Channels for Message Passing

Channels provide thread-safe communication between concurrent tasks:

### Creating and Using Channels

```c
channel<int> ch = channel<int>(10);

// Send values
ch.send(42);
ch.send(100);

// Receive values
int val = ch.recv();
show("Received: " + val);  // 42

ch.close();  // Close the channel
```

### Channel Methods

- **`ch.send(value)`** — send a value (blocks if channel is full)
- **`ch.recv()`** — receive a value (blocks if channel is empty)
- **`ch.try_send(value)`** — non-blocking send, returns 1 on success, 0 if full, -1 if closed
- **`ch.try_recv()`** — non-blocking receive
- **`ch.close()`** — close the channel
- **`ch.is_closed()`** — check if channel is closed
- **`ch.free()`** — release channel memory

### Producer-Consumer Pattern

```c
channel<string> jobs = channel<string>(100);

// Producer
fn produce() {
    for (int i = 0; i < 10; i++) {
        jobs.send("job_" + i);
    }
    jobs.close();
}

// Consumer
fn consume() {
    string job;
    while (jobs.try_recv(&job) > 0) {
        show("Processing: " + job);
    }
}
```

---

## Pattern Matching

**`match`** is a powerful way to inspect and destructure values. It's like a `switch` on steroids:

### Type matching

Check the runtime type of an `any` value:

```c
any value = 42;

match (value) {
    case int n: {
        show("Got integer: " + n);
    }
    case string s: {
        show("Got string: " + s);
    }
    case float f: {
        show("Got float: " + f);
    }
    default: {
        show("Got something else");
    }
}
```

The variable after the type (`n`, `s`, `f`) is bound to the value with that type.

### Matching results

Results have two cases: `Ok` for success, `Err` for failure:

```c
result r = parse_number("42");

match (r) {
    case Ok(v): {
        int n = as_int(v);
        show("Parsed: " + n);
    }
    case Err(e): {
        show("Error: " + e);
    }
}
```

### Matching with conditions

Add guards with `if`:

```c
match (value) {
    case int n if n > 0: {
        show("Positive: " + n);
    }
    case int n: {
        show("Non-positive: " + n);
    }
}
```

### Fallthrough

Cases don't fall through by default (unlike C `switch`). Each case needs its own body.

---

## Defer (cleanup on scope exit)

**`defer`** runs code when the current block exits, whether by normal flow or early return:

```c
void process_file(string path) {
    var file = open_file(path);
    defer { close_file(file); };  // Guaranteed to run on exit
    
    // Work with file...
    if (error) {
        return;  // defer still runs!
    }
    // More work...
}  // defer runs here too
```

**Key behaviors:**
- Deferred code runs in **reverse order** of declaration (LIFO)
- Runs on **any** exit: normal end, `return`, or even panic
- Perfect for cleanup: closing files, freeing resources, releasing locks

```c
void main() {
    show("start");
    defer { show("first defer"); }
    defer { show("second defer"); }
    show("middle");
}
// Output: start, middle, second defer, first defer
```

**Defer vs try/finally:**
- Defer is simpler and more flexible
- Can have multiple defers in one block
- No need for try block structure

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

## Async and Coroutines

Cortex supports **cooperative multitasking** with coroutines. Functions can pause execution and resume later, allowing multiple tasks to run concurrently without threads.

### Creating coroutines

Use `coroutine` keyword to define a coroutine function:

```c
coroutine void fetch_data(void* arg) {
    show("Starting fetch...");
    for (int i = 0; i < 3; i++) {
        sleep(0.5);  // Simulate work
        co_yield();   // Pause, let other code run
    }
    show("Fetch complete!");
}
```

**Key points:**
- `coroutine` marks a function that can pause
- `co_yield()` pauses execution and returns control to caller
- When resumed, execution continues after `co_yield()`

### Running coroutines

```c
void main() {
    // Create coroutine
    var task = async_create(fetch_data, null);
    
    // Run until complete
    while (async_is_running(task)) {
        async_resume(task);
        // Can do other work here
    }
    
    // Or wait for completion
    async_await(task);
    show("Task finished!");
}
```

### Coroutine API

| Function | Description |
|----------|-------------|
| `async_create(fn, arg)` | Create a new coroutine from a coroutine function |
| `async_resume(co)` | Resume a paused coroutine |
| `async_await(co)` | Block until coroutine completes |
| `async_is_running(co)` | Check if coroutine is still running |
| `co_yield()` | Pause current coroutine (inside coroutine only) |

### When to use coroutines

- **I/O operations**: Fetch data, read files without blocking
- **Game loops**: Spread work across frames
- **Animations**: Pause and resume over time
- **State machines**: Natural pause/resume points

```c
coroutine void animate_player(Player* p) {
    for (int i = 0; i < 10; i++) {
        p->x += 5;
        co_yield();  // Move a bit each frame
    }
}
```

**Note**: Coroutines are cooperative - they only yield at `co_yield()`. Long-running code without yields will block other coroutines.

---

## Threading and Channels

Cortex provides **true parallelism** with threads and thread-safe channels for communication.

### Spawning Threads

Use `spawn` to run a function in a new thread:

```c
void worker(void* arg) {
    show("Worker running in parallel!");
    thread_sleep_ms(1000);
    show("Worker done!");
}

void main() {
    show("Main starting...");
    spawn worker(null);       // Fire and forget
    
    // Or capture thread handle
    spawn t = worker(null);   // t is cortex_thread
    thread_join(t);           // Wait for completion
    show("Worker finished");
}
```

### Thread API

| Function | Description |
|----------|-------------|
| `spawn fn(args)` | Run function in new thread (fire and forget) |
| `spawn var = fn(args)` | Run and capture thread handle |
| `thread_join(t)` | Wait for thread to complete |
| `thread_is_running(t)` | Check if thread is still running |
| `thread_id()` | Get current thread ID |
| `thread_sleep_ms(ms)` | Sleep for milliseconds |

### Channels

Channels provide thread-safe communication between threads:

```c
void producer(void* arg) {
    cortex_channel ch = (cortex_channel)arg;
    for (int i = 1; i <= 5; i++) {
        channel_send(ch, &i);  // Send value
        show("Sent: " + to_string(i));
    }
    channel_close(ch);
}

void consumer(void* arg) {
    cortex_channel ch = (cortex_channel)arg;
    int value;
    while (channel_recv(ch, &value)) {
        show("Received: " + to_string(value));
    }
    show("Channel closed");
}

void main() {
    // Create channel for int values, capacity 10
    cortex_channel ch = channel_create(sizeof(int), 10);
    
    spawn producer(ch);
    spawn consumer(ch);
    
    thread_sleep_ms(100);  // Let threads finish
    channel_free(ch);
}
```

### Channel API

| Function | Description |
|----------|-------------|
| `channel_create(elem_size, capacity)` | Create a new channel |
| `channel_send(ch, &value)` | Send value (blocks if full) |
| `channel_recv(ch, &out)` | Receive value (blocks if empty) |
| `channel_try_send(ch, &value)` | Non-blocking send (returns 1=sent, 0=would block, -1=closed) |
| `channel_try_recv(ch, &out)` | Non-blocking receive (returns 1=received, 0=would block, -1=closed) |
| `channel_close(ch)` | Close channel (no more sends) |
| `channel_is_closed(ch)` | Check if channel is closed |
| `channel_free(ch)` | Free channel resources |

### When to Use

- **Coroutines** (`coroutine`/`co_yield`): Cooperative multitasking, game loops, state machines
- **Threads** (`spawn`): True parallelism, CPU-intensive work, blocking I/O
- **Channels**: Thread communication, producer-consumer patterns

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

## Automatic Memory Management with Cleanup

Cortex provides **automatic memory management** for extern functions using the `cleanup` annotation. No manual `free()` calls needed!

### The Problem

In C, you must manually free memory:

```c
void* buf = malloc(1024);
// ... use buf ...
free(buf);  // Don't forget!
```

Forget to free? Memory leak. Free twice? Crash. Free too early? Undefined behavior.

### The Cortex Solution

Annotate extern functions with their cleanup function:

```c
extern void* my_alloc(int size) cleanup(free);
extern void free(void* ptr);

void main() {
    var buf = my_alloc(1024);  // Automatically freed on scope exit!
    show("Using buffer...");
    // No free() needed - Cortex handles it
}
```

### How It Works

When you declare `extern void* my_alloc(int size) cleanup(free)`:

1. Cortex wraps the returned pointer in a **managed handle**
2. Uses GCC's `__attribute__((cleanup))` for automatic cleanup
3. When the variable goes out of scope, `free()` is called automatically

### Safe and Simple

```c
extern FILE* fopen(string path, string mode) cleanup(fclose);
extern void fclose(FILE* f);

void main() {
    var file = fopen("data.txt", "r");  // Auto-closed!
    // ... read file ...
    // fclose called automatically
}

extern void* sqlite3_open(string path) cleanup(sqlite3_close);
var db = sqlite3_open("my.db");  // Auto-closed!
```

### When to Use

| Function Type | Example | Cleanup |
|---------------|---------|---------|
| Memory allocation | `malloc`, `calloc` | `free` |
| File handles | `fopen` | `fclose` |
| Database connections | `sqlite3_open` | `sqlite3_close` |
| Network sockets | `socket` | `close` |
| Custom resources | Any allocator | Any cleanup function |

### Benefits

- **No memory leaks** - Cleanup is guaranteed
- **No use-after-free** - Pointer is nullified after cleanup
- **No double-free** - Cleanup runs exactly once
- **Simple** - Just add `cleanup(func)` to extern declaration

---

## Modules and Imports

Cortex supports organizing code into modules that can be imported:

### Import syntax

```c
// Import a single file
import "utils";

// Import from a subdirectory
import "math/vector";

// Import a module directory (loads mod.cx)
import "graphics";
```

### Module structure

```
project/
├── main.cx          # Entry point
├── utils.cx         # Imported as import "utils"
├── math/
│   └── vector.cx    # Imported as import "math/vector"
└── graphics/
    └── mod.cx       # Imported as import "graphics"
```

### What gets shared

All top-level declarations (functions, structs, enums, consts) are automatically available to importing files:

```c
// utils.cx
const int MAX_ITEMS = 100;

struct Item {
    string name;
    int value;
};

void process_item(Item* i) {
    println("Processing: ${i->name}");
}

// main.cx
import "utils";

void main() {
    var item = Item{ name: "Test", value: 42 };
    process_item(&item);
    println("Max: ${MAX_ITEMS}");
}
```

### Compiling with imports

```bash
# Cortex automatically resolves and compiles imported files
cortex -i main.cx -o myapp
```

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

### Declaring C functions with `extern`

When you need to call C functions that aren't in a standard header, use `extern`:

```c
// Declare a C function
extern int my_c_function(int a, int b);

// Declare with pointer types
extern void* malloc(int size);
extern void free(void* ptr);

void main() {
    int result = my_c_function(10, 20);
    show("Result: " + result);
}
```

**Type mapping between Cortex and C:**

| Cortex type | C type |
|-------------|--------|
| `int` | `int` |
| `float` | `float` |
| `double` | `double` |
| `string` | `char*` |
| `bool` | `int` (0/1) |
| `any` | `AnyValue` struct |
| `void` | `void` |

**Pointer types in extern:**
- Use `void*` for generic pointers
- Use `char*` for C strings
- Use `Type*` for typed pointers (e.g., `int*`)
- Cortex handles the conversion automatically

### Embedding raw C code

For code that can't be expressed in Cortex, embed raw C:

```c
@c int global_counter = 0;

@c #define MAX_SIZE 100

void main() {
    @c printf("Direct C: %d\n", global_counter);
}
```

The `@c` prefix passes the rest of the line directly to the generated C code.

### Linking libraries

Three ways to link:

1. **Auto-link from include**: `#include <raylib.h>` → links `raylib`
2. **Explicit pragma**: `#pragma link("mylib")`
3. **Shorthand**: `#use "raylib"` → includes and links

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

## Standard Library Overview

Cortex includes a standard library of built-in functions organized by category.

### I/O Functions

| Function | Description |
|----------|-------------|
| `print(s)` | Print without newline |
| `println(s)` | Print with newline |
| `show(s)` | Same as `println` |
| `say(s)` | Same as `print` |
| `input_line()` | Read a line from stdin |

### String Functions

| Function | Description |
|----------|-------------|
| `str_concat(a, b)` | Concatenate strings |
| `str_length(s)` | Get string length |
| `str_sub(s, start, len)` | Substring |
| `str_find(s, substr)` | Find substring position |
| `to_string(val)` | Convert value to string |

### Math Functions

| Function | Description |
|----------|-------------|
| `abs(n)` | Absolute value |
| `min(a, b)` | Minimum of two values |
| `max(a, b)` | Maximum of two values |
| `sqrt(n)` | Square root |
| `pow(base, exp)` | Power |
| `sin(x)`, `cos(x)`, `tan(x)` | Trigonometry |
| `floor(x)`, `ceil(x)`, `round(x)` | Rounding |

### Random and Time

| Function | Description |
|----------|-------------|
| `random_int(min, max)` | Random integer in range |
| `random_float(min, max)` | Random float in range |
| `get_time()` | Seconds since program start |
| `sleep(seconds)` | Pause execution |
| `wait(seconds)` | Same as `sleep` |

### Type Checking and Conversion

| Function | Description |
|----------|-------------|
| `is_type(val, "typename")` | Check runtime type |
| `as_int(val)` | Convert to int |
| `as_float(val)` | Convert to float |
| `as_string(val)` | Convert to string |
| `to_int(val)` | Parse string to int |
| `to_float(val)` | Parse string to float |

### Array Functions

| Function | Description |
|----------|-------------|
| `array_create()` | Create empty dynamic array |
| `array_push(a, val)` | Add element at end |
| `array_get(a, i)` | Get element at index |
| `array_set(a, i, val)` | Set element at index |
| `array_len(a)` | Get length |
| `array_free(a)` | Free array memory |

### Dictionary Functions

| Function | Description |
|----------|-------------|
| `dict_create()` | Create empty dictionary |
| `dict_set(d, key, val)` | Set key-value pair |
| `dict_get(d, key)` | Get value by key |
| `dict_has(d, key)` | Check if key exists |
| `dict_len(d)` | Get entry count |
| `dict_free(d)` | Free dictionary memory |

### Vector Functions

| Function | Description |
|----------|-------------|
| `make_vec2(x, y)` | Create 2D vector |
| `make_vec3(x, y, z)` | Create 3D vector |
| `vec2_add(a, b)` | Add vectors |
| `vec2_sub(a, b)` | Subtract vectors |
| `vec2_length(v)` | Get magnitude |
| `vec2_distance(a, b)` | Distance between points |
| `normalize(v)` | Unit vector |

---

## CLI Commands

The Cortex compiler is run from the command line:

### Basic usage

```bash
# Compile a single file
cortex -i program.cx -o program

# Compile and run immediately
cortex -i program.cx -run

# Compile with optimization
cortex -i program.cx -o program -O2
```

### Command-line options

| Option | Description |
|--------|-------------|
| `-i <file>` | Input file(s) to compile |
| `-o <name>` | Output executable name |
| `-run` | Compile and run immediately |
| `-O<level>` | Optimization level (0-3) |
| `-config <file>` | Use config file for library paths |
| `-use <lib>` | Link with a library |
| `-features <list>` | Enable features (e.g., `gui,network`) |
| `-verbose` | Show detailed output |
| `-help` | Show help message |

### Config files

For complex projects, use a JSON config file:

```json
{
  "include_paths": ["./include", "/usr/local/include"],
  "library_paths": ["./lib", "/usr/local/lib"],
  "libraries": ["raylib", "m"],
  "features": ["gui", "network"]
}
```

```bash
cortex -i game.cx -config game.json -o game
```

### Multi-file projects

```bash
# Compile multiple files
cortex -i main.cx utils.cx graphics.cx -o myapp

# Or use a project structure with imports
cortex -i main.cx -o myapp
```

---

## C Interop

Cortex provides seamless integration with C libraries, allowing you to leverage the entire C ecosystem while writing modern, safe code.

### The Philosophy

Cortex doesn't reinvent the wheel—it gives you C's performance and library ecosystem with modern ergonomics:

- **Use any C library** — Include headers, call functions, link libraries
- **Automatic memory safety** — Cleanup annotations prevent leaks
- **No FFI boilerplate** — Just declare and call
- **Zero runtime overhead** — Compiles to clean C

### Including C Headers

Use standard C `#include` syntax:

```c
// Standard library
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>

// Third-party libraries
#include <raylib.h>
#include <sqlite3.h>
#include <curl/curl.h>

// Your own headers
#include "mylib.h"
```

Cortex passes these includes directly to the C compiler.

### Declaring External Functions

Use `extern` to declare C functions:

```c
// Basic declaration
extern int printf(string format, ...);
extern double sqrt(double x);

// With pointer types
extern void* malloc(int size);
extern void free(void* ptr);

// With cleanup annotation
extern void* my_alloc(int size) cleanup(free);
```

**Type mapping:**

| Cortex type | C type |
|-------------|--------|
| `int` | `int` |
| `float` | `float` |
| `double` | `double` |
| `string` | `char*` |
| `bool` | `int` (0/1) |
| `void` | `void` |
| `any` | `AnyValue` struct |

### Automatic Memory Management with Cleanup

The `cleanup` annotation automatically frees resources when they go out of scope:

```c
extern void* malloc(int size) cleanup(free);
extern void free(void* ptr);

extern FILE* fopen(string path, string mode) cleanup(fclose);
extern void fclose(FILE* f);

void main() {
    var buf = malloc(1024);      // Auto-freed!
    var file = fopen("data.txt", "r");  // Auto-closed!
    
    // Use buf and file...
    
}  // free(buf) and fclose(file) called automatically
```

**How it works:**
1. Cortex wraps the returned pointer in a managed handle
2. Uses GCC's `__attribute__((cleanup))` 
3. Calls the cleanup function when the variable leaves scope

**Benefits:**
- No memory leaks
- No use-after-free
- No double-free
- No forgotten cleanup

### Managed vs Borrowed Pointers

**Managed pointers** (with cleanup):
```c
extern void* malloc(int size) cleanup(free);
var buf = malloc(1024);  // Cortex owns, auto-freed
```

**Borrowed pointers** (no cleanup):
```c
extern const char* getenv(string name);  // No cleanup - borrowed from C
var home = getenv("HOME");  // Don't free this!
```

**Rules of thumb:**
- If C documentation says "caller must free", use `cleanup`
- If pointer comes from internal storage, omit `cleanup`
- When in doubt, check the library docs

### Working with C Structs

For simple C structs, use Cortex structs with matching layout:

```c
// C: typedef struct { float x, y; } Vector2;
struct Vector2 {
    float x;
    float y;
}

// C: Vector2 Vector2Add(Vector2 a, Vector2 b);
extern Vector2 Vector2Add(Vector2 a, Vector2 b);

void main() {
    Vector2 v1 = { .x = 10.0, .y = 20.0 };
    Vector2 v2 = { .x = 5.0, .y = 3.0 };
    Vector2 sum = Vector2Add(v1, v2);
}
```

For complex C structs (unions, bitfields, etc.), use `void*` and accessor functions:

```c
extern void* sqlite3_open(string path);
extern int sqlite3_exec(void* db, string sql, void* callback, void* arg, string* err);
extern void sqlite3_close(void* db);

void main() {
    void* db = sqlite3_open("my.db");
    defer { sqlite3_close(db); };
    
    sqlite3_exec(db, "CREATE TABLE test(id INT)", null, null, null);
}
```

### Callbacks and Function Pointers

Pass Cortex functions to C as callbacks:

```c
// C: void register_handler(void (*handler)(int event));
extern void register_handler(void (*handler)(int event));

void my_handler(int event) {
    println("Event received: ${event}");
}

void main() {
    register_handler(my_handler);
}
```

For lambdas/closures, use the `[]` syntax:

```c
extern void sort_array(int* arr, int n, int (*compare)(int a, int b));

void main() {
    int[] nums = [3, 1, 4, 1, 5, 9];
    
    // Pass a comparison function
    sort_array(nums.data, nums.length, [](int a, int b) -> int {
        return a - b;  // Ascending order
    });
}
```

### Linking Libraries

Three ways to link:

**1. Auto-link from include:**
```c
#include <raylib.h>  // Automatically links -lraylib
```

**2. Explicit pragma:**
```c
#pragma link("mylib")
#pragma link("pthread")
```

**3. Command line:**
```bash
cortex -i main.cx -o app -use raylib -use pthread
```

### Platform-Specific Code

Use preprocessor conditionals for platform differences:

```c
#ifdef WINDOWS
    #pragma link("ws2_32")
    extern int WSAGetLastError();
#elif LINUX
    #pragma link("pthread")
    extern int errno;
#elif MACOS
    #pragma link("-framework Foundation")
#endif
```

### Embedding Raw C

For code that can't be expressed in Cortex:

```c
// Global C code
@c #define VERSION "1.0"
@c static int global_state = 0;

// Inline C code
void main() {
    @c printf("Direct C: %d\n", global_state);
    
    // Mix Cortex and C
    int x = 42;
    @c printf("x from Cortex: %d\n", x);
}
```

### Common C Libraries

**Standard library:**
```c
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void main() {
    printf("Hello from C!\n");
    var buf = malloc(1024);
    strcpy(buf, "Cortex");
}
```

**Math library:**
```c
#include <math.h>
#pragma link("m")

void main() {
    double x = sin(3.14159 / 2);
    double y = pow(2.0, 10.0);
    printf("sin(pi/2) = %f, 2^10 = %f\n", x, y);
}
```

**raylib (game development):**
```c
#include <raylib.h>

void main() {
    InitWindow(800, 450, "Cortex Game");
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

**SQLite:**
```c
#include <sqlite3.h>
#pragma link("sqlite3")

extern int sqlite3_open(string path, void** db) cleanup(sqlite3_close);
extern void sqlite3_close(void* db);

void main() {
    void* db;
    sqlite3_open("my.db", &db);  // Auto-closed!
    
    // Use database...
}
```

### Best Practices

**1. Always use cleanup annotations:**
```c
// Good - automatic cleanup
extern void* malloc(int size) cleanup(free);
var buf = malloc(1024);

// Bad - manual cleanup, error-prone
extern void* malloc(int size);
var buf = malloc(1024);
defer { free(buf); };  // Easy to forget
```

**2. Check return values:**
```c
extern FILE* fopen(string path, string mode) cleanup(fclose);

void main() {
    var file = fopen("data.txt", "r");
    if (file == null) {
        println("Failed to open file!");
        return;
    }
    // Use file...
}
```

**3. Use defer for non-cleanup resources:**
```c
extern void lock_mutex(void* m);
extern void unlock_mutex(void* m);

void main() {
    lock_mutex(my_mutex);
    defer { unlock_mutex(my_mutex); };
    
    // Critical section...
}
```

**4. Wrap C APIs in Cortex functions:**
```c
// Low-level C API
extern void* sqlite3_open(string path);
extern int sqlite3_exec(void* db, string sql, void* cb, void* arg, string* err);
extern void sqlite3_close(void* db);

// High-level Cortex wrapper
struct Database {
    void* handle;
    
    void exec(string sql) {
        sqlite3_exec(handle, sql, null, null, null);
    }
}

Database open_db(string path) {
    Database db;
    sqlite3_open(path, &db.handle);
    return db;
}

void close_db(Database db) {
    sqlite3_close(db.handle);
}

void main() {
    var db = open_db("my.db");
    defer { close_db(db); };
    
    db.exec("CREATE TABLE test(id INT)");
}
```

### C Interop Quick Reference

| Task | Syntax |
|------|--------|
| Include header | `#include <lib.h>` |
| Declare function | `extern int func(int x);` |
| With cleanup | `extern void* func() cleanup(free);` |
| Link library | `#pragma link("lib")` or `-use lib` |
| Raw C code | `@c int x = 0;` |
| Pass callback | `register_handler(my_func);` |
| C struct | Match layout in Cortex struct |
| C pointer | Use `void*` or typed pointer |

---

## Quick reference

| Topic           | Syntax / API |
|----------------|--------------|
| Entry point    | `void main() { }` |
| Print line     | `println("text");` or `show("text");` |
| Variables      | `int x = 5;`, `var x = 5;`, `any x = 5;` |
| Const          | `const x = 10;` or `const int x = 10;` |
| Conditionals   | `if (cond) { } else if (cond) { } else { }` |
| Loops          | `for (int i=0; i<n; i++)`, `while (cond)`, `repeat (n)`, `for (x in arr)` |
| Functions      | `int f(int a, int b) { return a+b; }` |
| Multiple returns | `(int, int) f() { return (a, b); }` |
| Struct         | `struct T { int x; }` then `T t; t.x = 1;` |
| Struct method  | `void move(int dx) { x = x + dx; }` → `t.move(5);` |
| Enum           | `enum E { A, B }` then `int e = A;` |
| Array literal  | `var a = [1, 2, 3];` then `a[0]` |
| 2D array       | `var m = [[1,2],[3,4]];` then `m[i][j]` |
| Dynamic array  | `array_create`, `array_push`, `array_get`, `array_len`, `array_free` |
| Dict literal   | `dict d = { "k": v };` |
| Dictionary     | `dict_create`, `dict_set`, `dict_get`, `dict_has`, `dict_free` |
| Result         | `result_ok(val)`, `result_err("msg")`, `result_is_ok(r)`, `result_value(r)` |
| Match          | `match (val) { case int n: ... case string s: ... default: ... }` |
| Match result   | `match (r) { case Ok(v): ... case Err(e): ... }` |
| Defer          | `defer { cleanup_code(); }` |
| Coroutine      | `coroutine void f() { co_yield(); }` |
| Async API      | `async_create(fn, arg)`, `async_resume(co)`, `async_await(co)` |
| Threading      | `spawn fn(args)`, `spawn t = fn(args)`, `thread_join(t)` |
| Channels       | `channel_create(size, cap)`, `channel_send(ch, &v)`, `channel_recv(ch, &out)` |
| Vectors        | `make_vec2(x,y)`, `vec2_add`, `vec2_length`, `normalize` |
| Random         | `random_int(min,max)`, `random_float(min,max)` |
| Time           | `get_time()`, `sleep(sec)`, `wait(sec)` |
| Tests          | `test "name" { assert_eq(a,b); }` then `test_run_all();` |
| Extern         | `extern int c_func(int a);` |
| C library      | `#include <lib.h>` and build with `-use lib` or config |
| Raw C          | `@c int x = 0;` |

---

## Where to go next

- **Examples** — See the `examples/` folder: `hello.cx`, `guess_game.cx`, `struct_methods.cx`, `app_file.cx`, and the raylib examples under `examples/raylib/`.
- **Full reference** — See the main [README](README.md) and [LANGUAGE_SPEC.md](LANGUAGE_SPEC.md) for the complete language and compiler options.
- **Building** — `go build -o cortex .` then `cortex -i file.cx -o program` or `cortex -i file.cx -run`.

You now have a full, beginner-oriented picture of Cortex: variables, control flow, functions, structs, enums, arrays, dicts, results, vectors, tests, and C libraries, all with examples.
