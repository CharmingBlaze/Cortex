# Cortex Changelog

All notable changes to Cortex will be documented in this file.

## [0.2.0] - 2026-03-15

### Added

#### Optional Types
- **Optional type syntax**: `int?`, `string?`, `User?`
- **Postfix `?` operator**: Check if optional has value (`maybe?` returns bool)
- **Postfix `!` operator**: Force unwrap optional (`maybe!` returns value)

#### Match Expressions
- **Match as expressions**: `string result = match n { 1 => "one", _ => "other" };`
- **Nested match**: Match expressions can be nested inside cases
- **Comma-separated cases**: Clean syntax with comma separators

#### Arrow Functions
- **Typed parameters**: `(int a, int b) => a + b`
- **Type inference with `fn`**: `fn add = (int x, int y) => x + y;`

#### Arrays
- **Array type syntax**: `int[]`, `string[]`, `int[][]`
- **Curly brace literals**: `int[] nums = {1, 2, 3, 4, 5};`

### Fixed
- Optional unwrap in if conditions now generates correct C code
- Postfix `?` and `!` operators properly distinguished from prefix `!` (logical NOT)
- Arrow functions with typed parameters parse correctly
- Array type declarations work with `Type[]` syntax

### Release Packages
- Windows (x64, ARM64) with bundled Zig CC compiler
- Linux (x64, ARM64) with bundled Zig CC compiler
- macOS (Intel, Apple Silicon) with bundled Zig CC compiler

**No external C compiler required!** Zig CC is bundled with all releases.

---

## [0.1.0] - 2024-03-15

### Added

#### Control Flow
- **`elif`** - Sugar for `else if` chains (Python/Swift/Kotlin style)
  ```c
  if x == 5 { ... } elif x == 10 { ... } else { ... }
  ```
- **`unless`** - Inverse of `if`, runs when condition is false (Ruby-style)
  ```c
  unless x < 0 println("non-negative");
  ```
- **`if let`** - Pattern matching for optional values
  ```c
  if let name = maybe_name { println("Hello, ${name}!"); }
  ```
- **Optional parentheses in if** - `if x > 5 { }` or `if (x > 5) { }`
- **Single-statement if bodies** - `if x > 0 println("positive");`
- **`loop { }`** - Sugar for `while (true) { }`

#### Operators
- **Compound assignment operators**: `+=`, `-=`, `*=`, `/=`
- **Increment/decrement operators**: `++`, `--` (prefix and postfix)

#### Variables
- **`const` keyword** - Declare constants with type inference
  - `const greeting = "Hello";`
  - `const int max_score = 100;`

#### Structs
- **Struct methods** - Define methods inside structs with implicit `self`
- **Dot notation** - Access struct fields and methods with `.`

#### Build System
- **CLI commands**: `cortex new`, `cortex run`, `cortex build`, `cortex bind`
- **Project configuration** via `cortex.toml`
- **C library binding generator** - Generate Cortex bindings from C headers

#### Examples
- `examples/raylib/raylib_pong_methods.cx` - Pong with struct methods
- `examples/test_operators.cx` - Compound and increment operators
- `examples/test_loop.cx` - Loop sugar
- `examples/test_if_modern.cx` - Optional if parentheses
- `examples/test_const.cx` - Const declarations

### Documentation
- Updated `LANGUAGE_SPEC.md` with operator syntax
- Added clean examples for all new features
- Polished if statement documentation

---

## Future Plans

### Planned Features
- `for (item in collection)` iteration
- Async/await support
- Actor model for concurrency
- GUI toolkit improvements

### Under Consideration
- Module system improvements
- Package manager integration
- LSP support for IDEs
- WebAssembly backend
