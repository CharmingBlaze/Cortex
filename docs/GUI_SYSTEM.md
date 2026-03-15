# Cortex GUI System Documentation

## Overview

The Cortex GUI System provides a complete, cross-platform graphical user interface framework powered by [GTK4](https://gtk.org/), a modern native C GUI toolkit. This system allows Cortex developers to create native desktop applications with a simple, easy-to-use API.

## Philosophy

- **Simple like BASIC**: Easy-to-learn commands for beginners
- **Familiar like C**: Syntax that C programmers recognize
- **Powerful**: Full access to modern GUI capabilities
- **Cross-platform**: Works on Windows, macOS, and Linux
- **Pointer-free**: All GUI objects use opaque handles for safety
- **Native**: Pure C implementation with GTK4 - no CGO or Go runtime

## Quick Start

```c
#include <gui_runtime.h>

void on_click(gui_event e) {
    gui_alert_info("Button clicked!");
}

fn main() {
    // Create window - one line!
    gui_start("My App", 800, 600);
    
    // Add widgets - super simple
    gui_add_label("Hello, World!");
    gui_add_button("Click Me", on_click);
    
    // Run the app
    gui_run();
}
```

## API Reference

### Application Lifecycle

#### `void gui_start(const char* title, int width, int height)`
Creates the main window and initializes GTK. One-liner to start your app.

```c
gui_start("My Application", 800, 600);
```

#### `void gui_run(void)`
Starts the GUI event loop. This function blocks until the application exits.

```c
gui_run();
```

#### `void gui_quit(void)`
Quits the GUI application.

```c
gui_quit();
```

### Main Window

#### `void gui_add(gui_widget w)`
Adds a widget to the main window (auto-creates vbox if needed).

```c
gui_add(gui_label("Hello"));
```

#### `void gui_add_to(gui_container c, gui_widget w)`
Adds a widget to a specific container.

```c
gui_add_to(vbox, gui_label("Hello"));
```

#### `void gui_set_title(const char* title)`
Changes the main window title.

```c
gui_set_title("New Title");
```

#### `void gui_set_size(int width, int height)`
Changes the main window size.

```c
gui_set_size(1024, 768);
```

### Widgets - Basic

#### `gui_widget gui_label(const char* text)`
Creates a text label.

```c
gui_widget label = gui_label("Hello, World!");
```

#### `void gui_set_text(gui_widget w, const char* text)`
Updates text on labels, entries, or buttons.

```c
gui_set_text(label, "Updated text");
```

#### `char* gui_get_text(gui_widget w)`
Gets text from labels, entries, or buttons. Caller must free with `gui_free()`.

```c
char* text = gui_get_text(entry);
println(text);
gui_free(text);
```

#### `gui_widget gui_button(const char* text, gui_callback on_click)`
Creates a clickable button.

```c
gui_widget btn = gui_button("Click Me", [](gui_event e) {
    println("Clicked!");
});
```

#### `gui_widget gui_button_ok(const char* text, gui_callback on_click)`
Creates a primary/suggested-action button.

```c
gui_widget btn = gui_button_ok("OK", on_ok);
```

### Widgets - Input Controls

#### `gui_widget gui_entry(const char* placeholder)`
Creates a single-line text input.

```c
gui_widget entry = gui_entry("Enter name...");
```

#### `gui_widget gui_entry_secret(const char* placeholder)`
Creates a password entry field.

```c
gui_widget password = gui_entry_secret("Password");
```

#### `gui_widget gui_entry_multi(const char* placeholder)`
Creates a multi-line text area.

```c
gui_widget area = gui_entry_multi("Enter long text...");
```

#### `gui_widget gui_check(const char* label)`
Creates a checkbox.

```c
gui_widget check = gui_check("Enable feature");
```

#### `bool gui_is_checked(gui_widget w)`
Gets the checkbox state.

```c
bool checked = gui_is_checked(check);
```

#### `void gui_set_checked(gui_widget w, bool checked)`
Sets the checkbox state.

```c
gui_set_checked(check, true);
```

#### `gui_widget gui_select(const char* options[], int count)`
Creates a dropdown/select.

```c
const char* options[] = {"Red", "Green", "Blue"};
gui_widget select = gui_select(options, 3);
```

#### `int gui_get_selected(gui_widget w)`
Gets the selected index.

```c
int index = gui_get_selected(select);
```

#### `gui_widget gui_slider(double min, double max)`
Creates a slider.

```c
gui_widget slider = gui_slider(0, 100);
```

#### `double gui_get_value(gui_widget w)`
Gets slider or progress value.

```c
double val = gui_get_value(slider);
```

#### `void gui_set_value(gui_widget w, double value)`
Sets slider or progress value.

```c
gui_set_value(slider, 75.0);
```

#### `gui_widget gui_progress(void)`
Creates a progress bar.

```c
gui_widget progress = gui_progress();
gui_set_value(progress, 0.5);  // 50%
```

### Visual Widgets

#### `gui_widget gui_separator(void)`
Creates a horizontal separator.

```c
gui_add(gui_separator());
```

#### `gui_widget gui_image(const char* path)`
Creates an image from a file.

```c
gui_widget img = gui_image("photo.png");
```

#### `gui_widget gui_spinner(void)`
Creates a loading spinner.

```c
gui_widget spinner = gui_spinner();
gui_spinner_start(spinner);
// ... load something ...
gui_spinner_stop(spinner);
```

### Layout Containers

#### `gui_container gui_vbox(void)`
Creates a vertical box layout.

```c
gui_container vbox = gui_vbox();
gui_add_to(vbox, gui_label("First"));
gui_add_to(vbox, gui_label("Second"));
```

#### `gui_container gui_hbox(void)`
Creates a horizontal box layout.

```c
gui_container hbox = gui_hbox();
gui_add_to(hbox, gui_button("A", NULL));
gui_add_to(hbox, gui_button("B", NULL));
```

#### `gui_container gui_grid(int columns)`
Creates a grid layout.

```c
gui_container grid = gui_grid(3);  // 3 columns
for (int i = 0; i < 9; i++) {
    gui_add_to(grid, gui_button("Btn", NULL));
}
```

#### `gui_container gui_scroll(gui_widget content)`
Creates a scrollable area.

```c
gui_container scroll = gui_scroll(content);
```

#### `gui_container gui_tabs(void)`
Creates a tab notebook.

```c
gui_container tabs = gui_tabs();
gui_tab_add(tabs, "Tab 1", gui_label("Content 1"));
gui_tab_add(tabs, "Tab 2", gui_label("Content 2"));
```

### Dialogs

#### `void gui_alert_info(const char* message)`
Shows an information dialog.

```c
gui_alert_info("Operation completed!");
```

#### `void gui_alert_error(const char* message)`
Shows an error dialog.

```c
gui_alert_error("Failed to save file!");
```

#### `void gui_alert_warn(const char* message)`
Shows a warning dialog.

```c
gui_alert_warn("This action cannot be undone.");
```

#### `void gui_confirm(const char* message, gui_callback on_result)`
Shows a confirmation dialog.

```c
gui_confirm("Delete this file?", [](gui_event e) {
    if (e.checked) {
        println("User confirmed");
    }
});
```

#### `void gui_pick_file(gui_callback on_result)`
Shows a file open dialog.

```c
gui_pick_file([](gui_event e) {
    if (e.text) {
        println("Selected: %s", e.text);
    }
});
```

#### `void gui_save_file(const char* default_name, gui_callback on_result)`
Shows a file save dialog.

```c
gui_save_file("untitled.txt", [](gui_event e) {
    if (e.text) {
        // Save file...
    }
});
```

#### `void gui_pick_folder(gui_callback on_result)`
Shows a folder select dialog.

```c
gui_pick_folder([](gui_event e) {
    if (e.text) {
        println("Folder: %s", e.text);
    }
});
```

### Widget State

#### `void gui_show(gui_widget w)` / `void gui_hide(gui_widget w)`
Shows or hides a widget.

```c
gui_show(widget);
gui_hide(widget);
```

#### `void gui_enable(gui_widget w)` / `void gui_disable(gui_widget w)`
Enables or disables a widget.

```c
gui_enable(button);
gui_disable(button);
```

#### `void gui_focus(gui_widget w)`
Gives keyboard focus to a widget.

```c
gui_focus(entry);
```

### Utility

#### `void gui_free(char* str)`
Frees a string returned by GUI functions.

```c
char* text = gui_get_text(entry);
gui_free(text);
```

#### `char* gui_clipboard_get(void)` / `void gui_clipboard_set(const char* text)`
Clipboard access.

```c
gui_clipboard_set("Copied text");
```

## Event System

Events are handled through callbacks that receive a `gui_event` structure:

```c
typedef struct {
    int type;           // Event type
    int64_t source;     // Widget that triggered
    double value;       // Numeric value (slider, progress)
    char* text;         // Text value (entry, file dialog)
    bool checked;       // Boolean value (checkbox, confirm)
} gui_event;
```

Event types:
- `GUI_CLICK` - Button click
- `GUI_CHANGE` - Value change
- `GUI_SELECT` - Selection change
- `GUI_CHECK` - Checkbox toggle
- `GUI_SUBMIT` - Form submit

## Convenience Macros

One-liners for common patterns:

```c
gui_add_label("Hello");        // Create and add label
gui_add_button("Click", cb);   // Create and add button
gui_add_entry("Type...");      // Create and add entry
gui_add_check("Option");       // Create and add checkbox
gui_add_separator();           // Add separator
gui_add_progress();            // Add progress bar
```

## Complete Examples

See the `examples/gui/` directory for complete working examples.

## Building GUI Applications

GUI applications are built the same way as regular Cortex programs:

```bash
cortex -i myapp.cx -o myapp
./myapp
```

The GUI runtime is automatically linked when GUI functions are detected.

## Architecture

The GUI system consists of two layers:

1. **Cortex Application** - Your code using the `gui_runtime.h` API
2. **GTK4 Backend** - `internal/gui_gtk4/` provides the native implementation

```
Cortex Code
    ↓
C Runtime (gui_runtime.h)
    ↓
GTK4 Backend (internal/gui_gtk4/)
    ↓
Native Platform (Win/Mac/Linux)
```

## Troubleshooting

### Window doesn't appear
Make sure to call `gui_run()` after adding widgets.

### Events not firing
Check that you're passing valid callback functions.

### Memory issues
Always use `gui_free()` for strings returned by GUI functions.

### GTK4 not found

**Windows:**
- For development: Install MSYS2 from https://msys2.org, then run `pacman -S mingw-w64-x86_64-gtk4`
- For distribution: Use `bundle_gtk.ps1` to create a portable bundle - no installation required on target machines

**Linux:** `sudo apt install libgtk-4-dev`

**macOS:** `brew install gtk4`

## Windows Distribution

To distribute your GUI app on Windows without requiring users to install anything:

```bash
# Compile your app
./cortex -i myapp.cx -o myapp.exe

# Bundle GTK4 (creates dist/ folder with everything)
powershell -ExecutionPolicy Bypass -File bundle_gtk.ps1 -AppExe myapp.exe

# The dist/ folder (47 MB) is portable - copy to any Windows machine!
```

## License

The Cortex GUI System is part of the Cortex compiler project.
