package level

import (
	"fmt"
	"strings"
)

type Level int

const (
	NONE Level = iota
	ERROR
	WARN
	INFO
	DEBUG
	FATAL
)

// String returns the string representation of a blog.Level.
func (l Level) String() string {
	// switch for perf
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

// FromString sets a blog.Level from a string, returning ErrInvalidLogLevel if the string is invalid.
// Case-insensitive. Example: "ERROR" -> ERROR, "error" -> ERROR, "Error" -> ERROR, etc.
func (l *Level) FromString(levelStr string) error {
	switch strings.ToUpper(levelStr) {
	case "NONE":
		*l = NONE
	case "ERROR":
		*l = ERROR
	case "WARN":
		*l = WARN
	case "INFO":
		*l = INFO
	case "DEBUG":
		*l = DEBUG
	case "FATAL":
		*l = FATAL
	default:
		return fmt.Errorf("blog: invalid log level")
	}
	return nil
}
