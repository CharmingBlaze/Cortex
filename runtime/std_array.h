// Cortex Standard Library - Array Module
// Provides array manipulation functions

#ifndef CORTEX_STD_ARRAY_H
#define CORTEX_STD_ARRAY_H

#include <stdlib.h>
#include <string.h>

// Dynamic array structure
typedef struct {
    void* data;
    size_t size;
    size_t capacity;
    size_t elem_size;
} std_array_t;

// Create a new dynamic array
static inline std_array_t* std_array_new(size_t elem_size, size_t initial_capacity) {
    std_array_t* arr = (std_array_t*)malloc(sizeof(std_array_t));
    arr->elem_size = elem_size;
    arr->size = 0;
    arr->capacity = initial_capacity > 0 ? initial_capacity : 8;
    arr->data = malloc(arr->capacity * elem_size);
    return arr;
}

// Free a dynamic array
static inline void std_array_free(std_array_t* arr) {
    if (arr) {
        free(arr->data);
        free(arr);
    }
}

// Get element at index
static inline void* std_array_get(std_array_t* arr, size_t index) {
    if (index >= arr->size) return NULL;
    return (char*)arr->data + index * arr->elem_size;
}

// Set element at index
static inline int std_array_set(std_array_t* arr, size_t index, const void* elem) {
    if (index >= arr->size) return 0;
    memcpy((char*)arr->data + index * arr->elem_size, elem, arr->elem_size);
    return 1;
}

// Internal: ensure capacity
static inline void std_array_ensure_capacity(std_array_t* arr, size_t needed) {
    if (needed > arr->capacity) {
        size_t new_cap = arr->capacity * 2;
        if (new_cap < needed) new_cap = needed;
        arr->data = realloc(arr->data, new_cap * arr->elem_size);
        arr->capacity = new_cap;
    }
}

// Push element to end
static inline void std_array_push(std_array_t* arr, const void* elem) {
    std_array_ensure_capacity(arr, arr->size + 1);
    memcpy((char*)arr->data + arr->size * arr->elem_size, elem, arr->elem_size);
    arr->size++;
}

// Pop element from end
static inline int std_array_pop(std_array_t* arr, void* out) {
    if (arr->size == 0) return 0;
    arr->size--;
    if (out) memcpy(out, (char*)arr->data + arr->size * arr->elem_size, arr->elem_size);
    return 1;
}

// Insert at index
static inline int std_array_insert(std_array_t* arr, size_t index, const void* elem) {
    if (index > arr->size) return 0;
    std_array_ensure_capacity(arr, arr->size + 1);
    // Shift elements right
    memmove((char*)arr->data + (index + 1) * arr->elem_size,
            (char*)arr->data + index * arr->elem_size,
            (arr->size - index) * arr->elem_size);
    memcpy((char*)arr->data + index * arr->elem_size, elem, arr->elem_size);
    arr->size++;
    return 1;
}

// Remove at index
static inline int std_array_remove_at(std_array_t* arr, size_t index) {
    if (index >= arr->size) return 0;
    // Shift elements left
    memmove((char*)arr->data + index * arr->elem_size,
            (char*)arr->data + (index + 1) * arr->elem_size,
            (arr->size - index - 1) * arr->elem_size);
    arr->size--;
    return 1;
}

// Clear array (keep capacity)
static inline void std_array_clear(std_array_t* arr) {
    arr->size = 0;
}

// Get size
static inline size_t std_array_size(std_array_t* arr) { return arr->size; }
static inline int std_array_is_empty(std_array_t* arr) { return arr->size == 0; }

// Find element (linear search, returns index or -1)
typedef int (*std_array_cmp_fn)(const void*, const void*);
static inline int std_array_find(std_array_t* arr, const void* elem, std_array_cmp_fn cmp) {
    for (size_t i = 0; i < arr->size; i++) {
        void* current = (char*)arr->data + i * arr->elem_size;
        if (cmp(current, elem) == 0) return (int)i;
    }
    return -1;
}

// Sort array
static inline void std_array_sort(std_array_t* arr, std_array_cmp_fn cmp) {
    qsort(arr->data, arr->size, arr->elem_size, cmp);
}

// Reverse array
static inline void std_array_reverse(std_array_t* arr) {
    for (size_t i = 0; i < arr->size / 2; i++) {
        void* a = (char*)arr->data + i * arr->elem_size;
        void* b = (char*)arr->data + (arr->size - 1 - i) * arr->elem_size;
        char tmp[256]; // Assume element size <= 256
        memcpy(tmp, a, arr->elem_size);
        memcpy(a, b, arr->elem_size);
        memcpy(b, tmp, arr->elem_size);
    }
}

// Integer array helpers
static inline int std_array_int_get(std_array_t* arr, size_t index) {
    int* p = (int*)std_array_get(arr, index);
    return p ? *p : 0;
}
static inline void std_array_int_push(std_array_t* arr, int val) {
    std_array_push(arr, &val);
}

// Float array helpers
static inline double std_array_float_get(std_array_t* arr, size_t index) {
    double* p = (double*)std_array_get(arr, index);
    return p ? *p : 0.0;
}
static inline void std_array_float_push(std_array_t* arr, double val) {
    std_array_push(arr, &val);
}

#endif // CORTEX_STD_ARRAY_H
