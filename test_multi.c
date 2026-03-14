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
#include "runtime/game.h"

void main() {    {
        AnyValue win = gui_window_create("Test", 300, 200);
gui_window_center(win);
gui_dialog_info(win, "Hello", "This is a Cortex GUI dialog!");
gui_run();
}}
