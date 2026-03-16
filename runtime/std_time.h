// Cortex Standard Library - Time Module
// Provides time and date functions

#ifndef CORTEX_STD_TIME_H
#define CORTEX_STD_TIME_H

#include <time.h>
#include <sys/timeb.h>

// Get current time as Unix timestamp (seconds since epoch)
static inline time_t std_time_now() {
    return time(NULL);
}

// Get current time in milliseconds since epoch
static inline long long std_time_now_ms() {
    struct timeb tb;
    ftime(&tb);
    return (long long)tb.time * 1000 + tb.millitm;
}

// Get current time in microseconds (Windows-specific)
#ifdef _WIN32
#include <windows.h>
static inline long long std_time_now_us() {
    LARGE_INTEGER frequency, counter;
    QueryPerformanceFrequency(&frequency);
    QueryPerformanceCounter(&counter);
    return (counter.QuadPart * 1000000LL) / frequency.QuadPart;
}
#else
#include <sys/time.h>
static inline long long std_time_now_us() {
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (long long)tv.tv_sec * 1000000LL + tv.tv_usec;
}
#endif

// Sleep for milliseconds
#ifdef _WIN32
static inline void std_time_sleep_ms(unsigned int ms) {
    Sleep(ms);
}
#else
#include <unistd.h>
static inline void std_time_sleep_ms(unsigned int ms) {
    usleep(ms * 1000);
}
#endif

// Sleep for seconds
static inline void std_time_sleep(double seconds) {
    std_time_sleep_ms((unsigned int)(seconds * 1000));
}

// Time structure
typedef struct {
    int year;
    int month;    // 1-12
    int day;      // 1-31
    int hour;     // 0-23
    int minute;   // 0-59
    int second;   // 0-59
    int weekday;  // 0-6 (Sunday = 0)
    int yearday;  // 0-365
} std_datetime_t;

// Convert timestamp to datetime
static inline std_datetime_t std_time_to_datetime(time_t timestamp) {
    struct tm* t = localtime(&timestamp);
    std_datetime_t dt;
    dt.year = t->tm_year + 1900;
    dt.month = t->tm_mon + 1;
    dt.day = t->tm_mday;
    dt.hour = t->tm_hour;
    dt.minute = t->tm_min;
    dt.second = t->tm_sec;
    dt.weekday = t->tm_wday;
    dt.yearday = t->tm_yday;
    return dt;
}

// Convert datetime to timestamp
static inline time_t std_time_from_datetime(std_datetime_t dt) {
    struct tm t = {0};
    t.tm_year = dt.year - 1900;
    t.tm_mon = dt.month - 1;
    t.tm_mday = dt.day;
    t.tm_hour = dt.hour;
    t.tm_min = dt.minute;
    t.tm_sec = dt.second;
    return mktime(&t);
}

// Get current datetime
static inline std_datetime_t std_time_datetime_now() {
    return std_time_to_datetime(time(NULL));
}

// Format time as string (ISO 8601 format)
static inline void std_time_format_iso(std_datetime_t dt, char* buf, size_t bufsize) {
    snprintf(buf, bufsize, "%04d-%02d-%02dT%02d:%02d:%02d",
             dt.year, dt.month, dt.day, dt.hour, dt.minute, dt.second);
}

// Format time as date string (YYYY-MM-DD)
static inline void std_time_format_date(std_datetime_t dt, char* buf, size_t bufsize) {
    snprintf(buf, bufsize, "%04d-%02d-%02d", dt.year, dt.month, dt.day);
}

// Format time as time string (HH:MM:SS)
static inline void std_time_format_time(std_datetime_t dt, char* buf, size_t bufsize) {
    snprintf(buf, bufsize, "%02d:%02d:%02d", dt.hour, dt.minute, dt.second);
}

// Timer structure for measuring elapsed time
typedef struct {
    long long start_time;
    int running;
} std_timer_t;

// Create and start a timer
static inline std_timer_t std_timer_start() {
    std_timer_t t;
    t.start_time = std_time_now_us();
    t.running = 1;
    return t;
}

// Stop timer and return elapsed microseconds
static inline long long std_timer_stop(std_timer_t* t) {
    if (!t->running) return 0;
    long long elapsed = std_time_now_us() - t->start_time;
    t->running = 0;
    return elapsed;
}

// Get elapsed time without stopping
static inline long long std_timer_elapsed_us(std_timer_t* t) {
    if (!t->running) return 0;
    return std_time_now_us() - t->start_time;
}

static inline double std_timer_elapsed_ms(std_timer_t* t) {
    return std_timer_elapsed_us(t) / 1000.0;
}

static inline double std_timer_elapsed_sec(std_timer_t* t) {
    return std_timer_elapsed_us(t) / 1000000.0;
}

// Reset timer
static inline void std_timer_reset(std_timer_t* t) {
    t->start_time = std_time_now_us();
    t->running = 1;
}

#endif // CORTEX_STD_TIME_H
