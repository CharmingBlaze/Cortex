// gui_runtime.h - Cortex GUI Runtime API
//
// A simple, intuitive GUI library. Build apps in minutes, not hours.
//
// Quick Example:
//   gui_start("My App", 800, 600);           // Create window
//   gui_add(gui_label("Hello, World!"));     // Add a label
//   gui_add(gui_button("Click", on_click));  // Add a button
//   gui_run();                                // Show and run
//
// That's it! No complex setup, no boilerplate.
//
// All widgets are automatically added to the main window's content area.
// Use containers (vbox, hbox, grid) for layout control.

#ifndef CORTEX_GUI_RUNTIME_H
#define CORTEX_GUI_RUNTIME_H

#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

// ============================================================================
// Types - Simple opaque handles
// ============================================================================

typedef int64_t gui_window;
typedef int64_t gui_widget;
typedef int64_t gui_container;

#define GUI_NULL 0
#define GUI_INVALID_HANDLE (-1)

// ============================================================================
// Events - Callback signature
// ============================================================================

// Event types
typedef enum {
    GUI_CLICK = 0,
    GUI_CHANGE = 1,
    GUI_SELECT = 2,
    GUI_CHECK = 3,
    GUI_SUBMIT = 4
} gui_event_type;

// Aliases for compatibility
#define GUI_EVENT_CLICK GUI_CLICK
#define GUI_EVENT_CHANGE GUI_CHANGE
#define GUI_EVENT_SELECT GUI_SELECT
#define GUI_EVENT_CHECK GUI_CHECK
#define GUI_EVENT_SUBMIT GUI_SUBMIT

typedef struct {
    gui_event_type type;  // Event type (click, change, etc.)
    int64_t source;       // Widget that triggered the event
    double value;         // Numeric value (slider, check)
    char* text;           // Text value (entry, select)
    bool checked;         // Boolean value (checkbox)
    void* data;           // Generic data pointer
    // Additional fields for GTK4 compatibility
    int key;              // Key code for key events
    float x, y;           // Mouse coordinates
    bool bool_val;        // Boolean value alias
} gui_event;

typedef void (*gui_callback)(gui_event e);
typedef void (*gui_event_callback)(gui_event e);  // Alias for compatibility

// ============================================================================
// Application Lifecycle - Super Simple
// ============================================================================

// One-liner to create window and start GUI
void gui_start(const char* title, int width, int height);

// Standard lifecycle
void gui_init(void);
void gui_run(void);      // Blocking - use for GUI-only apps
void gui_quit(void);

// Non-blocking mode - for integration with raylib/SDL/OpenGL
void gui_run_nonblock(void);   // Start without blocking
void gui_update(void);         // Process events (call in your main loop)
bool gui_is_running(void);     // Check if GUI is active

// ============================================================================
// Main Window - Auto-managed
// ============================================================================

// Add widget to main window (auto-creates vbox if needed)
void gui_add(gui_widget w);

// Add widget to specific container
void gui_add_to(gui_container c, gui_widget w);

// Window controls
void gui_set_title(const char* title);
void gui_set_size(int width, int height);
void gui_set_resizable(bool resizable);

// ============================================================================
// Layout Containers - Simple & Intuitive
// ============================================================================

gui_container gui_vbox(void);           // Vertical stack
gui_container gui_hbox(void);           // Horizontal stack
gui_container gui_grid(int columns);    // Grid layout
gui_container gui_scroll(gui_widget content);  // Scrollable area
gui_container gui_tabs(void);           // Tab notebook

// Layout control
void gui_end_row(void);                 // End horizontal row, start new row
void gui_spacing(int pixels);           // Add vertical spacing
void gui_padding(int left, int top, int right, int bottom);  // Set padding
void gui_set_spacing(int pixels);       // Set default widget spacing
void gui_set_margin(int pixels);        // Set margin from edges

// Section headers
gui_widget gui_header(const char* text); // Bold section header
gui_widget gui_subheader(const char* text); // Smaller section header

// Tab helpers
void gui_tab_add(gui_container tabs, const char* label, gui_widget content);

// ============================================================================
// Basic Widgets - Create & Use
// ============================================================================

// Label
gui_widget gui_label(const char* text);
void gui_set_text(gui_widget w, const char* text);
char* gui_get_text(gui_widget w);

// Button
gui_widget gui_button(const char* text, gui_callback on_click);
gui_widget gui_button_ok(const char* text, gui_callback on_click);  // Primary style

// Entry (text input)
gui_widget gui_entry(const char* placeholder);
gui_widget gui_entry_secret(const char* placeholder);  // Password field
gui_widget gui_entry_multi(const char* placeholder);   // Multiline

// Checkbox
gui_widget gui_check(const char* label);
bool gui_is_checked(gui_widget w);
void gui_set_checked(gui_widget w, bool checked);

// Dropdown/Select
gui_widget gui_select(const char* options[], int count);
int gui_get_selected(gui_widget w);
void gui_set_selected(gui_widget w, int index);

// Slider
gui_widget gui_slider(double min, double max);
double gui_get_value(gui_widget w);
void gui_set_value(gui_widget w, double value);

// Progress
gui_widget gui_progress(void);

// Radio Button
gui_widget gui_radio(const char* label, int group);
bool gui_is_selected(gui_widget w);
void gui_set_selected_radio(gui_widget w, bool selected);
int gui_get_radio_group(gui_widget w);

// Spin Button (numeric input)
gui_widget gui_spin(double min, double max, double step);

// List Box
gui_widget gui_list(const char* items[], int count);
int gui_get_list_selected(gui_widget w);
void gui_set_list_selected(gui_widget w, int index);
void gui_list_add(gui_widget w, const char* item);

// Group Box
gui_widget gui_group(const char* label);
void gui_group_add(gui_widget group, gui_widget widget);

// Color Button
gui_widget gui_color_button(uint8_t r, uint8_t g, uint8_t b);
void gui_get_color(gui_widget w, uint8_t* r, uint8_t* g, uint8_t* b);

// ============================================================================
// Visual Widgets
// ============================================================================

gui_widget gui_separator(void);
gui_widget gui_image(const char* path);
gui_widget gui_spinner(void);  // Loading indicator
void gui_spinner_start(gui_widget w);
void gui_spinner_stop(gui_widget w);

// ============================================================================
// Dialogs - Simple One-Liners
// ============================================================================

void gui_alert_info(const char* message);
void gui_alert_error(const char* message);
void gui_alert_warn(const char* message);
void gui_confirm(const char* message, gui_callback on_result);
void gui_pick_file(gui_callback on_result);
void gui_save_file(const char* default_name, gui_callback on_result);
void gui_pick_folder(gui_callback on_result);

// ============================================================================
// Widget State
// ============================================================================

void gui_show(gui_widget w);
void gui_hide(gui_widget w);
void gui_enable(gui_widget w);
void gui_disable(gui_widget w);
void gui_focus(gui_widget w);

// ============================================================================
// Utility
// ============================================================================

void gui_free(char* str);
char* gui_clipboard_get(void);
void gui_clipboard_set(const char* text);

// ============================================================================
// Convenience - One-Liners for Common Patterns
// ============================================================================

// Create and add in one step
#define gui_add_label(text)        gui_add(gui_label(text))
#define gui_add_button(text, cb)   gui_add(gui_button(text, cb))
#define gui_add_entry(ph)          gui_add(gui_entry(ph))
#define gui_add_check(label)       gui_add(gui_check(label))
#define gui_add_separator()        gui_add(gui_separator())
#define gui_add_progress()         gui_add(gui_progress())

// Container shortcuts
#define vbox_add(c, w)             gui_add_to(c, w)
#define hbox_add(c, w)             gui_add_to(c, w)

#ifdef __cplusplus
}
#endif

#endif // CORTEX_GUI_RUNTIME_H
