# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.5] - 2026-01-02

### Fixed
- Fixed pkg.go.dev license detection by removing empty LICENSE.md file
- Package documentation now displays correctly on pkg.go.dev

### Added
- Comprehensive package-level documentation in goroutine.go
- Enhanced README.md with detailed API reference and usage examples for all components
- Documentation improvements for better discoverability

### Documentation
- Added complete API reference for Async Resolve, SuperSlice, SafeChannel, and GoManager
- Added usage examples in package overview
- Enhanced documentation following Go conventions

## [1.0.4] - 2025-01-05

### Added
- SuperSlice methods: filter, map, and reduce operations
- Enhanced slice processing capabilities

## [1.0.2] - 2024-10-09

### Added
- Async resolver functionality
- Promise-like pattern for concurrent operations
- Support for waiting with timeout
- Multiple async task management

## [1.0.1] - 2024-07-29

### Added
- Struct recasting utilities
- Type conversion helpers for collecting data from multiple sources
- Basic type conversion functionality
- Functional resolution support

### Documentation
- Added comprehensive recasting examples
- Type conversion documentation

## [0.0.1] - 2024-06-19

### Added
- Initial release
- GoManager for goroutine management
- SafeChannel for thread-safe channel operations
- Basic concurrent processing utilities
- Named goroutine management with cancellation support
- Context-based lifecycle management

[1.0.5]: https://github.com/ivikasavnish/goroutine/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/ivikasavnish/goroutine/compare/v1.0.2...v1.0.4
[1.0.2]: https://github.com/ivikasavnish/goroutine/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/ivikasavnish/goroutine/compare/v0.0.1...v1.0.1
[0.0.1]: https://github.com/ivikasavnish/goroutine/releases/tag/v0.0.1
