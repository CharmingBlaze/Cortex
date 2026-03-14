// bridge.go - C ↔ Go bridge for the GUI system
//
// This file implements the low-level bridge that allows Cortex (C code) to call
// Go functions and vice versa. It uses cgo to export functions that can be called
// from C, and provides mechanisms for Go to call back into Cortex.

package gui_fyne

/*
#include <stdint.h>
#include <stdlib.h>

// Forward declaration - implemented in Cortex runtime
typedef void (*cortex_event_callback)(int event_type, int64_t source_handle, void* data);

// Callback trampoline that Go can call to invoke Cortex callbacks
static inline void call_cortex_callback(uintptr_t fn_ptr, int event_type, int64_t source, void* data) {
    ((cortex_event_callback)fn_ptr)(event_type, source, data);
}

// External callback handler - implemented in gui_runtime.c
extern void cortex_gui_event_handler(int event_type, int64_t source, void* data);

// Type aliases for CGO compatibility
typedef uint8_t uint8;
typedef float cfloat;
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// ============================================================================
// Exported C Functions - Called from gui_runtime.c
// ============================================================================

//export CortexGUI_Version
func CortexGUI_Version() C.int {
	return C.int(1)
}

//export CortexGUI_CreateWindow
func CortexGUI_CreateWindow(title *C.char, width, height C.int) C.int64_t {
	titleStr := C.GoString(title)
	if err := InitGUI(titleStr, float32(width), float32(height)); err != nil {
		return 0
	}
	return C.int64_t(1) // Window handle
}

//export CortexGUI_ShowWindow
func CortexGUI_ShowWindow(handle C.int64_t) {
	// Window is shown via RunGUI
}

//export CortexGUI_HideWindow
func CortexGUI_HideWindow(handle C.int64_t) {
	// TODO: Implement window hiding
}

//export CortexGUI_CloseWindow
func CortexGUI_CloseWindow(handle C.int64_t) {
	QuitGUI()
}

//export CortexGUI_SetWindowTitle
func CortexGUI_SetWindowTitle(handle C.int64_t, title *C.char) {
	SetWindowTitle(C.GoString(title))
}

//export CortexGUI_CenterWindow
func CortexGUI_CenterWindow(handle C.int64_t) {
	CenterWindow()
}

//export CortexGUI_SetWindowFixedSize
func CortexGUI_SetWindowFixedSize(handle C.int64_t, fixed C.int) {
	if fixed != 0 {
		SetWindowFixedSize()
	}
}

//export CortexGUI_FullscreenWindow
func CortexGUI_FullscreenWindow(handle C.int64_t, fullscreen C.int) {
	// TODO: Implement fullscreen
}

//export CortexGUI_SetWindowContent
func CortexGUI_SetWindowContent(window, content C.int64_t) {
	SetWindowContent(int64(content))
}

//export CortexGUI_CreateLabel
func CortexGUI_CreateLabel(text *C.char) C.int64_t {
	return C.int64_t(CreateLabel(C.GoString(text)))
}

//export CortexGUI_SetLabelText
func CortexGUI_SetLabelText(handle C.int64_t, text *C.char) {
	SetLabelText(int64(handle), C.GoString(text))
}

//export CortexGUI_CreateButton
func CortexGUI_CreateButton(label *C.char, callbackID C.int64_t) C.int64_t {
	// Store callback ID for later invocation
	cb, _ := Callbacks.Get(int64(callbackID))
	return C.int64_t(CreateButton(C.GoString(label), cb))
}

//export CortexGUI_CreateEntry
func CortexGUI_CreateEntry(placeholder *C.char, callbackID C.int64_t) C.int64_t {
	return C.int64_t(CreateEntry(C.GoString(placeholder)))
}

//export CortexGUI_GetEntryText
func CortexGUI_GetEntryText(handle C.int64_t) *C.char {
	text := GetEntryText(int64(handle))
	return C.CString(text)
}

//export CortexGUI_SetEntryText
func CortexGUI_SetEntryText(handle C.int64_t, text *C.char) {
	SetEntryText(int64(handle), C.GoString(text))
}

//export CortexGUI_CreateTextArea
func CortexGUI_CreateTextArea(placeholder *C.char, callbackID C.int64_t) C.int64_t {
	return C.int64_t(CreateTextArea(C.GoString(placeholder)))
}

//export CortexGUI_CreateCheck
func CortexGUI_CreateCheck(label *C.char, callbackID C.int64_t) C.int64_t {
	return C.int64_t(CreateCheckBox(C.GoString(label), false))
}

//export CortexGUI_GetCheckState
func CortexGUI_GetCheckState(handle C.int64_t) C.int {
	if GetCheckBoxState(int64(handle)) {
		return 1
	}
	return 0
}

//export CortexGUI_SetCheckState
func CortexGUI_SetCheckState(handle C.int64_t, checked C.int) {
	SetCheckBoxState(int64(handle), checked != 0)
}

//export CortexGUI_CreateSlider
func CortexGUI_CreateSlider(min, max, value C.double, callbackID C.int64_t) C.int64_t {
	return C.int64_t(CreateSlider(float64(min), float64(max), float64(value)))
}

//export CortexGUI_GetSliderValue
func CortexGUI_GetSliderValue(handle C.int64_t) C.double {
	return C.double(GetSliderValue(int64(handle)))
}

//export CortexGUI_SetSliderValue
func CortexGUI_SetSliderValue(handle C.int64_t, value C.double) {
	SetSliderValue(int64(handle), float64(value))
}

//export CortexGUI_CreateProgress
func CortexGUI_CreateProgress() C.int64_t {
	return C.int64_t(CreateProgressBar())
}

//export CortexGUI_SetProgressValue
func CortexGUI_SetProgressValue(handle C.int64_t, value C.double) {
	SetProgressValue(int64(handle), float64(value))
}

//export CortexGUI_CreateImage
func CortexGUI_CreateImage(filepath *C.char) C.int64_t {
	return C.int64_t(CreateImage(C.GoString(filepath)))
}

//export CortexGUI_SetImageFill
func CortexGUI_SetImageFill(handle C.int64_t, fillMode C.int) {
	SetImageFill(int64(handle), int(fillMode))
}

//export CortexGUI_CreateRectangle
func CortexGUI_CreateRectangle(r, g, b, a C.uchar) C.int64_t {
	color := fmt.Sprintf("#%02X%02X%02X", uint8(r), uint8(g), uint8(b))
	return C.int64_t(CreateRectangle(color))
}

//export CortexGUI_CreateCircle
func CortexGUI_CreateCircle(r, g, b, a C.uchar) C.int64_t {
	color := fmt.Sprintf("#%02X%02X%02X", uint8(r), uint8(g), uint8(b))
	return C.int64_t(CreateCircle(color))
}

//export CortexGUI_CreateLine
func CortexGUI_CreateLine(x1, y1, x2, y2 C.float) C.int64_t {
	return C.int64_t(CreateLine(float32(x1), float32(y1), float32(x2), float32(y2), "white"))
}

//export CortexGUI_SetLineColor
func CortexGUI_SetLineColor(handle C.int64_t, r, g, b, a C.uchar) {
	color := fmt.Sprintf("#%02X%02X%02X", uint8(r), uint8(g), uint8(b))
	SetFillColor(int64(handle), color)
}

//export CortexGUI_CreateVBox
func CortexGUI_CreateVBox(handles *C.int64_t, count C.int) C.int64_t {
	goHandles := make([]int64, int(count))
	for i := 0; i < int(count); i++ {
		goHandles[i] = int64(*(*C.int64_t)(unsafe.Pointer(uintptr(unsafe.Pointer(handles)) + uintptr(i)*8)))
	}
	return C.int64_t(CreateVBox(goHandles))
}

//export CortexGUI_CreateHBox
func CortexGUI_CreateHBox(handles *C.int64_t, count C.int) C.int64_t {
	goHandles := make([]int64, int(count))
	for i := 0; i < int(count); i++ {
		goHandles[i] = int64(*(*C.int64_t)(unsafe.Pointer(uintptr(unsafe.Pointer(handles)) + uintptr(i)*8)))
	}
	return C.int64_t(CreateHBox(goHandles))
}

//export CortexGUI_CreateGrid
func CortexGUI_CreateGrid(columns C.int, handles *C.int64_t, count C.int) C.int64_t {
	goHandles := make([]int64, int(count))
	for i := 0; i < int(count); i++ {
		goHandles[i] = int64(*(*C.int64_t)(unsafe.Pointer(uintptr(unsafe.Pointer(handles)) + uintptr(i)*8)))
	}
	return C.int64_t(CreateGrid(goHandles, int(columns)))
}

//export CortexGUI_AddToContainer
func CortexGUI_AddToContainer(container, widget C.int64_t) {
	// Containers are built at creation time
}

//export CortexGUI_ShowInfoDialog
func CortexGUI_ShowInfoDialog(window C.int64_t, title, message *C.char) {
	ShowInfoDialog(C.GoString(title), C.GoString(message))
}

//export CortexGUI_ShowErrorDialog
func CortexGUI_ShowErrorDialog(window C.int64_t, title, message *C.char) {
	ShowErrorDialog(C.GoString(title), C.GoString(message))
}

//export CortexGUI_ShowConfirmDialog
func CortexGUI_ShowConfirmDialog(window C.int64_t, title, message *C.char, callbackID C.int64_t) {
	cb, _ := Callbacks.Get(int64(callbackID))
	ShowConfirmDialog(C.GoString(title), C.GoString(message), cb)
}

//export CortexGUI_ShowFileOpenDialog
func CortexGUI_ShowFileOpenDialog(window, callbackID C.int64_t) {
	// TODO: Implement file dialogs
}

//export CortexGUI_ShowFileSaveDialog
func CortexGUI_ShowFileSaveDialog(window, callbackID C.int64_t) {
	// TODO: Implement file dialogs
}

//export CortexGUI_Run
func CortexGUI_Run() {
	RunGUI()
}

//export CortexGUI_Quit
func CortexGUI_Quit() {
	QuitGUI()
}

//export CortexGUI_Refresh
func CortexGUI_Refresh(handle C.int64_t) {
	Refresh(int64(handle))
}

//export CortexGUI_Resize
func CortexGUI_Resize(handle C.int64_t, width, height C.float) {
	Resize(int64(handle), float32(width), float32(height))
}

//export CortexGUI_Move
func CortexGUI_Move(handle C.int64_t, x, y C.float) {
	Move(int64(handle), float32(x), float32(y))
}

//export CortexGUI_Enable
func CortexGUI_Enable(handle C.int64_t) {
	Enable(int64(handle))
}

//export CortexGUI_Disable
func CortexGUI_Disable(handle C.int64_t) {
	Disable(int64(handle))
}

//export CortexGUI_IsEnabled
func CortexGUI_IsEnabled(handle C.int64_t) C.int {
	if IsEnabled(int64(handle)) {
		return 1
	}
	return 0
}

//export CortexGUI_FreeString
func CortexGUI_FreeString(str *C.char) {
	C.free(unsafe.Pointer(str))
}

// ============================================================================
// Internal Helper Functions
// ============================================================================

// InvokeCortexCallback calls a C function pointer from Go
// This is the bridge that routes events from Fyne back to Cortex lambdas
func InvokeCortexCallback(cFuncPtr uintptr, event GUIEvent) {
	if cFuncPtr == 0 {
		return
	}

	// Convert event data to something C can understand
	var dataPtr unsafe.Pointer

	switch d := event.Data.(type) {
	case string:
		// Allocate C string (caller must free or we leak - handled by runtime)
		dataPtr = unsafe.Pointer(C.CString(d))
	case bool:
		// Store bool as int
		val := C.int(0)
		if d {
			val = 1
		}
		dataPtr = unsafe.Pointer(&val)
	case int:
		val := C.int64_t(d)
		dataPtr = unsafe.Pointer(&val)
	case int64:
		val := C.int64_t(d)
		dataPtr = unsafe.Pointer(&val)
	case float64:
		// Store as double
		val := C.double(d)
		dataPtr = unsafe.Pointer(&val)
	default:
		dataPtr = nil
	}

	// Call the C callback trampoline
	C.call_cortex_callback(
		C.uintptr_t(cFuncPtr),
		C.int(event.Type),
		C.int64_t(event.SourceHandle),
		dataPtr,
	)
}

// ConvertCString converts a C string to Go string and frees the C memory.
func ConvertCString(cstr *C.char) string {
	if cstr == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
}

// CString allocates a C string from a Go string.
func CString(s string) *C.char {
	return C.CString(s)
}

// FreeCString frees a C string.
func FreeCString(cstr *C.char) {
	C.free(unsafe.Pointer(cstr))
}

// GetCInt returns a C.int value.
func GetCInt(v int) C.int {
	return C.int(v)
}

// GetCInt64 returns a C.int64_t value.
func GetCInt64(v int64) C.int64_t {
	return C.int64_t(v)
}

// GoInt converts C.int to Go int.
func GoInt(v C.int) int {
	return int(v)
}

// GoInt64 converts C.int64_t to Go int64.
func GoInt64(v C.int64_t) int64 {
	return int64(v)
}

// GoString converts C string to Go string.
func GoString(cstr *C.char) string {
	return C.GoString(cstr)
}
