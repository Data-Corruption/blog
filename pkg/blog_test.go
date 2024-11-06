package blog

import (
	"errors"
	"testing"
	"time"
)

/*


tempDir, err = os.MkdirTemp("", "example")


parallel test example:

func TestFirst(t *testing.T) {
  // Run sequentially
  t.Log("Running first test")
}

func TestSecond(t *testing.T) {
  // Run sequentially
  t.Log("Running second test")
}

func TestParallelTests(t *testing.T) {
  t.Run("ParallelTest1", func(t *testing.T) {
    t.Parallel()
    t.Log("Running ParallelTest1")
    // Use t like normal here
  })

  t.Run("ParallelTest2", func(t *testing.T) {
    t.Parallel()
    t.Log("Running ParallelTest2")
    // Use t like normal here
  })
}

*/

// Helper functions ===========================================================

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
		t.Errorf("NewLogger(Config{Level: &level, DirectoryPath: &path}, 255, 2) = %v; want ErrInvalidPath", err)
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
// - Test console output / formatting
// - Test file output / formatting
// - Test log level filtering
// - Test log rotation
// - Test console fallback

/*

func TestInvalidInitArgs(t *testing.T) {
	// Create a temporary directory.
	createTempDir()
	defer cleanupTempDir()

	// Test invalid directory argument.
	invalidDirArg := filepath.Join(tempDir, "invalid")
	err := Init(invalidDirArg, INFO)
	switch err {
	case ErrInvalidPath:
		break
	default:
		t.Errorf("Init(%s, INFO) = %v; want InvalidPathError", invalidDirArg, err)
	}

	// Test re-initialization.
	err = Init(tempDir, INFO)
	switch err {
	case ErrAlreadyInitialized:
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

func TestShouldLog(t *testing.T) {
	normalStartup()
	defer cleanupTempDir()

	SetLevel(WARN)
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		level    LogLevel
		expected bool
	}{
		{NONE, false},
		{ERROR, true},
		{WARN, true},
		{INFO, false},
		{DEBUG, false},
		{FATAL, true},
	}
	for _, test := range tests {
		actual := instance.shouldLog(test.level)
		if actual != test.expected {
			t.Errorf("shouldLog(%v) = %v; want %v", test.level, actual, test.expected)
		}
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
	log.SetOutput(&buf)

	// set useConsole to true and log something
	SetUseConsole(true)
	time.Sleep(100 * time.Millisecond)
	Info("This is a test")
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

func TestDualOutput(t *testing.T) {
	normalStartup()
	defer cleanupTempDir()

	// buffer to capture output
	var buf bytes.Buffer
	log.SetOutput(&buf)

	// set useConsole to true and log something
	SetUseConsole(true)
	time.Sleep(100 * time.Millisecond)
	Info("This is a test")
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

	// sleep for half of defaultFlushInterval
	time.Sleep(defaultFlushInterval / 2)

	// check if flushed too early
	if latestContainsData() {
		t.Errorf("Should not have flushed yet")
	}

	// sleep for the full defaultFlushInterval
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

*/
