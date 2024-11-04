package blog

type LogLevel int

const (
	NONE LogLevel = iota
	ERROR
	WARN
	INFO
	DEBUG
	FATAL
)

func (l LogLevel) String() string {
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
