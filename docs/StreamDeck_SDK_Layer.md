# Stream Deck SDK Layer
## Table of Contents
- [Package Overview](#package-overview)
- [Lifecycle of a Plugin Connection](#lifecycle-of-a-plugin-connection)
- [Threading & Concurrency Model](#threading--concurrency-model)
- [Outgoing Command Helpers](#outgoing-command-helpers)
- [Key Source Files](#key-source-files)
- [Reconnection Logic](#reconnection-logic)

This layer abstracts the Elgato Stream Deck WebSocket API into idiomatic Go types.

## Package Overview

`pkg/streamdeck` contains:

* Message structs for all events (keyDown, willAppear, didReceiveSettings, etc.).
* WebSocket client that handles registration and event dispatch.
* Helper functions for sending setTitle, setImage, and showAlert commands.

## Outgoing Command Helpers

`pkg/streamdeck/streamdeck.go` provides thin wrappers for the most common WebSocket commands so callers do not need to craft JSON manually.

| Go Method | SDK Command | Typical Use-Case |
| --- | --- | --- |
| `SetTitle(context, title string)` | `setTitle` | Update the button's text label. Used every polling tick for numeric readings. |
| `SetImage(context string, png []byte)` | `setImage` | Replace button icon with dynamically generated PNG (temperature graphs). Accepts raw PNG bytes which are base64-encoded internally. |
| `ShowAlert(context string)` *(planned)* | `showAlert` | Flash the red triangle on the key (not yet exposed but easy to add). |
| `SendToPropertyInspector(action, context string, payload interface{})` | `sendToPropertyInspector` | Push sensor & reading lists to the PI WebSocket for dropdown population. Payload struct is JSON-marshalled on the fly. |
| `SetSettings(context string, payload interface{})` | `setSettings` | Persist per-action settings; payload must be an anonymous struct that matches PI expectations. |

All helpers:

* Base64-encode binary data when required (images).
* Automatically include the mandatory `context` field.
* Return `error` if `json.Marshal` fails or the WebSocket write encounters an issue.

Because the WebSocket write side uses a mutex, helpers are safe to call concurrently from multiple goroutines.

## Lifecycle

1. Plugin binary starts, reads the `port`, `uuid`, and `registerEvent` from the Stream Deck runtime.
2. Establishes a WebSocket connection to `ws://localhost:{port}`.
3. Sends `registerEvent` + `uuid` to subscribe.
4. Receives events → unmarshalled into Go structs → forwarded to application layer.

## Threading Model

All incoming messages funnel through a single goroutine; outgoing commands use a buffered channel to avoid blocking.

## Key Files

| File | Description |
| --- | --- |
| `pkg/streamdeck/streamdeck.go` | Public API to dial the WebSocket and send/receive messages |
| `pkg/streamdeck/types.go` | Event and command struct definitions |

## Key Source Files

| File | Approx. LOC | Highlights |
| --- | --- | --- |
| `streamdeck.go` | ~300 | Core WebSocket client: connection, registration, JSON decoding, event dispatch, command helpers. Spawns reader goroutine and handles SIGINT for graceful shutdown. |
| `types.go` | ~200 | Auto-generated/hand-written Go structs that mirror Stream Deck's JSON schema for events (`EvWillAppear`, `EvSendToPlugin`, etc.) and commands (`cmdSetTitle`, `cmdSetImage`). Includes MarshalJSON helpers for optional fields. |
| `const.go` *(planned)* | <50 | Central place for action/event string constants—keeps magic strings out of code (e.g., `registerPlugin`, `setTitle`). Not implemented yet; current code uses literals. |
| `examples/streamdeck_echo/main.go` *(future)* | – | Example program for library consumers demonstrating echoing key presses back to titles. |

Reading these files in order gives newcomers a clear picture: schema definitions → WebSocket orchestration → helper commands.

## Reconnection Logic

The current implementation opts for a **simple external watchdog** rather than automatic reconnection inside the `pkg/streamdeck` package:

1. The application layer (see `internal/app/hwinfostreamdeckplugin/plugin.go`) starts a background goroutine that polls `client.Exited()` every second.
   If the subprocess or WebSocket dies, it re-establishes the helper and calls `StreamDeck.Connect()` again.
2. This keeps the transport code in `streamdeck.go` lean while still guaranteeing resiliency if the Stream Deck runtime restarts (common when upgrading firmware or quitting the app).

### Planned Native Auto-Reconnect

If you need library-level reconnection, the simplest enhancement path is:

```go
func (sd *StreamDeck) ConnectWithRetry(ctx context.Context, max int) error {
    var attempt int
    for {
        if err := sd.Connect(); err == nil {
            return nil
        }
        attempt++
        if max > 0 && attempt >= max {
            return fmt.Errorf("failed after %d attempts", attempt)
        }
        select {
        case <-time.After(time.Second):
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

Key considerations:

* **Back-off** – Start with 1 s delay; double on successive failures up to, say, 30 s.
* **Lost Context** – After reconnect you must resend `registerPlugin` and cache of any local state (e.g., current settings) because Stream Deck forgets previous connections.
* **In-flight Commands** – Wrap writes with a mutex and check `sd.conn == nil` before attempting to send; queue commands during downtime if necessary.

Until this is implemented, the external watchdog pattern remains sufficient for most users—Stream Deck restarts are infrequent and the reconnect time is <2 s.
