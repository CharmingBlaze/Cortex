package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
	Italic = "\033[3m"

	// Violet/Purple theme
	Violet      = "\033[38;5;141m"  // Light violet
	VioletBold  = "\033[1;38;5;141m"
	Purple      = "\033[38;5;98m"   // Medium purple
	Magenta     = "\033[38;5;198m"  // Pink magenta
	Lavender    = "\033[38;5;183m"  // Light lavender

	// Semantic colors
	Success     = "\033[38;5;114m"  // Green
	Warning     = "\033[38;5;214m"  // Orange
	Error       = "\033[38;5;203m"  // Red
	Info        = "\033[38;5;117m"  // Light blue
	Debug       = "\033[38;5;147m"  // Light purple

	// Progress bar colors
	ProgressBg  = "\033[48;5;60m"   // Purple background
	ProgressFg  = "\033[48;5;98m"   // Lighter purple
)

// ColorWriter provides colored output methods
type ColorWriter struct {
	stdout io.Writer
	stderr io.Writer
}

// NewColorWriter creates a new ColorWriter
func NewColorWriter() *ColorWriter {
	return &ColorWriter{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// PrintViolet prints in violet color
func (c *ColorWriter) PrintViolet(format string, args ...interface{}) {
	fmt.Fprintf(c.stdout, Violet+format+Reset+"\n", args...)
}

// PrintSuccess prints a success message in green
func (c *ColorWriter) PrintSuccess(format string, args ...interface{}) {
	fmt.Fprintf(c.stdout, Success+"✓ "+format+Reset+"\n", args...)
}

// PrintError prints an error message in red with violet accent
func (c *ColorWriter) PrintError(format string, args ...interface{}) {
	fmt.Fprintf(c.stderr, Error+"✗ "+VioletBold+"Error:"+Reset+Error+" "+format+Reset+"\n", args...)
}

// PrintWarning prints a warning message in orange
func (c *ColorWriter) PrintWarning(format string, args ...interface{}) {
	fmt.Fprintf(c.stderr, Warning+"⚠ "+format+Reset+"\n", args...)
}

// PrintInfo prints an info message in light blue
func (c *ColorWriter) PrintInfo(format string, args ...interface{}) {
	fmt.Fprintf(c.stdout, Info+"ℹ "+format+Reset+"\n", args...)
}

// PrintDebug prints a debug message in purple
func (c *ColorWriter) PrintDebug(format string, args ...interface{}) {
	fmt.Fprintf(c.stdout, Debug+"⚙ "+format+Reset+"\n", args...)
}

// PrintHeader prints a violet header with decoration
func (c *ColorWriter) PrintHeader(title string) {
	line := strings.Repeat("─", len(title)+4)
	fmt.Fprintf(c.stdout, "\n"+VioletBold+"  %s\n  │ %s │\n  %s"+Reset+"\n\n", line, title, line)
}

// PrintStep prints a compilation step in violet
func (c *ColorWriter) PrintStep(step, desc string) {
	fmt.Fprintf(c.stdout, Violet+"  %s"+Reset+" %s\n", step, desc)
}

// ProgressBar represents a progress bar
type ProgressBar struct {
	total     int
	current   int
	width     int
	desc      string
	startTime time.Time
	mu        sync.Mutex
	writer    io.Writer
	done      bool
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, desc string) *ProgressBar {
	return &ProgressBar{
		total:     total,
		current:   0,
		width:     30,
		desc:      desc,
		startTime: time.Now(),
		writer:    os.Stdout,
	}
}

// Add increments the progress bar
func (p *ProgressBar) Add(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current += n
	if p.current > p.total {
		p.current = p.total
	}
	p.render()
}

// Set sets the current progress
func (p *ProgressBar) Set(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = n
	if p.current > p.total {
		p.current = p.total
	}
	p.render()
}

// Complete marks the progress as done
func (p *ProgressBar) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.done = true
	p.current = p.total
	p.render()
	fmt.Fprintln(p.writer)
}

// render draws the progress bar
func (p *ProgressBar) render() {
	if p.done {
		return
	}
	
	percent := float64(p.current) / float64(p.total)
	if percent > 1.0 {
		percent = 1.0
	}
	
	filled := int(percent * float64(p.width))
	empty := p.width - filled
	
	// Build the bar
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	
	// Calculate elapsed time
	elapsed := time.Since(p.startTime).Seconds()
	
	// Build the line
	line := fmt.Sprintf("\r"+Violet+"  %s"+Reset+" ["+ProgressFg+"%s"+Reset+"] "+
		Violet+"%3.0f%%"+Reset+" "+Dim+"%d/%d"+Reset+" "+Lavender+"%.1fs"+Reset,
		p.desc, bar, percent*100, p.current, p.total, elapsed)
	
	fmt.Fprint(p.writer, line)
}

// Spinner shows an animated spinner for indeterminate progress
type Spinner struct {
	frames   []string
	current  int
	desc     string
	stopChan chan struct{}
	writer   io.Writer
	mu       sync.Mutex
}

// NewSpinner creates a new spinner
func NewSpinner(desc string) *Spinner {
	return &Spinner{
		frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		desc:     desc,
		stopChan: make(chan struct{}),
		writer:   os.Stdout,
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-s.stopChan:
				return
			case <-ticker.C:
				s.mu.Lock()
				frame := s.frames[s.current]
				s.current = (s.current + 1) % len(s.frames)
				fmt.Fprintf(s.writer, "\r"+Violet+"  %s "+Reset+"%s", frame, s.desc)
				s.mu.Unlock()
			}
		}
	}()
}

// Stop stops the spinner and shows completion
func (s *Spinner) Stop(success bool) {
	close(s.stopChan)
	time.Sleep(50 * time.Millisecond) // Let animation finish
	if success {
		fmt.Fprintf(s.writer, "\r"+Success+"  ✓ "+Reset+"%s\n", s.desc)
	} else {
		fmt.Fprintf(s.writer, "\r"+Error+"  ✗ "+Reset+"%s\n", s.desc)
	}
}

// Update changes the spinner description
func (s *Spinner) Update(desc string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.desc = desc
}

// Global color writer instance
var Colors = NewColorWriter()
