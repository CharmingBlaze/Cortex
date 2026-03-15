// gui_gtk4_internal.h - Internal GTK4 implementation details
//
// This header provides internal types, macros, and utilities for the GTK4 backend.
// It is NOT part of the public API and should not be included by user code.

#ifndef CORTEX_GUI_GTK4_INTERNAL_H
#define CORTEX_GUI_GTK4_INTERNAL_H

#include <gtk/gtk.h>
#include <glib.h>
#include <stdint.h>
#include <stdbool.h>
#include <stdlib.h>
#include <string.h>

// Include public API
#include "runtime/gui_runtime.h"

// ============================================================================
// Handle Registry
// ============================================================================

// Maximum number of widgets that can be tracked
#define GUI_MAX_WIDGETS 4096

// Widget types for internal tracking
typedef enum {
    WIDGET_TYPE_NONE = 0,
    WIDGET_TYPE_WINDOW,
    WIDGET_TYPE_CONTAINER,
    WIDGET_TYPE_LABEL,
    WIDGET_TYPE_BUTTON,
    WIDGET_TYPE_ENTRY,
    WIDGET_TYPE_CHECK,
    WIDGET_TYPE_RADIO,
    WIDGET_TYPE_SELECT,
    WIDGET_TYPE_SLIDER,
    WIDGET_TYPE_PROGRESS,
    WIDGET_TYPE_IMAGE,
    WIDGET_TYPE_SEPARATOR,
    WIDGET_TYPE_SPINNER,
    WIDGET_TYPE_TEXT_VIEW,
    WIDGET_TYPE_LIST_BOX,
    WIDGET_TYPE_TREE_VIEW,
    WIDGET_TYPE_GRID_VIEW,
    WIDGET_TYPE_CANVAS,
    WIDGET_TYPE_MENU,
    WIDGET_TYPE_MENU_ITEM,
    WIDGET_TYPE_TOOLBAR,
} WidgetType;

// Widget entry in the registry
typedef struct {
    GtkWidget *widget;          // The actual GTK widget
    WidgetType type;            // Widget type for casting
    gui_callback callback;      // Event callback (if any)
    void *user_data;            // User data for callbacks
    bool in_use;                // Is this slot in use?
} WidgetEntry;

// Global registry
typedef struct {
    WidgetEntry widgets[GUI_MAX_WIDGETS];
    int64_t next_handle;
    int64_t main_window;
    bool is_running;
    GMainLoop *main_loop;
} GUIRegistry;

// Global registry instance
extern GUIRegistry g_gui_registry;

// ============================================================================
// Handle Management
// ============================================================================

// Register a widget and return its handle
static inline int64_t gui_register_widget(GtkWidget *widget, WidgetType type) {
    for (int i = 1; i < GUI_MAX_WIDGETS; i++) {
        if (!g_gui_registry.widgets[i].in_use) {
            g_gui_registry.widgets[i].widget = widget;
            g_gui_registry.widgets[i].type = type;
            g_gui_registry.widgets[i].callback = NULL;
            g_gui_registry.widgets[i].user_data = NULL;
            g_gui_registry.widgets[i].in_use = true;
            return i;
        }
    }
    g_warning("GUI registry full!");
    return GUI_NULL;
}

// Get widget entry by handle
static inline WidgetEntry* gui_get_entry(int64_t handle) {
    if (handle <= 0 || handle >= GUI_MAX_WIDGETS) return NULL;
    if (!g_gui_registry.widgets[handle].in_use) return NULL;
    return &g_gui_registry.widgets[handle];
}

// Get GTK widget by handle
static inline GtkWidget* gui_get_widget(int64_t handle) {
    WidgetEntry *entry = gui_get_entry(handle);
    return entry ? entry->widget : NULL;
}

// Get widget type by handle
static inline WidgetType gui_get_type(int64_t handle) {
    WidgetEntry *entry = gui_get_entry(handle);
    return entry ? entry->type : WIDGET_TYPE_NONE;
}

// Unregister a widget
static inline void gui_unregister(int64_t handle) {
    if (handle > 0 && handle < GUI_MAX_WIDGETS) {
        g_gui_registry.widgets[handle].in_use = false;
        g_gui_registry.widgets[handle].widget = NULL;
        g_gui_registry.widgets[handle].callback = NULL;
        g_gui_registry.widgets[handle].user_data = NULL;
    }
}

// Set callback for a widget
static inline void gui_set_callback(int64_t handle, gui_callback callback) {
    WidgetEntry *entry = gui_get_entry(handle);
    if (entry) {
        entry->callback = callback;
    }
}

// Get callback for a widget
static inline gui_callback gui_get_callback(int64_t handle) {
    WidgetEntry *entry = gui_get_entry(handle);
    return entry ? entry->callback : NULL;
}

// ============================================================================
// Event Helpers
// ============================================================================

// Create and dispatch an event
static inline void gui_dispatch_event(int64_t source, gui_event_type type, 
                                       double value, const char *text, 
                                       int key, float x, float y, bool bool_val) {
    WidgetEntry *entry = gui_get_entry(source);
    if (entry && entry->callback) {
        gui_event event = {
            .type = type,
            .source = source,
            .value = value,
            .text = text ? g_strdup(text) : NULL,
            .key = key,
            .x = x,
            .y = y,
            .bool_val = bool_val,
        };
        entry->callback(event);
        if (event.text) g_free(event.text);
    }
}

// Convenience macros for common events
#define gui_dispatch_click(source) \
    gui_dispatch_event(source, GUI_EVENT_CLICK, 0, NULL, 0, 0, 0, false)

#define gui_dispatch_change(source, text) \
    gui_dispatch_event(source, GUI_EVENT_CHANGE, 0, text, 0, 0, 0, false)

#define gui_dispatch_check(source, checked) \
    gui_dispatch_event(source, GUI_EVENT_CHECK, 0, NULL, 0, 0, 0, checked)

#define gui_dispatch_slider(source, value) \
    gui_dispatch_event(source, GUI_EVENT_SLIDER, value, NULL, 0, 0, 0, false)

#define gui_dispatch_select(source, index, text) \
    gui_dispatch_event(source, GUI_EVENT_SELECT, (double)index, text, 0, 0, 0, false)

// ============================================================================
// String Helpers
// ============================================================================

// Duplicate string (caller must free with g_free)
static inline char* gui_strdup(const char *str) {
    return str ? g_strdup(str) : NULL;
}

// Free a string allocated by the GUI library
static inline void gui_strfree(char *str) {
    g_free(str);
}

// ============================================================================
// Color Helpers
// ============================================================================

// Create GDK RGBA from uint8 components
static inline GdkRGBA gui_make_rgba(uint8_t r, uint8_t g, uint8_t b, uint8_t a) {
    GdkRGBA color = {
        .red = r / 255.0,
        .green = g / 255.0,
        .blue = b / 255.0,
        .alpha = a / 255.0,
    };
    return color;
}

// ============================================================================
// Widget Type Checkers
// ============================================================================

#define IS_WINDOW(h)    (gui_get_type(h) == WIDGET_TYPE_WINDOW)
#define IS_CONTAINER(h) (gui_get_type(h) == WIDGET_TYPE_CONTAINER)
#define IS_LABEL(h)     (gui_get_type(h) == WIDGET_TYPE_LABEL)
#define IS_BUTTON(h)    (gui_get_type(h) == WIDGET_TYPE_BUTTON)
#define IS_ENTRY(h)     (gui_get_type(h) == WIDGET_TYPE_ENTRY)

// ============================================================================
// Module Initialization
// ============================================================================

// Initialize the registry
void gui_registry_init(void);

// Cleanup the registry
void gui_registry_cleanup(void);

#endif // CORTEX_GUI_GTK4_INTERNAL_H
