# Plugin Application Flow
## Table of Contents
- [Runtime Overview](#runtime-overview)
- [Action Manager](#action-manager)
- [Settings Structure](#settings-structure)
- [Event Handlers](#event-handlers)
  - [willAppear](#willappear)
  - [keyDown](#keydown)
  - [didReceiveSettings](#didreceivesettings)
- [Periodic Tile Update Cycle](#periodic-tile-update-cycle)
- [Error Handling](#error-handling)
- [Lifecycle Sequence Diagram](#full-lifecycle-sequence-diagram-todo)

Runtime behaviour within `internal/app/hwinfostreamdeckplugin`.

## Action Manager

* Maintains a map of active contexts → settings.
* Updates tiles every second via the Stream Deck SDK with fresh sensor data.

## Settings Structure

```go
type actionSettings struct {
    IsValid bool
    SensorID string
    Format string // e.g. "{{.Value}} °C"
    ValueTextStrokeColor string // Stroke colour for value text
    ValueTextStrokeSize float64 // Stroke size (0-5 pixels)
}
```

Settings are persisted by the Stream Deck runtime and delivered back on startup or when the user changes them in the Property Inspector.

## Event Handlers

| Event | Function | Purpose |
| --- | --- | --- |
| `willAppear` | `handleWillAppear` | Called when a key is added to the deck |
| `keyDown` | `handleKeyDown` | User pressed the button |
| `didReceiveSettings` | Updates the action's settings |

## Periodic Tile Update

`actionManager.Run()` launches a ticker; each tick iterates through actions and calls `updateTiles` with current readings for that sensor. The update process includes:

1. **Format Application**: Apply custom formatting tokens (`%b`, `%t`, `%u`)
2. **Stroke Rendering**: Apply stroke effects based on user configuration
3. **Text Rendering**: Draw text with stroke outline if enabled

## Error Handling

* Missing sensor → shows alert icon.
* HWiNFO signature dead → greys out tile.

---

> **TODO**: Sequence diagram of a full lifecycle from willAppear → didReceiveSettings → periodic updates.
