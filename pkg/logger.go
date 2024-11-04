package blog

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"
)

// Defaults for configuration settings.
const (
	MaxBufferSizeBytes  = 4096               // 4 KB
	MaxLogFileSizeBytes = 1024 * 1024 * 1024 // 1 GB
	FlushInterval       = 15 * time.Second
)

/*
Logger is a simple, thread-safe logger. It supports various log levels, file and or
console logging, basic performance tuning, automatic flushing, and size based log rotation.

Usage:

	// Create a new logger instance, setting all configuration options.
	var err error
	var logger *blog.Logger
	logger, err := blog.NewLogger(blog.Config{}, 255, os.Stdout) // 255 is the size of the message channel buffer, os.Stdout is the console output writer
	if err != nil {
		log.Printf("Error creating logger: %v", err)
	}

	// Log messages
	logger.Info("This is an info message.")
	logger.Infof("This is an info message with a format string: %v", err)

	// Log a fatal message and exit with the given exit code
	logger.Fatalf(1, 0, "This is a fatal message with a format string: %v", err)

	// Get the current logger configuration
	config := logger.GetConfigCopy()

	// Update a configuration settings at any time
	newLevel := blog.DEBUG
	logger.UpdateConfig(blog.Config{ Level: &newLevel }) // nil fields are ignored

	// Manually flush the log write buffer
	logger.Flush()

	// Synchronously flush the log write buffer with a timeout; 0 means block indefinitely
	logger.SyncFlush(0)

	// Synchronously flush and shutdown the logger with a timeout; 0 means block indefinitely
	logger.Shutdown(0)

Performance Notes:

  - Default max buf and file size are 4 KB and 1 GB respectively. The default flush interval is 15 seconds.
  - A single thread is used to handle all logging operations. This is generally fine for most applications.
*/
type Logger struct {
	// Configuration settings.
	config Config

	// Base logger instance, to avoid affecting the global logger when using console logging.
	baseLogger *log.Logger

	// Buffer for messages before they are written to console or file.
	writeBuffer bytes.Buffer

	// Primitives for communication between the logger and its main goroutine.
	msgChan       chan LogMessage // buffered to prevent blocking on high-frequency logging.
	getConfigChan chan chan Config
	setConfigChan chan Config // nil fields are ignored
	flushSignal   chan struct{}
	syncFlushChan chan chan struct{}
	shutdownChan  chan chan struct{}
}

// Config holds the configuration settings for the Logger.
type Config struct {
	Level              *LogLevel      // the minimum log level to write. Default is INFO.
	UseConsole         *bool          // whether to log to the console. Default is true.
	IncludeLocation    *bool          // whether to include the file and line number in the log message. Default is false.
	LocationSkip       *int           // number of stack frames to skip when getting the location. Default is 2.
	MaxBufferSizeBytes *int           // the maximum size of the write buffer before it is flushed. Default is 4 KB.
	MaxFileSizeBytes   *int           // the maximum size of the log file before it is rotated. Default is 1 GB.
	FlushInterval      *time.Duration // the interval at which the write buffer is flushed. Default is 15 seconds.
	DirectoryPath      *string        // the directory path where the log file is stored. Default is the current working directory ("."). To disable file logging, set this to an empty string.
}

// LogMessage represents a single log message.
type LogMessage struct {
	level     LogLevel
	exitCode  int // only used by FATAL messages
	timestamp time.Time
	location  string // e.g., "file.go:42"
	content   string
}

// NewLogger creates a new Logger instance with the provided configuration.
func NewLogger(cfg Config, msgChanSize int, consoleOut io.Writer) (*Logger, error) {
	// Set default values for any nil fields in the configuration.
	SetIfNil(&cfg.Level, INFO)
	SetIfNil(&cfg.UseConsole, true)
	SetIfNil(&cfg.IncludeLocation, false)
	SetIfNil(&cfg.LocationSkip, 2)
	SetIfNil(&cfg.MaxBufferSizeBytes, MaxBufferSizeBytes)
	SetIfNil(&cfg.FlushInterval, FlushInterval)
	SetIfNil(&cfg.MaxFileSizeBytes, MaxLogFileSizeBytes)
	SetIfNil(&cfg.DirectoryPath, "")

	// Create the logger instance.
	l := &Logger{
		config:        cfg,
		baseLogger:    log.New(consoleOut, "", 0),
		msgChan:       make(chan LogMessage, msgChanSize),
		getConfigChan: make(chan chan Config),
		setConfigChan: make(chan Config),
		flushSignal:   make(chan struct{}),
		syncFlushChan: make(chan chan struct{}),
		shutdownChan:  make(chan chan struct{}),
	}

	// Set the log directory path
	if err := l.setPath(l.config.DirectoryPath); err != nil {
		return nil, fmt.Errorf("failed to set log directory path: %w", err)
	}

	// Start the logger goroutine
	go l.run()

	// Return the logger instance
	return l, nil
}
