package blog

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Variables ==================================================================

var (
	tempDir    = ""
	captureBuf bytes.Buffer
)

// Helper functions ===========================================================

// createTempDir creates a temporary directory and stores its path in the tempDir variable.
func createTempDir() {
	// Create a temporary directory.
	var err error
	tempDir, err = ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}
}

// cleanupTempDir removes the temporary directory
func cleanupTempDir() {
	os.RemoveAll(tempDir)
}

// normalStartup puts the package in a normal startup state.
func normalStartup() {
	reset()
	cleanupTempDir()
	createTempDir()
	if err := Init(tempDir, INFO); err != nil {
		log.Fatalf("Init(%s, INFO) = %v; want nil", tempDir, err)
	}
}

// stripTimestamp removes the timestamp from a log line.
func stripTimestamp(line string) (string, error) {
	// Find the index of the first comma.
	firstComma := strings.Index(line, ",")
	if firstComma == -1 {
		return "", errors.New("not enough commas in log line")
	}

	// Find the index of the second comma, starting the search just after the first comma.
	secondComma := strings.Index(line[firstComma+1:], ",")
	if secondComma == -1 {
		return "", errors.New("not enough commas in log line")
	}

	// Return the part of the string after the second comma.
	return line[firstComma+1+secondComma+1:], nil
}

// Tests ======================================================================

func TestLogLevelFromString(t *testing.T) {
	tests := []struct {
		input         string
		expectedLevel LogLevel
		expectedOk    bool
	}{
		{"none", NONE, true},
		{"NONE", NONE, true},
		{"ERROR", ERROR, true},
		{"WARN", WARN, true},
		{"INFO", INFO, true},
		{"DEBUG", DEBUG, true},
		{"FATAL", FATAL, true},
		{"invalid", NONE, false},
		{"", NONE, false},
	}
	for _, test := range tests {
		level, ok := LogLevelFromString(test.input)
		if level != test.expectedLevel || ok != test.expectedOk {
			t.Errorf("LogLevelFromString(%s) = %v, %v; want %v, %v", test.input, level, ok, test.expectedLevel, test.expectedOk)
		}
	}
}

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
		actual := padString(test.input, test.length)
		if actual != test.expected {
			t.Errorf("padString(%s, %d) = %s; want %s", test.input, test.length, actual, test.expected)
		}
	}
}

func TestInvalidInitArgs(t *testing.T) {
	// Create a temporary directory.
	createTempDir()
	defer cleanupTempDir()

	// Initialize with invalid arguments.
	invalidDirArg := filepath.Join(tempDir, "invalid")
	err := Init(invalidDirArg, INFO)
	if err == nil {
		t.Errorf("Init(%s, INFO) = nil; want error", invalidDirArg)
	}
}

func TestInit(t *testing.T) {
	reset()
	createTempDir()
	defer cleanupTempDir()

	// Initialize with valid arguments.
	err := Init(tempDir, INFO)
	if err != nil {
		t.Errorf("Init(%s, INFO) = %v; want nil", tempDir, err)
	}
}

func TestConsoleOutput(t *testing.T) {
	normalStartup()
	defer cleanupTempDir()

	// buffer to capture output
	var buf bytes.Buffer

	// set output to buffer
	log.SetOutput(&buf)

	// test
	SetUseConsole(true)
	Info("This is a test")

	// sleep for 100ms to allow for channel to be processed
	time.Sleep(100 * time.Millisecond)

	// get output, strip timestamp
	actual, err := stripTimestamp(buf.String())
	if err != nil {
		t.Errorf("Error stripping timestamp: %v", err)
	}

	// redirect output back to stdout
	log.SetOutput(os.Stdout)

	expected := "INFO: This is a test\n"
	if actual != expected {
		t.Errorf("Console output = \"%s\"; want \"%s\"", actual, expected)
	}
}

/*
func TestAutoFlush(t *testing.T) {
	normalStartup()
	defer cleanupTempDir()

	//
}

func TestManualFlush(t *testing.T) {
	normalStartup()
	defer cleanupTempDir()

	//
}
*/
