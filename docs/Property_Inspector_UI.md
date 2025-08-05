# Property Inspector UI
## Table of Contents
- [File Layout](#file-layout)
- [Communication Flow](#communication-flow)
- [UX Guidelines](#ux-guidelines)
- [CSS & Theming Reference](#css--theming-reference-todo)
- [Screenshots](#screenshots-todo)

The Property Inspector (PI) is the HTML/JS interface that appears in Stream Deck when the user configures a button.

## File Layout

```
com.exension.hwinfo.sdPlugin/
  index_pi.html     # Entry point loaded by Stream Deck
  index_pi.js       # JS that handles Stream Deck PI API
  css/
    sdpi.css        # Elgato style baseline
    local.css       # Custom styles for this plugin
```

## Communication Flow

1. PI establishes a WebSocket to Stream Deck runtime (different port than plugin).
2. It receives global + action-specific settings.
3. JavaScript populates dropdowns with available sensors (sent from plugin via `sendToPropertyInspector`).
4. User changes are `setSettings` to persist.

## UX Guidelines

* Keep the dropdown searchable—sensor lists can be large.
* Show a warning banner when HWiNFO is not running.
* Validate format strings client-side.

## Format String Reference

The format string field supports standard Go format specifiers plus custom extensions:

### Standard Format Specifiers
* `%f` - Format as floating point (e.g., "123.45")
* `%d` - Format as integer (e.g., "123")
* `%.2f` - Format with 2 decimal places (e.g., "123.45")

### Custom Format Extensions (v2.1.0+)
* `%t` - Format with thousands separators (e.g., "1,234" instead of "1234")
* `%b` - Format as boolean text (0 → "NO", 1 → "YES")
* `%u` - Dynamic unit conversion (e.g., "1.5 GB/s" instead of "1500 MB/s")

### Examples
* `%.1f °C` -> "23.4 °C"
* `%t RPM` -> "1,234 RPM"
* `%t V` -> "1,234 V"
* `%b` -> "YES" (for value = 1)
* `%u` -> "1.5 GB/s" (for network sensors)

## Value Text Stroke Configuration

The Property Inspector includes controls for customizing the value text stroke:

### Stroke Colour
- **Location**: Under "Graph Colors" section
- **Control**: Colour picker for "Value Text Stroke"
- **Purpose**: Sets the outline colour for value text

### Stroke Size
- **Location**: Range slider below colour pickers
- **Range**: 0-5 pixels
- **Default**: 1 pixel
- **Purpose**: Controls the thickness of the text outline

### Real-time Preview
- Changes are applied immediately to show the effect
- Current value is displayed below the slider
- Stroke is disabled when size is set to 0

---

> **TODO**: Include screenshots and CSS class reference.
