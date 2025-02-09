package logger

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/Data-Corruption/blog/v3/internal/config"
	"github.com/Data-Corruption/blog/v3/internal/level"
	"github.com/Data-Corruption/blog/v3/internal/utils"
	"github.com/Data-Corruption/blog/v3/internal/utils/strutil"
)

/*
Logger is a simple, thread-safe logger. It supports various log levels, file and or
console logging, basic performance tuning, automatic flushing, and size based log rotation.
*/
type Logger struct {
	// Configuration settings.
	config *config.Config

	// Number of stack frames to skip when including the location of the log message. Default is 2, -1 to disable.
	locationSkip int // not configurable after creation for performance reasons

	// Buffer for messages before they are written to console or file.
	writeBuffer bytes.Buffer

	// True when the goroutine is running.
	Running      bool
	RunningMutex sync.Mutex

	// Config update method. Uses chans instead of a mutex for better performance.
	getConfigChan chan chan config.Config
	setConfigChan chan config.Config // nil fields are ignored

	messageChan   chan LogMessage
	flushSignal   chan struct{}
	syncFlushChan chan chan struct{}
	shutdownChan  chan chan struct{}
}

// LogMessage represents a single log message.
type LogMessage struct {
	level     level.Level
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
// normal usage, LocationSkip should be set to 2. Location information is only
// included for ERROR, DEBUG, and FATAL log levels.
//
// Returns an error if the log directory path cannot be set.
func NewLogger(cfg *config.Config, msgChanSize int, LocationSkip int) (*Logger, error) {
	// Create the logger instance.
	l := &Logger{
		config:        cfg,
		locationSkip:  LocationSkip,
		Running:       true,
		messageChan:   make(chan LogMessage, msgChanSize),
		getConfigChan: make(chan chan config.Config),
		setConfigChan: make(chan config.Config),
		flushSignal:   make(chan struct{}),
		syncFlushChan: make(chan chan struct{}),
		shutdownChan:  make(chan chan struct{}),
	}

	// Apply default values to the configuration.
	l.config.ApplyDefaults()

	// Set the log directory path
	if err := l.setPath(*l.config.DirectoryPath); err != nil {
		return nil, err
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
	l.RunningMutex.Lock()
	defer l.RunningMutex.Unlock()
	return utils.Ternary(l.Running, fmt.Errorf("logger failed to shutdown in time"), nil)
}

// Start restarts the logger goroutine after a shutdown.
func (l *Logger) Start() {
	l.RunningMutex.Lock()
	defer l.RunningMutex.Unlock()
	if !l.Running {
		l.Running = true
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
func (l *Logger) GetConfigCopy() config.Config {
	resp := make(chan config.Config)
	l.getConfigChan <- resp
	return <-resp
}

// UpdateConfig updates the logger configuration with the provided settings.
// Nil fields are ignored.
func (l *Logger) UpdateConfig(cfg config.Config) {
	l.setConfigChan <- cfg
}

// Log message functions. These are the main interface for logging messages.

func (l *Logger) Info(msg string)                   { l.qM(level.INFO, 0, "%s", msg) }
func (l *Logger) Infof(format string, args ...any)  { l.qM(level.INFO, 0, format, args...) }
func (l *Logger) Warn(msg string)                   { l.qM(level.WARN, 0, "%s", msg) }
func (l *Logger) Warnf(format string, args ...any)  { l.qM(level.WARN, 0, format, args...) }
func (l *Logger) Error(msg string)                  { l.qM(level.ERROR, 0, "%s", msg) }
func (l *Logger) Errorf(format string, args ...any) { l.qM(level.ERROR, 0, format, args...) }
func (l *Logger) Debug(msg string)                  { l.qM(level.DEBUG, 0, "%s", msg) }
func (l *Logger) Debugf(format string, args ...any) { l.qM(level.DEBUG, 0, format, args...) }

// Fatal attempts to log a message and exits the program. It exits with the given exit code either when the message is
// logged or the timeout duration is reached. A timeout of 0 means block indefinitely.
func (l *Logger) Fatal(exitCode int, timeout time.Duration, msg string) {
	l.qM(level.FATAL, exitCode, "%s", msg)
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
func (l *Logger) qM(lvl level.Level, exitCode int, format string, args ...any) {
	m := LogMessage{
		level:     lvl,
		exitCode:  exitCode,
		timestamp: time.Now(),
		location:  "",
		content:   fmt.Sprintf(format, args...),
	}
	if l.locationSkip != -1 {
		if (lvl == level.FATAL) || (lvl == level.ERROR) || (lvl == level.DEBUG) {
			if _, file, line, ok := runtime.Caller(l.locationSkip); ok {
				m.location = fmt.Sprintf("%s:%d", filepath.Base(file), line)
			}
		}
	}
	l.messageChan <- m
}

func (l *Logger) handleMessage(m LogMessage) {
	// Check if the message should be logged given the current log level
	if l.config.Level == nil || *l.config.Level == level.NONE {
		return
	}
	if m.level > *l.config.Level {
		return
	}
	// Create the message prefix
	prefix := m.timestamp.Format("[2006-01-02,15-04-05,") + m.level.String() + "] "
	prefix = strutil.Pad(prefix, 28)
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
	if l.config.ConsoleOut.L != nil {
		l.config.ConsoleOut.L.Print(m.content)
	}
	if m.level == level.FATAL {
		l.flush()
		os.Exit(m.exitCode)
	}
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
			l.RunningMutex.Lock()
			l.Running = false
			l.RunningMutex.Unlock()
			return
		case resp := <-l.getConfigChan:
			resp <- *l.config
		case cfg := <-l.setConfigChan:
			utils.CopyIfNotNil(l.config.Level, cfg.Level)
			utils.CopyIfNotNil(l.config.MaxBufferSizeBytes, cfg.MaxBufferSizeBytes)
			utils.CopyIfNotNil(l.config.MaxFileSizeBytes, cfg.MaxFileSizeBytes)
			if cfg.FlushInterval != nil {
				*l.config.FlushInterval = *cfg.FlushInterval
				restartTickerReq = true
			}
			if cfg.DirectoryPath != nil {
				l.setPath(*cfg.DirectoryPath)
			}
			if cfg.ConsoleOut != nil {
				l.config.ConsoleOut.L = cfg.ConsoleOut.L
			}
		}
	}
}
