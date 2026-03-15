# Contributing to Cortex

Thank you for your interest in contributing to Cortex! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)

## Code of Conduct

Be respectful, inclusive, and constructive. We welcome contributions from everyone.

## How Can I Contribute?

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/CharmingBlaze/Cortex/issues)
2. If not, create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - Code samples if applicable

### Suggesting Features

1. Open an issue with the label `enhancement`
2. Describe the feature and why it would be useful
3. Provide examples of how it would be used

### Contributing Code

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.21 or later
- GCC (MinGW on Windows, clang/gcc on macOS/Linux)
- Git

### Building from Source

```bash
git clone https://github.com/CharmingBlaze/Cortex.git
cd Cortex
go build -o cortex ./cmd/cortex
```

### Running Tests

```bash
# Run the test suite
go test ./...

# Run a specific Cortex file
./cortex -i examples/hello.cx -run
```

### Project Structure

```
Cortex/
├── cmd/cortex/          # CLI entry point
├── internal/
│   ├── ast/             # Abstract Syntax Tree definitions
│   ├── compiler/        # Lexer, Parser, Semantic Analysis, Code Generation
│   ├── config/          # Configuration handling
│   └── errors/          # Error reporting
├── runtime/             # C runtime libraries (async, threads, GUI, etc.)
├── examples/            # Example Cortex programs
├── LANGUAGE_GUIDE.md    # User documentation
└── LANGUAGE_SPEC.md     # Language specification
```

## Pull Request Process

1. **Update documentation** if you change behavior
2. **Add tests** for new features
3. **Follow existing code style** (see Coding Standards)
4. **Write clear commit messages**

### Commit Message Format

```
type: brief description

- Detailed explanation if needed
- Reference issues: #123

Types: feat, fix, docs, style, refactor, test, chore
```

## Coding Standards

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Add comments for exported functions
- Handle errors explicitly

### C Runtime Code

- Use C99 standard
- Include header guards
- Document public functions
- Handle NULL pointers gracefully

### Documentation

- Keep LANGUAGE_GUIDE.md user-focused
- Update README.md for major features
- Use code examples liberally

## Questions?

Open an issue with the `question` label or reach out to the maintainers.

---

Thank you for helping make Cortex better!
