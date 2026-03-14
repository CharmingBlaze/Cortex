# What’s Missing in Cortex

A single place to see what’s **not yet implemented**, what’s **partially done**, and what’s **planned**, so you know the current gaps.

---

## Language / runtime gaps

| Area | Missing | Notes |
|------|--------|--------|
| **Dict literals** | ~~`{ "a": 1, "b": 2 }`~~ | **Implemented.** Use `dict d = { "a": 1, "b": 2 };` |
| **Multi-dimensional arrays** | ~~`[[1,2],[3,4]]`, `arr[i][j]`~~ | **Implemented.** Use `var m = [[1,2],[3,4]];` and `m[i][j]` |
| **Lambda captures** | ~~`[x](int a) { return x + a; }`~~ | **Implemented.** By-value capture; passed as (fn, &env); use in call-argument position. |
| **Named / default parameters** | ~~`f(x: 1, y: 2)` or `void f(int x = 0)`~~ | **Implemented.** Named args merged by param order; default values in function params. |
| **Generics / templates** | `vector<int>`, `optional<string>` | LANGUAGE_SPEC mentions them as “wanted”; not implemented. |
| **Modules / namespaces** | `module math;`, `import math.vector;` | Multi-file is “merge” only; no real module system. |
| **Async/await** | `async` / `await` | Keywords gated by config; no real implementation (stretch). |
| **Actors / channels** | `spawn`, channels | Keywords gated; full implementation is stretch. |
| **Nested JSON** | ~~Arrays and nested objects~~ | **Implemented.** AnyValue has TYPE_DICT/TYPE_ARRAY; json_parse_value and json_stringify_any handle nested objects/arrays; use as_dict/as_array on values. |
| **`parse_number` (example)** | ~~No built-in~~ | **Implemented.** `parse_number(string)` → float, `parse_int(string)` → int (0 on failure). |

---

## Spec / doc vs reality

- **LANGUAGE_SPEC.md** describes many “wanted” features (generics, modules, pipeline operator, attributes, etc.) that are **not** implemented. Treat the spec as aspirational where it goes beyond the README/Implemented list.
- **README “What’s still missing” table**: Lambdas row now says **implemented**.
- **README “Removed Features”**: Preprocessor line correctly states Cortex has `#include`, `#pragma`, `#use`, `#define` (passed through); no full C preprocessor.

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
- Coroutines / yield (e.g. for timelines, tweening)
- ~~Modules / namespaces (minimal: `module "math";`, `math.func()`)~~ **Done**
- ~~Nested JSON in parse/stringify~~ **Done**
- ~~ECS helpers (entity_id, add_component, get_component)~~ **Done**
- Hot reloading (stretch)

---

## Quick “am I blocked?” checklist

- **Games / 2D** — vec2, random, time, raylib interop: **in place.**
- **Data in memory** — structs, enums, array, dict, result: **in place.**
- **Scripting-style** — var, any, type checks at runtime: **in place.**
- **C libs** — #include, -use, config: **in place.**
- **Async / concurrency** — **minimal:** `async`/`await` compile when `features.async` is enabled (run synchronously for now).
- **Rich JSON** — **in place** (nested parse/stringify, as_dict/as_array).
- **Generics / modules** — **modules:** minimal `module "name";` + `name.func()`; **generics** not in place.

So for “games and applications” the main remaining gaps are: async, generics, and a full module system (if you need them).
