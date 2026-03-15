// Cortex Async Runtime - libco wrapper for coroutines
// Based on libco: https://github.com/howerj/libco

#ifndef CORTEX_ASYNC_H
#define CORTEX_ASYNC_H

#include <stdint.h>
#include <stdbool.h>

// Coroutine handle
typedef void* co_t;

// Create a new coroutine
// entry: the function to run
// stack_size: size of the stack (0 = default 64KB)
co_t co_create(void (*entry)(void*), void* arg, int stack_size);

// Start or resume a coroutine
void co_resume(co_t co);

// Yield back to caller
void co_yield(void);

// Free a coroutine
void co_free(co_t co);

// Get current coroutine (null if main)
co_t co_current(void);

// Check if coroutine is finished
bool co_finished(co_t co);

// === Cortex async API ===

// Async task handle
typedef int64_t async_task;

// Create async task from a coroutine function
async_task async_create(void (*entry)(void*), void* arg);

// Await a task (blocks until complete)
void async_await(async_task task);

// Check if task is complete
bool async_is_complete(async_task task);

// Run all pending tasks
void async_run_all(void);

// === Simple scheduler ===

#define MAX_TASKS 64

typedef struct {
    co_t coroutine;
    bool finished;
    void* result;
} Task;

extern Task task_list[MAX_TASKS];
extern int task_count;

// Schedule a task to run
int schedule_task(co_t co);

// Run one iteration of the scheduler
void scheduler_tick(void);

#endif // CORTEX_ASYNC_H
