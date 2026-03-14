// api.go - High-level GUI API for Cortex
//
// This file provides the main GUI functions that Cortex code can call.
// It wraps Fyne widgets and provides a simple interface for GUI development.

package gui_fyne

import (
	"fmt"
	"image/color"
	"runtime"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// GUIApp holds the main Fyne application and window state.
type GUIApp struct {
	app    fyne.App
	window fyne.Window
}

// Global GUI application instance
var currentApp *GUIApp

func init() {
	// Lock the main goroutine to the OS thread for GUI operations
	runtime.LockOSThread()
}

// InitGUI initializes the GUI system. Must be called before any other GUI functions.
func InitGUI(title string, width, height float32) error {
	if currentApp != nil {
		return fmt.Errorf("GUI already initialized")
	}

	currentApp = &GUIApp{
		app: app.NewWithID("cortex.gui"),
	}
	currentApp.window = currentApp.app.NewWindow(title)
	currentApp.window.Resize(fyne.NewSize(width, height))

	return nil
}

// RunGUI starts the GUI event loop. This blocks until the window is closed.
func RunGUI() {
	if currentApp == nil {
		fmt.Println("GUI not initialized - call InitGUI first")
		return
	}
	// Lock this goroutine to the OS thread - required for GUI
	runtime.LockOSThread()
	currentApp.window.ShowAndRun()
}

// QuitGUI closes the GUI application.
func QuitGUI() {
	if currentApp != nil {
		currentApp.app.Quit()
		currentApp = nil
	}
}

// Window functions

// SetWindowTitle sets the window title.
func SetWindowTitle(title string) {
	if currentApp != nil {
		currentApp.window.SetTitle(title)
	}
}

// CenterWindow centers the window on screen.
func CenterWindow() {
	if currentApp != nil {
		currentApp.window.CenterOnScreen()
	}
}

// SetWindowFixedSize makes the window non-resizable.
func SetWindowFixedSize() {
	if currentApp != nil {
		currentApp.window.SetFixedSize(true)
	}
}

// SetWindowContent sets the main content of the window.
func SetWindowContent(handle int64) {
	if currentApp == nil {
		return
	}
	if obj, ok := Registry.Get(handle); ok {
		if w, ok := obj.(fyne.CanvasObject); ok {
			currentApp.window.SetContent(w)
		}
	}
}

// Widget creation functions

// CreateLabel creates a text label widget.
func CreateLabel(text string) int64 {
	label := widget.NewLabel(text)
	return Registry.Register(label)
}

// SetLabelText updates the text of a label.
func SetLabelText(handle int64, text string) {
	if obj, ok := Registry.Get(handle); ok {
		if label, ok := obj.(*widget.Label); ok {
			label.SetText(text)
		}
	}
}

// CreateButton creates a button widget with a callback.
func CreateButton(text string, callback uintptr) int64 {
	btn := widget.NewButton(text, func() {
		if callback != 0 {
			InvokeCortexCallback(callback, GUIEvent{
				Type:         EventClick,
				SourceHandle: 0,
				Data:         text,
			})
		}
	})
	handle := Registry.Register(btn)
	if callback != 0 {
		Callbacks.Set(handle, callback)
	}
	return handle
}

// CreateEntry creates a text entry widget.
func CreateEntry(placeholder string) int64 {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)
	return Registry.Register(entry)
}

// GetEntryText gets the text from an entry widget.
func GetEntryText(handle int64) string {
	if obj, ok := Registry.Get(handle); ok {
		if entry, ok := obj.(*widget.Entry); ok {
			return entry.Text
		}
	}
	return ""
}

// SetEntryText sets the text in an entry widget.
func SetEntryText(handle int64, text string) {
	if obj, ok := Registry.Get(handle); ok {
		if entry, ok := obj.(*widget.Entry); ok {
			entry.SetText(text)
		}
	}
}

// CreateTextArea creates a multi-line text entry.
func CreateTextArea(placeholder string) int64 {
	entry := widget.NewMultiLineEntry()
	entry.SetPlaceHolder(placeholder)
	return Registry.Register(entry)
}

// CreateCheckBox creates a checkbox widget.
func CreateCheckBox(text string, checked bool) int64 {
	cb := widget.NewCheck(text, func(b bool) {})
	cb.SetChecked(checked)
	return Registry.Register(cb)
}

// GetCheckBoxState returns whether the checkbox is checked.
func GetCheckBoxState(handle int64) bool {
	if obj, ok := Registry.Get(handle); ok {
		if cb, ok := obj.(*widget.Check); ok {
			return cb.Checked
		}
	}
	return false
}

// SetCheckBoxState sets the checkbox state.
func SetCheckBoxState(handle int64, checked bool) {
	if obj, ok := Registry.Get(handle); ok {
		if cb, ok := obj.(*widget.Check); ok {
			cb.SetChecked(checked)
		}
	}
}

// CreateSlider creates a slider widget.
func CreateSlider(min, max, value float64) int64 {
	slider := widget.NewSlider(float64(min), float64(max))
	slider.SetValue(float64(value))
	return Registry.Register(slider)
}

// GetSliderValue returns the current slider value.
func GetSliderValue(handle int64) float64 {
	if obj, ok := Registry.Get(handle); ok {
		if slider, ok := obj.(*widget.Slider); ok {
			return float64(slider.Value)
		}
	}
	return 0
}

// SetSliderValue sets the slider value.
func SetSliderValue(handle int64, value float64) {
	if obj, ok := Registry.Get(handle); ok {
		if slider, ok := obj.(*widget.Slider); ok {
			slider.SetValue(float64(value))
		}
	}
}

// Container functions

// CreateVBox creates a vertical box container.
func CreateVBox(handles []int64) int64 {
	var objects []fyne.CanvasObject
	for _, h := range handles {
		if obj, ok := Registry.Get(h); ok {
			if w, ok := obj.(fyne.CanvasObject); ok {
				objects = append(objects, w)
			}
		}
	}
	box := container.NewVBox(objects...)
	return Registry.Register(box)
}

// CreateHBox creates a horizontal box container.
func CreateHBox(handles []int64) int64 {
	var objects []fyne.CanvasObject
	for _, h := range handles {
		if obj, ok := Registry.Get(h); ok {
			if w, ok := obj.(fyne.CanvasObject); ok {
				objects = append(objects, w)
			}
		}
	}
	box := container.NewHBox(objects...)
	return Registry.Register(box)
}

// CreateGrid creates a grid container.
func CreateGrid(handles []int64, cols int) int64 {
	var objects []fyne.CanvasObject
	for _, h := range handles {
		if obj, ok := Registry.Get(h); ok {
			if w, ok := obj.(fyne.CanvasObject); ok {
				objects = append(objects, w)
			}
		}
	}
	grid := container.NewGridWrap(fyne.NewSize(100, 100), objects...)
	return Registry.Register(grid)
}

// Dialog functions

// ShowInfoDialog shows an information dialog.
func ShowInfoDialog(title, message string) {
	if currentApp != nil {
		dialog.ShowInformation(title, message, currentApp.window)
	}
}

// ShowErrorDialog shows an error dialog.
func ShowErrorDialog(title, message string) {
	if currentApp != nil {
		dialog.ShowError(fmt.Errorf("%s", message), currentApp.window)
	}
}

// ShowConfirmDialog shows a confirmation dialog.
func ShowConfirmDialog(title, message string, callback uintptr) {
	if currentApp != nil {
		dialog.ShowConfirm(title, message, func(b bool) {
			if callback != 0 {
				InvokeCortexCallback(callback, GUIEvent{
					Type:         EventClick,
					SourceHandle: 0,
					Data:         b,
				})
			}
		}, currentApp.window)
	}
}

// Drawing functions

// CreateRectangle creates a rectangle shape.
func CreateRectangle(color string) int64 {
	rect := canvas.NewRectangle(parseColor(color))
	return Registry.Register(rect)
}

// CreateCircle creates a circle shape.
func CreateCircle(color string) int64 {
	circle := canvas.NewCircle(parseColor(color))
	return Registry.Register(circle)
}

// CreateLine creates a line shape.
func CreateLine(x1, y1, x2, y2 float32, color string) int64 {
	line := canvas.NewLine(parseColor(color))
	line.Position1 = fyne.NewPos(x1, y1)
	line.Position2 = fyne.NewPos(x2, y2)
	return Registry.Register(line)
}

// SetFillColor sets the fill color of a shape.
func SetFillColor(handle int64, color string) {
	if obj, ok := Registry.Get(handle); ok {
		switch shape := obj.(type) {
		case *canvas.Rectangle:
			shape.FillColor = parseColor(color)
			shape.Refresh()
		case *canvas.Circle:
			shape.FillColor = parseColor(color)
			shape.Refresh()
		}
	}
}

// Helper functions

func parseColor(hex string) color.Color {
	// Default to white if empty
	if hex == "" {
		return color.White
	}

	// Remove # prefix if present
	hex = strings.TrimPrefix(hex, "#")

	// Parse hex color
	if len(hex) == 6 {
		r, _ := strconv.ParseUint(hex[0:2], 16, 8)
		g, _ := strconv.ParseUint(hex[2:4], 16, 8)
		b, _ := strconv.ParseUint(hex[4:6], 16, 8)
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
	}

	// Named colors
	switch strings.ToLower(hex) {
	case "red":
		return color.RGBA{R: 255, G: 0, B: 0, A: 255}
	case "green":
		return color.RGBA{R: 0, G: 255, B: 0, A: 255}
	case "blue":
		return color.RGBA{R: 0, G: 0, B: 255, A: 255}
	case "black":
		return color.Black
	case "white":
		return color.White
	case "yellow":
		return color.RGBA{R: 255, G: 255, B: 0, A: 255}
	case "cyan":
		return color.RGBA{R: 0, G: 255, B: 255, A: 255}
	case "magenta":
		return color.RGBA{R: 255, G: 0, B: 255, A: 255}
	case "gray", "grey":
		return color.RGBA{R: 128, G: 128, B: 128, A: 255}
	case "orange":
		return color.RGBA{R: 255, G: 165, B: 0, A: 255}
	case "purple":
		return color.RGBA{R: 128, G: 0, B: 128, A: 255}
	default:
		return color.White
	}
}

// Refresh refreshes a widget.
func Refresh(handle int64) {
	if obj, ok := Registry.Get(handle); ok {
		if w, ok := obj.(fyne.CanvasObject); ok {
			w.Refresh()
		}
	}
}

// Resize resizes a widget.
func Resize(handle int64, width, height float32) {
	if obj, ok := Registry.Get(handle); ok {
		if w, ok := obj.(fyne.CanvasObject); ok {
			w.Resize(fyne.NewSize(width, height))
		}
	}
}

// Move moves a widget.
func Move(handle int64, x, y float32) {
	if obj, ok := Registry.Get(handle); ok {
		if w, ok := obj.(fyne.CanvasObject); ok {
			w.Move(fyne.NewPos(x, y))
		}
	}
}

// Enable enables a widget.
func Enable(handle int64) {
	if obj, ok := Registry.Get(handle); ok {
		if w, ok := obj.(fyne.Disableable); ok {
			w.Enable()
		}
	}
}

// Disable disables a widget.
func Disable(handle int64) {
	if obj, ok := Registry.Get(handle); ok {
		if w, ok := obj.(fyne.Disableable); ok {
			w.Disable()
		}
	}
}

// IsEnabled checks if a widget is enabled.
func IsEnabled(handle int64) bool {
	if obj, ok := Registry.Get(handle); ok {
		if w, ok := obj.(fyne.Disableable); ok {
			return !w.Disabled()
		}
	}
	return false
}

// Image functions

// CreateImage creates an image widget from a file.
func CreateImage(path string) int64 {
	img := canvas.NewImageFromFile(path)
	img.FillMode = canvas.ImageFillContain
	return Registry.Register(img)
}

// SetImageFill sets how the image fills its container.
func SetImageFill(handle int64, mode int) {
	if obj, ok := Registry.Get(handle); ok {
		if img, ok := obj.(*canvas.Image); ok {
			switch mode {
			case 0:
				img.FillMode = canvas.ImageFillStretch
			case 1:
				img.FillMode = canvas.ImageFillContain
			case 2:
				img.FillMode = canvas.ImageFillOriginal
			}
			img.Refresh()
		}
	}
}

// Progress bar

// CreateProgressBar creates a progress bar widget.
func CreateProgressBar() int64 {
	pb := widget.NewProgressBar()
	return Registry.Register(pb)
}

// SetProgressValue sets the progress bar value.
func SetProgressValue(handle int64, value float64) {
	if obj, ok := Registry.Get(handle); ok {
		if pb, ok := obj.(*widget.ProgressBar); ok {
			pb.SetValue(float64(value))
		}
	}
}

// Spacer functions

// CreateVSpacer creates a vertical spacer.
func CreateVSpacer() int64 {
	spacer := layout.NewSpacer()
	return Registry.Register(container.NewVBox(spacer))
}

// CreateHSpacer creates a horizontal spacer.
func CreateHSpacer() int64 {
	spacer := layout.NewSpacer()
	return Registry.Register(container.NewHBox(spacer))
}
