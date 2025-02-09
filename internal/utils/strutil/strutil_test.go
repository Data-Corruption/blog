package strutil

import (
	"strings"
	"testing"
)

func TestPad(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		length   int
		expected string
	}{
		{"Shorter string", "hello", 10, "hello     "},
		{"Equal length", "hello", 5, "hello"},
		{"Longer string", "hello world", 5, "hello world"},
		{"Empty string", "", 3, "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Pad(tt.input, tt.length)
			if result != tt.expected {
				t.Errorf("Pad(%q, %d) = %q; expected %q", tt.input, tt.length, result, tt.expected)
			}
		})
	}
}

func TestRandom(t *testing.T) {
	// test various byte lengths
	tests := []struct {
		name string
		n    int
	}{
		{"Zero bytes", 0},
		{"One byte", 1},
		{"Sixteen bytes", 16},
		{"Thirty-two bytes", 32},
	}

	// The expected encoded length is ((n+2)/3)*4.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := Random(tt.n)
			if err != nil {
				t.Fatalf("Random(%d) returned error: %v", tt.n, err)
			}
			expectedLen := ((tt.n + 2) / 3) * 4
			if len(s) != expectedLen {
				t.Errorf("Random(%d) = %q has length %d; expected %d", tt.n, s, len(s), expectedLen)
			}
			// Verify the output is URL-safe (i.e. it should not contain '+' or '/')
			if strings.ContainsAny(s, "+/") {
				t.Errorf("Random(%d) = %q contains '+' or '/'", tt.n, s)
			}
		})
	}

	// Check that two different calls (with non-zero byte count) return different results.
	s1, err := Random(16)
	if err != nil {
		t.Fatalf("Random(16) returned error: %v", err)
	}
	s2, err := Random(16)
	if err != nil {
		t.Fatalf("Random(16) returned error: %v", err)
	}
	if s1 == s2 && s1 != "" {
		t.Errorf("Two calls to Random(16) produced the same result: %q", s1)
	}
}
