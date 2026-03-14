# Cortex Feature Implementation Plan

This document outlines the architecture and execution strategy for the next wave of Cortex compiler/runtime features. Each section pairs language/compiler work with runtime/tooling deliverables so we can iterate incrementally while keeping the system modular and testable.

## Global Guiding Principles

1. **Layered Architecture**: AST & parser feed semantic analysis, which feeds codegen, which targets the modular runtime (`runtime/core.*`). New features must fit this flow cleanly.
2. **Feature Flags Everywhere**: Async, actors, QoL, blockchain, etc. remain optional. Gating hooks must exist in parser (keyword acceptance), semantic analyzer (diagnostics), codegen (emission), and runtime (C macros).
3. **Focused Packages**: New subsystems get their own packages (e.g. `internal/lsp`, `internal/coroutine`, `internal/jsonext`). Keep public APIs narrow so we can swap implementations later.
4. **Diagnostics-First**: Every feature exposes good error messages (source range, helpful guidance) and pipes into CLI + upcoming LSP.
5. **Tests per Layer**: Parser fixtures, semantic unit tests, codegen golden files, runtime C tests (where practical), plus integration programs under `examples/`.

## Milestone 1 – Language Server Protocol (LSP)

Goal: ship a minimal-but-solid LSP server that reuses the compiler as a library.

- **Package Layout**: `internal/lsp` with submodules:
  - `protocol`: Go types mirroring LSP structures.
  - `server`: document store, connection loop, feature handlers.
  - `compilerhost`: thin adapter that runs the existing pipeline in-memory and returns diagnostics/AST info.
- **Features**:
  1. Document lifecycle (didOpen/didChange/didClose) with rope-style incremental text.
  2. Diagnostics: run parser+semantic after each change; stream structured errors back.
  3. Hover & go-to-definition: leverage scopes/symbol tables.
  4. Completion: surface keywords, locals, built-ins with context awareness.
- **Future-Proofing**: handler registry so rename/references/code actions plug in later without refactors.

## Milestone 2 – Dict Literals

- **Parser/AST**: ensure `{ "a": 1 }` is parsed as a `DictLiteralNode` with ordered entry list; support usage in expressions/arguments.
- **Semantics**: enforce string keys, convert values to `any` (implicit boxing). Error on non-string keys or duplicate keys (optional warning).
- **Codegen**: lower to calls on dict runtime API (`dict_create`, `dict_set_any`). Reuse temporary symbol allocation for inline literals.
- **Runtime**: ensure dict API exposes `dict_set_any` helper; extend QoL gating if needed.
- **Tests**: parsing fixture, semantic errors, codegen golden output, runtime example in `examples/dict_literals.cx`.

## Milestone 3 – 2D Array Literals & Indexing

- **Parser/AST**: allow nested array literals (`[[1,2],[3,4]]`). Tag `ArrayLiteralNode` with element dimensionality.
- **Semantics**: enforce rectangular shape (same length per row) unless ragged arrays explicitly allowed; ensure element types unify.
- **Codegen**: allocate runtime arrays with metadata (rows, cols) so `m[i][j]` compiles to two-stage bounds-checked access.
- **Runtime**: extend array helpers to accept 2D initialization data + metadata struct.
- **Tests**: invalid ragged arrays, out-of-range diagnostics, 2D iteration example.

## Milestone 4 – Lambda Captures (by-value)

- **Syntax**: `[x, y](params){ body }` capturing identifiers by value.
- **AST**: extend `LambdaNode` with capture list (names + capture mode for future by-ref support).
- **Semantics**: validate capture scope, forbid temporaries, handle shadowing, ensure only allowed use sites (call arguments for now).
- **Codegen**:
  - Emit environment struct per capture shape (dedup by hash to reuse definitions).
  - Allocate env struct, copy captured values, pass pointer + function symbol.
  - Generate helper signature `return_type lambda_fn(env*, params...)` and call conventions expected by runtime/stdlib.
- **Runtime**: ensure function pointer + env pair is representable (likely already in QoL runtime for lambdas).
- **Tests**: nested lambdas, captures of struct fields, duplicate captures error, etc.

## Milestone 5 – Nested JSON Support

- **Runtime** (`internal/jsonext`, `runtime/core.*`): rewrite `json_parse`/`json_stringify_any` to recursively handle dicts (`cortex_dict`) and arrays (`cortex_array`). Provide `as_dict(any)` / `as_array(any)` helpers.
- **Semantics**: register new helpers as built-ins (gated under QoL) and ensure type checker recognizes conversions.
- **Tests**: round-trip nested JSON, malformed error handling, large inputs.

## Milestone 6 – `parse_number` / `parse_int`

- **Runtime**: deterministic parsing (no locale). Use `strtod_l`/`strtol` equivalents or manual parsing for portability.
- **Semantic**: add built-ins returning `float` and `int`; optionally integrate with `Result` later.
- **Tests**: valid/invalid inputs, boundary conditions, whitespace handling.

## Milestone 7 – Named & Default Parameters

- **AST**: parameter nodes store optional default expressions; call argument nodes track name vs positional.
- **Parser**: enforce ordering rule (positional before named) and allow `param = expr` defaults.
- **Semantics**: default expressions must be compile-time constant; ensure no duplicate named args; fill omitted args with defaults; generate canonical argument ordering.
- **Codegen**: synthesize temporaries for defaults and call function with full positional parameter list.
- **Tests**: mixing named/positional, missing names, invalid defaults.

## Milestone 8 – Coroutines & `yield`

- **Syntax**: `coroutine foo() { ... yield ... }` + `yield` statement.
- **AST**: new node kinds for coroutine declarations and yield statements.
- **Semantics**: treat coroutine functions separately; enforce allowed statements inside; produce coroutine handle type.
- **Codegen**: transform coroutine body into state machine struct (program counter, locals, stack). Generate helper functions `coroutine_create`, `coroutine_resume`, `coroutine_destroy`.
- **Runtime**: add coroutine scheduler (manual resume) inside `runtime/core.*` or dedicated `runtime/coroutine.*` files.
- **Tests**: simple timeline, nested yields, kill/resume semantics.

## Milestone 9 – Async/Await (Synchronous Backend)

- **Parser**: tokens for `async` keyword (already planned) and `await` expression.
- **Semantics**: gate on `features.Async`; treat async functions similar to coroutines but with future objects; `await` allowed only inside async/coroutine contexts.
- **Codegen**: reuse coroutine state machine but run synchronously for now (await just calls and returns). Keep IR flexible for real async later.
- **Runtime**: minimal future/promise struct to satisfy ABI; synchronous completion path.
- **Tests**: nested async calls, awaits in loops, feature-flag errors when disabled.

## Milestone 10 – Hot Reloading (Stretch)

- **Design**: start with function-level hot reload via shared library rebind.
  - Abstract function pointer tables (`runtime/hotreload.*`).
  - Provide CLI workflow: compile → dlopen → patch function table.
- **Compiler Hooks**: emit metadata (symbol lists) for hot reload manager.
- **Tests**: integrate with dev harness (maybe a Go test simulating reload) plus documentation describing workflow.

## Cross-Cutting Concerns

- **Documentation**: update spec/guide per feature.
- **Examples**: add targeted programs under `examples/` to exercise each capability.
- **Tooling**: ensure LSP + CLI share configuration/loading logic for feature flags.
- **CI**: expand `go test ./...` plus runtime smoke tests (maybe via `examples/` runner).

---

This plan gives us a clear order of operations while leaving room for parallel work (e.g., LSP + dict literals). The next actionable step is to start Milestone 2 by auditing existing dict literal support and bringing parser/semantic/codegen/runtime in sync with the spec above.
