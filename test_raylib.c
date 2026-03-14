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

#include <raylib.h>

#pragma link("raylib") 

void main() {    {
        InitWindow(800, 450, "Cortex + raylib");
while ((!WindowShouldClose())) {
            BeginDrawing();
ClearBackground(RAYWHITE);
DrawText("Hello from Cortex!", 190, 200, 20, LIGHTGRAY);
EndDrawing();
}
CloseWindow();
}}
