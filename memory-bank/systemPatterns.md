# System Patterns

## Build System

### Prerequisites Check Pattern
- PowerShell script (`check-prerequisites.ps1`) validates all dependencies
- Fail-fast approach with clear error messages
- Automated environment setup where possible
- Visual feedback with color-coded status indicators

### Makefile Structure
```makefile
# Variables at top
GOCMD=go
GOBUILD=$(GOCMD) build
SDPLUGINDIR=./com.exension.hwinfo.sdPlugin

# PHONY declarations for virtual targets
.PHONY: build clean proto plugin debug release

# Standard targets
build: plugin        # Main build target
clean: ...          # Cleanup
release: build ...  # Distribution packaging

# Component-specific targets
plugin: proto       # Main plugin binary
proto: $(PROTOPB)   # Protocol buffers
debug: ...         # Debug builds
```

Key patterns:
- Standard target names (`build`, `clean`, `release`)
- Dependencies flow from high-level to low-level targets
- Automatic directory creation for build artifacts
- Fail-safe cleanup operations
- Self-contained without external dependencies

### Visual Studio Integration
- Automatic detection of VS2022 installations
- Environment setup via vcvars64.bat or VsDevCmd.bat
- Support for multiple VS editions (Enterprise, Professional, Community, BuildTools)
- PATH management for compiler tools

## Architecture Patterns

### Plugin Structure
- Main plugin executable (`cmd/hwinfo_streamdeck_plugin`)
- Helper plugin for gRPC service (`cmd/hwinfo-plugin`)
- Debug utility (`cmd/hwinfo_debugger`)

### Package Organization
```
internal/           # Private implementation
  hwinfo/          # Core HWiNFO integration
  app/             # Application logic
pkg/               # Public API
  streamdeck/      # SDK wrapper
  service/         # gRPC services
  graph/           # Visualization
```

### Communication Patterns
- WebSocket for Stream Deck communication
- Shared memory for HWiNFO sensor data
- gRPC for external service integration
- Windows mutex for resource coordination

## Error Handling

### Build-time Errors
- Prerequisites validation before build
- Clear error messages with fix instructions
- Proper exit codes for script integration
- Cleanup on failure

### Runtime Errors
- Graceful degradation when HWiNFO is unavailable
- Automatic reconnection for lost connections
- User feedback through Stream Deck UI
- Logging for diagnostics

## Testing Patterns

### Test Categories
- Unit tests alongside code
- Integration tests in examples/
- Performance benchmarks
- Manual UI testing

### Test Tools
- Go testing package
- Example programs
- Debug builds
- Sensor simulators

## Deployment Patterns

### Plugin Distribution
- `.streamDeckPlugin` bundle packaging
- Automatic installation script
- Version stamping from git
- Optional code signing

### Development Workflow
1. Check prerequisites (`check-prerequisites.ps1`)
2. Build plugin (`make build`)
3. Test locally
4. Package for distribution (`make release`)

---

*Document updated to reflect build system patterns and improvements.*
