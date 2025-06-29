# Technical Context

| Area | Details |
| --- | --- |
| Primary Language | Go 1.22+ (module mode) |
| Build Tooling | Makefile with tasks; batch scripts for Windows plugin packaging |
| Target Platform | Windows 10+ (Stream Deck & HWiNFO availability) |
| External Dependencies | HWiNFO shared memory interface & mutex, Elgato Stream Deck SDK (manifest) |
| Internal Packages | `internal/hwinfo` (sensor logic), `pkg/streamdeck` (SDK wrapper), `pkg/service` (gRPC) |
| Frontend | HTML/CSS/JS property inspector assets under `com.exension.hwinfo.sdPlugin` |
| Testing | Go `testing` pkg; example benches under `examples/` |
| Continuous Integration | Not yet defined |

## Technical Constraints

- Must build to 64-bit Windows binary (<5 MB).
- No CGO dependencies to keep builds simple.
- Avoid vendorizing Stream Deck SDK; treat as static assets.

## Build Prerequisites (Windows)

A PowerShell script `check-prerequisites.ps1` is provided to verify all build dependencies are properly installed and configured. Run it before attempting to build:

```powershell
.\check-prerequisites.ps1
```

Required Components:
- **Go 1.22+** – install from https://go.dev/dl/
- **Git** – required for cloning and version tagging
- **Visual Studio 2022 Build Tools** (or full VS) with:
  - MSVC v143 - VS 2022 C++ x64/x86 build tools
  - Windows 10/11 SDK
  - C++ CMake tools for Windows
- **Elgato Stream Deck software** ≥ v6.0
- **HWiNFO64** with "Shared Memory Support" enabled in settings

Optional Components:
- **PowerShell 7** – recommended for better build script support
- **protoc v3.x** – only needed when editing `.proto` files (auto-downloaded if missing)

## Build Environment Setup

1. Install required components listed above (use `check-prerequisites.ps1` to verify)
2. Configure HWiNFO:
   - Launch HWiNFO in "Sensors-only" mode
   - Open Settings and enable "Shared Memory Support"
   - Recommended: Configure to start with Windows
3. Ensure Stream Deck software is running
4. Clone repository and initialize:
   ```powershell
   git clone https://github.com/shayne/hwinfo-streamdeck.git
   cd hwinfo-streamdeck
   go mod download
   ```

## Development Tools

| Tool | Purpose | Installation |
| --- | --- | --- |
| `go` | Core build toolchain | Manual install required |
| `make` | Build automation | Included in Git for Windows |
| `protoc` | gRPC stub generation | Auto-downloaded to `.cache/` |
| `cl.exe` | MSVC compiler | VS2022 Build Tools required |
| `DistributionTool.exe` | Plugin packaging | Included with Stream Deck software |

## Environment Variables

No special environment variables are required for development. The Visual Studio environment is automatically configured by the build scripts when needed.

## IDE Support

The codebase is IDE-friendly and works well with:
- Visual Studio Code with Go extension
- GoLand
- Any LSP-compatible editor

## Debugging Tools

- Built-in debugger (`cmd/hwinfo_debugger`)
- Stream Deck logs (see Troubleshooting.md)
- HWiNFO sensor viewer
- Windows Event Viewer for system-level issues

---

*Document updated with build prerequisites script and detailed environment setup.*
