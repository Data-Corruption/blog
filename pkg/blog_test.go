// To run all test run the following in the project root:
// go test -v ./pkg
package blog

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Helper functions ===========================================================

func bufInit() (*Logger, *bytes.Buffer, error) {
	path := "" // no file output
	level := INFO
	buf := &bytes.Buffer{}
	cl := &ConsoleLogger{l: log.New(buf, "", 0)}
	instance, err := NewLogger(Config{Level: &level, DirectoryPath: &path, ConsoleOut: cl}, 255, 2)
	return instance, buf, err
}

func fileInit() (*Logger, *bytes.Buffer, string, error) {
	tempDir, err := os.MkdirTemp("", "example")
	if err != nil {
		return nil, nil, "", err
	}
	level := INFO
	buf := &bytes.Buffer{}
	cl := &ConsoleLogger{l: log.New(buf, "", 0)}
	instance, err := NewLogger(Config{Level: &level, DirectoryPath: &tempDir, ConsoleOut: cl}, 255, 2)
	return instance, buf, tempDir, err
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
	time.Sleep(20 * time.Millisecond)
	err := instance.Shutdown(0)
	if err != nil {
		t.Errorf("instance.Shutdown(0) = %v; want nil", err)
	}
}

// At this point we know we can stop the goroutine and safely inspect the logger state after doing so.

// TODO:
// - Test log level filtering
// - Test console fallback. call fallbackToConsole(), ensure it's no longer writing to file and only to stdout or test buf

func TestParallelTests(t *testing.T) {
	t.Run("ConsoleOutput", func(t *testing.T) {
		t.Parallel()

		instance, buf, err := bufInit()
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
			fmt.Println("Console output: ", actual)
		}

		buf.Reset()
		instance.UpdateConfig(Config{ConsoleOut: &ConsoleLogger{l: nil}})
		time.Sleep(20 * time.Millisecond)
		instance.Info(testMsg)
		time.Sleep(20 * time.Millisecond)

		actual = buf.String()
		if actual != "" {
			t.Errorf("Console output = \"%s\" after being disabled", actual)
		}

		instance.Shutdown(time.Second)
	})

	t.Run("FileOutput", func(t *testing.T) {
		t.Parallel()

		instance, _, dirPath, err := fileInit()
		if err != nil {
			t.Errorf("Error during normalInit: %v", err)
		}

		testMsg := "This is a test"
		instance.Info(testMsg)
		time.Sleep(20 * time.Millisecond)
		instance.SyncFlush(time.Second)

		file, err := os.ReadFile(filepath.Join(dirPath, "latest.log"))
		if err != nil {
			t.Errorf("Error reading log file: %v", err)
		}

		actual := string(file)
		if !strings.Contains(actual, testMsg) {
			t.Errorf("File output = \"%s\"; want something that contains \"%s\"", actual, testMsg)
		} else {
			actual = strings.TrimSuffix(actual, "\n")
			fmt.Println("File output: ", actual)
		}

		instance.Shutdown(time.Second)
	})

	t.Run("FileAutoFlush", func(t *testing.T) {
		t.Parallel()

		instance, _, dirPath, err := fileInit()
		if err != nil {
			t.Errorf("Error during normalInit: %v", err)
		}

		time.Sleep(20 * time.Millisecond)

		testMsg := "This is a test"
		instance.Info(testMsg)
		second := time.Millisecond * 100
		instance.UpdateConfig(Config{FlushInterval: &second})

		time.Sleep(time.Millisecond * 150)

		file, err := os.ReadFile(filepath.Join(dirPath, "latest.log"))
		if err != nil {
			t.Errorf("Error reading log file: %v", err)
		}

		instance.Shutdown(time.Second)

		actual := string(file)
		if !strings.Contains(actual, testMsg) {
			t.Errorf("File output = \"%s\"; want something that contains \"%s\"", actual, testMsg)
		}
	})

	t.Run("FileRotation", func(t *testing.T) {
		t.Parallel()

		instance, _, dirPath, err := fileInit()
		if err != nil {
			t.Errorf("Error during normalInit: %v", err)
		}

		time.Sleep(20 * time.Millisecond)

		s := 100
		instance.UpdateConfig(Config{MaxFileSizeBytes: &s, MaxBufferSizeBytes: &s})
		time.Sleep(20 * time.Millisecond)

		testMsg := strings.Repeat("This is a test", 100)
		instance.Info(testMsg) // creates a new file
		instance.Info(testMsg) // creates a new file
		instance.Info(testMsg) // creates a new file
		time.Sleep(20 * time.Millisecond)
		instance.Shutdown(time.Second)
		time.Sleep(time.Millisecond * 150)

		// check that there are 3 files
		files, err := os.ReadDir(dirPath)
		if err != nil {
			t.Errorf("Error reading directory: %v\nPath: %s", err, dirPath)
		}
		if len(files) != 3 {
			t.Errorf("len(files) = %d; want 3; Path: %s", len(files), dirPath)
		}
	})

	t.Run("Location", func(t *testing.T) {
		t.Parallel()

		if err := Init("", DEBUG, true, true); err != nil {
			log.Printf("Error initializing logger: %v", err)
		}

		// not sure how to test this yet

		// Log messages from anywhere in the program
		Debug("This is a debug message.")
		time.Sleep(100 * time.Millisecond)
		Cleanup(0)
	})
}
