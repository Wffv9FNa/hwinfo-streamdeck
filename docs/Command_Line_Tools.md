# Command-Line Tools
## Table of Contents
- [Tool Summary](#tool-summary)
- [Binary Reference Table](#binary-reference-table)
  - `hwinfo_streamdeck_plugin`
  - `hwinfo_debugger`
  - `hwinfo-plugin`
- [Common Flags & Environment Variables](#common-flags--environment-variables)
- [Logging & Verbosity Options](#logging--verbosity-options-todo)

Utilities under `cmd/`.

| Binary | Description | Sample Usage |
| --- | --- | --- |
| `hwinfo_streamdeck_plugin` | Production plugin launched by Stream Deck (rarely run manually) | `hwinfo_streamdeck_plugin --port 12345` |
| `hwinfo_debugger` | Dump all sensors & readings to console for debugging | `hwinfo_debugger --interval 2s` |
| `hwinfo-plugin` | gRPC side-car exposing HWService | `hwinfo-plugin -socket-path ./sock` |

All binaries support `--help` for detailed flags.

---

### Tool Summary

The repository ships three primary executables—each found under `cmd/` and copied into the `.sdPlugin` bundle (or installed separately during debugging):

* **`hwinfo_streamdeck_plugin`** – The core plugin binary. Normally launched by the Stream Deck runtime with CLI parameters (`--port`, `--pluginUUID`, etc.). It:
  * Opens a WebSocket to Stream Deck to receive key events.
  * Starts a helper gRPC side-car when required.
  * Polls HWiNFO shared memory and updates button titles/images every second.

* **`hwinfo_debugger`** – A console utility for development. It connects directly to HWiNFO shared memory (no Stream Deck) and dumps sensor readings to stdout in real-time. Useful for:
  * Verifying that shared memory integration works on a new machine.
  * Quickly identifying the `SensorUID` and `ReadingID` to pre-populate settings.

* **`hwinfo-plugin`** – A small side-car implementing `HardwareService` over gRPC using `hashicorp/go-plugin`. It is spawned by the main plugin binary and can also be run stand-alone to expose hardware data to external dashboards.

All three binaries are Windows-only (GOOS=windows), but can be cross-compiled from other hosts.

---

## Binary Reference Table

| Binary | Description | Key Flags / Params |
| --- | --- | --- |
| `hwinfo_streamdeck_plugin` | Core plugin binary run by Stream Deck. | `--port` (WS port), `--pluginUUID` (unique ID), `--registerEvent` (`registerPlugin`), `--info` (JSON with SD info). Flags are auto-supplied by Stream Deck. |
| `hwinfo_debugger` | Stand-alone console tool for sensor debugging. | Same flags as above when launched by Stream Deck, but when run manually you can omit them – the program will just log the args it receives. |
| `hwinfo-plugin` | gRPC side-car; no user-visible CLI flags. | None – parameters are passed over gRPC handshake. |

---

#### `hwinfo_streamdeck_plugin`

Launched by Stream Deck with CLI args. Manual example:

```powershell
hwinfo_streamdeck_plugin --port 12345 --pluginUUID com.exension.hwinfo --registerEvent registerPlugin --info '{}'
```

Key runtime behaviours:

* Connects to WebSocket `ws://127.0.0.1:{port}`.
* Registers with `registerPlugin` event.
* Reads shared memory every second.
* Spawns `hwinfo-plugin.exe` if not already running.

#### `hwinfo_debugger`

Primarily for developers; simply logs the arguments and then exits. In practice you embed your own debug logic or run the `examples/` binaries for benchmarks.

```powershell
hwinfo_debugger
```

Outputs the CLI args it was given to `APPDATA\Elgato\StreamDeck\Plugins\...\hwinfo.log`.

#### `hwinfo-plugin`

No CLI flags; invoked by the host via Hashicorp go-plugin. It:

* Streams shared memory using `internal/hwinfo/plugin.Service`.
* Serves gRPC on an automatically allocated TCP port provided in handshake env vars.

Not intended for direct user execution.

> **TODO**: Document log levels and environment variables.
