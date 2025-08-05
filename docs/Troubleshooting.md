# Troubleshooting
## Table of Contents
- [Problem / Cause / Fix Matrix](#problem--cause--fix-matrix)
- [Debug Log Locations](#debug-log-locations-todo)
- [Advanced Diagnostics](#advanced-diagnostics-todo)

Common problems and fixes.

| Symptom | Possible Cause | Fix |
| --- | --- | --- |
| Tile shows `—` | HWiNFO not running | Start HWiNFO or enable sensor shared memory in settings |
| Tile stuck on old value | Mutex locked by previous crash | Run `kill-streamdeck.bat` then restart Stream Deck |
| Plugin not listed | Manifest JSON invalid | Check `manifest.json` syntax via Stream Deck validator |
| gRPC client cannot connect | Port blocked | Verify `hwinfo-plugin` running and firewall rules |
| Format tokens not working | Stream Deck Software < 6.4 | Update to Stream Deck Software 6.4 or later |
| Stroke not visible | Stroke size set to 0 | Increase stroke size in Property Inspector |
| Dynamic unit conversion not working | Incorrect format token | Use `%u` for dynamic unit conversion, not `%t` |

---

> **TODO**: Add debug log locations and diagnostic flags.
