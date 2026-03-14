package gui_fyne

import (
	"image/color"
	"testing"
)

func TestWidgetRegistry(t *testing.T) {
	r := NewWidgetRegistry()

	// Test registration
	widget := "test_widget"
	handle := r.Register(widget)

	if handle == 0 {
		t.Error("Handle should not be 0")
	}

	// Test retrieval
	retrieved, ok := r.Get(handle)
	if !ok {
		t.Error("Should find registered widget")
	}
	if retrieved != widget {
		t.Errorf("Expected %v, got %v", widget, retrieved)
	}

	// Test removal
	r.Remove(handle)
	_, ok = r.Get(handle)
	if ok {
		t.Error("Should not find removed widget")
	}
}

func TestCallbackStore(t *testing.T) {
	s := NewCallbackStore()

	// Test set/get
	handle := int64(1)
	fnPtr := uintptr(0x12345678)

	s.Set(handle, fnPtr)

	retrieved, ok := s.Get(handle)
	if !ok {
		t.Error("Should find registered callback")
	}
	if retrieved != fnPtr {
		t.Errorf("Expected %v, got %v", fnPtr, retrieved)
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		input    string
		expected color.RGBA
	}{
		{"#FF0000", color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"red", color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"green", color.RGBA{R: 0, G: 255, B: 0, A: 255}},
		{"blue", color.RGBA{R: 0, G: 0, B: 255, A: 255}},
		{"white", color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"black", color.RGBA{R: 0, G: 0, B: 0, A: 255}},
	}

	for _, tc := range tests {
		c := parseColor(tc.input)
		rgba := color.RGBAModel.Convert(c).(color.RGBA)
		if rgba != tc.expected {
			t.Errorf("parseColor(%q) = %v, want %v", tc.input, rgba, tc.expected)
		}
	}
}
