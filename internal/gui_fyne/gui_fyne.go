// Package gui_fyne provides a Fyne-based GUI runtime for Cortex.
// This enables native GUI development with widgets, dialogs, and event handling.
package gui_fyne

import (
	"sync"
)

// GUIEvent represents an event from a GUI widget.
type GUIEvent struct {
	Type         int
	SourceHandle int64
	Data         interface{}
}

// Event types
const (
	EventClick = iota
	EventChange
	EventSubmit
	EventClose
	EventResize
	EventFocus
	EventBlur
)

// WidgetRegistry maps handles to widget implementations.
type WidgetRegistry struct {
	mu      sync.RWMutex
	widgets map[int64]interface{}
	nextID  int64
}

// NewWidgetRegistry creates a new widget registry.
func NewWidgetRegistry() *WidgetRegistry {
	return &WidgetRegistry{
		widgets: make(map[int64]interface{}),
		nextID:  1,
	}
}

// Register adds a widget and returns its handle.
func (r *WidgetRegistry) Register(widget interface{}) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.nextID
	r.nextID++
	r.widgets[id] = widget
	return id
}

// Get retrieves a widget by handle.
func (r *WidgetRegistry) Get(handle int64) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.widgets[handle]
	return w, ok
}

// Remove deletes a widget from the registry.
func (r *WidgetRegistry) Remove(handle int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.widgets, handle)
}

// Global registry for all GUI widgets
var Registry = NewWidgetRegistry()

// CallbackStore stores C function pointers for event callbacks.
type CallbackStore struct {
	mu        sync.RWMutex
	callbacks map[int64]uintptr // handle -> C function pointer
}

// NewCallbackStore creates a new callback store.
func NewCallbackStore() *CallbackStore {
	return &CallbackStore{
		callbacks: make(map[int64]uintptr),
	}
}

// Set stores a callback for a widget handle.
func (s *CallbackStore) Set(handle int64, fnPtr uintptr) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callbacks[handle] = fnPtr
}

// Get retrieves a callback by handle.
func (s *CallbackStore) Get(handle int64) (uintptr, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fn, ok := s.callbacks[handle]
	return fn, ok
}

// Global callback store
var Callbacks = NewCallbackStore()
