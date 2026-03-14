#include "core.h"

#include <ctype.h>
#include <dirent.h>
#include <math.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <time.h>
#ifdef _WIN32
#include <windows.h>
#else
#include <sys/time.h>
#include <unistd.h>
#endif

#if CORTEX_FEATURE_QOL
AnyValue make_any_int(int val) {
    AnyValue out = {TYPE_INT};
    out.data.int_val = val;
    return out;
}

AnyValue make_any_float(float val) {
    AnyValue out = {TYPE_FLOAT};
    out.data.float_val = val;
    return out;
}

AnyValue make_any_string(char* val) {
    AnyValue out = {TYPE_STRING};
    out.data.string_val = val;
    return out;
}

AnyValue make_any_bool(bool val) {
    AnyValue out = {TYPE_BOOL};
    out.data.bool_val = val;
    return out;
}

AnyValue make_any_vec2(vec2 val) {
    AnyValue out = {TYPE_VEC2};
    out.data.vec2_val = val;
    return out;
}

AnyValue make_any_vec3(vec3 val) {
    AnyValue out = {TYPE_VEC3};
    out.data.vec3_val = val;
    return out;
}

AnyValue make_any_null(void) {
    AnyValue out = {TYPE_NULL};
    return out;
}

char* type_of(AnyValue val) {
    switch (val.type) {
        case TYPE_INT: return "int";
        case TYPE_FLOAT: return "float";
        case TYPE_STRING: return "string";
        case TYPE_BOOL: return "bool";
        case TYPE_VEC2: return "vec2";
        case TYPE_VEC3: return "vec3";
        case TYPE_NULL: return "null";
        case TYPE_DICT: return "dict";
        case TYPE_ARRAY: return "array";
        default: return "unknown";
    }
}

bool is_type(AnyValue val, char* type_name) {
    return strcmp(type_of(val), type_name) == 0;
}

int as_int(AnyValue val) {
    if (val.type == TYPE_INT) return val.data.int_val;
    if (val.type == TYPE_FLOAT) return (int)val.data.float_val;
    return 0;
}

float as_float(AnyValue val) {
    if (val.type == TYPE_FLOAT) return val.data.float_val;
    if (val.type == TYPE_INT) return (float)val.data.int_val;
    return 0.0f;
}

char* as_string(AnyValue val) {
    if (val.type == TYPE_STRING) return val.data.string_val;
    return "";
}

bool as_bool(AnyValue val) {
    if (val.type == TYPE_BOOL) return val.data.bool_val;
    return false;
}
#else
AnyValue make_any_int(int val) { (void)val; return (AnyValue){TYPE_INT}; }
AnyValue make_any_float(float val) { (void)val; return (AnyValue){TYPE_FLOAT}; }
AnyValue make_any_string(char* val) { (void)val; return (AnyValue){TYPE_STRING}; }
AnyValue make_any_bool(bool val) { (void)val; return (AnyValue){TYPE_BOOL}; }
AnyValue make_any_vec2(vec2 val) { (void)val; return (AnyValue){TYPE_VEC2}; }
AnyValue make_any_vec3(vec3 val) { (void)val; return (AnyValue){TYPE_VEC3}; }
AnyValue make_any_null(void) { return (AnyValue){TYPE_NULL}; }
char* type_of(AnyValue val) { (void)val; return "unknown"; }
bool is_type(AnyValue val, char* type_name) { (void)val; (void)type_name; return false; }
int as_int(AnyValue val) { (void)val; return 0; }
float as_float(AnyValue val) { (void)val; return 0.0f; }
char* as_string(AnyValue val) { (void)val; return ""; }
bool as_bool(AnyValue val) { (void)val; return false; }
#endif

#if CORTEX_FEATURE_QOL
static char int_buffer[32];
static char float_buffer[64];
static char double_buffer[64];
static char bool_buffer[8];

char* toString_int(int val) {
    snprintf(int_buffer, sizeof(int_buffer), "%d", val);
    return int_buffer;
}

char* toString_float(float val) {
    snprintf(float_buffer, sizeof(float_buffer), "%f", val);
    return float_buffer;
}

char* toString_double(double val) {
    snprintf(double_buffer, sizeof(double_buffer), "%lf", val);
    return double_buffer;
}

char* toString_bool(bool val) {
    snprintf(bool_buffer, sizeof(bool_buffer), "%s", val ? "true" : "false");
    return bool_buffer;
}
#else
char* toString_int(int val) { (void)val; return ""; }
char* toString_float(float val) { (void)val; return ""; }
char* toString_double(double val) { (void)val; return ""; }
char* toString_bool(bool val) { (void)val; return ""; }
#endif

#if CORTEX_FEATURE_QOL
vec2 make_vec2(float x, float y) {
    vec2 v = {x, y};
    return v;
}

vec3 make_vec3(float x, float y, float z) {
    vec3 v = {x, y, z};
    return v;
}

float dot(vec2 a, vec2 b) {
    return a.x * b.x + a.y * b.y;
}

vec2 normalize(vec2 v) {
    float len = sqrtf(v.x * v.x + v.y * v.y);
    if (len == 0) return (vec2){0, 0};
    return (vec2){v.x / len, v.y / len};
}

float vec2_length(vec2 v) {
    return sqrtf(v.x * v.x + v.y * v.y);
}

float vec2_length_sq(vec2 v) {
    return v.x * v.x + v.y * v.y;
}

float vec2_distance(vec2 a, vec2 b) {
    float dx = b.x - a.x, dy = b.y - a.y;
    return sqrtf(dx * dx + dy * dy);
}

vec2 vec2_add(vec2 a, vec2 b) {
    return (vec2){a.x + b.x, a.y + b.y};
}

vec2 vec2_sub(vec2 a, vec2 b) {
    return (vec2){a.x - b.x, a.y - b.y};
}

vec2 vec2_scale(vec2 v, float s) {
    return (vec2){v.x * s, v.y * s};
}

vec2 vec2_normalize(vec2 v) {
    float len = sqrtf(v.x * v.x + v.y * v.y);
    if (len == 0) return (vec2){0, 0};
    return (vec2){v.x / len, v.y / len};
}

float vec2_dot(vec2 a, vec2 b) {
    return a.x * b.x + a.y * b.y;
}

vec2 vec2_lerp(vec2 a, vec2 b, float t) {
    return (vec2){a.x + (b.x - a.x) * t, a.y + (b.y - a.y) * t};
}

vec4 make_vec4(float x, float y, float z, float w) {
    return (vec4){x, y, z, w};
}

vec4 vec4_lerp(vec4 a, vec4 b, float t) {
    return (vec4){
        a.x + (b.x - a.x) * t,
        a.y + (b.y - a.y) * t,
        a.z + (b.z - a.z) * t,
        a.w + (b.w - a.w) * t
    };
}

float vec3_length(vec3 v) {
    return sqrtf(v.x * v.x + v.y * v.y + v.z * v.z);
}

float vec3_length_sq(vec3 v) {
    return v.x * v.x + v.y * v.y + v.z * v.z;
}

float vec3_dot(vec3 a, vec3 b) {
    return a.x * b.x + a.y * b.y + a.z * b.z;
}

vec3 vec3_normalize(vec3 v) {
    float len = sqrtf(v.x * v.x + v.y * v.y + v.z * v.z);
    if (len == 0) return (vec3){0, 0, 0};
    return (vec3){v.x / len, v.y / len, v.z / len};
}

vec3 vec3_add(vec3 a, vec3 b) {
    return (vec3){a.x + b.x, a.y + b.y, a.z + b.z};
}

vec3 vec3_sub(vec3 a, vec3 b) {
    return (vec3){a.x - b.x, a.y - b.y, a.z - b.z};
}

vec3 vec3_scale(vec3 v, float s) {
    return (vec3){v.x * s, v.y * s, v.z * s};
}

float vec3_distance(vec3 a, vec3 b) {
    float dx = b.x - a.x, dy = b.y - a.y, dz = b.z - a.z;
    return sqrtf(dx * dx + dy * dy + dz * dz);
}

vec3 vec3_lerp(vec3 a, vec3 b, float t) {
    return (vec3){a.x + (b.x - a.x) * t, a.y + (b.y - a.y) * t, a.z + (b.z - a.z) * t};
}

float clamp_float(float x, float lo, float hi) {
    if (x < lo) return lo;
    if (x > hi) return hi;
    return x;
}

float lerp_float(float a, float b, float t) {
    return a + (b - a) * t;
}

float min_float(float a, float b) {
    return a < b ? a : b;
}

float max_float(float a, float b) {
    return a > b ? a : b;
}

float sign_float(float x) {
    if (x > 0) return 1.0f;
    if (x < 0) return -1.0f;
    return 0.0f;
}

float wrap_float(float x, float lo, float hi) {
    float range = hi - lo;
    if (range <= 0) return lo;
    float d = x - lo;
    float r = fmodf(d, range);
    if (r < 0) r += range;
    return lo + r;
}

float round_float(float x) {
    return floorf(x + 0.5f);
}

float floor_float(float x) {
    return floorf(x);
}

float ceil_float(float x) {
    return ceilf(x);
}

/* --- Dynamic array --- */
#define CORTEX_ARRAY_INIT_CAP 16
struct cortex_array {
    AnyValue* data;
    int len;
    int cap;
};

cortex_array* array_create(void) {
    cortex_array* a = (cortex_array*)malloc(sizeof(cortex_array));
    if (!a) return NULL;
    a->cap = CORTEX_ARRAY_INIT_CAP;
    a->len = 0;
    a->data = (AnyValue*)malloc((size_t)a->cap * sizeof(AnyValue));
    if (!a->data) { free(a); return NULL; }
    return a;
}

void array_push(cortex_array* a, AnyValue val) {
    if (!a) return;
    if (a->len >= a->cap) {
        int newcap = a->cap * 2;
        AnyValue* newdata = (AnyValue*)realloc(a->data, (size_t)newcap * sizeof(AnyValue));
        if (!newdata) return;
        a->data = newdata;
        a->cap = newcap;
    }
    a->data[a->len++] = val;
}

AnyValue array_get(cortex_array* a, int index) {
    AnyValue out = make_any_null();
    if (!a || index < 0 || index >= a->len) return out;
    return a->data[index];
}

void array_set(cortex_array* a, int index, AnyValue val) {
    if (!a || index < 0 || index >= a->len) return;
    a->data[index] = val;
}

int array_len(cortex_array* a) {
    return a ? a->len : 0;
}

int array_capacity(cortex_array* a) {
    return a ? a->cap : 0;
}

void array_reserve(cortex_array* a, int min_cap) {
    if (!a || min_cap <= a->cap) return;
    AnyValue* newdata = (AnyValue*)realloc(a->data, (size_t)min_cap * sizeof(AnyValue));
    if (!newdata) return;
    a->data = newdata;
    a->cap = min_cap;
}

AnyValue array_pop(cortex_array* a) {
    AnyValue out = make_any_null();
    if (!a || a->len == 0) return out;
    return a->data[--a->len];
}

void array_insert(cortex_array* a, int index, AnyValue val) {
    if (!a || index < 0 || index > a->len) return;
    if (a->len >= a->cap) {
        int newcap = a->cap ? a->cap * 2 : CORTEX_ARRAY_INIT_CAP;
        AnyValue* newdata = (AnyValue*)realloc(a->data, (size_t)newcap * sizeof(AnyValue));
        if (!newdata) return;
        a->data = newdata;
        a->cap = newcap;
    }
    memmove(&a->data[index + 1], &a->data[index], (size_t)(a->len - index) * sizeof(AnyValue));
    a->data[index] = val;
    a->len++;
}

void array_remove_at(cortex_array* a, int index) {
    if (!a || index < 0 || index >= a->len) return;
    memmove(&a->data[index], &a->data[index + 1], (size_t)(a->len - 1 - index) * sizeof(AnyValue));
    a->len--;
}

void array_free(cortex_array* a) {
    if (a) {
        free(a->data);
        free(a);
    }
}

/* --- Dictionary (string key -> AnyValue), simple hash table --- */
#define CORTEX_DICT_INIT_CAP 32
#define CORTEX_DICT_LOAD 70  /* resize when load > 70% */

typedef struct dict_entry {
    char* key;
    AnyValue val;
    struct dict_entry* next;
} dict_entry;

struct cortex_dict {
    dict_entry** buckets;
    int cap;
    int count;
};

static unsigned dict_hash(const char* s) {
    unsigned h = 5381;
    while (*s) h = ((h << 5) + h) + (unsigned char)*s++;
    return h;
}

cortex_dict* dict_create(void) {
    cortex_dict* d = (cortex_dict*)malloc(sizeof(cortex_dict));
    if (!d) return NULL;
    d->cap = CORTEX_DICT_INIT_CAP;
    d->count = 0;
    d->buckets = (dict_entry**)calloc((size_t)d->cap, sizeof(dict_entry*));
    if (!d->buckets) { free(d); return NULL; }
    return d;
}

static void dict_grow(cortex_dict* d) {
    int newcap = d->cap * 2;
    dict_entry** newb = (dict_entry**)calloc((size_t)newcap, sizeof(dict_entry*));
    if (!newb) return;
    for (int i = 0; i < d->cap; i++) {
        for (dict_entry* e = d->buckets[i]; e; ) {
            dict_entry* next = e->next;
            unsigned h = dict_hash(e->key) % (unsigned)newcap;
            e->next = newb[h];
            newb[h] = e;
            e = next;
        }
    }
    free(d->buckets);
    d->buckets = newb;
    d->cap = newcap;
}

void dict_set(cortex_dict* d, const char* key, AnyValue val) {
    if (!d || !key) return;
    if (d->count * 100 >= d->cap * CORTEX_DICT_LOAD)
        dict_grow(d);
    unsigned h = dict_hash(key) % (unsigned)d->cap;
    for (dict_entry* e = d->buckets[h]; e; e = e->next) {
        if (strcmp(e->key, key) == 0) {
            e->val = val;
            return;
        }
    }
    dict_entry* e = (dict_entry*)malloc(sizeof(dict_entry));
    if (!e) return;
    e->key = strdup(key);
    e->val = val;
    e->next = d->buckets[h];
    d->buckets[h] = e;
    d->count++;
}

AnyValue dict_get(cortex_dict* d, const char* key) {
    AnyValue out = make_any_null();
    if (!d || !key) return out;
    unsigned h = dict_hash(key) % (unsigned)d->cap;
    for (dict_entry* e = d->buckets[h]; e; e = e->next) {
        if (strcmp(e->key, key) == 0) return e->val;
    }
    return out;
}

bool dict_has(cortex_dict* d, const char* key) {
    if (!d || !key) return false;
    unsigned h = dict_hash(key) % (unsigned)d->cap;
    for (dict_entry* e = d->buckets[h]; e; e = e->next) {
        if (strcmp(e->key, key) == 0) return true;
    }
    return false;
}

int dict_len(cortex_dict* d) {
    return d ? d->count : 0;
}

void dict_free(cortex_dict* d) {
    if (!d) return;
    for (int i = 0; i < d->cap; i++) {
        dict_entry* e = d->buckets[i];
        while (e) {
            dict_entry* next = e->next;
            free(e->key);
            free(e);
            e = next;
        }
    }
    free(d->buckets);
    free(d);
}

AnyValue make_any_dict(cortex_dict* d) {
    AnyValue out = {TYPE_DICT};
    out.data.dict_val = (void*)d;
    return out;
}

AnyValue make_any_array(cortex_array* a) {
    AnyValue out = {TYPE_ARRAY};
    out.data.array_val = (void*)a;
    return out;
}

cortex_dict* as_dict(AnyValue val) {
    if (val.type == TYPE_DICT && val.data.dict_val) return (cortex_dict*)val.data.dict_val;
    return NULL;
}

cortex_array* as_array(AnyValue val) {
    if (val.type == TYPE_ARRAY && val.data.array_val) return (cortex_array*)val.data.array_val;
    return NULL;
}

/* --- Events --- */
#define CORTEX_EVENT_CALLBACKS_INIT 8
struct cortex_event {
    cortex_event_callback* callbacks;
    int count;
    int cap;
};

cortex_event* event_create(void) {
    cortex_event* e = (cortex_event*)malloc(sizeof(cortex_event));
    if (!e) return NULL;
    e->cap = CORTEX_EVENT_CALLBACKS_INIT;
    e->count = 0;
    e->callbacks = (cortex_event_callback*)malloc((size_t)e->cap * sizeof(cortex_event_callback));
    if (!e->callbacks) { free(e); return NULL; }
    return e;
}

void event_subscribe(cortex_event* e, cortex_event_callback cb) {
    if (!e || !cb) return;
    if (e->count >= e->cap) {
        int newcap = e->cap * 2;
        cortex_event_callback* newb = (cortex_event_callback*)realloc(e->callbacks, (size_t)newcap * sizeof(cortex_event_callback));
        if (!newb) return;
        e->callbacks = newb;
        e->cap = newcap;
    }
    e->callbacks[e->count++] = cb;
}

void event_unsubscribe(cortex_event* e, cortex_event_callback cb) {
    if (!e || !cb) return;
    for (int i = 0; i < e->count; i++) {
        if (e->callbacks[i] == cb) {
            e->count--;
            memmove(&e->callbacks[i], &e->callbacks[i + 1], (size_t)(e->count - i) * sizeof(cortex_event_callback));
            return;
        }
    }
}

void event_emit(cortex_event* e, AnyValue val) {
    if (!e) return;
    for (int i = 0; i < e->count; i++)
        e->callbacks[i](val);
}

void event_free(cortex_event* e) {
    if (e) {
        free(e->callbacks);
        free(e);
    }
}

/* --- Result --- */
cortex_result result_ok(AnyValue v) {
    cortex_result r = { true, v, NULL };
    return r;
}

cortex_result result_err(const char* msg) {
    cortex_result r = { false, make_any_null(), msg ? strdup(msg) : NULL };
    return r;
}

bool result_is_ok(cortex_result r) {
    return r.ok;
}

AnyValue result_value(cortex_result r) {
    return r.value;
}

char* result_error(cortex_result r) {
    return r.error;
}

/* --- String utilities --- */
static char* strdup_safe(const char* s) {
    return s ? strdup(s) : NULL;
}
char* cortex_str_trim(const char* s) {
    if (!s) return NULL;
    while (*s && isspace((unsigned char)*s)) s++;
    size_t len = strlen(s);
    while (len > 0 && isspace((unsigned char)s[len - 1])) len--;
    char* out = (char*)malloc(len + 1);
    if (!out) return NULL;
    memcpy(out, s, len);
    out[len] = '\0';
    return out;
}
bool cortex_str_starts_with(const char* s, const char* prefix) {
    if (!s || !prefix) return false;
    size_t pl = strlen(prefix);
    return strncmp(s, prefix, pl) == 0;
}
bool cortex_str_ends_with(const char* s, const char* suffix) {
    if (!s || !suffix) return false;
    size_t sl = strlen(s), sul = strlen(suffix);
    return sl >= sul && strcmp(s + sl - sul, suffix) == 0;
}
char* cortex_str_to_lower(const char* s) {
    if (!s) return NULL;
    char* out = strdup(s);
    if (!out) return NULL;
    for (char* p = out; *p; p++) *p = (char)tolower((unsigned char)*p);
    return out;
}
char* cortex_str_to_upper(const char* s) {
    if (!s) return NULL;
    char* out = strdup(s);
    if (!out) return NULL;
    for (char* p = out; *p; p++) *p = (char)toupper((unsigned char)*p);
    return out;
}
char* cortex_str_replace(const char* s, const char* from, const char* to) {
    if (!s || !from || !to) return strdup_safe(s);
    size_t fl = strlen(from), tl = strlen(to);
    if (fl == 0) return strdup(s);
    size_t cap = strlen(s) + 1;
    char* out = (char*)malloc(cap);
    if (!out) return NULL;
    size_t j = 0;
    const char* p = s;
    while (*p) {
        if (strncmp(p, from, fl) == 0) {
            while (j + tl >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(out); return NULL; } out = n; }
            memcpy(out + j, to, tl + 1);
            j += tl;
            p += fl;
        } else {
            if (j + 1 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(out); return NULL; } out = n; }
            out[j++] = *p++;
        }
    }
    out[j] = '\0';
    return out;
}
char* cortex_str_join(const char** parts, int count, const char* sep) {
    if (!parts || count <= 0) return strdup("");
    if (!sep) sep = "";
    size_t seplen = strlen(sep), tot = 0;
    for (int i = 0; i < count; i++) tot += (parts[i] ? strlen(parts[i]) : 0) + (i > 0 ? seplen : 0);
    char* out = (char*)malloc(tot + 1);
    if (!out) return NULL;
    out[0] = '\0';
    char* cur = out;
    for (int i = 0; i < count; i++) {
        if (i > 0) { memcpy(cur, sep, seplen + 1); cur += seplen; }
        if (parts[i]) { size_t l = strlen(parts[i]); memcpy(cur, parts[i], l + 1); cur += l; }
    }
    return out;
}
cortex_array* cortex_str_split(const char* s, const char* delim) {
    cortex_array* out = array_create();
    if (!out || !s || !delim) return out;
    size_t dlen = strlen(delim);
    if (dlen == 0) { array_push(out, make_any_string(strdup(s))); return out; }
    const char* start = s;
    while (1) {
        const char* found = strstr(start, delim);
        const char* end = found ? found : start + strlen(start);
        size_t len = (size_t)(end - start);
        char* part = (char*)malloc(len + 1);
        if (!part) break;
        memcpy(part, start, len);
        part[len] = '\0';
        array_push(out, make_any_string(part));
        if (!found) break;
        start = found + dlen;
    }
    return out;
}
char* cortex_str_join_array(cortex_array* parts, const char* sep) {
    if (!parts || !sep) return strdup("");
    int n = array_len(parts);
    if (n == 0) return strdup("");
    size_t seplen = strlen(sep), tot = 0;
    for (int i = 0; i < n; i++) {
        AnyValue v = array_get(parts, i);
        tot += (v.type == TYPE_STRING && v.data.string_val) ? strlen(v.data.string_val) : 0;
        if (i > 0) tot += seplen;
    }
    char* out = (char*)malloc(tot + 1);
    if (!out) return NULL;
    out[0] = '\0';
    char* cur = out;
    for (int i = 0; i < n; i++) {
        if (i > 0) { memcpy(cur, sep, seplen + 1); cur += seplen; }
        AnyValue v = array_get(parts, i);
        if (v.type == TYPE_STRING && v.data.string_val) { size_t l = strlen(v.data.string_val); memcpy(cur, v.data.string_val, l + 1); cur += l; }
    }
    return out;
}

/* --- Math --- */
int clamp_int(int x, int lo, int hi) {
    if (x < lo) return lo;
    if (x > hi) return hi;
    return x;
}
double cortex_pow(double base, double exp) {
    return pow(base, exp);
}
AnyValue array_random_choice(cortex_array* a) {
    AnyValue out = make_any_null();
    if (!a || a->len == 0) return out;
    int i = rand() % a->len;
    return a->data[i];
}

/* --- File/path --- */
bool file_exists(const char* path) {
    if (!path) return false;
    FILE* f = fopen(path, "r");
    if (f) { fclose(f); return true; }
    return false;
}
char* path_join(const char* a, const char* b) {
    if (!a) return strdup_safe(b);
    if (!b) return strdup(a);
    size_t al = strlen(a), bl = strlen(b);
    int need_sep = (al > 0 && a[al - 1] != '/' && a[al - 1] != '\\' && bl > 0 && b[0] != '/' && b[0] != '\\');
    char* out = (char*)malloc(al + bl + (need_sep ? 2 : 1));
    if (!out) return NULL;
    memcpy(out, a, al + 1);
    if (need_sep) {
#ifdef _WIN32
        out[al] = '\\';
#else
        out[al] = '/';
#endif
        out[al + 1] = '\0';
        strcat(out, b);
    } else
        strcat(out, b);
    return out;
}
char* list_dir(const char* path) {
    if (!path) return strdup("[]");
    
    cortex_array* files = array_create();
    if (!files) return strdup("[]");
    
#ifdef _WIN32
    WIN32_FIND_DATAA findData;
    HANDLE hFind = INVALID_HANDLE_VALUE;
    
    // Append wildcard to path
    char searchPath[MAX_PATH];
    snprintf(searchPath, sizeof(searchPath), "%s\\*", path);
    
    hFind = FindFirstFileA(searchPath, &findData);
    if (hFind != INVALID_HANDLE_VALUE) {
        do {
            if (strcmp(findData.cFileName, ".") != 0 && strcmp(findData.cFileName, "..") != 0) {
                array_push(files, make_any_string(strdup(findData.cFileName)));
            }
        } while (FindNextFileA(hFind, &findData) != 0);
        FindClose(hFind);
    }
#else
    DIR* dir = opendir(path);
    if (dir) {
        struct dirent* entry;
        while ((entry = readdir(dir)) != NULL) {
            if (strcmp(entry->d_name, ".") != 0 && strcmp(entry->d_name, "..") != 0) {
                array_push(files, make_any_string(strdup(entry->d_name)));
            }
        }
        closedir(dir);
    }
#endif
    
    char* result = json_stringify_any(make_any_array(files));
    array_free(files);
    return result;
}

// --- String utilities (enhanced) ---
char* string_format(const char* fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    
    // First pass: calculate required size
    va_list ap_copy;
    va_copy(ap_copy, ap);
    int size = vsnprintf(NULL, 0, fmt, ap_copy);
    va_end(ap_copy);
    
    if (size < 0) {
        va_end(ap);
        return strdup("");
    }
    
    // Allocate and format
    char* result = (char*)malloc((size_t)size + 1);
    if (result) {
        vsnprintf(result, (size_t)size + 1, fmt, ap);
    }
    va_end(ap);
    
    return result ? result : strdup("");
}

char* string_pad_left(const char* s, int width, char pad) {
    if (!s) return NULL;
    int len = strlen(s);
    if (len >= width) return strdup(s);
    
    int pad_len = width - len;
    char* result = (char*)malloc((size_t)width + 1);
    if (!result) return NULL;
    
    // Add padding
    for (int i = 0; i < pad_len; i++) result[i] = pad;
    // Add string
    strcpy(result + pad_len, s);
    
    return result;
}

char* string_pad_right(const char* s, int width, char pad) {
    if (!s) return NULL;
    int len = strlen(s);
    if (len >= width) return strdup(s);
    
    char* result = (char*)malloc((size_t)width + 1);
    if (!result) return NULL;
    
    strcpy(result, s);
    // Add padding
    for (int i = len; i < width; i++) result[i] = pad;
    result[width] = '\0';
    
    return result;
}

char* string_center(const char* s, int width, char pad) {
    if (!s) return NULL;
    int len = strlen(s);
    if (len >= width) return strdup(s);
    
    int total_pad = width - len;
    int left_pad = total_pad / 2;
    int right_pad = total_pad - left_pad;
    
    char* result = (char*)malloc((size_t)width + 1);
    if (!result) return NULL;
    
    // Left padding
    for (int i = 0; i < left_pad; i++) result[i] = pad;
    // String
    strcpy(result + left_pad, s);
    // Right padding
    for (int i = len + left_pad; i < width; i++) result[i] = pad;
    result[width] = '\0';
    
    return result;
}

char* string_reverse(const char* s) {
    if (!s) return NULL;
    int len = strlen(s);
    char* result = (char*)malloc((size_t)len + 1);
    if (!result) return NULL;
    
    for (int i = 0; i < len; i++) {
        result[i] = s[len - 1 - i];
    }
    result[len] = '\0';
    
    return result;
}

int string_index_of(const char* s, const char* sub) {
    if (!s || !sub) return -1;
    char* found = strstr(s, sub);
    return found ? (int)(found - s) : -1;
}

int string_last_index_of(const char* s, const char* sub) {
    if (!s || !sub) return -1;
    
    int s_len = strlen(s);
    int sub_len = strlen(sub);
    if (sub_len > s_len) return -1;
    
    for (int i = s_len - sub_len; i >= 0; i--) {
        if (strncmp(s + i, sub, (size_t)sub_len) == 0) {
            return i;
        }
    }
    return -1;
}

bool string_contains(const char* s, const char* sub) {
    return string_index_of(s, sub) != -1;
}

// --- Math utilities (enhanced) ---
int abs_int(int x) { return x < 0 ? -x : x; }
float abs_float(float x) { return x < 0 ? -x : x; }
double abs_double(double x) { return x < 0 ? -x : x; }

int min_int(int a, int b) { return a < b ? a : b; }
int max_int(int a, int b) { return a > b ? a : b; }

double pow_double(double base, double exp) { return pow(base, exp); }
double sqrt_double(double x) { return sqrt(x); }
double sin_double(double x) { return sin(x); }
double cos_double(double x) { return cos(x); }
double tan_double(double x) { return tan(x); }
double floor_double(double x) { return floor(x); }
double ceil_double(double x) { return ceil(x); }
double round_double(double x) { return round(x); }

// --- System utilities ---
char* get_env(const char* name) {
    if (!name) return NULL;
    char* value = getenv(name);
    return value ? strdup(value) : NULL;
}

bool set_env(const char* name, const char* value) {
    if (!name || !value) return false;
#ifdef _WIN32
    return SetEnvironmentVariableA(name, value) != 0;
#else
    return setenv(name, value, 1) == 0;
#endif
}

char* get_cwd(void) {
    char* buffer = (char*)malloc(1024);
    if (!buffer) return NULL;
    
#ifdef _WIN32
    if (GetCurrentDirectoryA(1023, buffer) == 0) {
        free(buffer);
        return NULL;
    }
#else
    if (!getcwd(buffer, 1023)) {
        free(buffer);
        return NULL;
    }
#endif
    return buffer;
}

bool change_dir(const char* path) {
    if (!path) return false;
    return chdir(path) == 0;
}

int system_run(const char* command) {
    if (!command) return -1;
    return system(command);
}

char* get_username(void) {
#ifdef _WIN32
    char buffer[256];
    DWORD size = sizeof(buffer) - 1;
    if (GetUserNameA(buffer, &size)) {
        buffer[size] = '\0';
        return strdup(buffer);
    }
#else
    char* user = getenv("USER");
    if (!user) user = getenv("USERNAME");
    return user ? strdup(user) : strdup("unknown");
#endif
    return strdup("unknown");
}

// --- Memory utilities ---
void* mem_alloc(size_t size) { return malloc(size); }
void* mem_realloc(void* ptr, size_t size) { return realloc(ptr, size); }
void mem_free(void* ptr) { free(ptr); }
void* mem_copy(void* dest, const void* src, size_t n) { return memcpy(dest, src, n); }
void* mem_move(void* dest, const void* src, size_t n) { return memmove(dest, src, n); }
void* mem_set(void* ptr, int value, size_t n) { return memset(ptr, value, n); }
int mem_compare(const void* a, const void* b, size_t n) { return memcmp(a, b, n); }

/* --- Debug --- */
void debug_log(const char* fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    fprintf(stderr, "[cortex] ");
    vfprintf(stderr, fmt, ap);
    fprintf(stderr, "\n");
    va_end(ap);
}
void debug_assert(int condition, const char* msg, int line) {
    if (!condition) {
        fprintf(stderr, "cortex: debug_assert failed at line %d: %s\n", line, msg ? msg : "");
        abort();
    }
}
void dump_any(AnyValue v) {
    fprintf(stderr, "[cortex dump] type=%s ", type_of(v));
    switch (v.type) {
        case TYPE_INT: fprintf(stderr, "int_val=%d\n", v.data.int_val); break;
        case TYPE_FLOAT: fprintf(stderr, "float_val=%g\n", v.data.float_val); break;
        case TYPE_STRING: fprintf(stderr, "string_val=%s\n", v.data.string_val ? v.data.string_val : "(null)"); break;
        case TYPE_BOOL: fprintf(stderr, "bool_val=%s\n", v.data.bool_val ? "true" : "false"); break;
        default: fprintf(stderr, "\n"); break;
    }
}

/* --- Unit test --- */
#define CORTEX_MAX_TESTS 64
static struct { const char* name; void (*fn)(void); } cortex_tests[CORTEX_MAX_TESTS];
static int cortex_test_count = 0;
void test_register(const char* name, void (*fn)(void)) {
    if (cortex_test_count < CORTEX_MAX_TESTS && name && fn) {
        cortex_tests[cortex_test_count].name = name;
        cortex_tests[cortex_test_count].fn = fn;
        cortex_test_count++;
    }
}
int test_run_all(void) {
    int failed = 0;
    for (int i = 0; i < cortex_test_count; i++) {
        fprintf(stderr, "  test %s ... ", cortex_tests[i].name);
        fflush(stderr);
        cortex_tests[i].fn();
        fprintf(stderr, "ok\n");
    }
    return failed;
}
void assert_eq_int(int a, int b, const char* file, int line) {
    if (a != b) {
        fprintf(stderr, "%s:%d: assert_eq failed: %d != %d\n", file ? file : "?", line, a, b);
        abort();
    }
}
void assert_eq_float(float a, float b, float epsilon, const char* file, int line) {
    (void)file;
    (void)line;
    float d = a - b;
    if (d < 0) d = -d;
    if (d > epsilon) {
        fprintf(stderr, "assert_approx failed: %g != %g (eps %g)\n", (double)a, (double)b, (double)epsilon);
        abort();
    }
}
void assert_eq_str(const char* a, const char* b, const char* file, int line) {
    if (!a) a = ""; if (!b) b = "";
    if (strcmp(a, b) != 0) {
        fprintf(stderr, "%s:%d: assert_eq failed: \"%s\" != \"%s\"\n", file ? file : "?", line, a, b);
        abort();
    }
}

/* --- JSON (minimal: objects with string keys, values = number/string/bool/null) --- */
static AnyValue json_parse_value(const char** inp);

static const char* json_skip_ws(const char* p) {
    while (*p == ' ' || *p == '\t' || *p == '\n' || *p == '\r') p++;
    return p;
}
static char* json_parse_string(const char** inp) {
    const char* p = *inp;
    if (*p != '"') return NULL;
    p++;
    size_t cap = 32, len = 0;
    char* out = (char*)malloc(cap);
    if (!out) return NULL;
    while (*p && *p != '"') {
        if (*p == '\\') {
            p++;
            if (*p == 'n') { if (len + 1 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(out); return NULL; } out = n; } out[len++] = '\n'; p++; }
            else if (*p == 't') { if (len + 1 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(out); return NULL; } out = n; } out[len++] = '\t'; p++; }
            else if (*p == '"') { if (len + 1 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(out); return NULL; } out = n; } out[len++] = '"'; p++; }
            else if (*p == '\\') { if (len + 1 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(out); return NULL; } out = n; } out[len++] = '\\'; p++; }
            else { if (len + 1 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(out); return NULL; } out = n; } out[len++] = *p++; }
        } else {
            if (len + 1 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(out); return NULL; } out = n; }
            out[len++] = *p++;
        }
    }
    if (*p != '"') { free(out); return NULL; }
    p++;
    out[len] = '\0';
    *inp = p;
    return out;
}
/* Parse a JSON object {...}; updates *inp, returns new dict or NULL on error. */
static cortex_dict* json_parse_object(const char** inp) {
    const char* p = json_skip_ws(*inp);
    if (*p != '{') return NULL;
    cortex_dict* d = dict_create();
    if (!d) return NULL;
    p = json_skip_ws(p + 1);
    if (*p == '}') { *inp = p + 1; return d; }
    while (1) {
        char* key = json_parse_string(&p);
        if (!key) { dict_free(d); return NULL; }
        p = json_skip_ws(p);
        if (*p != ':') { free(key); dict_free(d); return NULL; }
        p = json_skip_ws(p + 1);
        AnyValue val = json_parse_value(&p);
        dict_set(d, key, val);
        free(key);
        p = json_skip_ws(p);
        if (*p == '}') { *inp = p + 1; return d; }
        if (*p != ',') { dict_free(d); return NULL; }
        p = json_skip_ws(p + 1);
    }
}

/* Parse a JSON array [...]; updates *inp, returns new array or NULL on error. */
static cortex_array* json_parse_array(const char** inp) {
    const char* p = json_skip_ws(*inp);
    if (*p != '[') return NULL;
    cortex_array* a = array_create();
    if (!a) return NULL;
    p = json_skip_ws(p + 1);
    if (*p == ']') { *inp = p + 1; return a; }
    while (1) {
        AnyValue val = json_parse_value(&p);
        array_push(a, val);
        p = json_skip_ws(p);
        if (*p == ']') { *inp = p + 1; return a; }
        if (*p != ',') { array_free(a); return NULL; }
        p = json_skip_ws(p + 1);
    }
}

static AnyValue json_parse_value(const char** inp) {
    const char* p = json_skip_ws(*inp);
    AnyValue out = make_any_null();
    if (*p == '"') {
        char* s = json_parse_string(&p);
        if (s) { out = make_any_string(s); *inp = p; }
        return out;
    }
    if (*p == '{') {
        cortex_dict* d = json_parse_object(inp);
        if (d) return make_any_dict(d);
        return out;
    }
    if (*p == '[') {
        cortex_array* a = json_parse_array(inp);
        if (a) return make_any_array(a);
        return out;
    }
    if (*p == 'n' && strncmp(p, "null", 4) == 0) { *inp = p + 4; return make_any_null(); }
    if (*p == 't' && strncmp(p, "true", 4) == 0) { *inp = p + 4; return make_any_bool(true); }
    if (*p == 'f' && strncmp(p, "false", 5) == 0) { *inp = p + 5; return make_any_bool(false); }
    if (*p == '-' || (*p >= '0' && *p <= '9')) {
        char* end = NULL;
        double d = strtod(p, &end);
        if (end && end > p) {
            *inp = end;
            if (d == (double)(int)d) return make_any_int((int)d);
            return make_any_float((float)d);
        }
    }
    *inp = p;
    return out;
}
cortex_dict* json_parse(const char* s) {
    cortex_dict* d = dict_create();
    if (!d || !s) return d;
    const char* p = json_skip_ws(s);
    if (*p != '{') return d;
    p = json_skip_ws(p + 1);
    if (*p == '}') return d;
    while (1) {
        char* key = json_parse_string(&p);
        if (!key) break;
        p = json_skip_ws(p);
        if (*p != ':') { free(key); break; }
        p = json_skip_ws(p + 1);
        AnyValue val = json_parse_value(&p);
        dict_set(d, key, val);
        free(key);
        p = json_skip_ws(p);
        if (*p == '}') break;
        if (*p != ',') break;
        p = json_skip_ws(p + 1);
    }
    return d;
}
static void json_append_escaped(char** out, size_t* cap, size_t* len, const char* s) {
    if (!s) return;
    for (const char* p = s; *p; p++) {
        if (*len + 8 >= *cap) { *cap *= 2; char* n = (char*)realloc(*out, *cap); if (!n) return; *out = n; }
        if (*p == '"' || *p == '\\') { (*out)[(*len)++] = '\\'; (*out)[(*len)++] = *p; }
        else if (*p == '\n') { (*out)[(*len)++] = '\\'; (*out)[(*len)++] = 'n'; }
        else if (*p == '\t') { (*out)[(*len)++] = '\\'; (*out)[(*len)++] = 't'; }
        else (*out)[(*len)++] = *p;
    }
}
char* json_stringify_any(AnyValue v) {
    size_t cap = 64, len = 0;
    char* out = (char*)malloc(cap);
    if (!out) return NULL;
    switch (v.type) {
        case TYPE_NULL: memcpy(out, "null", 5); len = 4; break;
        case TYPE_BOOL: memcpy(out, v.data.bool_val ? "true" : "false", v.data.bool_val ? 5 : 6); len = v.data.bool_val ? 4 : 5; break;
        case TYPE_INT: len = (size_t)snprintf(out, cap, "%d", v.data.int_val); break;
        case TYPE_FLOAT: len = (size_t)snprintf(out, cap, "%g", v.data.float_val); break;
        case TYPE_STRING: out[0] = '"'; len = 1; json_append_escaped(&out, &cap, &len, v.data.string_val); while (len + 2 > cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) break; out = n; } out[len++] = '"'; out[len] = '\0'; break;
        case TYPE_DICT: {
            free(out);
            return json_stringify_dict((cortex_dict*)v.data.dict_val);
        }
        case TYPE_ARRAY: {
            cortex_array* a = (cortex_array*)v.data.array_val;
            free(out);
            if (!a || a->len == 0) return strdup("[]");
            out = (char*)malloc(cap = 128);
            if (!out) return NULL;
            out[len++] = '[';
            for (int i = 0; i < a->len; i++) {
                char* elem = json_stringify_any(a->data[i]);
                if (elem) {
                    size_t elen = strlen(elem);
                    while (len + elen + 2 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(elem); free(out); return NULL; } out = n; }
                    memcpy(out + len, elem, elen + 1);
                    len += elen;
                    free(elem);
                }
                if (i < a->len - 1) { if (len + 1 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(out); return NULL; } out = n; } out[len++] = ','; }
            }
            if (len + 2 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(out); return NULL; } out = n; }
            out[len++] = ']';
            out[len] = '\0';
            return out;
        }
        default: memcpy(out, "null", 5); len = 4; break;
    }
    out[len] = '\0';
    return out;
}
char* json_stringify_dict(cortex_dict* d) {
    if (!d || d->count == 0) return strdup("{}");
    size_t cap = 128, len = 0;
    char* out = (char*)malloc(cap);
    if (!out) return NULL;
    out[len++] = '{';
    int first = 1;
    for (int i = 0; i < d->cap; i++) {
        for (dict_entry* e = d->buckets[i]; e; e = e->next) {
            if (!first) { if (len + 1 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(out); return NULL; } out = n; } out[len++] = ','; }
            first = 0;
            if (len + 4 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(out); return NULL; } out = n; }
            out[len++] = '"';
            size_t oldlen = len;
            json_append_escaped(&out, &cap, &len, e->key);
            if (len + 4 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(out); return NULL; } out = n; }
            out[len++] = '"';
            out[len++] = ':';
            char* valStr = json_stringify_any(e->val);
            if (valStr) {
                size_t vlen = strlen(valStr);
                while (len + vlen + 1 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(valStr); free(out); return NULL; } out = n; }
                memcpy(out + len, valStr, vlen + 1);
                len += vlen;
                free(valStr);
            }
        }
    }
    if (len + 2 >= cap) { cap *= 2; char* n = (char*)realloc(out, cap); if (!n) { free(out); return NULL; } out = n; }
    out[len++] = '}';
    out[len] = '\0';
    return out;
}

/* Parse string to number; returns 0 / 0.0 on parse failure. */
float parse_number(const char* s) {
    if (!s || !*s) return 0.0f;
    char* end = NULL;
    double d = strtod(s, &end);
    if (end == s) return 0.0f;
    return (float)d;
}
int parse_int(const char* s) {
    if (!s || !*s) return 0;
    char* end = NULL;
    long n = strtol(s, &end, 10);
    if (end == s) return 0;
    return (int)n;
}

/* --- ECS helpers: entity_id = 1-based index into global entity array (each entity = dict of components) --- */
static cortex_array* g_entities;

int entity_create(void) {
    if (!g_entities) g_entities = array_create();
    if (!g_entities) return 0;
    cortex_dict* d = dict_create();
    if (!d) return 0;
    array_push(g_entities, make_any_dict(d));
    return array_len(g_entities);
}

void add_component(int entity_id, const char* component_name, AnyValue val) {
    if (!g_entities || entity_id < 1 || entity_id > array_len(g_entities)) return;
    AnyValue ev = array_get(g_entities, entity_id - 1);
    cortex_dict* d = as_dict(ev);
    if (d) dict_set(d, component_name, val);
}

AnyValue get_component(int entity_id, const char* component_name) {
    AnyValue out = make_any_null();
    if (!g_entities || entity_id < 1 || entity_id > array_len(g_entities)) return out;
    AnyValue ev = array_get(g_entities, entity_id - 1);
    cortex_dict* d = as_dict(ev);
    if (d) return dict_get(d, component_name);
    return out;
}

bool has_component(int entity_id, const char* component_name) {
    if (!g_entities || entity_id < 1 || entity_id > array_len(g_entities)) return false;
    AnyValue ev = array_get(g_entities, entity_id - 1);
    cortex_dict* d = as_dict(ev);
    return d && dict_has(d, component_name);
}

void entity_remove(int entity_id) {
    if (!g_entities || entity_id < 1 || entity_id > array_len(g_entities)) return;
    AnyValue ev = array_get(g_entities, entity_id - 1);
    cortex_dict* d = as_dict(ev);
    if (d) dict_free(d);
    array_set(g_entities, entity_id - 1, make_any_null());
}

/* Stub for examples/lambda_capture.cx; link your own to override. */
void apply_twice(void* fn, void* env, int x) { (void)fn; (void)env; (void)x; }

#else
vec2 make_vec2(float x, float y) { (void)x; (void)y; return (vec2){0,0}; }
vec3 make_vec3(float x, float y, float z) { (void)x; (void)y; (void)z; return (vec3){0,0,0}; }
float dot(vec2 a, vec2 b) { (void)a; (void)b; return 0.0f; }
vec2 normalize(vec2 v) { (void)v; return (vec2){0,0}; }
float vec2_length(vec2 v) { (void)v; return 0.0f; }
float vec2_length_sq(vec2 v) { (void)v; return 0.0f; }
float vec2_distance(vec2 a, vec2 b) { (void)a; (void)b; return 0.0f; }
vec2 vec2_add(vec2 a, vec2 b) { (void)a; (void)b; return (vec2){0,0}; }
vec2 vec2_sub(vec2 a, vec2 b) { (void)a; (void)b; return (vec2){0,0}; }
vec2 vec2_scale(vec2 v, float s) { (void)v; (void)s; return (vec2){0,0}; }
float vec3_length(vec3 v) { (void)v; return 0.0f; }
float vec3_length_sq(vec3 v) { (void)v; return 0.0f; }
float vec3_dot(vec3 a, vec3 b) { (void)a; (void)b; return 0.0f; }
vec3 vec3_normalize(vec3 v) { (void)v; return (vec3){0,0,0}; }
vec3 vec3_add(vec3 a, vec3 b) { (void)a; (void)b; return (vec3){0,0,0}; }
vec3 vec3_sub(vec3 a, vec3 b) { (void)a; (void)b; return (vec3){0,0,0}; }
vec3 vec3_scale(vec3 v, float s) { (void)v; (void)s; return (vec3){0,0,0}; }
float vec3_distance(vec3 a, vec3 b) { (void)a; (void)b; return 0.0f; }
float clamp_float(float x, float lo, float hi) { (void)x; (void)lo; (void)hi; return 0.0f; }
float lerp_float(float a, float b, float t) { (void)a; (void)b; (void)t; return 0.0f; }
float min_float(float a, float b) { (void)a; (void)b; return 0.0f; }
float max_float(float a, float b) { (void)a; (void)b; return 0.0f; }
float sign_float(float x) { (void)x; return 0.0f; }
float wrap_float(float x, float lo, float hi) { (void)x; (void)lo; (void)hi; return 0.0f; }
float round_float(float x) { (void)x; return 0.0f; }
float floor_float(float x) { (void)x; return 0.0f; }
float ceil_float(float x) { (void)x; return 0.0f; }
cortex_array* array_create(void) { return NULL; }
void array_push(cortex_array* a, AnyValue val) { (void)a; (void)val; }
AnyValue array_get(cortex_array* a, int index) { (void)a; (void)index; return make_any_null(); }
void array_set(cortex_array* a, int index, AnyValue val) { (void)a; (void)index; (void)val; }
int array_len(cortex_array* a) { (void)a; return 0; }
int array_capacity(cortex_array* a) { (void)a; return 0; }
void array_reserve(cortex_array* a, int min_cap) { (void)a; (void)min_cap; }
AnyValue array_pop(cortex_array* a) { (void)a; return make_any_null(); }
void array_insert(cortex_array* a, int index, AnyValue val) { (void)a; (void)index; (void)val; }
void array_remove_at(cortex_array* a, int index) { (void)a; (void)index; }
void array_free(cortex_array* a) { (void)a; }
cortex_event* event_create(void) { return NULL; }
void event_subscribe(cortex_event* e, cortex_event_callback cb) { (void)e; (void)cb; }
void event_unsubscribe(cortex_event* e, cortex_event_callback cb) { (void)e; (void)cb; }
void event_emit(cortex_event* e, AnyValue val) { (void)e; (void)val; }
void event_free(cortex_event* e) { (void)e; }
cortex_dict* dict_create(void) { return NULL; }
void dict_set(cortex_dict* d, const char* key, AnyValue val) { (void)d; (void)key; (void)val; }
AnyValue dict_get(cortex_dict* d, const char* key) { (void)d; (void)key; return make_any_null(); }
bool dict_has(cortex_dict* d, const char* key) { (void)d; (void)key; return false; }
int dict_len(cortex_dict* d) { (void)d; return 0; }
void dict_free(cortex_dict* d) { (void)d; }
AnyValue make_any_dict(cortex_dict* d) { (void)d; return (AnyValue){TYPE_NULL}; }
AnyValue make_any_array(cortex_array* a) { (void)a; return (AnyValue){TYPE_NULL}; }
cortex_dict* as_dict(AnyValue val) { (void)val; return NULL; }
cortex_array* as_array(AnyValue val) { (void)val; return NULL; }
cortex_result result_ok(AnyValue v) { cortex_result r = { true, v, NULL }; return r; }
cortex_result result_err(const char* msg) { (void)msg; cortex_result r = { false, make_any_null(), NULL }; return r; }
bool result_is_ok(cortex_result r) { return r.ok; }
AnyValue result_value(cortex_result r) { return r.value; }
char* result_error(cortex_result r) { return r.error; }
cortex_array* cortex_str_split(const char* s, const char* delim) { (void)s; (void)delim; return NULL; }
char* cortex_str_join_array(cortex_array* parts, const char* sep) { (void)parts; (void)sep; return NULL; }
char* cortex_str_replace(const char* s, const char* from, const char* to) { (void)s; (void)from; (void)to; return NULL; }
char* cortex_str_trim(const char* s) { (void)s; return NULL; }
bool cortex_str_starts_with(const char* s, const char* prefix) { (void)s; (void)prefix; return false; }
bool cortex_str_ends_with(const char* s, const char* suffix) { (void)s; (void)suffix; return false; }
char* cortex_str_to_lower(const char* s) { (void)s; return NULL; }
char* cortex_str_to_upper(const char* s) { (void)s; return NULL; }
int clamp_int(int x, int lo, int hi) { (void)x; (void)lo; (void)hi; return 0; }
double cortex_pow(double base, double exp) { (void)base; (void)exp; return 0; }
AnyValue array_random_choice(cortex_array* a) { (void)a; return make_any_null(); }
bool file_exists(const char* path) { (void)path; return false; }
int append_file(const char* path, const char* content) { (void)path; (void)content; return -1; }
char* read_file_lines(const char* path) { (void)path; return strdup("[]"); }
int write_file_lines(const char* path, cortex_array* lines) { (void)path; (void)lines; return -1; }
long file_size(const char* path) { (void)path; return -1; }
bool file_copy(const char* src, const char* dst) { (void)src; (void)dst; return false; }
bool file_move(const char* src, const char* dst) { (void)src; (void)dst; return false; }
bool file_delete(const char* path) { (void)path; return false; }
time_t file_modified(const char* path) { (void)path; return 0; }
bool file_create_dir(const char* path) { (void)path; return false; }
char* string_format(const char* fmt, ...) { (void)fmt; return strdup(""); }
char* string_pad_left(const char* s, int width, char pad) { (void)s; (void)width; (void)pad; return NULL; }
char* string_pad_right(const char* s, int width, char pad) { (void)s; (void)width; (void)pad; return NULL; }
char* string_center(const char* s, int width, char pad) { (void)s; (void)width; (void)pad; return NULL; }
char* string_reverse(const char* s) { (void)s; return NULL; }
int string_index_of(const char* s, const char* sub) { (void)s; (void)sub; return -1; }
int string_last_index_of(const char* s, const char* sub) { (void)s; (void)sub; return -1; }
bool string_contains(const char* s, const char* sub) { (void)s; (void)sub; return false; }
int abs_int(int x) { (void)x; return 0; }
float abs_float(float x) { (void)x; return 0.0f; }
double abs_double(double x) { (void)x; return 0.0; }
int min_int(int a, int b) { (void)a; (void)b; return 0; }
int max_int(int a, int b) { (void)a; (void)b; return 0; }
double pow_double(double base, double exp) { (void)base; (void)exp; return 0.0; }
double sqrt_double(double x) { (void)x; return 0.0; }
double sin_double(double x) { (void)x; return 0.0; }
double cos_double(double x) { (void)x; return 0.0; }
double tan_double(double x) { (void)x; return 0.0; }
double floor_double(double x) { (void)x; return 0.0; }
double ceil_double(double x) { (void)x; return 0.0; }
double round_double(double x) { (void)x; return 0.0; }
char* get_env(const char* name) { (void)name; return NULL; }
bool set_env(const char* name, const char* value) { (void)name; (void)value; return false; }
char* get_cwd(void) { return NULL; }
bool change_dir(const char* path) { (void)path; return false; }
int system_run(const char* command) { (void)command; return -1; }
char* get_username(void) { return strdup("unknown"); }
void* mem_alloc(size_t size) { (void)size; return NULL; }
void* mem_realloc(void* ptr, size_t size) { (void)ptr; (void)size; return NULL; }
void mem_free(void* ptr) { (void)ptr; }
void* mem_copy(void* dest, const void* src, size_t n) { (void)dest; (void)src; (void)n; return NULL; }
void* mem_move(void* dest, const void* src, size_t n) { (void)dest; (void)src; (void)n; return NULL; }
void* mem_set(void* ptr, int value, size_t n) { (void)ptr; (void)value; (void)n; return NULL; }
int mem_compare(const void* a, const void* b, size_t n) { (void)a; (void)b; (void)n; return 0; }
char* path_join(const char* a, const char* b) { (void)a; (void)b; return NULL; }
void debug_log(const char* fmt, ...) { (void)fmt; }
void debug_assert(int condition, const char* msg, int line) { (void)condition; (void)msg; (void)line; }
void dump_any(AnyValue v) { (void)v; }
void test_register(const char* name, void (*fn)(void)) { (void)name; (void)fn; }
int test_run_all(void) { return 0; }
void assert_eq_int(int a, int b, const char* file, int line) { (void)a; (void)b; (void)file; (void)line; }
void assert_eq_float(float a, float b, float epsilon, const char* file, int line) { (void)a; (void)b; (void)epsilon; (void)file; (void)line; }
void assert_eq_str(const char* a, const char* b, const char* file, int line) { (void)a; (void)b; (void)file; (void)line; }
cortex_dict* json_parse(const char* s) { (void)s; return NULL; }
char* json_stringify_any(AnyValue v) { (void)v; return NULL; }
char* json_stringify_dict(cortex_dict* d) { (void)d; return NULL; }
float parse_number(const char* s) { (void)s; return 0.0f; }
int parse_int(const char* s) { (void)s; return 0; }
int entity_create(void) { return 0; }
void add_component(int entity_id, const char* component_name, AnyValue val) { (void)entity_id; (void)component_name; (void)val; }
AnyValue get_component(int entity_id, const char* component_name) { (void)entity_id; (void)component_name; return make_any_null(); }
bool has_component(int entity_id, const char* component_name) { (void)entity_id; (void)component_name; return false; }
void entity_remove(int entity_id) { (void)entity_id; }
void print_any(AnyValue v) { (void)v; }
void println_any(AnyValue v) { (void)v; }
int writeline_fmt(const char* fmt, ...) { (void)fmt; return 0; }
#endif

int random_int(int min, int max) {
    return min + rand() % (max - min + 1);
}

float random_float(float min, float max) {
    return min + ((float)rand() / RAND_MAX) * (max - min);
}

float get_time(void) {
    return (float)clock() / CLOCKS_PER_SEC;
}

void sleep_func(float seconds) {
#ifdef _WIN32
    Sleep((DWORD)(seconds * 1000));
#else
    struct timespec ts;
    ts.tv_sec = (time_t)seconds;
    ts.tv_nsec = (long)((seconds - ts.tv_sec) * 1e9);
    nanosleep(&ts, NULL);
#endif
}

void random_bytes(uint8_t* buffer, size_t len) {
    for (size_t i = 0; i < len; ++i) {
        buffer[i] = (uint8_t)(rand() & 0xFF);
    }
}

uint64_t unix_time_ms(void) {
#ifdef _WIN32
    FILETIME ft;
    GetSystemTimeAsFileTime(&ft);
    ULARGE_INTEGER uli;
    uli.LowPart = ft.dwLowDateTime;
    uli.HighPart = ft.dwHighDateTime;
    return (uli.QuadPart / 10000ULL) - 11644473600000ULL;
#else
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (uint64_t)tv.tv_sec * 1000ULL + (tv.tv_usec / 1000ULL);
#endif
}

void print_string(const char* str) {
    printf("%s", str);
}

void println_string(const char* str) {
    printf("%s\n", str);
}

#if CORTEX_FEATURE_QOL
void print_any(AnyValue v) {
    char buf[256];
    switch (v.type) {
        case TYPE_INT: snprintf(buf, sizeof(buf), "%d", v.data.int_val); print_string(buf); break;
        case TYPE_FLOAT: snprintf(buf, sizeof(buf), "%g", (double)v.data.float_val); print_string(buf); break;
        case TYPE_STRING: print_string(v.data.string_val ? v.data.string_val : ""); break;
        case TYPE_BOOL: print_string(v.data.bool_val ? "true" : "false"); break;
        case TYPE_NULL: print_string("null"); break;
        default: snprintf(buf, sizeof(buf), "(value)"); print_string(buf); break;
    }
}

void println_any(AnyValue v) {
    print_any(v);
    print_string("\n");
}

int writeline_fmt(const char* fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    int n = vprintf(fmt, ap);
    va_end(ap);
    putchar('\n');
    return n + 1;
}
#endif

char* cortex_strcat(const char* a, const char* b) {
    if (!a) a = "";
    if (!b) b = "";
    size_t na = strlen(a), nb = strlen(b);
    char* out = (char*)malloc(na + nb + 1);
    if (!out) return NULL;
    memcpy(out, a, na + 1);
    memcpy(out + na, b, nb + 1);
    return out;
}

#define INPUT_LINE_BUF 4096
char* input_line(void) {
    static char buf[INPUT_LINE_BUF];
    if (!fgets(buf, (int)sizeof(buf), stdin))
        return NULL;
    size_t n = strlen(buf);
    if (n > 0 && buf[n - 1] == '\n') {
        buf[n - 1] = '\0';
        n--;
    }
    char* out = (char*)malloc(n + 1);
    if (!out) return NULL;
    memcpy(out, buf, n + 1);
    return out;
}

void waitkey(void) {
#ifdef _WIN32
    // Windows: wait for any key press
    system("pause >nul");
#else
    // Linux/macOS: wait for Enter key
    printf("Press Enter to continue...");
    getchar();
#endif
}

char* read_file(const char* path) {
    FILE* f = fopen(path, "rb");
    if (!f) return NULL;
    if (fseek(f, 0, SEEK_END) != 0) { fclose(f); return NULL; }
    long size = ftell(f);
    if (size < 0 || fseek(f, 0, SEEK_SET) != 0) { fclose(f); return NULL; }
    char* out = (char*)malloc((size_t)size + 1);
    if (!out) { fclose(f); return NULL; }
    size_t n = fread(out, 1, (size_t)size, f);
    out[n] = '\0';
    fclose(f);
    return out;
}

int write_file(const char* path, const char* content) {
    FILE* f = fopen(path, "wb");
    if (!f) return -1;
    size_t len = strlen(content);
    int ok = (fwrite(content, 1, len, f) == len);
    fclose(f);
    return ok ? 0 : -1;
}

int append_file(const char* path, const char* content) {
    FILE* f = fopen(path, "a");
    if (!f) return -1;
    size_t len = strlen(content);
    int ok = (fwrite(content, 1, len, f) == len);
    fclose(f);
    return ok ? 0 : -1;
}

char* read_file_lines(const char* path) {
    FILE* f = fopen(path, "r");
    if (!f) return NULL;
    
    // Count lines first
    int line_count = 0;
    char ch;
    while ((ch = fgetc(f)) != EOF) {
        if (ch == '\n') line_count++;
    }
    rewind(f);
    
    // Allocate array for lines
    cortex_array* lines = array_create();
    if (!lines) { fclose(f); return NULL; }
    
    char buffer[1024];
    while (fgets(buffer, sizeof(buffer), f)) {
        // Remove trailing newline
        size_t len = strlen(buffer);
        if (len > 0 && buffer[len-1] == '\n') {
            buffer[len-1] = '\0';
            len--;
        }
        if (len > 0 && buffer[len-1] == '\r') {
            buffer[len-1] = '\0';
        }
        array_push(lines, make_any_string(strdup(buffer)));
    }
    fclose(f);
    
    // Convert to JSON array string
    return json_stringify_any(make_any_array(lines));
}

int write_file_lines(const char* path, cortex_array* lines) {
    FILE* f = fopen(path, "w");
    if (!f) return -1;
    
    for (int i = 0; i < array_len(lines); i++) {
        AnyValue v = array_get(lines, i);
        if (v.type == TYPE_STRING && v.data.string_val) {
            fprintf(f, "%s\n", v.data.string_val);
        }
    }
    fclose(f);
    return 0;
}

long file_size(const char* path) {
    FILE* f = fopen(path, "rb");
    if (!f) return -1;
    if (fseek(f, 0, SEEK_END) != 0) { fclose(f); return -1; }
    long size = ftell(f);
    fclose(f);
    return size;
}

bool file_copy(const char* src, const char* dst) {
    FILE* fsrc = fopen(src, "rb");
    if (!fsrc) return false;
    
    FILE* fdst = fopen(dst, "wb");
    if (!fdst) { fclose(fsrc); return false; }
    
    char buffer[4096];
    size_t n;
    bool ok = true;
    while ((n = fread(buffer, 1, sizeof(buffer), fsrc)) > 0) {
        if (fwrite(buffer, 1, n, fdst) != n) {
            ok = false;
            break;
        }
    }
    
    fclose(fsrc);
    fclose(fdst);
    return ok;
}

bool file_move(const char* src, const char* dst) {
    if (file_copy(src, dst)) {
        return remove(src) == 0;
    }
    return false;
}

bool file_delete(const char* path) {
    return remove(path) == 0;
}

time_t file_modified(const char* path) {
    struct stat st;
    if (stat(path, &st) != 0) return 0;
    return st.st_mtime;
}

bool file_create_dir(const char* path) {
#ifdef _WIN32
    return CreateDirectoryA(path, NULL) != 0;
#else
    return mkdir(path, 0755) == 0;
#endif
}

int cortex_bounds_check(int len, int index, int line) {
    if (index < 0 || index >= len) {
        fprintf(stderr, "cortex: array bounds check failed (index=%d, len=%d) at line %d\n", index, len, line);
        abort();
    }
    return index;
}

void cortex_assert_fail(int line, const char* msg) {
    fprintf(stderr, "cortex: assertion failed at line %d: %s\n", line, msg ? msg : "");
    abort();
}

// --- Blockchain helpers ---
#if CORTEX_FEATURE_BLOCKCHAIN

typedef struct {
    uint32_t state[8];
    uint64_t bitlen;
    uint32_t datalen;
    uint8_t data[64];
} sha256_ctx;

static const uint32_t sha256_k[64] = {
    0x428a2f98,0x71374491,0xb5c0fbcf,0xe9b5dba5,0x3956c25b,0x59f111f1,0x923f82a4,0xab1c5ed5,
    0xd807aa98,0x12835b01,0x243185be,0x550c7dc3,0x72be5d74,0x80deb1fe,0x9bdc06a7,0xc19bf174,
    0xe49b69c1,0xefbe4786,0x0fc19dc6,0x240ca1cc,0x2de92c6f,0x4a7484aa,0x5cb0a9dc,0x76f988da,
    0x983e5152,0xa831c66d,0xb00327c8,0xbf597fc7,0xc6e00bf3,0xd5a79147,0x06ca6351,0x14292967,
    0x27b70a85,0x2e1b2138,0x4d2c6dfc,0x53380d13,0x650a7354,0x766a0abb,0x81c2c92e,0x92722c85,
    0xa2bfe8a1,0xa81a664b,0xc24b8b70,0xc76c51a3,0xd192e819,0xd6990624,0xf40e3585,0x106aa070,
    0x19a4c116,0x1e376c08,0x2748774c,0x34b0bcb5,0x391c0cb3,0x4ed8aa4a,0x5b9cca4f,0x682e6ff3,
    0x748f82ee,0x78a5636f,0x84c87814,0x8cc70208,0x90befffa,0xa4506ceb,0xbef9a3f7,0xc67178f2
};

#define ROTRIGHT(a,b) (((a) >> (b)) | ((a) << (32-(b))))
#define CH(x,y,z) (((x) & (y)) ^ (~(x) & (z)))
#define MAJ(x,y,z) (((x) & (y)) ^ ((x) & (z)) ^ ((y) & (z)))
#define EP0(x) (ROTRIGHT(x,2) ^ ROTRIGHT(x,13) ^ ROTRIGHT(x,22))
#define EP1(x) (ROTRIGHT(x,6) ^ ROTRIGHT(x,11) ^ ROTRIGHT(x,25))
#define SIG0(x) (ROTRIGHT(x,7) ^ ROTRIGHT(x,18) ^ ((x) >> 3))
#define SIG1(x) (ROTRIGHT(x,17) ^ ROTRIGHT(x,19) ^ ((x) >> 10))

static void sha256_transform(sha256_ctx* ctx, const uint8_t data[]) {
    uint32_t m[64];
    for (uint32_t i = 0, j = 0; i < 16; ++i, j += 4)
        m[i] = (data[j] << 24) | (data[j+1] << 16) | (data[j+2] << 8) | data[j+3];
    for (uint32_t i = 16; i < 64; ++i)
        m[i] = SIG1(m[i-2]) + m[i-7] + SIG0(m[i-15]) + m[i-16];

    uint32_t a = ctx->state[0];
    uint32_t b = ctx->state[1];
    uint32_t c = ctx->state[2];
    uint32_t d = ctx->state[3];
    uint32_t e = ctx->state[4];
    uint32_t f = ctx->state[5];
    uint32_t g = ctx->state[6];
    uint32_t h = ctx->state[7];

    for (uint32_t i = 0; i < 64; ++i) {
        uint32_t t1 = h + EP1(e) + CH(e,f,g) + sha256_k[i] + m[i];
        uint32_t t2 = EP0(a) + MAJ(a,b,c);
        h = g;
        g = f;
        f = e;
        e = d + t1;
        d = c;
        c = b;
        b = a;
        a = t1 + t2;
    }

    ctx->state[0] += a;
    ctx->state[1] += b;
    ctx->state[2] += c;
    ctx->state[3] += d;
    ctx->state[4] += e;
    ctx->state[5] += f;
    ctx->state[6] += g;
    ctx->state[7] += h;
}

static void sha256_init(sha256_ctx* ctx) {
    ctx->datalen = 0;
    ctx->bitlen = 0;
    ctx->state[0] = 0x6a09e667;
    ctx->state[1] = 0xbb67ae85;
    ctx->state[2] = 0x3c6ef372;
    ctx->state[3] = 0xa54ff53a;
    ctx->state[4] = 0x510e527f;
    ctx->state[5] = 0x9b05688c;
    ctx->state[6] = 0x1f83d9ab;
    ctx->state[7] = 0x5be0cd19;
}

static void sha256_update(sha256_ctx* ctx, const uint8_t* data, size_t len) {
    for (size_t i = 0; i < len; ++i) {
        ctx->data[ctx->datalen++] = data[i];
        if (ctx->datalen == 64) {
            sha256_transform(ctx, ctx->data);
            ctx->bitlen += 512;
            ctx->datalen = 0;
        }
    }
}

static void sha256_final(sha256_ctx* ctx, uint8_t hash[32]) {
    uint32_t i = ctx->datalen;
    if (ctx->datalen < 56) {
        ctx->data[i++] = 0x80;
        while (i < 56) ctx->data[i++] = 0x00;
    } else {
        ctx->data[i++] = 0x80;
        while (i < 64) ctx->data[i++] = 0x00;
        sha256_transform(ctx, ctx->data);
        memset(ctx->data, 0, 56);
    }

    ctx->bitlen += ctx->datalen * 8;
    ctx->data[63] = ctx->bitlen;
    ctx->data[62] = ctx->bitlen >> 8;
    ctx->data[61] = ctx->bitlen >> 16;
    ctx->data[60] = ctx->bitlen >> 24;
    ctx->data[59] = ctx->bitlen >> 32;
    ctx->data[58] = ctx->bitlen >> 40;
    ctx->data[57] = ctx->bitlen >> 48;
    ctx->data[56] = ctx->bitlen >> 56;
    sha256_transform(ctx, ctx->data);

    for (i = 0; i < 4; ++i) {
        hash[i]      = (ctx->state[0] >> (24 - i * 8)) & 0xff;
        hash[i + 4]  = (ctx->state[1] >> (24 - i * 8)) & 0xff;
        hash[i + 8]  = (ctx->state[2] >> (24 - i * 8)) & 0xff;
        hash[i + 12] = (ctx->state[3] >> (24 - i * 8)) & 0xff;
        hash[i + 16] = (ctx->state[4] >> (24 - i * 8)) & 0xff;
        hash[i + 20] = (ctx->state[5] >> (24 - i * 8)) & 0xff;
        hash[i + 24] = (ctx->state[6] >> (24 - i * 8)) & 0xff;
        hash[i + 28] = (ctx->state[7] >> (24 - i * 8)) & 0xff;
    }
}

void sha256_hash(const uint8_t* data, size_t len, uint8_t out[32]) {
    sha256_ctx ctx;
    sha256_init(&ctx);
    sha256_update(&ctx, data, len);
    sha256_final(&ctx, out);
}

void sha256_double(const uint8_t* data, size_t len, uint8_t out[32]) {
    uint8_t tmp[32];
    sha256_hash(data, len, tmp);
    sha256_hash(tmp, 32, out);
}

// --- Keccak, Merkle, Base58, blockheader hashing (ported from generator.go) ---
static const uint64_t keccak_round_constants[24] = {
    0x0000000000000001ULL, 0x0000000000008082ULL, 0x800000000000808AULL,
    0x8000000080008000ULL, 0x000000000000808BULL, 0x0000000080000001ULL,
    0x8000000080008081ULL, 0x8000000000008009ULL, 0x000000000000008AULL,
    0x0000000000000088ULL, 0x0000000080008009ULL, 0x000000008000000AULL,
    0x000000008000808BULL, 0x800000000000008BULL, 0x8000000000008089ULL,
    0x8000000000008003ULL, 0x8000000000008002ULL, 0x8000000000000080ULL,
    0x000000000000800AULL, 0x800000008000000AULL, 0x8000000080008081ULL,
    0x8000000000008080ULL, 0x0000000080000001ULL, 0x8000000080008008ULL
};

static uint64_t keccak_rotl(uint64_t x, uint64_t y) {
    return (x << y) | (x >> (64 - y));
}

static void keccakf(uint64_t state[25]) {
    static const uint32_t rho[25] = {
        0,  1, 62, 28, 27,
       36, 44,  6, 55, 20,
        3, 10, 43, 25, 39,
       41, 45, 15, 21,  8,
       18,  2, 61, 56, 14
    };
    static const uint32_t pi[25] = {
        0,  6, 12, 18, 24,
        3,  9, 10, 16, 22,
        1,  7, 13, 19, 20,
        4,  5, 11, 17, 23,
        2,  8, 14, 15, 21
    };
    for (int round = 0; round < 24; ++round) {
        uint64_t C[5];
        for (int i = 0; i < 5; ++i)
            C[i] = state[i] ^ state[i + 5] ^ state[i + 10] ^ state[i + 15] ^ state[i + 20];
        uint64_t D[5];
        for (int i = 0; i < 5; ++i)
            D[i] = C[(i + 4) % 5] ^ keccak_rotl(C[(i + 1) % 5], 1);
        for (int i = 0; i < 25; i += 5)
            for (int j = 0; j < 5; ++j)
                state[i + j] ^= D[j];

        uint64_t B[25];
        for (int i = 0; i < 25; ++i)
            B[pi[i]] = keccak_rotl(state[i], rho[i]);

        for (int i = 0; i < 25; i += 5)
            for (int j = 0; j < 5; ++j)
                state[i + j] = B[i + j] ^ ((~B[i + ((j + 1) % 5)]) & B[i + ((j + 2) % 5)]);

        state[0] ^= keccak_round_constants[round];
    }
}

void keccak256(const uint8_t* data, size_t len, uint8_t out[32]) {
    uint64_t state[25];
    memset(state, 0, sizeof(state));
    size_t rate = 136;
    size_t offset = 0;

    while (len >= rate) {
        for (size_t i = 0; i < rate / 8; ++i) {
            uint64_t t = 0;
            for (size_t j = 0; j < 8; ++j)
                t |= ((uint64_t)data[offset + i * 8 + j]) << (8 * j);
            state[i] ^= t;
        }
        keccakf(state);
        offset += rate;
        len -= rate;
    }

    uint8_t block[136];
    memset(block, 0, sizeof(block));
    memcpy(block, data + offset, len);
    block[len] = 0x01;
    block[rate - 1] |= 0x80;

    for (size_t i = 0; i < rate / 8; ++i) {
        uint64_t t = 0;
        for (size_t j = 0; j < 8; ++j)
            t |= ((uint64_t)block[i * 8 + j]) << (8 * j);
        state[i] ^= t;
    }
    keccakf(state);

    for (int i = 0; i < 4; ++i) {
        uint64_t t = state[i];
        for (int j = 0; j < 8; ++j)
            out[i * 8 + j] = (uint8_t)((t >> (8 * j)) & 0xff);
    }
}

void merkle_root(const uint8_t* hashes, size_t count, uint8_t out[32]) {
    if (count == 0) {
        memset(out, 0, 32);
        return;
    }
    uint8_t* buffer = (uint8_t*)malloc(count * 32);
    memcpy(buffer, hashes, count * 32);
    size_t current = count;
    while (current > 1) {
        size_t next = 0;
        uint8_t* next_buf = (uint8_t*)malloc(((current + 1) / 2) * 32);
        for (size_t i = 0; i < current; i += 2) {
            uint8_t hash[32];
            if (i + 1 < current) {
                sha256_double(buffer + i * 32, 64, hash);
            } else {
                memcpy(hash, buffer + i * 32, 32);
            }
            memcpy(next_buf + next * 32, hash, 32);
            next++;
        }
        free(buffer);
        buffer = next_buf;
        current = next;
    }
    memcpy(out, buffer, 32);
    free(buffer);
}

void blockheader_hash(const cortex_block_header* header, uint8_t out[32]) {
    uint8_t buffer[32 + 32 + 8 + 4 + 4];
    memcpy(buffer, header->previous_hash, 32);
    memcpy(buffer + 32, header->merkle_root, 32);
    memcpy(buffer + 64, &header->timestamp, 8);
    memcpy(buffer + 72, &header->difficulty, 4);
    memcpy(buffer + 76, &header->nonce, 4);
    sha256_double(buffer, sizeof(buffer), out);
}

#endif // CORTEX_FEATURE_BLOCKCHAIN
