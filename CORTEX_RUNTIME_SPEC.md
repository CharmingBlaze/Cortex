# Cortex Concurrency, Modules, and Runtime Specification

## 1. Concurrency Model

### 1.1 Executors & Scheduling
- Cortex uses a **work-stealing scheduler** per process.
- Each actor/async task runs on logical fibers multiplexed over OS threads.
- Scheduler guarantees **fair scheduling** for ready tasks; blocking calls must use `await` wrappers to avoid stalling threads.

### 1.2 Async/Await
```c
aSYNC fn fetch_data(url: string) -> result<string, net_error> {
    let conn = await net::open(url);
    defer { conn.close(); }
    let bytes = await conn.read_all();
    return ok(bytes);
}
```
- `async fn` transforms into a state machine; `await` yields to the scheduler.
- Exceptions inside async functions bubble through the `result` if declared, otherwise terminate the task and log.

### 1.3 Actors
```c
actor Worker {
    inbox: channel<Job>;

    async fn run(mut self) {
        while (let job = await self.inbox.recv()) {
            job.execute();
        }
    }
}
```
- `actor` is syntactic sugar for a struct with an inbox channel and a spawned async task.
- Actors own their state; cross-actor mutation requires message passing.
- Channels are typed (`channel<T>`), bounded by default (`channel<T>(capacity = 128)`), blocking semantics use `await recv()`.

### 1.4 Pipelines & Tasks
- `spawn expr` schedules an async block immediately.
- `pipeline` operator can be combined with async tasks: `jobs |> map_async(process) |> await_all();`
- Race detector can be enabled via `--race` flag; it instruments mutable borrows and channel edges.

## 2. Error Handling

### 2.1 Optional & Result Syntax
```c
fn find_user(id: int) -> optional<User> { ... }
let maybe = find_user(42);
match maybe {
    some(user) => println(user.name);
    none => println("missing");
}

fn parse(text: string) -> result<int, parse_error> { ... }
let value = try parse("123?"); // propagates err automatically
```
- `optional<T>` sugar: `T?`.
- `result<T, E>` sugar: `T ! E`.
- `try expr` works on `result`; on error it returns early with `err`.
- `throw err` wraps value in `err` for current result.

### 2.2 Defer, Catch, Finally
```c
try {
    critical();
} catch (err) {
    log(err);
} finally {
    cleanup();
}
```
- `defer` executes when scope exits, even if `throw` or `try` propagate.
- `catch` can only appear with `try`.

## 3. Module & Package System

### 3.1 File Layout
- Each file begins with optional `module` declaration: `module game::ecs;`
- Directory structure mirrors module namespaces.
- `import foo::bar as baz;` resolves relative to project roots defined in `cortex.toml`.

### 3.2 Visibility
- `pub` keyword exports symbols from the current module.
- Symbols are private by default.
- Re-export via `pub use foo::bar;`.

### 3.3 `cortex.toml`
```toml
[project]
name = "cortex-game"
version = "0.1.0"

[build]
target = ["wasm32", "win64"]
opt-level = "speed"

deps = {
    "graphics" = "github:studio/graphics#main",
    "network" = "1.4"
}
```
- `[build] target` accepts multiple triples; toolchain performs multi-output builds.
- Dependency resolver supports git URLs, version ranges, local paths.

## 4. Generics & Traits

### 4.1 Generic Functions & Types
```c
fn clamp<T: Comparable>(value: T, min: T, max: T) -> T { ... }
struct Option<T> { ... }
```
- Monomorphization by default; `@erased` forces type-erasure for dynamic dispatch.
- Traits define constraints:
```c
trait Comparable {
    fn cmp(self, other: Self) -> ordering;
}
```
- Implementations attach via `impl Comparable for vec2 { ... }`.

### 4.2 Variance & Lifetimes
- Types are invariant unless annotated with `@covariant` or `@contravariant`.
- Lifetimes expressed as `'a` for reference types when interacting with borrow checker; elided by default.

## 5. Standard Library Contracts

| Area | Key Modules |
| --- | --- |
| Collections | `std::array`, `std::vector`, `std::map`, `std::set`, `std::queue` |
| Strings & Text | `std::string`, `std::text`, Unicode normalization helpers |
| Math | `std::math` (scalar), `std::vec`, `std::matrix`, `std::complex` |
| IO | `std::fs`, `std::net`, `std::sys`, `std::process` |
| Concurrency | `std::channel`, `std::actor`, `std::task`, `std::sync` |
| Crypto/Blockchain | `std::crypto::sha`, `std::crypto::keccak`, `std::blockchain` |
| Testing | `std::test` (assertions, parameterized tests) |

Each API is versioned; incompatible changes require bumping the minor version in `cortex.toml`.

## 6. Runtime & Build Behavior

### 6.1 Compiler Flags
- `--target <triple>`: `win64`, `linux64`, `macos-arm64`, `wasm32`, `wasm32-threaded`.
- `--opt <debug|speed|size>`: Controls inlining, GC heuristics.
- `--test`: Runs `test` blocks without producing final binary.
- `--watch`: Hot reload; rebuilds modules on change.
- `--race`: Enables race detector.

### 6.2 Linking & Artifacts
- Default output: single executable per target.
- `--emit c` stores generated C files for auditing.
- `--module <name>` compiles to `.cortexmod` library for reuse.

### 6.3 Hot Reload Lifecycle
- `hot_reload { ... }` block runs post-reload.
- Compiler emits reloadable shared library segments; runtime swaps vtables safely.

### 6.4 Error Reporting
- Diagnostics follow `path:line:col: error: message` format.
- Compiler exposes JSON diagnostic stream with `--json` for IDE integration.

## 7. Example Build Flow

```
$ cortex build --target win64 --opt speed
$ cortex test --json > results.json
$ cortex run --watch
```

This runtime specification completes the high-level design alongside the grammar and memory documents, enabling compiler and tooling implementers to proceed with full coverage of Cortex’s language surface.
