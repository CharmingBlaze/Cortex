// Generated Cortex Program
#define CORTEX_FEATURE_ASYNC 0
#define CORTEX_FEATURE_ACTORS 0
#define CORTEX_FEATURE_BLOCKCHAIN 0
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

void main() {
    int width = 800;
int height = 450;
InitWindow(width, height, "Cortex - Bouncing Ball");
SetTargetFPS(60);
float x = (width / (int)2.0);
float y = (height / (int)2.0);
float vx = 4.0;
float vy = 3.0;
float radius = 20.0;
while ((!WindowShouldClose())) {
        x = (x + vx);
y = (y + vy);
if ((((x + radius) > width) || ((x - radius) < 0))) {
            vx = (-vx);
}

if ((((y + radius) > height) || ((y - radius) < 0))) {
            vy = (-vy);
}

BeginDrawing();
ClearBackground(RAYWHITE);
DrawCircle(x, y, radius, MAROON);
DrawText("Bouncing Ball Demo", 10, 10, 20, DARKGRAY);
DrawText("Press ESC to exit", 10, 35, 18, LIGHTGRAY);
EndDrawing();
}
CloseWindow();
}

