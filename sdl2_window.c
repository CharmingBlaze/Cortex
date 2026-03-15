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

#include <SDL.h>

void main() {
    if ((SDL_Init(SDL_INIT_VIDEO) < 0)) {
        println_string("SDL could not initialize!");
return;
}

AnyValue window = SDL_CreateWindow("Cortex SDL2 Demo", SDL_WINDOWPOS_CENTERED, SDL_WINDOWPOS_CENTERED, 800, 600, SDL_WINDOW_SHOWN);
if ((window == NULL)) {
        println_string("Window could not be created!");
SDL_Quit();
return;
}

AnyValue screenSurface = SDL_GetWindowSurface(window);
SDL_FillRect(screenSurface, NULL, SDL_MapRGB(screenSurface.format, 255, 100, 100));
SDL_UpdateWindowSurface(window);
SDL_Delay(3000);
SDL_DestroyWindow(window);
SDL_Quit();
}

