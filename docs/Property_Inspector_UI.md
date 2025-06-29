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

---

> **TODO**: Include screenshots and CSS class reference.
