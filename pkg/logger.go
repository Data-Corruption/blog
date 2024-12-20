package blog

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
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
*/
type Logger struct {
	// Configuration settings.
	config Config

	// Number of stack frames to skip when including the location of the log message. Default is 2, -1 to disable.
	locationSkip int // not configurable after creation for performance reasons

	// Buffer for messages before they are written to console or file.
	writeBuffer bytes.Buffer

	// True when the goroutine is running.
	running      bool
	runningMutex sync.Mutex

	// Config update method. Uses chans instead of a mutex for better performance.
	getConfigChan chan chan Config
	setConfigChan chan Config // nil fields are ignored

	messageChan   chan LogMessage
	flushSignal   chan struct{}
	syncFlushChan chan chan struct{}
	shutdownChan  chan chan struct{}
}

// ConsoleLogger wraps *log.Logger to allow nil value semantics for disabled state
type ConsoleLogger struct {
	l *log.Logger
}

// Config holds the configuration settings for the Logger.
type Config struct {
	Level              *LogLevel      // the minimum log level to write. Default is INFO.
	MaxBufferSizeBytes *int           // the maximum size of the write buffer before it is flushed. Default is 4 KB.
	MaxFileSizeBytes   *int           // the maximum size of the log file before it is rotated. Default is 1 GB.
	FlushInterval      *time.Duration // the interval at which the write buffer is flushed. Default is 15 seconds.
	DirectoryPath      *string        // the directory path where the log file is stored. Default is the current working directory ("."). To disable file logging, set this to an empty string.
	ConsoleOut         *ConsoleLogger // the logger to write to the console. Default is ConsoleLogger{l: nil}. When l is nil, console logging is disabled. This is configurable for easy testing.
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
// It initializes all channels and starts the background logging goroutine.
//
// The msgChanSize parameter controls the buffer size of the message channel,
// where 0 means unbuffered. LocationSkip controls the number of stack frames
// to skip when including the location in log messages (-1 to disable). For
// normal usage, LocationSkip should be set to 2.
//
// Returns an error if the log directory path cannot be set.
func NewLogger(cfg Config, msgChanSize int, LocationSkip int) (*Logger, error) {
	// Set default values for any nil fields in the configuration.
	if cfg.Level == nil {
		defaultLogLevel := INFO
		cfg.Level = &defaultLogLevel
	}
	if cfg.MaxBufferSizeBytes == nil {
		defaultMaxBufferSizeBytes := MaxBufferSizeBytes
		cfg.MaxBufferSizeBytes = &defaultMaxBufferSizeBytes
	}
	if cfg.MaxFileSizeBytes == nil {
		defaultMaxLogFileSizeBytes := MaxLogFileSizeBytes
		cfg.MaxFileSizeBytes = &defaultMaxLogFileSizeBytes
	}
	if cfg.FlushInterval == nil {
		defaultFlushInterval := FlushInterval
		cfg.FlushInterval = &defaultFlushInterval
	}
	if cfg.DirectoryPath == nil {
		defaultDirectoryPath := ""
		cfg.DirectoryPath = &defaultDirectoryPath
	}
	if cfg.ConsoleOut == nil {
		cfg.ConsoleOut = &ConsoleLogger{}
	}

	// Create the logger instance.
	l := &Logger{
		config:        cfg,
		locationSkip:  LocationSkip,
		running:       true,
		messageChan:   make(chan LogMessage, msgChanSize),
		getConfigChan: make(chan chan Config),
		setConfigChan: make(chan Config),
		flushSignal:   make(chan struct{}),
		syncFlushChan: make(chan chan struct{}),
		shutdownChan:  make(chan chan struct{}),
	}

	// Set the log directory path
	if err := l.setPath(*l.config.DirectoryPath); err != nil {
		return nil, errors.Join(ErrInvalidPath, err)
	}

	// Start the logger goroutine
	go l.run()

	// Return the logger instance
	return l, nil
}

// Shutdown synchronously flushes and waits for the logger to shutdown it's goroutine for the given timeout duration.
// A timeout of 0 means block indefinitely.
// You may want to time.Sleep(20 * time.Millisecond) before calling this function to ensure all log messages are buffered.
func (l *Logger) Shutdown(timeout time.Duration) error {
	done := make(chan struct{})
	l.shutdownChan <- done
	if timeout == 0 {
		<-done
		return nil
	}
	select {
	case <-done:
	case <-time.After(timeout):
	}
	l.runningMutex.Lock()
	defer l.runningMutex.Unlock()
	return Ternary(l.running, fmt.Errorf("logger failed to shutdown in time"), nil)
}

// Start restarts the logger goroutine after a shutdown.
func (l *Logger) Start() {
	l.runningMutex.Lock()
	defer l.runningMutex.Unlock()
	if !l.running {
		l.running = true
		go l.run()
	}
}

// Flush asynchronously flushes the log write buffer.
func (l *Logger) Flush() {
	l.flushSignal <- struct{}{}
}

// SyncFlush synchronously flushes the log write buffer with the given timeout duration.
// A timeout of 0 means block indefinitely.
func (l *Logger) SyncFlush(timeout time.Duration) {
	done := make(chan struct{})
	l.syncFlushChan <- done
	select {
	case <-done:
	case <-time.After(timeout):
	}
}

// GetConfigCopy returns a copy of the current logger configuration.
func (l *Logger) GetConfigCopy() Config {
	resp := make(chan Config)
	l.getConfigChan <- resp
	return <-resp
}

// UpdateConfig updates the logger configuration with the provided settings.
// Nil fields are ignored.
func (l *Logger) UpdateConfig(cfg Config) {
	l.setConfigChan <- cfg
}

// Log message functions. These are the main interface for logging messages.

func (l *Logger) Info(msg string)                   { l.qM(INFO, 0, msg) }
func (l *Logger) Infof(format string, args ...any)  { l.qM(INFO, 0, format, args...) }
func (l *Logger) Warn(msg string)                   { l.qM(WARN, 0, msg) }
func (l *Logger) Warnf(format string, args ...any)  { l.qM(WARN, 0, format, args...) }
func (l *Logger) Error(msg string)                  { l.qM(ERROR, 0, msg) }
func (l *Logger) Errorf(format string, args ...any) { l.qM(ERROR, 0, format, args...) }
func (l *Logger) Debug(msg string)                  { l.qM(DEBUG, 0, msg) }
func (l *Logger) Debugf(format string, args ...any) { l.qM(DEBUG, 0, format, args...) }

// Fatal attempts to log a message and exits the program. It exits with the given exit code either when the message is
// logged or the timeout duration is reached. A timeout of 0 means block indefinitely.
func (l *Logger) Fatal(exitCode int, timeout time.Duration, msg string) {
	l.qM(FATAL, exitCode, msg)
	time.Sleep(timeout)
	fmt.Printf("Fatal message failed to log in time: %s\n", msg)
	os.Exit(exitCode)
}

// Fatalf is a convenience function that calls Fatal with a format string.
func (l *Logger) Fatalf(exitCode int, timeout time.Duration, format string, args ...any) {
	l.Fatal(exitCode, timeout, fmt.Sprintf(format, args...))
}

// Internal functions

// qM is a helper function to create and enqueue a log message.
func (l *Logger) qM(level LogLevel, exitCode int, format string, args ...any) {
	m := LogMessage{
		level:     level,
		exitCode:  exitCode,
		timestamp: time.Now(),
		location:  "",
		content:   fmt.Sprintf(format, args...),
	}
	if l.locationSkip != -1 {
		if (level == FATAL) || (level == ERROR) || (level == DEBUG) {
			if _, file, line, ok := runtime.Caller(l.locationSkip); ok {
				m.location = fmt.Sprintf("%s:%d", filepath.Base(file), line)
			}
		}
	}
	l.messageChan <- m
}

// fallbackToConsole disables file logging and enables console logging if not already enabled. Also passes the given error through.
func (l *Logger) fallbackToConsole() {
	*l.config.DirectoryPath = ""
	if l.config.ConsoleOut.l == nil {
		l.config.ConsoleOut = &ConsoleLogger{log.New(os.Stdout, "", 0)}
	}
}

// setPath sets the directory path for the log file. If there are any issues, it enables console logging and returns an error.
func (l *Logger) setPath(path string) error {
	// Handle the special case of an empty path
	if path == "" {
		l.fallbackToConsole()
		return nil
	}
	// Check if the path exists and is a directory
	cleanedPath := filepath.Clean(path)
	fileInfo, err := os.Stat(cleanedPath)
	if err != nil {
		l.fallbackToConsole()
		return fmt.Errorf("failed to stat path: %w", err)
	}
	if !fileInfo.IsDir() {
		l.fallbackToConsole()
		return fmt.Errorf("path is not a directory: %s", cleanedPath)
	}
	// Set the directory path
	*l.config.DirectoryPath = cleanedPath
	return nil
}

func (l *Logger) handleMessage(m LogMessage) {
	// Check if the message should be logged given the current log level
	if l.config.Level == nil || *l.config.Level == NONE {
		return
	}
	if m.level > *l.config.Level {
		return
	}
	// Create the message prefix
	prefix := m.timestamp.Format("[2006-01-02,15-04-05,") + m.level.String() + "] "
	prefix = PadString(prefix, 28)
	// Add location if it exists
	if m.location != "" {
		prefix += "[" + m.location + "] "
	}
	// Format the message
	m.content = prefix + m.content + "\n"
	// If file logging is enabled, write the message to the log file
	if *l.config.DirectoryPath != "" {
		l.writeBuffer.WriteString(m.content)
		if l.writeBuffer.Len() >= *l.config.MaxBufferSizeBytes {
			l.flush()
		}
	}
	// If console logging is enabled, write the message to the console
	if l.config.ConsoleOut.l != nil {
		l.config.ConsoleOut.l.Print(m.content)
	}
	if m.level == FATAL {
		l.flush()
		os.Exit(m.exitCode)
	}
}

// getLatestPath returns the path to the latest.log file.
func (l *Logger) getLatestPath() string {
	return filepath.Join(*l.config.DirectoryPath, "latest.log")
}

// handleFileOverflow renames latest.log to the current timestamp and creates a new latest.log.
func (l *Logger) handleFileOverflow() (*os.File, error) {
	if *l.config.DirectoryPath == "" {
		return nil, fmt.Errorf("file logging is disabled")
	}
	// Create a new name for the current log file
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	path := filepath.Join(*l.config.DirectoryPath, timestamp+".log")
	// If path already exists, add a random string suffix to the timestamp
	if _, err := os.Stat(path); err == nil {
		randomString, err := GenRandomString(8)
		if err != nil {
			return nil, err
		}
		path = filepath.Join(*l.config.DirectoryPath, timestamp+"_"+randomString+".log")
	}
	// Rename latest.log to the current timestamp
	if err := os.Rename(l.getLatestPath(), path); err != nil {
		return nil, err
	}
	// Create a new latest.log
	return os.OpenFile(l.getLatestPath(), os.O_CREATE|os.O_WRONLY, 0644)
}

// handleFlushError prints the error to the console, sets use console to true and dir path to nil,
// effectively disabling file logging, and prints the remaining write buffer to the console.
func (l *Logger) handleFlushError(err error) {
	l.fallbackToConsole()
	// print the remaining write buffer to the console
	l.config.ConsoleOut.l.Print("Failed to write to log file: " + err.Error() + "\n")
	l.config.ConsoleOut.l.Print(l.writeBuffer.String())
	l.writeBuffer.Reset()
}

// flush writes the buffered log to the file and resets the buffer.
func (l *Logger) flush() {
	if (l.writeBuffer.Len() == 0) || (*l.config.DirectoryPath == "") {
		return
	}
	// Open the log file
	f, err := os.OpenFile(l.getLatestPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		l.handleFlushError(err)
		return
	}
	// Check if the log file is too large
	fileInfo, err := f.Stat()
	if err != nil {
		l.handleFlushError(err)
		f.Close()
		return
	}
	// If the log file is too large, create a new log file
	if fileInfo.Size() >= int64(*l.config.MaxFileSizeBytes) {
		f.Close()
		if f, err = l.handleFileOverflow(); err != nil {
			l.handleFlushError(err)
			return
		}
	}
	// Write the buffered log to the file
	if _, err := f.Write(l.writeBuffer.Bytes()); err != nil {
		l.handleFlushError(err)
		f.Close()
		return
	}
	// Reset the buffer and close the file
	l.writeBuffer.Reset()
	f.Close()
}

// run is the main loop for the logger goroutine.
func (l *Logger) run() {
	ticker := time.NewTicker(*l.config.FlushInterval)
	restartTickerReq := false
	defer ticker.Stop()

	for {
		if restartTickerReq {
			restartTickerReq = false
			ticker.Stop()
			if *l.config.FlushInterval > 0 {
				ticker = time.NewTicker(*l.config.FlushInterval)
			}
		}
		select {
		case m := <-l.messageChan:
			l.handleMessage(m)
		case <-l.flushSignal:
			l.flush()
		case <-ticker.C:
			l.flush()
		case done := <-l.syncFlushChan:
			l.flush()
			done <- struct{}{}
		case done := <-l.shutdownChan:
			l.flush()
			done <- struct{}{}
			l.runningMutex.Lock()
			l.running = false
			l.runningMutex.Unlock()
			return
		case resp := <-l.getConfigChan:
			resp <- l.config
		case cfg := <-l.setConfigChan:
			CopyNotNil(l.config.Level, cfg.Level)
			CopyNotNil(l.config.MaxBufferSizeBytes, cfg.MaxBufferSizeBytes)
			CopyNotNil(l.config.MaxFileSizeBytes, cfg.MaxFileSizeBytes)
			if cfg.FlushInterval != nil {
				*l.config.FlushInterval = *cfg.FlushInterval
				restartTickerReq = true
			}
			if cfg.DirectoryPath != nil {
				l.setPath(*cfg.DirectoryPath)
			}
			if cfg.ConsoleOut != nil {
				l.config.ConsoleOut.l = cfg.ConsoleOut.l
			}
		}
	}
}
