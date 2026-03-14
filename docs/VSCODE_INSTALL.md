# Cortex Language - VSCode Extension Installation

This document explains how to install the Cortex language extension for VSCode.

---

## Quick Install (Windows)

### Option 1: Double-click the batch file
Navigate to `scripts/` folder and double-click `install-extension.bat`

### Option 2: Run PowerShell script
```powershell
cd scripts
.\install-vscode-extension.ps1
```

### Option 3: Run from project root
```powershell
.\scripts\install-extension.bat
```

---

## Manual Installation

### Method 1: Copy to Extensions Folder

1. **Locate the extension source:**
   ```
   Simple C/vscode-extension/
   ```

2. **Copy the entire folder to VSCode extensions:**
   - **Windows**: Copy to `%USERPROFILE%\.vscode\extensions\cortex-language\`
   - **macOS**: Copy to `~/.vscode/extensions/cortex-language/`
   - **Linux**: Copy to `~/.vscode/extensions/cortex-language/`

3. **Restart VSCode**

### Method 2: Create Extension Manually

1. Create folder: `%USERPROFILE%\.vscode\extensions\cortex-language\`

2. Create `package.json` inside:
   ```json
   {
     "name": "cortex-language",
     "displayName": "Cortex Language Support",
     "version": "0.1.0",
     "engines": { "vscode": "^1.50.0" },
     "contributes": {
       "languages": [{
         "id": "cortex",
         "aliases": ["Cortex", "cx"],
         "extensions": [".cx"],
         "configuration": "./language-configuration.json"
       }],
       "grammars": [{
         "language": "cortex",
         "scopeName": "source.cortex",
         "path": "./syntaxes/cortex.tmLanguage.json"
       }]
     }
   }
   ```

3. Create `language-configuration.json`:
   ```json
   {
     "comments": { "lineComment": "//", "blockComment": ["/*", "*/"] },
     "brackets": [["{", "}"], ["[", "]"], ["(", ")"]],
     "autoClosingPairs": [["{", "}"], ["[", "]"], ["(", ")"], ["\"", "\""]]
   }
   ```

4. Create `syntaxes/cortex.tmLanguage.json` (copy from `vscode-extension/syntaxes/`)

5. **Restart VSCode**

---

## Verify Installation

1. Open VSCode
2. Open any `.cx` file
3. Check syntax highlighting is working
4. If not, press `Ctrl+K M` (or `Cmd+K M` on Mac) and select "Cortex"

---

## Workspace Mode (No Installation Needed)

The `.vscode/` folder in the project root already contains language configuration.
When you open this project as a workspace, VSCode automatically recognizes `.cx` files.

Files in `.vscode/`:
- `settings.json` - File associations
- `extensions.json` - Extension recommendations
- `language-configuration.json` - Language settings
- `syntaxes/cortex.tmLanguage.json` - Syntax grammar

---

## Troubleshooting

**Syntax highlighting not working?**
1. Press `Ctrl+Shift+P`
2. Type "Change Language Mode"
3. Select "Cortex"

**Extension not appearing?**
1. Check the folder is in the correct location
2. Ensure `package.json` is valid JSON
3. Restart VSCode completely

**Want to uninstall?**
Delete the `cortex-language` folder from your extensions directory.
