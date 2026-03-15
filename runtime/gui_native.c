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
    gui_callback callback;
    struct NativeWidget* next;
    // Shape data for custom drawing
    uint8_t color_r, color_g, color_b, color_a;
    int shape_x, shape_y, shape_w, shape_h;
    int line_x1, line_y1, line_x2, line_y2;
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

// Auto-layout state
static int layout_next_y = 15;  // Next Y position for auto-layout
static int layout_margin = 15;  // Margin from edges
static int layout_spacing = 12; // Spacing between widgets
static int layout_row_height = 32; // Default row height
static int layout_current_x = 15; // Current X for horizontal layout
static bool layout_is_horizontal = false; // In horizontal box mode?
static int layout_row_start_y = 15; // Y position at start of current row

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

// Get next Y position and advance for auto-layout
static int get_next_layout_y(int height) {
    int y = layout_next_y;
    if (!layout_is_horizontal) {
        layout_next_y += height + layout_spacing + 4; // Extra padding
        layout_current_x = layout_margin;
    }
    return y;
}

// Get next X position for horizontal layout
static int get_next_layout_x(int width) {
    int x = layout_current_x;
    layout_current_x += width + layout_spacing + 8; // Extra horizontal spacing
    return x;
}

// Reset layout for new window
static void reset_layout(void) {
    layout_next_y = layout_margin;
    layout_current_x = layout_margin;
    layout_is_horizontal = false;
    layout_row_start_y = layout_margin;
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
                event.type = GUI_CLICK;
                event.source = widget->id;
                
                switch (widget->type) {
                    case WIDGET_BUTTON:
                        if (widget->callback) widget->callback(event);
                        break;
                        
                    case WIDGET_CHECKBOX: {
                        int checked = (SendMessage(ctrl_hwnd, BM_GETCHECK, 0, 0) == BST_CHECKED);
                        event.type = GUI_CHECK;
                        event.checked = checked;
                        if (widget->callback) widget->callback(event);
                        break;
                    }
                        
                    case WIDGET_ENTRY:
                    case WIDGET_TEXTAREA:
                        if (notification == EN_CHANGE) {
                            event.type = GUI_CHANGE;
                            if (widget->callback) widget->callback(event);
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
                event.type = GUI_CHANGE;
                event.source = widget->id;
                event.value = (double)pos;
                if (widget->callback) widget->callback(event);
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
                    if (w->type == WIDGET_RECTANGLE) {
                        // Draw filled rectangle
                        HBRUSH brush = CreateSolidBrush(RGB(w->color_r, w->color_g, w->color_b));
                        RECT rc = {w->shape_x, w->shape_y, w->shape_x + w->shape_w, w->shape_y + w->shape_h};
                        FillRect(hdc, &rc, brush);
                        DeleteObject(brush);
                    }
                    else if (w->type == WIDGET_CIRCLE) {
                        // Draw filled ellipse
                        HBRUSH brush = CreateSolidBrush(RGB(w->color_r, w->color_g, w->color_b));
                        HBRUSH oldBrush = (HBRUSH)SelectObject(hdc, brush);
                        HPEN pen = CreatePen(PS_NULL, 0, 0);
                        HPEN oldPen = (HPEN)SelectObject(hdc, pen);
                        Ellipse(hdc, w->shape_x, w->shape_y, w->shape_x + w->shape_w, w->shape_y + w->shape_h);
                        SelectObject(hdc, oldBrush);
                        SelectObject(hdc, oldPen);
                        DeleteObject(brush);
                        DeleteObject(pen);
                    }
                    else if (w->type == WIDGET_LINE) {
                        // Draw line
                        HPEN pen = CreatePen(PS_SOLID, 2, RGB(w->color_r, w->color_g, w->color_b));
                        HPEN oldPen = (HPEN)SelectObject(hdc, pen);
                        MoveToEx(hdc, w->line_x1, w->line_y1, NULL);
                        LineTo(hdc, w->line_x2, w->line_y2);
                        SelectObject(hdc, oldPen);
                        DeleteObject(pen);
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
        return GUI_NULL;
    }
    
    if (window_count >= MAX_WINDOWS) {
        return GUI_NULL;
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
        return GUI_NULL;
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

void gui_window_set_content_native(gui_window window, gui_container content) {
    // Find the window structure
    for (int i = 0; i < window_count; i++) {
        if (windows[i].hwnd == (HWND)window) {
            windows[i].content = (NativeContainer*)content;
            // Also set the container's hwnd to the window
            if (content) {
                ((NativeContainer*)content)->hwnd = (HWND)window;
            }
            break;
        }
    }
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

gui_widget gui_label(const char* text) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    int y = get_next_layout_y(24);
    int x = layout_margin;
    
    HWND hwnd = CreateWindowExA(
        0, "STATIC", text,
        WS_VISIBLE | WS_CHILD | SS_LEFT,
        x, y, 350, 24,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
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

gui_widget gui_button(const char* label, gui_callback on_click) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    int height = 36;
    int width = 100;
    int y, x;
    
    if (layout_is_horizontal) {
        y = layout_row_start_y;
        x = get_next_layout_x(width);
        // Track max height in this row
        if (height > layout_row_height) layout_row_height = height;
    } else {
        y = get_next_layout_y(height);
        x = layout_margin;
    }
    
    HWND hwnd = CreateWindowExA(
        0, "BUTTON", label,
        WS_VISIBLE | WS_CHILD | BS_PUSHBUTTON,
        x, y, width, height,
        current_window->hwnd, (HMENU)get_next_id(), g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_BUTTON;
    w->id = get_next_id();
    w->callback = on_click;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

gui_widget gui_entry(const char* placeholder) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    int height = 28;
    int y = get_next_layout_y(height);
    int x = layout_margin;
    
    HWND hwnd = CreateWindowExA(
        WS_EX_CLIENTEDGE, "EDIT", "",
        WS_VISIBLE | WS_CHILD | ES_AUTOHSCROLL,
        x, y, 350, height,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    // Set placeholder text (gray)
    if (placeholder) SendMessage(hwnd, EM_SETCUEBANNER, TRUE, (LPARAM)placeholder);
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_ENTRY;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

char* gui_get_text(gui_widget entry) {
    int len = GetWindowTextLength((HWND)entry) + 1;
    char* buffer = (char*)malloc(len);
    if (buffer) {
        GetWindowTextA((HWND)entry, buffer, len);
    }
    return buffer;
}

void gui_set_text(gui_widget widget, const char* text) {
    SetWindowTextA((HWND)widget, text);
}

void gui_entry_set_text(gui_widget entry, const char* text) {
    SetWindowTextA((HWND)entry, text);
}

gui_widget gui_entry_multi(const char* placeholder) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    HWND hwnd = CreateWindowExA(
        WS_EX_CLIENTEDGE, "EDIT", "",
        WS_VISIBLE | WS_CHILD | ES_MULTILINE | ES_AUTOVSCROLL | WS_VSCROLL,
        10, 10, 400, 200,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_TEXTAREA;
    w->id = get_next_id();
    w->callback = NULL;
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

gui_widget gui_check(const char* label) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    int height = 28;
    int y = get_next_layout_y(height);
    int x = layout_margin;
    
    HWND hwnd = CreateWindowExA(
        0, "BUTTON", label,
        WS_VISIBLE | WS_CHILD | BS_AUTOCHECKBOX,
        x, y, 300, height,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_CHECKBOX;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

bool gui_is_checked(gui_widget checkbox) {
    return SendMessage((HWND)checkbox, BM_GETCHECK, 0, 0) == BST_CHECKED;
}

void gui_set_checked(gui_widget checkbox, bool checked) {
    SendMessage((HWND)checkbox, BM_SETCHECK, checked ? BST_CHECKED : BST_UNCHECKED, 0);
}

gui_widget gui_slider(double min, double max) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    int height = 35;
    int y = get_next_layout_y(height);
    int x = layout_margin;
    
    HWND hwnd = CreateWindowExA(
        0, TRACKBAR_CLASSA, "",
        WS_VISIBLE | WS_CHILD | TBS_HORZ | TBS_AUTOTICKS,
        x, y, 400, height,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    SendMessage(hwnd, TBM_SETRANGE, TRUE, MAKELPARAM((int)min, (int)max));
    SendMessage(hwnd, TBM_SETTICFREQ, 10, 0); // Add tick marks
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_SLIDER;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

double gui_get_value(gui_widget slider) {
    if (IsWindow((HWND)slider)) {
        return (double)SendMessage((HWND)slider, TBM_GETPOS, 0, 0);
    }
    return 0.0;
}

void gui_set_value(gui_widget widget, double value) {
    if (IsWindow((HWND)widget)) {
        LONG_PTR style = GetWindowLongPtr((HWND)widget, GWL_STYLE);
        if (style & TBS_HORZ) {
            // Slider
            SendMessage((HWND)widget, TBM_SETPOS, TRUE, (LPARAM)(int)value);
        } else {
            // Progress bar
            SendMessage((HWND)widget, PBM_SETPOS, (int)(value * 100), 0);
        }
    }
}

gui_widget gui_progress(void) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    int height = 28;
    int y = get_next_layout_y(height);
    int x = layout_margin;
    
    HWND hwnd = CreateWindowExA(
        0, PROGRESS_CLASSA, "",
        WS_VISIBLE | WS_CHILD | PBS_SMOOTH,
        x, y, 300, height,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_PROGRESS;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

// ============================================================================
// Additional Widgets
// ============================================================================

// Radio button groups
static int radio_group_id = 0;
static HWND radio_groups[32] = {0}; // Track first HWND per group

gui_widget gui_radio(const char* label, int group) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    int height = 24;
    int y = get_next_layout_y(height);
    int x = layout_margin;
    
    // First radio in group starts a new group
    DWORD style = WS_VISIBLE | WS_CHILD | BS_AUTORADIOBUTTON;
    if (radio_groups[group] == NULL) {
        style |= WS_GROUP; // Start of group
    }
    
    HWND hwnd = CreateWindowExA(
        0, "BUTTON", label,
        style,
        x, y, 250, height,
        current_window->hwnd, (HMENU)get_next_id(), g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    // Track first radio in group
    if (radio_groups[group] == NULL) {
        radio_groups[group] = hwnd;
    }
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_CHECKBOX; // Reuse checkbox type
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

gui_widget gui_spin(double min, double max, double step) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    int height = 28;
    int y = get_next_layout_y(height);
    int x = layout_margin;
    
    // Create up-down control with buddy edit
    HWND hwndEdit = CreateWindowExA(
        WS_EX_CLIENTEDGE, "EDIT", "",
        WS_VISIBLE | WS_CHILD | ES_NUMBER | ES_RIGHT,
        x, y, 100, height,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    HWND hwndSpin = CreateWindowExA(
        0, UPDOWN_CLASSA, "",
        WS_VISIBLE | WS_CHILD | UDS_SETBUDDYINT | UDS_ALIGNRIGHT,
        x + 100, y, 30, height,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwndSpin) return GUI_NULL;
    
    // Set buddy and range
    SendMessage(hwndSpin, UDM_SETBUDDY, (WPARAM)hwndEdit, 0);
    SendMessage(hwndSpin, UDM_SETRANGE, 0, MAKELPARAM((int)max, (int)min));
    SendMessage(hwndSpin, UDM_SETPOS, 0, 0);
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwndSpin; // Store spin control
    w->type = WIDGET_SLIDER; // Reuse slider type
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    return (gui_widget)hwndSpin;
}

gui_widget gui_list(const char* items[], int count) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    int height = 120;
    int y = get_next_layout_y(height);
    int x = layout_margin;
    
    HWND hwnd = CreateWindowExA(
        WS_EX_CLIENTEDGE, "LISTBOX", "",
        WS_VISIBLE | WS_CHILD | LBS_NOTIFY | WS_VSCROLL | LBS_NOINTEGRALHEIGHT,
        x, y, 250, height,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    for (int i = 0; i < count; i++) {
        SendMessageA(hwnd, LB_ADDSTRING, 0, (LPARAM)items[i]);
    }
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_NONE;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

int gui_get_list_selected(gui_widget w) {
    return (int)SendMessageA((HWND)w, LB_GETCURSEL, 0, 0);
}

void gui_set_list_selected(gui_widget w, int index) {
    SendMessageA((HWND)w, LB_SETCURSEL, index, 0);
}

void gui_list_add(gui_widget w, const char* item) {
    SendMessageA((HWND)w, LB_ADDSTRING, 0, (LPARAM)item);
}

gui_widget gui_group(const char* label) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    int height = 150;
    int y = get_next_layout_y(height);
    int x = layout_margin;
    
    HWND hwnd = CreateWindowExA(
        0, "BUTTON", label,
        WS_VISIBLE | WS_CHILD | BS_GROUPBOX,
        x, y, 280, height,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_NONE;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

void gui_group_add(gui_widget group, gui_widget widget) {
    // Reparent widget to group box
    SetParent((HWND)widget, (HWND)group);
}

gui_widget gui_color_button(uint8_t r, uint8_t g, uint8_t b) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    int height = 30;
    int width = 50;
    int y = get_next_layout_y(height);
    int x = layout_margin;
    
    HWND hwnd = CreateWindowExA(
        0, "BUTTON", "",
        WS_VISIBLE | WS_CHILD | BS_PUSHBUTTON,
        x, y, width, height,
        current_window->hwnd, (HMENU)get_next_id(), g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    // Set button color via custom draw (simplified - just store color)
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_BUTTON;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    w->color_r = r;
    w->color_g = g;
    w->color_b = b;
    w->color_a = 255;
    
    return (gui_widget)hwnd;
}

void gui_get_color(gui_widget w, uint8_t* r, uint8_t* g, uint8_t* b) {
    NativeWidget* widget = find_widget_by_hwnd((HWND)w);
    if (widget) {
        *r = widget->color_r;
        *g = widget->color_g;
        *b = widget->color_b;
    }
}


// ============================================================================
// Simplified API (matches GTK4 API)
// ============================================================================

void gui_start(const char* title, int width, int height) {
    gui_window win = gui_window_create_native(title, width, height);
    reset_layout(); // Reset layout for new window
    gui_window_show_native(win);
}

void gui_init(void) {
    register_window_class();
}

void gui_run(void) {
    gui_run_native();
}

void gui_run_nonblock(void) {
    // Windows GUI is already non-blocking in this implementation
}

void gui_update(void) {
    MSG msg;
    while (PeekMessage(&msg, NULL, 0, 0, PM_REMOVE)) {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }
}

bool gui_is_running(void) {
    return current_window ? current_window->running : false;
}

void gui_quit(void) {
    gui_quit_native();
}

void gui_set_title(const char* title) {
    if (current_window && current_window->hwnd) {
        SetWindowTextA(current_window->hwnd, title);
    }
}

void gui_set_size(int width, int height) {
    if (current_window && current_window->hwnd) {
        SetWindowPos(current_window->hwnd, NULL, 0, 0, width, height, SWP_NOMOVE | SWP_NOZORDER);
    }
}

void gui_set_resizable(bool resizable) {
    if (current_window && current_window->hwnd) {
        LONG_PTR style = GetWindowLongPtr(current_window->hwnd, GWL_STYLE);
        if (resizable) {
            style |= WS_THICKFRAME | WS_MAXIMIZEBOX;
        } else {
            style &= ~(WS_THICKFRAME | WS_MAXIMIZEBOX);
        }
        SetWindowLongPtr(current_window->hwnd, GWL_STYLE, style);
    }
}

gui_widget gui_button_ok(const char* label, gui_callback on_click) {
    return gui_button(label, on_click);  // Same as regular button for now
}

gui_widget gui_entry_secret(const char* placeholder) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    HWND hwnd = CreateWindowExA(
        WS_EX_CLIENTEDGE, "EDIT", "",
        WS_VISIBLE | WS_CHILD | ES_AUTOHSCROLL | ES_PASSWORD,
        10, 10, 200, 25,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_ENTRY;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

gui_widget gui_select(const char* options[], int count) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    int height = 100; // Dropdown height
    int y = get_next_layout_y(24); // Position at row height
    int x = layout_margin;
    
    HWND hwnd = CreateWindowExA(
        0, "COMBOBOX", "",
        WS_VISIBLE | WS_CHILD | CBS_DROPDOWNLIST | WS_VSCROLL,
        x, y, 250, height,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    for (int i = 0; i < count; i++) {
        SendMessageA(hwnd, CB_ADDSTRING, 0, (LPARAM)options[i]);
    }
    SendMessageA(hwnd, CB_SETCURSEL, 0, 0);
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_NONE;  // Dropdown
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

int gui_get_selected(gui_widget dropdown) {
    return (int)SendMessageA((HWND)dropdown, CB_GETCURSEL, 0, 0);
}

void gui_set_selected(gui_widget dropdown, int index) {
    SendMessageA((HWND)dropdown, CB_SETCURSEL, index, 0);
}

gui_widget gui_separator(void) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    int height = 2;
    int y = get_next_layout_y(height) + 8; // Add extra padding around separator
    int x = layout_margin;
    
    HWND hwnd = CreateWindowExA(
        0, "STATIC", "",
        WS_VISIBLE | WS_CHILD | SS_ETCHEDHORZ,
        x, y, 350, height,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    layout_next_y += 8; // Extra padding after separator
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_NONE;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    return (gui_widget)hwnd;
}

gui_widget gui_image(const char* path) {
    // Not implemented for native Windows - placeholder
    (void)path;
    return GUI_NULL;
}

gui_widget gui_spinner(void) {
    // Not implemented for native Windows - placeholder
    return GUI_NULL;
}

void gui_spinner_start(gui_widget w) {
    (void)w;
}

void gui_spinner_stop(gui_widget w) {
    (void)w;
}

gui_container gui_scroll(gui_widget content) {
    (void)content;
    return gui_vbox();  // Placeholder
}

gui_container gui_tabs(void) {
    return gui_vbox();  // Placeholder
}

void gui_tab_add(gui_container tabs, const char* label, gui_widget content) {
    (void)tabs;
    (void)label;
    (void)content;
}

// Dialogs
void gui_alert_info(const char* message) {
    MessageBoxA(current_window ? current_window->hwnd : NULL, message, "Info", MB_OK | MB_ICONINFORMATION);
}

void gui_alert_error(const char* message) {
    MessageBoxA(current_window ? current_window->hwnd : NULL, message, "Error", MB_OK | MB_ICONERROR);
}

void gui_alert_warn(const char* message) {
    MessageBoxA(current_window ? current_window->hwnd : NULL, message, "Warning", MB_OK | MB_ICONWARNING);
}

void gui_confirm(const char* message, gui_callback on_result) {
    int result = MessageBoxA(current_window ? current_window->hwnd : NULL, message, "Confirm", MB_YESNO | MB_ICONQUESTION);
    gui_event event = {0};
    event.type = GUI_CLICK;
    event.checked = (result == IDYES);
    if (on_result) on_result(event);
}

void gui_pick_file(gui_callback on_result) {
    char filename[MAX_PATH] = {0};
    OPENFILENAMEA ofn = {0};
    ofn.lStructSize = sizeof(ofn);
    ofn.hwndOwner = current_window ? current_window->hwnd : NULL;
    ofn.lpstrFile = filename;
    ofn.nMaxFile = MAX_PATH;
    ofn.Flags = OFN_PATHMUSTEXIST | OFN_FILEMUSTEXIST;
    
    if (GetOpenFileNameA(&ofn)) {
        gui_event event = {0};
        event.type = GUI_SELECT;
        event.text = _strdup(filename);
        if (on_result) on_result(event);
    }
}

void gui_save_file(const char* default_name, gui_callback on_result) {
    char filename[MAX_PATH] = {0};
    if (default_name) strncpy(filename, default_name, MAX_PATH - 1);
    
    OPENFILENAMEA ofn = {0};
    ofn.lStructSize = sizeof(ofn);
    ofn.hwndOwner = current_window ? current_window->hwnd : NULL;
    ofn.lpstrFile = filename;
    ofn.nMaxFile = MAX_PATH;
    ofn.Flags = OFN_PATHMUSTEXIST | OFN_OVERWRITEPROMPT;
    
    if (GetSaveFileNameA(&ofn)) {
        gui_event event = {0};
        event.type = GUI_SELECT;
        event.text = _strdup(filename);
        if (on_result) on_result(event);
    }
}

void gui_pick_folder(gui_callback on_result) {
    // Simplified - use file dialog
    gui_pick_file(on_result);
}

void gui_free(char* str) {
    free(str);
}

char* gui_clipboard_get(void) {
    if (!OpenClipboard(NULL)) return _strdup("");
    HANDLE hData = GetClipboardData(CF_TEXT);
    if (!hData) {
        CloseClipboard();
        return _strdup("");
    }
    char* text = _strdup((char*)GlobalLock(hData));
    GlobalUnlock(hData);
    CloseClipboard();
    return text ? text : _strdup("");
}

void gui_clipboard_set(const char* text) {
    if (!text) return;
    if (!OpenClipboard(NULL)) return;
    EmptyClipboard();
    HGLOBAL hMem = GlobalAlloc(GMEM_MOVEABLE, strlen(text) + 1);
    if (hMem) {
        memcpy(GlobalLock(hMem), text, strlen(text) + 1);
        GlobalUnlock(hMem);
        SetClipboardData(CF_TEXT, hMem);
    }
    CloseClipboard();
}

void gui_focus(gui_widget w) {
    SetFocus((HWND)w);
}

void gui_hide(gui_widget w) {
    ShowWindow((HWND)w, SW_HIDE);
}

void gui_show(gui_widget w) {
    ShowWindow((HWND)w, SW_SHOW);
}

void gui_enable(gui_widget w) {
    EnableWindow((HWND)w, TRUE);
}

void gui_disable(gui_widget w) {
    EnableWindow((HWND)w, FALSE);
}

// ============================================================================
// Shape Widgets (Custom Drawing)
// ============================================================================

gui_widget gui_rectangle_create_native(uint8_t r, uint8_t g, uint8_t b, uint8_t a) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = NULL; // No native window handle for shapes
    w->type = WIDGET_RECTANGLE;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    w->color_r = r;
    w->color_g = g;
    w->color_b = b;
    w->color_a = a;
    w->shape_x = 0;
    w->shape_y = 0;
    w->shape_w = 50;
    w->shape_h = 50;
    
    return (gui_widget)w;
}

gui_widget gui_circle_create_native(uint8_t r, uint8_t g, uint8_t b, uint8_t a) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = NULL;
    w->type = WIDGET_CIRCLE;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    w->color_r = r;
    w->color_g = g;
    w->color_b = b;
    w->color_a = a;
    w->shape_x = 0;
    w->shape_y = 0;
    w->shape_w = 50;
    w->shape_h = 50;
    
    return (gui_widget)w;
}

gui_widget gui_line_create_native(float x1, float y1, float x2, float y2) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = NULL;
    w->type = WIDGET_LINE;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    w->color_r = 0;
    w->color_g = 0;
    w->color_b = 0;
    w->color_a = 255;
    w->line_x1 = (int)x1;
    w->line_y1 = (int)y1;
    w->line_x2 = (int)x2;
    w->line_y2 = (int)y2;
    
    return (gui_widget)w;
}

void gui_line_set_color_native(gui_widget line, uint8_t r, uint8_t g, uint8_t b, uint8_t a) {
    NativeWidget* w = (NativeWidget*)line;
    if (w && w->type == WIDGET_LINE) {
        w->color_r = r;
        w->color_g = g;
        w->color_b = b;
        w->color_a = a;
    }
}

// ============================================================================
// Layout Containers
// ============================================================================

gui_container gui_vbox(void) {
    if (container_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    NativeContainer* c = &containers[container_count++];
    c->hwnd = current_window ? current_window->hwnd : NULL;
    c->widgets = NULL;
    c->widget_count = 0;
    c->id = get_next_id();
    
    return (gui_container)c;
}

gui_container gui_hbox(void) {
    layout_is_horizontal = true;
    layout_row_start_y = layout_next_y; // Remember row start
    layout_current_x = layout_margin; // Use current margin
    layout_row_height = 32; // Reset row height
    gui_container c = gui_vbox();
    return c;
}

void gui_end_row(void) {
    if (layout_is_horizontal) {
        // Advance Y by the max height in this row
        layout_next_y = layout_row_start_y + layout_row_height + layout_spacing + 4;
        layout_is_horizontal = false;
        layout_current_x = layout_margin;
        layout_row_height = 32; // Reset for next row
    }
}

void gui_spacing(int pixels) {
    layout_next_y += pixels;
}

void gui_set_spacing(int pixels) {
    layout_spacing = pixels;
}

void gui_set_margin(int pixels) {
    layout_margin = pixels;
    // Also update current_x if we haven't started laying out yet
    if (layout_next_y <= layout_margin) {
        layout_current_x = layout_margin;
        layout_next_y = layout_margin;
    }
}

gui_widget gui_header(const char* text) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    // Add extra space before header
    layout_next_y += 8;
    
    int y = get_next_layout_y(28);
    int x = layout_margin;
    
    HWND hwnd = CreateWindowExA(
        0, "STATIC", text,
        WS_VISIBLE | WS_CHILD | SS_LEFT,
        x, y, 400, 28,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    // Make font bold
    HFONT hFont = (HFONT)GetStockObject(DEFAULT_GUI_FONT);
    LOGFONTA lf = {0};
    GetObjectA(hFont, sizeof(lf), &lf);
    lf.lfWeight = FW_BOLD;
    lf.lfHeight = -18; // 14pt
    HFONT hBoldFont = CreateFontIndirectA(&lf);
    SendMessage(hwnd, WM_SETFONT, (WPARAM)hBoldFont, TRUE);
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_LABEL;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    // Add space after header
    layout_next_y += 4;
    
    return (gui_widget)hwnd;
}

gui_widget gui_subheader(const char* text) {
    if (!current_window || widget_count >= MAX_WIDGETS) {
        return GUI_NULL;
    }
    
    layout_next_y += 4;
    
    int y = get_next_layout_y(22);
    int x = layout_margin;
    
    HWND hwnd = CreateWindowExA(
        0, "STATIC", text,
        WS_VISIBLE | WS_CHILD | SS_LEFT,
        x, y, 400, 22,
        current_window->hwnd, NULL, g_hInstance, NULL
    );
    
    if (!hwnd) return GUI_NULL;
    
    // Make font semi-bold
    HFONT hFont = (HFONT)GetStockObject(DEFAULT_GUI_FONT);
    LOGFONTA lf = {0};
    GetObjectA(hFont, sizeof(lf), &lf);
    lf.lfWeight = FW_SEMIBOLD;
    lf.lfHeight = -15; // 12pt
    HFONT hSemiFont = CreateFontIndirectA(&lf);
    SendMessage(hwnd, WM_SETFONT, (WPARAM)hSemiFont, TRUE);
    
    NativeWidget* w = &widgets[widget_count++];
    w->hwnd = hwnd;
    w->type = WIDGET_LABEL;
    w->id = get_next_id();
    w->callback = NULL;
    w->next = NULL;
    
    layout_next_y += 2;
    
    return (gui_widget)hwnd;
}

gui_container gui_grid(int columns) {
    (void)columns;
    return gui_vbox();
}

void gui_add(gui_widget widget) {
    if (!current_window || !widget) return;
    
    // Just show the widget - it's already positioned by auto-layout
    ShowWindow((HWND)widget, SW_SHOW);
    
    // If we were in horizontal mode, advance Y after adding last widget in row
    // (This is a simplification - real hbox would track this better)
}

void gui_add_to(gui_container container, gui_widget widget) {
    if (!container || !widget) return;
    
    NativeContainer* c = (NativeContainer*)container;
    NativeWidget* w = NULL;
    
    // Check if widget is a native HWND (regular widget) or a NativeWidget pointer (shape)
    if (IsWindow((HWND)widget)) {
        // Regular widget - find the NativeWidget by HWND
        w = find_widget_by_hwnd((HWND)widget);
    } else {
        // Shape widget - widget IS the NativeWidget pointer
        w = (NativeWidget*)widget;
    }
    
    if (!w) return;
    
    // Add widget to container's list (but don't reposition - auto-layout handles it)
    w->next = c->widgets;
    c->widgets = w;
    c->widget_count++;
    
    // Just show the widget
    if (w->hwnd) {
        ShowWindow(w->hwnd, SW_SHOW);
    }
    
    // Trigger repaint for shapes
    if (c->hwnd) {
        InvalidateRect(c->hwnd, NULL, TRUE);
    }
}

// ============================================================================
// Widget Management
// ============================================================================

void gui_refresh(gui_widget widget) {
    NativeWidget* w = (NativeWidget*)widget;
    if (w && w->hwnd) {
        InvalidateRect(w->hwnd, NULL, TRUE);
        UpdateWindow(w->hwnd);
    }
}

void gui_resize(gui_widget widget, float width, float height) {
    NativeWidget* w = (NativeWidget*)widget;
    if (!w) return;
    
    if (w->hwnd) {
        SetWindowPos(w->hwnd, NULL, 0, 0, (int)width, (int)height, SWP_NOMOVE | SWP_NOZORDER);
    } else {
        // Shape widget
        w->shape_w = (int)width;
        w->shape_h = (int)height;
    }
}

void gui_move(gui_widget widget, float x, float y) {
    NativeWidget* w = (NativeWidget*)widget;
    if (!w) return;
    
    if (w->hwnd) {
        SetWindowPos(w->hwnd, NULL, (int)x, (int)y, 0, 0, SWP_NOSIZE | SWP_NOZORDER);
    } else {
        // Shape widget
        w->shape_x = (int)x;
        w->shape_y = (int)y;
    }
}

bool gui_is_enabled(gui_widget widget) {
    return IsWindowEnabled((HWND)widget) != 0;
}

#endif // _WIN32
