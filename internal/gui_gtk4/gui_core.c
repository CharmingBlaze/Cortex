// gui_core.c - GTK4 Core Module
//
// Handles application lifecycle, main loop, and window management.
// Simplified API with auto-managed main window.

#include "gui_gtk4_internal.h"

// ============================================================================
// Global State
// ============================================================================

GUIRegistry g_gui_registry = {0};

// Additional state for simplified API (accessible from other modules)
GtkWidget *g_main_window = NULL;
GtkWidget *g_main_vbox = NULL;
static bool g_initialized = false;

// ============================================================================
// Registry Management
// ============================================================================

void gui_registry_init(void) {
    memset(&g_gui_registry, 0, sizeof(GUIRegistry));
    g_gui_registry.next_handle = 1;
    g_gui_registry.main_window = GUI_NULL;
    g_gui_registry.is_running = false;
    g_gui_registry.main_loop = NULL;
}

void gui_registry_cleanup(void) {
    for (int i = 0; i < GUI_MAX_WIDGETS; i++) {
        g_gui_registry.widgets[i].in_use = false;
    }
}

// ============================================================================
// Simplified Application Lifecycle
// ============================================================================

static void on_window_close(GtkWindow *window, gpointer user_data) {
    (void)window;
    (void)user_data;
    gui_quit();
}

void gui_start(const char* title, int width, int height) {
    // Initialize GTK
    gtk_init();
    
    // Initialize registry
    gui_registry_init();
    
    // Create main window
    g_main_window = gtk_window_new();
    gtk_window_set_title(GTK_WINDOW(g_main_window), title);
    gtk_window_set_default_size(GTK_WINDOW(g_main_window), width, height);
    
    // Create main vbox as default container
    g_main_vbox = gtk_box_new(GTK_ORIENTATION_VERTICAL, 8);
    gtk_widget_set_margin_start(g_main_vbox, 12);
    gtk_widget_set_margin_end(g_main_vbox, 12);
    gtk_widget_set_margin_top(g_main_vbox, 12);
    gtk_widget_set_margin_bottom(g_main_vbox, 12);
    
    gtk_window_set_child(GTK_WINDOW(g_main_window), g_main_vbox);
    
    // Connect close handler
    g_signal_connect(g_main_window, "close-request", G_CALLBACK(on_window_close), NULL);
    
    // Register window
    g_gui_registry.main_window = gui_register_widget(g_main_window, WIDGET_TYPE_WINDOW);
    
    g_initialized = true;
}

void gui_init(void) {
    if (!g_initialized) {
        gtk_init();
        gui_registry_init();
        g_initialized = true;
    }
}

void gui_run(void) {
    if (!g_initialized) {
        gui_start("Cortex App", 800, 600);
    }
    
    // Show main window
    if (g_main_window) {
        gtk_window_present(GTK_WINDOW(g_main_window));
    }
    
    g_gui_registry.is_running = true;
    g_gui_registry.main_loop = g_main_loop_new(NULL, TRUE);
    
    g_main_loop_run(g_gui_registry.main_loop);
    
    g_main_loop_unref(g_gui_registry.main_loop);
    g_gui_registry.main_loop = NULL;
}

// Non-blocking mode for integration with raylib/SDL/OpenGL
void gui_run_nonblock(void) {
    if (!g_initialized) {
        gui_start("Cortex App", 800, 600);
    }
    
    // Show main window
    if (g_main_window) {
        gtk_window_present(GTK_WINDOW(g_main_window));
    }
    
    g_gui_registry.is_running = true;
    // Don't create/run main loop - caller will call gui_update()
}

void gui_update(void) {
    if (!g_gui_registry.is_running) return;
    
    // Process pending GTK events without blocking
    GMainContext *context = g_main_context_default();
    while (g_main_context_pending(context)) {
        g_main_context_iteration(context, FALSE);
    }
}

bool gui_is_running(void) {
    return g_gui_registry.is_running;
}

void gui_quit(void) {
    if (!g_gui_registry.is_running) return;
    
    g_gui_registry.is_running = false;
    
    if (g_gui_registry.main_loop) {
        g_main_loop_quit(g_gui_registry.main_loop);
    }
    
    gui_registry_cleanup();
}

// ============================================================================
// Main Window Helpers
// ============================================================================

void gui_add(gui_widget w) {
    if (!g_main_vbox) {
        gui_start("Cortex App", 800, 600);
    }
    
    GtkWidget *widget = gui_get_widget(w);
    if (widget && g_main_vbox) {
        gtk_box_append(GTK_BOX(g_main_vbox), widget);
    }
}

void gui_add_to(gui_container c, gui_widget w) {
    GtkWidget *container = gui_get_widget(c);
    GtkWidget *widget = gui_get_widget(w);
    
    if (container && widget) {
        if (GTK_IS_BOX(container)) {
            gtk_box_append(GTK_BOX(container), widget);
        } else if (GTK_IS_GRID(container)) {
            int pos = GPOINTER_TO_INT(g_object_get_data(G_OBJECT(container), "pos"));
            int cols = GPOINTER_TO_INT(g_object_get_data(G_OBJECT(container), "columns"));
            if (cols <= 0) cols = 1;
            gtk_grid_attach(GTK_GRID(container), widget, pos % cols, pos / cols, 1, 1);
            g_object_set_data(G_OBJECT(container), "pos", GINT_TO_POINTER(pos + 1));
        } else if (GTK_IS_NOTEBOOK(container)) {
            GtkWidget *label = gtk_label_new("Tab");
            gtk_notebook_append_page(GTK_NOTEBOOK(container), widget, label);
        }
    }
}

void gui_set_title(const char* title) {
    if (g_main_window && title) {
        gtk_window_set_title(GTK_WINDOW(g_main_window), title);
    }
}

void gui_set_size(int width, int height) {
    if (g_main_window) {
        gtk_window_set_default_size(GTK_WINDOW(g_main_window), width, height);
    }
}

void gui_set_resizable(bool resizable) {
    if (g_main_window) {
        gtk_window_set_resizable(GTK_WINDOW(g_main_window), resizable);
    }
}

// ============================================================================
// Utility Functions
// ============================================================================

void gui_free(char* str) {
    g_free(str);
}

char* gui_clipboard_get(void) {
    GdkDisplay *display = gdk_display_get_default();
    if (!display) return g_strdup("");
    
    GdkClipboard *clipboard = gdk_display_get_clipboard(display);
    if (!clipboard) return g_strdup("");
    
    // Note: GTK4 clipboard is async, returning empty for sync API
    return g_strdup("");
}

void gui_clipboard_set(const char* text) {
    if (!text) return;
    
    GdkDisplay *display = gdk_display_get_default();
    if (!display) return;
    
    GdkClipboard *clipboard = gdk_display_get_clipboard(display);
    if (!clipboard) return;
    
    gdk_clipboard_set_text(clipboard, text);
}
