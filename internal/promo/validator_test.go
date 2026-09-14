package promo

import (
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestFileIndex - checks coupon length and two-file presence rules.
func TestFileIndex(t *testing.T) {
	directory := t.TempDir()
	paths := []string{
		writeCoupons(t, directory, "one.gz", "HAPPYHRS", "HAPPYHRS", "SUPER100"),
		writeCoupons(t, directory, "two.gz", "xx HAPPYHRS yy", "FIFTYOFF"),
		writeCoupons(t, directory, "three.gz", "FIFTYOFF"),
	}
	index := NewValidator()
	if err := index.LoadFiles(context.Background(), paths); err != nil {
		t.Fatalf("LoadFiles() error = %v", err)
	}
	tests := []struct {
		name string
		code string
		want bool
	}{
		{name: "present in two files", code: "HAPPYHRS", want: true},
		{name: "second valid code", code: "FIFTYOFF", want: true},
		{name: "duplicate in only one file", code: "SUPER100", want: false},
		{name: "too short", code: "SHORT", want: false},
		{name: "too long", code: "LONGCOUPON11", want: false},
		{name: "unknown", code: "UNKNOWN1", want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := index.Validate(context.Background(), test.code); got != test.want {
				t.Fatalf("Validate(%q) = %v, want %v", test.code, got, test.want)
			}
		})
	}
}

// TestFileEngine - checks that valid coupons discount and invalid ones do not.
func TestFileEngine(t *testing.T) {
	index := NewValidator()
	index.counts = map[string]int{
		"HAPPYHRS": 2,
		"SUPER100": 1,
	}
	engine := NewFileEngine(index, PercentOff(0.10))
	if got := engine.Discount(context.Background(), "HAPPYHRS", 1330); got != 133 {
		t.Fatalf("valid coupon discount = %d, want 133", got)
	}
	if got := engine.Discount(context.Background(), "SUPER100", 1330); got != 0 {
		t.Fatalf("invalid coupon discount = %d, want 0", got)
	}
	if got := engine.Discount(context.Background(), "", 1330); got != 0 {
		t.Fatalf("empty coupon discount = %d, want 0", got)
	}
}

// writeCoupons - writes a gzip coupon dump containing the given lines.
func writeCoupons(t *testing.T, directory, name string, lines ...string) string {
	t.Helper()
	path := filepath.Join(directory, name)
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create coupon file: %v", err)
	}
	writer := gzip.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.Write([]byte(line + "\n")); err != nil {
			t.Fatalf("write coupon: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close coupon file: %v", err)
	}
	return path
}
