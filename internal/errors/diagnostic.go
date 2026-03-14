// Package errors provides structured diagnostics for the Cortex compiler.
// Use for lexer, parser, semantic, and codegen errors with line/column and suggestions.
package errors

import (
	"fmt"
	"strings"
)

// Severity indicates diagnostic level.
type Severity int

const (
	SeverityError   Severity = iota
	SeverityWarning
)

// Code is a stable error code for tooling and docs.
type Code string

const (
	ErrLexUnexpectedChar   Code = "E001"
	ErrParseUnexpected     Code = "E002"
	ErrParseExpected       Code = "E003"
	ErrSemanticRedef       Code = "E004"
	ErrSemanticTypeMismatch Code = "E005"
	ErrSemanticUnknown     Code = "E006"
	ErrArrayBounds         Code = "E007"
)

// Diagnostic is a single compiler message with location and optional suggestion.
type Diagnostic struct {
	Severity  Severity
	Code      Code
	Line      int
	Column    int
	Message   string
	Suggestion string
	File      string
}

// Collector gathers diagnostics and implements error interface.
type Collector struct {
	list []Diagnostic
}

// NewCollector returns a new diagnostic collector.
func NewCollector() *Collector {
	return &Collector{}
}

// Add appends a diagnostic.
func (c *Collector) Add(d Diagnostic) {
	c.list = append(c.list, d)
}

// AddError is a shorthand for adding an error at a location.
func (c *Collector) AddError(code Code, line, column int, message, suggestion string) {
	c.Add(Diagnostic{
		Severity:  SeverityError,
		Code:      code,
		Line:      line,
		Column:    column,
		Message:   message,
		Suggestion: suggestion,
	})
}

// Err returns an error if any diagnostics are errors.
func (c *Collector) Err() error {
	if c.HasErrors() {
		return c
	}
	return nil
}

// HasErrors returns true if any diagnostic has SeverityError.
func (c *Collector) HasErrors() bool {
	for _, d := range c.list {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}

// Error implements error.
func (c *Collector) Error() string {
	return c.String()
}

// String returns formatted diagnostics (one per line).
func (c *Collector) String() string {
	var b strings.Builder
	for _, d := range c.list {
		b.WriteString(d.String())
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

// List returns a copy of the diagnostic list.
func (c *Collector) List() []Diagnostic {
	out := make([]Diagnostic, len(c.list))
	copy(out, c.list)
	return out
}

// String formats one diagnostic for display.
func (d Diagnostic) String() string {
	loc := ""
	if d.Line > 0 {
		loc = fmt.Sprintf("%d:%d", d.Line, d.Column)
		if d.File != "" {
			loc = d.File + ":" + loc
		}
		if loc != "" {
			loc = loc + ": "
		}
	}
	sev := "error"
	if d.Severity == SeverityWarning {
		sev = "warning"
	}
	msg := fmt.Sprintf("%s%s [%s] %s", loc, sev, d.Code, d.Message)
	if d.Suggestion != "" {
		msg += "\n  hint: " + d.Suggestion
	}
	return msg
}
