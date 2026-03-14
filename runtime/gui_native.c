// gui_native.c - Native Windows GUI implementation
// This provides a complete GUI system using native Windows API

#ifdef _WIN32
#include <windows.h>
#include <commctrl.h>
#include "gui_runtime.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#pragma comment(lib, "comctl32.lib")
#pragma comment(linker,"\"/manifestdependency:type='win32' name='Microsoft.Windows.Common-Controls' version='6.0.0.0' processorArchitecture='*' publicKeyToken='6595b64144ccf1df' language='*'\"")

// ============================================================================
// Widget Types
// ============================================================================

typedef enum {
    WIDGET_NONE = 0,
    WIDGET_LABEL,
    WIDGET_BUTTON,
    WIDGET_ENTRY,
    WIDGET_TEXTAREA,
    WIDGET_CHECKBOX,
    WIDGET_SLIDER,
    WIDGET_PROGRESS,
    WIDGET_IMAGE,
    WIDGET_RECTANGLE,
    WIDGET_CIRCLE,
    WIDGET_LINE
} WidgetType;

// ============================================================================
// Widget Structure
// ============================================================================

typedef struct NativeWidget {
    HWND hwnd;
    WidgetType type;
    int64_t id;
    gui_event_callback callback;
    struct NativeWidget* next;
} NativeWidget;

typedef struct NativeContainer {
    HWND hwnd;
    NativeWidget* widgets;
    int widget_count;
    int64_t id;
} NativeContainer;

typedef struct NativeWindow {
    HWND hwnd;
    NativeContainer* content;
    int running;
    int64_t id;
} NativeWindow;

// ============================================================================
// Global State
// ============================================================================

#define MAX_WINDOWS 16
#define MAX_WIDGETS 256

static NativeWindow windows[MAX_WINDOWS];
static NativeWidget widgets[MAX_WIDGETS];
static NativeContainer containers[MAX_WIDGETS];

static int window_count = 0;
static int widget_count = 0;
static int container_count = 0;
static int64_t next_id = 1;

static NativeWindow* current_window = NULL;
static HINSTANCE g_hInstance = NULL;

// Forward declarations
LRESULT CALLBACK CortexWindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam);
LRESULT CALLBACK CortexButtonProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam);
LRESULT CALLBACK CortexEditProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam);

static WNDPROC original_button_proc = NULL;
static WNDPROC original_edit_proc = NULL;

// ============================================================================
// Helper Functions
// ============================================================================

static int64_t get_next_id(void) {
    return next_id++;
}

static NativeWidget* find_widget_by_hwnd(HWND hwnd) {
    for (int i = 0; i < widget_count; i++) {
        if (widgets[i].hwnd == hwnd) return &widgets[i];
    }
    return NULL;
}

static NativeWidget* find_widget_by_id(int64_t id) {
    for (int i = 0; i < widget_count; i++) {
        if (widgets[i].id == id) return &widgets[i];
    }
    return NULL;
}

static NativeWindow* find_window_by_hwnd(HWND hwnd) {
    for (int i = 0; i < window_count; i++) {
        if (windows[i].hwnd == hwnd) return &windows[i];
    }
    return NULL;
}

// ============================================================================
// Dialog Functions
// ============================================================================

void gui_dialog_info_native(gui_window parent, const char* title, const char* message) {
    MessageBoxA((HWND)parent, message, title, MB_OK | MB_ICONINFORMATION);
}

void gui_dialog_error_native(gui_window parent, const char* title, const char* message) {
    MessageBoxA((HWND)parent, message, title, MB_OK | MB_ICONERROR);
}

int gui_dialog_confirm_native(gui_window parent, const char* title, const char* message) {
    int result = MessageBoxA((HWND)parent, message, title, MB_YESNO | MB_ICONQUESTION);
    return (result == IDYES) ? 1 : 0;
}

// ============================================================================
// Window Procedure
// ============================================================================

LRESULT CALLBACK CortexWindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam) {
    NativeWindow* win = find_window_by_hwnd(hwnd);
    
    switch (uMsg) {
        case WM_CLOSE:
            if (win) win->running = 0;
            DestroyWindow(hwnd);
            return 0;
            
        case WM_DESTROY:
            PostQuitMessage(0);
            return 0;
            
        case WM_COMMAND: {
            HWND ctrl_hwnd = (HWND)lParam;
            int notification = HIWORD(wParam);
            
            NativeWidget* widget = find_widget_by_hwnd(ctrl_hwnd);
            if (widget && widget->callback) {
                gui_event event = {0};
                event.type = GUI_EVENT_CLICK;
                event.source = widget->id;
                
                switch (widget->type) {
                    case WIDGET_BUTTON:
                        widget->callback(event);
                        break;
                        
                    case WIDGET_CHECKBOX: {
                        int checked = (SendMessage(ctrl_hwnd, BM_GETCHECK, 0, 0) == BST_CHECKED);
                        event.type = GUI_EVENT_CHANGE;
                        event.data = (void*)(intptr_t)checked;
                        widget->callback(event);
                        break;
                    }
                        
                    case WIDGET_ENTRY:
                    case WIDGET_TEXTAREA:
                        if (notification == EN_CHANGE) {
                            event.type = GUI_EVENT_CHANGE;
                            widget->callback(event);
                        }
                        break;
                }
            }
            return 0;
        }
        
        case WM_HSCROLL: {
            HWND ctrl_hwnd = (HWND)lParam;
            NativeWidget* widget = find_widget_by_hwnd(ctrl_hwnd);
            
            if (widget && widget->type == WIDGET_SLIDER && widget->callback) {
                int pos = (int)SendMessage(ctrl_hwnd, TBM_GETPOS, 0, 0);
                gui_event event = {0};
                event.type = GUI_EVENT_CHANGE;
                event.source = widget->id;
                event.data = (void*)(intptr_t)pos;
                widget->callback(event);
            }
            return 0;
        }
        
        case WM_SIZE: {
            // Repaint on resize
            InvalidateRect(hwnd, NULL, TRUE);
            return 0;
        }
        
        case WM_PAINT: {
            PAINTSTRUCT ps;
            HDC hdc = BeginPaint(hwnd, &ps);
            
            // Draw any custom graphics widgets
            if (win && win->content) {
                NativeWidget* w = win->content->widgets;
                while (w) {
                    if (w->type == WIDGET_RECTANGLE || w->type == WIDGET_CIRCLE || w->type == WIDGET_LINE) {
                        // Custom drawing handled by widget state
                    }
                    w = w->next;
                }
            }
            
            EndPaint(hwnd, &ps);
            return 0;
        }
    }
    
    return DefWindowProcA(hwnd, uMsg, wParam, lParam);
}

// ============================================================================
// Initialization
// ============================================================================

static int register_window_class(void) {
    static int registered = 0;
    
    if (!registered) {
        INITCOMMONCONTROLSEX icc = {0};
        icc.dwSize = sizeof(icc);
        icc.dwICC = ICC_STANDARD_CLASSES | ICC_BAR_CLASSES;
        InitCommonControlsEx(&icc);
        
        WNDCLASSA wc = {0};
        wc.lpfnWndProc = CortexWindowProc;
        wc.hInstance = GetModuleHandle(NULL);
        wc.lpszClassName = "CortexWindow";
        wc.hbrBackground = (HBRUSH)(COLOR_WINDOW + 1);
        wc.hCursor = LoadCursor(NULL, IDC_ARROW);
        wc.hIcon = LoadIcon(NULL, IDI_APPLICATION);
        
        if (!RegisterClassA(&wc)) {
            return 0;
        }
        registered = 1;
    }
    
    return 1;
}

// ============================================================================
// Window Management
// ============================================================================

gui_window gui_window_create_native(const char* title, int width, int height) {
    if (!register_window_class()) {
        return GUI_INVALID_HANDLE;
    }
    
    if (window_count >= MAX_WINDOWS) {
        return GUI_INVALID_HANDLE;
    }
    
    HWND hwnd = CreateWindowExA(
        0,
        "CortexWindow",
        title,
        WS_OVERLAPPEDWINDOW,
        CW_USEDEFAULT, CW_USEDEFAULT,
        width, height,
        NULL, NULL,
        GetModuleHandle(NULL),
        NULL
    );
    
    if (!hwnd) {
        return GUI_INVALID_HANDLE;
    }
    
    g_hInstance = GetModuleHandle(NULL);
    
    NativeWindow* win = &windows[window_count++];
    win->hwnd = hwnd;
    win->content = NULL;
    win->running = 1;
    win->id = get_next_id();
    
    current_window = win;
    
    return (gui_window)hwnd;
}

void gui_window_show_native(gui_window window) {
    ShowWindow((HWND)window, SW_SHOW);
    UpdateWindow((HWND)window);
}

void gui_window_center_native(gui_window window) {
    RECT rc;
    GetWindowRect((HWND)window, &rc);
    int width = rc.right - rc.left;
    int height = rc.bottom - rc.top;
    
    int screen_width = GetSystemMetrics(SM_CXSCREEN);
    int screen_height = GetSystemMetrics(SM_CYSCREEN);
    
    int x = (screen_width - width) / 2;
    int y = (screen_height - height) / 2;
    
    SetWindowPos((HWND)window, NULL, x, y, 0, 0, SWP_NOSIZE | SWP_NOZORDER);
}

void gui_run_native(void) {
    MSG msg;
    while (GetMessage(&msg, NULL, 0, 0)) {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }
}

void gui_quit_native(void) {
    for (int i = 0; i < window_count; i++) {
        if (windows[i].hwnd) {
            PostMessage(windows[i].hwnd, WM_CLOSE, 0, 0);
        }
    }
}

// ============================================================================
// Widget Creation
// ============================================================================

gui_widget gui_label_create(const char* text) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_INVALID_HANDLE;
    }
    
    HWND hwnd = CreateWindowExA(
        0, "STATIC", text,
        WS_VISIBLE | WS_CHILD | SS_LEFT,
        10, 10, 200, 25,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_INVALID_HANDLE;
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_LABEL;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

void gui_label_set_text(gui_widget label, const char* text) {
    SetWindowTextA((HWND)label, text);
}

gui_widget gui_button_create(const char* label, gui_event_callback on_click) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_INVALID_HANDLE;
    }
    
    HWND hwnd = CreateWindowExA(
        0, "BUTTON", label,
        WS_VISIBLE | WS_CHILD | BS_PUSHBUTTON,
        10, 10, 100, 30,
        current_window->hwnd, (HMENU)get_next_id(), g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_INVALID_HANDLE;
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_BUTTON;
    w->id = get_next_id();
    w->callback = on_click;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

gui_widget gui_entry_create(const char* placeholder, gui_event_callback on_change) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_INVALID_HANDLE;
    }
    
    HWND hwnd = CreateWindowExA(
        WS_EX_CLIENTEDGE, "EDIT", placeholder,
        WS_VISIBLE | WS_CHILD | ES_AUTOHSCROLL,
        10, 10, 200, 25,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_INVALID_HANDLE;
    
    // Set placeholder text (gray)
    SendMessage(hwnd, EM_SETCUEBANNER, TRUE, (LPARAM)placeholder);
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_ENTRY;
    w->id = get_next_id();
    w->callback = on_change;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

char* gui_entry_get_text(gui_widget entry) {
    int len = GetWindowTextLength((HWND)entry) + 1;
    char* buffer = (char*)malloc(len);
    if (buffer) {
        GetWindowTextA((HWND)entry, buffer, len);
    }
    return buffer;
}

void gui_entry_set_text(gui_widget entry, const char* text) {
    SetWindowTextA((HWND)entry, text);
}

gui_widget gui_textarea_create(const char* placeholder, gui_event_callback on_change) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_INVALID_HANDLE;
    }
    
    HWND hwnd = CreateWindowExA(
        WS_EX_CLIENTEDGE, "EDIT", "",
        WS_VISIBLE | WS_CHILD | ES_MULTILINE | ES_AUTOVSCROLL | WS_VSCROLL,
        10, 10, 300, 150,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_INVALID_HANDLE;
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_TEXTAREA;
    w->id = get_next_id();
    w->callback = on_change;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

char* gui_textarea_get_text(gui_widget textarea) {
    int len = GetWindowTextLength((HWND)textarea) + 1;
    char* buffer = (char*)malloc(len);
    if (buffer) {
        GetWindowTextA((HWND)textarea, buffer, len);
    }
    return buffer;
}

void gui_textarea_set_text(gui_widget textarea, const char* text) {
    SetWindowTextA((HWND)textarea, text);
}

gui_widget gui_checkbox_create(const char* label, gui_event_callback on_change) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_INVALID_HANDLE;
    }
    
    HWND hwnd = CreateWindowExA(
        0, "BUTTON", label,
        WS_VISIBLE | WS_CHILD | BS_AUTOCHECKBOX,
        10, 10, 200, 25,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_INVALID_HANDLE;
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_CHECKBOX;
    w->id = get_next_id();
    w->callback = on_change;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

bool gui_checkbox_get_state(gui_widget checkbox) {
    return SendMessage((HWND)checkbox, BM_GETCHECK, 0, 0) == BST_CHECKED;
}

void gui_checkbox_set_state(gui_widget checkbox, bool checked) {
    SendMessage((HWND)checkbox, BM_SETCHECK, checked ? BST_CHECKED : BST_UNCHECKED, 0);
}

gui_widget gui_slider_create(double min, double max, double value, gui_event_callback on_change) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_INVALID_HANDLE;
    }
    
    HWND hwnd = CreateWindowExA(
        0, TRACKBAR_CLASSA, "",
        WS_VISIBLE | WS_CHILD | TBS_HORZ,
        10, 10, 200, 30,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_INVALID_HANDLE;
    
    SendMessage(hwnd, TBM_SETRANGE, TRUE, MAKELPARAM((int)min, (int)max));
    SendMessage(hwnd, TBM_SETPOS, TRUE, (LPARAM)(int)value);
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_SLIDER;
    w->id = get_next_id();
    w->callback = on_change;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

double gui_slider_get_value(gui_widget slider) {
    return (double)SendMessage((HWND)slider, TBM_GETPOS, 0, 0);
}

void gui_slider_set_value(gui_widget slider, double value) {
    SendMessage((HWND)slider, TBM_SETPOS, TRUE, (LPARAM)(int)value);
}

gui_widget gui_progress_create(void) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_INVALID_HANDLE;
    }
    
    HWND hwnd = CreateWindowExA(
        0, PROGRESS_CLASSA, "",
        WS_VISIBLE | WS_CHILD | PBS_SMOOTH,
        10, 10, 200, 25,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_INVALID_HANDLE;
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_PROGRESS;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

void gui_progress_set_value(gui_widget progress, double value) {
    SendMessage((HWND)progress, PBM_SETPOS, (int)(value * 100), 0);
}

// ============================================================================
// Layout Containers
// ============================================================================

gui_container gui_vbox_create(void) {
    if (container_count >= MAX_WIDGETS) {
        return GUI_INVALID_HANDLE;
    }
    
    NativeContainer* c = &containers[container_count++];
    c->hwnd = current_window ? current_window->hwnd : NULL;
    c->widgets = NULL;
    c->widget_count = 0;
    c->id = get_next_id();
    
    return (gui_container)c;
}

gui_container gui_hbox_create(void) {
    return gui_vbox_create(); // Same implementation, layout handled in add
}

gui_container gui_grid_create(int columns) {
    (void)columns;
    return gui_vbox_create();
}

void gui_container_add(gui_container container, gui_widget widget) {
    if (!container || !widget) return;
    
    NativeContainer* c = (NativeContainer*)container;
    NativeWidget* w = find_widget_by_hwnd((HWND)widget);
    
    if (w) {
        w->next = c->widgets;
        c->widgets = w;
        c->widget_count++;
        
        // Position widget based on container type
        int y = 10 + (c->widget_count - 1) * 35;
        int x = 10;
        
        SetWindowPos((HWND)widget, NULL, x, y, 0, 0, SWP_NOSIZE | SWP_NOZORDER);
    }
}

// ============================================================================
// Widget Management
// ============================================================================

void gui_refresh(gui_widget widget) {
    InvalidateRect((HWND)widget, NULL, TRUE);
    UpdateWindow((HWND)widget);
}

void gui_resize(gui_widget widget, float width, float height) {
    SetWindowPos((HWND)widget, NULL, 0, 0, (int)width, (int)height, SWP_NOMOVE | SWP_NOZORDER);
}

void gui_move(gui_widget widget, float x, float y) {
    SetWindowPos((HWND)widget, NULL, (int)x, (int)y, 0, 0, SWP_NOSIZE | SWP_NOZORDER);
}

void gui_enable(gui_widget widget) {
    EnableWindow((HWND)widget, TRUE);
}

void gui_disable(gui_widget widget) {
    EnableWindow((HWND)widget, FALSE);
}

bool gui_is_enabled(gui_widget widget) {
    return IsWindowEnabled((HWND)widget) != 0;
}

void gui_hide(gui_widget widget) {
    ShowWindow((HWND)widget, SW_HIDE);
}

void gui_show(gui_widget widget) {
    ShowWindow((HWND)widget, SW_SHOW);
}

#endif // _WIN32
