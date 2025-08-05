# Extending the Project
## Table of Contents
- [Adding a New Reading Type](#adding-a-new-reading-type)
- [Creating Additional Stream Deck Actions](#creating-additional-stream-deck-actions)
- [Internationalisation](#internationalisation)
- [Performance Considerations](#performance-considerations-todo)

Guidelines for adding new capabilities.

## Adding a New Sensor Reading Type

1. Update `ReadingType` enum in `pkg/service/interface.go`.
2. Modify `internal/hwinfo/reading.go` to map the raw HWiNFO type.
3. Update PI dropdown list to include the new label.

## Adding New Formatting Extensions

The plugin supports custom formatting extensions that can be added by:

1. **Define the format token** in `internal/app/hwinfostreamdeckplugin/plugin.go`
2. **Implement the formatting logic** in `handleFormatString()` function
3. **Update Property Inspector** to document the new token
4. **Add examples** to the documentation

### Example: Adding a Percentage Formatter

```go
// In handleFormatString()
if strings.Contains(format, "%p") {
    // Convert to percentage
    percentage := value * 100
    return strings.ReplaceAll(format, "%p", fmt.Sprintf("%.1f", percentage))
}
```

### Backward Compatibility

All new formatting extensions should:
- Be backward compatible with existing configurations
- Have clear documentation and examples
- Include proper error handling

## Creating Additional Stream Deck Action

* Copy existing action handler for reference.
* Register new UUID in `manifest.json`.
* Implement new action in JS PI if UI required.

## Internationalisation

_All static strings are currently English only._ Consider externalising to JSON for localisation.

---

> **TODO**: Document performance considerations when polling more frequently.
