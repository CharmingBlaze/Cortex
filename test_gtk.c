#include <gtk/gtk.h>

int main(int argc, char **argv) {
    gtk_init();
    
    GtkWidget *window = gtk_window_new();
    gtk_window_set_title(GTK_WINDOW(window), "GTK4 Test");
    gtk_window_set_default_size(GTK_WINDOW(window), 400, 300);
    
    GtkWidget *label = gtk_label_new("Hello GTK4!");
    gtk_window_set_child(GTK_WINDOW(window), label);
    
    g_signal_connect(window, "close-request", G_CALLBACK(gtk_window_destroy), NULL);
    
    gtk_window_present(GTK_WINDOW(window));
    
    while (true) {
        g_main_context_iteration(g_main_context_default(), TRUE);
    }
    
    return 0;
}
