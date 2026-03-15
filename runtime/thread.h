// Cortex Thread Runtime - Cross-platform threading and channels
// Uses pthreads on POSIX, Windows threads on Windows

#ifndef CORTEX_THREAD_H
#define CORTEX_THREAD_H

#include <stdbool.h>
#include <stdint.h>

// === Thread Handle ===
typedef void* cortex_thread;

// === Channel Handle ===
typedef void* cortex_channel;

// === Thread Functions ===

// Spawn a new thread running function with arg
// Returns thread handle
cortex_thread thread_spawn(void (*fn)(void*), void* arg);

// Wait for thread to complete
void thread_join(cortex_thread t);

// Check if thread is still running
bool thread_is_running(cortex_thread t);

// Get current thread ID
uint64_t thread_id(void);

// Sleep current thread (milliseconds)
void thread_sleep_ms(int ms);

// === Channel Functions ===

// Create a channel for passing values of a given size
// elem_size: size of each element in bytes
// capacity: max buffered items (0 = unbounded, not recommended)
cortex_channel channel_create(int elem_size, int capacity);

// Send value to channel (blocks if channel full)
// Returns true on success, false if channel closed
bool channel_send(cortex_channel ch, void* value);

// Receive value from channel (blocks if channel empty)
// Returns true on success, false if channel closed
bool channel_recv(cortex_channel ch, void* out);

// Try to send without blocking
// Returns: 1 = sent, 0 = would block, -1 = closed
int channel_try_send(cortex_channel ch, void* value);

// Try to receive without blocking
// Returns: 1 = received, 0 = would block, -1 = closed
int channel_try_recv(cortex_channel ch, void* out);

// Close channel (no more sends allowed)
void channel_close(cortex_channel ch);

// Check if channel is closed
bool channel_is_closed(cortex_channel ch);

// Free channel resources
void channel_free(cortex_channel ch);

// === Convenience Macros for Type-Safe Channels ===

// Create channel for type T with capacity N
#define channel_of(T, N) channel_create(sizeof(T), N)

// Send typed value
#define channel_send_typed(ch, val) ({ \
    __typeof__(val) _v = val; \
    channel_send(ch, &_v); \
})

// Receive typed value
#define channel_recv_typed(ch, T) ({ \
    T _v; \
    channel_recv(ch, &_v) ? _v : (T){0}; \
})

#endif // CORTEX_THREAD_H
