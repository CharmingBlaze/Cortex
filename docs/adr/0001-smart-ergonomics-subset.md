# ADR 0001: Smart ergonomics subset (pointer-free)

## Status

Accepted (directional). Implementation is incremental.

## Context

Cortex targets **simple C + smart ergonomics** (similar in spirit to [Jule](https://github.com/julelang/jule): intentional interop, safety-minded defaults) while **compiling to C**, not to C++. Users must **not** write `*T` / `&` in normal code; see [POINTER_FREE_AND_FFI.md](../POINTER_FREE_AND_FFI.md).

## Decision

Adopt a **small, named** set of “smart” features. Each must have a **documented lowering to C** and **interop rules** for `extern` / `@c`.

### 1. Scoped resources (priority)

- Extend **`defer`** and **`cleanup` on `extern`** as the primary RAII-style story.
- Optional later: constructor/destructor **sugar** on structs that lowers to init/fini calls—only if it does not require C++.

### 2. Optional / result types

- Keep and refine **`optional`** (and related result-style patterns) for control flow without exceptions.
- No C++ exceptions; no Jule-style “exceptionals” unless explicitly chosen later.

### 3. Methods / UFCS

- Allow **method call syntax** where it reduces noise for **structs and bindings**, as long as lowering stays **plain C** (first parameter + name mangling or static functions).

### 4. Explicit escapes

- **`@c { ... }`**: raw C injection; pointer semantics allowed **only** here (or in carefully documented `extern` lines).
- **`extern`** signatures that mirror C headers may still mention pointer types textually; **generated bindings** should prefer **handles** and **documented wrappers** where possible.

## Compile-time capabilities (deferred)

Large compile-time features (full reflection, macro systems, Jule-scale `comptime`) are **out of scope** until:

1. Pointer-free FFI and **binder** quality are stable.
2. **Strict / bounds** diagnostics are in good shape.

**Near-term** work stays with existing **constant folding** / optimizer paths and **const** where already supported; revisit **compile-time evaluation** only when it unlocks **bindings** or **serialization**, not as a standalone goal.

## Consequences

- Compiler and **binder** teams optimize for **handle + slice + string** lowering, not for exposing C pointer syntax to users.
- New syntax proposals must answer: **how does this lower to C**, and **does it leak pointers into `.cx`?**
