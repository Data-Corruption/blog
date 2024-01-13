package blog

import (
	"bytes"
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Variables ==================================================================

var (
	tempDir = ""
)

// Helper functions ===========================================================

// createTempDir creates a temporary directory and stores its path in the tempDir variable.
func createTempDir() {
	// Create a temporary directory.
	var err error
	tempDir, err = os.MkdirTemp("", "example")
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

// stripFirstLine removes the first line from a string. It returns an error if the string does not contain a newline character.
func stripFirstLine(s string) (string, error) {
	// Find the index of the first newline character.
	newlineIndex := strings.Index(s, "\n")

	// If there's no newline, return an error.
	if newlineIndex == -1 {
		return "", errors.New("no newline character in string")
	}

	// Return the string after the first newline.
	return s[newlineIndex+1:], nil
}

// latestContainsData returns true if the latest log file exists and is not empty, false otherwise.
func latestContainsData() bool {
	// Check if the latest log file exists.
	if _, err := os.Stat(filepath.Join(tempDir, "latest.log")); err != nil {
		return false
	}

	// Check if the latest log file is empty.
	info, err := os.Stat(filepath.Join(tempDir, "latest.log"))
	if err != nil {
		return false
	}
	return info.Size() != 0
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
		{"erR", ERROR, true},
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

func TestLogLevelToString(t *testing.T) {
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
		actual := test.input.toString()
		if actual != test.expected {
			t.Errorf("LogLevelToString(%v) = %s; want %s", test.input, actual, test.expected)
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

	// Test invalid directory argument.
	invalidDirArg := filepath.Join(tempDir, "invalid")
	err := Init(invalidDirArg, INFO)
	switch err.(type) {
	case InvalidPathError:
		break
	default:
		t.Errorf("Init(%s, INFO) = %v; want InvalidPathError", invalidDirArg, err)
	}

	// Test re-initialization.
	err = Init(tempDir, INFO)
	switch err.(type) {
	case AlreadyInitializedError:
		break
	default:
		t.Errorf("Init(%s, INFO) = %v; want AlreadyInitializedError", tempDir, err)
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

// TestGenLogPath tests the genLogPath function. It should return a path with the following format: <tempDir>/YYYY-MM-DD_HH-MM-SS.log
func TestGenLogPath(t *testing.T) {
	normalStartup()
	defer cleanupTempDir()

	path := instance.genLogPath()

	// Create a regular expression to match the expected pattern.
	expectedPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(tempDir) + `[\/\\]\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}\.log$`)

	// Check if the path matches the expected pattern.
	if !expectedPattern.MatchString(path) {
		t.Errorf("Path '%s' does not match the expected format '<tempDir>/YYYY-MM-DD_HH-MM-SS.log'", path)
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

	// print string with timestamp
	log.Print("Actual: \"" + buf.String() + "\"")

	expected := "INFO]  This is a test\n"
	if actual != expected {
		t.Errorf("Console output = \"%s\"; want \"%s\"", actual, expected)
	}
}

func TestHandleFlushError(t *testing.T) {
	normalStartup()
	defer cleanupTempDir()

	// buffer to capture output
	var buf bytes.Buffer
	log.SetOutput(&buf)

	// log something and sleep for 100ms to allow for channel to be processed
	Info("This is a test")
	time.Sleep(100 * time.Millisecond)

	// manually trigger flush error
	instance.handleFlushError(errors.New("Test error"))

	// check if variables were set correctly
	if instance.useConsole != true {
		t.Errorf("useConsole = %v; want true", instance.useConsole)
	}
	if instance.dirPath != "" {
		t.Errorf("dirPath = %s; want \"\"", instance.dirPath)
	}

	// check if buffer starts with expected string
	actual := buf.String()
	expected := "Falling back to console logging due to an error flushing the log write buffer: Test error\n"
	if !strings.HasPrefix(actual, expected) {
		t.Errorf("Console output = \"%s\"; want \"%s\"", actual, expected)
	}

	// strip the first line from actual (the error message)
	actual, err := stripFirstLine(actual)
	if err != nil {
		t.Errorf("Error stripping first line: %v", err)
	}

	// strip timestamp from actual (which now should just be the log message)
	actual, err = stripTimestamp(actual)
	if err != nil {
		t.Errorf("Error stripping timestamp: %v", err)
	}

	// check if the log message was logged to console
	expected = "INFO]  This is a test\n"
	if actual != expected {
		t.Errorf("Console output = \"%s\"; want \"%s\"", actual, expected)
	}

	// redirect output back to stdout
	log.SetOutput(os.Stdout)
}

func TestAutoFlush(t *testing.T) {
	normalStartup()
	defer cleanupTempDir()

	// log something
	Info("This is a test")

	// sleep for half of defaultFlushInterval
	time.Sleep(defaultFlushInterval / 2)

	// check if flushed too early
	if latestContainsData() {
		t.Errorf("Should not have flushed yet")
	}

	// sleep for the other half of defaultFlushInterval
	time.Sleep(defaultFlushInterval)

	// the file should contain data now
	if !latestContainsData() {
		t.Errorf("Should have flushed by now")
	}
}

func TestSetFlushInterval(t *testing.T) {
	normalStartup()
	defer cleanupTempDir()

	newFlushInterval := 1 * time.Second

	// set flush interval to 1 second
	SetFlushInterval(newFlushInterval)

	// sleep for 100ms to allow for channel to be processed
	time.Sleep(100 * time.Millisecond)

	// check if flush interval was set correctly
	copyOfInstance := getCopyOfInstance()
	if copyOfInstance.flushInterval != newFlushInterval {
		t.Errorf("Flush interval = %v; want %v", copyOfInstance.flushInterval, newFlushInterval)
	}

	// log something
	Info("This is a test")

	// sleep for half of newFlushInterval
	time.Sleep(newFlushInterval / 2)

	// check if flushed too early
	if latestContainsData() {
		t.Errorf("Should not have flushed yet")
	}

	// sleep for the other half of newFlushInterval
	time.Sleep(newFlushInterval)

	// the file should contain data now
	if !latestContainsData() {
		t.Errorf("Should have flushed by now")
	}
}

func TestManualFlush(t *testing.T) {
	normalStartup()
	defer cleanupTempDir()

	// log something
	Info("This is a test")

	// sleep for 100ms to allow for channel to be processed
	time.Sleep(100 * time.Millisecond)

	// manually flush
	Flush()

	// sleep for 100ms to allow for channel to be processed and file to be written
	time.Sleep(100 * time.Millisecond)

	// the file should contain data now
	if !latestContainsData() {
		t.Errorf("Should have flushed by now")
	}
}

func TestSyncFlush(t *testing.T) {
	normalStartup()
	defer cleanupTempDir()

	// log something
	Info("This is a test")

	// sleep for 100ms to allow for channel to be processed
	time.Sleep(100 * time.Millisecond)

	// manually sync flush
	SyncFlush(0)

	// the file should contain data now
	if !latestContainsData() {
		t.Errorf("Should have flushed by now")
	}
}

// test flush on max write size hit
func TestMaxWriteSize(t *testing.T) {
	normalStartup()
	defer cleanupTempDir()

	newMaxBufSize := 100
	testString := strings.Repeat("a", newMaxBufSize*2)

	SetMaxWriteBufSize(newMaxBufSize)
	time.Sleep(100 * time.Millisecond)
	Info(testString)
	time.Sleep(100 * time.Millisecond)

	// the file should contain data now
	if !latestContainsData() {
		t.Errorf("Should have flushed by now")
	}
}
