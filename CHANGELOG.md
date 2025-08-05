# Changelog

All notable changes to this project will be documented in this file.

## [2.1.0] - 2025-01-07

### Requirements
- Stream Deck Software 6.4 or later (updated from 4.1)

### Added
- **Boolean Formatting**: Use `%b` in format string to convert 0/1 values to YES/NO
- **Thousands Separator**: Use `%t` in format string to add thousands separators (e.g., "1,234" instead of "1234")
- **Dynamic Unit Conversion**: Use `%u` in format string for automatic unit conversion (e.g., "1.5 GB/s" instead of "1500 MB/s")
- **Value Text Stroke**: Add stroke/outline to value text for better readability with customizable stroke colour and size
- **Improved Degree Symbol Support**: Enhanced handling of degree symbol (°) with proper UTF-8 and ISO8859-1 encoding support

### Changed
- Updated Property Inspector UI layout to accommodate new stroke colour and size controls
- Improved text rendering with configurable stroke effects

### Technical Details
- Dynamic unit conversion supports data transfer rates (KB/s, MB/s, GB/s) and storage units
- Stroke size is configurable from 0 to 5 pixels with real-time preview
- All new formatting options are backward compatible with existing configurations

## [2.0.5] - Previous Release

Initial stable release with core HWiNFO sensor reading functionality.
