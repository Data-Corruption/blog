package blog

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestConsoleLogging(t *testing.T) {
	// Backup the original stdout.
	origStdout := os.Stdout

	// Create a pipe to capture stdout.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = origStdout
	}()

	// Initialize logger with console enabled and file logging disabled.
	if err := Init("", INFO, false, true); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	// Ensure the logger is cleaned up after the test.
	defer Cleanup(1 * time.Second)

	// Log a test message.
	testMsg := "Hello, stdout logging!"
	if err := Info(testMsg); err != nil {
		t.Errorf("Error logging info: %v", err)
	}

	// Synchronously flush the log buffer.
	if err := SyncFlush(1 * time.Second); err != nil {
		t.Errorf("Error flushing logs: %v", err)
	}

	// Give a moment for the asynchronous logging to complete.
	time.Sleep(50 * time.Millisecond)

	// Close the writer so we can read the captured output.
	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to read captured output: %v", err)
	}
	output := buf.String()

	if !strings.Contains(output, testMsg) {
		t.Errorf("Expected stdout to contain %q, got %q", testMsg, output)
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{NONE, "NONE"},
		{ERROR, "ERROR"},
		{WARN, "WARN"},
		{INFO, "INFO"},
		{DEBUG, "DEBUG"},
		{FATAL, "FATAL"},
	}

	for _, tc := range tests {
		got := tc.level.String()
		if got != tc.expected {
			t.Errorf("Level(%d).String() = %q, want %q", tc.level, got, tc.expected)
		}
	}
}

func TestLevelFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
		wantErr  bool
	}{
		// Valid (case-insensitive) strings.
		{"none", NONE, false},
		{"None", NONE, false},
		{"NONE", NONE, false},
		{"error", ERROR, false},
		{"Error", ERROR, false},
		{"ERROR", ERROR, false},
		{"warn", WARN, false},
		{"Warn", WARN, false},
		{"WARN", WARN, false},
		{"info", INFO, false},
		{"Info", INFO, false},
		{"INFO", INFO, false},
		{"debug", DEBUG, false},
		{"Debug", DEBUG, false},
		{"DEBUG", DEBUG, false},
		{"fatal", FATAL, false},
		{"Fatal", FATAL, false},
		{"FATAL", FATAL, false},
		// Invalid input.
		{"invalid", NONE, true},
	}

	for _, tc := range tests {
		// Initialize to a known value different from the expected valid result.
		var lvl Level = INFO

		err := lvl.FromString(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("FromString(%q): expected error but got none", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("FromString(%q): unexpected error: %v", tc.input, err)
			}
			// Ensure the conversion is case-insensitive.
			// For comparison, we use the String() method.
			if strings.ToUpper(lvl.String()) != tc.expected.String() {
				t.Errorf("FromString(%q) set level = %v, want %v", tc.input, lvl, tc.expected)
			}
		}
	}
}
