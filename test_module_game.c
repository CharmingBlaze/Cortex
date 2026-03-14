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

int main() {
    int x = math__add(1, 2);
int y = math__mul(3, 4);
println_string((cortex_strcat("", toString_int(x))));
println_string((cortex_strcat("", toString_int(y))));
return 0;
}

