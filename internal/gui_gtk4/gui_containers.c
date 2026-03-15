// gui_containers.c - GTK4 Container Module
//
// Simple layout containers for Cortex GUI.

#include "gui_gtk4_internal.h"

// ============================================================================
// Layout State
// ============================================================================

static int layout_spacing = 8;
static int layout_margin = 10;

void gui_set_spacing(int spacing) {
    layout_spacing = spacing;
}

void gui_set_margin(int margin) {
    layout_margin = margin;
}

int gui_get_layout_spacing(void) {
    return layout_spacing;
}

// ============================================================================
// Spacing Widget
// ============================================================================

// External reference to main container from gui_core.c
extern GtkWidget *g_main_vbox;

void gui_spacing(int pixels) {
    // Add vertical spacing by adding an empty box with height
    GtkWidget *box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 0);
    gtk_widget_set_size_request(box, -1, pixels);
    if (g_main_vbox) {
        gtk_box_append(GTK_BOX(g_main_vbox), box);
    }
}

// ============================================================================
// Section Headers
// ============================================================================

gui_widget gui_header(const char* text) {
    GtkWidget *label = gtk_label_new(text);
    gtk_widget_set_margin_top(label, layout_margin);
    gtk_widget_set_margin_bottom(label, layout_spacing / 2);
    
    // Make bold using Pango markup
    char *markup = g_markup_printf_escaped("<b>%s</b>", text);
    gtk_label_set_markup(GTK_LABEL(label), markup);
    g_free(markup);
    
    return gui_register_widget(label, WIDGET_TYPE_LABEL);
}

gui_widget gui_subheader(const char* text) {
    GtkWidget *label = gtk_label_new(text);
    gtk_widget_set_margin_top(label, layout_spacing / 2);
    gtk_widget_set_margin_bottom(label, layout_spacing / 4);
    
    // Make slightly bold using Pango markup
    char *markup = g_markup_printf_escaped("<span weight='semibold'>%s</span>", text);
    gtk_label_set_markup(GTK_LABEL(label), markup);
    g_free(markup);
    
    return gui_register_widget(label, WIDGET_TYPE_LABEL);
}

// ============================================================================
// Box Containers
// ============================================================================

gui_container gui_vbox(void) {
    GtkWidget *box = gtk_box_new(GTK_ORIENTATION_VERTICAL, layout_spacing);
    gtk_widget_set_margin_start(box, layout_margin);
    gtk_widget_set_margin_end(box, layout_margin);
    gtk_widget_set_margin_top(box, layout_margin);
    gtk_widget_set_margin_bottom(box, layout_margin);
    return gui_register_widget(box, WIDGET_TYPE_CONTAINER);
}

gui_container gui_hbox(void) {
    GtkWidget *box = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, layout_spacing);
    return gui_register_widget(box, WIDGET_TYPE_CONTAINER);
}

void gui_end_row(void) {
    // No-op in GTK4 - hbox automatically handles layout
}

// ============================================================================
// Grid Container
// ============================================================================

gui_container gui_grid(int columns) {
    GtkWidget *grid = gtk_grid_new();
    gtk_grid_set_column_homogeneous(GTK_GRID(grid), TRUE);
    gtk_grid_set_row_spacing(GTK_GRID(grid), layout_spacing);
    gtk_grid_set_column_spacing(GTK_GRID(grid), layout_spacing);
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
