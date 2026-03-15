// Cortex Thread Runtime - Cross-platform threading and channels
// Uses pthreads on POSIX, Windows threads on Windows

#include "thread.h"
#include <stdlib.h>
#include <string.h>

// === Platform Detection ===
#if defined(_WIN32) || defined(__CYGWIN__)
    #define CORTEX_USE_WINDOWS_THREADS
    #include <windows.h>
    #include <process.h>
#else
    #define CORTEX_USE_PTHREADS
    #include <pthread.h>
    #include <unistd.h>
    #include <time.h>
#endif

// === Thread Implementation ===

#ifdef CORTEX_USE_WINDOWS_THREADS
// Windows thread wrapper
typedef struct {
    void (*fn)(void*);
    void* arg;
    HANDLE handle;
    volatile int running;
} win_thread_data;

static unsigned __stdcall thread_wrapper(void* arg) {
    win_thread_data* data = (win_thread_data*)arg;
    data->running = 1;
    data->fn(data->arg);
    data->running = 0;
    return 0;
}

cortex_thread thread_spawn(void (*fn)(void*), void* arg) {
    win_thread_data* data = (win_thread_data*)malloc(sizeof(win_thread_data));
    if (!data) return NULL;
    
    data->fn = fn;
    data->arg = arg;
    data->running = 0;
    
    uintptr_t handle = _beginthreadex(NULL, 0, thread_wrapper, data, 0, NULL);
    if (handle == 0) {
        free(data);
        return NULL;
    }
    data->handle = (HANDLE)handle;
    return data;
}

void thread_join(cortex_thread t) {
    if (!t) return;
    win_thread_data* data = (win_thread_data*)t;
    WaitForSingleObject(data->handle, INFINITE);
    CloseHandle(data->handle);
    free(data);
}

bool thread_is_running(cortex_thread t) {
    if (!t) return false;
    win_thread_data* data = (win_thread_data*)t;
    return data->running != 0;
}

uint64_t thread_id(void) {
    return (uint64_t)GetCurrentThreadId();
}

void thread_sleep_ms(int ms) {
    Sleep((DWORD)ms);
}

#else
// POSIX pthreads implementation
typedef struct {
    void (*fn)(void*);
    void* arg;
    pthread_t thread;
    volatile int running;
} pthread_data;

static void* thread_wrapper(void* arg) {
    pthread_data* data = (pthread_data*)arg;
    data->running = 1;
    data->fn(data->arg);
    data->running = 0;
    return NULL;
}

cortex_thread thread_spawn(void (*fn)(void*), void* arg) {
    pthread_data* data = (pthread_data*)malloc(sizeof(pthread_data));
    if (!data) return NULL;
    
    data->fn = fn;
    data->arg = arg;
    data->running = 0;
    
    if (pthread_create(&data->thread, NULL, thread_wrapper, data) != 0) {
        free(data);
        return NULL;
    }
    return data;
}

void thread_join(cortex_thread t) {
    if (!t) return;
    pthread_data* data = (pthread_data*)t;
    pthread_join(data->thread, NULL);
    free(data);
}

bool thread_is_running(cortex_thread t) {
    if (!t) return false;
    pthread_data* data = (pthread_data*)t;
    return data->running != 0;
}

uint64_t thread_id(void) {
    return (uint64_t)pthread_self();
}

void thread_sleep_ms(int ms) {
    usleep(ms * 1000);
}

#endif

// === Channel Implementation ===

// Channel structure with mutex protection
typedef struct {
    char* buffer;           // Ring buffer storage
    int elem_size;          // Size of each element
    int capacity;           // Max elements
    int head;               // Write position
    int tail;               // Read position
    int count;              // Current elements
    int closed;             // Channel closed flag
    
#ifdef CORTEX_USE_WINDOWS_THREADS
    CRITICAL_SECTION lock;
    HANDLE not_empty;       // Signal: data available
    HANDLE not_full;        // Signal: space available
#else
    pthread_mutex_t lock;
    pthread_cond_t not_empty;
    pthread_cond_t not_full;
#endif
} channel_impl;

cortex_channel channel_create(int elem_size, int capacity) {
    if (elem_size <= 0) return NULL;
    if (capacity <= 0) capacity = 64;  // Default capacity
    
    channel_impl* ch = (channel_impl*)malloc(sizeof(channel_impl));
    if (!ch) return NULL;
    
    ch->buffer = (char*)malloc(elem_size * capacity);
    if (!ch->buffer) {
        free(ch);
        return NULL;
    }
    
    ch->elem_size = elem_size;
    ch->capacity = capacity;
    ch->head = 0;
    ch->tail = 0;
    ch->count = 0;
    ch->closed = 0;
    
#ifdef CORTEX_USE_WINDOWS_THREADS
    InitializeCriticalSection(&ch->lock);
    ch->not_empty = CreateEvent(NULL, FALSE, FALSE, NULL);
    ch->not_full = CreateEvent(NULL, FALSE, FALSE, NULL);
#else
    pthread_mutex_init(&ch->lock, NULL);
    pthread_cond_init(&ch->not_empty, NULL);
    pthread_cond_init(&ch->not_full, NULL);
#endif
    
    return ch;
}

bool channel_send(cortex_channel c, void* value) {
    if (!c || !value) return false;
    channel_impl* ch = (channel_impl*)c;
    
#ifdef CORTEX_USE_WINDOWS_THREADS
    EnterCriticalSection(&ch->lock);
#else
    pthread_mutex_lock(&ch->lock);
#endif
    
    // Wait for space
    while (ch->count >= ch->capacity && !ch->closed) {
#ifdef CORTEX_USE_WINDOWS_THREADS
        LeaveCriticalSection(&ch->lock);
        WaitForSingleObject(ch->not_full, INFINITE);
        EnterCriticalSection(&ch->lock);
#else
        pthread_cond_wait(&ch->not_full, &ch->lock);
#endif
    }
    
    if (ch->closed) {
#ifdef CORTEX_USE_WINDOWS_THREADS
        LeaveCriticalSection(&ch->lock);
#else
        pthread_mutex_unlock(&ch->lock);
#endif
        return false;
    }
    
    // Copy value into buffer
    memcpy(ch->buffer + ch->head * ch->elem_size, value, ch->elem_size);
    ch->head = (ch->head + 1) % ch->capacity;
    ch->count++;
    
#ifdef CORTEX_USE_WINDOWS_THREADS
    SetEvent(ch->not_empty);
    LeaveCriticalSection(&ch->lock);
#else
    pthread_cond_signal(&ch->not_empty);
    pthread_mutex_unlock(&ch->lock);
#endif
    
    return true;
}

bool channel_recv(cortex_channel c, void* out) {
    if (!c || !out) return false;
    channel_impl* ch = (channel_impl*)c;
    
#ifdef CORTEX_USE_WINDOWS_THREADS
    EnterCriticalSection(&ch->lock);
#else
    pthread_mutex_lock(&ch->lock);
#endif
    
    // Wait for data
    while (ch->count == 0 && !ch->closed) {
#ifdef CORTEX_USE_WINDOWS_THREADS
        LeaveCriticalSection(&ch->lock);
        WaitForSingleObject(ch->not_empty, INFINITE);
        EnterCriticalSection(&ch->lock);
#else
        pthread_cond_wait(&ch->not_empty, &ch->lock);
#endif
    }
    
    if (ch->count == 0) {
#ifdef CORTEX_USE_WINDOWS_THREADS
        LeaveCriticalSection(&ch->lock);
#else
        pthread_mutex_unlock(&ch->lock);
#endif
        return false;
    }
    
    // Copy value from buffer
    memcpy(out, ch->buffer + ch->tail * ch->elem_size, ch->elem_size);
    ch->tail = (ch->tail + 1) % ch->capacity;
    ch->count--;
    
#ifdef CORTEX_USE_WINDOWS_THREADS
    SetEvent(ch->not_full);
    LeaveCriticalSection(&ch->lock);
#else
    pthread_cond_signal(&ch->not_full);
    pthread_mutex_unlock(&ch->lock);
#endif
    
    return true;
}

int channel_try_send(cortex_channel c, void* value) {
    if (!c || !value) return -1;
    channel_impl* ch = (channel_impl*)c;
    
#ifdef CORTEX_USE_WINDOWS_THREADS
    EnterCriticalSection(&ch->lock);
#else
    pthread_mutex_lock(&ch->lock);
#endif
    
    if (ch->closed) {
#ifdef CORTEX_USE_WINDOWS_THREADS
        LeaveCriticalSection(&ch->lock);
#else
        pthread_mutex_unlock(&ch->lock);
#endif
        return -1;
    }
    
    if (ch->count >= ch->capacity) {
#ifdef CORTEX_USE_WINDOWS_THREADS
        LeaveCriticalSection(&ch->lock);
#else
        pthread_mutex_unlock(&ch->lock);
#endif
        return 0;
    }
    
    memcpy(ch->buffer + ch->head * ch->elem_size, value, ch->elem_size);
    ch->head = (ch->head + 1) % ch->capacity;
    ch->count++;
    
#ifdef CORTEX_USE_WINDOWS_THREADS
    SetEvent(ch->not_empty);
    LeaveCriticalSection(&ch->lock);
#else
    pthread_cond_signal(&ch->not_empty);
    pthread_mutex_unlock(&ch->lock);
#endif
    
    return 1;
}

int channel_try_recv(cortex_channel c, void* out) {
    if (!c || !out) return -1;
    channel_impl* ch = (channel_impl*)c;
    
#ifdef CORTEX_USE_WINDOWS_THREADS
    EnterCriticalSection(&ch->lock);
#else
    pthread_mutex_lock(&ch->lock);
#endif
    
    if (ch->count == 0) {
        int closed = ch->closed;
#ifdef CORTEX_USE_WINDOWS_THREADS
        LeaveCriticalSection(&ch->lock);
#else
        pthread_mutex_unlock(&ch->lock);
#endif
        return closed ? -1 : 0;
    }
    
    memcpy(out, ch->buffer + ch->tail * ch->elem_size, ch->elem_size);
    ch->tail = (ch->tail + 1) % ch->capacity;
    ch->count--;
    
#ifdef CORTEX_USE_WINDOWS_THREADS
    SetEvent(ch->not_full);
    LeaveCriticalSection(&ch->lock);
#else
    pthread_cond_signal(&ch->not_full);
    pthread_mutex_unlock(&ch->lock);
#endif
    
    return 1;
}

void channel_close(cortex_channel c) {
    if (!c) return;
    channel_impl* ch = (channel_impl*)c;
    
#ifdef CORTEX_USE_WINDOWS_THREADS
    EnterCriticalSection(&ch->lock);
    ch->closed = 1;
    SetEvent(ch->not_empty);
    SetEvent(ch->not_full);
    LeaveCriticalSection(&ch->lock);
#else
    pthread_mutex_lock(&ch->lock);
    ch->closed = 1;
    pthread_cond_broadcast(&ch->not_empty);
    pthread_cond_broadcast(&ch->not_full);
    pthread_mutex_unlock(&ch->lock);
#endif
}

bool channel_is_closed(cortex_channel c) {
    if (!c) return true;
    channel_impl* ch = (channel_impl*)c;
    return ch->closed != 0;
}

void channel_free(cortex_channel c) {
    if (!c) return;
    channel_impl* ch = (channel_impl*)c;
    
#ifdef CORTEX_USE_WINDOWS_THREADS
    DeleteCriticalSection(&ch->lock);
    CloseHandle(ch->not_empty);
    CloseHandle(ch->not_full);
#else
    pthread_mutex_destroy(&ch->lock);
    pthread_cond_destroy(&ch->not_empty);
    pthread_cond_destroy(&ch->not_full);
#endif
    
    free(ch->buffer);
    free(ch);
}
