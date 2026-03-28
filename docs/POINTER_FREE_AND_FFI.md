# Pointer-free Cortex and C interop

This document is the **normative roadmap** for how Cortex stays **free of user-visible pointers** while still compiling to C and calling real C libraries.

## Goals

- In `.cx` source, users do **not** write C pointer syntax: no `*T`, no `&x`, no pointer arithmetic.
- **Sequences** are **fixed arrays** (known length) and, as the type system grows, **slices** (view: data + length)—not raw addresses.
- **Strings** are values; the compiler lowers them to `char *` / string buffers at `extern` call sites where required.
- **Opaque handles** (from bindings or `extern`) represent C pointers without exposing `*` in Cortex.
- **Escape hatches** are explicit: `@c { ... }` and narrowly documented `extern` declarations.

## What the compiler does today

- **Array literals** and **indexing** (`arr[i]`) are lowered to C arrays or compatible representations.
- **`for (x in collection)`** iterates without exposing addresses.
- **`extern` functions** may still use pointer types in their **signature text** when mirroring C headers; prefer generated **bindings** that wrap common patterns.
- **`#include`** pulls in C headers; **[`internal/clibs`](../internal/clibs/)** infers link names and loads **`configs/<lib>.json`** when present ([Binding Guide](BINDING_GUIDE.md)).
- **`@c { ... }`** injects raw C into generated output—use only when the pointer-free surface is insufficient.

## FFI lowering (target behavior)

| Cortex concept | Typical C at the boundary |
|----------------|---------------------------|
| `string` / string literal | `char *` to a temporary or stable buffer; callee must not retain unless documented |
| Fixed array / future slice | `T *` + length (`size_t`); synthesized at call site |
| Struct value passed where C wants pointer | Address of temporary or stack value—**lifetime** ends after the call unless API is documented otherwise |
| Opaque handle | `void *` or typedef behind a Cortex `struct` name |

### Stable pointers and callbacks

If a C API **stores** a pointer past the call (callbacks, async completion, global registration), the **temporary lowering** above is **unsound**. Cortex should **diagnose** this where possible and direct users to:

- **`@c`** blocks with explicit C management, and/or
- A future **pinned buffer** / **arena** type, and/or
- Wrapper C code in the project.

## Slices (roadmap)

**Today:** fixed-size arrays and array literals are the main sequence model.

**Planned:** a **slice** type as a **view** `(reference, length)` in the language, lowered to `T *` + `size_t` only in generated C—never as user-written `*T`.

## Related documents

- [Binding Guide](BINDING_GUIDE.md) — `#include`, `configs/`, `cortex bind`
- [Language Spec](../LANGUAGE_SPEC.md) — syntax and current guarantees
- [ADR 0001: Smart ergonomics subset](adr/0001-smart-ergonomics-subset.md) — scoped resources, optional, methods (pointer-free)
