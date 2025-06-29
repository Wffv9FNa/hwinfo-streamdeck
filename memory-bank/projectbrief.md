# Project Brief: HWiNFO Stream Deck Plugin

## Vision
Enable Stream Deck devices to display and interact with real-time hardware sensor data collected by HWiNFO, giving users an at-a-glance overview of their system performance.

## Core Requirements
- Read live sensor values from HWiNFO via its shared-memory interface.
- Expose one or more Stream Deck actions that render sensor values on buttons.
- Provide a Property Inspector UI for configuring which sensor to display and formatting options.
- Offer a debugging utility for development and troubleshooting.
- Buildable as a standalone `.streamDeckPlugin` bundle for easy install.

## Success Criteria
- Plugin installs on Windows Stream Deck software without manual steps.
- Users can pick any HWiNFO sensor and see values updating with <1 s latency.
- No noticeable impact on system performance.
- Project is self-contained Go code and minimal JS/HTML assets.

## Out of Scope (for now)
- macOS/Linux support (HWiNFO is Windows-only).
- Historical charting or logging.
- Auto-update mechanism.
