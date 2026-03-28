package binder

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSyntheticHeaderAST(t *testing.T) {
	if _, _, _, err := acquireHostConfig(); err != nil {
		t.Skip("host C toolchain / cpp not available:", err)
	}

	h := filepath.Join("testdata", "synthetic.h")
	b := NewBinder("synth")
	opt := ParseOptions{SkipPreprocess: true}
	if err := b.ParseHeaderWithOptions(h, opt); err != nil {
		t.Fatal(err)
	}
	fn, st, en, _ := b.Stats()
	if fn < 1 {
		t.Fatalf("expected at least one function, got %d", fn)
	}
	if st < 1 {
		t.Fatalf("expected at least one struct, got %d", st)
	}
	if en < 1 {
		t.Fatalf("expected at least one enum, got %d", en)
	}

	out := b.GenerateCortex()
	if !strings.Contains(out, "draw_point") {
		t.Fatalf("generated output missing draw_point:\n%s", out)
	}
	if !strings.Contains(out, "struct Point") {
		t.Fatalf("generated output missing struct Point:\n%s", out)
	}
}

func TestLegacyBindRegex(t *testing.T) {
	h := filepath.Join("testdata", "synthetic.h")
	b := NewBinder("synth")
	if err := b.ParseHeaderWithOptions(h, ParseOptions{LegacyBind: true}); err != nil {
		t.Fatal(err)
	}
	if len(b.functions) == 0 {
		t.Fatal("legacy bind expected at least one function")
	}
}
