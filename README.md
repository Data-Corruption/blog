# Blog · [![Tests](https://github.com/Data-Corruption/blog/actions/workflows/tests.yml/badge.svg)](https://github.com/Data-Corruption/blog/actions/workflows/tests.yml) [![codecov](https://codecov.io/github/Data-Corruption/blog/graph/badge.svg?token=HGC6QI86EG)](https://codecov.io/github/Data-Corruption/blog) ![Go Report](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat) [![Go Reference](https://pkg.go.dev/badge/github.com/Data-Corruption/blog.svg)](https://pkg.go.dev/github.com/Data-Corruption/blog) ![GitHub License](https://img.shields.io/github/license/Data-Corruption/blog)

Blog is a simple async logger with file rotation and console logging.

This project is born from a personal need for a performant logger and a desire to deepen my understanding of Go. While it's primarily a personal endeavor, I welcome anyone who finds it useful for their own projects and am open to evolving it based on user feedback and needs.

## Features

- **Standard Library Only**
- **Thread-Safe**
- **Customizable At Runtime**
- **Log File Rotation**

## Getting Started

```sh
go get github.com/Data-Corruption/blog/v2
```

Basic Example:

```go
import "github.com/Data-Corruption/blog/v2/pkg"

func main() {
  // Init blog.
  //
  // Parameters:
  //   - DirPath: Path for log files. "." for current working directory or "" to disable file logging.
  //   - Level: Desired logging level for filtering messages.
  //   - IncludeLocation: When true, adds source file and line number to log messages (e.g., "main.go:42").
  //   - EnableConsole: When true, enables logging to the console in addition to files.
  //
  if err := blog.Init("logs", blog.INFO, false, true); err != nil {
    log.Printf("Error initializing logger: %v", err)
  }

  // Log messages from anywhere in the program
  blog.Info("This is an info message.")

  // Log messages with formatting
  blog.Warnf("This is an warn message with a format string: %v", err)

  // Synchronously cleanup the logger with a timeout; 0 means block indefinitely.
  // This should be called at the end of the program.
  blog.Cleanup(0)

  // for all other functions see `blog.go`. For access to the raw logger, see `logger.go`.
}
```

## FAQ

In this section, you'll find answers to common questions and troubleshooting tips for using our project effectively.

<details>
<summary><b>How to Handle Log File Rotation</b></summary>

Question: What happens when the log file reaches its maximum size, and how can I manage it?

Answer: Blog automatically handles log file rotation based on the size limit you set. Once the latest.log file exceeds the specified maximum size, it's renamed with the current date and time, and a new latest.log file is created. You can adjust the maximum file size using `blog.SetMaxFileSizeBytes(size)`. This ensures your logs are manageable and prevents excessive file growth.
</details>

<details>
<summary><b>Changing Settings Dynamically</b></summary>

Question: Can I change the logger's settings at runtime, and how?

Answer: Yes, you can dynamically adjust various settings in the logger. Due to the async nature of the logger these settings may take a few ms to update. Here is a list of available methods to update settings:

- `SetLevel(level LogLevel)`
- `SetConsole(enable bool)`
- `SetMaxBufferSizeBytes(size int)` Larger values will increase memory usage and reduce the frequency of disk writes.
- `SetMaxFileSizeBytes(size int)`
- `SetDirectoryPath(path string)` "." for current directory and "" to disable file logging.
- `SetFlushInterval(d time.Duration)` To disable automatic flushing, set to 0

</details>

<details>
<summary><b>Configuring Buffer and Flush Settings</b></summary>

**Question**: How can I optimize performance by configuring the internal buffer and flush intervals?

**Answer**: Blog optimizes log writing using a rolling buffer, which automatically flushes based on two configurable events:

- **Buffer Size Limit Reached**: When the buffer accumulates to a certain size, it triggers a flush. You can set this threshold with `blog.SetMaxBufferSizeBytes(size)`. The default size is 4KB. Larger values will reduce the frequency of disk writes but also increase memory usage.
- **Time Interval Elapsed**: The buffer also flushes periodically after a specified time interval, ensuring logs are written even during low activity. Set this interval with `blog.SetFlushInterval(amountOfTime)`. The default interval is 5 seconds. Shortening this time ensures more frequent writes, while lengthening it can reduce disk I/O. To disable entirely set this to 0.

</details>

<details>
<summary><b>Using Blog in Concurrent Environments</b></summary>

**Question**: Is Blog suitable for concurrent environments, and are there any special considerations for synchronous operations?

**Answer**: Blog is inherently safe for concurrent use in applications. Keep in mind it is asynchronous. Tf you require synchronous logging, I recommend checking out one of GO's many libs that support sync operation, like [zap](https://github.com/uber-go/zap).
</details>

<details>
<summary><b>Logging Issue: Message not appearing in the log file</b></summary>
  
**Question**: After logging a message, flushing, and then reading the log file, why doesn't it contain my message?

**Answer**: This is likely due to the asynchronous nature of our logging system. These processes may require some time to execute. To resolve this:

- **Step 1**: Wait for a few milliseconds after logging your message before flushing or cleanup.
- **Step 2**: Similarly, wait for a few milliseconds after flushing before you attempt to read the log file.

</details>

## Contributing

Contributions, issues, and feature requests are welcome! Feel free to check [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to contribute.

## License

Distributed under the MIT License. See [LICENSE.md](LICENSE.md) for more information.
