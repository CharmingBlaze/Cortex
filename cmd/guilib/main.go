// cmd/guilib/main.go - GUI Library Build Command
//
// The Cortex GUI is now a native GTK4 C library (internal/gui_gtk4).
// This command is kept for compatibility but the GUI is built via Makefile.
//
// Build the GTK4 GUI library:
//
//	cd internal/gui_gtk4 && make
package main

func main() {
	// GUI library is now built from internal/gui_gtk4/Makefile
	// Run: make -C internal/gui_gtk4
}
