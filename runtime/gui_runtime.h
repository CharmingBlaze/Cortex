// gui_runtime.h - Cortex GUI Runtime API
//
// This header provides the C API that Cortex-compiled code uses to create
// and manage GUI applications. All functions use opaque handles (int64) to
// reference GUI objects, keeping Cortex code pointer-safe.
//
// Example Cortex code:
//   window w = gui_window("Hello", 800, 600);
//   gui_label(w, "Hello World");
//   gui_button(w, "Click Me", [](event e) {
//       println("Clicked!");
//   });
//   gui_run();

#ifndef CORTEX_GUI_RUNTIME_H
#define CORTEX_GUI_RUNTIME_H

#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

// ============================================================================
// Opaque Handle Types
// ============================================================================

typedef int64_t gui_window;
typedef int64_t gui_widget;
typedef int64_t gui_container;
typedef int64_t gui_callback_id;

#define GUI_INVALID_HANDLE 0

// ============================================================================
// Event System
// ============================================================================

typedef enum {
    GUI_EVENT_CLICK = 0,
    GUI_EVENT_CHANGE = 1,
    GUI_EVENT_KEY = 2,
    GUI_EVENT_MOUSE = 3,
    GUI_EVENT_WINDOW = 4,
    GUI_EVENT_CUSTOM = 5
} gui_event_type;

typedef struct {
    gui_event_type type;
    int64_t source;
    void* data;  // Event-specific data (string, bool, number, etc.)
} gui_event;

// Callback type for event handlers
typedef void (*gui_event_callback)(gui_event event);

// ============================================================================
// Application Lifecycle
// ============================================================================

// Initialize the GUI system (called automatically by gui_window)
void gui_init(void);

// Run the GUI event loop (blocks until application exits)
void gui_run(void);

// Quit the GUI application
void gui_quit(void);

// Check if GUI is running
bool gui_is_running(void);

// Get GUI API version
int gui_version(void);

// ============================================================================
// Window Management
// ============================================================================

// Create a new window with title and size
gui_window gui_window_create(const char* title, int width, int height);

// Show a window
void gui_window_show(gui_window window);

// Hide a window
void gui_window_hide(gui_window window);

// Close and destroy a window
void gui_window_close(gui_window window);

// Set window title
void gui_window_set_title(gui_window window, const char* title);

// Center window on screen
void gui_window_center(gui_window window);

// Set fixed size (prevent resizing)
void gui_window_set_fixed_size(gui_window window, bool fixed);

// Set fullscreen mode
void gui_window_set_fullscreen(gui_window window, bool fullscreen);

// Set the main content of a window
void gui_window_set_content(gui_window window, gui_container content);

// ============================================================================
// Widget Creation - Basic
// ============================================================================

// Create a label widget
gui_widget gui_label_create(const char* text);

// Update label text
void gui_label_set_text(gui_widget label, const char* text);

// Create a button with callback
gui_widget gui_button_create(const char* label, gui_event_callback on_click);

// Create a text entry field
gui_widget gui_entry_create(const char* placeholder, gui_event_callback on_change);

// Get entry text (caller must free result)
char* gui_entry_get_text(gui_widget entry);

// Set entry text
void gui_entry_set_text(gui_widget entry, const char* text);

// Create a multi-line text area
gui_widget gui_textarea_create(const char* placeholder, gui_event_callback on_change);

// Get textarea text (caller must free result)
char* gui_textarea_get_text(gui_widget textarea);

// Set textarea text
void gui_textarea_set_text(gui_widget textarea, const char* text);

// ============================================================================
// Widget Creation - Input Controls
// ============================================================================

// Create a checkbox
gui_widget gui_checkbox_create(const char* label, gui_event_callback on_change);

// Get checkbox state
bool gui_checkbox_get_state(gui_widget checkbox);

// Set checkbox state
void gui_checkbox_set_state(gui_widget checkbox, bool checked);

// Create a slider
gui_widget gui_slider_create(double min, double max, double value, gui_event_callback on_change);

// Get slider value
double gui_slider_get_value(gui_widget slider);

// Set slider value
void gui_slider_set_value(gui_widget slider, double value);

// Create a progress bar
gui_widget gui_progress_create(void);

// Set progress (0.0 to 1.0)
void gui_progress_set_value(gui_widget progress, double value);

// ============================================================================
// Widget Creation - Graphics
// ============================================================================

// Create an image from file
gui_widget gui_image_create(const char* filepath);

// Image fill modes
#define GUI_IMAGE_FILL_ORIGINAL 0
#define GUI_IMAGE_FILL_STRETCH 1
#define GUI_IMAGE_FILL_CONTAIN 2

// Set image fill mode
void gui_image_set_fill(gui_widget image, int fill_mode);

// Create a colored rectangle
gui_widget gui_rectangle_create(uint8_t r, uint8_t g, uint8_t b, uint8_t a);

// Create a colored circle
gui_widget gui_circle_create(uint8_t r, uint8_t g, uint8_t b, uint8_t a);

// Create a line
gui_widget gui_line_create(float x1, float y1, float x2, float y2);

// Set line color
void gui_line_set_color(gui_widget line, uint8_t r, uint8_t g, uint8_t b, uint8_t a);

// ============================================================================
// Layout Containers
// ============================================================================

// Create vertical box container
gui_container gui_vbox_create(void);

// Create horizontal box container
gui_container gui_hbox_create(void);

// Create grid container with columns
gui_container gui_grid_create(int columns);

// Add widget to container
void gui_container_add(gui_container container, gui_widget widget);

// Add multiple widgets to container (variadic helper)
void gui_container_add_many(gui_container container, int count, ...);

// ============================================================================
// Advanced Layout
// ============================================================================

// Border layout positions
#define GUI_BORDER_TOP 0
#define GUI_BORDER_BOTTOM 1
#define GUI_BORDER_LEFT 2
#define GUI_BORDER_RIGHT 3
#define GUI_BORDER_CENTER 4

// Create border layout container
gui_container gui_border_create(
    gui_widget top,
    gui_widget bottom,
    gui_widget left,
    gui_widget right,
    gui_widget center
);

// Set padding for a container
void gui_container_set_padding(gui_container container, int padding);

// ============================================================================
// Dialogs
// ============================================================================

// Show information dialog
void gui_dialog_info(gui_window parent, const char* title, const char* message);

// Show error dialog
void gui_dialog_error(gui_window parent, const char* title, const char* message);

// Show confirmation dialog with callback
void gui_dialog_confirm(
    gui_window parent,
    const char* title,
    const char* message,
    gui_event_callback on_result
);

// Show file open dialog with callback
// Callback receives filepath string (empty if cancelled)
void gui_dialog_file_open(gui_window parent, gui_event_callback on_result);

// Show file save dialog with callback
// Callback receives filepath string (empty if cancelled)
void gui_dialog_file_save(gui_window parent, gui_event_callback on_result);

// ============================================================================
// Widget Management
// ============================================================================

// Refresh/redraw a widget
void gui_refresh(gui_widget widget);

// Resize a widget
void gui_resize(gui_widget widget, float width, float height);

// Move a widget to position
void gui_move(gui_widget widget, float x, float y);

// Enable a widget
void gui_enable(gui_widget widget);

// Disable a widget
void gui_disable(gui_widget widget);

// Check if widget is enabled
bool gui_is_enabled(gui_widget widget);

// Hide a widget
void gui_hide(gui_widget widget);

// Show a widget
void gui_show(gui_widget widget);

// ============================================================================
// Utility Functions
// ============================================================================

// Free a string returned by GUI functions
void gui_free_string(char* str);

// Convert event data to string (helper)
const char* gui_event_get_string(gui_event event);

// Convert event data to bool (helper)
bool gui_event_get_bool(gui_event event);

// Convert event data to double (helper)
double gui_event_get_double(gui_event event);

// ============================================================================
// Convenience Macros for Cortex
// ============================================================================

// Simplified window creation
#define gui_window(title, w, h) gui_window_create(title, w, h)

// Simplified label creation
#define gui_label(parent, text) gui_container_add(parent, gui_label_create(text))

// Simplified button creation
#define gui_button(parent, label, callback) gui_container_add(parent, gui_button_create(label, callback))

// Simplified entry creation
#define gui_entry(parent, placeholder, callback) gui_container_add(parent, gui_entry_create(placeholder, callback))

#ifdef __cplusplus
}
#endif

#endif // CORTEX_GUI_RUNTIME_H
