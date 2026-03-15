// gui_dialogs.c - GTK4 Dialogs Module
//
// Simple dialog functions for Cortex GUI.

#include "gui_gtk4_internal.h"

// ============================================================================
// Simple Alert Dialogs
// ============================================================================

void gui_alert_info(const char* message) {
    GtkWidget *dialog = gtk_message_dialog_new(
        NULL,
        GTK_DIALOG_MODAL,
        GTK_MESSAGE_INFO,
        GTK_BUTTONS_OK,
        "%s", message
    );
    g_signal_connect(dialog, "response", G_CALLBACK(gtk_window_destroy), NULL);
    gtk_window_present(GTK_WINDOW(dialog));
}

void gui_alert_error(const char* message) {
    GtkWidget *dialog = gtk_message_dialog_new(
        NULL,
        GTK_DIALOG_MODAL,
        GTK_MESSAGE_ERROR,
        GTK_BUTTONS_OK,
        "%s", message
    );
    g_signal_connect(dialog, "response", G_CALLBACK(gtk_window_destroy), NULL);
    gtk_window_present(GTK_WINDOW(dialog));
}

void gui_alert_warn(const char* message) {
    GtkWidget *dialog = gtk_message_dialog_new(
        NULL,
        GTK_DIALOG_MODAL,
        GTK_MESSAGE_WARNING,
        GTK_BUTTONS_OK,
        "%s", message
    );
    g_signal_connect(dialog, "response", G_CALLBACK(gtk_window_destroy), NULL);
    gtk_window_present(GTK_WINDOW(dialog));
}

// ============================================================================
// Confirm Dialog
// ============================================================================

static void on_confirm_response(GtkDialog *dialog, int response_id, gpointer user_data) {
    gui_callback callback = (gui_callback)user_data;
    
    if (callback) {
        gui_event e = {
            .type = GUI_CLICK,
            .source = 0,
            .checked = (response_id == GTK_RESPONSE_OK || response_id == GTK_RESPONSE_YES),
        };
        callback(e);
    }
    
    gtk_window_destroy(GTK_WINDOW(dialog));
}

void gui_confirm(const char* message, gui_callback on_result) {
    GtkWidget *dialog = gtk_message_dialog_new(
        NULL,
        GTK_DIALOG_MODAL,
        GTK_MESSAGE_QUESTION,
        GTK_BUTTONS_YES_NO,
        "%s", message
    );
    
    g_signal_connect(dialog, "response", G_CALLBACK(on_confirm_response), on_result);
    gtk_window_present(GTK_WINDOW(dialog));
}

// ============================================================================
// File Dialogs
// ============================================================================

typedef struct {
    gui_callback callback;
} FileDialogData;

static void on_file_open_response(GtkFileDialog *dialog, GAsyncResult *result, gpointer user_data) {
    FileDialogData *data = (FileDialogData *)user_data;
    
    GFile *file = gtk_file_dialog_open_finish(dialog, result, NULL);
    
    if (data->callback) {
        char *path = file ? g_file_get_path(file) : NULL;
        
        gui_event e = {
            .type = GUI_SELECT,
            .source = 0,
            .text = path,
        };
        data->callback(e);
        
        g_free(path);
    }
    
    if (file) g_object_unref(file);
    g_object_unref(dialog);
    g_free(data);
}

void gui_pick_file(gui_callback on_result) {
    GtkFileDialog *dialog = gtk_file_dialog_new();
    gtk_file_dialog_set_title(dialog, "Open File");
    
    FileDialogData *data = g_new(FileDialogData, 1);
    data->callback = on_result;
    
    gtk_file_dialog_open(dialog, NULL, NULL, (GAsyncReadyCallback)on_file_open_response, data);
}

static void on_file_save_response(GtkFileDialog *dialog, GAsyncResult *result, gpointer user_data) {
    FileDialogData *data = (FileDialogData *)user_data;
    
    GFile *file = gtk_file_dialog_save_finish(dialog, result, NULL);
    
    if (data->callback) {
        char *path = file ? g_file_get_path(file) : NULL;
        
        gui_event e = {
            .type = GUI_SELECT,
            .source = 0,
            .text = path,
        };
        data->callback(e);
        
        g_free(path);
    }
    
    if (file) g_object_unref(file);
    g_object_unref(dialog);
    g_free(data);
}

void gui_save_file(const char* default_name, gui_callback on_result) {
    GtkFileDialog *dialog = gtk_file_dialog_new();
    gtk_file_dialog_set_title(dialog, "Save File");
    
    if (default_name) {
        gtk_file_dialog_set_initial_name(dialog, default_name);
    }
    
    FileDialogData *data = g_new(FileDialogData, 1);
    data->callback = on_result;
    
    gtk_file_dialog_save(dialog, NULL, NULL, (GAsyncReadyCallback)on_file_save_response, data);
}

static void on_folder_open_response(GtkFileDialog *dialog, GAsyncResult *result, gpointer user_data) {
    FileDialogData *data = (FileDialogData *)user_data;
    
    GFile *folder = gtk_file_dialog_select_folder_finish(dialog, result, NULL);
    
    if (data->callback) {
        char *path = folder ? g_file_get_path(folder) : NULL;
        
        gui_event e = {
            .type = GUI_SELECT,
            .source = 0,
            .text = path,
        };
        data->callback(e);
        
        g_free(path);
    }
    
    if (folder) g_object_unref(folder);
    g_object_unref(dialog);
    g_free(data);
}

void gui_pick_folder(gui_callback on_result) {
    GtkFileDialog *dialog = gtk_file_dialog_new();
    gtk_file_dialog_set_title(dialog, "Select Folder");
    
    FileDialogData *data = g_new(FileDialogData, 1);
    data->callback = on_result;
    
    gtk_file_dialog_select_folder(dialog, NULL, NULL, (GAsyncReadyCallback)on_folder_open_response, data);
}
