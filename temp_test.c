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
#include "runtime/async.h"
#include "runtime/game.h"

typedef struct {  } test_coroutine_frame;

static void test_coroutine_entry(void* _arg) {
    test_coroutine_frame* _f = (test_coroutine_frame*)_arg;
println_string("Coroutine: step 1")
co_yield()
println_string("Coroutine: step 2")
co_yield()
println_string("Coroutine: step 3")
}

void test_coroutine() {
    test_coroutine_frame* _frame = malloc(sizeof(test_coroutine_frame));
co_t _co = co_create(test_coroutine_entry, _frame, 0);
co_resume(_co);
co_free(_co);
free(_frame);
}

void main() {
    println_string("Starting coroutine test...");
test_coroutine();
println_string("Coroutine test complete!");
}

