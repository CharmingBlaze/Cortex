package binder

// builtinPrelude is a minimal built-in declaration layer for modernc.org/cc/v3.
// Derived from modernc.org/cc/v3 all_test.go (parserTestBuiltin), BSD-style license.
const builtinPrelude = `
#define __DI__
#define __FUNCTION__ __func__
#define __HI__
#define __PRETTY_FUNCTION__ __func__
#define __QI__
#define __SI__
#define __builtin_add_overflow(...) 0
#define __builtin_constant_p(x) (__builtin_constant_p_impl(0, x))
#define __builtin_expect(exp, c) (exp)
#define __builtin_mul_overflow(...) 0
#define __builtin_offsetof(type, member) ((__SIZE_TYPE__)&(((type*)0)->member))
#define __builtin_sub_overflow(...) 0
#define __builtin_va_arg(ap, type) ((type)__builtin_va_arg_impl(ap))
#define __builtin_va_end(ap)
#define __builtin_va_start(ap, v)
#define __declspec(...)
#define __extension__
#define __read_only__ 0
#define __sync_synchronize(...)
#define __word__
#define __write_only__ 0
#define asm __asm__

#ifndef __GNUC__
#define __attribute__(x)
#endif

#ifdef __PTRDIFF_TYPE__
typedef __PTRDIFF_TYPE__ ptrdiff_t;
#endif

#ifdef __SIZE_TYPE__
typedef __SIZE_TYPE__ size_t;
#endif

#ifdef __WCHAR_TYPE__
typedef __WCHAR_TYPE__ wchar_t;
#endif

typedef void *__builtin_va_list;

#if __SIZEOF_POINTER__ == 8
#else
typedef long double __float128;
#endif

#ifdef __UINT16_TYPE__
__UINT16_TYPE__ __builtin_bswap16 (__UINT16_TYPE__ x);
#endif

#ifdef __UINT32_TYPE__
__UINT32_TYPE__ __builtin_bswap32 (__UINT32_TYPE__ x);
#endif

#ifdef __UINT64_TYPE__
__UINT64_TYPE__ __builtin_bswap64 (__UINT64_TYPE__ x);
#endif

#ifdef __SIZEOF_INT128__
typedef __INT64_TYPE__ __int128_t[2];
typedef __UINT64_TYPE__ __uint128_t[2];
#endif

#if defined(__MINGW32__) || defined(__MINGW64__)
int gnu_printf(const char *format, ...);
int gnu_scanf(const char *format, ...);
int ms_scanf(const char *format, ...);
int ms_printf(const char *format, ...);
#endif

typedef struct { char real, imag; } __COMPLEX_CHAR_TYPE__;
typedef struct { double real, imag; } __COMPLEX_DOUBLE_TYPE__;
typedef struct { float real, imag; } __COMPLEX_FLOAT_TYPE__;
typedef struct { int real, imag; } __COMPLEX_INT_TYPE__;
typedef struct { long double real, imag; } __COMPLEX_LONG_DOUBLE_TYPE__;
typedef struct { long real, imag; } __COMPLEX_LONG_TYPE__;
typedef struct { long long real, imag; } __COMPLEX_LONG_LONG_TYPE__;
typedef struct { long long unsigned real, imag; } __COMPLEX_LONG_LONG_UNSIGNED_TYPE__;
typedef struct { long unsigned real, imag; } __COMPLEX_LONG_UNSIGNED_TYPE__;
typedef struct { short real, imag; } __COMPLEX_SHORT_TYPE__;
typedef struct { unsigned real, imag; } __COMPLEX_UNSIGNED_TYPE__;
typedef struct { unsigned short real, imag; } __COMPLEX_SHORT_UNSIGNED_TYPE__;

int __builtin_clzll (unsigned long long);
int __builtin_constant_p_impl(int, ...);
int __printf__ ( const char * format, ... );
int __scanf__ ( const char *format, ... );
void *__builtin_alloca (size_t);
void *__builtin_extract_return_addr (void *addr);
void *__builtin_frame_address (unsigned int level);
void *__builtin_malloc(size_t);
void *__builtin_stack_save(void);
void *__builtin_va_arg_impl(void* ap);

`
