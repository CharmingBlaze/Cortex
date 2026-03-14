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
        printf("=== Smart Dynamic Typing Demo ===\n");
printf("\n--- Intelligent Type Inference ---\n");
float result = 13.14;
char* text = 42;
int flag = true;
printf("result (10 + 3.14): %f\n", result);
printf("text: %s\n", text);
if (flag) {
            printf("flag (5 > 3): true\n");
} else {
            printf("flag (5 > 3): false\n");
}
printf("\n--- Automatic Type Promotion ---\n");
int int_val = 42;
float float_val = (int_val + (int)0.5);
float mixed_result = (int_val * (int)2.0);
printf("int_val: %d\n", int_val);
printf("float_val (42 + 0.5): %f\n", float_val);
printf("mixed_result (42 * 2.0): %f\n", mixed_result);
printf("\n--- Smart String Operations ---\n");
char* name = "World";
char* greeting = (cortex_strcat("Hello ", name));
char* message = 42;
char* combined = (cortex_strcat(greeting, message));
printf("name: %s\n", name);
printf("greeting: %s\n", greeting);
printf("message: %s\n", message);
printf("combined: %s\n", combined);
printf("\n--- Type-Aware Operations ---\n");
int a = 10;
float b = 3.5;
float c = ((float)a + b);
printf("a (int): %d\n", a);
printf("b (float): %f\n", b);
printf("c (a + b): %f\n", c);
char* str_num = 42;
printf("str_num: %s\n", str_num);
int is_positive = (c > 0);
int is_large = (c > 100);
int both_true = (is_positive && is_large);
if (is_positive) {
            printf("is_positive: true\n");
} else {
            printf("is_positive: false\n");
}
if (is_large) {
            printf("is_large: true\n");
} else {
            printf("is_large: false\n");
}
if (both_true) {
            printf("both_true: true\n");
} else {
            printf("both_true: false\n");
}
printf("\n=== Smart Dynamic Typing Demo Completed! ===\n");
}}
