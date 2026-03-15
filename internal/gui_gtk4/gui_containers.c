// gui_containers.c - GTK4 Container Module
//
// Simple layout containers for Cortex GUI.

#include "gui_gtk4_internal.h"

// ============================================================================
// Box Containers
// ============================================================================

gui_container gui_vbox(void) {
    GtkWidget *box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 8);
    return gui_register_widget(box, WIDGET_TYPE_CONTAINER);
}

gui_container gui_hbox(void) {
    GtkWidget *box = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 8);
    return gui_register_widget(box, WIDGET_TYPE_CONTAINER);
}

// ============================================================================
// Grid Container
// ============================================================================

gui_container gui_grid(int columns) {
    GtkWidget *grid = gtk_grid_new();
    gtk_grid_set_column_homogeneous(GTK_GRID(grid), TRUE);
    gtk_grid_set_row_spacing(GTK_GRID(grid), 6);
    gtk_grid_set_column_spacing(GTK_GRID(grid), 6);
    g_object_set_data(G_OBJECT(grid), "columns", GINT_TO_POINTER(columns));
    g_object_set_data(G_OBJECT(grid), "pos", GINT_TO_POINTER(0));
    return gui_register_widget(grid, WIDGET_TYPE_CONTAINER);
}

// ============================================================================
// Scroll Container
// ============================================================================

gui_container gui_scroll(gui_widget content) {
    GtkWidget *scroll = gtk_scrolled_window_new();
    gtk_scrolled_window_set_policy(GTK_SCROLLED_WINDOW(scroll),
                                   GTK_POLICY_AUTOMATIC,
                                   GTK_POLICY_AUTOMATIC);
    
    GtkWidget *content_w = gui_get_widget(content);
    if (content_w) {
        gtk_scrolled_window_set_child(GTK_SCROLLED_WINDOW(scroll), content_w);
    }
    
    return gui_register_widget(scroll, WIDGET_TYPE_CONTAINER);
}

// ============================================================================
// Tab Container (Notebook)
// ============================================================================

gui_container gui_tabs(void) {
    GtkWidget *notebook = gtk_notebook_new();
    return gui_register_widget(notebook, WIDGET_TYPE_CONTAINER);
}

void gui_tab_add(gui_container tabs, const char* label, gui_widget content) {
    GtkWidget *notebook = gui_get_widget(tabs);
    GtkWidget *content_w = gui_get_widget(content);
    
    if (notebook && content_w) {
        GtkWidget *label_w = gtk_label_new(label);
        gtk_notebook_append_page(GTK_NOTEBOOK(notebook), content_w, label_w);
    }
}
