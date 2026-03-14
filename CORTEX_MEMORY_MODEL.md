# Cortex Memory & Ownership Model

## 1. Goals

1. **Predictability** – No dangling pointers or manual free, deterministic semantics for structs/arrays.
2. **Performance** – Zero-cost value types, escape analysis, arenas for hot paths.
3. **Interop** – Clear rules when passing data to/from C via `extern`.
4. **Safety** – Default immutability (`let`), explicit `var` for mutation, borrow-checked references in critical paths.

## 2. Runtime Architecture

### 2.1 Heap Manager
- Primary heap uses **region-based arenas** with fallback **generational tracing GC**.
- Each async task/actor receives a default arena; arenas recycle when the task completes.
- Large objects (≥ 1 MiB) are allocated via OS virtual memory and tracked separately.
- The GC is **incremental, tri-colour, mostly concurrent**.

### 2.2 Stack Frames
- Functions run on native stacks; `async` functions use segmented stacks.
- Value types (scalars, structs without heap members) live on the stack unless they escape.
- Escape analysis decides when to promote stack allocations to heap.

## 3. Type Semantics

### 3.1 Value Types
- Primitives (`int`, `float`, `bool`, `vec2`, `vec3`, `char`) are **copy-by-value**.
- `struct` defaults to copy-by-value. Use `@ref` attribute to mark reference semantics.
- Arrays declared as `array[T]` are value types containing a pointer/length pair; copying clones metadata, not contents. Use `.clone()` for deep copy.

### 3.2 Reference Types
- `string`, `any`, `vec2[]`, `var` dynamic objects, actors, lambda captures, and `optional<T>` when `T` is reference are heap-managed with reference counts plus cycle detection (part of GC).

### 3.3 `var` and `any`
- `var` binds to immutable value by default; mutation requires `var mut name` or `name = ...` for smart values.
- `any` stores an `AnyValue` described in `generator.go`; conversions call `as_*` helpers.

## 4. Ownership Rules

### 4.1 Default Ownership
- Each heap allocation has a single logical owner (function, struct, actor).
- Passing ownership transfers via move semantics: `let b = move(a);` invalidates `a` until re-assigned.
- Copying reference types increments RC automatically; drop decrements.

### 4.2 Borrowing
- `ref`/`out` parameters borrow mutable references with compile-time checking:
  - Only one mutable borrow active at a time.
  - Multiple immutable borrows allowed if no mutable borrow exists.
- Borrow scopes end at function exit or when explicitly released (`endborrow name;`).

### 4.3 Defer & Finalizers
- `defer { ... }` runs even if the block returns or throws.
- Finalizers can be registered via `@finalizer(fn)` attribute on struct definitions.

## 5. Interaction with C Interop

### 5.1 Passing Data to C
- Value types pass by value (copied onto C stack).
- Reference types expose raw pointers via `as_ptr(value)`; developer must ensure lifetime extends past the C call.
- `extern` functions returning pointers must wrap them in `c_ptr[T]`, which requires explicit `.free()`.

### 5.2 Taking Data from C
- `extern` declarations returning structs/arrays treat them as value types; Cortex copies the memory.
- To avoid copies, use `c_view[T]` which references C memory without ownership; unsafe operations require `@unsafe` annotation.

## 6. Optional/Result + Memory
- `optional<T>` holds either an inline `T` (if value type) or reference to heap-managed box.
- `result<T, E>` stores `union { T ok; E err; }` plus discriminant; destructors run when the result drops.

## 7. GC & Performance Tuning

- Compiler flag `--gc=off` switches to pure arena RC (for embedded targets).
- `@arena` attribute forces allocations into a named arena.
- `gc_collect()` triggers a collection; `gc_pause()` suspends GC during critical sections.

## 8. Examples

```c
struct Texture {
    @ref
    handle: c_ptr<void>;
    width: int;
    height: int;
}

actor Loader {
    queue: channel<string>;

    async fn start(mut self) {
        defer { println("loader closed"); }
        while (let path = await self.queue.recv()) {
            let tex = load_texture(path);
            self.process(tex);
            // tex drops here; GC frees handle if @finalizer defined
        }
    }
}

fn borrow_example(ref mut data: array<int>) {
    defer { println("borrow done"); }
    data[0] = 42;        // allowed: mutable ref
    let first = &data[0]; // compile error: immutable borrow while mutable active
}
```

This document defines the baseline for cortex compiler/runtime implementers. Further sections (concurrency, modules, generics, build system) will build upon these ownership semantics.
