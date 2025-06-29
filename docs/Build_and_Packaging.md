# Build & Packaging
## Table of Contents
- [Requirements](#requirements)
- [Makefile Targets](#makefile-targets)
  - [`build`](#makefile-targets)
  - [`release`](#makefile-targets)
  - [`clean`](#makefile-targets)
- [Packaging Steps](#packaging-steps)
  - [Compile Binary](#packaging-steps)
  - [Assemble Bundle](#packaging-steps)
  - [Zip & Rename](#packaging-steps)
  - [Install](#packaging-steps)
- [Versioning & Code-Signing](#versioning--code-signing-todo)

## Requirements

* Go 1.22
* Windows SDK (for CGO headers)
* `streamdeck-plugin` utility (optional helper for bundling)

## Makefile Targets

| Target | Description |
| --- | --- |
| `make build` | Build the plugin binary for local OS |
| `make release` | Cross-compile, assemble `.streamDeckPlugin`, and zip artifacts |
| `make clean` | Remove build output |

## Packaging Steps

1. `make build` to compile `hwinfo_streamdeck_plugin.exe`.
2. Copy binary and front-end assets into `com.exension.hwinfo.sdPlugin` bundle.
3. Zip the folder and rename to `.streamDeckPlugin`.
4. Double-click the file or run `install-plugin.bat` to install.

## Cross-Compilation

The plugin must be built for **64-bit Windows** because both Stream Deck and HWiNFO are Windows-only. From macOS/Linux you can run:

```bash
# Produce static Windows binaries
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o com.exension.hwinfo.sdPlugin/hwinfo.exe ./cmd/hwinfo_streamdeck_plugin
```

> No CGO is used, so cross-compiling does not require a Windows tool-chain.

## Detailed Target Notes

| Target | Steps Performed |
| --- | --- |
| `make build` | 1) Compile `hwinfo.exe` (plugin) and `hwinfo-plugin.exe` (gRPC helper) into the `.sdPlugin` folder.<br>2) Optionally copy pre-built helper from sibling repo.<br>3) Runs `install-plugin.bat` which calls Elgato's CLI to install/update the plugin in the local Stream Deck profile. |
| `make proto` | Regenerate Go gRPC stubs from `*.proto` files using `protoc` cached inside `.cache/protoc/`. |
| `make debug` | Builds an alternative binary (`cmd/hwinfo_debugger`) and installs it into the bundle for local console debugging. |
| `make release` | 1) Cleans previous build artifact.<br>2) Calls `DistributionTool.exe` to create the final `com.exension.hwinfo.streamDeckPlugin` file inside `build/`. |

Windows users can also run the helper batch scripts directly:

* `install-plugin.bat` – installs/updates the plugin for the current user.
* `kill-streamdeck.bat` / `start-streamdeck.bat` – restart the Stream Deck process when needed.
* `make-release.bat` – thin wrapper around Elgato's `DistributionTool.exe` for CI environments that lack `make`.

## CI Example (GitHub Actions)

```yaml
name: build-release
on:
  push:
    tags: [ 'v*' ]
jobs:
  win:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Cache protoc
        uses: actions/cache@v4
        with:
          path: .cache
          key: protoc-3
      - name: Build plugin
        run: make release
      - name: Upload Artifact
        uses: actions/upload-artifact@v4
        with:
          name: hwinfo-streamdeck
          path: build/com.exension.hwinfo.streamDeckPlugin
```

## Versioning & Code-Signing

### Version stamping

The binary exposes a `Version` string compiled via `-ldflags`:

```bash
go build -ldflags "-X main.Version=$(git describe --tags --always)" -o hwinfo.exe ./cmd/hwinfo_streamdeck_plugin
```

The `manifest.json` inside the `.sdPlugin` bundle must also be updated with the same semantic version (`"Version": "1.4.0"`).

### Windows code-signing (optional)

If you distribute the plugin publicly, sign the executables to eliminate SmartScreen warnings:

```bash
signtool sign /fd SHA256 /a /tr http://timestamp.digicert.com /td SHA256 hwinfo.exe
signtool sign /fd SHA256 /a /tr http://timestamp.digicert.com /td SHA256 hwinfo-plugin.exe
```

After signing, re-run `DistributionTool.exe` so the `.streamDeckPlugin` contains the signed binaries.

---

*Document complete — review whenever build tooling changes.*
