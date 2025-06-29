# System Patterns

## Project Organization

### Development Scripts Pattern
All development and verification scripts are stored in the `/dev` directory:
- `check-prerequisites.ps1` - Environment setup validation
- `verify-hwinfo.ps1` - HWiNFO configuration checks
- Additional verification and development tools

Key patterns:
- Keep all development scripts in `/dev`
- Use PowerShell for Windows-specific tooling
- Consistent naming (verify-*, check-*)
- Clear exit codes for automation
- Detailed error messages with fix steps

### PowerShell Compatibility Pattern
```powershell
# Prefer native cmdlets over Win32 API
# Instead of:
Add-Type @"
    using System;
    public class Win32 {
        [DllImport("user32.dll")]
        public static extern IntPtr FindWindow(string className, string windowName);
    }
"@
$window = [Win32]::FindWindow($null, "Window Title")

# Use:
$windows = Get-Process | Where-Object {
    $_.MainWindowTitle -match "Window Title" -and
    $_.MainWindowHandle -ne 0
}

# Get process path (works across PowerShell versions)
$processPath = $null
try {
    $processPath = (Get-Process "ProcessName" | Select-Object -First 1).MainModule.FileName
} catch {
    Write-Host "Could not determine process path"
}

# Use path if available
if ($processPath) {
    # Process path-dependent operations
} else {
    # Fallback behavior
}
```

Key patterns:
- Prefer PowerShell cmdlets over Win32 API
- Use MainWindowTitle and MainWindowHandle for window detection
- Filter out processes without windows (MainWindowHandle -ne 0)
- Use MainModule.FileName instead of deprecated Path property
- Handle access denied errors gracefully
- Provide fallback behavior when information unavailable
- Support both Windows PowerShell and PowerShell Core
- Maintain backward compatibility

### Windows API Interop Pattern
```powershell
# Define Windows API functions
Add-Type @"
    using System;
    using System.Runtime.InteropServices;
    public class Win32 {
        [DllImport("user32.dll")]
        public static extern int EnumWindows(EnumWindowsProc lpEnumFunc, IntPtr lParam);

        public delegate bool EnumWindowsProc(IntPtr hwnd, IntPtr lParam);
    }
"@

# Create callback and delegate
$callback = {
    param([IntPtr]$hwnd, [IntPtr]$lparam)
    # Process window
    return $true
}

$delegate = [System.Runtime.InteropServices.Marshal]::GetDelegateForFunctionPointer(
    [System.Runtime.InteropServices.Marshal]::GetFunctionPointerForDelegate($callback),
    [Type]$delegateType
)
```

Key patterns:
- Use Add-Type for P/Invoke definitions
- Handle callbacks via script blocks
- Convert script blocks to delegates
- Add debug output for diagnostics
- Handle API errors gracefully
- Support both 32-bit and 64-bit processes

## Build System

### Prerequisites Check Pattern
- PowerShell script (`check-prerequisites.ps1`) validates all dependencies
- Fail-fast approach with clear error messages
- Automated environment setup where possible
- Visual feedback with color-coded status indicators

### HWiNFO Verification Pattern
```powershell
# Verify HWiNFO configuration
$sharedMemMutex = [System.Threading.Mutex]::OpenExisting("Global\HWiNFO_SM2_MUTEX")
try {
    if ($sharedMemMutex.WaitOne(100)) {
        try {
            # Shared memory is accessible
        }
        finally {
            $sharedMemMutex.ReleaseMutex()
        }
    }
}
catch {
    # Shared memory not available
}
```

Key patterns:
- Registry-based installation detection
- Process path validation for portable check
- Mutex-based shared memory verification
- Window title detection for mode check
- Color-coded status output
- Detailed error messages with fix steps
- Exit codes for automation

### Process Management Pattern
```makefile
# Process termination target
kill-processes:
    -@taskkill /F /IM StreamDeck.exe /T >nul 2>&1
    -@taskkill /F /IM hwinfo.exe /T >nul 2>&1
    -@taskkill /F /IM hwinfo-plugin.exe /T >nul 2>&1
    -@timeout /t 2 /nobreak >nul

# Build targets depend on process termination
plugin: kill-processes
debug: kill-processes
clean: kill-processes
```

Key patterns:
- Terminate processes before file operations
- Silent error handling for non-existent processes
- Wait period to ensure cleanup
- Dependency-based process management

### HWiNFO Integration Pattern
```go
// Shared memory access pattern with retries
func ReadBytes() ([]byte, error) {
    maxRetries := 5
    retryDelay := 1 * time.Second

    var lastErr error
    for i := 0; i < maxRetries; i++ {
        err := mutex.Lock()
        if err != nil {
            lastErr = fmt.Errorf("failed to acquire mutex: %w", err)
            time.Sleep(retryDelay)
            continue
        }

        hnd, err := openFileMapping()
        if err != nil {
            mutex.Unlock()
            lastErr = fmt.Errorf("failed to open file mapping: %w", err)
            time.Sleep(retryDelay)
            continue
        }

        // Validate shared memory signature
        if !isValidSignature(hnd) {
            mutex.Unlock()
            lastErr = fmt.Errorf("invalid shared memory signature")
            time.Sleep(retryDelay)
            continue
        }

        data := copyBytes(addr)
        unmapViewOfFile(addr)
        windows.CloseHandle(windows.Handle(unsafe.Pointer(hnd)))
        mutex.Unlock()
        return data, nil
    }

    return nil, fmt.Errorf("failed to read shared memory after %d retries: %v", maxRetries, lastErr)
}

// Plugin error handling pattern
type Plugin struct {
    c      *plugin.Client
    cmd    *exec.Cmd
    hw     hwsensorsservice.HardwareService
    sd     *streamdeck.StreamDeck
    // ... other fields ...
}

func (p *Plugin) startClient() error {
    cmd := exec.Command("./hwinfo-plugin.exe")
    p.cmd = cmd

    client := plugin.NewClient(&plugin.ClientConfig{
        HandshakeConfig:  hwsensorsservice.Handshake,
        Plugins:          hwsensorsservice.PluginMap,
        Cmd:             cmd,
        AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
        AutoMTLS:        true,
    })

    // Connect via RPC with detailed error handling
    rpcClient, err := client.Client()
    if err != nil {
        return fmt.Errorf("failed to connect to plugin: %w", err)
    }

    // Request the plugin with error context
    raw, err := rpcClient.Dispense("hwinfoplugin")
    if err != nil {
        return fmt.Errorf("failed to dispense plugin: %w", err)
    }

    p.c = client
    p.hw = raw.(hwsensorsservice.HardwareService)

    return nil
}
```

Key patterns:
- Retry mechanism for shared memory access
- Proper mutex handling with deferred unlocks
- Signature validation before data access
- Detailed error messages with context
- Clean resource management
- Process monitoring and auto-restart
- Debug logging for troubleshooting

### Shared Memory Validation Pattern
```go
// Constants for shared memory validation
const (
    HWiNFO_SIGNATURE = 0x53695748 // "HWiS" in little-endian
    HWiNFO_DEAD_SIGNATURE = 0x44414544 // "DEAD" in little-endian
)

// Validation functions
func isValidSignature(addr uintptr) bool {
    var d []byte
    dh := (*reflect.SliceHeader)(unsafe.Pointer(&d))
    dh.Data = addr
    dh.Len, dh.Cap = C.sizeof_HWiNFO_SENSORS_SHARED_MEM2, C.sizeof_HWiNFO_SENSORS_SHARED_MEM2
    cheader := C.PHWiNFO_SENSORS_SHARED_MEM2(unsafe.Pointer(&d[0]))

    switch cheader.dwSignature {
    case HWiNFO_SIGNATURE:
        return true
    case HWiNFO_DEAD_SIGNATURE:
        return false // HWiNFO is shutting down
    default:
        return false // Invalid or uninitialized shared memory
    }
}
```

Key patterns:
- Explicit signature constants
- Proper byte order handling
- State detection (active vs shutdown)
- Safe memory access
- Clear error conditions
- Resource cleanup

### Plugin Process Management Pattern
```go
// Plugin process management without external dependencies
type Plugin struct {
    c      *plugin.Client
    cmd    *exec.Cmd
    // ... other fields ...
}

// Auto-restart monitoring
go func() {
    for {
        if p.c != nil && p.c.Exited() {
            log.Println("Plugin process exited, attempting to restart...")
            if err := p.startClient(); err != nil {
                log.Printf("Failed to restart plugin: %v\n", err)
            }
        }
        time.Sleep(1 * time.Second)
    }
}()
```

Key patterns:
- Simple process tracking
- Automatic restart on failure
- Detailed error logging
- Resource cleanup on exit
- Graceful shutdown handling

### Makefile Structure
```makefile
# Variables at top
GOCMD=go
GOBUILD=$(GOCMD) build
SDPLUGINDIR=./com.exension.hwinfo.sdPlugin

# PHONY declarations for virtual targets
.PHONY: build clean proto plugin debug release kill-processes

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
- Process management integration

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
- Process isolation and management

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
- Process management for build-time coordination

### HWiNFO Data Access Pattern
```go
// Sensor data access pattern
type Sensor struct {
    ID string      // Unique sensor identifier
    Name string    // User-friendly name
    Readings []Reading // Associated readings
}

// Reading data access pattern
type Reading struct {
    ID int32       // Reading identifier
    Type string    // Reading type (temp, voltage, etc.)
    Label string   // User-friendly label
    Value float64  // Current value
    Unit string    // Measurement unit
}

// Error handling pattern
type Result struct {
    Shmem *SharedMemory
    Err error
}
```

Key patterns:
- Strongly typed sensor/reading data
- Error propagation through channels
- Automatic reconnection on failure
- Configuration validation
- User feedback for issues

## Error Handling

### Build-time Errors
- Prerequisites validation before build
- Clear error messages with fix instructions
- Proper exit codes for script integration
- Cleanup on failure
- Silent handling of expected process termination errors

### Runtime Errors
- Graceful degradation when HWiNFO is unavailable
- Automatic reconnection for lost connections
- User feedback through Stream Deck UI
- Logging for diagnostics
- Process cleanup on abnormal termination
- HWiNFO configuration validation

## Testing Patterns

### Test Categories
- Unit tests alongside code
- Integration tests in examples/
- Performance benchmarks
- Manual UI testing
- Process management testing
- HWiNFO integration testing

### Test Tools
- Go testing package
- Example programs
- Debug builds
- Sensor simulators
- Process monitoring tools
- HWiNFO configuration validator

## Deployment Patterns

### Plugin Distribution
- `.streamDeckPlugin` bundle packaging
- Automatic installation script
- Version stamping from git
- Optional code signing
- Process management during installation
- HWiNFO configuration check

### Development Workflow
1. Check prerequisites (`check-prerequisites.ps1`)
   - Validate HWiNFO configuration
   - Check for installed vs portable version
2. Build plugin (`make build`)
   - Automatic process termination
   - File operation safety
3. Test locally
   - Verify HWiNFO integration
   - Check sensor data access
4. Package for distribution (`make release`)

## Error Handling

### User Feedback Pattern
```powershell
# Color-coded status messages
Write-Status "Component check" $success
if (-not $success) {
    Write-Host "Error: Component failed"
    Write-Host "Fix: Follow these steps..."
    exit $errorCode
}
```

Key patterns:
- Color-coded status indicators
- Consistent message formatting
- Clear error messages
- Step-by-step fix instructions
- Meaningful exit codes

## Integration Patterns

### HWiNFO Integration
- Shared memory access via mutex
- Process validation before operations
- Configuration verification
- Mode detection via window title
- Installation path validation

### Stream Deck Integration
- Process lifecycle management
- Plugin installation verification
- Action registration
- UI feedback for configuration issues

## Development Workflow

### Verification First Pattern
1. Check installation and configuration
2. Validate running processes
3. Verify feature availability
4. Provide fix instructions if needed
5. Exit with meaningful codes

### Documentation Pattern
- Clear prerequisites
- Step-by-step setup instructions
- Troubleshooting guides
- Exit code documentation
- Configuration requirements

---

*Document updated to reflect build system patterns, process management, HWiNFO integration patterns, and improvements.*
