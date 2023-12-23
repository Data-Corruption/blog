// Package blog implements a simple, thread-safe singleton logger.
// It supports various log levels and can write to files or the console.
package blog

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Variables ==================================================================

// default constants define the standard behavior and limits of the logger.
const (
	defaultMaxLogBufSize   = 255
	defaultMaxWriteBufSize = 4096
	defaultMaxFileSize     = 1024 * 1024 * 1024 // 1 GB
	defaultFlushInterval   = 5 * time.Second
)

var (
	// instance is the singleton instance of the logger.
	instance *logger = nil
	// run channels
	manualFlushChan       chan struct{}      = make(chan struct{})
	logMsgChan            chan message       = make(chan message, defaultMaxLogBufSize)
	updateLevel           chan LogLevel      = make(chan LogLevel)
	updateUseConsole      chan bool          = make(chan bool)
	updateMaxWriteBufSize chan int           = make(chan int)
	updateMaxFileSize     chan int           = make(chan int)
	updateFlushInterval   chan time.Duration = make(chan time.Duration)
	updateDirPath         chan string        = make(chan string)
	// used for testing
	reqStateChan chan struct{}  = make(chan struct{})
	resStateChan chan logger    = make(chan logger)
	runExitChan  chan struct{}  = make(chan struct{})
	runWaitGroup sync.WaitGroup = sync.WaitGroup{}
)

// Types Definitions ==========================================================

// LogLevel defines the severity levels for logging messages.
type LogLevel int

const (
	NONE LogLevel = iota
	ERROR
	WARN
	INFO
	DEBUG
	FATAL
)

// AlreadyInitializedError indicates Init has already been called.
type AlreadyInitializedError struct{}

// InvalidPathError occurs when a provided directory path is invalid or inaccessible.
type InvalidPathError struct {
	Path string
}

// message represents a single log entry.
type message struct {
	level     LogLevel
	exitCode  int // only used by FATAL messages
	timestamp time.Time
	content   string
}

// logger is the main struct for the blog package, handling all logging operations.
type logger struct {
	level           LogLevel
	useConsole      bool
	maxWriteBufSize int
	maxFileSize     int
	flushInterval   time.Duration
	dirPath         string
	latestPath      string // dirPath/latest.log
	writeBuffer     bytes.Buffer
}

// Exported Functions =========================================================

// Init sets up the logger with the specified directory path and log level.
// It returns an error if called more than once or if the directory path is invalid.
// On error, logging falls back to the console. See AlreadyInitializedError and InvalidPathError.
func Init(dirPath string, level LogLevel) error {
	if instance != nil {
		return AlreadyInitializedError{}
	}

	instance = &logger{
		level:           level,
		useConsole:      false,
		dirPath:         "",
		maxWriteBufSize: defaultMaxWriteBufSize,
		maxFileSize:     defaultMaxFileSize,
		flushInterval:   defaultFlushInterval,
		latestPath:      "",
		writeBuffer:     bytes.Buffer{},
	}

	err := instance.setDirPath(dirPath)
	if err != nil {
		instance.useConsole = true
		instance.dirPath = ""
		err = InvalidPathError{dirPath}
	}

	// start the run goroutine
	runWaitGroup.Add(1)
	go instance.run()

	return err
}

// LogLevelFromString converts a string to a LogLevel, returning false if the string is unrecognized.
func LogLevelFromString(levelStr string) (LogLevel, bool) {
	switch strings.ToUpper(levelStr) {
	case "NONE":
		return NONE, true
	case "ERROR":
		return ERROR, true
	case "WARN":
		return WARN, true
	case "INFO":
		return INFO, true
	case "DEBUG":
		return DEBUG, true
	case "FATAL":
		return FATAL, true
	default:
		return NONE, false
	}
}

// Flush manually flushes the log write buffer.
func Flush() { manualFlushChan <- struct{}{} }

// SetLevel sets the log level.
func SetLevel(level LogLevel) { updateLevel <- level }

// SetUseConsole sets whether or not to log to the console.
func SetUseConsole(use bool) { updateUseConsole <- use }

// SetMaxWriteBufSize sets the maximum size of the log write buffer.
func SetMaxWriteBufSize(size int) { updateMaxWriteBufSize <- size }

// SetMaxFileSize sets the maximum size of the log file. When the log file reaches
// this size, it is renamed to the current timestamp and a new log file is created.
func SetMaxFileSize(size int) { updateMaxFileSize <- size }

// SetDirPath sets the directory path for the log files. If dirPath is empty, the
// current working directory is used.
func SetDirPath(path string) { updateDirPath <- path }

// SetFlushInterval sets the interval at which the log write buffer is automatically
// flushed to the log file.
func SetFlushInterval(d time.Duration) { updateFlushInterval <- d }

// Error logs an error message.
func Error(msg string) { logMsgChan <- message{ERROR, 0, time.Now(), msg} }

// Warn logs a warning message.
func Warn(msg string) { logMsgChan <- message{WARN, 0, time.Now(), msg} }

// Info logs an info message.
func Info(msg string) { logMsgChan <- message{INFO, 0, time.Now(), msg} }

// Debug logs a debug message.
func Debug(msg string) { logMsgChan <- message{DEBUG, 0, time.Now(), msg} }

// Fatal logs a fatal message and exits with the given exit code.
func Fatal(msg string, c int) { logMsgChan <- message{FATAL, c, time.Now(), msg} }

// ======== Unexported Functions ========

func padString(s string, length int) string {
	if len(s) < length {
		return s + strings.Repeat(" ", length-len(s))
	}
	return s
}

func (e AlreadyInitializedError) Error() string {
	return "blog: already initialized"
}

func (e InvalidPathError) Error() string {
	return fmt.Sprintf("blog: invalid path: %s", e.Path)
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

func (l *logger) genLogPath() string {
	if l.dirPath == "" {
		panic("dirPath in blog is nil, this should never happen")
	}
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return filepath.Join(l.dirPath, timestamp+".log")
}

func (l *logger) setDirPath(path string) error {
	// if path is empty, set to current working directory
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		l.dirPath = cwd
		l.latestPath = filepath.Join(cwd, "latest.log")
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

func (l *logger) handleMessage(msg message) {
	// Check if the message should be logged given the current log level
	if !(msg.level != NONE && (msg.level == FATAL || msg.level <= l.level)) {
		return
	}
	// Format the message
	msg.content = fmt.Sprintf("%s,%s: %s\n", msg.timestamp.Format("2006-01-02,15-04-05"), msg.level.toString(), msg.content)
	// If a file path is set, write the message to the log file
	if l.dirPath != "" {
		l.writeBuffer.WriteString(msg.content)
		if l.writeBuffer.Len() >= l.maxWriteBufSize {
			l.flush()
		}
	}
	// If console logging is enabled, write the message to the console
	if l.useConsole {
		log.Print(msg.content)
	}
}

// handleFileOverflow renames latest.log to the current timestamp and creates a new latest.log.
func (l *logger) handleFileOverflow() (*os.File, error) {
	// rename latest.log to the current timestamp
	if err := os.Rename(l.latestPath, l.genLogPath()); err != nil {
		return nil, err
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
	// print the error to the console
	log.Printf("Falling back to console logging due to an error flushing the log write buffer: %v", err)
	// set use console to true and dir path to nil which will disable file logging
	l.useConsole = true
	l.dirPath = ""
	// print the remaining write buffer to the console
	log.Print(l.writeBuffer)
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
		case <-manualFlushChan:
			l.flush()
		case msg := <-logMsgChan:
			l.handleMessage(msg)
			if msg.level == FATAL {
				l.flush()
				os.Exit(1)
			}
		case level := <-updateLevel:
			l.level = level
		case useConsole := <-updateUseConsole:
			l.useConsole = useConsole
		case maxWriteBufSize := <-updateMaxWriteBufSize:
			l.maxWriteBufSize = maxWriteBufSize
		case maxFileSize := <-updateMaxFileSize:
			l.maxFileSize = maxFileSize
		case flushInterval := <-updateFlushInterval:
			l.flushInterval = flushInterval
			shouldRestart = true
		case dirPath := <-updateDirPath:
			err := l.setDirPath(dirPath)
			if err != nil {
				l.useConsole = true
				l.dirPath = ""
			}
		case <-reqStateChan:
			resStateChan <- *l
		case <-runExitChan:
			runWaitGroup.Done()
			return
		}
	}
}

// Test Related ===============================================================

// getCopyOfInstance is used for testing. It returns a copy of the current logger instance.
// The purpose of this is to allow reading state without blocking the run goroutine.
func getCopyOfInstance() logger {
	reqStateChan <- struct{}{}
	return <-resStateChan
}

// reset is used for testing. It shuts down the run goroutine and resets all variables.
func reset() {
	if instance == nil {
		return
	}
	close(runExitChan)
	runWaitGroup.Wait()
	instance = nil
	// reset run channels and wait group
	manualFlushChan = make(chan struct{})
	logMsgChan = make(chan message, defaultMaxLogBufSize)
	updateLevel = make(chan LogLevel)
	updateUseConsole = make(chan bool)
	updateMaxWriteBufSize = make(chan int)
	updateMaxFileSize = make(chan int)
	updateFlushInterval = make(chan time.Duration)
	updateDirPath = make(chan string)
	reqStateChan = make(chan struct{})
	resStateChan = make(chan logger)
	runExitChan = make(chan struct{})
	runWaitGroup = sync.WaitGroup{}
}
