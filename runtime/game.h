#ifndef CORTEX_RUNTIME_GAME_H
#define CORTEX_RUNTIME_GAME_H

#include "core.h"
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

// ============================================================================
// Easing Functions (t in [0,1])
// ============================================================================

float ease_linear(float t);
float ease_quad_in(float t);
float ease_quad_out(float t);
float ease_quad_in_out(float t);
float ease_cubic_in(float t);
float ease_cubic_out(float t);
float ease_cubic_in_out(float t);
float ease_quart_in(float t);
float ease_quart_out(float t);
float ease_quart_in_out(float t);
float ease_quint_in(float t);
float ease_quint_out(float t);
float ease_quint_in_out(float t);
float ease_sine_in(float t);
float ease_sine_out(float t);
float ease_sine_in_out(float t);
float ease_expo_in(float t);
float ease_expo_out(float t);
float ease_expo_in_out(float t);
float ease_circ_in(float t);
float ease_circ_out(float t);
float ease_circ_in_out(float t);
float ease_back_in(float t);
float ease_back_out(float t);
float ease_back_in_out(float t);
float ease_elastic_in(float t);
float ease_elastic_out(float t);
float ease_elastic_in_out(float t);
float ease_bounce_in(float t);
float ease_bounce_out(float t);
float ease_bounce_in_out(float t);

// ============================================================================
// Vector Interpolation
// ============================================================================

vec2 vec2_lerp(vec2 a, vec2 b, float t);
vec3 vec3_lerp(vec3 a, vec3 b, float t);
vec2 vec2_smoothstep(vec2 a, vec2 b, float t);
vec3 vec3_smoothstep(vec3 a, vec3 b, float t);

// ============================================================================
// Rectangle / AABB
// ============================================================================

typedef struct { float x, y, w, h; } cortex_rect;

bool rect_contains_point(cortex_rect r, float px, float py);
bool rect_overlaps(cortex_rect a, cortex_rect b);
cortex_rect rect_from_center(float cx, float cy, float w, float h);
cortex_rect rect_from_corners(float x1, float y1, float x2, float y2);
cortex_rect rect_union(cortex_rect a, cortex_rect b);
cortex_rect rect_intersection(cortex_rect a, cortex_rect b);
float rect_area(cortex_rect r);
vec2 rect_center(cortex_rect r);
vec2 rect_min(cortex_rect r);
vec2 rect_max(cortex_rect r);
cortex_rect rect_expand(cortex_rect r, float amount);
cortex_rect rect_translate(cortex_rect r, float dx, float dy);

// ============================================================================
// Circle / Sphere
// ============================================================================

typedef struct { float x, y, r; } cortex_circle;
typedef struct { float x, y, z, r; } cortex_sphere;

bool circle_contains_point(cortex_circle c, float px, float py);
bool circle_overlaps_circle(cortex_circle a, cortex_circle b);
bool circle_overlaps_rect(cortex_circle c, cortex_rect r);
float circle_area(cortex_circle c);
float circle_circumference(cortex_circle c);
cortex_circle circle_from_center(float cx, float cy, float r);

bool sphere_contains_point(cortex_sphere s, float px, float py, float pz);
bool sphere_overlaps_sphere(cortex_sphere a, cortex_sphere b);
float sphere_volume(cortex_sphere s);
float sphere_surface_area(cortex_sphere s);

// ============================================================================
// Angle Helpers (radians)
// ============================================================================

float vec2_angle(vec2 v);
vec2 vec2_from_angle(float radians);
float angle_normalize(float radians);
float angle_lerp(float a, float b, float t);
float angle_distance(float a, float b);
float deg_to_rad(float degrees);
float rad_to_deg(float radians);

// ============================================================================
// Collision Detection
// ============================================================================

bool point_in_rect(float px, float py, cortex_rect r);
bool point_in_circle(float px, float py, cortex_circle c);
bool point_in_sphere(float px, float py, float pz, cortex_sphere s);
bool rect_collides_rect(cortex_rect a, cortex_rect b);
bool circle_collides_circle(cortex_circle a, cortex_circle b);
bool sphere_collides_sphere(cortex_sphere a, cortex_sphere b);
bool line_intersects_rect(float x1, float y1, float x2, float y2, cortex_rect r);
bool line_intersects_circle(float x1, float y1, float x2, float y2, cortex_circle c);

// Raycasting
typedef struct { vec2 origin; vec2 direction; } ray2d;
typedef struct { vec3 origin; vec3 direction; } ray3d;
typedef struct { bool hit; float distance; vec2 point; vec2 normal; } ray2d_hit;
typedef struct { bool hit; float distance; vec3 point; vec3 normal; } ray3d_hit;

ray2d_hit ray2d_intersect_rect(ray2d ray, cortex_rect r);
ray2d_hit ray2d_intersect_circle(ray2d ray, cortex_circle c);

// ============================================================================
// Physics / Motion
// ============================================================================

typedef struct {
    vec2 position;
    vec2 velocity;
    vec2 acceleration;
    float mass;
    float drag;
    float restitution; // bounciness 0-1
} physics_body_2d;

void physics_body_2d_update(physics_body_2d* body, float dt);
void physics_body_2d_apply_force(physics_body_2d* body, vec2 force);
void physics_body_2d_apply_impulse(physics_body_2d* body, vec2 impulse);
void physics_body_2d_apply_gravity(physics_body_2d* body, float g);

// Simple collision response
void physics_resolve_rect_rect(physics_body_2d* a, cortex_rect ra, physics_body_2d* b, cortex_rect rb);
void physics_resolve_circle_circle(physics_body_2d* a, cortex_circle ca, physics_body_2d* b, cortex_circle cb);

// ============================================================================
// Particle System
// ============================================================================

typedef struct {
    vec2 position;
    vec2 velocity;
    vec4 color;         // r, g, b, a
    float size;
    float lifetime;
    float age;
    bool active;
} particle;

typedef struct {
    particle* particles;
    int count;
    int capacity;
    
    // Emitter settings
    vec2 emitter_position;
    vec2 emitter_velocity_min;
    vec2 emitter_velocity_max;
    vec4 color_start;
    vec4 color_end;
    float size_start;
    float size_end;
    float lifetime_min;
    float lifetime_max;
    float emission_rate;
    float emission_accumulator;
} particle_system;

particle_system* particle_system_create(int capacity);
void particle_system_free(particle_system* ps);
void particle_system_emit(particle_system* ps, int count);
void particle_system_update(particle_system* ps, float dt);
void particle_system_render(particle_system* ps); // For future rendering integration

// ============================================================================
// Timer / Animation
// ============================================================================

typedef struct {
    float duration;
    float elapsed;
    bool looping;
    bool playing;
    float progress; // 0-1
} animation_timer;

animation_timer animation_timer_create(float duration, bool looping);
void animation_timer_start(animation_timer* timer);
void animation_timer_stop(animation_timer* timer);
void animation_timer_reset(animation_timer* timer);
void animation_timer_update(animation_timer* timer, float dt);
bool animation_timer_is_complete(animation_timer* timer);

// ============================================================================
// Input State
// ============================================================================

typedef struct {
    bool pressed[256];      // Current frame
    bool held[256];         // Held down
    bool released[256];     // Just released
    vec2 mouse_position;
    vec2 mouse_delta;
    bool mouse_buttons[3];  // Left, Middle, Right
} input_state;

void input_update(input_state* state);
void input_key_down(input_state* state, int key);
void input_key_up(input_state* state, int key);
void input_mouse_move(input_state* state, float x, float y);
void input_mouse_button(input_state* state, int button, bool pressed);
bool input_is_pressed(input_state* state, int key);
bool input_is_held(input_state* state, int key);
bool input_is_released(input_state* state, int key);

// ============================================================================
// Game State Management
// ============================================================================

typedef struct game_state game_state;
typedef void (*state_init_fn)(game_state* state);
typedef void (*state_update_fn)(game_state* state, float dt);
typedef void (*state_render_fn)(game_state* state);
typedef void (*state_cleanup_fn)(game_state* state);

struct game_state {
    const char* name;
    state_init_fn init;
    state_update_fn update;
    state_render_fn render;
    state_cleanup_fn cleanup;
    void* data; // User data
    bool initialized;
};

// ============================================================================
// Noise / Random
// ============================================================================

// Perlin noise
float noise_perlin_1d(float x);
float noise_perlin_2d(float x, float y);
float noise_perlin_3d(float x, float y, float z);

// Seeded random
void random_seed(uint32_t seed);
int random_range(int min, int max);
float random_range_float(float min, float max);
vec2 random_point_in_circle(float radius);
vec2 random_point_on_circle(float radius);
vec3 random_point_in_sphere(float radius);
vec3 random_point_on_sphere(float radius);

// ============================================================================
// Utility
// ============================================================================

float approach(float current, float target, float max_delta);
float precise_approach(float current, float target, float max_delta);
vec2 vec2_approach(vec2 current, vec2 target, float max_distance);
vec3 vec3_approach(vec3 current, vec3 target, float max_distance);

// Screen shake
typedef struct {
    vec2 offset;
    float intensity;
    float duration;
    float elapsed;
    float frequency;
} screen_shake;

void screen_shake_start(screen_shake* shake, float intensity, float duration, float frequency);
void screen_shake_update(screen_shake* shake, float dt);
vec2 screen_shake_get_offset(screen_shake* shake);

#ifdef __cplusplus
}
#endif

#endif
