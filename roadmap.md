# Cortex Roadmap

This document outlines the planned development trajectory for Cortex.

## Current Status: Alpha

Cortex is in active development. The core language is functional, but APIs and features may change.

---

## Completed Features ✓

### Core Language
- [x] C-like syntax with type inference (`var`)
- [x] Functions, structs, enums
- [x] Control flow (if, while, for, for-in, match)
- [x] String interpolation
- [x] Array and dictionary literals
- [x] Lambdas with captures
- [x] Pattern matching

### Memory Management
- [x] Automatic cleanup annotations (`cleanup`)
- [x] Managed handles with GCC cleanup attribute
- [x] `defer` for scope-based cleanup

### Concurrency
- [x] Coroutines (cooperative multitasking)
- [x] Threads with `spawn`
- [x] Channels for thread communication
- [x] Cross-platform threading (pthreads/Windows threads)

### C Interop
- [x] `#include` for C headers
- [x] `extern` declarations
- [x] Automatic library linking from includes
- [x] Raw C code embedding (`rawc`)

### Standard Library
- [x] String operations
- [x] Array operations
- [x] File I/O
- [x] Network basics
- [x] GUI integration (Raylib)

### Tooling
- [x] CLI compiler
- [x] Run mode (`-run`)
- [x] Multi-file compilation
- [x] Module imports
- [x] Test framework (`test`, `assert_eq`)

---

## Near-Term Goals (v0.2)

### Language Improvements
- [ ] Generic types for structs and functions
- [ ] Operator overloading
- [ ] Better error messages with suggestions
- [ ] Compile-time constants and constant folding
- [ ] Named function arguments

### Standard Library Expansion
- [ ] Comprehensive string library (split, join, trim, etc.)
- [ ] JSON parsing and generation
- [ ] Regular expressions
- [ ] Date/time utilities
- [ ] Path/file utilities

### Tooling
- [ ] Language Server Protocol (LSP) support
- [ ] VS Code extension
- [ ] Package manager (`cortex pkg`)
- [ ] Build system integration (`cortex build`)

---

## Medium-Term Goals (v0.3)

### Advanced Features
- [ ] Interface types (duck typing)
- [ ] Reflection capabilities
- [ ] SIMD intrinsics
- [ ] Inline assembly support
- [ ] Memory profiling tools

### Ecosystem
- [ ] Package registry
- [ ] Documentation generator
- [ ] Formatter (`cortex fmt`)
- [ ] Linter (`cortex lint`)

### Platforms
- [ ] WebAssembly target
- [ ] ARM embedded support
- [ ] Cross-compilation

---

## Long-Term Vision (v1.0)

### Language Maturity
- [ ] Full generic type system
- [ ] Advanced type inference
- [ ] Formal language specification
- [ ] Comprehensive standard library

### Production Ready
- [ ] Stable API guarantees
- [ ] Performance benchmarks
- [ ] Security audit
- [ ] Enterprise support options

### Community
- [ ] Active contributor base
- [ ] Third-party package ecosystem
- [ ] Educational materials
- [ ] Conference presence

---

## Future Possibilities

Ideas being explored:

- **Gradual typing**: Mix static and dynamic typing
- **Effect system**: Track side effects in types
- **Linear types**: Ownership without borrow checking complexity
- **Metaprogramming**: Compile-time code generation
- **Distributed computing**: Built-in actor model

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to help make these goals a reality.

## Version History

| Version | Target Date | Focus |
|---------|-------------|-------|
| v0.1 | Current | Core language, basic stdlib |
| v0.2 | TBD | Generics, LSP, expanded stdlib |
| v0.3 | TBD | Interfaces, package ecosystem |
| v1.0 | TBD | Production ready |

---

*This roadmap is subject to change based on community feedback and development progress.*
