/* raylib_helper.c - Small helpers for using raylib from Cortex.
 * Linked automatically when you use -l raylib or #pragma link("raylib").
 * Provides Vec2() so Cortex can build Vector2 values without C struct literals.
 */
#if defined(__has_include)
#if __has_include("raylib.h")
#include "raylib.h"
#else
#include <raylib.h>
#endif
#else
#include <raylib.h>
#endif

Vector2 Vec2(float x, float y) {
    return (Vector2){ x, y };
}
