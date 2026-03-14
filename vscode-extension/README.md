# Cortex Language VSCode Extension

This extension provides syntax highlighting and language support for the Cortex programming language (`.cx` files).

## Features

- Syntax highlighting for `.cx` files
- Auto-closing brackets, quotes, and parentheses
- Comment toggling support
- Bracket matching
- Indentation rules

## Installation

### Automatic Installation (Windows)

Run the install script from the project root:

```powershell
.\scripts\install-vscode-extension.ps1
```

Or simply double-click `install-extension.bat`.

### Manual Installation

#### Method 1: Copy to Extensions Folder

1. Copy the `vscode-extension` folder to your VSCode extensions directory:
   - **Windows**: `%USERPROFILE%\.vscode\extensions\cortex-language`
   - **macOS**: `~/.vscode/extensions/cortex-language`
   - **Linux**: `~/.vscode/extensions/cortex-language`

2. Restart VSCode.

#### Method 2: Install as VSIX Package

1. If you have `vsce` installed:
   ```bash
   cd vscode-extension
   vsce package
   code --install-extension cortex-language-0.1.0.vsix
   ```

2. Or use the pre-built VSIX (if available):
   ```bash
   code --install-extension cortex-language-0.1.0.vsix
   ```

#### Method 3: Workspace Recommendation

For project-specific use, the extension files are already in `.vscode/` folder.
VSCode will automatically recognize the language configuration when opening this workspace.

## File Associations

The extension automatically associates `.cx` files with the Cortex language.

You can also manually set the language mode:
- Press `Ctrl+K M` (or `Cmd+K M` on macOS)
- Select "Cortex"

## Supported Syntax

- Keywords: `if`, `else`, `elif`, `for`, `while`, `do`, `repeat`, `until`, `switch`, `case`, `default`, `break`, `continue`, `return`, `fn`, `func`, `function`, `void`, `int`, `float`, `double`, `string`, `bool`, `char`, `var`, `const`, `let`, `struct`, `enum`, `import`, `export`, `module`, `extern`, `null`, `true`, `false`
- Types: `int`, `float`, `double`, `string`, `bool`, `char`, `void`, `any`, `dict`, `array`, `result`
- Comments: `//` line comments, `/* */` block comments
- Strings: `"double"`, `'single'`, `` `backtick` ``
- Numbers: integers, floats, hex (`0x`), binary (`0b`)

## License

MIT
