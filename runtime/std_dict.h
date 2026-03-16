// Cortex Standard Library - Dict (Dictionary/Map) Module
// Provides key-value dictionary functions

#ifndef CORTEX_STD_DICT_H
#define CORTEX_STD_DICT_H

#include <stdlib.h>
#include <string.h>

// Hash map entry
typedef struct std_dict_entry_t {
    char* key;
    void* value;
    struct std_dict_entry_t* next;
} std_dict_entry_t;

// Hash map structure
typedef struct {
    std_dict_entry_t** buckets;
    size_t bucket_count;
    size_t size;
    size_t value_size;
} std_dict_t;

// Simple string hash function
static inline unsigned int std_dict_hash(const char* key, size_t bucket_count) {
    unsigned int hash = 5381;
    int c;
    while ((c = *key++)) {
        hash = ((hash << 5) + hash) + c;
    }
    return hash % bucket_count;
}

// Create a new dictionary
static inline std_dict_t* std_dict_new(size_t value_size, size_t initial_buckets) {
    std_dict_t* dict = (std_dict_t*)malloc(sizeof(std_dict_t));
    dict->bucket_count = initial_buckets > 0 ? initial_buckets : 16;
    dict->buckets = (std_dict_entry_t**)calloc(dict->bucket_count, sizeof(std_dict_entry_t*));
    dict->size = 0;
    dict->value_size = value_size;
    return dict;
}

// Free a dictionary
static inline void std_dict_free(std_dict_t* dict) {
    if (!dict) return;
    for (size_t i = 0; i < dict->bucket_count; i++) {
        std_dict_entry_t* entry = dict->buckets[i];
        while (entry) {
            std_dict_entry_t* next = entry->next;
            free(entry->key);
            free(entry->value);
            free(entry);
            entry = next;
        }
    }
    free(dict->buckets);
    free(dict);
}

// Get value by key
static inline void* std_dict_get(std_dict_t* dict, const char* key) {
    if (!dict || !key) return NULL;
    unsigned int idx = std_dict_hash(key, dict->bucket_count);
    std_dict_entry_t* entry = dict->buckets[idx];
    while (entry) {
        if (strcmp(entry->key, key) == 0) {
            return entry->value;
        }
        entry = entry->next;
    }
    return NULL;
}

// Check if key exists
static inline int std_dict_has(std_dict_t* dict, const char* key) {
    return std_dict_get(dict, key) != NULL;
}

// Set key-value pair
static inline void std_dict_set(std_dict_t* dict, const char* key, const void* value) {
    if (!dict || !key) return;
    unsigned int idx = std_dict_hash(key, dict->bucket_count);
    std_dict_entry_t* entry = dict->buckets[idx];
    
    // Check if key already exists
    while (entry) {
        if (strcmp(entry->key, key) == 0) {
            memcpy(entry->value, value, dict->value_size);
            return;
        }
        entry = entry->next;
    }
    
    // Create new entry
    entry = (std_dict_entry_t*)malloc(sizeof(std_dict_entry_t));
    entry->key = strdup(key);
    entry->value = malloc(dict->value_size);
    memcpy(entry->value, value, dict->value_size);
    entry->next = dict->buckets[idx];
    dict->buckets[idx] = entry;
    dict->size++;
}

// Remove key
static inline int std_dict_remove(std_dict_t* dict, const char* key) {
    if (!dict || !key) return 0;
    unsigned int idx = std_dict_hash(key, dict->bucket_count);
    std_dict_entry_t* entry = dict->buckets[idx];
    std_dict_entry_t* prev = NULL;
    
    while (entry) {
        if (strcmp(entry->key, key) == 0) {
            if (prev) prev->next = entry->next;
            else dict->buckets[idx] = entry->next;
            free(entry->key);
            free(entry->value);
            free(entry);
            dict->size--;
            return 1;
        }
        prev = entry;
        entry = entry->next;
    }
    return 0;
}

// Get size
static inline size_t std_dict_size(std_dict_t* dict) { return dict->size; }
static inline int std_dict_is_empty(std_dict_t* dict) { return dict->size == 0; }

// Clear dictionary
static inline void std_dict_clear(std_dict_t* dict) {
    if (!dict) return;
    for (size_t i = 0; i < dict->bucket_count; i++) {
        std_dict_entry_t* entry = dict->buckets[i];
        while (entry) {
            std_dict_entry_t* next = entry->next;
            free(entry->key);
            free(entry->value);
            free(entry);
            entry = next;
        }
        dict->buckets[i] = NULL;
    }
    dict->size = 0;
}

// Get all keys (returns array of strings, caller must free each string and the array)
static inline char** std_dict_keys(std_dict_t* dict, size_t* out_count) {
    if (!dict || dict->size == 0) {
        *out_count = 0;
        return NULL;
    }
    char** keys = (char**)malloc(dict->size * sizeof(char*));
    size_t idx = 0;
    for (size_t i = 0; i < dict->bucket_count; i++) {
        std_dict_entry_t* entry = dict->buckets[i];
        while (entry) {
            keys[idx++] = strdup(entry->key);
            entry = entry->next;
        }
    }
    *out_count = idx;
    return keys;
}

// Integer dict helpers
static inline int std_dict_int_get(std_dict_t* dict, const char* key, int default_val) {
    int* p = (int*)std_dict_get(dict, key);
    return p ? *p : default_val;
}
static inline void std_dict_int_set(std_dict_t* dict, const char* key, int val) {
    std_dict_set(dict, key, &val);
}

// Float dict helpers
static inline double std_dict_float_get(std_dict_t* dict, const char* key, double default_val) {
    double* p = (double*)std_dict_get(dict, key);
    return p ? *p : default_val;
}
static inline void std_dict_float_set(std_dict_t* dict, const char* key, double val) {
    std_dict_set(dict, key, &val);
}

// String dict helpers
static inline const char* std_dict_string_get(std_dict_t* dict, const char* key, const char* default_val) {
    char** p = (char**)std_dict_get(dict, key);
    return p ? *p : default_val;
}
static inline void std_dict_string_set(std_dict_t* dict, const char* key, const char* val) {
    std_dict_set(dict, key, &val);
}

#endif // CORTEX_STD_DICT_H
