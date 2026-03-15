// Cortex Async Runtime - libco wrapper for coroutines
// Implements a simple cooperative scheduler

#include "async.h"
#include <stdlib.h>
#include <string.h>

// libco is a tiny coroutine library
// We embed a minimal implementation here for portability

// === Platform-specific context switching ===
#if defined(_WIN32) || defined(__CYGWIN__)
    #define CO_USE_WINDOWS_FIBERS
    #include <windows.h>
#else
    #define CO_USE_UCONTEXT
    #include <ucontext.h>
#endif

// Coroutine state
struct co_state {
    #ifdef CO_USE_WINDOWS_FIBERS
    LPVOID fiber;
    LPVOID caller;
    #else
    ucontext_t ctx;
    ucontext_t* caller;
    void* stack;
    #endif
    void (*entry)(void*);
    void* arg;
    bool finished;
};

// Thread-local current coroutine
#ifdef _MSC_VER
    __declspec(thread) struct co_state* current_co = NULL;
#else
    __thread struct co_state* current_co = NULL;
#endif

// Track if main thread is converted to fiber
#ifdef CO_USE_WINDOWS_FIBERS
static int main_thread_is_fiber = 0;
static LPVOID main_fiber = NULL;

static void ensure_main_fiber(void) {
    if (!main_thread_is_fiber) {
        main_fiber = ConvertThreadToFiber(NULL);
        if (main_fiber) {
            main_thread_is_fiber = 1;
        }
    }
}
#endif

// Default stack size
#define DEFAULT_STACK_SIZE (64 * 1024)

// === libco implementation ===

#ifdef CO_USE_WINDOWS_FIBERS
// Windows fiber-based implementation

static void __stdcall co_entry(void* arg) {
    struct co_state* co = (struct co_state*)arg;
    current_co = co;
    co->entry(co->arg);
    co->finished = true;
    current_co = NULL;
    SwitchToFiber(co->caller);
}

co_t co_create(void (*entry)(void*), void* arg, int stack_size) {
    struct co_state* co = (struct co_state*)malloc(sizeof(struct co_state));
    if (!co) return NULL;
    
    co->entry = entry;
    co->arg = arg;
    co->finished = false;
    co->caller = NULL;
    
    if (stack_size == 0) stack_size = DEFAULT_STACK_SIZE;
    co->fiber = CreateFiber(stack_size, co_entry, co);
    
    if (!co->fiber) {
        free(co);
        return NULL;
    }
    
    return co;
}

void co_resume(co_t handle) {
    struct co_state* co = (struct co_state*)handle;
    if (!co || co->finished) return;
    
    ensure_main_fiber();
    co->caller = main_fiber;
    
    current_co = co;
    SwitchToFiber(co->fiber);
}

void co_yield(void) {
    if (!current_co) return;
    struct co_state* co = current_co;
    current_co = NULL;
    SwitchToFiber(co->caller);
}

void co_free(co_t handle) {
    struct co_state* co = (struct co_state*)handle;
    if (!co) return;
    if (co->fiber) DeleteFiber(co->fiber);
    free(co);
}

#else
// POSIX ucontext-based implementation

static void co_entry_wrapper(int a, int b, int c, int d) {
    // Reconstruct pointer from integers
    void* arg = (void*)((size_t)a | ((size_t)b << 16) | ((size_t)c << 32) | ((size_t)d << 48));
    struct co_state* co = current_co;
    co->entry(arg);
    co->finished = true;
    co_yield();
}

co_t co_create(void (*entry)(void*), void* arg, int stack_size) {
    struct co_state* co = (struct co_state*)malloc(sizeof(struct co_state));
    if (!co) return NULL;
    
    co->entry = entry;
    co->arg = arg;
    co->finished = false;
    co->caller = NULL;
    
    if (stack_size == 0) stack_size = DEFAULT_STACK_SIZE;
    co->stack = malloc(stack_size);
    
    if (!co->stack) {
        free(co);
        return NULL;
    }
    
    getcontext(&co->ctx);
    co->ctx.uc_stack.ss_sp = co->stack;
    co->ctx.uc_stack.ss_size = stack_size;
    co->ctx.uc_link = NULL;
    
    // Split pointer into 4 ints for makecontext
    size_t ptr = (size_t)arg;
    int a = ptr & 0xFFFF;
    int b = (ptr >> 16) & 0xFFFF;
    int c = (ptr >> 32) & 0xFFFF;
    int d = (ptr >> 48) & 0xFFFF;
    
    makecontext(&co->ctx, (void(*)())co_entry_wrapper, 4, a, b, c, d);
    
    return co;
}

void co_resume(co_t handle) {
    struct co_state* co = (struct co_state*)handle;
    if (!co || co->finished) return;
    
    static ucontext_t main_ctx;
    co->caller = &main_ctx;
    current_co = co;
    swapcontext(&main_ctx, &co->ctx);
}

void co_yield(void) {
    if (!current_co) return;
    struct co_state* co = current_co;
    current_co = NULL;
    swapcontext(&co->ctx, co->caller);
}

void co_free(co_t handle) {
    struct co_state* co = (struct co_state*)handle;
    if (!co) return;
    if (co->stack) free(co->stack);
    free(co);
}

#endif

co_t co_current(void) {
    return current_co;
}

bool co_finished(co_t handle) {
    struct co_state* co = (struct co_state*)handle;
    return co ? co->finished : true;
}

// === Cortex async API ===

Task task_list[MAX_TASKS];
int task_count = 0;

int schedule_task(co_t co) {
    if (task_count >= MAX_TASKS) return -1;
    
    task_list[task_count].coroutine = co;
    task_list[task_count].finished = false;
    task_list[task_count].result = NULL;
    
    return task_count++;
}

async_task async_create(void (*entry)(void*), void* arg) {
    co_t co = co_create(entry, arg, 0);
    if (!co) return -1;
    return schedule_task(co);
}

void async_await(async_task task) {
    if (task < 0 || task >= task_count) return;
    
    Task* t = &task_list[task];
    while (!t->finished && !co_finished(t->coroutine)) {
        co_resume(t->coroutine);
    }
    t->finished = true;
}

bool async_is_complete(async_task task) {
    if (task < 0 || task >= task_count) return true;
    return task_list[task].finished || co_finished(task_list[task].coroutine);
}

void async_run_all(void) {
    bool any_running = true;
    while (any_running) {
        any_running = false;
        for (int i = 0; i < task_count; i++) {
            if (!task_list[i].finished && !co_finished(task_list[i].coroutine)) {
                co_resume(task_list[i].coroutine);
                any_running = true;
            }
        }
    }
}

void scheduler_tick(void) {
    for (int i = 0; i < task_count; i++) {
        if (!task_list[i].finished && !co_finished(task_list[i].coroutine)) {
            co_resume(task_list[i].coroutine);
        }
    }
}
