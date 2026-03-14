// graphics.c - Image, Sprite, and Vector Graphics Library
// Part 1: Color functions

#include "graphics.h"
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <float.h>

#ifndef M_PI
#define M_PI 3.14159265358979323846
#endif

// ============================================================================
// Utility Functions
// ============================================================================

uint8_t clamp_uint8(int value) {
    if (value < 0) return 0;
    if (value > 255) return 255;
    return (uint8_t)value;
}

float clamp_float_01(float value) {
    if (value < 0.0f) return 0.0f;
    if (value > 1.0f) return 1.0f;
    return value;
}

int next_power_of_two(int n) {
    if (n <= 0) return 1;
    n--;
    n |= n >> 1;
    n |= n >> 2;
    n |= n >> 4;
    n |= n >> 8;
    n |= n >> 16;
    return n + 1;
}

bool is_power_of_two(int n) {
    return n > 0 && (n & (n - 1)) == 0;
}

// ============================================================================
// Color Creation and Manipulation
// ============================================================================

color color_make(uint8_t r, uint8_t g, uint8_t b, uint8_t a) {
    color c = {r, g, b, a};
    return c;
}

color color_from_hex(uint32_t hex) {
    color c;
    c.r = (hex >> 24) & 0xFF;
    c.g = (hex >> 16) & 0xFF;
    c.b = (hex >> 8) & 0xFF;
    c.a = hex & 0xFF;
    return c;
}

color color_from_hsv(float h, float s, float v) {
    return color_from_hsva(h, s, v, 1.0f);
}

color color_from_hsva(float h, float s, float v, float a) {
    color c = {0, 0, 0, (uint8_t)(a * 255)};
    
    if (s <= 0.0f) {
        c.r = c.g = c.b = (uint8_t)(v * 255);
        return c;
    }
    
    h = fmodf(h, 360.0f);
    if (h < 0) h += 360.0f;
    h /= 60.0f;
    
    int i = (int)h;
    float f = h - i;
    float p = v * (1.0f - s);
    float q = v * (1.0f - s * f);
    float t = v * (1.0f - s * (1.0f - f));
    
    switch (i) {
        case 0: c.r = (uint8_t)(v * 255); c.g = (uint8_t)(t * 255); c.b = (uint8_t)(p * 255); break;
        case 1: c.r = (uint8_t)(q * 255); c.g = (uint8_t)(v * 255); c.b = (uint8_t)(p * 255); break;
        case 2: c.r = (uint8_t)(p * 255); c.g = (uint8_t)(v * 255); c.b = (uint8_t)(t * 255); break;
        case 3: c.r = (uint8_t)(p * 255); c.g = (uint8_t)(q * 255); c.b = (uint8_t)(v * 255); break;
        case 4: c.r = (uint8_t)(t * 255); c.g = (uint8_t)(p * 255); c.b = (uint8_t)(v * 255); break;
        default: c.r = (uint8_t)(v * 255); c.g = (uint8_t)(p * 255); c.b = (uint8_t)(q * 255); break;
    }
    
    return c;
}

colorf colorf_make(float r, float g, float b, float a) {
    colorf c = {r, g, b, a};
    return c;
}

color colorf_to_color(colorf c) {
    color result;
    result.r = clamp_uint8((int)(c.r * 255));
    result.g = clamp_uint8((int)(c.g * 255));
    result.b = clamp_uint8((int)(c.b * 255));
    result.a = clamp_uint8((int)(c.a * 255));
    return result;
}

colorf color_to_colorf(color c) {
    colorf result;
    result.r = c.r / 255.0f;
    result.g = c.g / 255.0f;
    result.b = c.b / 255.0f;
    result.a = c.a / 255.0f;
    return result;
}

// ============================================================================
// Color Operations
// ============================================================================

color color_lerp(color a, color b, float t) {
    t = clamp_float_01(t);
    color result;
    result.r = (uint8_t)(a.r + (b.r - a.r) * t);
    result.g = (uint8_t)(a.g + (b.g - a.g) * t);
    result.b = (uint8_t)(a.b + (b.b - a.b) * t);
    result.a = (uint8_t)(a.a + (b.a - a.a) * t);
    return result;
}

color color_blend(color src, color dst) {
    // Alpha blending: src over dst
    float src_a = src.a / 255.0f;
    float dst_a = dst.a / 255.0f;
    float out_a = src_a + dst_a * (1.0f - src_a);
    
    if (out_a <= 0.0f) return COLOR_CLEAR;
    
    color result;
    result.a = (uint8_t)(out_a * 255);
    result.r = (uint8_t)((src.r * src_a + dst.r * dst_a * (1.0f - src_a)) / out_a);
    result.g = (uint8_t)((src.g * src_a + dst.g * dst_a * (1.0f - src_a)) / out_a);
    result.b = (uint8_t)((src.b * src_a + dst.b * dst_a * (1.0f - src_a)) / out_a);
    return result;
}

color color_multiply(color a, color b) {
    color result;
    result.r = (uint8_t)((a.r * b.r) / 255);
    result.g = (uint8_t)((a.g * b.g) / 255);
    result.b = (uint8_t)((a.b * b.b) / 255);
    result.a = (uint8_t)((a.a * b.a) / 255);
    return result;
}

color color_tint(color c, color tint) {
    color result;
    result.r = (uint8_t)((c.r * tint.r) / 255);
    result.g = (uint8_t)((c.g * tint.g) / 255);
    result.b = (uint8_t)((c.b * tint.b) / 255);
    result.a = (uint8_t)((c.a * tint.a) / 255);
    return result;
}

color color_fade(color c, float alpha) {
    color result = c;
    result.a = (uint8_t)(c.a * clamp_float_01(alpha));
    return result;
}

color color_brightness(color c, float factor) {
    color result;
    result.r = clamp_uint8((int)(c.r * factor));
    result.g = clamp_uint8((int)(c.g * factor));
    result.b = clamp_uint8((int)(c.b * factor));
    result.a = c.a;
    return result;
}

color color_contrast(color c, float factor) {
    factor = (factor + 1.0f) * 128.0f;
    color result;
    result.r = clamp_uint8((int)(factor + (c.r - 128) * factor / 128));
    result.g = clamp_uint8((int)(factor + (c.g - 128) * factor / 128));
    result.b = clamp_uint8((int)(factor + (c.b - 128) * factor / 128));
    result.a = c.a;
    return result;
}

color color_saturation(color c, float factor) {
    float gray = 0.299f * c.r + 0.587f * c.g + 0.114f * c.b;
    color result;
    result.r = clamp_uint8((int)(gray + (c.r - gray) * factor));
    result.g = clamp_uint8((int)(gray + (c.g - gray) * factor));
    result.b = clamp_uint8((int)(gray + (c.b - gray) * factor));
    result.a = c.a;
    return result;
}

color color_invert(color c) {
    color result;
    result.r = 255 - c.r;
    result.g = 255 - c.g;
    result.b = 255 - c.b;
    result.a = c.a;
    return result;
}

color color_grayscale(color c) {
    uint8_t gray = (uint8_t)(0.299f * c.r + 0.587f * c.g + 0.114f * c.b);
    color result = {gray, gray, gray, c.a};
    return result;
}

color color_sepia(color c) {
    color result;
    result.r = clamp_uint8((int)(0.393f * c.r + 0.769f * c.g + 0.189f * c.b));
    result.g = clamp_uint8((int)(0.349f * c.r + 0.686f * c.g + 0.168f * c.b));
    result.b = clamp_uint8((int)(0.272f * c.r + 0.534f * c.g + 0.131f * c.b));
    result.a = c.a;
    return result;
}

// ============================================================================
// Color Queries
// ============================================================================

uint32_t color_to_hex(color c) {
    return ((uint32_t)c.r << 24) | ((uint32_t)c.g << 16) | ((uint32_t)c.b << 8) | c.a;
}

uint32_t color_to_rgba(color c) {
    return ((uint32_t)c.r << 24) | ((uint32_t)c.g << 16) | ((uint32_t)c.b << 8) | c.a;
}

void color_to_hsv(color c, float* h, float* s, float* v) {
    float r = c.r / 255.0f;
    float g = c.g / 255.0f;
    float b = c.b / 255.0f;
    
    float max_val = fmaxf(r, fmaxf(g, b));
    float min_val = fminf(r, fminf(g, b));
    float delta = max_val - min_val;
    
    *v = max_val;
    
    if (delta < 0.0001f) {
        *h = 0;
        *s = 0;
        return;
    }
    
    *s = delta / max_val;
    
    if (max_val == r) {
        *h = 60.0f * fmodf((g - b) / delta, 6.0f);
    } else if (max_val == g) {
        *h = 60.0f * ((b - r) / delta + 2.0f);
    } else {
        *h = 60.0f * ((r - g) / delta + 4.0f);
    }
    
    if (*h < 0) *h += 360.0f;
}

bool color_equals(color a, color b) {
    return a.r == b.r && a.g == b.g && a.b == b.b && a.a == b.a;
}

bool color_transparent(color c) {
    return c.a == 0;
}

// ============================================================================
// Image Creation and Destruction
// ============================================================================

image* image_create(int width, int height, int channels) {
    if (width <= 0 || height <= 0 || channels < 1 || channels > 4) return NULL;
    
    image* img = (image*)malloc(sizeof(image));
    if (!img) return NULL;
    
    img->width = width;
    img->height = height;
    img->channels = channels;
    img->data = (uint8_t*)calloc((size_t)width * height * channels, sizeof(uint8_t));
    img->owns_data = true;
    
    if (!img->data) {
        free(img);
        return NULL;
    }
    
    return img;
}

image* image_from_data(int width, int height, int channels, uint8_t* data, bool copy) {
    if (width <= 0 || height <= 0 || channels < 1 || channels > 4 || !data) return NULL;
    
    image* img = (image*)malloc(sizeof(image));
    if (!img) return NULL;
    
    img->width = width;
    img->height = height;
    img->channels = channels;
    
    if (copy) {
        img->data = (uint8_t*)malloc((size_t)width * height * channels);
        if (!img->data) {
            free(img);
            return NULL;
        }
        memcpy(img->data, data, (size_t)width * height * channels);
        img->owns_data = true;
    } else {
        img->data = data;
        img->owns_data = false;
    }
    
    return img;
}

image* image_clone(const image* img) {
    if (!img || !img->data) return NULL;
    return image_from_data(img->width, img->height, img->channels, img->data, true);
}

image* image_subimage(const image* img, int x, int y, int w, int h) {
    if (!img || !img->data) return NULL;
    if (x < 0 || y < 0 || w <= 0 || h <= 0) return NULL;
    if (x + w > img->width || y + h > img->height) return NULL;
    
    image* sub = image_create(w, h, img->channels);
    if (!sub) return NULL;
    
    for (int row = 0; row < h; row++) {
        const uint8_t* src = img->data + ((y + row) * img->width + x) * img->channels;
        uint8_t* dst = sub->data + row * w * img->channels;
        memcpy(dst, src, (size_t)w * img->channels);
    }
    
    return sub;
}

void image_free(image* img) {
    if (!img) return;
    if (img->owns_data && img->data) {
        free(img->data);
    }
    free(img);
}

// ============================================================================
// Image Properties
// ============================================================================

int image_width(const image* img) { return img ? img->width : 0; }
int image_height(const image* img) { return img ? img->height : 0; }
int image_channels(const image* img) { return img ? img->channels : 0; }

size_t image_size(const image* img) {
    if (!img) return 0;
    return (size_t)img->width * img->height * img->channels;
}

bool image_valid(const image* img) {
    return img && img->data && img->width > 0 && img->height > 0;
}

vec2 image_center(const image* img) {
    if (!img) return make_vec2(0, 0);
    return make_vec2(img->width / 2.0f, img->height / 2.0f);
}

cortex_rect image_bounds(const image* img) {
    if (!img) return (cortex_rect){0, 0, 0, 0};
    return (cortex_rect){0, 0, (float)img->width, (float)img->height};
}

// ============================================================================
// Pixel Access
// ============================================================================

color image_get_pixel(const image* img, int x, int y) {
    if (!image_valid(img) || x < 0 || y < 0 || x >= img->width || y >= img->height) {
        return COLOR_CLEAR;
    }
    
    const uint8_t* pixel = img->data + (y * img->width + x) * img->channels;
    
    color c = {0, 0, 0, 255};
    switch (img->channels) {
        case 1: c.r = c.g = c.b = pixel[0]; break;
        case 2: c.r = c.g = c.b = pixel[0]; c.a = pixel[1]; break;
        case 3: c.r = pixel[0]; c.g = pixel[1]; c.b = pixel[2]; break;
        case 4: c.r = pixel[0]; c.g = pixel[1]; c.b = pixel[2]; c.a = pixel[3]; break;
    }
    return c;
}

void image_set_pixel(image* img, int x, int y, color c) {
    if (!image_valid(img) || x < 0 || y < 0 || x >= img->width || y >= img->height) return;
    
    uint8_t* pixel = img->data + (y * img->width + x) * img->channels;
    
    switch (img->channels) {
        case 1: pixel[0] = (uint8_t)(0.299f * c.r + 0.587f * c.g + 0.114f * c.b); break;
        case 2: pixel[0] = (uint8_t)(0.299f * c.r + 0.587f * c.g + 0.114f * c.b); pixel[1] = c.a; break;
        case 3: pixel[0] = c.r; pixel[1] = c.g; pixel[2] = c.b; break;
        case 4: pixel[0] = c.r; pixel[1] = c.g; pixel[2] = c.b; pixel[3] = c.a; break;
    }
}

uint8_t image_get_channel(const image* img, int x, int y, int channel) {
    if (!image_valid(img) || x < 0 || y < 0 || x >= img->width || y >= img->height) return 0;
    if (channel < 0 || channel >= img->channels) return 0;
    
    return img->data[(y * img->width + x) * img->channels + channel];
}

void image_set_channel(image* img, int x, int y, int channel, uint8_t value) {
    if (!image_valid(img) || x < 0 || y < 0 || x >= img->width || y >= img->height) return;
    if (channel < 0 || channel >= img->channels) return;
    
    img->data[(y * img->width + x) * img->channels + channel] = value;
}

color* image_get_pixel_ptr(image* img, int x, int y) {
    if (!image_valid(img) || x < 0 || y < 0 || x >= img->width || y >= img->height) return NULL;
    if (img->channels != 4) return NULL;
    return (color*)(img->data + (y * img->width + x) * 4);
}

const color* image_get_pixel_ptr_const(const image* img, int x, int y) {
    if (!image_valid(img) || x < 0 || y < 0 || x >= img->width || y >= img->height) return NULL;
    if (img->channels != 4) return NULL;
    return (const color*)(img->data + (y * img->width + x) * 4);
}

// ============================================================================
// Image Operations
// ============================================================================

void image_clear(image* img, color c) {
    if (!image_valid(img)) return;
    
    if (img->channels == 4) {
        uint32_t rgba = color_to_rgba(c);
        uint32_t* pixels = (uint32_t*)img->data;
        size_t count = (size_t)img->width * img->height;
        for (size_t i = 0; i < count; i++) {
            pixels[i] = rgba;
        }
    } else {
        for (int y = 0; y < img->height; y++) {
            for (int x = 0; x < img->width; x++) {
                image_set_pixel(img, x, y, c);
            }
        }
    }
}

void image_fill_rect(image* img, int x, int y, int w, int h, color c) {
    if (!image_valid(img)) return;
    
    for (int py = y; py < y + h; py++) {
        for (int px = x; px < x + w; px++) {
            if (px >= 0 && py >= 0 && px < img->width && py < img->height) {
                image_set_pixel(img, px, py, c);
            }
        }
    }
}

void image_copy(image* dst, int dx, int dy, const image* src, int sx, int sy, int sw, int sh) {
    if (!image_valid(dst) || !image_valid(src)) return;
    
    // Clamp source rect
    if (sx < 0) { sw += sx; dx -= sx; sx = 0; }
    if (sy < 0) { sh += sy; dy -= sy; sy = 0; }
    if (sx + sw > src->width) sw = src->width - sx;
    if (sy + sh > src->height) sh = src->height - sy;
    
    // Clamp dest rect
    if (dx < 0) { sw += dx; sx -= dx; dx = 0; }
    if (dy < 0) { sh += dy; sy -= dy; dy = 0; }
    if (dx + sw > dst->width) sw = dst->width - dx;
    if (dy + sh > dst->height) sh = dst->height - dy;
    
    if (sw <= 0 || sh <= 0) return;
    
    // Copy row by row
    for (int row = 0; row < sh; row++) {
        const uint8_t* src_row = src->data + ((sy + row) * src->width + sx) * src->channels;
        uint8_t* dst_row = dst->data + ((dy + row) * dst->width + dx) * dst->channels;
        
        if (src->channels == dst->channels) {
            memcpy(dst_row, src_row, (size_t)sw * src->channels);
        } else {
            for (int px = 0; px < sw; px++) {
                color c = image_get_pixel(src, sx + px, sy + row);
                image_set_pixel(dst, dx + px, dy + row, c);
            }
        }
    }
}

void image_paste(image* dst, int dx, int dy, const image* src) {
    if (!image_valid(src)) return;
    image_copy(dst, dx, dy, src, 0, 0, src->width, src->height);
}

void image_blend(image* dst, int dx, int dy, const image* src, float alpha) {
    if (!image_valid(dst) || !image_valid(src)) return;
    
    for (int y = 0; y < src->height; y++) {
        for (int x = 0; x < src->width; x++) {
            int dst_x = dx + x;
            int dst_y = dy + y;
            
            if (dst_x >= 0 && dst_y >= 0 && dst_x < dst->width && dst_y < dst->height) {
                color src_c = image_get_pixel(src, x, y);
                color dst_c = image_get_pixel(dst, dst_x, dst_y);
                
                src_c.a = (uint8_t)(src_c.a * alpha);
                color blended = color_blend(src_c, dst_c);
                image_set_pixel(dst, dst_x, dst_y, blended);
            }
        }
    }
}

// ============================================================================
// Image Transformations
// ============================================================================

image* image_resize_nearest(const image* img, int new_width, int new_height) {
    if (!image_valid(img)) return NULL;
    
    image* result = image_create(new_width, new_height, img->channels);
    if (!result) return NULL;
    
    float x_ratio = (float)img->width / new_width;
    float y_ratio = (float)img->height / new_height;
    
    for (int y = 0; y < new_height; y++) {
        for (int x = 0; x < new_width; x++) {
            int src_x = (int)(x * x_ratio);
            int src_y = (int)(y * y_ratio);
            color c = image_get_pixel(img, src_x, src_y);
            image_set_pixel(result, x, y, c);
        }
    }
    
    return result;
}

image* image_resize_bilinear(const image* img, int new_width, int new_height) {
    if (!image_valid(img)) return NULL;
    
    image* result = image_create(new_width, new_height, img->channels);
    if (!result) return NULL;
    
    float x_ratio = (float)(img->width - 1) / (new_width - 1);
    float y_ratio = (float)(img->height - 1) / (new_height - 1);
    
    for (int y = 0; y < new_height; y++) {
        for (int x = 0; x < new_width; x++) {
            float src_x = x * x_ratio;
            float src_y = y * y_ratio;
            
            int x0 = (int)src_x;
            int y0 = (int)src_y;
            int x1 = x0 + 1 < img->width ? x0 + 1 : x0;
            int y1 = y0 + 1 < img->height ? y0 + 1 : y0;
            
            float fx = src_x - x0;
            float fy = src_y - y0;
            
            color c00 = image_get_pixel(img, x0, y0);
            color c10 = image_get_pixel(img, x1, y0);
            color c01 = image_get_pixel(img, x0, y1);
            color c11 = image_get_pixel(img, x1, y1);
            
            color c;
            c.r = (uint8_t)((1-fx)*(1-fy)*c00.r + fx*(1-fy)*c10.r + (1-fx)*fy*c01.r + fx*fy*c11.r);
            c.g = (uint8_t)((1-fx)*(1-fy)*c00.g + fx*(1-fy)*c10.g + (1-fx)*fy*c01.g + fx*fy*c11.g);
            c.b = (uint8_t)((1-fx)*(1-fy)*c00.b + fx*(1-fy)*c10.b + (1-fx)*fy*c01.b + fx*fy*c11.b);
            c.a = (uint8_t)((1-fx)*(1-fy)*c00.a + fx*(1-fy)*c10.a + (1-fx)*fy*c01.a + fx*fy*c11.a);
            
            image_set_pixel(result, x, y, c);
        }
    }
    
    return result;
}

image* image_resize(const image* img, int new_width, int new_height) {
    return image_resize_bilinear(img, new_width, new_height);
}

image* image_scale(const image* img, float scale_x, float scale_y) {
    if (!image_valid(img)) return NULL;
    return image_resize(img, (int)(img->width * scale_x), (int)(img->height * scale_y));
}

image* image_rotate_90(const image* img, int times) {
    if (!image_valid(img)) return NULL;
    
    times = ((times % 4) + 4) % 4;
    if (times == 0) return image_clone(img);
    
    if (times == 2) return image_rotate_180(img);
    
    image* result = image_create(img->height, img->width, img->channels);
    if (!result) return NULL;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            color c = image_get_pixel(img, x, y);
            if (times == 1) {
                image_set_pixel(result, img->height - 1 - y, x, c);
            } else {
                image_set_pixel(result, y, img->width - 1 - x, c);
            }
        }
    }
    
    return result;
}

image* image_rotate_180(const image* img) {
    if (!image_valid(img)) return NULL;
    
    image* result = image_create(img->width, img->height, img->channels);
    if (!result) return NULL;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            color c = image_get_pixel(img, x, y);
            image_set_pixel(result, img->width - 1 - x, img->height - 1 - y, c);
        }
    }
    
    return result;
}

image* image_flip_h(image* img) {
    if (!image_valid(img)) return NULL;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width / 2; x++) {
            color c1 = image_get_pixel(img, x, y);
            color c2 = image_get_pixel(img, img->width - 1 - x, y);
            image_set_pixel(img, x, y, c2);
            image_set_pixel(img, img->width - 1 - x, y, c1);
        }
    }
    return img;
}

image* image_flip_v(image* img) {
    if (!image_valid(img)) return NULL;
    
    for (int y = 0; y < img->height / 2; y++) {
        for (int x = 0; x < img->width; x++) {
            color c1 = image_get_pixel(img, x, y);
            color c2 = image_get_pixel(img, x, img->height - 1 - y);
            image_set_pixel(img, x, y, c2);
            image_set_pixel(img, x, img->height - 1 - y, c1);
        }
    }
    return img;
}

image* image_flip_both(image* img) {
    return image_flip_h(image_flip_v(img));
}

image* image_crop(const image* img, int x, int y, int w, int h) {
    return image_subimage(img, x, y, w, h);
}

// ============================================================================
// Image Filters
// ============================================================================

void image_grayscale(image* img) {
    if (!image_valid(img)) return;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            color c = image_get_pixel(img, x, y);
            image_set_pixel(img, x, y, color_grayscale(c));
        }
    }
}

void image_sepia(image* img) {
    if (!image_valid(img)) return;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            color c = image_get_pixel(img, x, y);
            image_set_pixel(img, x, y, color_sepia(c));
        }
    }
}

void image_invert(image* img) {
    if (!image_valid(img)) return;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            color c = image_get_pixel(img, x, y);
            image_set_pixel(img, x, y, color_invert(c));
        }
    }
}

void image_brightness(image* img, float factor) {
    if (!image_valid(img)) return;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            color c = image_get_pixel(img, x, y);
            image_set_pixel(img, x, y, color_brightness(c, factor));
        }
    }
}

void image_contrast(image* img, float factor) {
    if (!image_valid(img)) return;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            color c = image_get_pixel(img, x, y);
            image_set_pixel(img, x, y, color_contrast(c, factor));
        }
    }
}

void image_saturation(image* img, float factor) {
    if (!image_valid(img)) return;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            color c = image_get_pixel(img, x, y);
            image_set_pixel(img, x, y, color_saturation(c, factor));
        }
    }
}

void image_gamma(image* img, float gamma) {
    if (!image_valid(img)) return;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            color c = image_get_pixel(img, x, y);
            colorf cf = color_to_colorf(c);
            cf.r = powf(cf.r, gamma);
            cf.g = powf(cf.g, gamma);
            cf.b = powf(cf.b, gamma);
            image_set_pixel(img, x, y, colorf_to_color(cf));
        }
    }
}

void image_tint(image* img, color tint) {
    if (!image_valid(img)) return;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            color c = image_get_pixel(img, x, y);
            image_set_pixel(img, x, y, color_tint(c, tint));
        }
    }
}

void image_fade(image* img, float alpha) {
    if (!image_valid(img)) return;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            color c = image_get_pixel(img, x, y);
            image_set_pixel(img, x, y, color_fade(c, alpha));
        }
    }
}

void image_blur_box(image* img, int radius) {
    if (!image_valid(img) || radius < 1) return;
    
    image* temp = image_clone(img);
    if (!temp) return;
    
    int size = radius * 2 + 1;
    (void)size; // used for area calculation conceptually
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            int r = 0, g = 0, b = 0, a = 0, count = 0;
            
            for (int dy = -radius; dy <= radius; dy++) {
                for (int dx = -radius; dx <= radius; dx++) {
                    int nx = x + dx;
                    int ny = y + dy;
                    if (nx >= 0 && ny >= 0 && nx < img->width && ny < img->height) {
                        color c = image_get_pixel(temp, nx, ny);
                        r += c.r; g += c.g; b += c.b; a += c.a;
                        count++;
                    }
                }
            }
            
            color c = {(uint8_t)(r/count), (uint8_t)(g/count), (uint8_t)(b/count), (uint8_t)(a/count)};
            image_set_pixel(img, x, y, c);
        }
    }
    
    image_free(temp);
}

void image_threshold(image* img, uint8_t threshold) {
    if (!image_valid(img)) return;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            color c = image_get_pixel(img, x, y);
            uint8_t gray = (uint8_t)(0.299f * c.r + 0.587f * c.g + 0.114f * c.b);
            uint8_t val = gray >= threshold ? 255 : 0;
            image_set_pixel(img, x, y, (color){val, val, val, c.a});
        }
    }
}

void image_pixelate(image* img, int block_size) {
    if (!image_valid(img) || block_size < 1) return;
    
    for (int by = 0; by < img->height; by += block_size) {
        for (int bx = 0; bx < img->width; bx += block_size) {
            int r = 0, g = 0, b = 0, a = 0, count = 0;
            
            for (int y = by; y < by + block_size && y < img->height; y++) {
                for (int x = bx; x < bx + block_size && x < img->width; x++) {
                    color c = image_get_pixel(img, x, y);
                    r += c.r; g += c.g; b += c.b; a += c.a;
                    count++;
                }
            }
            
            color avg = {(uint8_t)(r/count), (uint8_t)(g/count), (uint8_t)(b/count), (uint8_t)(a/count)};
            
            for (int y = by; y < by + block_size && y < img->height; y++) {
                for (int x = bx; x < bx + block_size && x < img->width; x++) {
                    image_set_pixel(img, x, y, avg);
                }
            }
        }
    }
}

void image_convolve(image* img, const float* kernel, int kernel_size) {
    if (!image_valid(img) || !kernel || kernel_size < 1) return;
    
    image* temp = image_clone(img);
    if (!temp) return;
    
    int half = kernel_size / 2;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            float r = 0, g = 0, b = 0;
            
            for (int ky = 0; ky < kernel_size; ky++) {
                for (int kx = 0; kx < kernel_size; kx++) {
                    int nx = x + kx - half;
                    int ny = y + ky - half;
                    
                    if (nx >= 0 && ny >= 0 && nx < img->width && ny < img->height) {
                        color c = image_get_pixel(temp, nx, ny);
                        float k = kernel[ky * kernel_size + kx];
                        r += c.r * k;
                        g += c.g * k;
                        b += c.b * k;
                    }
                }
            }
            
            color orig = image_get_pixel(temp, x, y);
            color result = {clamp_uint8((int)r), clamp_uint8((int)g), clamp_uint8((int)b), orig.a};
            image_set_pixel(img, x, y, result);
        }
    }
    
    image_free(temp);
}

void image_sharpen(image* img, float amount) {
    float kernel[9] = {
        0, -amount, 0,
        -amount, 1 + 4*amount, -amount,
        0, -amount, 0
    };
    image_convolve(img, kernel, 3);
}

void image_emboss(image* img) {
    float kernel[9] = {
        -2, -1, 0,
        -1, 1, 1,
        0, 1, 2
    };
    image_convolve(img, kernel, 3);
}

void image_edge_detect(image* img) {
    float kernel[9] = {
        -1, -1, -1,
        -1, 8, -1,
        -1, -1, -1
    };
    image_convolve(img, kernel, 3);
}

// ============================================================================
// Image Utilities
// ============================================================================

image* image_convert(const image* img, int channels) {
    if (!image_valid(img) || channels < 1 || channels > 4) return NULL;
    if (img->channels == channels) return image_clone(img);
    
    image* result = image_create(img->width, img->height, channels);
    if (!result) return NULL;
    
    for (int y = 0; y < img->height; y++) {
        for (int x = 0; x < img->width; x++) {
            color c = image_get_pixel(img, x, y);
            image_set_pixel(result, x, y, c);
        }
    }
    
    return result;
}

bool image_has_alpha(const image* img) {
    return img && img->channels >= 2 && img->channels <= 4;
}

color image_average_color(const image* img) {
    if (!image_valid(img)) return COLOR_CLEAR;
    
    long long r = 0, g = 0, b = 0, a = 0;
    size_t count = (size_t)img->width * img->height;
    
    for (size_t i = 0; i < count; i++) {
        color c = image_get_pixel(img, i % img->width, i / img->width);
        r += c.r; g += c.g; b += c.b; a += c.a;
    }
    
    return (color){(uint8_t)(r/count), (uint8_t)(g/count), (uint8_t)(b/count), (uint8_t)(a/count)};
}

color bilinear_interpolate(const image* img, float x, float y) {
    if (!image_valid(img)) return COLOR_CLEAR;
    
    int x0 = (int)x;
    int y0 = (int)y;
    int x1 = x0 + 1;
    int y1 = y0 + 1;
    
    if (x0 < 0) x0 = 0; 
    if (x0 >= img->width) x0 = img->width - 1; 
    if (y0 < 0) y0 = 0; 
    if (y0 >= img->height) y0 = img->height - 1; 
    if (x1 < 0) x1 = 0; 
    if (x1 >= img->width) x1 = img->width - 1; 
    if (y1 < 0) y1 = 0; 
    if (y1 >= img->height) y1 = img->height - 1;
    
    float fx = x - x0;
    float fy = y - y0;
    
    color c00 = image_get_pixel(img, x0, y0);
    color c10 = image_get_pixel(img, x1, y0);
    color c01 = image_get_pixel(img, x0, y1);
    color c11 = image_get_pixel(img, x1, y1);
    
    return color_lerp(color_lerp(c00, c10, fx), color_lerp(c01, c11, fx), fy);
}

bool point_in_image(const image* img, int x, int y) {
    return img && x >= 0 && y >= 0 && x < img->width && y < img->height;
}

bool rect_in_image(const image* img, int x, int y, int w, int h) {
    return img && x >= 0 && y >= 0 && x + w <= img->width && y + h <= img->height;
}

// ============================================================================
// Sprite Sheet Management
// ============================================================================

sprite_sheet* sprite_sheet_create(image* img) {
    if (!img) return NULL;
    
    sprite_sheet* sheet = (sprite_sheet*)calloc(1, sizeof(sprite_sheet));
    if (!sheet) return NULL;
    
    sheet->sheet = img;
    sheet->frames = NULL;
    sheet->frame_count = 0;
    sheet->animations = NULL;
    sheet->animation_count = 0;
    
    return sheet;
}

sprite_sheet* sprite_sheet_from_grid(image* img, int frame_width, int frame_height, int columns, int rows) {
    sprite_sheet* sheet = sprite_sheet_create(img);
    if (!sheet) return NULL;
    
    sprite_sheet_add_frames_grid(sheet, 0, 0, frame_width, frame_height, columns, rows);
    return sheet;
}

void sprite_sheet_free(sprite_sheet* sheet) {
    if (!sheet) return;
    
    if (sheet->frames) free(sheet->frames);
    if (sheet->animations) free(sheet->animations);
    free(sheet);
}

int sprite_sheet_add_frame(sprite_sheet* sheet, int x, int y, int w, int h) {
    if (!sheet) return -1;
    
    int new_count = sheet->frame_count + 1;
    sprite_frame* new_frames = (sprite_frame*)realloc(sheet->frames, new_count * sizeof(sprite_frame));
    if (!new_frames) return -1;
    
    sheet->frames = new_frames;
    sheet->frames[sheet->frame_count] = (sprite_frame){x, y, w, h, 0.5f, 0.5f};
    sheet->frame_count++;
    
    return sheet->frame_count - 1;
}

int sprite_sheet_add_frames_grid(sprite_sheet* sheet, int start_x, int start_y, int frame_w, int frame_h, int columns, int rows) {
    if (!sheet) return 0;
    
    int added = 0;
    for (int row = 0; row < rows; row++) {
        for (int col = 0; col < columns; col++) {
            int x = start_x + col * frame_w;
            int y = start_y + row * frame_h;
            if (sprite_sheet_add_frame(sheet, x, y, frame_w, frame_h) >= 0) {
                added++;
            }
        }
    }
    return added;
}

void sprite_sheet_set_pivot(sprite_sheet* sheet, int frame, float pivot_x, float pivot_y) {
    if (!sheet || frame < 0 || frame >= sheet->frame_count) return;
    sheet->frames[frame].pivot_x = pivot_x;
    sheet->frames[frame].pivot_y = pivot_y;
}

sprite_frame* sprite_sheet_get_frame(sprite_sheet* sheet, int frame) {
    if (!sheet || frame < 0 || frame >= sheet->frame_count) return NULL;
    return &sheet->frames[frame];
}

int sprite_sheet_add_animation(sprite_sheet* sheet, int start_frame, int frame_count, float duration, bool looping) {
    if (!sheet || start_frame < 0 || frame_count <= 0) return -1;
    
    int new_count = sheet->animation_count + 1;
    sprite_animation* new_anim = (sprite_animation*)realloc(sheet->animations, new_count * sizeof(sprite_animation));
    if (!new_anim) return -1;
    
    sheet->animations = new_anim;
    sheet->animations[sheet->animation_count] = (sprite_animation){start_frame, frame_count, duration, looping};
    sheet->animation_count++;
    
    return sheet->animation_count - 1;
}

sprite_animation* sprite_sheet_get_animation(sprite_sheet* sheet, int animation) {
    if (!sheet || animation < 0 || animation >= sheet->animation_count) return NULL;
    return &sheet->animations[animation];
}

// ============================================================================
// Sprite Instance
// ============================================================================

sprite_instance* sprite_instance_create(sprite_sheet* sheet) {
    if (!sheet) return NULL;
    
    sprite_instance* inst = (sprite_instance*)calloc(1, sizeof(sprite_instance));
    if (!inst) return NULL;
    
    inst->sheet = sheet;
    inst->current_frame = 0;
    inst->current_animation = -1;
    inst->frame_time = 0;
    inst->speed = 1.0f;
    inst->playing = false;
    inst->flipped_h = false;
    inst->flipped_v = false;
    inst->scale_x = 1.0f;
    inst->scale_y = 1.0f;
    inst->rotation = 0;
    inst->alpha = 1.0f;
    inst->tint = COLOR_WHITE;
    
    return inst;
}

void sprite_instance_free(sprite_instance* sprite) {
    if (sprite) free(sprite);
}

void sprite_instance_play(sprite_instance* sprite, int animation) {
    if (!sprite || !sprite->sheet) return;
    if (animation < 0 || animation >= sprite->sheet->animation_count) return;
    
    sprite->current_animation = animation;
    sprite->current_frame = sprite->sheet->animations[animation].start_frame;
    sprite->frame_time = 0;
    sprite->playing = true;
}

void sprite_instance_stop(sprite_instance* sprite) {
    if (sprite) sprite->playing = false;
}

void sprite_instance_reset(sprite_instance* sprite) {
    if (!sprite) return;
    sprite->current_frame = 0;
    sprite->frame_time = 0;
    sprite->playing = false;
}

void sprite_instance_update(sprite_instance* sprite, float dt) {
    if (!sprite || !sprite->playing || !sprite->sheet) return;
    if (sprite->current_animation < 0) return;
    
    sprite_animation* anim = &sprite->sheet->animations[sprite->current_animation];
    float frame_duration = anim->duration / anim->frame_count;
    
    sprite->frame_time += dt * sprite->speed;
    
    if (sprite->frame_time >= frame_duration) {
        sprite->frame_time -= frame_duration;
        sprite->current_frame++;
        
        if (sprite->current_frame >= anim->start_frame + anim->frame_count) {
            if (anim->looping) {
                sprite->current_frame = anim->start_frame;
            } else {
                sprite->current_frame = anim->start_frame + anim->frame_count - 1;
                sprite->playing = false;
            }
        }
    }
}

void sprite_instance_set_frame(sprite_instance* sprite, int frame) {
    if (!sprite || !sprite->sheet) return;
    if (frame < 0 || frame >= sprite->sheet->frame_count) return;
    sprite->current_frame = frame;
    sprite->frame_time = 0;
}

void sprite_instance_next_frame(sprite_instance* sprite) {
    if (!sprite || !sprite->sheet) return;
    sprite->current_frame = (sprite->current_frame + 1) % sprite->sheet->frame_count;
}

void sprite_instance_prev_frame(sprite_instance* sprite) {
    if (!sprite || !sprite->sheet) return;
    sprite->current_frame = (sprite->current_frame - 1 + sprite->sheet->frame_count) % sprite->sheet->frame_count;
}

void sprite_instance_set_scale(sprite_instance* sprite, float scale) {
    if (sprite) sprite->scale_x = sprite->scale_y = scale;
}

void sprite_instance_set_scale_xy(sprite_instance* sprite, float scale_x, float scale_y) {
    if (sprite) { sprite->scale_x = scale_x; sprite->scale_y = scale_y; }
}

void sprite_instance_set_rotation(sprite_instance* sprite, float radians) {
    if (sprite) sprite->rotation = radians;
}

void sprite_instance_set_alpha(sprite_instance* sprite, float alpha) {
    if (sprite) sprite->alpha = clamp_float_01(alpha);
}

void sprite_instance_set_tint(sprite_instance* sprite, color tint) {
    if (sprite) sprite->tint = tint;
}

void sprite_instance_set_flip(sprite_instance* sprite, bool h, bool v) {
    if (sprite) { sprite->flipped_h = h; sprite->flipped_v = v; }
}

// ============================================================================
// Sprite Rendering
// ============================================================================

void sprite_draw(const sprite_instance* sprite, image* target, float x, float y) {
    if (!sprite || !sprite->sheet || !target || !sprite->sheet->sheet) return;
    if (sprite->current_frame < 0 || sprite->current_frame >= sprite->sheet->frame_count) return;
    
    sprite_frame* frame = &sprite->sheet->frames[sprite->current_frame];
    image* src = sprite->sheet->sheet;
    
    // Calculate source position with flip
    int sx = frame->x;
    int sy = frame->y;
    int sw = frame->width;
    int sh = frame->height;
    
    // Calculate destination with pivot
    float dx = x - frame->pivot_x * sw * sprite->scale_x;
    float dy = y - frame->pivot_y * sh * sprite->scale_y;
    
    // Draw scaled (simple nearest-neighbor for now)
    for (int py = 0; py < sh; py++) {
        for (int px = 0; px < sw; px++) {
            int src_x = sprite->flipped_h ? (sx + sw - 1 - px) : (sx + px);
            int src_y = sprite->flipped_v ? (sy + sh - 1 - py) : (sy + py);
            
            color c = image_get_pixel(src, src_x, src_y);
            
            // Apply tint and alpha
            c = color_tint(c, sprite->tint);
            c = color_fade(c, sprite->alpha);
            
            // Draw scaled pixel
            for (float sy2 = 0; sy2 < sprite->scale_y; sy2++) {
                for (float sx2 = 0; sx2 < sprite->scale_x; sx2++) {
                    int dst_x = (int)(dx + px * sprite->scale_x + sx2);
                    int dst_y = (int)(dy + py * sprite->scale_y + sy2);
                    
                    if (dst_x >= 0 && dst_y >= 0 && dst_x < target->width && dst_y < target->height) {
                        color dst_c = image_get_pixel(target, dst_x, dst_y);
                        image_set_pixel(target, dst_x, dst_y, color_blend(c, dst_c));
                    }
                }
            }
        }
    }
}

void sprite_draw_scaled(const sprite_instance* sprite, image* target, float x, float y, float scale) {
    if (!sprite) return;
    float old_sx = sprite->scale_x, old_sy = sprite->scale_y;
    sprite_instance_set_scale((sprite_instance*)sprite, scale);
    sprite_draw(sprite, target, x, y);
    ((sprite_instance*)sprite)->scale_x = old_sx;
    ((sprite_instance*)sprite)->scale_y = old_sy;
}

cortex_rect sprite_get_bounds(const sprite_instance* sprite) {
    if (!sprite || !sprite->sheet || sprite->current_frame < 0) {
        return (cortex_rect){0, 0, 0, 0};
    }
    
    sprite_frame* frame = &sprite->sheet->frames[sprite->current_frame];
    return (cortex_rect){
        -frame->pivot_x * frame->width * sprite->scale_x,
        -frame->pivot_y * frame->height * sprite->scale_y,
        frame->width * sprite->scale_x,
        frame->height * sprite->scale_y
    };
}

vec2 sprite_get_size(const sprite_instance* sprite) {
    if (!sprite || !sprite->sheet || sprite->current_frame < 0) {
        return make_vec2(0, 0);
    }
    
    sprite_frame* frame = &sprite->sheet->frames[sprite->current_frame];
    return make_vec2(frame->width * sprite->scale_x, frame->height * sprite->scale_y);
}

vec2 sprite_get_pivot(const sprite_instance* sprite) {
    if (!sprite || !sprite->sheet || sprite->current_frame < 0) {
        return make_vec2(0.5f, 0.5f);
    }
    
    sprite_frame* frame = &sprite->sheet->frames[sprite->current_frame];
    return make_vec2(frame->pivot_x, frame->pivot_y);
}

// ============================================================================
// Canvas Creation
// ============================================================================

canvas* canvas_create(int width, int height) {
    image* img = image_create(width, height, 4);
    if (!img) return NULL;
    
    canvas* cv = canvas_from_image(img);
    if (!cv) {
        image_free(img);
        return NULL;
    }
    cv->target = img;
    return cv;
}

canvas* canvas_from_image(image* img) {
    if (!img) return NULL;
    
    canvas* cv = (canvas*)calloc(1, sizeof(canvas));
    if (!cv) return NULL;
    
    cv->target = img;
    cv->clip_enabled = false;
    cv->fill_color = COLOR_WHITE;
    cv->stroke_color = COLOR_BLACK;
    cv->stroke_width = 1.0f;
    
    // Identity transform
    cv->transform[0] = 1; cv->transform[1] = 0;
    cv->transform[2] = 0; cv->transform[3] = 1;
    cv->transform[4] = 0; cv->transform[5] = 0;
    
    return cv;
}

void canvas_free(canvas* cv) {
    if (cv) free(cv);
}

// ============================================================================
// Canvas State
// ============================================================================

void canvas_set_fill(canvas* cv, color c) { if (cv) cv->fill_color = c; }
void canvas_set_stroke(canvas* cv, color c) { if (cv) cv->stroke_color = c; }
void canvas_set_stroke_width(canvas* cv, float width) { if (cv) cv->stroke_width = width; }

void canvas_set_clip(canvas* cv, int x, int y, int w, int h) {
    if (!cv) return;
    cv->clip_x = x; cv->clip_y = y; cv->clip_w = w; cv->clip_h = h;
    cv->clip_enabled = true;
}

void canvas_clear_clip(canvas* cv) { if (cv) cv->clip_enabled = false; }

void canvas_clear(canvas* cv, color c) {
    if (cv && cv->target) image_clear(cv->target, c);
}

// ============================================================================
// Canvas Transformations
// ============================================================================

void canvas_reset_transform(canvas* cv) {
    if (!cv) return;
    cv->transform[0] = 1; cv->transform[1] = 0;
    cv->transform[2] = 0; cv->transform[3] = 1;
    cv->transform[4] = 0; cv->transform[5] = 0;
}

void canvas_translate(canvas* cv, float tx, float ty) {
    if (!cv) return;
    cv->transform[4] += cv->transform[0] * tx + cv->transform[2] * ty;
    cv->transform[5] += cv->transform[1] * tx + cv->transform[3] * ty;
}

void canvas_scale(canvas* cv, float sx, float sy) {
    if (!cv) return;
    cv->transform[0] *= sx; cv->transform[2] *= sy;
    cv->transform[1] *= sx; cv->transform[3] *= sy;
}

void canvas_rotate(canvas* cv, float radians) {
    if (!cv) return;
    float c = cosf(radians), s = sinf(radians);
    float t0 = cv->transform[0], t1 = cv->transform[1];
    float t2 = cv->transform[2], t3 = cv->transform[3];
    cv->transform[0] = t0 * c + t2 * s;
    cv->transform[1] = t1 * c + t3 * s;
    cv->transform[2] = t2 * c - t0 * s;
    cv->transform[3] = t3 * c - t1 * s;
}

void screen_to_world(canvas* cv, float sx, float sy, float* wx, float* wy) {
    if (!cv || !wx || !wy) return;
    float det = cv->transform[0] * cv->transform[3] - cv->transform[1] * cv->transform[2];
    if (fabsf(det) < 0.0001f) { *wx = sx; *wy = sy; return; }
    float dx = sx - cv->transform[4];
    float dy = sy - cv->transform[5];
    *wx = (cv->transform[3] * dx - cv->transform[2] * dy) / det;
    *wy = (cv->transform[0] * dy - cv->transform[1] * dx) / det;
}

void world_to_screen(canvas* cv, float wx, float wy, float* sx, float* sy) {
    if (!cv || !sx || !sy) return;
    *sx = cv->transform[0] * wx + cv->transform[2] * wy + cv->transform[4];
    *sy = cv->transform[1] * wx + cv->transform[3] * wy + cv->transform[5];
}

// ============================================================================
// Canvas Basic Shapes
// ============================================================================

void canvas_draw_point(canvas* cv, float x, float y) {
    if (!cv || !cv->target) return;
    float sx, sy; world_to_screen(cv, x, y, &sx, &sy);
    int px = (int)sx, py = (int)sy;
    if (px >= 0 && py >= 0 && px < cv->target->width && py < cv->target->height) {
        image_set_pixel(cv->target, px, py, cv->stroke_color);
    }
}

void canvas_draw_line(canvas* cv, float x1, float y1, float x2, float y2) {
    if (!cv || !cv->target) return;
    float sx1, sy1, sx2, sy2;
    world_to_screen(cv, x1, y1, &sx1, &sy1);
    world_to_screen(cv, x2, y2, &sx2, &sy2);
    draw_line(cv->target, (int)sx1, (int)sy1, (int)sx2, (int)sy2, cv->stroke_color);
}

void canvas_draw_rect(canvas* cv, float x, float y, float w, float h) {
    if (!cv || !cv->target) return;
    float sx, sy; world_to_screen(cv, x, y, &sx, &sy);
    draw_rect(cv->target, (int)sx, (int)sy, (int)w, (int)h, cv->stroke_color);
}

void canvas_fill_rect(canvas* cv, float x, float y, float w, float h) {
    if (!cv || !cv->target) return;
    float sx, sy; world_to_screen(cv, x, y, &sx, &sy);
    fill_rect(cv->target, (int)sx, (int)sy, (int)w, (int)h, cv->fill_color);
}

void canvas_draw_circle(canvas* cv, float cx, float cy, float radius) {
    if (!cv || !cv->target) return;
    float sx, sy; world_to_screen(cv, cx, cy, &sx, &sy);
    draw_circle(cv->target, (int)sx, (int)sy, (int)radius, cv->stroke_color);
}

void canvas_fill_circle(canvas* cv, float cx, float cy, float radius) {
    if (!cv || !cv->target) return;
    float sx, sy; world_to_screen(cv, cx, cy, &sx, &sy);
    fill_circle(cv->target, (int)sx, (int)sy, (int)radius, cv->fill_color);
}

void canvas_draw_ellipse(canvas* cv, float cx, float cy, float rx, float ry) {
    if (!cv || !cv->target) return;
    float sx, sy; world_to_screen(cv, cx, cy, &sx, &sy);
    draw_ellipse(cv->target, (int)sx, (int)sy, (int)rx, (int)ry, cv->stroke_color);
}

void canvas_fill_ellipse(canvas* cv, float cx, float cy, float rx, float ry) {
    if (!cv || !cv->target) return;
    float sx, sy; world_to_screen(cv, cx, cy, &sx, &sy);
    fill_ellipse(cv->target, (int)sx, (int)sy, (int)rx, (int)ry, cv->fill_color);
}

void canvas_draw_triangle(canvas* cv, float x1, float y1, float x2, float y2, float x3, float y3) {
    if (!cv || !cv->target) return;
    float sx1, sy1, sx2, sy2, sx3, sy3;
    world_to_screen(cv, x1, y1, &sx1, &sy1);
    world_to_screen(cv, x2, y2, &sx2, &sy2);
    world_to_screen(cv, x3, y3, &sx3, &sy3);
    draw_triangle(cv->target, (int)sx1, (int)sy1, (int)sx2, (int)sy2, (int)sx3, (int)sy3, cv->stroke_color);
}

void canvas_fill_triangle(canvas* cv, float x1, float y1, float x2, float y2, float x3, float y3) {
    if (!cv || !cv->target) return;
    float sx1, sy1, sx2, sy2, sx3, sy3;
    world_to_screen(cv, x1, y1, &sx1, &sy1);
    world_to_screen(cv, x2, y2, &sx2, &sy2);
    world_to_screen(cv, x3, y3, &sx3, &sy3);
    fill_triangle(cv->target, (int)sx1, (int)sy1, (int)sx2, (int)sy2, (int)sx3, (int)sy3, cv->fill_color);
}

// ============================================================================
// Canvas Image Drawing
// ============================================================================

void canvas_draw_image(canvas* cv, const image* img, float x, float y) {
    if (!cv || !cv->target || !img) return;
    float sx, sy; world_to_screen(cv, x, y, &sx, &sy);
    image_paste(cv->target, (int)sx, (int)sy, img);
}

void canvas_draw_image_scaled(canvas* cv, const image* img, float x, float y, float scale) {
    if (!cv || !cv->target || !img) return;
    image* scaled = image_scale(img, scale, scale);
    if (scaled) {
        canvas_draw_image(cv, scaled, x, y);
        image_free(scaled);
    }
}

// ============================================================================
// Drawing Primitives - Standalone Functions
// ============================================================================

void draw_line(image* img, int x1, int y1, int x2, int y2, color c) {
    if (!image_valid(img)) return;
    
    int dx = abs(x2 - x1);
    int dy = abs(y2 - y1);
    int sx = x1 < x2 ? 1 : -1;
    int sy = y1 < y2 ? 1 : -1;
    int err = dx - dy;
    
    while (1) {
        if (x1 >= 0 && y1 >= 0 && x1 < img->width && y1 < img->height) {
            image_set_pixel(img, x1, y1, c);
        }
        if (x1 == x2 && y1 == y2) break;
        int e2 = 2 * err;
        if (e2 > -dy) { err -= dy; x1 += sx; }
        if (e2 < dx) { err += dx; y1 += sy; }
    }
}

void draw_rect(image* img, int x, int y, int w, int h, color c) {
    if (!image_valid(img) || w <= 0 || h <= 0) return;
    draw_line(img, x, y, x + w - 1, y, c);
    draw_line(img, x + w - 1, y, x + w - 1, y + h - 1, c);
    draw_line(img, x + w - 1, y + h - 1, x, y + h - 1, c);
    draw_line(img, x, y + h - 1, x, y, c);
}

void fill_rect(image* img, int x, int y, int w, int h, color c) {
    image_fill_rect(img, x, y, w, h, c);
}

void draw_circle(image* img, int cx, int cy, int radius, color c) {
    if (!image_valid(img) || radius < 0) return;
    
    int x = radius, y = 0;
    int err = 0;
    
    while (x >= y) {
        if (cx + x >= 0 && cx + x < img->width && cy + y >= 0 && cy + y < img->height)
            image_set_pixel(img, cx + x, cy + y, c);
        if (cx + y >= 0 && cx + y < img->width && cy + x >= 0 && cy + x < img->height)
            image_set_pixel(img, cx + y, cy + x, c);
        if (cx - y >= 0 && cx - y < img->width && cy + x >= 0 && cy + x < img->height)
            image_set_pixel(img, cx - y, cy + x, c);
        if (cx - x >= 0 && cx - x < img->width && cy + y >= 0 && cy + y < img->height)
            image_set_pixel(img, cx - x, cy + y, c);
        if (cx - x >= 0 && cx - x < img->width && cy - y >= 0 && cy - y < img->height)
            image_set_pixel(img, cx - x, cy - y, c);
        if (cx - y >= 0 && cx - y < img->width && cy - x >= 0 && cy - x < img->height)
            image_set_pixel(img, cx - y, cy - x, c);
        if (cx + y >= 0 && cx + y < img->width && cy - x >= 0 && cy - x < img->height)
            image_set_pixel(img, cx + y, cy - x, c);
        if (cx + x >= 0 && cx + x < img->width && cy - y >= 0 && cy - y < img->height)
            image_set_pixel(img, cx + x, cy - y, c);
        
        y++;
        err += 1 + 2 * y;
        if (2 * (err - x) + 1 > 0) {
            x--;
            err += 1 - 2 * x;
        }
    }
}

void fill_circle(image* img, int cx, int cy, int radius, color c) {
    if (!image_valid(img) || radius < 0) return;
    
    for (int y = -radius; y <= radius; y++) {
        for (int x = -radius; x <= radius; x++) {
            if (x * x + y * y <= radius * radius) {
                int px = cx + x;
                int py = cy + y;
                if (px >= 0 && py >= 0 && px < img->width && py < img->height) {
                    image_set_pixel(img, px, py, c);
                }
            }
        }
    }
}

void draw_ellipse(image* img, int cx, int cy, int rx, int ry, color c) {
    if (!image_valid(img) || rx < 0 || ry < 0) return;
    
    float dx, dy, d1, d2, x, y;
    x = 0; y = ry;
    
    d1 = ry * ry - rx * rx * ry + 0.25f * rx * rx;
    dx = 2 * ry * ry * x;
    dy = 2 * rx * rx * y;
    
    while (dx < dy) {
        if (cx + (int)x >= 0 && cx + (int)x < img->width && cy + (int)y >= 0 && cy + (int)y < img->height)
            image_set_pixel(img, cx + (int)x, cy + (int)y, c);
        if (cx - (int)x >= 0 && cx - (int)x < img->width && cy + (int)y >= 0 && cy + (int)y < img->height)
            image_set_pixel(img, cx - (int)x, cy + (int)y, c);
        if (cx + (int)x >= 0 && cx + (int)x < img->width && cy - (int)y >= 0 && cy - (int)y < img->height)
            image_set_pixel(img, cx + (int)x, cy - (int)y, c);
        if (cx - (int)x >= 0 && cx - (int)x < img->width && cy - (int)y >= 0 && cy - (int)y < img->height)
            image_set_pixel(img, cx - (int)x, cy - (int)y, c);
        
        if (d1 < 0) {
            x++;
            dx = dx + 2 * ry * ry;
            d1 = d1 + dx + ry * ry;
        } else {
            x++; y--;
            dx = dx + 2 * ry * ry;
            dy = dy - 2 * rx * rx;
            d1 = d1 + dx - dy + ry * ry;
        }
    }
    
    d2 = ry * ry * ((x + 0.5f) * (x + 0.5f)) + rx * rx * ((y - 1) * (y - 1)) - rx * rx * ry * ry;
    
    while (y >= 0) {
        if (cx + (int)x >= 0 && cx + (int)x < img->width && cy + (int)y >= 0 && cy + (int)y < img->height)
            image_set_pixel(img, cx + (int)x, cy + (int)y, c);
        if (cx - (int)x >= 0 && cx - (int)x < img->width && cy + (int)y >= 0 && cy + (int)y < img->height)
            image_set_pixel(img, cx - (int)x, cy + (int)y, c);
        if (cx + (int)x >= 0 && cx + (int)x < img->width && cy - (int)y >= 0 && cy - (int)y < img->height)
            image_set_pixel(img, cx + (int)x, cy - (int)y, c);
        if (cx - (int)x >= 0 && cx - (int)x < img->width && cy - (int)y >= 0 && cy - (int)y < img->height)
            image_set_pixel(img, cx - (int)x, cy - (int)y, c);
        
        if (d2 > 0) {
            y--;
            dy = dy - 2 * rx * rx;
            d2 = d2 + rx * rx - dy;
        } else {
            y--; x++;
            dx = dx + 2 * ry * ry;
            dy = dy - 2 * rx * rx;
            d2 = d2 + dx - dy + rx * rx;
        }
    }
}

void fill_ellipse(image* img, int cx, int cy, int rx, int ry, color c) {
    if (!image_valid(img) || rx < 0 || ry < 0) return;
    
    for (int y = -ry; y <= ry; y++) {
        for (int x = -rx; x <= rx; x++) {
            if ((x * x * ry * ry + y * y * rx * rx) <= rx * rx * ry * ry) {
                int px = cx + x;
                int py = cy + y;
                if (px >= 0 && py >= 0 && px < img->width && py < img->height) {
                    image_set_pixel(img, px, py, c);
                }
            }
        }
    }
}

void draw_triangle(image* img, int x1, int y1, int x2, int y2, int x3, int y3, color c) {
    draw_line(img, x1, y1, x2, y2, c);
    draw_line(img, x2, y2, x3, y3, c);
    draw_line(img, x3, y3, x1, y1, c);
}

void fill_triangle(image* img, int x1, int y1, int x2, int y2, int x3, int y3, color c) {
    if (!image_valid(img)) return;
    
    // Sort vertices by y
    if (y1 > y2) { int t; t=y1; y1=y2; y2=t; t=x1; x1=x2; x2=t; }
    if (y1 > y3) { int t; t=y1; y1=y3; y3=t; t=x1; x1=x3; x3=t; }
    if (y2 > y3) { int t; t=y2; y2=y3; y3=t; t=x2; x2=x3; x3=t; }
    
    if (y1 == y3) { // Degenerate
        draw_line(img, x1, y1, x2, y2, c);
        draw_line(img, x2, y2, x3, y3, c);
        return;
    }
    
    int total_height = y3 - y1;
    
    for (int i = 0; i < total_height; i++) {
        bool second_half = i > y2 - y1 || y2 == y1;
        int segment_height = second_half ? y3 - y2 : y2 - y1;
        if (segment_height == 0) continue;
        
        float alpha = (float)i / total_height;
        float beta = second_half ? (float)(i - (y2 - y1)) / segment_height : (float)i / segment_height;
        
        int ax = x1 + (int)((x3 - x1) * alpha);
        int ay = y1 + i;
        int bx = second_half ? x2 + (int)((x3 - x2) * beta) : x1 + (int)((x2 - x1) * beta);
        
        if (ax > bx) { int t = ax; ax = bx; bx = t; }
        
        for (int x = ax; x <= bx; x++) {
            if (x >= 0 && ay >= 0 && x < img->width && ay < img->height) {
                image_set_pixel(img, x, ay, c);
            }
        }
    }
}

void draw_polygon(image* img, const int* x, const int* y, int count, color c) {
    if (!image_valid(img) || !x || !y || count < 2) return;
    for (int i = 0; i < count; i++) {
        int next = (i + 1) % count;
        draw_line(img, x[i], y[i], x[next], y[next], c);
    }
}

void fill_polygon(image* img, const int* x, const int* y, int count, color c) {
    if (!image_valid(img) || !x || !y || count < 3) return;
    
    // Find bounding box
    int miny = y[0], maxy = y[0];
    for (int i = 1; i < count; i++) {
        if (y[i] < miny) miny = y[i];
        if (y[i] > maxy) maxy = y[i];
    }
    
    // Scanline fill
    for (int scany = miny; scany <= maxy; scany++) {
        int nodes = 0;
        int node_x[64]; // Max 64 nodes per scanline
        
        for (int i = 0; i < count; i++) {
            int j = (i + 1) % count;
            if ((y[i] <= scany && y[j] > scany) || (y[j] <= scany && y[i] > scany)) {
                float slope = (float)(x[j] - x[i]) / (y[j] - y[i]);
                node_x[nodes++] = (int)(x[i] + slope * (scany - y[i]));
                if (nodes >= 64) break;
            }
        }
        
        // Sort nodes
        for (int i = 1; i < nodes; i++) {
            int temp = node_x[i];
            int j = i - 1;
            while (j >= 0 && node_x[j] > temp) {
                node_x[j + 1] = node_x[j];
                j--;
            }
            node_x[j + 1] = temp;
        }
        
        // Fill between pairs
        for (int i = 0; i < nodes - 1; i += 2) {
            if (node_x[i + 1] > node_x[i]) {
                for (int x = node_x[i]; x <= node_x[i + 1]; x++) {
                    if (x >= 0 && scany >= 0 && x < img->width && scany < img->height) {
                        image_set_pixel(img, x, scany, c);
                    }
                }
            }
        }
    }
}

void draw_arc(image* img, int cx, int cy, int radius, float start_angle, float end_angle, color c) {
    if (!image_valid(img) || radius < 0) return;
    
    float step = 1.0f / radius;
    for (float a = start_angle; a < end_angle; a += step) {
        int x = cx + (int)(radius * cosf(a));
        int y = cy + (int)(radius * sinf(a));
        if (x >= 0 && y >= 0 && x < img->width && y < img->height) {
            image_set_pixel(img, x, y, c);
        }
    }
}

void fill_arc(image* img, int cx, int cy, int radius, float start_angle, float end_angle, color c) {
    if (!image_valid(img) || radius < 0) return;
    
    float step = 1.0f / radius;
    for (float a = start_angle; a < end_angle; a += step) {
        int ex = cx + (int)(radius * cosf(a));
        int ey = cy + (int)(radius * sinf(a));
        draw_line(img, cx, cy, ex, ey, c);
    }
}

void draw_bezier(image* img, int x1, int y1, int x2, int y2, int x3, int y3, int x4, int y4, color c) {
    if (!image_valid(img)) return;
    
    int steps = 20;
    float prev_x = (float)x1, prev_y = (float)y1;
    
    for (int i = 1; i <= steps; i++) {
        float t = (float)i / steps;
        float t2 = t * t;
        float t3 = t2 * t;
        float mt = 1 - t;
        float mt2 = mt * mt;
        float mt3 = mt2 * mt;
        
        float x = mt3 * x1 + 3 * mt2 * t * x2 + 3 * mt * t2 * x3 + t3 * x4;
        float y = mt3 * y1 + 3 * mt2 * t * y2 + 3 * mt * t2 * y3 + t3 * y4;
        
        draw_line(img, (int)prev_x, (int)prev_y, (int)x, (int)y, c);
        prev_x = x; prev_y = y;
    }
}

void draw_bezier_quad(image* img, int x1, int y1, int x2, int y2, int x3, int y3, color c) {
    if (!image_valid(img)) return;
    
    int steps = 20;
    float prev_x = (float)x1, prev_y = (float)y1;
    
    for (int i = 1; i <= steps; i++) {
        float t = (float)i / steps;
        float mt = 1 - t;
        
        float x = mt * mt * x1 + 2 * mt * t * x2 + t * t * x3;
        float y = mt * mt * y1 + 2 * mt * t * y2 + t * t * y3;
        
        draw_line(img, (int)prev_x, (int)prev_y, (int)x, (int)y, c);
        prev_x = x; prev_y = y;
    }
}

void draw_spline(image* img, const int* x, const int* y, int count, color c) {
    if (!image_valid(img) || !x || !y || count < 2) return;
    if (count == 2) { draw_line(img, x[0], y[0], x[1], y[1], c); return; }
    
    for (int i = 0; i < count - 1; i++) {
        int x1 = x[i], y1 = y[i];
        int x2 = x[i + 1], y2 = y[i + 1];
        
        if (i > 0) {
            x1 = (x[i] + x[i + 1]) / 2;
            y1 = (y[i] + y[i + 1]) / 2;
        }
        if (i < count - 2) {
            x2 = (x[i + 1] + x[i + 2]) / 2;
            y2 = (y[i + 1] + y[i + 2]) / 2;
        }
        
        draw_bezier_quad(img, x1, y1, x[i + 1], y[i + 1], x2, y2, c);
    }
}

// ============================================================================
// Bitmap Font
// ============================================================================

bitmap_font* bitmap_font_create(image* atlas, int char_w, int char_h, int columns, int first_char, int last_char) {
    if (!atlas || char_w <= 0 || char_h <= 0 || columns <= 0) return NULL;
    
    bitmap_font* font = (bitmap_font*)malloc(sizeof(bitmap_font));
    if (!font) return NULL;
    
    font->atlas = atlas;
    font->char_width = char_w;
    font->char_height = char_h;
    font->columns = columns;
    font->first_char = first_char;
    font->last_char = last_char;
    font->spacing = 1;
    
    return font;
}

bitmap_font* bitmap_font_load(const char* filepath, int char_w, int char_h, int columns, int first_char, int last_char) {
    // Note: This requires image_load which would need file format support
    // For now, return NULL - user should load image separately
    return NULL;
}

void bitmap_font_free(bitmap_font* font) {
    if (font) free(font);
}

void bitmap_font_set_spacing(bitmap_font* font, int spacing) {
    if (font) font->spacing = spacing;
}

void bitmap_font_draw(const bitmap_font* font, image* target, int x, int y, const char* text) {
    bitmap_font_draw_color(font, target, x, y, text, COLOR_WHITE);
}

void bitmap_font_draw_color(const bitmap_font* font, image* target, int x, int y, const char* text, color c) {
    if (!font || !font->atlas || !target || !text) return;
    
    int cur_x = x;
    int cur_y = y;
    
    for (const char* ch = text; *ch; ch++) {
        if (*ch == '\n') {
            cur_x = x;
            cur_y += font->char_height;
            continue;
        }
        if (*ch == '\r') continue;
        
        int char_idx = *ch - font->first_char;
        if (char_idx < 0 || char_idx > font->last_char - font->first_char) {
            cur_x += font->char_width + font->spacing;
            continue;
        }
        
        int col = char_idx % font->columns;
        int row = char_idx / font->columns;
        int src_x = col * font->char_width;
        int src_y = row * font->char_height;
        
        // Copy character from atlas to target
        for (int py = 0; py < font->char_height; py++) {
            for (int px = 0; px < font->char_width; px++) {
                color pixel = image_get_pixel(font->atlas, src_x + px, src_y + py);
                if (pixel.a > 0) {
                    pixel = color_tint(pixel, c);
                    int dst_x = cur_x + px;
                    int dst_y = cur_y + py;
                    if (dst_x >= 0 && dst_y >= 0 && dst_x < target->width && dst_y < target->height) {
                        color dst_c = image_get_pixel(target, dst_x, dst_y);
                        image_set_pixel(target, dst_x, dst_y, color_blend(pixel, dst_c));
                    }
                }
            }
        }
        
        cur_x += font->char_width + font->spacing;
    }
}

void bitmap_font_draw_scaled(const bitmap_font* font, image* target, float x, float y, const char* text, float scale) {
    if (!font || !target || !text || scale <= 0) return;
    
    int cur_x = (int)x;
    int cur_y = (int)y;
    int scaled_w = (int)(font->char_width * scale);
    int scaled_h = (int)(font->char_height * scale);
    int scaled_spacing = (int)(font->spacing * scale);
    
    for (const char* ch = text; *ch; ch++) {
        if (*ch == '\n') {
            cur_x = (int)x;
            cur_y += scaled_h;
            continue;
        }
        if (*ch == '\r') continue;
        
        int char_idx = *ch - font->first_char;
        if (char_idx < 0 || char_idx > font->last_char - font->first_char) {
            cur_x += scaled_w + scaled_spacing;
            continue;
        }
        
        int col = char_idx % font->columns;
        int row = char_idx / font->columns;
        int src_x = col * font->char_width;
        int src_y = row * font->char_height;
        
        for (int py = 0; py < scaled_h; py++) {
            for (int px = 0; px < scaled_w; px++) {
                int src_px = (int)(px / scale);
                int src_py = (int)(py / scale);
                color pixel = image_get_pixel(font->atlas, src_x + src_px, src_y + src_py);
                if (pixel.a > 0) {
                    int dst_x = cur_x + px;
                    int dst_y = cur_y + py;
                    if (dst_x >= 0 && dst_y >= 0 && dst_x < target->width && dst_y < target->height) {
                        color dst_c = image_get_pixel(target, dst_x, dst_y);
                        image_set_pixel(target, dst_x, dst_y, color_blend(pixel, dst_c));
                    }
                }
            }
        }
        
        cur_x += scaled_w + scaled_spacing;
    }
}

int bitmap_font_measure_width(const bitmap_font* font, const char* text) {
    if (!font || !text) return 0;
    
    int max_width = 0, cur_width = 0;
    
    for (const char* ch = text; *ch; ch++) {
        if (*ch == '\n') {
            if (cur_width > max_width) max_width = cur_width;
            cur_width = 0;
        } else if (*ch != '\r') {
            cur_width += font->char_width + font->spacing;
        }
    }
    
    if (cur_width > max_width) max_width = cur_width;
    return max_width > 0 ? max_width - font->spacing : 0;
}

int bitmap_font_measure_height(const bitmap_font* font, const char* text) {
    if (!font || !text) return 0;
    
    int lines = 1;
    for (const char* ch = text; *ch; ch++) {
        if (*ch == '\n') lines++;
    }
    return lines * font->char_height;
}

// ============================================================================
// Texture Atlas
// ============================================================================

texture_atlas* texture_atlas_create(int width, int height, int padding) {
    image* img = image_create(width, height, 4);
    if (!img) return NULL;
    
    texture_atlas* ta = (texture_atlas*)malloc(sizeof(texture_atlas));
    if (!ta) { image_free(img); return NULL; }
    
    ta->atlas = img;
    ta->regions = NULL;
    ta->region_count = 0;
    ta->padding = padding;
    
    return ta;
}

texture_atlas* texture_atlas_from_images(image** images, int count, int padding) {
    if (!images || count <= 0) return NULL;
    
    // Calculate required size
    int total_area = 0;
    int max_w = 0, max_h = 0;
    for (int i = 0; i < count; i++) {
        if (images[i]) {
            total_area += (images[i]->width + padding) * (images[i]->height + padding);
            if (images[i]->width > max_w) max_w = images[i]->width;
            if (images[i]->height > max_h) max_h = images[i]->height;
        }
    }
    
    int size = next_power_of_two((int)sqrtf((float)total_area));
    if (size < max_w + padding) size = next_power_of_two(max_w + padding);
    if (size < max_h + padding) size = next_power_of_two(max_h + padding);
    
    texture_atlas* ta = texture_atlas_create(size, size, padding);
    if (!ta) return NULL;
    
    texture_atlas_pack(ta, images, count);
    return ta;
}

void texture_atlas_free(texture_atlas* ta) {
    if (!ta) return;
    if (ta->atlas) image_free(ta->atlas);
    if (ta->regions) free(ta->regions);
    free(ta);
}

int texture_atlas_add_region(texture_atlas* ta, int x, int y, int w, int h) {
    if (!ta) return -1;
    
    int new_count = ta->region_count + 1;
    int* new_regions = (int*)realloc(ta->regions, new_count * 4 * sizeof(int));
    if (!new_regions) return -1;
    
    ta->regions = new_regions;
    int idx = ta->region_count * 4;
    ta->regions[idx] = x;
    ta->regions[idx + 1] = y;
    ta->regions[idx + 2] = w;
    ta->regions[idx + 3] = h;
    ta->region_count++;
    
    return ta->region_count - 1;
}

int texture_atlas_add_image(texture_atlas* ta, const image* img) {
    if (!ta || !img) return -1;
    
    // Simple row packing
    int x = ta->padding, y = ta->padding;
    int row_height = 0;
    
    for (int i = 0; i < ta->region_count; i++) {
        int rx = ta->regions[i * 4];
        int ry = ta->regions[i * 4 + 1];
        int rw = ta->regions[i * 4 + 2];
        int rh = ta->regions[i * 4 + 3];
        
        if (ry == y) {
            x = rx + rw + ta->padding;
            if (rh > row_height) row_height = rh;
        }
        if (ry > y || (ry == y && rx + rw + ta->padding > x)) {
            if (x + img->width + ta->padding > ta->atlas->width) {
                y += row_height + ta->padding;
                x = ta->padding;
                row_height = 0;
            }
        }
    }
    
    if (x + img->width > ta->atlas->width) {
        x = ta->padding;
        y += row_height + ta->padding;
        row_height = 0;
    }
    
    if (y + img->height > ta->atlas->height) return -1;
    
    image_copy(ta->atlas, x, y, img, 0, 0, img->width, img->height);
    return texture_atlas_add_region(ta, x, y, img->width, img->height);
}

int texture_atlas_pack(texture_atlas* ta, image** images, int count) {
    if (!ta || !images || count <= 0) return 0;
    
    int added = 0;
    for (int i = 0; i < count; i++) {
        if (images[i] && texture_atlas_add_image(ta, images[i]) >= 0) {
            added++;
        }
    }
    return added;
}

cortex_rect texture_atlas_get_region(const texture_atlas* ta, int region) {
    if (!ta || region < 0 || region >= ta->region_count) {
        return (cortex_rect){0, 0, 0, 0};
    }
    int idx = region * 4;
    return (cortex_rect){(float)ta->regions[idx], (float)ta->regions[idx + 1], 
                         (float)ta->regions[idx + 2], (float)ta->regions[idx + 3]};
}

void texture_atlas_draw_region(const texture_atlas* ta, int region, image* target, int x, int y) {
    if (!ta || !ta->atlas || !target || region < 0 || region >= ta->region_count) return;
    
    int idx = region * 4;
    int rx = ta->regions[idx];
    int ry = ta->regions[idx + 1];
    int rw = ta->regions[idx + 2];
    int rh = ta->regions[idx + 3];
    
    image_copy(target, x, y, ta->atlas, rx, ry, rw, rh);
}

// ============================================================================
// Blend Modes
// ============================================================================

color blend_colors(color src, color dst, blend_mode mode) {
    if (src.a == 0) return dst;
    if (dst.a == 0) return src;
    
    colorf s = color_to_colorf(src);
    colorf d = color_to_colorf(dst);
    colorf r = {0, 0, 0, s.a + d.a * (1 - s.a)};
    
    if (r.a <= 0) return COLOR_CLEAR;
    
    switch (mode) {
        case BLEND_NORMAL:
            r.r = (s.r * s.a + d.r * d.a * (1 - s.a)) / r.a;
            r.g = (s.g * s.a + d.g * d.a * (1 - s.a)) / r.a;
            r.b = (s.b * s.a + d.b * d.a * (1 - s.a)) / r.a;
            break;
            
        case BLEND_MULTIPLY:
            r.r = s.r * d.r;
            r.g = s.g * d.g;
            r.b = s.b * d.b;
            break;
            
        case BLEND_SCREEN:
            r.r = 1 - (1 - s.r) * (1 - d.r);
            r.g = 1 - (1 - s.g) * (1 - d.g);
            r.b = 1 - (1 - s.b) * (1 - d.b);
            break;
            
        case BLEND_OVERLAY:
            r.r = d.r < 0.5f ? 2 * s.r * d.r : 1 - 2 * (1 - s.r) * (1 - d.r);
            r.g = d.g < 0.5f ? 2 * s.g * d.g : 1 - 2 * (1 - s.g) * (1 - d.g);
            r.b = d.b < 0.5f ? 2 * s.b * d.b : 1 - 2 * (1 - s.b) * (1 - d.b);
            break;
            
        case BLEND_DARKEN:
            r.r = fminf(s.r, d.r);
            r.g = fminf(s.g, d.g);
            r.b = fminf(s.b, d.b);
            break;
            
        case BLEND_LIGHTEN:
            r.r = fmaxf(s.r, d.r);
            r.g = fmaxf(s.g, d.g);
            r.b = fmaxf(s.b, d.b);
            break;
            
        case BLEND_COLOR_DODGE:
            r.r = d.r < 0.001f ? d.r : fminf(1, s.r / (1 - d.r));
            r.g = d.g < 0.001f ? d.g : fminf(1, s.g / (1 - d.g));
            r.b = d.b < 0.001f ? d.b : fminf(1, s.b / (1 - d.b));
            break;
            
        case BLEND_COLOR_BURN:
            r.r = d.r > 0.999f ? d.r : 1 - fminf(1, (1 - s.r) / d.r);
            r.g = d.g > 0.999f ? d.g : 1 - fminf(1, (1 - s.g) / d.g);
            r.b = d.b > 0.999f ? d.b : 1 - fminf(1, (1 - s.b) / d.b);
            break;
            
        case BLEND_HARD_LIGHT:
            r.r = s.r < 0.5f ? 2 * s.r * d.r : 1 - 2 * (1 - s.r) * (1 - d.r);
            r.g = s.g < 0.5f ? 2 * s.g * d.g : 1 - 2 * (1 - s.g) * (1 - d.g);
            r.b = s.b < 0.5f ? 2 * s.b * d.b : 1 - 2 * (1 - s.b) * (1 - d.b);
            break;
            
        case BLEND_SOFT_LIGHT:
            r.r = d.r < 0.5f ? d.r * (1 + 2 * s.r - 1) : d.r + (2 * s.r - 1) * (sqrtf(d.r) - d.r);
            r.g = d.g < 0.5f ? d.g * (1 + 2 * s.g - 1) : d.g + (2 * s.g - 1) * (sqrtf(d.g) - d.g);
            r.b = d.b < 0.5f ? d.b * (1 + 2 * s.b - 1) : d.b + (2 * s.b - 1) * (sqrtf(d.b) - d.b);
            break;
            
        case BLEND_DIFFERENCE:
            r.r = fabsf(s.r - d.r);
            r.g = fabsf(s.g - d.g);
            r.b = fabsf(s.b - d.b);
            break;
            
        case BLEND_EXCLUSION:
            r.r = s.r + d.r - 2 * s.r * d.r;
            r.g = s.g + d.g - 2 * s.g * d.g;
            r.b = s.b + d.b - 2 * s.b * d.b;
            break;
            
        case BLEND_ADD:
            r.r = fminf(1, s.r + d.r);
            r.g = fminf(1, s.g + d.g);
            r.b = fminf(1, s.b + d.b);
            break;
            
        case BLEND_SUBTRACT:
            r.r = fmaxf(0, s.r - d.r);
            r.g = fmaxf(0, s.g - d.g);
            r.b = fmaxf(0, s.b - d.b);
            break;
            
        case BLEND_DIVIDE:
            r.r = d.r > 0.001f ? s.r / d.r : 1;
            r.g = d.g > 0.001f ? s.g / d.g : 1;
            r.b = d.b > 0.001f ? s.b / d.b : 1;
            break;
            
        default:
            r = s;
            break;
    }
    
    return colorf_to_color(r);
}

void image_blend_mode(image* dst, int x, int y, const image* src, blend_mode mode) {
    if (!image_valid(dst) || !image_valid(src)) return;
    
    for (int sy = 0; sy < src->height; sy++) {
        for (int sx = 0; sx < src->width; sx++) {
            int dx = x + sx;
            int dy = y + sy;
            
            if (dx >= 0 && dy >= 0 && dx < dst->width && dy < dst->height) {
                color src_c = image_get_pixel(src, sx, sy);
                color dst_c = image_get_pixel(dst, dx, dy);
                color result = blend_colors(src_c, dst_c, mode);
                image_set_pixel(dst, dx, dy, result);
            }
        }
    }
}

void canvas_set_blend_mode(canvas* cv, blend_mode mode) {
    // Note: This would require storing blend mode in canvas struct
    // For now, this is a placeholder
    (void)cv; (void)mode;
}

// ============================================================================
// Additional Utility Functions
// ============================================================================

color trilinear_interpolate(const image* img, float x, float y, float level) {
    // For now, just use bilinear - level would be for mipmaps
    (void)level;
    return bilinear_interpolate(img, x, y);
}