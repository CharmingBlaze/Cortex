#ifndef CORTEX_RUNTIME_GRAPHICS_H
#define CORTEX_RUNTIME_GRAPHICS_H

#include "core.h"
#include <stdbool.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// ============================================================================
// Color Types
// ============================================================================

typedef struct { uint8_t r, g, b, a; } color;
typedef struct { float r, g, b, a; } colorf;

// Rectangle type for bounds
typedef struct { float x, y, w, h; } cortex_rect;

// Predefined colors
#define COLOR_CLEAR      ((color){0, 0, 0, 0})
#define COLOR_BLACK      ((color){0, 0, 0, 255})
#define COLOR_WHITE      ((color){255, 255, 255, 255})
#define COLOR_RED        ((color){255, 0, 0, 255})
#define COLOR_GREEN      ((color){0, 255, 0, 255})
#define COLOR_BLUE       ((color){0, 0, 255, 255})
#define COLOR_YELLOW     ((color){255, 255, 0, 255})
#define COLOR_CYAN       ((color){0, 255, 255, 255})
#define COLOR_MAGENTA    ((color){255, 0, 255, 255})
#define COLOR_ORANGE     ((color){255, 165, 0, 255})
#define COLOR_PURPLE     ((color){128, 0, 128, 255})
#define COLOR_PINK       ((color){255, 192, 203, 255})
#define COLOR_GRAY       ((color){128, 128, 128, 255})
#define COLOR_LIGHTGRAY  ((color){192, 192, 192, 255})
#define COLOR_DARKGRAY   ((color){64, 64, 64, 255})
#define COLOR_BROWN      ((color){139, 69, 19, 255})
#define COLOR_NAVY       ((color){0, 0, 128, 255})
#define COLOR_TEAL       ((color){0, 128, 128, 255})
#define COLOR_MAROON     ((color){128, 0, 0, 255})
#define COLOR_OLIVE      ((color){128, 128, 0, 255})
#define COLOR_LIME       ((color){0, 255, 128, 255})
#define COLOR_AQUA       ((color){0, 255, 255, 255})
#define COLOR_FUCHSIA    ((color){255, 0, 255, 255})
#define COLOR_SILVER     ((color){192, 192, 192, 255})

// Color creation and manipulation
color color_make(uint8_t r, uint8_t g, uint8_t b, uint8_t a);
color color_from_hex(uint32_t hex);
color color_from_hsv(float h, float s, float v);
color color_from_hsva(float h, float s, float v, float a);
colorf colorf_make(float r, float g, float b, float a);
color colorf_to_color(colorf c);
colorf color_to_colorf(color c);

// Color operations
color color_lerp(color a, color b, float t);
color color_blend(color src, color dst);
color color_multiply(color a, color b);
color color_tint(color c, color tint);
color color_fade(color c, float alpha);
color color_brightness(color c, float factor);
color color_contrast(color c, float factor);
color color_saturation(color c, float factor);
color color_invert(color c);
color color_grayscale(color c);
color color_sepia(color c);

// Color queries
uint32_t color_to_hex(color c);
uint32_t color_to_rgba(color c);
void color_to_hsv(color c, float* h, float* s, float* v);
bool color_equals(color a, color b);
bool color_transparent(color c);

// ============================================================================
// Image Type
// ============================================================================

typedef struct {
    int width;
    int height;
    int channels;       // 1=gray, 2=gray+alpha, 3=rgb, 4=rgba
    uint8_t* data;      // Pixel data, row-major, top-down
    bool owns_data;     // If true, free data on image_free
} image;

// Image creation
image* image_create(int width, int height, int channels);
image* image_from_data(int width, int height, int channels, uint8_t* data, bool copy);
image* image_from_file(const char* filepath);
image* image_from_memory(const uint8_t* data, size_t size);
image* image_clone(const image* img);
image* image_subimage(const image* img, int x, int y, int w, int h);

// Image destruction
void image_free(image* img);

// Image properties
int image_width(const image* img);
int image_height(const image* img);
int image_channels(const image* img);
size_t image_size(const image* img);
bool image_valid(const image* img);
vec2 image_center(const image* img);
cortex_rect image_bounds(const image* img);

// Pixel access
color image_get_pixel(const image* img, int x, int y);
void image_set_pixel(image* img, int x, int y, color c);
uint8_t image_get_channel(const image* img, int x, int y, int channel);
void image_set_channel(image* img, int x, int y, int channel, uint8_t value);
color* image_get_pixel_ptr(image* img, int x, int y);
const color* image_get_pixel_ptr_const(const image* img, int x, int y);

// Image operations
void image_clear(image* img, color c);
void image_fill_rect(image* img, int x, int y, int w, int h, color c);
void image_copy(image* dst, int dx, int dy, const image* src, int sx, int sy, int sw, int sh);
void image_paste(image* dst, int dx, int dy, const image* src);
void image_blend(image* dst, int dx, int dy, const image* src, float alpha);

// Image transformations
image* image_resize(const image* img, int new_width, int new_height);
image* image_resize_nearest(const image* img, int new_width, int new_height);
image* image_resize_bilinear(const image* img, int new_width, int new_height);
image* image_scale(const image* img, float scale_x, float scale_y);
image* image_rotate_90(const image* img, int times);
image* image_rotate_180(const image* img);
image* image_flip_h(image* img);
image* image_flip_v(image* img);
image* image_flip_both(image* img);
image* image_crop(const image* img, int x, int y, int w, int h);
image* image_trim(const image* img, color bg);

// Image filters
void image_grayscale(image* img);
void image_sepia(image* img);
void image_invert(image* img);
void image_brightness(image* img, float factor);
void image_contrast(image* img, float factor);
void image_saturation(image* img, float factor);
void image_gamma(image* img, float gamma);
void image_tint(image* img, color tint);
void image_fade(image* img, float alpha);
void image_blur_box(image* img, int radius);
void image_blur_gaussian(image* img, int radius);
void image_sharpen(image* img, float amount);
void image_emboss(image* img);
void image_edge_detect(image* img);
void image_threshold(image* img, uint8_t threshold);
void image_posterize(image* img, int levels);
void image_pixelate(image* img, int block_size);
void image_dither(image* img, int levels);
void image_convolve(image* img, const float* kernel, int kernel_size);

// Image effects
void image_noise(image* img, float amount, bool monochrome);
void image_vignette(image* img, float intensity, float radius);
void image_vignette_oval(image* img, float intensity);
void image_scanlines(image* img, float intensity, int spacing);
void image_glitch(image* img, float amount);
void image_chromatic_aberration(image* img, float amount);

// Image saving
bool image_save_png(const image* img, const char* filepath);
bool image_save_bmp(const image* img, const char* filepath);
bool image_save_tga(const image* img, const char* filepath);
bool image_save_jpg(const image* img, const char* filepath, int quality);

// Image utilities
image* image_load(const char* filepath);
bool image_save(const image* img, const char* filepath);
image* image_convert(const image* img, int channels);
image* image_premultiply_alpha(image* img);
image* image_unpremultiply_alpha(image* img);
bool image_has_alpha(const image* img);
color image_average_color(const image* img);
color image_dominant_color(const image* img);
void image_histogram(const image* img, int hist_r[256], int hist_g[256], int hist_b[256]);

// ============================================================================
// Sprite Type
// ============================================================================

typedef struct {
    int x, y;           // Position in sprite sheet
    int width, height;  // Size of sprite
    float pivot_x;      // Pivot point (0-1 normalized)
    float pivot_y;
} sprite_frame;

typedef struct {
    int start_frame;
    int frame_count;
    float duration;     // Total duration in seconds
    bool looping;
} sprite_animation;

typedef struct {
    image* sheet;                   // Source image
    sprite_frame* frames;            // Array of frames
    int frame_count;                 // Number of frames
    sprite_animation* animations;    // Array of animations
    int animation_count;             // Number of animations
} sprite_sheet;

typedef struct {
    sprite_sheet* sheet;
    int current_frame;
    int current_animation;
    float frame_time;
    float speed;
    bool playing;
    bool flipped_h;
    bool flipped_v;
    float scale_x;
    float scale_y;
    float rotation;
    float alpha;
    color tint;
} sprite_instance;

// Sprite sheet creation
sprite_sheet* sprite_sheet_create(image* img);
sprite_sheet* sprite_sheet_load(const char* filepath);
sprite_sheet* sprite_sheet_from_grid(image* img, int frame_width, int frame_height, int columns, int rows);
void sprite_sheet_free(sprite_sheet* sheet);

// Sprite frame management
int sprite_sheet_add_frame(sprite_sheet* sheet, int x, int y, int w, int h);
int sprite_sheet_add_frames_grid(sprite_sheet* sheet, int start_x, int start_y, int frame_w, int frame_h, int columns, int rows);
void sprite_sheet_set_pivot(sprite_sheet* sheet, int frame, float pivot_x, float pivot_y);
sprite_frame* sprite_sheet_get_frame(sprite_sheet* sheet, int frame);

// Animation management
int sprite_sheet_add_animation(sprite_sheet* sheet, int start_frame, int frame_count, float duration, bool looping);
int sprite_sheet_add_animation_frames(sprite_sheet* sheet, const int* frames, int frame_count, float duration, bool looping);
sprite_animation* sprite_sheet_get_animation(sprite_sheet* sheet, int animation);

// Sprite instance
sprite_instance* sprite_instance_create(sprite_sheet* sheet);
void sprite_instance_free(sprite_instance* sprite);
void sprite_instance_play(sprite_instance* sprite, int animation);
void sprite_instance_stop(sprite_instance* sprite);
void sprite_instance_reset(sprite_instance* sprite);
void sprite_instance_update(sprite_instance* sprite, float dt);
void sprite_instance_set_frame(sprite_instance* sprite, int frame);
void sprite_instance_next_frame(sprite_instance* sprite);
void sprite_instance_prev_frame(sprite_instance* sprite);

// Sprite instance properties
void sprite_instance_set_position(sprite_instance* sprite, float x, float y);
void sprite_instance_set_scale(sprite_instance* sprite, float scale);
void sprite_instance_set_scale_xy(sprite_instance* sprite, float scale_x, float scale_y);
void sprite_instance_set_rotation(sprite_instance* sprite, float radians);
void sprite_instance_set_alpha(sprite_instance* sprite, float alpha);
void sprite_instance_set_tint(sprite_instance* sprite, color tint);
void sprite_instance_set_flip(sprite_instance* sprite, bool h, bool v);

// Sprite rendering (draws to target image)
void sprite_draw(const sprite_instance* sprite, image* target, float x, float y);
void sprite_draw_scaled(const sprite_instance* sprite, image* target, float x, float y, float scale);
void sprite_draw_rotated(const sprite_instance* sprite, image* target, float x, float y, float rotation);
void sprite_draw_ext(const sprite_instance* sprite, image* target, float x, float y, float scale_x, float scale_y, float rotation, float alpha);

// Sprite bounds
cortex_rect sprite_get_bounds(const sprite_instance* sprite);
vec2 sprite_get_size(const sprite_instance* sprite);
vec2 sprite_get_pivot(const sprite_instance* sprite);

// ============================================================================
// Vector Graphics - Canvas
// ============================================================================

typedef struct {
    image* target;
    int clip_x, clip_y, clip_w, clip_h;
    bool clip_enabled;
    color fill_color;
    color stroke_color;
    float stroke_width;
    float transform[6];     // 2x3 affine transform matrix
} canvas;

// Canvas creation
canvas* canvas_create(int width, int height);
canvas* canvas_from_image(image* img);
void canvas_free(canvas* cv);

// Canvas state
void canvas_set_fill(canvas* cv, color c);
void canvas_set_stroke(canvas* cv, color c);
void canvas_set_stroke_width(canvas* cv, float width);
void canvas_set_clip(canvas* cv, int x, int y, int w, int h);
void canvas_clear_clip(canvas* cv);
void canvas_clear(canvas* cv, color c);

// Transformations
void canvas_reset_transform(canvas* cv);
void canvas_translate(canvas* cv, float tx, float ty);
void canvas_scale(canvas* cv, float sx, float sy);
void canvas_rotate(canvas* cv, float radians);
void canvas_transform(canvas* cv, float a, float b, float c, float d, float e, float f);
void canvas_set_transform(canvas* cv, float a, float b, float c, float d, float e, float f);
void canvas_push_transform(canvas* cv);
void canvas_pop_transform(canvas* cv);

// Basic shapes
void canvas_draw_point(canvas* cv, float x, float y);
void canvas_draw_line(canvas* cv, float x1, float y1, float x2, float y2);
void canvas_draw_line_width(canvas* cv, float x1, float y1, float x2, float y2, float width);
void canvas_draw_rect(canvas* cv, float x, float y, float w, float h);
void canvas_fill_rect(canvas* cv, float x, float y, float w, float h);
void canvas_draw_rect_rounded(canvas* cv, float x, float y, float w, float h, float radius);
void canvas_fill_rect_rounded(canvas* cv, float x, float y, float w, float h, float radius);

// Circle and ellipse
void canvas_draw_circle(canvas* cv, float cx, float cy, float radius);
void canvas_fill_circle(canvas* cv, float cx, float cy, float radius);
void canvas_draw_ellipse(canvas* cv, float cx, float cy, float rx, float ry);
void canvas_fill_ellipse(canvas* cv, float cx, float cy, float rx, float ry);
void canvas_draw_arc(canvas* cv, float cx, float cy, float radius, float start_angle, float end_angle);
void canvas_fill_arc(canvas* cv, float cx, float cy, float radius, float start_angle, float end_angle);
void canvas_draw_pie(canvas* cv, float cx, float cy, float radius, float start_angle, float end_angle);
void canvas_fill_pie(canvas* cv, float cx, float cy, float radius, float start_angle, float end_angle);

// Polygon
void canvas_draw_triangle(canvas* cv, float x1, float y1, float x2, float y2, float x3, float y3);
void canvas_fill_triangle(canvas* cv, float x1, float y1, float x2, float y2, float x3, float y3);
void canvas_draw_polygon(canvas* cv, const float* points, int count);
void canvas_fill_polygon(canvas* cv, const float* points, int count);
void canvas_draw_polyline(canvas* cv, const float* points, int count);

// Paths
typedef struct canvas_path canvas_path;

canvas_path* canvas_path_create(void);
void canvas_path_free(canvas_path* path);
void canvas_path_move_to(canvas_path* path, float x, float y);
void canvas_path_line_to(canvas_path* path, float x, float y);
void canvas_path_arc_to(canvas_path* path, float x1, float y1, float x2, float y2, float radius);
void canvas_path_arc(canvas_path* path, float cx, float cy, float radius, float start, float end, bool ccw);
void canvas_path_curve_to(canvas_path* path, float cp1x, float cp1y, float cp2x, float cp2y, float x, float y);
void canvas_path_quad_to(canvas_path* path, float cpx, float cpy, float x, float y);
void canvas_path_rect(canvas_path* path, float x, float y, float w, float h);
void canvas_path_circle(canvas_path* path, float cx, float cy, float radius);
void canvas_path_ellipse(canvas_path* path, float cx, float cy, float rx, float ry);
void canvas_path_close(canvas_path* path);
void canvas_path_reset(canvas_path* path);

void canvas_draw_path(canvas* cv, const canvas_path* path);
void canvas_fill_path(canvas* cv, const canvas_path* path);

// Text (basic bitmap font support)
void canvas_draw_text(canvas* cv, float x, float y, const char* text);
void canvas_draw_text_color(canvas* cv, float x, float y, const char* text, color c);
vec2 canvas_measure_text(canvas* cv, const char* text);

// Image drawing
void canvas_draw_image(canvas* cv, const image* img, float x, float y);
void canvas_draw_image_scaled(canvas* cv, const image* img, float x, float y, float scale);
void canvas_draw_image_rect(canvas* cv, const image* img, float dx, float dy, float dw, float dh, int sx, int sy, int sw, int sh);
void canvas_draw_image_tiled(canvas* cv, const image* img, float x, float y, float w, float h);
void canvas_draw_image_9patch(canvas* cv, const image* img, float x, float y, float w, float h, int border);

// Gradients
typedef struct {
    float x1, y1, x2, y2;
    color start_color;
    color end_color;
    bool radial;
} gradient;

gradient* gradient_create_linear(float x1, float y1, float x2, float y2, color start, color end);
gradient* gradient_create_radial(float cx, float cy, float inner_r, float outer_r, color start, color end);
void gradient_free(gradient* g);
void gradient_add_stop(gradient* g, float position, color c);

void canvas_fill_rect_gradient(canvas* cv, float x, float y, float w, float h, const gradient* g);
void canvas_fill_circle_gradient(canvas* cv, float cx, float cy, float radius, const gradient* g);
void canvas_fill_path_gradient(canvas* cv, const canvas_path* path, const gradient* g);

// ============================================================================
// Vector Graphics - Primitives (standalone functions)
// ============================================================================

// Line drawing algorithms
void draw_line(image* img, int x1, int y1, int x2, int y2, color c);
void draw_line_aa(image* img, float x1, float y1, float x2, float y2, color c);
void draw_line_thick(image* img, float x1, float y1, float x2, float y2, float width, color c);
void draw_line_dashed(image* img, float x1, float y1, float x2, float y2, float dash_len, float gap_len, color c);

// Shape drawing
void draw_rect(image* img, int x, int y, int w, int h, color c);
void fill_rect(image* img, int x, int y, int w, int h, color c);
void draw_rect_rounded(image* img, int x, int y, int w, int h, int radius, color c);
void fill_rect_rounded(image* img, int x, int y, int w, int h, int radius, color c);

void draw_circle(image* img, int cx, int cy, int radius, color c);
void fill_circle(image* img, int cx, int cy, int radius, color c);
void draw_circle_aa(image* img, float cx, float cy, float radius, color c);
void fill_circle_aa(image* img, float cx, float cy, float radius, color c);

void draw_ellipse(image* img, int cx, int cy, int rx, int ry, color c);
void fill_ellipse(image* img, int cx, int cy, int rx, int ry, color c);

void draw_triangle(image* img, int x1, int y1, int x2, int y2, int x3, int y3, color c);
void fill_triangle(image* img, int x1, int y1, int x2, int y2, int x3, int y3, color c);

void draw_polygon(image* img, const int* x, const int* y, int count, color c);
void fill_polygon(image* img, const int* x, const int* y, int count, color c);

// Arc and curve drawing
void draw_arc(image* img, int cx, int cy, int radius, float start_angle, float end_angle, color c);
void fill_arc(image* img, int cx, int cy, int radius, float start_angle, float end_angle, color c);
void draw_bezier(image* img, int x1, int y1, int x2, int y2, int x3, int y3, int x4, int y4, color c);
void draw_bezier_quad(image* img, int x1, int y1, int x2, int y2, int x3, int y3, color c);
void draw_spline(image* img, const int* x, const int* y, int count, color c);

// ============================================================================
// Text Rendering (Bitmap Font)
// ============================================================================

typedef struct {
    image* atlas;
    int char_width;
    int char_height;
    int columns;
    int first_char;
    int last_char;
    int spacing;
} bitmap_font;

bitmap_font* bitmap_font_create(image* atlas, int char_w, int char_h, int columns, int first_char, int last_char);
bitmap_font* bitmap_font_load(const char* filepath, int char_w, int char_h, int columns, int first_char, int last_char);
void bitmap_font_free(bitmap_font* font);
void bitmap_font_set_spacing(bitmap_font* font, int spacing);

void bitmap_font_draw(const bitmap_font* font, image* target, int x, int y, const char* text);
void bitmap_font_draw_color(const bitmap_font* font, image* target, int x, int y, const char* text, color c);
void bitmap_font_draw_scaled(const bitmap_font* font, image* target, float x, float y, const char* text, float scale);
int bitmap_font_measure_width(const bitmap_font* font, const char* text);
int bitmap_font_measure_height(const bitmap_font* font, const char* text);

// ============================================================================
// Texture Atlas
// ============================================================================

typedef struct {
    image* atlas;
    int* regions;       // x, y, w, h for each region
    int region_count;
    int padding;
} texture_atlas;

texture_atlas* texture_atlas_create(int width, int height, int padding);
texture_atlas* texture_atlas_from_images(image** images, int count, int padding);
void texture_atlas_free(texture_atlas* ta);

int texture_atlas_add_region(texture_atlas* ta, int x, int y, int w, int h);
int texture_atlas_add_image(texture_atlas* ta, const image* img);
int texture_atlas_pack(texture_atlas* ta, image** images, int count);

cortex_rect texture_atlas_get_region(const texture_atlas* ta, int region);
void texture_atlas_draw_region(const texture_atlas* ta, int region, image* target, int x, int y);

// ============================================================================
// Blend Modes
// ============================================================================

typedef enum {
    BLEND_NORMAL,
    BLEND_MULTIPLY,
    BLEND_SCREEN,
    BLEND_OVERLAY,
    BLEND_DARKEN,
    BLEND_LIGHTEN,
    BLEND_COLOR_DODGE,
    BLEND_COLOR_BURN,
    BLEND_HARD_LIGHT,
    BLEND_SOFT_LIGHT,
    BLEND_DIFFERENCE,
    BLEND_EXCLUSION,
    BLEND_HUE,
    BLEND_SATURATION,
    BLEND_COLOR,
    BLEND_LUMINOSITY,
    BLEND_ADD,
    BLEND_SUBTRACT,
    BLEND_DIVIDE
} blend_mode;

color blend_colors(color src, color dst, blend_mode mode);
void image_blend_mode(image* dst, int x, int y, const image* src, blend_mode mode);
void canvas_set_blend_mode(canvas* cv, blend_mode mode);

// ============================================================================
// Utility Functions
// ============================================================================

// Coordinate conversion
void screen_to_world(canvas* cv, float screen_x, float screen_y, float* world_x, float* world_y);
void world_to_screen(canvas* cv, float world_x, float world_y, float* screen_x, float* screen_y);

// Bounds checking
bool point_in_image(const image* img, int x, int y);
bool rect_in_image(const image* img, int x, int y, int w, int h);

// Color interpolation
color bilinear_interpolate(const image* img, float x, float y);
color trilinear_interpolate(const image* img, float x, float y, float level);

// Utility
uint8_t clamp_uint8(int value);
float clamp_float_01(float value);
int next_power_of_two(int n);
bool is_power_of_two(int n);

#ifdef __cplusplus
}
#endif

#endif /* CORTEX_RUNTIME_GRAPHICS_H */
