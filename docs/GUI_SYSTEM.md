# Cortex GUI System Documentation

## Overview

The Cortex GUI System provides a complete, cross-platform graphical user interface framework powered by [Fyne](https://fyne.io/), a modern Go GUI toolkit. This system allows Cortex developers to create native desktop applications with a simple, BASIC-like API.

## Philosophy

- **Simple like BASIC**: Easy-to-learn commands for beginners
- **Familiar like C**: Syntax that C programmers recognize
- **Powerful**: Full access to modern GUI capabilities
- **Cross-platform**: Works on Windows, macOS, and Linux
- **Pointer-free**: All GUI objects use opaque handles for safety

## Quick Start

```c
#include <gui_runtime.h>

void main() {
    // Create a window
    window w = gui_window("My App", 800, 600);
    
    // Add a label
    gui_label(w, "Hello, World!");
    
    // Add a button
    gui_button(w, "Click Me", [](event e) {
        println("Button clicked!");
    });
    
    // Show and run
    gui_window_show(w);
    gui_run();
}
```

## API Reference

### Application Lifecycle

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

#### `int gui_version(void)`
Returns the GUI API version number.

```c
int version = gui_version();
```

### Window Management

#### `gui_window gui_window_create(const char* title, int width, int height)`
Creates a new window with the specified title and size.

```c
gui_window w = gui_window_create("My Application", 800, 600);
```

#### `void gui_window_show(gui_window window)`
Makes a window visible.

```c
gui_window_show(w);
```

#### `void gui_window_hide(gui_window window)`
Hides a window.

```c
gui_window_hide(w);
```

#### `void gui_window_close(gui_window window)`
Closes and destroys a window.

```c
gui_window_close(w);
```

#### `void gui_window_set_title(gui_window window, const char* title)`
Changes the window title.

```c
gui_window_set_title(w, "New Title");
```

#### `void gui_window_center(gui_window window)`
Centers the window on the screen.

```c
gui_window_center(w);
```

#### `void gui_window_set_fixed_size(gui_window window, bool fixed)`
Prevents or allows window resizing.

```c
gui_window_set_fixed_size(w, true);  // Fixed size
gui_window_set_fixed_size(w, false); // Resizable
```

#### `void gui_window_set_fullscreen(gui_window window, bool fullscreen)`
Toggles fullscreen mode.

```c
gui_window_set_fullscreen(w, true);
```

### Widgets - Basic

#### `gui_widget gui_label_create(const char* text)`
Creates a text label.

```c
gui_widget label = gui_label_create("Hello, World!");
```

#### `void gui_label_set_text(gui_widget label, const char* text)`
Updates label text.

```c
gui_label_set_text(label, "Updated text");
```

#### `gui_widget gui_button_create(const char* label, gui_event_callback on_click)`
Creates a clickable button.

```c
gui_widget btn = gui_button_create("Click Me", [](event e) {
    println("Clicked!");
});
```

#### `gui_widget gui_entry_create(const char* placeholder, gui_event_callback on_change)`
Creates a single-line text input field.

```c
gui_widget entry = gui_entry_create("Enter name...", [](event e) {
    const char* text = gui_event_get_string(e);
    println("Text changed: " + text);
});
```

#### `char* gui_entry_get_text(gui_widget entry)`
Gets the current text from an entry field. Caller must free the result.

```c
char* text = gui_entry_get_text(entry);
println(text);
gui_free_string(text);
```

#### `void gui_entry_set_text(gui_widget entry, const char* text)`
Sets the text in an entry field.

```c
gui_entry_set_text(entry, "New text");
```

#### `gui_widget gui_textarea_create(const char* placeholder, gui_event_callback on_change)`
Creates a multi-line text area.

```c
gui_widget area = gui_textarea_create("Enter long text...", [](event e) {
    const char* text = gui_event_get_string(e);
    println(text);
});
```

### Widgets - Input Controls

#### `gui_widget gui_checkbox_create(const char* label, gui_event_callback on_change)`
Creates a checkbox.

```c
gui_widget check = gui_checkbox_create("Enable feature", [](event e) {
    bool checked = gui_event_get_bool(e);
    println(checked ? "Enabled" : "Disabled");
});
```

#### `bool gui_checkbox_get_state(gui_widget checkbox)`
Gets the checkbox state.

```c
bool is_checked = gui_checkbox_get_state(check);
```

#### `void gui_checkbox_set_state(gui_widget checkbox, bool checked)`
Sets the checkbox state.

```c
gui_checkbox_set_state(check, true);
```

#### `gui_widget gui_slider_create(double min, double max, double value, gui_event_callback on_change)`
Creates a slider control.

```c
gui_widget slider = gui_slider_create(0.0, 100.0, 50.0, [](event e) {
    double value = gui_event_get_double(e);
    println("Value: " + value);
});
```

#### `double gui_slider_get_value(gui_widget slider)`
Gets the current slider value.

```c
double val = gui_slider_get_value(slider);
```

#### `void gui_slider_set_value(gui_widget slider, double value)`
Sets the slider value.

```c
gui_slider_set_value(slider, 75.0);
```

#### `gui_widget gui_progress_create(void)`
Creates a progress bar.

```c
gui_widget progress = gui_progress_create();
```

#### `void gui_progress_set_value(gui_widget progress, double value)`
Sets the progress value (0.0 to 1.0).

```c
gui_progress_set_value(progress, 0.5);  // 50%
```

### Widgets - Graphics

#### `gui_widget gui_image_create(const char* filepath)`
Creates an image from a file.

```c
gui_widget img = gui_image_create("photo.png");
```

#### `void gui_image_set_fill(gui_widget image, int fill_mode)`
Sets how the image fills its space.

```c
gui_image_set_fill(img, GUI_IMAGE_FILL_CONTAIN);
```

Fill modes:
- `GUI_IMAGE_FILL_ORIGINAL` - Original size
- `GUI_IMAGE_FILL_STRETCH` - Stretch to fit
- `GUI_IMAGE_FILL_CONTAIN` - Scale to fit, preserve aspect ratio

#### `gui_widget gui_rectangle_create(uint8_t r, uint8_t g, uint8_t b, uint8_t a)`
Creates a colored rectangle.

```c
gui_widget rect = gui_rectangle_create(255, 0, 0, 255);  // Red, fully opaque
```

#### `gui_widget gui_circle_create(uint8_t r, uint8_t g, uint8_t b, uint8_t a)`
Creates a colored circle.

```c
gui_widget circle = gui_circle_create(0, 255, 0, 128);  // Green, semi-transparent
```

#### `gui_widget gui_line_create(float x1, float y1, float x2, float y2)`
Creates a line.

```c
gui_widget line = gui_line_create(0, 0, 100, 100);
```

#### `void gui_line_set_color(gui_widget line, uint8_t r, uint8_t g, uint8_t b, uint8_t a)`
Sets the line color.

```c
gui_line_set_color(line, 0, 0, 255, 255);  // Blue
```

### Layout Containers

#### `gui_container gui_vbox_create(void)`
Creates a vertical box layout.

```c
gui_container vbox = gui_vbox_create();
gui_container_add(vbox, label);
gui_container_add(vbox, button);
gui_container_add(vbox, entry);
```

#### `gui_container gui_hbox_create(void)`
Creates a horizontal box layout.

```c
gui_container hbox = gui_hbox_create();
gui_container_add(hbox, btn1);
gui_container_add(hbox, btn2);
gui_container_add(hbox, btn3);
```

#### `gui_container gui_grid_create(int columns)`
Creates a grid layout with specified columns.

```c
gui_container grid = gui_grid_create(3);  // 3 columns
for (int i = 0; i < 9; i++) {
    gui_container_add(grid, gui_button_create("Btn", NULL));
}
```

#### `void gui_container_add(gui_container container, gui_widget widget)`
Adds a widget to a container.

```c
gui_container_add(vbox, my_widget);
```

### Dialogs

#### `void gui_dialog_info(gui_window parent, const char* title, const char* message)`
Shows an information dialog.

```c
gui_dialog_info(w, "Info", "Operation completed successfully!");
```

#### `void gui_dialog_error(gui_window parent, const char* title, const char* message)`
Shows an error dialog.

```c
gui_dialog_error(w, "Error", "Failed to save file!");
```

#### `void gui_dialog_confirm(gui_window parent, const char* title, const char* message, gui_event_callback on_result)`
Shows a confirmation dialog.

```c
gui_dialog_confirm(w, "Confirm", "Delete this file?", [](event e) {
    if (gui_event_get_bool(e)) {
        println("User confirmed");
    } else {
        println("User cancelled");
    }
});
```

#### `void gui_dialog_file_open(gui_window parent, gui_event_callback on_result)`
Shows a file open dialog.

```c
gui_dialog_file_open(w, [](event e) {
    const char* path = gui_event_get_string(e);
    if (strlen(path) > 0) {
        println("Selected: " + path);
    }
});
```

#### `void gui_dialog_file_save(gui_window parent, gui_event_callback on_result)`
Shows a file save dialog.

```c
gui_dialog_file_save(w, [](event e) {
    const char* path = gui_event_get_string(e);
    // Save file...
});
```

### Widget Management

#### `void gui_refresh(gui_widget widget)`
Requests a redraw of a widget.

```c
gui_refresh(label);
```

#### `void gui_resize(gui_widget widget, float width, float height)`
Changes the size of a widget.

```c
gui_resize(button, 200, 50);
```

#### `void gui_move(gui_widget widget, float x, float y)`
Moves a widget to a position.

```c
gui_move(label, 100, 50);
```

#### `void gui_enable(gui_widget widget)`
Enables a widget.

```c
gui_enable(button);
```

#### `void gui_disable(gui_widget widget)`
Disables a widget.

```c
gui_disable(button);
```

#### `bool gui_is_enabled(gui_widget widget)`
Checks if a widget is enabled.

```c
if (gui_is_enabled(button)) {
    // ...
}
```

### Utility Functions

#### `void gui_free_string(char* str)`
Frees a string returned by GUI functions.

```c
char* text = gui_entry_get_text(entry);
// ... use text ...
gui_free_string(text);
```

#### `const char* gui_event_get_string(gui_event event)`
Extracts string data from an event.

```c
const char* text = gui_event_get_string(e);
```

#### `bool gui_event_get_bool(gui_event event)`
Extracts boolean data from an event.

```c
bool checked = gui_event_get_bool(e);
```

#### `double gui_event_get_double(gui_event event)`
Extracts numeric data from an event.

```c
double value = gui_event_get_double(e);
```

## Event System

Events are handled through callbacks that receive a `gui_event` structure:

```c
typedef struct {
    gui_event_type type;
    int64_t source;
    void* data;
} gui_event;
```

Event types:
- `GUI_EVENT_CLICK` - Button click
- `GUI_EVENT_CHANGE` - Value change
- `GUI_EVENT_KEY` - Keyboard input
- `GUI_EVENT_MOUSE` - Mouse action
- `GUI_EVENT_WINDOW` - Window event
- `GUI_EVENT_CUSTOM` - Custom event

## Complete Examples

See the `examples/gui/` directory for complete working examples.

## Building GUI Applications

GUI applications are built the same way as regular Cortex programs:

```bash
cortex -i myapp.cx -o myapp
cortex build -run
```

The GUI runtime is automatically linked when GUI functions are detected.

## Architecture

The GUI system consists of three layers:

1. **Cortex Application** - Your code using the `gui_runtime.h` API
2. **C Runtime** - `gui_runtime.c` provides the C implementation
3. **Go/Fyne Bridge** - `internal/gui_fyne/` connects to the Fyne toolkit

```
Cortex Code
    ↓
C Runtime (gui_runtime.h/c)
    ↓
Go Bridge (internal/gui_fyne/)
    ↓
Fyne Toolkit
    ↓
Native Platform (Win/Mac/Linux)
```

## Troubleshooting

### Window doesn't appear
Make sure to call `gui_window_show()` before `gui_run()`.

### Events not firing
Check that you're passing valid callback functions to widget creation functions.

### Memory issues
Always use `gui_free_string()` for strings returned by GUI functions.

### Compilation errors
Ensure you're including `<gui_runtime.h>` and linking with the GUI runtime.

## License

The Cortex GUI System is part of the Cortex compiler project.
