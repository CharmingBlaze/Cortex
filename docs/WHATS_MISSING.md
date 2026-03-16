# What's Missing in Cortex

A single place to see what's **not yet implemented**, what's **partially done**, and what's **planned**, so you know the current gaps.

---

## Language / runtime gaps

| Area | Missing | Notes |
|------|--------|--------|
| **Dict literals** | ~~`{ "a": 1, "b": 2 }`~~ | **Implemented.** Use `dict d = { "a": 1, "b": 2 };` |
| **Multi-dimensional arrays** | ~~`[[1,2],[3,4]]`, `arr[i][j]`~~ | **Implemented.** Use `var m = [[1,2],[3,4]];` and `m[i][j]` |
| **Lambda captures** | ~~`[x](int a) { return x + a; }`~~ | **Implemented.** By-value capture; passed as (fn, &env); use in call-argument position. |
| **Named / default parameters** | ~~`f(x: 1, y: 2)` or `void f(int x = 0)`~~ | **Implemented.** Named args merged by param order; default values in function params. |
| **Generics / templates** | ~~`vector<int>`, `optional<string>`~~ | **Implemented.** Generic type parsing and code generation for vector<T> and optional<T>. |
| **Modules / namespaces** | ~~`module math;`, `import math.vector;`~~ | **Implemented.** `module "name";` + `name.func()` syntax supported. |
| **Async/await** | ~~`async` / `await`~~ | **Implemented.** Coroutine-based async with yield, async_run_all(). |
| **Actors / channels** | ~~`spawn`, channels~~ | **Implemented.** Thread-safe channels with send/recv, spawn for threads. |
| **Nested JSON** | ~~Arrays and nested objects~~ | **Implemented.** AnyValue has TYPE_DICT/TYPE_ARRAY; json_parse_value and json_stringify_any handle nested objects/arrays; use as_dict/as_array on values. |
| **`parse_number` (example)** | ~~No built-in~~ | **Implemented.** `parse_number(string)` → float, `parse_int(string)` → int (0 on failure). |
| **Method syntax** | ~~`arr.push(x)`, `s.len()`~~ | **Implemented.** Method calls on array, string, dict, vector types. |
| **Null coalescing** | ~~`??` operator~~ | **Implemented.** `value ?? default` syntax. |
| **Optional chaining** | ~~`?.` operator~~ | **Implemented.** `obj?.member` safe access. |
| **Range operators** | ~~`..` and `..<`~~ | **Implemented.** `0..10` inclusive, `0..<n` exclusive. |
| **Try/catch/throw** | ~~Exception handling~~ | **Implemented.** Structured exception handling with try/catch/throw. |

---

## Tooling & ecosystem

| Area | Missing |
|------|--------|
| **IDE / LSP** | No language server; no rich diagnostics over the wire. |
| **Structured diagnostics** | **Partially wired:** semantic analyzer supports `SetDiagnosticsCollector()`; undefined-identifier and feature-gate errors emit structured diagnostics (line, column, code, suggestion). Lexer/parser can be wired similarly. |
| **Optimizer** | **Wired.** `internal/optimizer` runs after semantic; constant folding (e.g. `2+3` → `5`) is enabled. See README Compiler Architecture. |
| **Hot reload** | Not implemented (stretch). |
| **ECS helpers** | ~~No built-in~~ | **Implemented.** `entity_create()`, `add_component(id, "name", val)`, `get_component(id, "name")`, `has_component(id, "name")`, `entity_remove(id)`. |

---

## Future enhancements (from README)

- IDE integration (LSP, diagnostics)
- ~~Dictionary/map literals~~ **Done**
- ~~Multi-dimensional array literals and indexing~~ **Done**
- ~~Lambda captures (by-value)~~ **Done**
- Coroutines / yield (e.g. for timelines, tweening) — **Done**
- ~~Modules / namespaces (minimal: `module "math";`, `math.func()`)~~ **Done**
- ~~Nested JSON in parse/stringify~~ **Done**
- ~~ECS helpers (entity_id, add_component, get_component)~~ **Done**
- ~~Generics (vector<T>, optional<T>)~~ **Done**
- ~~Async/await with concurrency~~ **Done**
- ~~Channels for message passing~~ **Done**
- Hot reloading (stretch)

---

## Quick "am I blocked?" checklist

- **Games / 2D** — vec2, random, time, raylib interop: **in place.**
- **Data in memory** — structs, enums, array, dict, result: **in place.**
- **Scripting-style** — var, any, type checks at runtime: **in place.**
- **C libs** — #include, -use, config: **in place.**
- **Async / concurrency** — **in place:** `async`/`await` with coroutines, channels for message passing.
- **Rich JSON** — **in place** (nested parse/stringify, as_dict/as_array).
- **Generics / modules** — **in place:** generic types `vector<T>`, `optional<T>`; modules with `module "name";` + `name.func()`.

**For games and applications, all major features are now implemented!**
