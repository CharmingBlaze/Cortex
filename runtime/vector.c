/* Vector<T> generic type implementation for Cortex */

#include <stdlib.h>
#include <string.h>
#include "core.h"

/* Int vector operations */
void vector_push_int(cortex_vector_int* v, int val) {
    if (v->size >= v->capacity) {
        int new_cap = v->capacity == 0 ? 8 : v->capacity * 2;
        v->data = realloc(v->data, new_cap * sizeof(int));
        v->capacity = new_cap;
    }
    v->data[v->size++] = val;
}

int vector_pop_int(cortex_vector_int* v) {
    if (v->size == 0) return 0;
    return v->data[--v->size];
}

int vector_get_int(cortex_vector_int* v, int idx) {
    if (idx < 0 || idx >= v->size) return 0;
    return v->data[idx];
}

void vector_set_int(cortex_vector_int* v, int idx, int val) {
    if (idx >= 0 && idx < v->size) {
        v->data[idx] = val;
    }
}

int vector_len_int(cortex_vector_int* v) {
    return v->size;
}

void vector_free_int(cortex_vector_int* v) {
    free(v->data);
    v->data = NULL;
    v->size = 0;
    v->capacity = 0;
}

/* Float vector operations */
void vector_push_float(cortex_vector_float* v, float val) {
    if (v->size >= v->capacity) {
        int new_cap = v->capacity == 0 ? 8 : v->capacity * 2;
        v->data = realloc(v->data, new_cap * sizeof(float));
        v->capacity = new_cap;
    }
    v->data[v->size++] = val;
}

float vector_pop_float(cortex_vector_float* v) {
    if (v->size == 0) return 0.0f;
    return v->data[--v->size];
}

float vector_get_float(cortex_vector_float* v, int idx) {
    if (idx < 0 || idx >= v->size) return 0.0f;
    return v->data[idx];
}

void vector_set_float(cortex_vector_float* v, int idx, float val) {
    if (idx >= 0 && idx < v->size) {
        v->data[idx] = val;
    }
}

int vector_len_float(cortex_vector_float* v) {
    return v->size;
}

void vector_free_float(cortex_vector_float* v) {
    free(v->data);
    v->data = NULL;
    v->size = 0;
    v->capacity = 0;
}

/* String vector operations */
void vector_push_string(cortex_vector_string* v, char* val) {
    if (v->size >= v->capacity) {
        int new_cap = v->capacity == 0 ? 8 : v->capacity * 2;
        v->data = realloc(v->data, new_cap * sizeof(char*));
        v->capacity = new_cap;
    }
    v->data[v->size++] = val;
}

char* vector_pop_string(cortex_vector_string* v) {
    if (v->size == 0) return NULL;
    return v->data[--v->size];
}

char* vector_get_string(cortex_vector_string* v, int idx) {
    if (idx < 0 || idx >= v->size) return NULL;
    return v->data[idx];
}

void vector_set_string(cortex_vector_string* v, int idx, char* val) {
    if (idx >= 0 && idx < v->size) {
        v->data[idx] = val;
    }
}

int vector_len_string(cortex_vector_string* v) {
    return v->size;
}

void vector_free_string(cortex_vector_string* v) {
    for (int i = 0; i < v->size; i++) {
        free(v->data[i]);
    }
    free(v->data);
    v->data = NULL;
    v->size = 0;
    v->capacity = 0;
}

/* Double vector operations */
void vector_push_double(cortex_vector_double* v, double val) {
    if (v->size >= v->capacity) {
        int new_cap = v->capacity == 0 ? 8 : v->capacity * 2;
        v->data = realloc(v->data, new_cap * sizeof(double));
        v->capacity = new_cap;
    }
    v->data[v->size++] = val;
}

double vector_pop_double(cortex_vector_double* v) {
    if (v->size == 0) return 0.0;
    return v->data[--v->size];
}

double vector_get_double(cortex_vector_double* v, int idx) {
    if (idx < 0 || idx >= v->size) return 0.0;
    return v->data[idx];
}

void vector_set_double(cortex_vector_double* v, int idx, double val) {
    if (idx >= 0 && idx < v->size) {
        v->data[idx] = val;
    }
}

int vector_len_double(cortex_vector_double* v) {
    return v->size;
}

void vector_free_double(cortex_vector_double* v) {
    free(v->data);
    v->data = NULL;
    v->size = 0;
    v->capacity = 0;
}

/* Bool vector operations */
void vector_push_bool(cortex_vector_bool* v, bool val) {
    if (v->size >= v->capacity) {
        int new_cap = v->capacity == 0 ? 8 : v->capacity * 2;
        v->data = realloc(v->data, new_cap * sizeof(bool));
        v->capacity = new_cap;
    }
    v->data[v->size++] = val;
}

bool vector_pop_bool(cortex_vector_bool* v) {
    if (v->size == 0) return false;
    return v->data[--v->size];
}

bool vector_get_bool(cortex_vector_bool* v, int idx) {
    if (idx < 0 || idx >= v->size) return false;
    return v->data[idx];
}

void vector_set_bool(cortex_vector_bool* v, int idx, bool val) {
    if (idx >= 0 && idx < v->size) {
        v->data[idx] = val;
    }
}

int vector_len_bool(cortex_vector_bool* v) {
    return v->size;
}

void vector_free_bool(cortex_vector_bool* v) {
    free(v->data);
    v->data = NULL;
    v->size = 0;
    v->capacity = 0;
}

/* Vec2 vector operations */
void vector_push_vec2(cortex_vector_vec2* v, vec2 val) {
    if (v->size >= v->capacity) {
        int new_cap = v->capacity == 0 ? 8 : v->capacity * 2;
        v->data = realloc(v->data, new_cap * sizeof(vec2));
        v->capacity = new_cap;
    }
    v->data[v->size++] = val;
}

vec2 vector_pop_vec2(cortex_vector_vec2* v) {
    if (v->size == 0) return make_vec2(0, 0);
    return v->data[--v->size];
}

vec2 vector_get_vec2(cortex_vector_vec2* v, int idx) {
    if (idx < 0 || idx >= v->size) return make_vec2(0, 0);
    return v->data[idx];
}

void vector_set_vec2(cortex_vector_vec2* v, int idx, vec2 val) {
    if (idx >= 0 && idx < v->size) {
        v->data[idx] = val;
    }
}

int vector_len_vec2(cortex_vector_vec2* v) {
    return v->size;
}

void vector_free_vec2(cortex_vector_vec2* v) {
    free(v->data);
    v->data = NULL;
    v->size = 0;
    v->capacity = 0;
}

/* Vec3 vector operations */
void vector_push_vec3(cortex_vector_vec3* v, vec3 val) {
    if (v->size >= v->capacity) {
        int new_cap = v->capacity == 0 ? 8 : v->capacity * 2;
        v->data = realloc(v->data, new_cap * sizeof(vec3));
        v->capacity = new_cap;
    }
    v->data[v->size++] = val;
}

vec3 vector_pop_vec3(cortex_vector_vec3* v) {
    if (v->size == 0) return make_vec3(0, 0, 0);
    return v->data[--v->size];
}

vec3 vector_get_vec3(cortex_vector_vec3* v, int idx) {
    if (idx < 0 || idx >= v->size) return make_vec3(0, 0, 0);
    return v->data[idx];
}

void vector_set_vec3(cortex_vector_vec3* v, int idx, vec3 val) {
    if (idx >= 0 && idx < v->size) {
        v->data[idx] = val;
    }
}

int vector_len_vec3(cortex_vector_vec3* v) {
    return v->size;
}

void vector_free_vec3(cortex_vector_vec3* v) {
    free(v->data);
    v->data = NULL;
    v->size = 0;
    v->capacity = 0;
}
