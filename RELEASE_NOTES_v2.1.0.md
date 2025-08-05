# HWiNFO Stream Deck Plugin v2.1.0

## 🎉 New Features

### 📊 Enhanced Formatting Options

**Boolean Formatting (`%b`)**
- Convert 0/1 values to YES/NO for better readability
- Perfect for sensors that return binary states

**Thousands Separator (`%t`)**
- Add thousands separators to large numbers
- Example: "1,234" instead of "1234"
- Makes large sensor readings much more readable

**Dynamic Unit Conversion (`%u`)**
- Automatically convert units based on value size
- Perfect for network sensors: KB/s → MB/s → GB/s
- Example: "1.5 GB/s" instead of "1500 MB/s"
- Supports data transfer rates and storage units

### 🎨 Improved Text Rendering

**Value Text Stroke**
- Add stroke/outline to value text for better readability
- Customizable stroke colour to match your theme
- Configurable stroke size (0-5 pixels) for perfect visibility
- Real-time preview in Property Inspector

### 🔧 Technical Improvements

**Enhanced Degree Symbol Support**
- Improved handling of degree symbol (°)
- Proper UTF-8 and ISO8859-1 encoding support
- Ensures correct display across different systems

**UI Improvements**
- Updated Property Inspector layout for new controls
- Better organisation of colour pickers
- Improved user experience with 2-column layout

## 🚀 How to Use New Features

### Boolean Formatting
```
Format: %b
Example: "CPU Usage: %b" → "CPU Usage: YES" or "CPU Usage: NO"
```

### Thousands Separator
```
Format: %t
Example: "Memory: %t MB" → "Memory: 1,234 MB"
```

### Dynamic Unit Conversion
```
Format: %u
Example: "Network: %u" → "Network: 1.5 GB/s" (instead of "1500 MB/s")
```

### Value Text Stroke
1. Open Property Inspector for any HWiNFO action
2. Under "Graph Colors" section, find "Value Text Stroke"
3. Choose your stroke colour
4. Adjust stroke size (0-5) for optimal visibility

## 🔄 Backward Compatibility

All new formatting options are **fully backward compatible**:
- Existing configurations will continue to work unchanged
- New format verbs are optional and additive
- No breaking changes to existing functionality

## 📦 Installation

1. Download the `com.exension.hwinfo.streamDeckPlugin` file from this release
2. Double-click the file to install the plugin
3. Restart Stream Deck if prompted
4. Add HWiNFO actions to your Stream Deck and configure the new formatting options!

## 🐛 Bug Fixes

- Fixed degree symbol display issues on certain systems
- Improved text rendering performance
- Enhanced error handling for malformed sensor data

## 📋 Requirements

- Stream Deck Software 6.4 or later
- Windows 10 or later
- HWiNFO64 v7.0 or later

---

**Note**: This plugin is not affiliated with HWiNFO64. For more information and to download HWiNFO64, visit https://www.hwinfo.com

## 🔗 Links

- [GitHub Repository](https://github.com/shayne/hwinfo-streamdeck)
- [HWiNFO64 Download](https://www.hwinfo.com)
- [Stream Deck SDK Documentation](https://developer.elgato.com/documentation/stream-deck/sdk/overview/)
