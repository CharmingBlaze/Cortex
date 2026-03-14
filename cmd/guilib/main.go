// cmd/guilib/main.go - Build the GUI shared library
//
// This builds the Fyne-based GUI runtime as a DLL/shared library
// that can be linked with Cortex-compiled programs.
package main

import _ "cortex/internal/gui_fyne"

func main() {
	// This package exists only to build the shared library.
	// All functionality is in the gui_fyne package.
	// Build with: go build -buildmode=c-shared -o cortex_gui.dll ./cmd/guilib
}
