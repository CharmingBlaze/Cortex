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

void main() {
    int width = 800;
int height = 450;
InitWindow(width, height, "Cortex Particle System");
SetTargetFPS(60);
float px[100];
float py[100];
float pvx[100];
float pvy[100];
float plife[100];
int particle_count = 0;
float emitter_x = (width / (int)2.0);
float emitter_y = (height / (int)2.0);
while ((!WindowShouldClose())) {
        emitter_x = GetMouseX();
emitter_y = GetMouseY();
if (((IsMouseButtonDown(MOUSE_LEFT_BUTTON) != 0) && (particle_count < 100))) {
            px[particle_count] = emitter_x;
py[particle_count] = emitter_y;
pvx[particle_count] = ((double)(random_float(0.0, 1.0) - (float)0.5) * 4.0);
pvy[particle_count] = ((double)((double)(random_float(0.0, 1.0) - (float)0.5) * 4.0) - 2.0);
plife[particle_count] = 60.0;
particle_count = (particle_count + (int)1);
}

int i = 0;
while ((i < particle_count)) {
            px[i] = (px[i] + pvx[i]);
py[i] = (py[i] + pvy[i]);
pvy[i] = ((double)pvy[i] + 0.05);
plife[i] = ((double)plife[i] - 1.0);
i = (i + (int)1);
}
int write_idx = 0;
i = 0;
while ((i < particle_count)) {
            if ((plife[i] > 0)) {
                px[write_idx] = px[i];
py[write_idx] = py[i];
pvx[write_idx] = pvx[i];
pvy[write_idx] = pvy[i];
plife[write_idx] = plife[i];
write_idx = (write_idx + (int)1);
}

i = (i + (int)1);
}
particle_count = write_idx;
BeginDrawing();
ClearBackground(BLACK);
i = 0;
while ((i < particle_count)) {
            DrawCircle(px[i], py[i], 4, GOLD);
i = (i + (int)1);
}
DrawText("Click to spawn particles!", 10, 10, 20, WHITE);
DrawText(toString_int(particle_count), 10, 35, 18, GRAY);
EndDrawing();
}
CloseWindow();
}

