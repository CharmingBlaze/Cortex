# Outstanding Work

## Done (recent)
- **Runtime/codegen**: Rely on `runtime/core.h` and `runtime/core.c`; `toString` split into `toString_int`/`toString_float`/`toString_double`/`toString_bool` for C compatibility; string concatenation via `cortex_strcat`.
- **Config**: Sample configs added (`minimal.json`, `full_features.json`, `games_only.json`); README documents `-config` and feature flags; config file is respected (no overwriting explicit `false`).
- **Parser/semantic**: Do-while added; if/else fixed (block exit on `}`/`else`); variable decl semicolon after declaration; for-loop initializer uses `parseVariableDeclarationRest` so type is not consumed twice; postfix `++`/`--`; variable shadowing allowed (current-scope-only check).
- **Codegen**: Expression statement semicolons; print/println → `print_string`/`println_string`; for-loop initializer no extra semicolon; `ResolvedType` on identifiers from semantic for better expression types.
- **CLI**: Help lists `-config`; runtime dir resolution via `CORTEX_ROOT`, cwd, or executable-relative.
- **Full-featured**: Pointer types in declarations (`void*`, `char*`, multiple `*`); optional parameter names in `extern`; feature gates (async/await and actor/channel/spawn rejected when feature disabled); string/char escape sequences (`\n`, `\t`, `\\`, `\"`, etc.) in lexer and C output escaping; extern redeclaration of builtins allowed; automated tests for pipeline, extern pointers, feature gating, and string escapes.

## 1. Configuration & tests (optional)
- Scripts that compile under different configs and verify gating.
- More test coverage (parser edge cases, codegen).

## 2. Implemented (this pass)
- **defer** — `defer { block }` runs at block exit and before return (defer stack in codegen).
- **String interpolation** — `"text ${id} more"` in parser (identifier only); codegen emits cortex_strcat chain.
- **Multiple return values** — `(T1, T2) f()` and `return (a, b)`; codegen emits C struct.
- **Array literals** — `[e1, e2, e3]`; variable decl emits `type name[] = {...}; int name_len = N`.
- **Range-based for** — `for (x in arr)`; requires `arr` from array literal; uses `arr_len`.
- **Pattern matching** — `match (val) { case type var: ... case literal: ... default: ... }`; codegen emits if/else with is_type and as_*.

## 3. Stretch
- CLI `--enable` flags; end-to-end examples for blockchain + async when gating is in place.
- Lambda codegen (AST + parse exist; semantic/codegen still “not implemented”).
- Document and enforce QoL bundle for helpers.
