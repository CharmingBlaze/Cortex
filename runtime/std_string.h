// Cortex Standard Library - String Module
// Provides string manipulation functions

#ifndef CORTEX_STD_STRING_H
#define CORTEX_STD_STRING_H

#include <string.h>
#include <stdlib.h>
#include <ctype.h>

// Length and capacity
static inline size_t std_string_len(const char* s) { return strlen(s); }
static inline int std_string_is_empty(const char* s) { return s == NULL || s[0] == '\0'; }

// Comparison
static inline int std_string_cmp(const char* a, const char* b) { return strcmp(a, b); }
static inline int std_string_cmp_n(const char* a, const char* b, size_t n) { return strncmp(a, b, n); }
static inline int std_string_eq(const char* a, const char* b) { return strcmp(a, b) == 0; }
static inline int std_string_ne(const char* a, const char* b) { return strcmp(a, b) != 0; }

// Case-insensitive comparison
static inline int std_string_cmp_i(const char* a, const char* b) { return strcasecmp(a, b); }
static inline int std_string_eq_i(const char* a, const char* b) { return strcasecmp(a, b) == 0; }

// Search
static inline char* std_string_find(const char* s, const char* substr) { return strstr(s, substr); }
static inline char* std_string_find_char(const char* s, char c) { return strchr(s, c); }
static inline char* std_string_find_last(const char* s, char c) { return strrchr(s, c); }
static inline int std_string_contains(const char* s, const char* substr) { return strstr(s, substr) != NULL; }
static inline int std_string_starts_with(const char* s, const char* prefix) {
    size_t len = strlen(prefix);
    return strncmp(s, prefix, len) == 0;
}
static inline int std_string_ends_with(const char* s, const char* suffix) {
    size_t slen = strlen(s);
    size_t suflen = strlen(suffix);
    if (suflen > slen) return 0;
    return strcmp(s + slen - suflen, suffix) == 0;
}

// Character classification
static inline int std_char_is_digit(char c) { return isdigit(c); }
static inline int std_char_is_alpha(char c) { return isalpha(c); }
static inline int std_char_is_alnum(char c) { return isalnum(c); }
static inline int std_char_is_space(char c) { return isspace(c); }
static inline int std_char_is_upper(char c) { return isupper(c); }
static inline int std_char_is_lower(char c) { return islower(c); }
static inline int std_char_is_print(char c) { return isprint(c); }

// Case conversion
static inline char std_char_to_upper(char c) { return toupper(c); }
static inline char std_char_to_lower(char c) { return tolower(c); }

// String case conversion (in-place)
static inline void std_string_to_upper(char* s) {
    for (; *s; s++) *s = toupper(*s);
}
static inline void std_string_to_lower(char* s) {
    for (; *s; s++) *s = tolower(*s);
}

// Trimming
static inline char* std_string_trim_left(char* s) {
    while (isspace(*s)) s++;
    return s;
}
static inline char* std_string_trim_right(char* s) {
    char* end = s + strlen(s) - 1;
    while (end > s && isspace(*end)) end--;
    *(end + 1) = '\0';
    return s;
}
static inline char* std_string_trim(char* s) {
    s = std_string_trim_left(s);
    return std_string_trim_right(s);
}

// Copy and concat
static inline char* std_string_copy(char* dest, const char* src) { return strcpy(dest, src); }
static inline char* std_string_copy_n(char* dest, const char* src, size_t n) { return strncpy(dest, src, n); }
static inline char* std_string_cat(char* dest, const char* src) { return strcat(dest, src); }
static inline char* std_string_cat_n(char* dest, const char* src, size_t n) { return strncat(dest, src, n); }

// Substring (returns new allocated string, caller must free)
static inline char* std_string_substr(const char* s, size_t start, size_t len) {
    size_t slen = strlen(s);
    if (start >= slen) return strdup("");
    if (start + len > slen) len = slen - start;
    char* result = (char*)malloc(len + 1);
    memcpy(result, s + start, len);
    result[len] = '\0';
    return result;
}

// Replace single character
static inline void std_string_replace_char(char* s, char old, char new_char) {
    for (; *s; s++) if (*s == old) *s = new_char;
}

// Count occurrences
static inline size_t std_string_count(const char* s, char c) {
    size_t count = 0;
    for (; *s; s++) if (*s == c) count++;
    return count;
}

// Reverse string in place
static inline void std_string_reverse(char* s) {
    size_t len = strlen(s);
    for (size_t i = 0; i < len / 2; i++) {
        char tmp = s[i];
        s[i] = s[len - 1 - i];
        s[len - 1 - i] = tmp;
    }
}

#endif // CORTEX_STD_STRING_H
