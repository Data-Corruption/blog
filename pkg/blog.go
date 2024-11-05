/*
Package blog implements a simple, thread-safe singleton logger.
It supports various log levels and can write to files or the console.

Usage:

	// Init the logger with a dir path(or "" to disable file logging), a log level, and whether to include line numbers.
	if err := blog.Init("logs", blog.INFO, false); err != nil {
		log.Printf("Error initializing logger: %v", err)
	}

	// Log messages from anywhere in the program
	blog.Info("This is an info message.")

	// Log messages with formatting
	blog.Infof("This is an info message with a format string: %v", err)

	// Log a fatal message and exit with the given exit code
	blog.Fatalf(1, 0, "This is a fatal message with a format string: %v", err)

	// Manually flush the log write buffer
	blog.Flush()

	// Synchronously flush the log write buffer with a timeout; 0 means block indefinitely
	blog.SyncFlush(0)

	// Synchronously cleanup the logger with a timeout; 0 means block indefinitely
	blog.Cleanup(0)

The logger can be configured with the following functions at any time:
  - SetLevel(LogLevel) sets the log level.
  - SetUseConsole(bool) sets whether or not to log to the console.
  - SetDirPath(string) sets the directory path for the log files. If dirPath is empty, the current working directory is used.
  - SetMaxWriteBufSize(int) sets the maximum size(in bytes) of the log write buffer.
  - SetMaxFileSize(int) sets the maximum size(in bytes) of the log file before it is renamed and a new log file is created.
  - SetFlushInterval(time.Duration) sets the interval at which the log write buffer is automatically flushed to the log file.

Performance Notes:
  - Default max buf and file size are 4 KB and 1 GB respectively. The default flush interval is 15 seconds.
  - A single thread is used to handle all logging operations.
    The channel that feeds it messages is buffered to 255 via a constant. If the buffer becomes full, log funcs will block.
    This shouldn't be an issue as if the flush fails for whatever reason, the logger will fall back to console logging.
    Worst case, you parallel log in mass and the blocking becomes a bottleneck.

For contributors:

	The approach is pretty straightforward. The logger is a type with a bunch of channels for communication and vars for configuration.
	When created it starts a goroutine that listens for messages/config updates via the chans then handles them.
	The public functions don't interact with the logger directly, they do so using the channels.

	Tests should create their own logger instances using newLogger() then use the 'r' prefixed member functions to interact with it.
	newLogger() lets you set the output writer for testing purposes. The public Init sets it to os.Stdout
	The public functions create and use a singleton instance of the logger.

	This has some nice benefits:
	- Easily test multiple logger instances in parallel.
	- Users don't need to manage the logger instance themselves.
*/
package blog

import (
	"fmt"
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

// Init sets up the logger with the specified directory path and log level.
// It returns an error if called more than once or if the directory path is invalid.
// On error, logging falls back to the console. See ErrAlreadyInitialized and ErrInvalidPath.
func Init(dirPath string, level LogLevel, includeLocation, enableConsole bool) error {
	if instance != nil {
		return ErrAlreadyInitialized
	}
	pathCopy := dirPath
	levelCopy := level
	var err error
	instance, err = NewLogger(Config{Level: &levelCopy, DirectoryPath: &pathCopy}, 255, 2, nil)
	return err
}

// Cleanup flushes the log write buffer and exits the logger. If timeout is 0, Cleanup blocks indefinitely.
func Cleanup(timeout time.Duration) error {
	if err := instanceGuard(); err != nil {
		return err
	}
	return instance.Shutdown(timeout)
}

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
