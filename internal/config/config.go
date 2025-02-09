package config

import (
	"log"
	"time"

	"github.com/Data-Corruption/blog/v3/internal/level"
	"github.com/Data-Corruption/blog/v3/internal/utils"
)

var (
	DefaultLevel              level.Level   = level.INFO
	DefaultMaxBufferSizeBytes int           = 4096               // 4 KB
	DefaultMaxFileSizeBytes   int           = 1024 * 1024 * 1024 // 1 GB
	DefaultFlushInterval      time.Duration = 15 * time.Second   // 15 seconds
	DefaultDirectoryPath      string        = "."
)

// ConsoleLogger wraps *log.Logger to allow nil value semantics for disabled state
type ConsoleLogger struct {
	L *log.Logger
}

// Config holds the configuration settings for the Logger.
type Config struct {
	Level              *level.Level   // the minimum log level to write. Default is INFO.
	MaxBufferSizeBytes *int           // the maximum size of the write buffer before it is flushed. Default is 4 KB.
	MaxFileSizeBytes   *int           // the maximum size of the log file before it is rotated. Default is 1 GB.
	FlushInterval      *time.Duration // the interval at which the write buffer is flushed. Default is 15 seconds.
	DirectoryPath      *string        // the directory path where the log file is stored. Default is the current working directory ("."). To disable file logging, set this to an empty string.
	ConsoleOut         *ConsoleLogger // the logger to write to the console. Default is ConsoleLogger{l: nil}. When l is nil, console logging is disabled. This is configurable for easy testing.
}

// ApplyDefaults applies the default values to the given Config if they are nil.
func (cfg *Config) ApplyDefaults() {
	if cfg == nil {
		cfg = &Config{}
	}
	utils.SetDefaultIfNil(&cfg.Level, &DefaultLevel)
	utils.SetDefaultIfNil(&cfg.MaxBufferSizeBytes, &DefaultMaxBufferSizeBytes)
	utils.SetDefaultIfNil(&cfg.MaxFileSizeBytes, &DefaultMaxFileSizeBytes)
	utils.SetDefaultIfNil(&cfg.FlushInterval, &DefaultFlushInterval)
	utils.SetDefaultIfNil(&cfg.DirectoryPath, &DefaultDirectoryPath)
	if cfg.ConsoleOut == nil {
		cfg.ConsoleOut = &ConsoleLogger{}
	}
}
