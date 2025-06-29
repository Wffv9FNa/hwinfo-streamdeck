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

---

> **TODO**: Add debug log locations and diagnostic flags.
