# Blog · [![Tests](https://github.com/Data-Corruption/blog/actions/workflows/tests.yml/badge.svg)](https://github.com/Data-Corruption/blog/actions/workflows/tests.yml) [![codecov](https://codecov.io/github/Data-Corruption/blog/graph/badge.svg?token=HGC6QI86EG)](https://codecov.io/github/Data-Corruption/blog) ![Go Report](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)

Blog is a small logging package for Go applications. It's designed to be simple, lightweight, and easy to integrate, while still offering the robust functionality needed for most real-world production scenarios.

## Features

### Easy Integration
- **Simple Integration**: Seamlessly integrate Blog with just a few lines of code to enable logging capabilities.
- **Standard Library Only**: Rely's exclusively on Go's standard library, ensuring a lightweight solution minimizing your third-party dependencies.

### Logging Essentials
- **Level-Based Logging**: Choose from multiple log levels (DEBUG, INFO, WARN, ERROR, FATAL) for precise control over logging output.
- **Concurrent Safe**: Ensure dependable logging across concurrent routines with a thread-safe design.
- **Customizable At Runtime**: Adjust settings during runtime, including log level, output directory, limits, etc.

### Efficient Log Management
- **Log File Rotation**: Benefit from automatic size-based log file rotation to manage space efficiently.
- **Transparent Optimizations**: Experience optimized log writing with a rolling buffer. This feature handles size and time-based buffering strategies mostly behind the scenes, but you can manually flush the buffer or configure the maximum size and interval for automatic flushes for finer control.

## Getting Started

```sh
go get github.com/Data-Corruption/blog
```

Basic Example:

```go
import "github.com/Data-Corruption/blog"

func main() {
    // gracefully flush remaining buffer before an exit.
    defer blog.Flush()
    // init logger to write in current working directory at level INFO
    if err := blog.Init("", blog.INFO); err != nil {
        blog.Error("Falling back to console due to out dir issue: " + err.Error())
    }
    blog.Info("This is an informational message")
    // ... rest of your code ...
}
```

Advanced Example:

```go
import "github.com/Data-Corruption/blog"

func main() {
    defer blog.Flush()

    // Convert a string to a LogLevel
    level, ok := blog.LogLevelFromString("Info")
    if !ok {
        // handle unknown string level
    }

    // Initialization error handling
	if err := blog.Init("example/dir", level); err != nil {
		switch err.(type) {
		case blog.AlreadyInitializedError:
			// handle already initialized case
		case blog.InvalidPathError:
			// handle invalid path case
		default:
			// handle other errors
		}
		blog.Error("Initialization failed:", err)
	}

    // Edit log level
    blog.Debug("This will not be logged")
    blog.SetLevel(blog.DEBUG)
    blog.Debug("This will be logged")
    
    // Edit output directory
    blog.Info("This will be logged to ./latest.log")
    blog.SetDirPath("example/dir")
    blog.Info("This will be logged to example/dir/latest.log")

    // Enable console output
    blog.SetUseConsole(true) // if init failed this will already be true
    blog.Info("This will be logged to both file and console")

    // Edit buffer configuration
    blog.SetMaxWriteBufSize(4096) // flush whenever buffer exceeds 4096 chars
    blog.SetFlushInterval(5 * time.Second) // auto flush every 5 seconds

    // Edit file limits
    blog.SetMaxFileSize(1024 * 1024 * 1024)
    // When latest.log exceeds 1GB it is renamed to the 
    // current date and time, then a new latest.log is created.
}
```

## Why Blog?

Blog attempts to stand out by balancing simplicity with capability. It addresses the core needs of logging without the overhead of more complex logging frameworks. It's aimed at projects where a lightweight, no-fuss logger is desired, but reliability and a basic feature set are still required.

Blog is also born out of me wanting to learn how to develop and distribute go projects / libs.

## Contributing

Contributions, issues, and feature requests are welcome! Feel free to check [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to contribute.

## License

Distributed under the MIT License. See [LICENSE.md](LICENSE.md) for more information.
