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


float f1 = parse_number("3.14159");
float f2 = parse_number("-42.5");
float f3 = parse_number("0.001");
float f4 = parse_number("invalid");
int i1 = parse_int("42");
int i2 = parse_int("-100");
int i3 = parse_int("0");
int i4 = parse_int("3.14");
int i5 = parse_int("not_a_number");
println_string("parse_number tests:")
println_string((cortex_strcat("  3.14159 -> ", as_string(f1))))
println_string((cortex_strcat("  -42.5 -> ", as_string(f2))))
println_string((cortex_strcat("  0.001 -> ", as_string(f3))))
println_string((cortex_strcat("  invalid -> ", as_string(f4))))
println_string("parse_int tests:")
println_string((cortex_strcat("  42 -> ", as_string(i1))))
println_string((cortex_strcat("  -100 -> ", as_string(i2))))
println_string((cortex_strcat("  0 -> ", as_string(i3))))
println_string((cortex_strcat((cortex_strcat("  3.14 -> ", as_string(i4))), " (truncated)")))
println_string((cortex_strcat("  not_a_number -> ", as_string(i5))))
float edge1 = parse_number("");
float edge2 = parse_number("   ");
float edge3 = parse_number("1e10");
println_string("Edge cases:")
println_string((cortex_strcat("  empty -> ", as_string(edge1))))
println_string((cortex_strcat("  whitespace -> ", as_string(edge2))))
println_string((cortex_strcat("  scientific -> ", as_string(edge3))))
println_string("Milestone 6 tests complete!")
