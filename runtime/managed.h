// Cortex Managed Memory - Automatic cleanup for extern functions
// Uses GCC/Clang cleanup attribute for automatic resource management

#ifndef CORTEX_MANAGED_H
#define CORTEX_MANAGED_H

#include <stdlib.h>

// === Cleanup Helper Functions ===
// These are called automatically when variables go out of scope

// Generic cleanup function for void* pointers
static inline void cortex_cleanup_free(void** ptr) {
    if (ptr && *ptr) {
        free(*ptr);
        *ptr = NULL;
    }
}

// Cleanup function that takes a cleanup function pointer
typedef void (*cortex_cleanup_fn)(void*);

static inline void cortex_cleanup_with(void** ptr, cortex_cleanup_fn cleanup) {
    if (ptr && *ptr && cleanup) {
        cleanup(*ptr);
        *ptr = NULL;
    }
}

// === Managed Pointer Wrapper ===
// Wraps a pointer with its cleanup function for the cleanup attribute
typedef struct {
    void* ptr;
    cortex_cleanup_fn cleanup;
} cortex_managed;

// Cleanup function for cortex_managed (used by __attribute__((cleanup)))
static inline void cortex_managed_cleanup(cortex_managed* handle) {
    if (handle && handle->ptr && handle->cleanup) {
        handle->cleanup(handle->ptr);
        handle->ptr = NULL;
    }
}

#endif // CORTEX_MANAGED_H
