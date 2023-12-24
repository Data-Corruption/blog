# Blog · [![Tests](https://github.com/Data-Corruption/blog/actions/workflows/tests.yml/badge.svg)](https://github.com/Data-Corruption/blog/actions/workflows/tests.yml) [![codecov](https://codecov.io/github/Data-Corruption/blog/graph/badge.svg?token=HGC6QI86EG)](https://codecov.io/github/Data-Corruption/blog) ![Go Report](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat) [![Go Reference](https://pkg.go.dev/badge/github.com/Data-Corruption/blog.svg)](https://pkg.go.dev/github.com/Data-Corruption/blog) ![GitHub License](https://img.shields.io/github/license/Data-Corruption/blog)

Blog is a async logging package for Go, designed with simplicity and performance in mind. It serves as a straightforward solution for developers looking for efficient, async logging without the complexities often found in larger frameworks. Currently, Blog is focused purely on async operations, with no immediate plans to incorporate synchronous logging.

This project is born from a personal need for a performant logger and a desire to deepen my understanding of Go. While it's primarily a personal endeavor, I welcome anyone who finds it useful for their own projects and am open to evolving it based on user feedback and needs.

**Quick Links:** [Features](#features) · [Getting Started](#getting-started) · [FAQ](#faq) · [Contributing](#contributing) · [License](#license)

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
- **Transparent Optimizations**: Configurable optimized log writing with a rolling buffer. This feature handles size and time-based buffering strategies.

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

  blog.Error("This is an error")
  blog.Warn("This is a warning")
  blog.Info("This is information")
  blog.Debug("This is verbose debug data")
  blog.Fatal("This is a fatal msg that will close the app")
}
```

## FAQ

In this section, you'll find answers to common questions and troubleshooting tips for using our project effectively.

<details>
<summary><b>How to Handle Log File Rotation</b></summary>

Question: What happens when the log file reaches its maximum size, and how can I manage it?

Answer: Blog automatically handles log file rotation based on the size limit you set. Once the latest.log file exceeds the specified maximum size, it's renamed with the current date and time, and a new latest.log file is created. You can adjust the maximum file size using `blog.SetMaxFileSize(size)`. This ensures your logs are manageable and prevents excessive file growth.
</details>

<details>
<summary><b>Changing Settings Dynamically</b></summary>

Question: Can I change the logger's settings at runtime, and how?

Answer: Yes, you can dynamically adjust various settings in the logger. Due to the async nature of the logger these settings may take a few ms to update. Here is a list of available methods to update settings:
- `SetLevel(level LogLevel)`
- `SetUseConsole(use bool)`
- `SetMaxWriteBufSize(size int)`
- `SetMaxFileSize(size int)`
- `SetDirPath(path string)`
- `SetFlushInterval(d time.Duration)`
</details>

<details>
<summary><b>Understanding LogLevelFromString Functionality</b></summary>

Question: What does blog.LogLevelFromString("string") do, and how should I handle unknown log levels?

Answer: The function blog.LogLevelFromString("string") converts a string representation of a log level (like "info" or "debug") into a Blog's LogLevel. If the string doesn't match any known log levels, it returns false as the second return value. The case of the characters is irrelevant as they are all upcased before checking.
</details>

<details>
<summary><b>Configuring Buffer and Flush Settings</b></summary>

**Question**: How can I optimize performance by configuring the internal buffer and flush intervals?

**Answer**: Blog optimizes log writing using a rolling buffer, which automatically flushes based on two configurable events:

  - **Buffer Size Limit Reached**: When the buffer accumulates to a certain size, it triggers a flush. You can set this threshold with `blog.SetMaxWriteBufSize(size)`. The default size is 4KB. Adjusting this allows you to balance between performance and real-time logging based on your application's needs.

  - **Time Interval Elapsed**: The buffer also flushes periodically after a specified time interval, ensuring logs are written even during low activity. Set this interval with `blog.SetFlushInterval(amountOfTime)`. The default interval is 5 seconds. Shortening this time ensures more frequent writes, while lengthening it can reduce disk I/O in less critical applications.

Both settings are crucial for tailoring Blog's performance to match your specific logging requirements and operational environment.
</details>

<details>
<summary><b>Using Blog in Concurrent Environments</b></summary>

**Question**: Is Blog suitable for concurrent environments, and are there any special considerations for synchronous operations?

**Answer**: Blog is inherently safe for concurrent use in applications utilizing multiple goroutines. It manages access to log files asynchronously, ensuring thread safety without the need for additional synchronization in most scenarios. However, due to its asynchronous nature, if you require synchronous logging, i recommend checking out one of GO's many libs that support sync operation, here is one if interested:
- **zap**: https://github.com/uber-go/zap
</details>

<details>
<summary><b>Logging Issue: Message not appearing in the log file</b></summary>
  
**Question**: After logging a message, flushing, and then reading the log file, why doesn't it contain my message?

**Answer**: This is likely due to the asynchronous nature of our logging system, which utilizes goroutines and channels. These processes require some time to execute. To resolve this:
  - **Step 1**: Wait for a few milliseconds after logging your message before flushing.
  - **Step 2**: Similarly, wait for a few milliseconds after flushing before you attempt to read the log file.o
These steps ensure that the system has enough time to process your requests.

</details>


## Contributing

Contributions, issues, and feature requests are welcome! Feel free to check [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to contribute.

## License

Distributed under the MIT License. See [LICENSE.md](LICENSE.md) for more information.
