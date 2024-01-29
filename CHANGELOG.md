# Changelog for "blog" project

## [v0.2.1-beta] - 2024-01-13

### Added
- **dualOutTest:** added test for dual output (console and file)

### Changed
- **shouldLog:** moved should log calculation into a dedicated function to allow easier testing of it

### Fixed
- None in this release.

### Deprecated
- None in this release.

### Removed
- None in this release.

### Security
- None in this release.

## [v0.2.0-beta] - 2024-01-13

### Added
- **SyncFlush:** added a SyncFlush method that only returns after the flush has been completed. This is particularly useful for cleanup situations.

### Changed
- None in this release.

### Fixed
- None in this release.

### Deprecated
- None in this release.

### Removed
- None in this release.

### Security
- None in this release.

## [v0.1.3-beta] - 2023-12-24

### Added
- None in this release.

### Changed
- None in this release.

### Fixed
- **Typo:** fixed a typo in the readme
- **Formatting:** messages now use padding and look a little better imo

### Deprecated
- None in this release.

### Removed
- None in this release.

### Security
- None in this release.

## [v0.1.2-beta] - 2023-12-24

### Added
- **Tests:** Added more tests

### Changed
- **README:** Improved the readme

### Fixed
- **handleFlushError:** Fixed an issue with the remaining buffer not being properly logged to console.

### Deprecated
- None in this release.

### Removed
- None in this release.

### Security
- None in this release.

## [v0.1.1-beta] - 2023-12-21

### Added
- **Tests:** Added tests for initialization, console out, file out, and automatic flushing.

### Changed
- **Channels:** Moved the channels out of the logger struct and into the var group for easier passing of logger in testing.

### Fixed
- **SetFlushInterval:** Fixed the restart logic in the main loop used to apply changes to the flush interval.

### Deprecated
- None in this release.

### Removed
- None in this release.

### Security
- None in this release.

## [v0.1.0-beta] - 2023-12-21

### Added
- **Source Code:** Added the initial source code. This includes the primary functionality of the logging system with features like multiple log levels (ERROR, WARN, INFO, DEBUG, FATAL), console and file output, and automatic buffer flushing.
- **Documentation:**
  - `README.md`: Created a comprehensive README file explaining the purpose of the project, how to install and use it.
  - `LICENSE.md`: Added a LICENSE file to clearly state the terms under which this project can be used.
  - `CONTRIBUTIONS.md`: Introduced a CONTRIBUTIONS file to guide potential contributors on how to effectively participate in the development of this project.

### Changed
- None in this release.

### Fixed
- None in this release.

### Deprecated
- None in this release.

### Removed
- None in this release.

### Security
- None in this release.