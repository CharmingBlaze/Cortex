// Cortex Standard Library - Math Module
// Provides common math functions

#ifndef CORTEX_STD_MATH_H
#define CORTEX_STD_MATH_H

#include <math.h>

// Constants
#define STD_PI 3.14159265358979323846
#define STD_E  2.71828182845904523536
#define STD_PHI 1.61803398874989484820  // Golden ratio

// Basic math functions (wrap C math library)
static inline double std_math_abs(double x) { return fabs(x); }
static inline double std_math_sqrt(double x) { return sqrt(x); }
static inline double std_math_pow(double base, double exp) { return pow(base, exp); }
static inline double std_math_exp(double x) { return exp(x); }
static inline double std_math_log(double x) { return log(x); }
static inline double std_math_log10(double x) { return log10(x); }
static inline double std_math_log2(double x) { return log2(x); }

// Trigonometric functions
static inline double std_math_sin(double x) { return sin(x); }
static inline double std_math_cos(double x) { return cos(x); }
static inline double std_math_tan(double x) { return tan(x); }
static inline double std_math_asin(double x) { return asin(x); }
static inline double std_math_acos(double x) { return acos(x); }
static inline double std_math_atan(double x) { return atan(x); }
static inline double std_math_atan2(double y, double x) { return atan2(y, x); }

// Hyperbolic functions
static inline double std_math_sinh(double x) { return sinh(x); }
static inline double std_math_cosh(double x) { return cosh(x); }
static inline double std_math_tanh(double x) { return tanh(x); }

// Rounding functions
static inline double std_math_floor(double x) { return floor(x); }
static inline double std_math_ceil(double x) { return ceil(x); }
static inline double std_math_round(double x) { return round(x); }
static inline double std_math_trunc(double x) { return trunc(x); }

// Min/Max/Clamp
static inline double std_math_min(double a, double b) { return a < b ? a : b; }
static inline double std_math_max(double a, double b) { return a > b ? a : b; }
static inline double std_math_clamp(double x, double lo, double hi) {
    return x < lo ? lo : (x > hi ? hi : x);
}

// Utility
static inline double std_math_sign(double x) { return (x > 0) - (x < 0); }
static inline int std_math_is_nan(double x) { return isnan(x); }
static inline int std_math_is_inf(double x) { return isinf(x); }
static inline int std_math_is_finite(double x) { return isfinite(x); }

// Interpolation
static inline double std_math_lerp(double a, double b, double t) {
    return a + t * (b - a);
}
static inline double std_math_smoothstep(double edge0, double edge1, double x) {
    double t = std_math_clamp((x - edge0) / (edge1 - edge0), 0.0, 1.0);
    return t * t * (3.0 - 2.0 * t);
}

// Random number generation
#include <stdlib.h>
static inline void std_math_srand(unsigned int seed) { srand(seed); }
static inline int std_math_rand() { return rand(); }
static inline double std_math_rand_range(double min, double max) {
    return min + (max - min) * ((double)rand() / RAND_MAX);
}

#endif // CORTEX_STD_MATH_H
