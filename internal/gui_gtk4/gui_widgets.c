// gui_widgets.c - GTK4 Widgets Module
//
// Simple, intuitive widget functions for Cortex GUI.

#include "gui_gtk4_internal.h"

// ============================================================================
// Label Widget
// ============================================================================

gui_widget gui_label(const char* text) {
    GtkWidget *label = gtk_label_new(text);
    gtk_label_set_wrap(GTK_LABEL(label), TRUE);
    return gui_register_widget(label, WIDGET_TYPE_LABEL);
}

void gui_set_text(gui_widget w, const char* text) {
    GtkWidget *widget = gui_get_widget(w);
    if (!widget) return;
    
    if (GTK_IS_LABEL(widget)) {
        gtk_label_set_text(GTK_LABEL(widget), text);
    } else if (GTK_IS_ENTRY(widget)) {
        gtk_editable_set_text(GTK_EDITABLE(widget), text);
    } else if (GTK_IS_BUTTON(widget)) {
        gtk_button_set_label(GTK_BUTTON(widget), text);
    }
}

char* gui_get_text(gui_widget w) {
    GtkWidget *widget = gui_get_widget(w);
    if (!widget) return g_strdup("");
    
    if (GTK_IS_LABEL(widget)) {
        return g_strdup(gtk_label_get_text(GTK_LABEL(widget)));
    } else if (GTK_IS_ENTRY(widget)) {
        return g_strdup(gtk_editable_get_text(GTK_EDITABLE(widget)));
    } else if (GTK_IS_BUTTON(widget)) {
        return g_strdup(gtk_button_get_label(GTK_BUTTON(widget)));
    }
    return g_strdup("");
}

// ============================================================================
// Button Widget
// ============================================================================

static void on_button_clicked(GtkButton *button, gpointer user_data) {
    int64_t handle = (int64_t)GPOINTER_TO_INT(user_data);
    gui_callback cb = gui_get_callback(handle);
    if (cb) {
        gui_event e = { .type = GUI_CLICK, .source = handle };
        cb(e);
    }
    (void)button;
}

gui_widget gui_button(const char* text, gui_callback on_click) {
    GtkWidget *button = gtk_button_new_with_label(text);
    int64_t handle = gui_register_widget(button, WIDGET_TYPE_BUTTON);
    
    if (on_click) {
        gui_set_callback(handle, on_click);
        g_signal_connect(button, "clicked", G_CALLBACK(on_button_clicked), 
                         GINT_TO_POINTER(handle));
    }
    
    return handle;
}

gui_widget gui_button_ok(const char* text, gui_callback on_click) {
    gui_widget btn = gui_button(text, on_click);
    GtkWidget *widget = gui_get_widget(btn);
    if (widget) {
        gtk_widget_add_css_class(widget, "suggested-action");
    }
    return btn;
}

// ============================================================================
// Entry Widget (Text Input)
// ============================================================================

gui_widget gui_entry(const char* placeholder) {
    GtkWidget *entry = gtk_entry_new();
    if (placeholder) {
        gtk_entry_set_placeholder_text(GTK_ENTRY(entry), placeholder);
    }
    return gui_register_widget(entry, WIDGET_TYPE_ENTRY);
}

gui_widget gui_entry_secret(const char* placeholder) {
    GtkWidget *entry = gtk_entry_new();
    gtk_entry_set_visibility(GTK_ENTRY(entry), FALSE);
    if (placeholder) {
        gtk_entry_set_placeholder_text(GTK_ENTRY(entry), placeholder);
    }
    return gui_register_widget(entry, WIDGET_TYPE_ENTRY);
}

gui_widget gui_entry_multi(const char* placeholder) {
    GtkWidget *view = gtk_text_view_new();
    gtk_text_view_set_wrap_mode(GTK_TEXT_VIEW(view), GTK_WRAP_WORD_CHAR);
    
    if (placeholder) {
        GtkTextBuffer *buffer = gtk_text_view_get_buffer(GTK_TEXT_VIEW(view));
        gtk_text_buffer_set_text(buffer, placeholder, -1);
    }
    
    return gui_register_widget(view, WIDGET_TYPE_TEXT_VIEW);
}

// ============================================================================
// Checkbox Widget
// ============================================================================

static void on_check_toggled(GtkCheckButton *check, gpointer user_data) {
    int64_t handle = (int64_t)GPOINTER_TO_INT(user_data);
    gui_callback cb = gui_get_callback(handle);
    if (cb) {
        gui_event e = { 
            .type = GUI_CHECK, 
            .source = handle,
            .checked = gtk_check_button_get_active(check)
        };
        cb(e);
    }
    (void)check;
}

gui_widget gui_check(const char* label) {
    GtkWidget *check = gtk_check_button_new_with_label(label);
    int64_t handle = gui_register_widget(check, WIDGET_TYPE_CHECK);
    g_signal_connect(check, "toggled", G_CALLBACK(on_check_toggled), 
                     GINT_TO_POINTER(handle));
    return handle;
}

bool gui_is_checked(gui_widget w) {
    GtkWidget *widget = gui_get_widget(w);
    if (widget && GTK_IS_CHECK_BUTTON(widget)) {
        return gtk_check_button_get_active(GTK_CHECK_BUTTON(widget));
    }
    return false;
}

void gui_set_checked(gui_widget w, bool checked) {
    GtkWidget *widget = gui_get_widget(w);
    if (widget && GTK_IS_CHECK_BUTTON(widget)) {
        gtk_check_button_set_active(GTK_CHECK_BUTTON(widget), checked);
    }
}

// ============================================================================
// Select (Dropdown) Widget
// ============================================================================

gui_widget gui_select(const char* options[], int count) {
    GtkStringList *list = gtk_string_list_new(NULL);
    for (int i = 0; i < count; i++) {
        gtk_string_list_append(list, options[i]);
    }
    
    GtkWidget *dropdown = gtk_drop_down_new(G_LIST_MODEL(list), NULL);
    return gui_register_widget(dropdown, WIDGET_TYPE_SELECT);
}

int gui_get_selected(gui_widget w) {
    GtkWidget *widget = gui_get_widget(w);
    if (widget && GTK_IS_DROP_DOWN(widget)) {
        return (int)gtk_drop_down_get_selected(GTK_DROP_DOWN(widget));
    }
    return -1;
}

void gui_set_selected(gui_widget w, int index) {
    GtkWidget *widget = gui_get_widget(w);
    if (widget && GTK_IS_DROP_DOWN(widget)) {
        gtk_drop_down_set_selected(GTK_DROP_DOWN(widget), (guint)index);
    }
}

// ============================================================================
// Slider Widget
// ============================================================================

gui_widget gui_slider(double min, double max) {
    GtkWidget *slider = gtk_scale_new_with_range(GTK_ORIENTATION_HORIZONTAL, min, max, 1.0);
    gtk_widget_set_hexpand(slider, TRUE);
    return gui_register_widget(slider, WIDGET_TYPE_SLIDER);
}

double gui_get_value(gui_widget w) {
    GtkWidget *widget = gui_get_widget(w);
    if (!widget) return 0.0;
    
    if (GTK_IS_RANGE(widget)) {
        return gtk_range_get_value(GTK_RANGE(widget));
    } else if (GTK_IS_PROGRESS_BAR(widget)) {
        return gtk_progress_bar_get_fraction(GTK_PROGRESS_BAR(widget));
    }
    return 0.0;
}

void gui_set_value(gui_widget w, double value) {
    GtkWidget *widget = gui_get_widget(w);
    if (!widget) return;
    
    if (GTK_IS_RANGE(widget)) {
        gtk_range_set_value(GTK_RANGE(widget), value);
    } else if (GTK_IS_PROGRESS_BAR(widget)) {
        gtk_progress_bar_set_fraction(GTK_PROGRESS_BAR(widget), value);
    }
}

// ============================================================================
// Progress Bar Widget
// ============================================================================

gui_widget gui_progress(void) {
    GtkWidget *progress = gtk_progress_bar_new();
    gtk_widget_set_hexpand(progress, TRUE);
    return gui_register_widget(progress, WIDGET_TYPE_PROGRESS);
}

// ============================================================================
// Visual Widgets
// ============================================================================

gui_widget gui_separator(void) {
    GtkWidget *sep = gtk_separator_new(GTK_ORIENTATION_HORIZONTAL);
    return gui_register_widget(sep, WIDGET_TYPE_SEPARATOR);
}

gui_widget gui_image(const char* path) {
    GtkWidget *image = path ? gtk_image_new_from_file(path) : gtk_image_new();
    return gui_register_widget(image, WIDGET_TYPE_IMAGE);
}

gui_widget gui_spinner(void) {
    GtkWidget *spinner = gtk_spinner_new();
    return gui_register_widget(spinner, WIDGET_TYPE_SPINNER);
}

void gui_spinner_start(gui_widget w) {
    GtkWidget *widget = gui_get_widget(w);
    if (widget && GTK_IS_SPINNER(widget)) {
        gtk_spinner_start(GTK_SPINNER(widget));
    }
}

void gui_spinner_stop(gui_widget w) {
    GtkWidget *widget = gui_get_widget(w);
    if (widget && GTK_IS_SPINNER(widget)) {
        gtk_spinner_stop(GTK_SPINNER(widget));
    }
}

// ============================================================================
// Radio Button Widget
// ============================================================================

// In GTK4, radio buttons are created by linking check buttons via set_group
// We store the first button of each group to link subsequent buttons
static GtkWidget *radio_group_leaders[16] = {NULL};  // Support up to 16 radio groups

gui_widget gui_radio(const char* label, int group) {
    GtkWidget *radio;
    if (group < 0 || group >= 16) group = 0;
    
    radio = gtk_check_button_new_with_label(label);
    
    if (radio_group_leaders[group] != NULL) {
        // Link to the first button in this group
        gtk_check_button_set_group(GTK_CHECK_BUTTON(radio), 
                                   GTK_CHECK_BUTTON(radio_group_leaders[group]));
    } else {
        // First button in group - save as leader
        radio_group_leaders[group] = radio;
    }
    
    return gui_register_widget(radio, WIDGET_TYPE_CHECK);
}

void gui_radio_reset_group(int group) {
    if (group >= 0 && group < 16) {
        radio_group_leaders[group] = NULL;
    }
}

// ============================================================================
// Spin Button Widget
// ============================================================================

gui_widget gui_spin(double min, double max, double step) {
    GtkWidget *spin = gtk_spin_button_new_with_range(min, max, step);
    gtk_widget_set_hexpand(spin, TRUE);
    return gui_register_widget(spin, WIDGET_TYPE_ENTRY);
}

// ============================================================================
// Widget State Functions
// ============================================================================

void gui_show(gui_widget w) {
    GtkWidget *widget = gui_get_widget(w);
    if (widget) gtk_widget_set_visible(widget, TRUE);
}

void gui_hide(gui_widget w) {
    GtkWidget *widget = gui_get_widget(w);
    if (widget) gtk_widget_set_visible(widget, FALSE);
}

void gui_enable(gui_widget w) {
    GtkWidget *widget = gui_get_widget(w);
    if (widget) gtk_widget_set_sensitive(widget, TRUE);
}

void gui_disable(gui_widget w) {
    GtkWidget *widget = gui_get_widget(w);
    if (widget) gtk_widget_set_sensitive(widget, FALSE);
}

void gui_focus(gui_widget w) {
    GtkWidget *widget = gui_get_widget(w);
    if (widget) gtk_widget_grab_focus(widget);
}
