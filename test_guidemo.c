// Generated Cortex Program
#define CORTEX_FEATURE_ASYNC 1
#define CORTEX_FEATURE_ACTORS 1
#define CORTEX_FEATURE_BLOCKCHAIN 1
#define CORTEX_FEATURE_QOL 1

#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <stdbool.h>
#include <time.h>
#include <string.h>
#include "runtime/core.h"
#include "runtime/gui_runtime.h"
#include "runtime/game.h"

void on_button_click() {
    gui_dialog_info(0, "Hello", "Button was clicked!");
}

void on_checkbox_change(int checked) {
    if (checked) {
        gui_dialog_info(0, "Checkbox", "Checkbox is now checked!");
}

void on_confirm_result(int confirmed) {
        if (confirmed) {
            gui_dialog_info(0, "Result", "User confirmed the action!");
} else {
            gui_dialog_info(0, "Result", "User cancelled the action.");
}

}

int main() {
        gui_window win = gui_window_create("Cortex GUI Demo", 400, 300);
gui_widget label = gui_label_create("Welcome to Cortex GUI!");
gui_widget button = gui_button_create("Click Me", on_button_click);
gui_widget checkbox = gui_checkbox_create("Enable Feature", NULL);
gui_widget entry = gui_entry_create("Type here...", NULL);
gui_container vbox = gui_vbox_create();
gui_container_add(vbox, label);
gui_container_add(vbox, button);
gui_container_add(vbox, checkbox);
gui_container_add(vbox, entry);
gui_window_set_content(win, vbox);
gui_window_center(win);
gui_run();
return 0;
}

}

