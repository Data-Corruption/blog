# Changelog for "blog" project

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