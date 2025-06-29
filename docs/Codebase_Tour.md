# Codebase Tour
## Table of Contents
- [Overview & Directory Map](#overview--directory-map)
- [`cmd/` – Executables](#cmd--executables)
- [`internal/` – Private Packages](#internal--private-packages)
  - [`hwinfo/`](#hwinfo)
  - [`app/hwinfostreamdeckplugin/`](#apphwinfostreamdeckplugin)
- [`pkg/` – Public Packages](#pkg--public-packages)
  - [`streamdeck/`](#streamdeck)
  - [`service/`](#service)
  - [`graph/`](#graph)
- [Front-End Assets](#front-end-assets-comextensionhwinfosdplugin)
- [Examples & Benchmarks](#examples--benchmarks-examples)
- [Future Expansion Notes](#future-expansion-notes-todo)

Quick reference to every directory & key file.

| Path | Purpose |
| --- | --- |
| `cmd/hwinfo_streamdeck_plugin` | Main plugin executable |
| `cmd/hwinfo_debugger` | CLI for streaming sensor values to stdout |
| `cmd/hwinfo-plugin` | gRPC side-car exposing `pkg/service` implementation |
| `internal/hwinfo` | Shared-memory parsing, sensor abstraction |
| `internal/app/hwinfostreamdeckplugin` | App layer orchestrating sensors ↔ Stream Deck |
| `pkg/streamdeck` | WebSocket SDK wrapper & message structs |
| `pkg/service` | gRPC proto + server/client helpers |
| `com.exension.hwinfo.sdPlugin` | Plugin bundle assets & manifest |
| `examples/` | Sample programs & benchmarks |

## Directory Details

### `cmd/`

| Binary | Purpose | Typical Invocation |
| --- | --- | --- |
| `hwinfo_streamdeck_plugin` | Main plugin executable launched by Stream Deck. Establishes WebSocket connection, polls sensors, updates button titles/images. | _Normally auto-started by Stream Deck._ For manual run:<br>`hwinfo_streamdeck_plugin --port 12345 --pluginUUID com.exention.hwinfo` |
| `hwinfo_debugger` | Developer-only CLI that prints all sensors/readings to stdout for rapid testing without Stream Deck. | `hwinfo_debugger --interval 2s --filter "CPU Package"` |
| `hwinfo-plugin` | HashiCorp go-plugin side-car that exposes the HWService gRPC interface so external programs can query sensors. | `hwinfo-plugin -socket-path C:\tmp\hw.sock -log-level debug` |

Flags are self-documenting via `--help`. All binaries target Windows x64 but can run under WSL/MinGW for development.

### `internal/`

> Private packages that encapsulate domain logic and should not be imported by external code.

#### `hwinfo/` — Hardware-data domain layer

| Item | Kind | Role |
| --- | --- | --- |
| `SharedMemory` | struct | Snapshot wrapper around raw `HWiNFO_SENSORS_SHARED_MEM2` bytes with helper accessors (signature, version, iterators) |
| `ReadSharedMem()` | func | Copies the shared-memory region into Go memory and returns a `SharedMemory` instance |
| `StreamSharedMem()` | func (goroutine) | Emits `hwinfo.Result` (snapshot or error) every second via a channel |
| `Sensor` / `Reading` | structs | Lightweight views over sensor and reading elements inside the snapshot |
| Sub-pkg `mutex` | package | Acquires/releases the named Win32 mutex used by HWiNFO to protect shared memory |
| Sub-pkg `shmem` | package | Opens the `Global\HWiNFO_SENS_SM2` file mapping and returns a byte slice |
| Sub-pkg `plugin` | package | Provides a higher-level `Service` that caches sensors/readings and exposes helper methods (`SensorIDByIdx`, `ReadingsBySensorID`) |

#### `app/hwinfostreamdeckplugin/` — Application layer

| Item | Kind | Role |
| --- | --- | --- |
| `Plugin` | struct | Main orchestrator; bridges HWiNFO service, Stream Deck SDK, and action manager |
| `NewPlugin()` | func | Bootstraps helper gRPC side-car, creates `streamdeck.StreamDeck` instance |
| `RunForever()` | method | Enters event loop; handles reconnection, shutdown, and delegates |
| `actionManager` | struct | Thread-safe registry of active Stream Deck contexts; ticks every second to update tiles |
| `handlers.go` | file | Implements Stream Deck delegate callbacks (`OnKeyDown`, `OnWillAppear`, etc.) |
| `delegate.go` | file | Glue between Stream Deck events and `actionManager` settings |

These two internal packages are where most business logic lives; changes here will ripple through the rest of the system.

### `pkg/`

Public, reusable packages that could be consumed by external tools or other Go modules.

#### `streamdeck/` — WebSocket SDK Wrapper

| Item | Kind | Role |
| --- | --- | --- |
| `StreamDeck` | struct | Manages WebSocket connection, event registration, and outgoing commands (setTitle, setImage, etc.) |
| `EventDelegate` | interface | Callback contract implemented by the application layer to receive SDK events |
| `Connect()/Close()` | methods | Establish and tear down WebSocket connection |
| `ListenAndWait()` | method | Starts message reader goroutine and blocks on signals |
| `SendToPropertyInspector()` | method | Convenience wrapper for `sendToPropertyInspector` payload |

#### `service/` — gRPC Service Definition & Helpers

| Item | Kind | Role |
| --- | --- | --- |
| `proto/` | directory | Contains `hwservice.proto` and generated `*_pb.go` stubs |
| `HardwareService` | interface | Abstract contract with `PollTime()`, `Sensors()`, and `ReadingsForSensorID()` |
| `HardwareServicePlugin` | struct | Hashicorp go-plugin implementation enabling the helper process |
| `Handshake` / `PluginMap` | vars | Shared constants between host and plugin |

#### `graph/` — Lightweight Image Graphing

| Item | Kind | Role |
| --- | --- | --- |
| `Graph` | struct | Renders rolling histogram of sensor values to an in-memory PNG (256×256 by default) |
| `Update()` | method | Push new datapoint; internal slice slides and re-renders lazily |
| `EncodePNG()` | method | Returns compressed PNG; used by plugin to set button image |
| `FontFaceManager` | helper | Caches truetype faces to avoid re-parsing font file every draw |

These public packages are versioned via Go modules; external consumers should import them using `github.com/shayne/hwinfo-streamdeck/pkg/...`.

### Front-End Assets (`com.exension.hwinfo.sdPlugin`)

All files in this bundle are packaged into the final `.streamDeckPlugin` zip and loaded by the Stream Deck runtime.

| Path | Type | Purpose |
| --- | --- | --- |
| `manifest.json` | JSON | Declares plugin metadata, executable paths, property-inspector HTML, action UUIDs, icons, and OS requirements. Must stay in sync with binary version and icon filenames. |
| `index_pi.html` | HTML | Property Inspector UI; references `sdpi.css`, `local.css`, and `index_pi.js`. |
| `index_pi.js` | JavaScript | Handles WebSocket to Stream Deck for PI, populates sensor/reading dropdowns, persists settings, DOM manipulation. |
| `css/sdpi.css` | CSS | Stock style sheet shipped by Elgato; provides SDPI layout primitives. |
| `css/xsdpi.css` | CSS | Retina/high-DPI overrides for SDPI widgets. |
| `css/local.css` | CSS | Project-specific tweaks (colors, spacing). |
| `css/reset.min.css` | CSS | Minimal CSS reset for consistent styling across browsers. |
| `css/*.svg / *.png` | Images | Small UI glyphs used by SDPI components (caret, checkmarks, calendar icons). |
| `defaultImage.png` | PNG | Default button image (1×). |
| `defaultImage@2x.png` | PNG | Retina version of default button image. |
| `icon.png` / `icon@2x.png` | PNG | Plugin list icon displayed in Stream Deck store/manager. |
| `pluginIcon.png` | PNG | Square icon for action selection list. |
| `launch-hwinfo.png` | PNG | Graphic shown on button when HWiNFO isn't running. |
| `DejaVuSans-Bold.ttf` | TTF | Embedded font used by `pkg/graph` for PNG rendering on buttons. |

_Development tip:_ open `index_pi.html` in a desktop browser and use the built-in mock WebSocket shim inside `index_pi.js` to iterate on UI without launching Stream Deck.

## Overview & Directory Map

```text
hwinfo-streamdeck/
├─ cmd/
│  ├─ hwinfo_streamdeck_plugin/    # Main plugin executable (WS client)
│  ├─ hwinfo_debugger/             # Console sensor viewer for devs
│  └─ hwinfo-plugin/               # gRPC side-car exposing HWService
├─ internal/
│  ├─ hwinfo/                      # Shared-memory parsing & sensor model
│  │  ├─ mutex/                    # Win32 mutex helpers
│  │  ├─ shmem/                    # Memory-mapped file helpers
│  │  └─ util/                     # Small C/Go interop helpers
│  └─ app/
│     └─ hwinfostreamdeckplugin/   # Application layer (action manager)
├─ pkg/
│  ├─ streamdeck/                  # WebSocket SDK wrapper & types
│  ├─ service/
│  │  ├─ proto/                    # gRPC proto definitions
│  │  └─ ...                       # Generated stubs & helpers
│  └─ graph/                       # Experimental sensor graph render
├─ com.exension.hwinfo.sdPlugin/   # Property Inspector assets, manifest
├─ examples/                       # Sample programs & benchmarks
├─ docs/                           # Project documentation (this folder)
├─ build/                          # Release artifacts (git-ignored)
├─ go.mod / go.sum                 # Module metadata & deps
└─ Makefile                        # Build automation
```

Use the tree above to orient yourself quickly; the detailed tables that follow describe each directory in depth.

---

> **TODO**: Document individual structs, interfaces, and significant functions as the project stabilises.
