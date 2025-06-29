# Product Context

## Why does this project exist?

Enthusiasts and professionals often monitor CPU/GPU temperatures, clocks, power usage, and other metrics. HWiNFO exposes exhaustive sensor data but its stock UI can be bulky or hidden behind other windows. Stream Deck devices are ubiquitous on many desks; using them as mini hardware dashboards frees up screen space and looks cool.

## Problems it solves

1. Provides a constantly visible hardware monitoring surface.
2. Avoids alt-tabbing or overlay clutter.
3. Enables quick access to multiple sensor readings and potential actions (e.g., open HWiNFO, toggle logging).

## User Experience Goals

- Setup should be as easy as installing any Stream Deck plugin.
- Selecting sensors should be guided (dropdown lists pulled from active HWiNFO session).
- Sensor value text must be readable at a glance.
- Graceful degradation when HWiNFO is not running.

## Stakeholders

- PC enthusiasts / streamers.
- IT admins wanting quick diagnostics.
- Open-source contributors.
