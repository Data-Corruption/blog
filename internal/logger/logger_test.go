package logger

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Data-Corruption/blog/v3/internal/config"
	LogLevel "github.com/Data-Corruption/blog/v3/internal/level"
)

// helper to return pointer values for simple types.
func ptr[T any](v T) *T { return &v }

// Test that messages are printed to the console.
func TestLoggerConsole(t *testing.T) {
	// Use an empty directory to disable file logging.
	buf := new(bytes.Buffer)
	config := &config.Config{
		DirectoryPath: ptr(""),
		Level:         ptr(LogLevel.INFO),
		ConsoleOut:    &config.ConsoleLogger{L: log.New(buf, "", 0)},
	}
	logInst, err := NewLogger(config, 255, 2)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logInst.Shutdown(time.Second)

	msg := "Test console message"
	logInst.Info(msg)
	time.Sleep(50 * time.Millisecond) // Allow the run loop to pick up the message
	// Force a flush (rather than sleeping) to ensure asynchronous writes are processed.
	logInst.SyncFlush(time.Second)

	output := buf.String()
	if !strings.Contains(output, msg) {
		t.Errorf("expected console output to contain %q, got %q", msg, output)
	}
}

// Test that messages are written to a file.
func TestLoggerFile(t *testing.T) {
	// Create a temporary directory for file logging.
	tempDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Even though we inject a console buffer, we will check the file.
	buf := new(bytes.Buffer)
	config := &config.Config{
		DirectoryPath: ptr(tempDir),
		Level:         ptr(LogLevel.INFO),
		ConsoleOut:    &config.ConsoleLogger{L: log.New(buf, "", 0)},
	}
	logInst, err := NewLogger(config, 255, 2)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logInst.Shutdown(time.Second)

	msg := "Test file message"
	logInst.Info(msg)
	time.Sleep(50 * time.Millisecond) // Allow the run loop to pick up the message
	logInst.SyncFlush(time.Second)

	// Read the latest.log file.
	logFile := filepath.Join(tempDir, "latest.log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if !strings.Contains(string(data), msg) {
		t.Errorf("expected file log to contain %q, got %q", msg, string(data))
	}
}

// Test that log level filtering works.
func TestLoggerLogLevelFiltering(t *testing.T) {
	buf := new(bytes.Buffer)
	config := &config.Config{
		DirectoryPath: ptr(""), // disable file logging
		Level:         ptr(LogLevel.WARN),
		ConsoleOut:    &config.ConsoleLogger{L: log.New(buf, "", 0)},
	}
	logInst, err := NewLogger(config, 255, 2)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logInst.Shutdown(time.Second)

	logInst.Info("Info message")      // should be filtered out
	logInst.Warn("Warn message")      // should be logged
	time.Sleep(50 * time.Millisecond) // Allow the run loop to pick up the message
	logInst.SyncFlush(time.Second)

	output := buf.String()
	if strings.Contains(output, "Info message") {
		t.Errorf("info message should be filtered out at WARN level")
	}
	if !strings.Contains(output, "Warn message") {
		t.Errorf("warn message should be logged")
	}
}

// Test updating configuration dynamically.
func TestLoggerConfigUpdate(t *testing.T) {
	buf := new(bytes.Buffer)
	cfg := &config.Config{
		DirectoryPath: ptr(""),
		Level:         ptr(LogLevel.INFO),
		ConsoleOut:    &config.ConsoleLogger{L: log.New(buf, "", 0)},
	}
	logInst, err := NewLogger(cfg, 255, 2)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logInst.Shutdown(time.Second)

	logInst.Info("Initial message")
	time.Sleep(50 * time.Millisecond) // Allow the run loop to pick up the message
	logInst.SyncFlush(time.Second)
	if !strings.Contains(buf.String(), "Initial message") {
		t.Errorf("expected initial message to be logged")
	}
	buf.Reset()

	// Update the config to disable console logging.
	logInst.UpdateConfig(config.Config{
		ConsoleOut: &config.ConsoleLogger{L: nil},
	})
	time.Sleep(50 * time.Millisecond) // Allow the run loop to pick up the message
	logInst.Info("This should not appear")
	time.Sleep(50 * time.Millisecond) // Allow the run loop to pick up the message
	logInst.SyncFlush(time.Second)
	if buf.Len() != 0 {
		t.Errorf("expected no console output after disabling console logging, got %q", buf.String())
	}
}
