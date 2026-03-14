#include "game.h"
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <float.h>

#ifndef M_PI
#define M_PI 3.14159265358979323846
#endif

// ============================================================================
// Easing Functions
// ============================================================================

float ease_linear(float t) { return t; }

float ease_quad_in(float t) { return t * t; }
float ease_quad_out(float t) { return t * (2.f - t); }
float ease_quad_in_out(float t) { return t < 0.5f ? 2.f * t * t : -1.f + (4.f - 2.f * t) * t; }

float ease_cubic_in(float t) { return t * t * t; }
float ease_cubic_out(float t) { float u = t - 1.f; return u * u * u + 1.f; }
float ease_cubic_in_out(float t) { return t < 0.5f ? 4.f * t * t * t : 1.f + (t - 1.f) * (2.f * (t - 1.f)) * (2.f * (t - 1.f) + 2.f); }

float ease_quart_in(float t) { return t * t * t * t; }
float ease_quart_out(float t) { float u = t - 1.f; return 1.f - u * u * u * u; }
float ease_quart_in_out(float t) { float u = t - 1.f; return t < 0.5f ? 8.f * t * t * t * t : 1.f - 8.f * u * u * u * u; }

float ease_quint_in(float t) { return t * t * t * t * t; }
float ease_quint_out(float t) { float u = t - 1.f; return 1.f + u * u * u * u * u; }
float ease_quint_in_out(float t) { float u = t - 1.f; return t < 0.5f ? 16.f * t * t * t * t * t : 1.f + 16.f * u * u * u * u * u; }

float ease_sine_in(float t) { return 1.f - cosf(t * (float)M_PI / 2.f); }
float ease_sine_out(float t) { return sinf(t * (float)M_PI / 2.f); }
float ease_sine_in_out(float t) { return -0.5f * (cosf((float)M_PI * t) - 1.f); }

float ease_expo_in(float t) { return t == 0.f ? 0.f : powf(2.f, 10.f * (t - 1.f)); }
float ease_expo_out(float t) { return t == 1.f ? 1.f : 1.f - powf(2.f, -10.f * t); }
float ease_expo_in_out(float t) {
    if (t == 0.f) return 0.f;
    if (t == 1.f) return 1.f;
    return t < 0.5f ? 0.5f * powf(2.f, 20.f * t - 10.f) : 1.f - 0.5f * powf(2.f, -20.f * t + 10.f);
}

float ease_circ_in(float t) { return 1.f - sqrtf(1.f - t * t); }
float ease_circ_out(float t) { return sqrtf(1.f - (t - 1.f) * (t - 1.f)); }
float ease_circ_in_out(float t) { return t < 0.5f ? 0.5f * (1.f - sqrtf(1.f - 4.f * t * t)) : 0.5f * (1.f + sqrtf(-3.f + 8.f * t - 4.f * t * t)); }

float ease_back_in(float t) { float s = 1.70158f; return t * t * ((s + 1.f) * t - s); }
float ease_back_out(float t) { float s = 1.70158f; float u = t - 1.f; return 1.f + u * u * ((s + 1.f) * u + s); }
float ease_back_in_out(float t) {
    float s = 1.70158f * 1.525f;
    return t < 0.5f ? 0.5f * (4.f * t * t * ((s + 1.f) * 2.f * t - s)) : 0.5f * (1.f + 2.f * (t - 1.f) * (t - 1.f) * ((s + 1.f) * 2.f * (t - 1.f) + s));
}

float ease_elastic_in(float t) {
    if (t == 0.f) return 0.f;
    if (t == 1.f) return 1.f;
    float p = 0.3f;
    float s = p / 4.f;
    return -powf(2.f, 10.f * (t - 1.f)) * sinf((t - 1.f - s) * (2.f * (float)M_PI) / p);
}

float ease_elastic_out(float t) {
    if (t == 0.f) return 0.f;
    if (t == 1.f) return 1.f;
    float p = 0.3f;
    float s = p / 4.f;
    return 1.f + powf(2.f, -10.f * t) * sinf((t - s) * (2.f * (float)M_PI) / p);
}

float ease_elastic_in_out(float t) {
    if (t == 0.f) return 0.f;
    if (t == 1.f) return 1.f;
    float p = 0.3f * 1.5f;
    float s = p / 4.f;
    return t < 0.5f
        ? -0.5f * powf(2.f, 20.f * t - 10.f) * sinf((20.f * t - 10.75f - s) * (2.f * (float)M_PI) / p)
        : 1.f + 0.5f * powf(2.f, -20.f * t + 10.f) * sinf((20.f * t - 10.75f - s) * (2.f * (float)M_PI) / p);
}

float ease_bounce_in(float t) { return 1.f - ease_bounce_out(1.f - t); }
float ease_bounce_out(float t) {
    if (t < 4.f / 11.f) return (121.f * t * t) / 16.f;
    if (t < 8.f / 11.f) return (363.f / 40.f * t * t) - (99.f / 10.f * t) + 17.f / 5.f;
    if (t < 9.f / 10.f) return (4356.f / 361.f * t * t) - (35442.f / 1805.f * t) + 16061.f / 1805.f;
    return (54.f / 5.f * t * t) - (513.f / 25.f * t) + 268.f / 25.f;
}
float ease_bounce_in_out(float t) { return t < 0.5f ? 0.5f * ease_bounce_in(t * 2.f) : 0.5f * ease_bounce_out(t * 2.f - 1.f) + 0.5f; }

// ============================================================================
// Vector Interpolation
// ============================================================================

vec2 vec2_smoothstep(vec2 a, vec2 b, float t) {
    float s = t * t * (3.f - 2.f * t);
    return vec2_lerp(a, b, s);
}

vec3 vec3_smoothstep(vec3 a, vec3 b, float t) {
    float s = t * t * (3.f - 2.f * t);
    return vec3_lerp(a, b, s);
}

// ============================================================================
// Rectangle / AABB
// ============================================================================

bool rect_contains_point(cortex_rect r, float px, float py) {
    return px >= r.x && px <= r.x + r.w && py >= r.y && py <= r.y + r.h;
}

bool rect_overlaps(cortex_rect a, cortex_rect b) {
    return a.x < b.x + b.w && a.x + a.w > b.x && a.y < b.y + b.h && a.y + a.h > b.y;
}

cortex_rect rect_from_center(float cx, float cy, float w, float h) {
    return (cortex_rect){ cx - w / 2.f, cy - h / 2.f, w, h };
}

cortex_rect rect_from_corners(float x1, float y1, float x2, float y2) {
    float xmin = fminf(x1, x2), xmax = fmaxf(x1, x2);
    float ymin = fminf(y1, y2), ymax = fmaxf(y1, y2);
    return (cortex_rect){ xmin, ymin, xmax - xmin, ymax - ymin };
}

cortex_rect rect_union(cortex_rect a, cortex_rect b) {
    float x1 = fminf(a.x, b.x), y1 = fminf(a.y, b.y);
    float x2 = fmaxf(a.x + a.w, b.x + b.w), y2 = fmaxf(a.y + a.h, b.y + b.h);
    return (cortex_rect){ x1, y1, x2 - x1, y2 - y1 };
}

cortex_rect rect_intersection(cortex_rect a, cortex_rect b) {
    float x1 = fmaxf(a.x, b.x), y1 = fmaxf(a.y, b.y);
    float x2 = fminf(a.x + a.w, b.x + b.w), y2 = fminf(a.y + a.h, b.y + b.h);
    if (x2 <= x1 || y2 <= y1) return (cortex_rect){ 0, 0, 0, 0 };
    return (cortex_rect){ x1, y1, x2 - x1, y2 - y1 };
}

float rect_area(cortex_rect r) { return r.w * r.h; }
vec2 rect_center(cortex_rect r) { return (vec2){ r.x + r.w / 2.f, r.y + r.h / 2.f }; }
vec2 rect_min(cortex_rect r) { return (vec2){ r.x, r.y }; }
vec2 rect_max(cortex_rect r) { return (vec2){ r.x + r.w, r.y + r.h }; }
cortex_rect rect_expand(cortex_rect r, float amount) { return (cortex_rect){ r.x - amount, r.y - amount, r.w + amount * 2.f, r.h + amount * 2.f }; }
cortex_rect rect_translate(cortex_rect r, float dx, float dy) { return (cortex_rect){ r.x + dx, r.y + dy, r.w, r.h }; }

// ============================================================================
// Circle / Sphere
// ============================================================================

bool circle_contains_point(cortex_circle c, float px, float py) {
    float dx = px - c.x, dy = py - c.y;
    return dx * dx + dy * dy <= c.r * c.r;
}

bool circle_overlaps_circle(cortex_circle a, cortex_circle b) {
    float dx = b.x - a.x, dy = b.y - a.y;
    float dist_sq = dx * dx + dy * dy;
    float radius_sum = a.r + b.r;
    return dist_sq <= radius_sum * radius_sum;
}

bool circle_overlaps_rect(cortex_circle c, cortex_rect r) {
    float closest_x = fmaxf(r.x, fminf(c.x, r.x + r.w));
    float closest_y = fmaxf(r.y, fminf(c.y, r.y + r.h));
    float dx = c.x - closest_x, dy = c.y - closest_y;
    return dx * dx + dy * dy <= c.r * c.r;
}

float circle_area(cortex_circle c) { return (float)M_PI * c.r * c.r; }
float circle_circumference(cortex_circle c) { return 2.f * (float)M_PI * c.r; }
cortex_circle circle_from_center(float cx, float cy, float r) { return (cortex_circle){ cx, cy, r }; }

bool sphere_contains_point(cortex_sphere s, float px, float py, float pz) {
    float dx = px - s.x, dy = py - s.y, dz = pz - s.z;
    return dx * dx + dy * dy + dz * dz <= s.r * s.r;
}

bool sphere_overlaps_sphere(cortex_sphere a, cortex_sphere b) {
    float dx = b.x - a.x, dy = b.y - a.y, dz = b.z - a.z;
    float dist_sq = dx * dx + dy * dy + dz * dz;
    float radius_sum = a.r + b.r;
    return dist_sq <= radius_sum * radius_sum;
}

float sphere_volume(cortex_sphere s) { return (4.f / 3.f) * (float)M_PI * s.r * s.r * s.r; }
float sphere_surface_area(cortex_sphere s) { return 4.f * (float)M_PI * s.r * s.r; }

// ============================================================================
// Angle Helpers
// ============================================================================

float vec2_angle(vec2 v) { return (float)atan2((double)v.y, (double)v.x); }
vec2 vec2_from_angle(float radians) { return (vec2){ (float)cos((double)radians), (float)sin((double)radians) }; }

float angle_normalize(float radians) {
    while (radians > (float)M_PI) radians -= 2.f * (float)M_PI;
    while (radians < -(float)M_PI) radians += 2.f * (float)M_PI;
    return radians;
}

float angle_lerp(float a, float b, float t) {
    float diff = angle_normalize(b - a);
    return angle_normalize(a + diff * t);
}

float angle_distance(float a, float b) {
    return fabsf(angle_normalize(b - a));
}

float deg_to_rad(float degrees) { return degrees * ((float)M_PI / 180.f); }
float rad_to_deg(float radians) { return radians * (180.f / (float)M_PI); }

// ============================================================================
// Collision Detection
// ============================================================================

bool point_in_rect(float px, float py, cortex_rect r) { return rect_contains_point(r, px, py); }
bool point_in_circle(float px, float py, cortex_circle c) { return circle_contains_point(c, px, py); }
bool point_in_sphere(float px, float py, float pz, cortex_sphere s) { return sphere_contains_point(s, px, py, pz); }
bool rect_collides_rect(cortex_rect a, cortex_rect b) { return rect_overlaps(a, b); }
bool circle_collides_circle(cortex_circle a, cortex_circle b) { return circle_overlaps_circle(a, b); }
bool sphere_collides_sphere(cortex_sphere a, cortex_sphere b) { return sphere_overlaps_sphere(a, b); }

bool line_intersects_rect(float x1, float y1, float x2, float y2, cortex_rect r) {
    // Liang-Barsky algorithm
    float dx = x2 - x1, dy = y2 - y1;
    float p[4] = { -dx, dx, -dy, dy };
    float q[4] = { x1 - r.x, r.x + r.w - x1, y1 - r.y, r.y + r.h - y1 };
    float u1 = 0.f, u2 = 1.f;
    
    for (int i = 0; i < 4; i++) {
        if (p[i] == 0.f) {
            if (q[i] < 0.f) return false;
        } else {
            float t = q[i] / p[i];
            if (p[i] < 0.f) { if (t > u2) return false; if (t > u1) u1 = t; }
            else { if (t < u1) return false; if (t < u2) u2 = t; }
        }
    }
    return u1 <= u2;
}

bool line_intersects_circle(float x1, float y1, float x2, float y2, cortex_circle c) {
    float dx = x2 - x1, dy = y2 - y1;
    float fx = x1 - c.x, fy = y1 - c.y;
    float a = dx * dx + dy * dy;
    float b = 2.f * (fx * dx + fy * dy);
    float cc = fx * fx + fy * fy - c.r * c.r;
    float discriminant = b * b - 4.f * a * cc;
    if (discriminant < 0.f) return false;
    discriminant = sqrtf(discriminant);
    float t1 = (-b - discriminant) / (2.f * a);
    float t2 = (-b + discriminant) / (2.f * a);
    return (t1 >= 0.f && t1 <= 1.f) || (t2 >= 0.f && t2 <= 1.f);
}

// ============================================================================
// Raycasting
// ============================================================================

ray2d_hit ray2d_intersect_rect(ray2d ray, cortex_rect r) {
    ray2d_hit hit = {0};
    
    // Slab method
    float tmin = 0.f, tmax = FLT_MAX;
    float inv_dx = ray.direction.x != 0.f ? 1.f / ray.direction.x : FLT_MAX;
    float inv_dy = ray.direction.y != 0.f ? 1.f / ray.direction.y : FLT_MAX;
    
    float tx1 = (r.x - ray.origin.x) * inv_dx;
    float tx2 = (r.x + r.w - ray.origin.x) * inv_dx;
    tmin = fmaxf(tmin, fminf(tx1, tx2));
    tmax = fminf(tmax, fmaxf(tx1, tx2));
    
    float ty1 = (r.y - ray.origin.y) * inv_dy;
    float ty2 = (r.y + r.h - ray.origin.y) * inv_dy;
    tmin = fmaxf(tmin, fminf(ty1, ty2));
    tmax = fminf(tmax, fmaxf(ty1, ty2));
    
    if (tmin <= tmax && tmax >= 0.f) {
        hit.hit = true;
        hit.distance = tmin > 0.f ? tmin : tmax;
        hit.point = vec2_add(ray.origin, vec2_scale(ray.direction, hit.distance));
        
        // Calculate normal
        if (fabsf(hit.point.x - r.x) < 0.001f) hit.normal = (vec2){ -1, 0 };
        else if (fabsf(hit.point.x - (r.x + r.w)) < 0.001f) hit.normal = (vec2){ 1, 0 };
        else if (fabsf(hit.point.y - r.y) < 0.001f) hit.normal = (vec2){ 0, -1 };
        else hit.normal = (vec2){ 0, 1 };
    }
    
    return hit;
}

ray2d_hit ray2d_intersect_circle(ray2d ray, cortex_circle c) {
    ray2d_hit hit = {0};
    
    vec2 oc = vec2_sub(ray.origin, (vec2){ c.x, c.y });
    float a = vec2_dot(ray.direction, ray.direction);
    float b = 2.f * vec2_dot(oc, ray.direction);
    float cc = vec2_dot(oc, oc) - c.r * c.r;
    float discriminant = b * b - 4.f * a * cc;
    
    if (discriminant >= 0.f) {
        float t = (-b - sqrtf(discriminant)) / (2.f * a);
        if (t >= 0.f) {
            hit.hit = true;
            hit.distance = t;
            hit.point = vec2_add(ray.origin, vec2_scale(ray.direction, t));
            hit.normal = vec2_normalize(vec2_sub(hit.point, (vec2){ c.x, c.y }));
        }
    }
    
    return hit;
}

// ============================================================================
// Physics / Motion
// ============================================================================

void physics_body_2d_update(physics_body_2d* body, float dt) {
    if (!body) return;
    
    // v = v + a * dt
    body->velocity = vec2_add(body->velocity, vec2_scale(body->acceleration, dt));
    
    // Apply drag
    body->velocity = vec2_scale(body->velocity, 1.f - body->drag * dt);
    
    // p = p + v * dt
    body->position = vec2_add(body->position, vec2_scale(body->velocity, dt));
}

void physics_body_2d_apply_force(physics_body_2d* body, vec2 force) {
    if (!body || body->mass == 0.f) return;
    body->acceleration = vec2_add(body->acceleration, vec2_scale(force, 1.f / body->mass));
}

void physics_body_2d_apply_impulse(physics_body_2d* body, vec2 impulse) {
    if (!body || body->mass == 0.f) return;
    body->velocity = vec2_add(body->velocity, vec2_scale(impulse, 1.f / body->mass));
}

void physics_body_2d_apply_gravity(physics_body_2d* body, float g) {
    if (!body) return;
    body->acceleration.y += g;
}

void physics_resolve_rect_rect(physics_body_2d* a, cortex_rect ra, physics_body_2d* b, cortex_rect rb) {
    if (!a || !b) return;
    
    // Calculate overlap
    float overlap_x = fminf(ra.x + ra.w, rb.x + rb.w) - fmaxf(ra.x, rb.x);
    float overlap_y = fminf(ra.y + ra.h, rb.y + rb.h) - fmaxf(ra.y, rb.y);
    
    if (overlap_x <= 0.f || overlap_y <= 0.f) return;
    
    // Resolve along smallest axis
    if (overlap_x < overlap_y) {
        float sign = (ra.x + ra.w / 2.f < rb.x + rb.w / 2.f) ? -1.f : 1.f;
        a->position.x += sign * overlap_x / 2.f;
        b->position.x -= sign * overlap_x / 2.f;
        
        // Bounce
        float restitution = (a->restitution + b->restitution) / 2.f;
        a->velocity.x *= -restitution;
        b->velocity.x *= -restitution;
    } else {
        float sign = (ra.y + ra.h / 2.f < rb.y + rb.h / 2.f) ? -1.f : 1.f;
        a->position.y += sign * overlap_y / 2.f;
        b->position.y -= sign * overlap_y / 2.f;
        
        float restitution = (a->restitution + b->restitution) / 2.f;
        a->velocity.y *= -restitution;
        b->velocity.y *= -restitution;
    }
}

void physics_resolve_circle_circle(physics_body_2d* a, cortex_circle ca, physics_body_2d* b, cortex_circle cb) {
    if (!a || !b) return;
    
    float dx = cb.x - ca.x, dy = cb.y - ca.y;
    float dist = sqrtf(dx * dx + dy * dy);
    float overlap = ca.r + cb.r - dist;
    
    if (overlap <= 0.f) return;
    
    // Normalize
    float nx = dx / dist, ny = dy / dist;
    
    // Separate
    a->position.x -= nx * overlap / 2.f;
    a->position.y -= ny * overlap / 2.f;
    b->position.x += nx * overlap / 2.f;
    b->position.y += ny * overlap / 2.f;
    
    // Bounce
    float restitution = (a->restitution + b->restitution) / 2.f;
    float rel_vx = a->velocity.x - b->velocity.x;
    float rel_vy = a->velocity.y - b->velocity.y;
    float rel_v_n = rel_vx * nx + rel_vy * ny;
    
    if (rel_v_n > 0.f) {
        float impulse = -(1.f + restitution) * rel_v_n / (1.f / a->mass + 1.f / b->mass);
        a->velocity.x += impulse * nx / a->mass;
        a->velocity.y += impulse * ny / a->mass;
        b->velocity.x -= impulse * nx / b->mass;
        b->velocity.y -= impulse * ny / b->mass;
    }
}

// ============================================================================
// Particle System
// ============================================================================

static uint32_t g_random_seed = 1;

static float rand_float(void) {
    g_random_seed = g_random_seed * 1103515245 + 12345;
    return (float)((g_random_seed >> 16) & 0x7FFF) / 32768.f;
}

particle_system* particle_system_create(int capacity) {
    particle_system* ps = (particle_system*)malloc(sizeof(particle_system));
    if (!ps) return NULL;
    
    ps->particles = (particle*)calloc(capacity, sizeof(particle));
    if (!ps->particles) { free(ps); return NULL; }
    
    ps->count = 0;
    ps->capacity = capacity;
    ps->emitter_position = (vec2){ 0, 0 };
    ps->emitter_velocity_min = (vec2){ -50, -50 };
    ps->emitter_velocity_max = (vec2){ 50, 50 };
    ps->color_start = (vec4){ 1, 1, 1, 1 };
    ps->color_end = (vec4){ 0, 0, 0, 0 };
    ps->size_start = 5.f;
    ps->size_end = 0.f;
    ps->lifetime_min = 1.f;
    ps->lifetime_max = 2.f;
    ps->emission_rate = 10.f;
    ps->emission_accumulator = 0.f;
    
    return ps;
}

void particle_system_free(particle_system* ps) {
    if (ps) {
        free(ps->particles);
        free(ps);
    }
}

void particle_system_emit(particle_system* ps, int count) {
    if (!ps) return;
    
    for (int i = 0; i < count && ps->count < ps->capacity; i++) {
        particle* p = &ps->particles[ps->count++];
        p->position = ps->emitter_position;
        
        p->velocity.x = ps->emitter_velocity_min.x + rand_float() * (ps->emitter_velocity_max.x - ps->emitter_velocity_min.x);
        p->velocity.y = ps->emitter_velocity_min.y + rand_float() * (ps->emitter_velocity_max.y - ps->emitter_velocity_min.y);
        
        p->color = ps->color_start;
        p->size = ps->size_start;
        p->lifetime = ps->lifetime_min + rand_float() * (ps->lifetime_max - ps->lifetime_min);
        p->age = 0.f;
        p->active = true;
    }
}

void particle_system_update(particle_system* ps, float dt) {
    if (!ps) return;
    
    // Emit based on rate
    ps->emission_accumulator += ps->emission_rate * dt;
    int to_emit = (int)ps->emission_accumulator;
    if (to_emit > 0) {
        particle_system_emit(ps, to_emit);
        ps->emission_accumulator -= to_emit;
    }
    
    // Update particles
    for (int i = 0; i < ps->count; i++) {
        particle* p = &ps->particles[i];
        if (!p->active) continue;
        
        p->position = vec2_add(p->position, vec2_scale(p->velocity, dt));
        p->age += dt;
        
        float t = p->age / p->lifetime;
        p->color.x = ps->color_start.x + (ps->color_end.x - ps->color_start.x) * t;
        p->color.y = ps->color_start.y + (ps->color_end.y - ps->color_start.y) * t;
        p->color.z = ps->color_start.z + (ps->color_end.z - ps->color_start.z) * t;
        p->color.w = ps->color_start.w + (ps->color_end.w - ps->color_start.w) * t;
        p->size = ps->size_start + (ps->size_end - ps->size_start) * t;
        
        if (p->age >= p->lifetime) {
            p->active = false;
        }
    }
    
    // Compact inactive particles
    int write = 0;
    for (int read = 0; read < ps->count; read++) {
        if (ps->particles[read].active) {
            if (write != read) ps->particles[write] = ps->particles[read];
            write++;
        }
    }
    ps->count = write;
}

void particle_system_render(particle_system* ps) {
    (void)ps; // Placeholder for rendering integration
}

// ============================================================================
// Timer / Animation
// ============================================================================

animation_timer animation_timer_create(float duration, bool looping) {
    animation_timer t = { duration, 0.f, looping, false, 0.f };
    return t;
}

void animation_timer_start(animation_timer* timer) { if (timer) timer->playing = true; }
void animation_timer_stop(animation_timer* timer) { if (timer) timer->playing = false; }
void animation_timer_reset(animation_timer* timer) { if (timer) { timer->elapsed = 0.f; timer->progress = 0.f; } }
bool animation_timer_is_complete(animation_timer* timer) { return timer ? timer->progress >= 1.f : false; }

void animation_timer_update(animation_timer* timer, float dt) {
    if (!timer || !timer->playing) return;
    
    timer->elapsed += dt;
    timer->progress = timer->elapsed / timer->duration;
    
    if (timer->progress >= 1.f) {
        if (timer->looping) {
            timer->elapsed = fmodf(timer->elapsed, timer->duration);
            timer->progress = timer->elapsed / timer->duration;
        } else {
            timer->progress = 1.f;
            timer->playing = false;
        }
    }
}

// ============================================================================
// Input State
// ============================================================================

void input_update(input_state* state) {
    if (!state) return;
    memset(state->pressed, 0, sizeof(state->pressed));
    memset(state->released, 0, sizeof(state->released));
    state->mouse_delta = (vec2){ 0, 0 };
}

void input_key_down(input_state* state, int key) {
    if (!state || key < 0 || key >= 256) return;
    state->pressed[key] = true;
    state->held[key] = true;
}

void input_key_up(input_state* state, int key) {
    if (!state || key < 0 || key >= 256) return;
    state->held[key] = false;
    state->released[key] = true;
}

void input_mouse_move(input_state* state, float x, float y) {
    if (!state) return;
    state->mouse_delta.x = x - state->mouse_position.x;
    state->mouse_delta.y = y - state->mouse_position.y;
    state->mouse_position = (vec2){ x, y };
}

void input_mouse_button(input_state* state, int button, bool pressed) {
    if (!state || button < 0 || button >= 3) return;
    state->mouse_buttons[button] = pressed;
}

bool input_is_pressed(input_state* state, int key) { return state && key >= 0 && key < 256 && state->pressed[key]; }
bool input_is_held(input_state* state, int key) { return state && key >= 0 && key < 256 && state->held[key]; }
bool input_is_released(input_state* state, int key) { return state && key >= 0 && key < 256 && state->released[key]; }

// ============================================================================
// Noise / Random
// ============================================================================

// Simple Perlin noise implementation
static float fade(float t) { return t * t * t * (t * (t * 6.f - 15.f) + 10.f); }
static float lerp_noise(float a, float b, float t) { return a + t * (b - a); }
static float grad(int hash, float x) { return (hash & 1) ? x : -x; }
static float grad2(int hash, float x, float y) { int h = hash & 3; return ((h & 1) ? -x : x) + ((h & 2) ? -y : y); }
static float grad3(int hash, float x, float y, float z) {
    int h = hash & 15;
    float u = h < 8 ? x : y;
    float v = h < 4 ? y : (h == 12 || h == 14 ? x : z);
    return ((h & 1) ? -u : u) + ((h & 2) ? -v : v);
}

static int perm[512];
static int perm_initialized = 0;

static void init_perm(void) {
    if (perm_initialized) return;
    for (int i = 0; i < 256; i++) perm[i] = i;
    for (int i = 255; i > 0; i--) {
        int j = rand() % (i + 1);
        int tmp = perm[i]; perm[i] = perm[j]; perm[j] = tmp;
    }
    for (int i = 0; i < 256; i++) perm[256 + i] = perm[i];
    perm_initialized = 1;
}

float noise_perlin_1d(float x) {
    init_perm();
    int X = (int)floorf(x) & 255;
    x -= floorf(x);
    float u = fade(x);
    return lerp_noise(grad(perm[X], x), grad(perm[X + 1], x - 1), u);
}

float noise_perlin_2d(float x, float y) {
    init_perm();
    int X = (int)floorf(x) & 255, Y = (int)floorf(y) & 255;
    x -= floorf(x); y -= floorf(y);
    float u = fade(x), v = fade(y);
    int A = perm[X] + Y, B = perm[X + 1] + Y;
    return lerp_noise(
        lerp_noise(grad2(perm[A], x, y), grad2(perm[B], x - 1, y), u),
        lerp_noise(grad2(perm[A + 1], x, y - 1), grad2(perm[B + 1], x - 1, y - 1), u),
        v
    );
}

float noise_perlin_3d(float x, float y, float z) {
    init_perm();
    int X = (int)floorf(x) & 255, Y = (int)floorf(y) & 255, Z = (int)floorf(z) & 255;
    x -= floorf(x); y -= floorf(y); z -= floorf(z);
    float u = fade(x), v = fade(y), w = fade(z);
    int A = perm[X] + Y, AA = perm[A] + Z, AB = perm[A + 1] + Z;
    int B = perm[X + 1] + Y, BA = perm[B] + Z, BB = perm[B + 1] + Z;
    return lerp_noise(
        lerp_noise(
            lerp_noise(grad3(perm[AA], x, y, z), grad3(perm[BA], x - 1, y, z), u),
            lerp_noise(grad3(perm[AB], x, y - 1, z), grad3(perm[BB], x - 1, y - 1, z), u),
            v
        ),
        lerp_noise(
            lerp_noise(grad3(perm[AA + 1], x, y, z - 1), grad3(perm[BA + 1], x - 1, y, z - 1), u),
            lerp_noise(grad3(perm[AB + 1], x, y - 1, z - 1), grad3(perm[BB + 1], x - 1, y - 1, z - 1), u),
            v
        ),
        w
    );
}

void random_seed(uint32_t seed) { g_random_seed = seed; srand(seed); }

int random_range(int min, int max) {
    if (min >= max) return min;
    return min + rand() % (max - min + 1);
}

float random_range_float(float min, float max) {
    return min + rand_float() * (max - min);
}

vec2 random_point_in_circle(float radius) {
    float angle = rand_float() * 2.f * (float)M_PI;
    float r = sqrtf(rand_float()) * radius;
    return (vec2){ cosf(angle) * r, sinf(angle) * r };
}

vec2 random_point_on_circle(float radius) {
    float angle = rand_float() * 2.f * (float)M_PI;
    return (vec2){ cosf(angle) * radius, sinf(angle) * radius };
}

vec3 random_point_in_sphere(float radius) {
    float theta = rand_float() * 2.f * (float)M_PI;
    float phi = acosf(2.f * rand_float() - 1.f);
    float r = powf(rand_float(), 1.f / 3.f) * radius;
    return (vec3){ r * sinf(phi) * cosf(theta), r * sinf(phi) * sinf(theta), r * cosf(phi) };
}

vec3 random_point_on_sphere(float radius) {
    float theta = rand_float() * 2.f * (float)M_PI;
    float phi = acosf(2.f * rand_float() - 1.f);
    return (vec3){ radius * sinf(phi) * cosf(theta), radius * sinf(phi) * sinf(theta), radius * cosf(phi) };
}

// ============================================================================
// Utility
// ============================================================================

float approach(float current, float target, float max_delta) {
    if (current < target) return fminf(current + max_delta, target);
    return fmaxf(current - max_delta, target);
}

float precise_approach(float current, float target, float max_delta) {
    float diff = target - current;
    if (fabsf(diff) <= max_delta) return target;
    return current + (diff > 0 ? max_delta : -max_delta);
}

vec2 vec2_approach(vec2 current, vec2 target, float max_distance) {
    vec2 diff = vec2_sub(target, current);
    float dist = vec2_length(diff);
    if (dist <= max_distance) return target;
    return vec2_add(current, vec2_scale(vec2_normalize(diff), max_distance));
}

vec3 vec3_approach(vec3 current, vec3 target, float max_distance) {
    vec3 diff = vec3_sub(target, current);
    float dist = vec3_length(diff);
    if (dist <= max_distance) return target;
    return vec3_add(current, vec3_scale(vec3_normalize(diff), max_distance));
}

// ============================================================================
// Screen Shake
// ============================================================================

void screen_shake_start(screen_shake* shake, float intensity, float duration, float frequency) {
    if (!shake) return;
    shake->intensity = intensity;
    shake->duration = duration;
    shake->elapsed = 0.f;
    shake->frequency = frequency;
    shake->offset = (vec2){ 0, 0 };
}

void screen_shake_update(screen_shake* shake, float dt) {
    if (!shake || shake->elapsed >= shake->duration) {
        if (shake) shake->offset = (vec2){ 0, 0 };
        return;
    }
    
    shake->elapsed += dt;
    
    float progress = shake->elapsed / shake->duration;
    float decay = 1.f - progress;
    
    shake->offset.x = sinf(shake->elapsed * shake->frequency * 2.f * (float)M_PI) * shake->intensity * decay;
    shake->offset.y = cosf(shake->elapsed * shake->frequency * 2.f * (float)M_PI * 1.3f) * shake->intensity * decay;
}

vec2 screen_shake_get_offset(screen_shake* shake) {
    return shake ? shake->offset : (vec2){ 0, 0 };
}
