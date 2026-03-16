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

#include <std.h>

void main() {
    println_string("=== Std Library Test ===");
println_string("");
println_string("--- std.math ---");
println_string((cortex_strcat("PI: ", toString_int(STD_PI))));
println_string((cortex_strcat("E: ", toString_int(STD_E))));
println_string((cortex_strcat("abs(-42): ", toString_int(std_math_abs((-42))))));
println_string((cortex_strcat("sqrt(16): ", toString_int(std_math_sqrt(16)))));
println_string((cortex_strcat("pow(2, 8): ", toString_int(std_math_pow(2, 8)))));
println_string((cortex_strcat("sin(0): ", toString_int(std_math_sin(0)))));
println_string((cortex_strcat("cos(0): ", toString_int(std_math_cos(0)))));
println_string((cortex_strcat("floor(3.7): ", toString_int(std_math_floor(3.7)))));
println_string((cortex_strcat("ceil(3.2): ", toString_int(std_math_ceil(3.2)))));
println_string((cortex_strcat("min(5, 10): ", toString_int(std_math_min(5, 10)))));
println_string((cortex_strcat("max(5, 10): ", toString_int(std_math_max(5, 10)))));
println_string((cortex_strcat("clamp(15, 0, 10): ", toString_int(std_math_clamp(15, 0, 10)))));
println_string((cortex_strcat("lerp(0, 100, 0.5): ", toString_int(std_math_lerp(0, 100, 0.5)))));
println_string("");
println_string("--- std.time ---");
int now = std_time_now();
println_string((cortex_strcat("Current timestamp: ", toString_int(now))));
println_string("");
println_string("=== All tests passed! ===");
}

