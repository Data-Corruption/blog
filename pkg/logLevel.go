package blog

import "strings"

type LogLevel int

const (
	NONE LogLevel = iota
	ERROR
	WARN
	INFO
	DEBUG
	FATAL
)

// String returns the string representation of a LogLevel.
func (l *LogLevel) String() string {
	// switch for perf
	switch *l {
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

// FromString sets a LogLevel from a string, returning ErrInvalidLogLevel if the string is invalid.
// Case-insensitive. Example: "ERROR" -> ERROR, "error" -> ERROR, "Error" -> ERROR, etc.
func (l *LogLevel) FromString(levelStr string) error {
	fromStrMap := map[string]LogLevel{"NONE": NONE, "ERROR": ERROR, "WARN": WARN, "INFO": INFO, "DEBUG": DEBUG, "FATAL": FATAL}
	if level, ok := fromStrMap[strings.ToUpper(levelStr)]; ok {
		*l = level
		return nil
	}
	return ErrInvalidLogLevel
}
