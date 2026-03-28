#ifndef CORTEX_RUNTIME_CORE_H
#define CORTEX_RUNTIME_CORE_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#ifndef CORTEX_FEATURE_ASYNC
#define CORTEX_FEATURE_ASYNC 1
#endif
#ifndef CORTEX_FEATURE_ACTORS
#define CORTEX_FEATURE_ACTORS 1
#endif
#ifndef CORTEX_FEATURE_BLOCKCHAIN
#define CORTEX_FEATURE_BLOCKCHAIN 1
#endif
#ifndef CORTEX_FEATURE_QOL
#define CORTEX_FEATURE_QOL 1
#endif

#ifdef __cplusplus
extern "C" {
#endif

typedef enum {
    TYPE_INT,
    TYPE_FLOAT,
    TYPE_STRING,
    TYPE_BOOL,
    TYPE_VEC2,
    TYPE_VEC3,
    TYPE_NULL,
    TYPE_ANY,
    TYPE_DICT,
    TYPE_ARRAY
} DataType;

typedef struct {
    float x;
    float y;
} vec2;

typedef struct {
    float x;
    float y;
    float z;
} vec3;

typedef struct {
    float x;
    float y;
    float z;
    float w;
} vec4;

/* Slice views (Cortex slice<T>): borrowed range; ptr is not owned by the slice */
typedef struct { int* ptr; int len; } cortex_slice_int;
typedef struct { float* ptr; int len; } cortex_slice_float;
typedef struct { double* ptr; int len; } cortex_slice_double;

/* Optional type support: T? -> cortex_optional_T with has_value flag */
typedef struct { bool has_value; int value; } cortex_optional_int;
typedef struct { bool has_value; float value; } cortex_optional_float;
typedef struct { bool has_value; double value; } cortex_optional_double;
typedef struct { bool has_value; char value; } cortex_optional_char;
typedef struct { bool has_value; bool value; } cortex_optional_bool;
typedef struct { bool has_value; char* value; } cortex_optional_string;
typedef struct { bool has_value; vec2 value; } cortex_optional_vec2;
typedef struct { bool has_value; vec3 value; } cortex_optional_vec3;
typedef struct { bool has_value; void* value; } cortex_optional_ptr;

/* Optional constructors */
#define optional_none_int() ((cortex_optional_int){.has_value = false})
#define optional_some_int(v) ((cortex_optional_int){.has_value = true, .value = (v)})
#define optional_none_float() ((cortex_optional_float){.has_value = false})
#define optional_some_float(v) ((cortex_optional_float){.has_value = true, .value = (v)})
#define optional_none_string() ((cortex_optional_string){.has_value = false})
#define optional_some_string(v) ((cortex_optional_string){.has_value = true, .value = (v)})

/* Vector<T> generic type support: typed dynamic arrays */
typedef struct { int* data; int size; int capacity; } cortex_vector_int;
typedef struct { float* data; int size; int capacity; } cortex_vector_float;
typedef struct { double* data; int size; int capacity; } cortex_vector_double;
typedef struct { char** data; int size; int capacity; } cortex_vector_string;
typedef struct { bool* data; int size; int capacity; } cortex_vector_bool;
typedef struct { vec2* data; int size; int capacity; } cortex_vector_vec2;
typedef struct { vec3* data; int size; int capacity; } cortex_vector_vec3;
typedef struct { void** data; int size; int capacity; } cortex_vector_ptr;

/* Vector constructors and methods */
#define vector_create_int() ((cortex_vector_int){.data = NULL, .size = 0, .capacity = 0})
#define vector_create_float() ((cortex_vector_float){.data = NULL, .size = 0, .capacity = 0})
#define vector_create_string() ((cortex_vector_string){.data = NULL, .size = 0, .capacity = 0})

/* Vector operations - implemented in vector.c */
void vector_push_int(cortex_vector_int* v, int val);
int vector_pop_int(cortex_vector_int* v);
int vector_get_int(cortex_vector_int* v, int idx);
void vector_set_int(cortex_vector_int* v, int idx, int val);
int vector_len_int(cortex_vector_int* v);
void vector_free_int(cortex_vector_int* v);

void vector_push_float(cortex_vector_float* v, float val);
float vector_pop_float(cortex_vector_float* v);
float vector_get_float(cortex_vector_float* v, int idx);
void vector_set_float(cortex_vector_float* v, int idx, float val);
int vector_len_float(cortex_vector_float* v);
void vector_free_float(cortex_vector_float* v);

void vector_push_string(cortex_vector_string* v, char* val);
char* vector_pop_string(cortex_vector_string* v);
char* vector_get_string(cortex_vector_string* v, int idx);
void vector_set_string(cortex_vector_string* v, int idx, char* val);
int vector_len_string(cortex_vector_string* v);
void vector_free_string(cortex_vector_string* v);

/* Forward decl for AnyValue union (dict/array stored as void* to avoid circular deps) */
struct cortex_dict;
struct cortex_array;

typedef struct {
    DataType type;
    union {
        int int_val;
        float float_val;
        char* string_val;
        bool bool_val;
        vec2 vec2_val;
        vec3 vec3_val;
        void* dict_val;   /* cortex_dict* when type == TYPE_DICT */
        void* array_val;  /* cortex_array* when type == TYPE_ARRAY */
    } data;
} AnyValue;

AnyValue make_any_int(int val);
AnyValue make_any_float(float val);
AnyValue make_any_string(char* val);
AnyValue make_any_bool(bool val);
AnyValue make_any_vec2(vec2 val);
AnyValue make_any_vec3(vec3 val);
AnyValue make_any_null(void);

char* type_of(AnyValue val);
bool is_type(AnyValue val, char* type_name);
int as_int(AnyValue val);
float as_float(AnyValue val);
char* as_string(AnyValue val);
bool as_bool(AnyValue val);

/* C has no overloading; use type-suffixed names for codegen */
char* toString_int(int val);
char* toString_float(float val);
char* toString_double(double val);
char* toString_bool(bool val);

vec2 make_vec2(float x, float y);
vec3 make_vec3(float x, float y, float z);
vec4 make_vec4(float x, float y, float z, float w);
float dot(vec2 a, vec2 b);
vec2 normalize(vec2 v);

/* Game math: vector ops */
float vec2_dot(vec2 a, vec2 b);
float vec2_length(vec2 v);
float vec2_length_sq(vec2 v);
float vec2_distance(vec2 a, vec2 b);
vec2 vec2_add(vec2 a, vec2 b);
vec2 vec2_sub(vec2 a, vec2 b);
vec2 vec2_scale(vec2 v, float s);
vec2 vec2_normalize(vec2 v);
vec2 vec2_lerp(vec2 a, vec2 b, float t);
float vec3_length(vec3 v);
float vec3_length_sq(vec3 v);
float vec3_dot(vec3 a, vec3 b);
vec3 vec3_normalize(vec3 v);
vec3 vec3_add(vec3 a, vec3 b);
vec3 vec3_sub(vec3 a, vec3 b);
vec3 vec3_scale(vec3 v, float s);
float vec3_distance(vec3 a, vec3 b);
vec3 vec3_lerp(vec3 a, vec3 b, float t);
vec4 vec4_lerp(vec4 a, vec4 b, float t);

/* Game math: scalars */
float clamp_float(float x, float lo, float hi);
float lerp_float(float a, float b, float t);
float min_float(float a, float b);
float max_float(float a, float b);

int random_int(int min, int max);
float random_float(float min, float max);
float get_time(void);
void sleep_func(float seconds);
void random_bytes(uint8_t* buffer, size_t len);
uint64_t unix_time_ms(void);

void print_string(const char* str);
void println_string(const char* str);
/* Print one AnyValue (converts to string); for print(a, b, c) from Cortex */
void print_any(AnyValue v);
/* Print one AnyValue and newline; for writeline(x) from Cortex */
void println_any(AnyValue v);
/* C-style writeline: printf format + newline. Use writeline(fmt, ...) in Cortex. */
int writeline_fmt(const char* fmt, ...);
/* Allocates a new string; caller should free. Used for Cortex string +. */
char* cortex_strcat(const char* a, const char* b);

/* Application I/O: reads until newline; caller must free result. Returns NULL on EOF. */
char* input_line(void);
/* Application I/O: wait for any key press (Windows only, cross-platform stub) */
void waitkey(void);
/* Application I/O: read entire file; caller must free. Returns NULL on error. */
char* read_file(const char* path);
/* Application I/O: write string to file. Returns 0 on success. */
int write_file(const char* path, const char* content);
/* Application I/O: append string to file. Returns 0 on success. */
int append_file(const char* path, const char* content);
/* Application I/O: read file as array of lines (JSON array string). Caller frees. */
char* read_file_lines(const char* path);
/* File utilities: get file size, copy, move, delete, modified time */
long file_size(const char* path);
bool file_copy(const char* src, const char* dst);
bool file_move(const char* src, const char* dst);
bool file_delete(const char* path);
time_t file_modified(const char* path);
bool file_create_dir(const char* path);

/* Array bounds checking: returns index if in range, else aborts. Use as arr[cortex_bounds_check(len, index, line)]. */
int cortex_bounds_check(int len, int index, int line);

/* Assert failure: prints message and line, then aborts. Called by assert() codegen. */
void cortex_assert_fail(int line, const char* msg);

/* Game/app math: sign, wrap, round, floor, ceil (QOL) */
float sign_float(float x);
float wrap_float(float x, float lo, float hi);
float round_float(float x);
float floor_float(float x);
float ceil_float(float x);

#if CORTEX_FEATURE_QOL
/* Dynamic array: holds AnyValue elements, grows as needed. Handle = cortex_array* */
typedef struct cortex_array cortex_array;

/* Application I/O: write array of strings as lines. Returns 0 on success. */
int write_file_lines(const char* path, cortex_array* lines);
cortex_array* array_create(void);
void array_push(cortex_array* a, AnyValue val);
AnyValue array_get(cortex_array* a, int index);
void array_set(cortex_array* a, int index, AnyValue val);
int array_len(cortex_array* a);
int array_capacity(cortex_array* a);
void array_reserve(cortex_array* a, int min_cap);
AnyValue array_pop(cortex_array* a);
void array_insert(cortex_array* a, int index, AnyValue val);
void array_remove_at(cortex_array* a, int index);
void array_free(cortex_array* a);

/* Events: subscribe callbacks, emit value. Handle = cortex_event* */
typedef void (*cortex_event_callback)(AnyValue val);
typedef struct cortex_event cortex_event;
cortex_event* event_create(void);
void event_subscribe(cortex_event* e, cortex_event_callback cb);
void event_unsubscribe(cortex_event* e, cortex_event_callback cb);
void event_emit(cortex_event* e, AnyValue val);
void event_free(cortex_event* e);

/* Dictionary: string keys, AnyValue values. Handle = cortex_dict* */
typedef struct cortex_dict cortex_dict;
typedef struct cortex_array cortex_array;
cortex_dict* dict_create(void);
void dict_set(cortex_dict* d, const char* key, AnyValue val);
AnyValue dict_get(cortex_dict* d, const char* key);
bool dict_has(cortex_dict* d, const char* key);
int dict_len(cortex_dict* d);
void dict_free(cortex_dict* d);

/* AnyValue boxing for dict/array (nested JSON, closures, etc.) */
AnyValue make_any_dict(cortex_dict* d);
AnyValue make_any_array(cortex_array* a);
cortex_dict* as_dict(AnyValue val);
cortex_array* as_array(AnyValue val);

/* Result<T,E>-like: ok flag + value or error message */
typedef struct { bool ok; AnyValue value; char* error; } cortex_result;
cortex_result result_ok(AnyValue v);
cortex_result result_err(const char* msg);
bool result_is_ok(cortex_result r);
AnyValue result_value(cortex_result r);
char* result_error(cortex_result r);

/* String utilities (caller frees returned strings where noted) */
cortex_array* cortex_str_split(const char* s, const char* delim);
char* cortex_str_join_array(cortex_array* parts, const char* sep);
char* cortex_str_replace(const char* s, const char* from, const char* to);
char* cortex_str_trim(const char* s);
bool cortex_str_starts_with(const char* s, const char* prefix);
bool cortex_str_ends_with(const char* s, const char* suffix);
char* cortex_str_to_lower(const char* s);
char* cortex_str_to_upper(const char* s);

/* Enhanced string utilities */
char* string_format(const char* fmt, ...);
char* string_pad_left(const char* s, int width, char pad);
char* string_pad_right(const char* s, int width, char pad);
char* string_center(const char* s, int width, char pad);
char* string_reverse(const char* s);
int string_index_of(const char* s, const char* sub);
int string_last_index_of(const char* s, const char* sub);
bool string_contains(const char* s, const char* sub);

/* Math */
int clamp_int(int x, int lo, int hi);
double cortex_pow(double base, double exp);
AnyValue array_random_choice(cortex_array* a);

/* Enhanced math utilities */
int abs_int(int x);
float abs_float(float x);
double abs_double(double x);
int min_int(int a, int b);
int max_int(int a, int b);
double min_double(double a, double b);
double max_double(double a, double b);
float min_float(float a, float b);
float max_float(float a, float b);
double cortex_sqrt(double x);
double cortex_sin(double x);
double cortex_cos(double x);
double cortex_tan(double x);
double cortex_floor(double x);
double cortex_ceil(double x);
double cortex_round(double x);

/* System utilities */
char* get_env(const char* name);
bool set_env(const char* name, const char* value);
char* get_cwd(void);
bool change_dir(const char* path);
int system_run(const char* command);
char* get_username(void);

/* Memory utilities */
void* mem_alloc(size_t size);
void* mem_realloc(void* ptr, size_t size);
void mem_free(void* ptr);
void* mem_copy(void* dest, const void* src, size_t size);
void* mem_move(void* dest, const void* src, size_t size);
void* mem_set(void* ptr, int value, size_t size);
int mem_compare(const void* a, const void* b, size_t size);

/* File/path */
bool file_exists(const char* path);
char* list_dir(const char* path);
char* path_join(const char* a, const char* b);

/* Debug */
void debug_log(const char* fmt, ...);
void debug_assert(int condition, const char* msg, int line);
void dump_any(AnyValue v);

/* Unit test */
void test_register(const char* name, void (*fn)(void));
int test_run_all(void);
void assert_eq_int(int a, int b, const char* file, int line);
void assert_eq_float(float a, float b, float epsilon, const char* file, int line);
void assert_eq_str(const char* a, const char* b, const char* file, int line);

/* JSON (minimal: parse to dict, stringify from any) */
cortex_dict* json_parse(const char* s);
char* json_stringify_any(AnyValue v);
char* json_stringify_dict(cortex_dict* d);

/* Parse string to number (0 or 0.0 on failure; use result pattern in Cortex for error handling) */
float parse_number(const char* s);
int parse_int(const char* s);

/* ECS helpers: entity = id (int); components stored per-entity as dict (name -> AnyValue) */
int entity_create(void);
void add_component(int entity_id, const char* component_name, AnyValue val);
AnyValue get_component(int entity_id, const char* component_name);
bool has_component(int entity_id, const char* component_name);
void entity_remove(int entity_id);
#endif

#if CORTEX_FEATURE_BLOCKCHAIN
void sha256_hash(const uint8_t* data, size_t len, uint8_t out[32]);
void sha256_double(const uint8_t* data, size_t len, uint8_t out[32]);
void keccak256(const uint8_t* data, size_t len, uint8_t out[32]);
void merkle_root(const uint8_t* hashes, size_t count, uint8_t out[32]);
void hex_encode(const uint8_t* data, size_t len, char* out);
size_t hex_decode(const char* hex, uint8_t* out, size_t max_len);
int base58_encode(const uint8_t* data, size_t len, char* out, size_t out_len);
int base58_decode(const char* input, uint8_t* out, size_t out_len);

typedef struct {
    uint8_t previous_hash[32];
    uint8_t merkle_root[32];
    uint64_t timestamp;
    uint32_t difficulty;
    uint32_t nonce;
} cortex_block_header;

void blockheader_hash(const cortex_block_header* header, uint8_t out[32]);
#endif

#ifdef __cplusplus
}
#endif

#endif // CORTEX_RUNTIME_CORE_H
