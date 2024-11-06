package blog

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"
)

// Helper functions ===========================================================

func normalInit(path string) (*Logger, *bytes.Buffer, error) {
	level := INFO
	buf := &bytes.Buffer{}
	cl := &ConsoleLogger{l: log.New(buf, "", 0)}
	instance, err := NewLogger(Config{Level: &level, DirectoryPath: &path, ConsoleOut: cl}, 255, 2)
	return instance, buf, err
}

// Tests ======================================================================

// TestLogLevelFromString tests the .FromString method of the LogLevel type.
func TestLogLevelFromString(t *testing.T) {
	tests := []struct {
		input         string
		expectedLevel LogLevel
		expectedErr   error
	}{
		{"none", NONE, nil},
		{"NONE", NONE, nil},
		{"ERrOR", ERROR, nil},
		{"WARN", WARN, nil},
		{"INFO", INFO, nil},
		{"DeBUG", DEBUG, nil},
		{"FATAL", FATAL, nil},
		{"invAlid", NONE, ErrInvalidLogLevel},
		{"", NONE, ErrInvalidLogLevel},
	}
	for _, test := range tests {
		var level LogLevel
		actualErr := level.FromString(test.input)
		if level != test.expectedLevel || actualErr != test.expectedErr {
			t.Errorf("LogLevel FromString(%s) = %v, %v; want %v, %v", test.input, level, actualErr, test.expectedLevel, test.expectedErr)
		}
	}
}

// TestLogLevelString tests the .String method of the LogLevel type.
func TestLogLevelString(t *testing.T) {
	tests := []struct {
		input    LogLevel
		expected string
	}{
		{NONE, "NONE"},
		{ERROR, "ERROR"},
		{WARN, "WARN"},
		{INFO, "INFO"},
		{DEBUG, "DEBUG"},
		{FATAL, "FATAL"},
		{LogLevel(100), "?"},
	}
	for _, test := range tests {
		actual := test.input.String()
		if actual != test.expected {
			t.Errorf("LogLevelToString(%v) = %s; want %s", test.input, actual, test.expected)
		}
	}
}

// TestPadString tests the PadString function.
func TestPadString(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"", 0, ""},
		{"", 1, " "},
		{"", 2, "  "},
		{"", 3, "   "},
		{"abc", 2, "abc"},
		{"abc", 3, "abc"},
		{"abc", 4, "abc "},
		{"abc", 5, "abc  "},
	}
	for _, test := range tests {
		actual := PadString(test.input, test.length)
		if actual != test.expected {
			t.Errorf("padString(%s, %d) = %s; want %s", test.input, test.length, actual, test.expected)
		}
	}
}

func TestInit(t *testing.T) {
	path := ""
	level := INFO
	instance, err := NewLogger(Config{Level: &level, DirectoryPath: &path}, 255, 2)
	if err != nil {
		t.Errorf("NewLogger(Config{Level: &level, DirectoryPath: &path}, 255, 2) = %v; want nil", err)
	}
	instance.Shutdown(time.Second)
}

func TestInvalidDirectoryPath(t *testing.T) {
	// set path to a invalid directory not allowed on win or linux
	path := "/foo/bar/<>:\"/\\|?*"
	level := INFO
	_, err := NewLogger(Config{Level: &level, DirectoryPath: &path}, 255, 2)
	if !errors.Is(err, ErrInvalidPath) {
		t.Errorf("NewLogger() = %v; when given an invalid path, want ErrInvalidPath", err)
	}
}

func TestShutdown(t *testing.T) {
	path := ""
	level := INFO
	instance, _ := NewLogger(Config{Level: &level, DirectoryPath: &path}, 255, 2)
	time.Sleep(100 * time.Millisecond)
	err := instance.Shutdown(time.Second)
	if err != nil {
		t.Errorf("instance.Shutdown(time.Second) = %v; want nil", err)
	}
}

// At this point we know we can stop the goroutine and safely inspect the logger state after doing so.

// TODO - parallel tests
// - Test console output / formatting - do in a way so we can reuse when testing public instance
// - Test file output / formatting
// - Test log level filtering
// - Test log rotation
// - Test console fallback. call fallbackToConsole(), ensure it's no longer writing to file and only to stdout or test buf

// test public instance

func TestParallelTests(t *testing.T) {
	t.Run("ConsoleOutput", func(t *testing.T) {
		t.Parallel()

		instance, buf, err := normalInit("")
		if err != nil {
			t.Errorf("Error during normalInit: %v", err)
		}

		testMsg := "This is a test"
		instance.Info(testMsg)
		time.Sleep(20 * time.Millisecond)

		actual := buf.String()
		if !strings.Contains(actual, testMsg) {
			t.Errorf("Console output = \"%s\"; want something that contains \"%s\"", actual, testMsg)
		} else {
			actual = strings.TrimSuffix(actual, "\n")
			fmt.Println("Example output: ", actual)
		}
		instance.Shutdown(time.Second)
	})
}

// tempDir, err = os.MkdirTemp("", "example")
