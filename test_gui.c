#include <stdio.h>
#include "runtime/gui_runtime.h"

void on_click(gui_event e) {
    printf("Button clicked!\n");
    gui_alert_info("Button clicked!");
}

int main() {
    printf("Starting GUI...\n");
    
    gui_start("Test GUI", 600, 400);
    printf("Window created\n");
    
    gui_set_spacing(16);
    gui_set_margin(25);
    
    gui_header("Test Section");
    gui_add(gui_entry("Type here..."));
    
    gui_container row = gui_hbox();
    gui_add_to(row, gui_button("Button 1", on_click));
    gui_add_to(row, gui_button("Button 2", on_click));
    gui_add_to(row, gui_button("Button 3", on_click));
    gui_end_row();
    
    printf("Widgets added, running...\n");
    gui_run();
    
    printf("GUI closed\n");
    return 0;
}
