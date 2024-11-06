/*
Package blog is a simple async logger with file rotation and console logging.

Usage:

	// Init blog.
	//
	// Parameters:
	//   - DirPath: Path for log files. "." for current working directory or "" to disable file logging.
	//   - Level: Desired logging level for filtering messages.
	//   - IncludeLocation: When true, adds source file and line number to log messages (e.g., "main.go:42").
	//   - EnableConsole: When true, enables logging to the console in addition to files.
	//
	if err := blog.Init("logs", blog.INFO, false, true); err != nil {
		log.Printf("Error initializing logger: %v", err)
	}

	// Log messages from anywhere in the program
	blog.Info("This is an info message.")

	// Log messages with formatting
	blog.Warnf("This is an warn message with a format string: %v", err)

	// Synchronously cleanup the logger with a timeout; 0 means block indefinitely.
	// This should be called at the end of the program.
	blog.Cleanup(0)

	// for all other functions see `blog.go`. For access to the raw logger, see `logger.go`.

# Performance Notes

Defaults; All of these are modifiable at runtime via the public functions:
  - Max buffer size:   4 KB.
  - Max log file size: 1 GB. When this is reached the file is rotated.
  - Flush interval:    15 seconds. For automatic flushing in low traffic scenarios.

A single thread is used to handle all logging operations.
The channel that feeds it messages is buffered to 255 in the instance managed by the public functions.
If you need control over it, you can create your own instance of the raw logger. Note interfacing with
the raw logger is is different from the simplified public functions.

# For contributors

The approach is pretty straightforward. There is a slightly lower abstraction level logger in logger.go.
This file creates and manages an instance of it for the common use case of a high abstraction singleton logger.

The logger is a struct with a few channels for communication and vars for configuration.
When created it starts a goroutine that listens for messages/config updates via the chans then handles them.
The logger's public functions don't interact with it's state directly, they do so through the channels.
This makes it thread-safe and more performant, as relying on go's event system is better than mutexes in this case.

This has some nice benefits:
  - Easily test multiple logger instances in parallel.
  - Users don't need to manage the logger instance themselves.
*/
package blog

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	ErrAlreadyInitialized = fmt.Errorf("blog: already initialized")
	ErrInvalidLogLevel    = fmt.Errorf("blog: invalid log level")
	ErrUninitialized      = fmt.Errorf("blog: uninitialized")
	ErrShutdown           = fmt.Errorf("blog: logger has been shut down")
	ErrInvalidPath        = fmt.Errorf("blog: invalid path")

	instance *Logger = nil
)

// Init sets up the logger with the specified configuration parameters.
//
// Parameters:
//   - DirPath: Directory path for log files. Use "." for current working directory or "" to disable file logging.
//   - Level: Desired logging level for filtering messages.
//   - IncludeLocation: When true, adds source file and line number to log messages (e.g., "main.go:42").
//   - EnableConsole: When true, enables logging to the console in addition to files.
//
// Returns:
//   - ErrAlreadyInitialized if logger was previously initialized,
//   - ErrInvalidPath if the directory path is invalid for any reason,
func Init(
	DirPath string,
	Level LogLevel,
	IncludeLocation bool,
	EnableConsole bool,
) error {
	if instance != nil {
		return ErrAlreadyInitialized
	}
	pathCopy := DirPath
	levelCopy := Level
	cout := Ternary(EnableConsole, &ConsoleLogger{l: log.New(os.Stdout, "", 0)}, nil)
	var err error
	instance, err = NewLogger(Config{Level: &levelCopy, DirectoryPath: &pathCopy, ConsoleOut: cout}, 255, 2)
	return err
}

// Cleanup flushes the log write buffer and exits the logger. If timeout is 0, Cleanup blocks indefinitely.
func Cleanup(timeout time.Duration) error { return a(func() { instance.Shutdown(timeout) }) }

// ==== Logging Functions ===

func Error(msg string) error                  { return a(func() { instance.Error(msg) }) }
func Errorf(format string, args ...any) error { return a(func() { instance.Errorf(format, args...) }) }
func Warn(msg string) error                   { return a(func() { instance.Warn(msg) }) }
func Warnf(format string, args ...any) error  { return a(func() { instance.Warnf(format, args...) }) }
func Info(msg string) error                   { return a(func() { instance.Info(msg) }) }
func Infof(format string, args ...any) error  { return a(func() { instance.Infof(format, args...) }) }
func Debug(msg string) error                  { return a(func() { instance.Debug(msg) }) }
func Debugf(format string, args ...any) error { return a(func() { instance.Debugf(format, args...) }) }

// Fatal logs a fatal message and exits with the given exit code.
// This function will not return, it will exit the program after attempting to log the message.
func Fatal(exitCode int, timeout time.Duration, msg string) error {
	return a(func() { instance.Fatal(exitCode, timeout, msg) })
}

// Fatalf logs a fatal message with a format string and exits with the given exit code.
// This function will not return, it will exit the program after attempting to log the message.
func Fatalf(exitCode int, timeout time.Duration, format string, args ...any) error {
	return a(func() { instance.Fatalf(exitCode, timeout, format, args...) })
}

// SetLevel sets the log level.
func SetLevel(level LogLevel) error {
	return a(func() { instance.UpdateConfig(Config{Level: &level}) })
}

// ==== Buffer controls ====

// Flush manually flushes the log write buffer.
func Flush() error { return a(func() { instance.Flush() }) }

// SyncFlush synchronously flushes the log write buffer and blocks until the flush is complete or the
// timeout is reached. If timeout is 0, SyncFlush blocks indefinitely.
func SyncFlush(timeout time.Duration) error { return a(func() { instance.SyncFlush(timeout) }) }

// SetMaxBufferSizeBytes sets the maximum size of the log write buffer. Larger values will increase memory
// usage and reduce the frequency of disk writes.
func SetMaxBufferSizeBytes(size int) error {
	return a(func() { instance.UpdateConfig(Config{MaxBufferSizeBytes: &size}) })
}

// SetFlushInterval sets the interval at which the log write buffer is automatically flushed to the log file.
// This happens regardless of the buffer size. A value of 0 disables automatic flushing.
func SetFlushInterval(d time.Duration) error {
	return a(func() { instance.UpdateConfig(Config{FlushInterval: &d}) })
}

// ==== File controls ====

// SetMaxFileSizeBytes sets the maximum size of the log file. When the log file reaches
// this size, it is renamed to the current timestamp and a new log file is created.
func SetMaxFileSizeBytes(size int) error {
	return a(func() { instance.UpdateConfig(Config{MaxFileSizeBytes: &size}) })
}

// SetDirectoryPath sets the directory path for the log files. To disable file logging, use an empty string.
func SetDirectoryPath(path string) error {
	return a(func() { instance.UpdateConfig(Config{DirectoryPath: &path}) })
}

// === helpers ===

// instanceGuard is a helper function that checks if the logger instance is initialized and not shutdown.
func instanceGuard() error {
	if instance == nil {
		return ErrUninitialized
	}
	instance.runningMutex.Lock()
	running := instance.running
	instance.runningMutex.Unlock()
	return Ternary(running, nil, ErrShutdown)
}

// a is a helper function for methods that don't return anything.
func a(f func()) error {
	if err := instanceGuard(); err != nil {
		return err
	}
	f()
	return nil
}
