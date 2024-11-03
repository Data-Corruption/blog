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
  - Default max buf and file size are 4 KB and 1 GB respectively.
  - A single thread is used to handle all logging operations.
    The channel that feeds it messages is buffered to 255 via a constant. If the buffer becomes full, log funcs will block.
    This shouldn't be an issue as if the flush fails for whatever reason, the logger will fall back to console logging.
    Worst case, you parallel log in mass and the blocking becomes a bottleneck.

For contributors:

	The approach is pretty straightforward. The logger is a type with a bunch of channels for communication and vars for configuration.
	When created it starts a goroutine that listens for messages and config updates then handles them.
	The public functions don't interact with the logger directly, they do so using the channels.

	Tests should create their own logger instances using newLogger() and use the 'r' prefixed functions to interact with them.
	newLogger() lets you set the output writer for testing purposes. The public Init sets it to os.Stdout
	The public functions create and use a singleton instance of the logger.

	This has some nice benefits:
	- Easily test multiple logger instances in parallel.
	- Users don't need to manage the logger instance themselves.
*/
package blog

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Public =====================================================================

// ==== Types ====

type LogLevel int

const (
	NONE LogLevel = iota
	ERROR
	WARN
	INFO
	DEBUG
	FATAL
)

// ==== Variables ====

var (
	ErrAlreadyInitialized = fmt.Errorf("blog: already initialized")
	ErrInvalidLogLevel    = fmt.Errorf("blog: invalid log level")
	ErrUninitialized      = fmt.Errorf("blog: uninitialized")
	ErrInvalidPath        = fmt.Errorf("blog: invalid path")
)

// ==== Functions ====

// Init sets up the logger with the specified directory path and log level.
// It returns an error if called more than once or if the directory path is invalid.
// On error, logging falls back to the console. See ErrAlreadyInitialized and ErrInvalidPath.
func Init(dirPath string, level LogLevel, includeLineNum bool) error {
	if instance != nil {
		return ErrAlreadyInitialized
	}
	var err error
	instance, err = newLogger(dirPath, level, includeLineNum, os.Stdout)
	return err
}

// Cleanup flushes the log write buffer and exits the logger. If timeout is 0, Cleanup blocks indefinitely.
func Cleanup(timeout time.Duration) error {
	return c(func() error { return instance.rCleanup(timeout) })
}

// LogLevelFromString converts a string to a LogLevel, returning ErrInvalidLogLevel if the string is invalid.
func LogLevelFromString(levelStr string) (LogLevel, error) {
	fromStrMap := map[string]LogLevel{"NONE": NONE, "ERROR": ERROR, "WARN": WARN, "INFO": INFO, "DEBUG": DEBUG, "FATAL": FATAL}
	if level, ok := fromStrMap[strings.ToUpper(levelStr)]; ok {
		return level, nil
	}
	return NONE, ErrInvalidLogLevel
}

// Flush manually flushes the log write buffer.
func Flush() error { return c(func() error { return instance.rFlush() }) }

// SyncFlush synchronously flushes the log write buffer and blocks until the flush is complete or the timeout is reached. If timeout is 0, SyncFlush blocks indefinitely.
func SyncFlush(timeout time.Duration) error {
	return c(func() error { return instance.rSyncFlush(timeout) })
}

// SetLevel sets the log level.
func SetLevel(level LogLevel) error {
	return c(func() error { return instance.rSetLevel(level) })
}

// SetUseConsole sets whether or not to log to the console.
func SetUseConsole(use bool) error { return c(func() error { return instance.rSetUseConsole(use) }) }

// SetMaxWriteBufSize sets the maximum size of the log write buffer.
func SetMaxWriteBufSize(size int) error {
	return c(func() error { return instance.rSetMaxWriteBufSize(size) })
}

// SetMaxFileSize sets the maximum size of the log file. When the log file reaches
// this size, it is renamed to the current timestamp and a new log file is created.
func SetMaxFileSize(size int) error { return c(func() error { return instance.rSetMaxFileSize(size) }) }

// SetDirPath sets the directory path for the log files. When dirPath is an empty string, file logging is disabled.
func SetDirPath(path string) error { return c(func() error { return instance.rSetDirPath(path) }) }

// SetFlushInterval sets the interval at which the log write buffer is automatically
// flushed to the log file.
func SetFlushInterval(d time.Duration) error {
	return c(func() error { return instance.rSetFlushInterval(d) })
}

// ==== Logging Functions ====

func Error(msg string) error { return c(func() error { return instance.rError(msg) }) }
func Errorf(format string, args ...any) error {
	return c(func() error { return instance.rErrorf(format, args...) })
}
func Warn(msg string) error { return c(func() error { return instance.rWarn(msg) }) }
func Warnf(format string, args ...any) error {
	return c(func() error { return instance.rWarnf(format, args...) })
}
func Info(msg string) error { return c(func() error { return instance.rInfo(msg) }) }
func Infof(format string, args ...any) error {
	return c(func() error { return instance.rInfof(format, args...) })
}
func Debug(msg string) error { return c(func() error { return instance.rDebug(msg) }) }
func Debugf(format string, args ...any) error {
	return c(func() error { return instance.rDebugf(format, args...) })
}

// Fatal logs a fatal message and exits with the given exit code.
// This function will not return, it will exit the program after attempting to log the message.
func Fatal(exitCode int, timeout time.Duration, msg string) error {
	return c(func() error { return instance.rFatal(exitCode, timeout, msg) })
}

// Fatalf logs a fatal message with a format string and exits with the given exit code.
// This function will not return, it will exit the program after attempting to log the message.
func Fatalf(exitCode int, timeout time.Duration, format string, args ...any) error {
	return c(func() error { return instance.rFatalf(exitCode, timeout, format, args...) })
}

// Internal ===================================================================

// ==== Types ====

// message represents a single log entry.
type message struct {
	level     LogLevel
	exitCode  int // only used by FATAL messages
	timestamp time.Time
	lineNum   string // [file:line]
	content   string
}

// logger is the main struct for the blog package, handling all logging operations.
type logger struct {
	stdLogger       *log.Logger // dedicated logger instance for removing timestamps without affecting the global logger
	level           LogLevel
	useConsole      bool
	includeLineNum  bool
	maxWriteBufSize int
	maxFileSize     int
	flushInterval   time.Duration
	dirPath         string
	latestPath      string // dirPath/latest.log
	// run channels
	flushChan             chan struct{}
	logMsgChan            chan message
	updateLevel           chan LogLevel
	updateUseConsole      chan bool
	updateMaxWriteBufSize chan int
	updateMaxFileSize     chan int
	updateFlushInterval   chan time.Duration
	updateDirPath         chan string
	// sync flush stuff
	syncFlushChan  chan struct{}
	syncFlushDone  chan struct{}
	syncFlushMutex sync.Mutex
	// exit stuff
	runExitChan  chan struct{}
	runWaitGroup sync.WaitGroup
	// log buffer
	writeBuffer bytes.Buffer
}

// ==== Variables ====

// default constants define the standard behavior and limits of the logger.
const (
	defaultMaxMsgChanBufSize = 255
	defaultMaxWriteBufSize   = 4096               // 4 KB
	defaultMaxFileSize       = 1024 * 1024 * 1024 // 1 GB
	defaultFlushInterval     = 5 * time.Second
)

// instance is the singleton instance of the logger.
// Tests don't use this or the public functions.
// They create their own logger instances and use 'r' prefixed functions.
var instance *logger = nil

// ==== Functions ====

// helper function that only executes the given function if the logger is initialized.
func c(f func() error) error {
	if instance == nil {
		return ErrUninitialized
	}
	return f()
}

func padString(s string, length int) string {
	if len(s) < length {
		return s + strings.Repeat(" ", length-len(s))
	}
	return s
}

func (l LogLevel) toString() string {
	switch l {
	case NONE:
		return "NONE"
	case ERROR:
		return "ERROR"
	case WARN:
		return "WARN"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case FATAL:
		return "FATAL"
	default:
		return "?"
	}
}

func newLogger(dirPath string, level LogLevel, includeLineNum bool, consoleOutput io.Writer) (*logger, error) {
	newLogger := &logger{
		stdLogger:             log.New(consoleOutput, "", 0),
		level:                 level,
		useConsole:            false,
		includeLineNum:        includeLineNum,
		dirPath:               "",
		maxWriteBufSize:       defaultMaxWriteBufSize,
		maxFileSize:           defaultMaxFileSize,
		flushInterval:         defaultFlushInterval,
		latestPath:            "",
		flushChan:             make(chan struct{}),
		logMsgChan:            make(chan message, defaultMaxMsgChanBufSize),
		updateLevel:           make(chan LogLevel),
		updateUseConsole:      make(chan bool),
		updateMaxWriteBufSize: make(chan int),
		updateMaxFileSize:     make(chan int),
		updateFlushInterval:   make(chan time.Duration),
		updateDirPath:         make(chan string),
		syncFlushChan:         make(chan struct{}),
		syncFlushDone:         make(chan struct{}),
		syncFlushMutex:        sync.Mutex{},
		runExitChan:           make(chan struct{}),
		runWaitGroup:          sync.WaitGroup{},
		writeBuffer:           bytes.Buffer{},
	}
	// set the log directory path
	err := newLogger.setPath(dirPath)
	if err != nil {
		newLogger.useConsole = true
		newLogger.dirPath = ""
		err = ErrInvalidPath
	}
	// start the run goroutine and return the logger
	newLogger.runWaitGroup.Add(1)
	go newLogger.run()
	return newLogger, err
}

func (l *logger) rCleanup(timeout time.Duration) error {
	// wait until run exits or timeout
	done := make(chan struct{})
	go func() {
		l.rSyncFlush(timeout)
		l.runExitChan <- struct{}{}
		l.runWaitGroup.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
	}
	instance = nil
	return nil
}

func (l *logger) rFlush() error { l.flushChan <- struct{}{}; return nil }

func (l *logger) rSyncFlush(timeout time.Duration) error {
	l.syncFlushMutex.Lock()
	defer l.syncFlushMutex.Unlock()
	l.syncFlushChan <- struct{}{}
	if timeout == 0 {
		<-l.syncFlushDone
	} else {
		select {
		case <-l.syncFlushDone:
		case <-time.After(timeout):
		}
	}
	return nil
}

func (l *logger) rSetLevel(level LogLevel) error          { l.updateLevel <- level; return nil }
func (l *logger) rSetUseConsole(use bool) error           { l.updateUseConsole <- use; return nil }
func (l *logger) rSetMaxWriteBufSize(size int) error      { l.updateMaxWriteBufSize <- size; return nil }
func (l *logger) rSetMaxFileSize(size int) error          { l.updateMaxFileSize <- size; return nil }
func (l *logger) rSetDirPath(path string) error           { l.updateDirPath <- path; return nil }
func (l *logger) rSetFlushInterval(d time.Duration) error { l.updateFlushInterval <- d; return nil }

func (l *logger) rError(msg string) error { l.qMsg(ERROR, 0, msg); return nil }
func (l *logger) rErrorf(format string, args ...any) error {
	l.qMsg(ERROR, 0, format, args...)
	return nil
}
func (l *logger) rWarn(msg string) error { l.qMsg(WARN, 0, msg); return nil }
func (l *logger) rWarnf(format string, args ...any) error {
	l.qMsg(WARN, 0, format, args...)
	return nil
}
func (l *logger) rInfo(msg string) error { l.qMsg(INFO, 0, msg); return nil }
func (l *logger) rInfof(format string, args ...any) error {
	l.qMsg(INFO, 0, format, args...)
	return nil
}
func (l *logger) rDebug(msg string) error { l.qMsg(DEBUG, 0, msg); return nil }
func (l *logger) rDebugf(format string, args ...any) error {
	l.qMsg(DEBUG, 0, format, args...)
	return nil
}

func (l *logger) rFatal(exitCode int, timeout time.Duration, msg string) error {
	l.qMsg(FATAL, exitCode, msg)
	time.Sleep(40 * time.Millisecond)
	l.rSyncFlush(0)
	os.Exit(exitCode)
	return nil
}
func (l *logger) rFatalf(exitCode int, timeout time.Duration, format string, args ...any) error {
	l.qMsg(FATAL, exitCode, format, args...)
	time.Sleep(40 * time.Millisecond)
	l.rSyncFlush(0)
	os.Exit(exitCode)
	return nil
}

// qMsg queues a message to be logged.
func (l *logger) qMsg(level LogLevel, code int, format string, args ...any) {
	lineNum := ""
	if l.includeLineNum {
		if _, file, line, ok := runtime.Caller(2); ok {
			lineNum = fmt.Sprintf("[%s:%d]", filepath.Base(file), line)
		}
	}
	l.logMsgChan <- message{level, code, time.Now(), lineNum, fmt.Sprintf(format, args...)}
}

func (l *logger) genLogPath() string {
	if l.dirPath == "" {
		return ""
	}
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return filepath.Join(l.dirPath, timestamp+".log")
}

func (l *logger) setPath(path string) error {
	// if path is empty, clear the dir path and latest path
	if path == "" {
		l.dirPath = ""
		l.latestPath = ""
		return nil
	}
	// check if the path exists and is a directory
	cleanedPath := filepath.Clean(path)
	fileInfo, err := os.Stat(cleanedPath)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return os.ErrNotExist
	}
	// set the dir path and update the latest path
	l.dirPath = cleanedPath
	l.latestPath = filepath.Join(cleanedPath, "latest.log")
	return nil
}

func (l *logger) shouldLog(level LogLevel) bool {
	if level == FATAL {
		return true
	} else if level == NONE {
		return false
	}
	return level <= l.level
}

func (l *logger) handleMessage(msg message) {
	// Check if the message should be logged given the current log level
	if !l.shouldLog(msg.level) {
		return
	}
	// create the message prefix
	prefix := msg.timestamp.Format("[2006-01-02,15-04-05,") + msg.level.toString() + "] "
	// make sure the prefix is at least 28 characters long
	prefix = padString(prefix, 28)
	// add the line number if it exists
	if msg.lineNum != "" {
		prefix += msg.lineNum + " "
	}
	// Format the message
	msg.content = prefix + msg.content + "\n"
	// If a file path is set, write the message to the log file
	if l.dirPath != "" {
		l.writeBuffer.WriteString(msg.content)
		if l.writeBuffer.Len() >= l.maxWriteBufSize {
			l.flush()
		}
	}
	// If console logging is enabled, write the message to the console
	if l.useConsole {
		l.stdLogger.Print(msg.content)
	}
}

// handleFileOverflow renames latest.log to the current timestamp and creates a new latest.log.
func (l *logger) handleFileOverflow() (*os.File, error) {
	// rename latest.log to the current timestamp
	if newName := l.genLogPath(); newName == "" {
		return nil, ErrInvalidPath
	} else {
		if err := os.Rename(l.latestPath, newName); err != nil {
			return nil, err
		}
	}
	// create a new latest.log
	f, err := os.OpenFile(l.latestPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// handleFlushError prints the error to the console, sets use console to true and dir path to nil,
// effectively disabling file logging, and prints the remaining write buffer to the console.
func (l *logger) handleFlushError(err error) {
	log.Printf("Falling back to console logging due to an error flushing the log write buffer: %v", err)
	l.useConsole = true
	l.dirPath = "" // disable file logging
	// print the remaining write buffer to the console
	log.Print(l.writeBuffer.String())
	l.writeBuffer.Reset()
}

func (l *logger) flush() {
	if (l.writeBuffer.Len() == 0) || (l.dirPath == "") {
		return
	}
	// Open the log file
	f, err := os.OpenFile(l.latestPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
	if fileInfo.Size() >= int64(l.maxFileSize) {
		f.Close()
		// Rename latest.log to the current timestamp and create a new latest.log
		f, err = l.handleFileOverflow()
		if err != nil {
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

// run is the main loop for the logger.
func (l *logger) run() {
	defer l.runWaitGroup.Done()
	var ticker *time.Ticker = time.NewTicker(l.flushInterval)
	defer ticker.Stop()
	shouldRestart := false

	for {
		if shouldRestart {
			ticker.Stop()
			ticker = time.NewTicker(l.flushInterval)
			shouldRestart = false
		}
		select {
		case <-ticker.C:
			l.flush()
		case <-l.flushChan:
			l.flush()
		case <-l.syncFlushChan:
			l.flush()
			l.syncFlushDone <- struct{}{}
		case msg := <-l.logMsgChan:
			if msg.level == FATAL {
				l.useConsole = true
			}
			l.handleMessage(msg)
			if msg.level == FATAL {
				l.flush()
				os.Exit(msg.exitCode)
			}
		case level := <-l.updateLevel:
			l.level = level
		case useConsole := <-l.updateUseConsole:
			l.useConsole = useConsole
		case maxWriteBufSize := <-l.updateMaxWriteBufSize:
			l.maxWriteBufSize = maxWriteBufSize
		case maxFileSize := <-l.updateMaxFileSize:
			l.maxFileSize = maxFileSize
		case flushInterval := <-l.updateFlushInterval:
			l.flushInterval = flushInterval
			shouldRestart = true
		case dirPath := <-l.updateDirPath:
			err := l.setPath(dirPath)
			if err != nil {
				l.useConsole = true
				l.dirPath = ""
			}
		case <-l.runExitChan:
			return
		}
	}
}
