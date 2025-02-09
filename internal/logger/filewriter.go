package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Data-Corruption/blog/v3/internal/config"
	"github.com/Data-Corruption/blog/v3/internal/utils/strutil"
)

// fallbackToConsole disables file logging and enables console logging if not already enabled. Also passes the given error through.
func (l *Logger) fallbackToConsole() {
	*l.config.DirectoryPath = ""
	if l.config.ConsoleOut.L == nil {
		l.config.ConsoleOut = &config.ConsoleLogger{L: log.New(os.Stdout, "", 0)}
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
		return fmt.Errorf("blog: failed to stat path: %w", err)
	}
	if !fileInfo.IsDir() {
		l.fallbackToConsole()
		return fmt.Errorf("blog: path is not a directory: %s", cleanedPath)
	}
	// Set the directory path
	*l.config.DirectoryPath = cleanedPath
	return nil
}

// getLatestPath returns the path to the latest.log file.
func (l *Logger) getLatestPath() string {
	return filepath.Join(*l.config.DirectoryPath, "latest.log")
}

// rotatedFilename returns a new path for latest.log to be renamed to.
func rotatedFilename(dir string) (string, error) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	name := timestamp + ".log"
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); err == nil {
		randomSuffix, err := strutil.Random(8)
		if err != nil {
			return "", err
		}
		name = fmt.Sprintf("%s_%s.log", timestamp, randomSuffix)
		path = filepath.Join(dir, name)
	}
	return path, nil
}

// handleFlushError prints the error to the console, sets use console to true and dir path to nil,
// effectively disabling file logging, and prints the remaining write buffer to the console.
func (l *Logger) handleFlushError(err error) {
	l.fallbackToConsole()
	// print the remaining write buffer to the console
	l.config.ConsoleOut.L.Printf("failed to write to log file: %v", err)
	l.config.ConsoleOut.L.Print(l.writeBuffer.String())
	l.writeBuffer.Reset()
}

func (l *Logger) rotateLogFile() error {
	// Get the new filename
	path, err := rotatedFilename(*l.config.DirectoryPath)
	if err != nil {
		return fmt.Errorf("failed to get rotated filename: %w", err)
	}
	// Rename latest.log to the current timestamp
	if err := os.Rename(l.getLatestPath(), path); err != nil {
		return fmt.Errorf("failed to rename latest.log: %w", err)
	}
	// Create a new latest.log with the write buffer
	if overflow, err := l.writeIfUnderMaxFileSize(); err != nil {
		return fmt.Errorf("failed to write to latest.log: %w", err)
	} else if overflow {
		return fmt.Errorf("rotated log file is still too large")
	}
	return nil
}

// flush writes the buffered log to the filesystem and resets the buffer.
func (l *Logger) flush() {
	if (l.writeBuffer.Len() == 0) || (*l.config.DirectoryPath == "") {
		return
	}
	// write the buffer to the file
	if overflow, err := l.writeIfUnderMaxFileSize(); err != nil {
		l.handleFlushError(fmt.Errorf("blog: failed to write to log file: %w", err))
		return
	} else if overflow {
		if err := l.rotateLogFile(); err != nil {
			l.handleFlushError(fmt.Errorf("blog: failed to rotate log file: %w", err))
		}
	}
}

// write writes the buffered log to the file if the file is under the maximum size.
// Returns true if the file was too large and needs to be rotated.
func (l *Logger) writeIfUnderMaxFileSize() (bool, error) {
	// Open the log file
	f, err := os.OpenFile(l.getLatestPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false, fmt.Errorf("failed to open log file: %w", err)
	}
	defer f.Close()
	// Check if the log file is too large
	fileInfo, err := f.Stat()
	if err != nil {
		return false, fmt.Errorf("failed to stat log file: %w", err)
	}
	// If the log file is too large, return true
	if fileInfo.Size() >= int64(*l.config.MaxFileSizeBytes) {
		return true, nil
	}
	// Write the buffered log to the file
	if _, err := f.Write(l.writeBuffer.Bytes()); err != nil {
		return false, fmt.Errorf("failed to write to log file: %w", err)
	}
	// Reset the buffer
	l.writeBuffer.Reset()
	return false, nil
}
