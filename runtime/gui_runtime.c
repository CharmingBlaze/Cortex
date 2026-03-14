// gui_runtime.c - Cortex GUI Runtime Implementation
//
// This file implements the C API for the Cortex GUI system.
// Uses native Windows API for simple dialogs and windows.

#include "gui_runtime.h"
#include "core.h"
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#ifdef _WIN32
#include <windows.h>

// Native Windows implementations
void gui_dialog_info_native(gui_window parent, const char* title, const char* message);
void gui_dialog_error_native(gui_window parent, const char* title, const char* message);
int gui_dialog_confirm_native(gui_window parent, const char* title, const char* message);
gui_window gui_window_create_native(const char* title, int width, int height);
void gui_window_show_native(gui_window window);
void gui_window_center_native(gui_window window);
void gui_run_native(void);
void gui_quit_native(void);
#endif

// ============================================================================
// Go Function Declarations (from bridge.go)
// These are implemented in Go and called via cgo.
// Only used on non-Windows platforms.
// ============================================================================

#ifndef _WIN32
extern void* CortexGUI_CreateWindow(const char* title, int width, int height);
extern void CortexGUI_ShowWindow(int64_t handle);
extern void CortexGUI_HideWindow(int64_t handle);
extern void CortexGUI_CloseWindow(int64_t handle);
extern void CortexGUI_SetWindowTitle(int64_t handle, const char* title);
extern void CortexGUI_CenterWindow(int64_t handle);
extern void CortexGUI_SetWindowFixedSize(int64_t handle, int fixed);
extern void CortexGUI_FullscreenWindow(int64_t handle, int fullscreen);
extern void CortexGUI_SetWindowContent(int64_t window, int64_t content);

extern int64_t CortexGUI_CreateLabel(const char* text);
extern void CortexGUI_SetLabelText(int64_t handle, const char* text);

extern int64_t CortexGUI_CreateButton(const char* label, int64_t callback_id);
extern int64_t CortexGUI_CreateEntry(const char* placeholder, int64_t callback_id);
extern char* CortexGUI_GetEntryText(int64_t handle);
extern void CortexGUI_SetEntryText(int64_t handle, const char* text);

extern int64_t CortexGUI_CreateTextArea(const char* placeholder, int64_t callback_id);
extern int64_t CortexGUI_CreateCheck(const char* label, int64_t callback_id);
extern int CortexGUI_GetCheckState(int64_t handle);
extern void CortexGUI_SetCheckState(int64_t handle, int checked);

extern int64_t CortexGUI_CreateSlider(double min, double max, double value, int64_t callback_id);
extern double CortexGUI_GetSliderValue(int64_t handle);
extern void CortexGUI_SetSliderValue(int64_t handle, double value);

extern int64_t CortexGUI_CreateProgress(void);
extern void CortexGUI_SetProgressValue(int64_t handle, double value);

extern int64_t CortexGUI_CreateImage(const char* filepath);
extern void CortexGUI_SetImageFill(int64_t handle, int fill_mode);

extern int64_t CortexGUI_CreateRectangle(uint8_t r, uint8_t g, uint8_t b, uint8_t a);
extern int64_t CortexGUI_CreateCircle(uint8_t r, uint8_t g, uint8_t b, uint8_t a);
extern int64_t CortexGUI_CreateLine(float x1, float y1, float x2, float y2);
extern void CortexGUI_SetLineColor(int64_t handle, uint8_t r, uint8_t g, uint8_t b, uint8_t a);

extern int64_t CortexGUI_CreateVBox(int64_t* handles, int count);
extern int64_t CortexGUI_CreateHBox(int64_t* handles, int count);
extern int64_t CortexGUI_CreateGrid(int columns, int64_t* handles, int count);
extern void CortexGUI_AddToContainer(int64_t container, int64_t widget);

extern void CortexGUI_ShowInfoDialog(int64_t window, const char* title, const char* message);
extern void CortexGUI_ShowErrorDialog(int64_t window, const char* title, const char* message);
extern void CortexGUI_ShowConfirmDialog(int64_t window, const char* title, const char* message, int64_t callback_id);
extern void CortexGUI_ShowFileOpenDialog(int64_t window, int64_t callback_id);
extern void CortexGUI_ShowFileSaveDialog(int64_t window, int64_t callback_id);

extern void CortexGUI_Run(void);
extern void CortexGUI_Quit(void);
extern void CortexGUI_Refresh(int64_t handle);
extern void CortexGUI_Resize(int64_t handle, float width, float height);
extern void CortexGUI_Move(int64_t handle, float x, float y);
extern void CortexGUI_Enable(int64_t handle);
extern void CortexGUI_Disable(int64_t handle);
extern int CortexGUI_IsEnabled(int64_t handle);

extern void CortexGUI_FreeString(char* str);
extern int CortexGUI_Version(void);
#endif

// ============================================================================
// Callback Registry
// ============================================================================

#define MAX_CALLBACKS 1024

static struct {
    gui_event_callback func;
    int64_t id;
} callback_registry[MAX_CALLBACKS];

static int callback_count = 0;
static int64_t next_callback_id = 1;

// Internal function to register a callback and get an ID
static int64_t register_callback(gui_event_callback callback) {
    if (callback_count >= MAX_CALLBACKS || callback == NULL) {
        return 0;
    }
    
    int64_t id = next_callback_id++;
    callback_registry[callback_count].func = callback;
    callback_registry[callback_count].id = id;
    callback_count++;
    
    return id;
}

// Called from Go when an event occurs
void cortex_gui_event_handler(int event_type, int64_t source, void* data) {
    // Find the callback
    for (int i = 0; i < callback_count; i++) {
        if (callback_registry[i].id == source) {
            gui_event event = {
                .type = (gui_event_type)event_type,
                .source = source,
                .data = data
            };
            callback_registry[i].func(event);
            return;
        }
    }
}

// ============================================================================
// Application Lifecycle
// ============================================================================

void gui_init(void) {
    // Initialization happens automatically
}

void gui_run(void) {
#ifdef _WIN32
    gui_run_native();
#else
    CortexGUI_Run();
#endif
}

void gui_quit(void) {
#ifdef _WIN32
    gui_quit_native();
#else
    CortexGUI_Quit();
#endif
}

bool gui_is_running(void) {
    return true;
}

int gui_version(void) {
    return 1;
}

// ============================================================================
// Window Management
// ============================================================================

gui_window gui_window_create(const char* title, int width, int height) {
#ifdef _WIN32
    return gui_window_create_native(title, width, height);
#else
    return (gui_window)CortexGUI_CreateWindow(title, width, height);
#endif
}

void gui_window_show(gui_window window) {
#ifdef _WIN32
    gui_window_show_native(window);
#else
    CortexGUI_ShowWindow(window);
#endif
}

void gui_window_hide(gui_window window) {
#ifdef _WIN32
    ShowWindow((HWND)window, SW_HIDE);
#else
    CortexGUI_HideWindow(window);
#endif
}

void gui_window_close(gui_window window) {
#ifdef _WIN32
    DestroyWindow((HWND)window);
#else
    CortexGUI_CloseWindow(window);
#endif
}

void gui_window_set_title(gui_window window, const char* title) {
#ifdef _WIN32
    SetWindowTextA((HWND)window, title);
#else
    CortexGUI_SetWindowTitle(window, title);
#endif
}

void gui_window_center(gui_window window) {
#ifdef _WIN32
    gui_window_center_native(window);
#else
    CortexGUI_CenterWindow(window);
#endif
}

void gui_window_set_fixed_size(gui_window window, bool fixed) {
#ifdef _WIN32
    (void)window; (void)fixed;
#else
    CortexGUI_SetWindowFixedSize(window, fixed ? 1 : 0);
#endif
}

void gui_window_set_fullscreen(gui_window window, bool fullscreen) {
#ifdef _WIN32
    (void)window; (void)fullscreen;
#else
    CortexGUI_FullscreenWindow(window, fullscreen ? 1 : 0);
#endif
}

void gui_window_set_content(gui_window window, gui_container content) {
#ifdef _WIN32
    (void)window; (void)content;
#else
    CortexGUI_SetWindowContent(window, content);
#endif
}

// ============================================================================
// Widget Creation - Basic
// ============================================================================

#ifndef _WIN32
gui_widget gui_label_create(const char* text) {
    return (gui_widget)CortexGUI_CreateLabel(text);
}

void gui_label_set_text(gui_widget label, const char* text) {
    CortexGUI_SetLabelText(label, text);
}

gui_widget gui_button_create(const char* label, gui_event_callback on_click) {
    int64_t callback_id = register_callback(on_click);
    return (gui_widget)CortexGUI_CreateButton(label, callback_id);
}

gui_widget gui_entry_create(const char* placeholder, gui_event_callback on_change) {
    int64_t callback_id = register_callback(on_change);
    return (gui_widget)CortexGUI_CreateEntry(placeholder, callback_id);
}

char* gui_entry_get_text(gui_widget entry) {
    return CortexGUI_GetEntryText(entry);
}

void gui_entry_set_text(gui_widget entry, const char* text) {
    CortexGUI_SetEntryText(entry, text);
}

gui_widget gui_textarea_create(const char* placeholder, gui_event_callback on_change) {
    int64_t callback_id = register_callback(on_change);
    return (gui_widget)CortexGUI_CreateTextArea(placeholder, callback_id);
}

char* gui_textarea_get_text(gui_widget textarea) {
    return CortexGUI_GetEntryText(textarea);
}

void gui_textarea_set_text(gui_widget textarea, const char* text) {
    CortexGUI_SetEntryText(textarea, text);
}

// ============================================================================
// Widget Creation - Input Controls
// ============================================================================

gui_widget gui_checkbox_create(const char* label, gui_event_callback on_change) {
    int64_t callback_id = register_callback(on_change);
    return (gui_widget)CortexGUI_CreateCheck(label, callback_id);
}

bool gui_checkbox_get_state(gui_widget checkbox) {
    return CortexGUI_GetCheckState(checkbox) != 0;
}

void gui_checkbox_set_state(gui_widget checkbox, bool checked) {
    CortexGUI_SetCheckState(checkbox, checked ? 1 : 0);
}

gui_widget gui_slider_create(double min, double max, double value, gui_event_callback on_change) {
    int64_t callback_id = register_callback(on_change);
    return (gui_widget)CortexGUI_CreateSlider(min, max, value, callback_id);
}

double gui_slider_get_value(gui_widget slider) {
    return CortexGUI_GetSliderValue(slider);
}

void gui_slider_set_value(gui_widget slider, double value) {
    CortexGUI_SetSliderValue(slider, value);
}

gui_widget gui_progress_create(void) {
    return (gui_widget)CortexGUI_CreateProgress();
}

void gui_progress_set_value(gui_widget progress, double value) {
    CortexGUI_SetProgressValue(progress, value);
}
#endif

// ============================================================================
// Widget Creation - Graphics
// ============================================================================

gui_widget gui_image_create(const char* filepath) {
#ifdef _WIN32
    (void)filepath;
    return GUI_INVALID_HANDLE;
#else
    return (gui_widget)CortexGUI_CreateImage(filepath);
#endif
}

void gui_image_set_fill(gui_widget image, int fill_mode) {
#ifdef _WIN32
    (void)image; (void)fill_mode;
#else
    CortexGUI_SetImageFill(image, fill_mode);
#endif
}

gui_widget gui_rectangle_create(uint8_t r, uint8_t g, uint8_t b, uint8_t a) {
#ifdef _WIN32
    (void)r; (void)g; (void)b; (void)a;
    return GUI_INVALID_HANDLE;
#else
    return (gui_widget)CortexGUI_CreateRectangle(r, g, b, a);
#endif
}

gui_widget gui_circle_create(uint8_t r, uint8_t g, uint8_t b, uint8_t a) {
#ifdef _WIN32
    (void)r; (void)g; (void)b; (void)a;
    return GUI_INVALID_HANDLE;
#else
    return (gui_widget)CortexGUI_CreateCircle(r, g, b, a);
#endif
}

gui_widget gui_line_create(float x1, float y1, float x2, float y2) {
#ifdef _WIN32
    (void)x1; (void)y1; (void)x2; (void)y2;
    return GUI_INVALID_HANDLE;
#else
    return (gui_widget)CortexGUI_CreateLine(x1, y1, x2, y2);
#endif
}

void gui_line_set_color(gui_widget line, uint8_t r, uint8_t g, uint8_t b, uint8_t a) {
#ifdef _WIN32
    (void)line; (void)r; (void)g; (void)b; (void)a;
#else
    CortexGUI_SetLineColor(line, r, g, b, a);
#endif
}

// ============================================================================
// Layout Containers
// ============================================================================

#ifndef _WIN32
gui_container gui_vbox_create(void) {
    // We'll build this dynamically using AddToContainer
    // For now, create an empty container
    return GUI_INVALID_HANDLE;
}

gui_container gui_hbox_create(void) {
    return GUI_INVALID_HANDLE;
}

gui_container gui_grid_create(int columns) {
    (void)columns;
    return GUI_INVALID_HANDLE;
}

void gui_container_add(gui_container container, gui_widget widget) {
    if (container == GUI_INVALID_HANDLE) {
        return;
    }
    CortexGUI_AddToContainer(container, widget);
}
#else
// Windows implementations are in gui_native.c
#endif

void gui_container_add_many(gui_container container, int count, ...) {
    (void)container;
    (void)count;
    // TODO: Implement variadic version
}

// ============================================================================
// Advanced Layout
// ============================================================================

gui_container gui_border_create(
    gui_widget top,
    gui_widget bottom,
    gui_widget left,
    gui_widget right,
    gui_widget center
) {
    // This will be implemented with specific Go calls
    (void)top; (void)bottom; (void)left; (void)right; (void)center;
    return GUI_INVALID_HANDLE;
}

void gui_container_set_padding(gui_container container, int padding) {
    (void)container;
    (void)padding;
    // TODO: Implement
}

// ============================================================================
// Dialogs
// ============================================================================

void gui_dialog_info(gui_window parent, const char* title, const char* message) {
#ifdef _WIN32
    gui_dialog_info_native(parent, title, message);
#else
    CortexGUI_ShowInfoDialog(parent, title, message);
#endif
}

void gui_dialog_error(gui_window parent, const char* title, const char* message) {
#ifdef _WIN32
    gui_dialog_error_native(parent, title, message);
#else
    CortexGUI_ShowErrorDialog(parent, title, message);
#endif
}

void gui_dialog_confirm(
    gui_window parent,
    const char* title,
    const char* message,
    gui_event_callback on_result
) {
#ifdef _WIN32
    int result = gui_dialog_confirm_native(parent, title, message);
    if (on_result) {
        gui_event event = {.type = GUI_EVENT_CLICK, .data = (void*)(intptr_t)result};
        on_result(event);
    }
#else
    int64_t callback_id = register_callback(on_result);
    CortexGUI_ShowConfirmDialog(parent, title, message, callback_id);
#endif
}

void gui_dialog_file_open(gui_window parent, gui_event_callback on_result) {
#ifdef _WIN32
    (void)parent; (void)on_result;
#else
    int64_t callback_id = register_callback(on_result);
    CortexGUI_ShowFileOpenDialog(parent, callback_id);
#endif
}

void gui_dialog_file_save(gui_window parent, gui_event_callback on_result) {
#ifdef _WIN32
    (void)parent; (void)on_result;
#else
    int64_t callback_id = register_callback(on_result);
    CortexGUI_ShowFileSaveDialog(parent, callback_id);
#endif
}

// ============================================================================
// Widget Management
// ============================================================================

#ifndef _WIN32
void gui_refresh(gui_widget widget) {
    CortexGUI_Refresh(widget);
}

void gui_resize(gui_widget widget, float width, float height) {
    CortexGUI_Resize(widget, width, height);
}

void gui_move(gui_widget widget, float x, float y) {
    CortexGUI_Move(widget, x, y);
}

void gui_enable(gui_widget widget) {
    CortexGUI_Enable(widget);
}

void gui_disable(gui_widget widget) {
    CortexGUI_Disable(widget);
}

bool gui_is_enabled(gui_widget widget) {
    return CortexGUI_IsEnabled(widget) != 0;
}

void gui_hide(gui_widget widget) {
    // TODO: Implement via Go
    (void)widget;
}

void gui_show(gui_widget widget) {
    // TODO: Implement via Go
    (void)widget;
}
#endif

// ============================================================================
// Utility Functions
// ============================================================================

void gui_free_string(char* str) {
    if (str) {
#ifdef _WIN32
        free(str);
#else
        CortexGUI_FreeString(str);
#endif
    }
}

const char* gui_event_get_string(gui_event event) {
    if (event.data == NULL) {
        return "";
    }
    // In the Go implementation, string data is passed as a C string pointer
    return (const char*)event.data;
}

bool gui_event_get_bool(gui_event event) {
    if (event.data == NULL) {
        return false;
    }
    // In the Go implementation, bool is passed as int 0 or 1
    return (intptr_t)event.data != 0;
}

double gui_event_get_double(gui_event event) {
    if (event.data == NULL) {
        return 0.0;
    }
    // This is a simplification - proper implementation would need type checking
    return (double)(intptr_t)event.data;
}
