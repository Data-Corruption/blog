# Changelog for "blog" project

## [v3.0.1] - 2025-02-09

### Fixed

- **Level:** Exported the LogLevel as Level to allow for easier access.

## [v3.0.0] - 2025-02-09

### Changed

- **Refactor:** Rewrote a few functions and reorganized the codebase into separate packages for better organization and maintainability.
- **Simplified Import:** It's now `"github.com/Data-Corruption/blog/v3"`.

## [v2.0.2] - 2024-07-21

### Added

- **Fatal Fix:** Fatal and Fatalf will now print the msg to console in the event they fail the timeout.

## [v2.0.0] - 2024-07-21

### Added

- **Cleanup():** You should now call this instead of SyncFlush at the end of your program.

### Changed

- **Init():** Now takes an additional two boolean arguments: IncludeLocation and EnableConsole.

### Deprecated

- **SetUseConsole():** Renamed to SetConsole.
- **SetDirPath():** Renamed to SetDirectoryPath.
- **SetMaxWriteBufSize():** Renamed to SetMaxBufferSizeBytes.
- **SetMaxFileSize():** Renamed to SetMaxFileSizeBytes.

## [v1.0.1] - 2024-07-21

### Changed

- **Fatal Logging:** Fatal() and Fatalf() now ignore use console setting and always print.

## [v1.0.0] - 2024-06-16

### Changed

- **Fatal Logging:** Fatal() and Fatalf() are now blocking / sync.

## [v0.3.0] - 2024-01-13

### Added

- **Formatting Options:** Added Errorf, Warnf, etc.

## [v0.2.2-beta] - 2024-01-13

### Added

- **ErrInvalidLogLevel:** LogLevelFromString() now returns this error as its second return value instead of an ok bool

### Deprecated

- **AlreadyInitializedError:** changed to ErrAlreadyInitialized
- **InvalidPathError:** changed to ErrInvalidPath

## [v0.2.1-beta] - 2024-01-13

### Added

- **dualOutTest:** added test for dual output (console and file)

### Changed

- **shouldLog:** moved should log calculation into a dedicated function to allow easier testing of it

## [v0.2.0-beta] - 2024-01-13

### Added

- **SyncFlush:** added a SyncFlush method that only returns after the flush has been completed. This is particularly useful for cleanup situations.

## [v0.1.3-beta] - 2023-12-24

### Fixed

- **Typo:** fixed a typo in the readme
- **Formatting:** messages now use padding and look a little better imo

## [v0.1.2-beta] - 2023-12-24

### Added

- **Tests:** Added more tests

### Changed

- **README:** Improved the readme

### Fixed

- **handleFlushError:** Fixed an issue with the remaining buffer not being properly logged to console.

## [v0.1.1-beta] - 2023-12-21

### Added

- **Tests:** Added tests for initialization, console out, file out, and automatic flushing.

### Changed

- **Channels:** Moved the channels out of the logger struct and into the var group for easier passing of logger in testing.

### Fixed

- **SetFlushInterval:** Fixed the restart logic in the main loop used to apply changes to the flush interval.

## [v0.1.0-beta] - 2023-12-21

### Added

- **Source Code:** Added the initial source code. This includes the primary functionality of the logging system with features like multiple log levels (ERROR, WARN, INFO, DEBUG, FATAL), console and file output, and automatic buffer flushing.
- **Documentation:**
  - `README.md`: Created a comprehensive README file explaining the purpose of the project, how to install and use it.
  - `LICENSE.md`: Added a LICENSE file to clearly state the terms under which this project can be used.
  - `CONTRIBUTIONS.md`: Introduced a CONTRIBUTIONS file to guide potential contributors on how to effectively participate in the development of this project.
