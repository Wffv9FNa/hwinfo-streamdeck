# HWiNFO Subsystem
## Table of Contents
- [Shared Memory Basics](#shared-memory-basics)
- [Key Files](#key-files)
  - [`internal/hwinfo/hwinfo.go`](#key-files)
  - [`internal/hwinfo/shmem/shmem.go`](#key-files)
  - [`internal/hwinfo/mutex/mutex.go`](#key-files)
  - [`internal/hwinfo/sensor.go`](#key-files)
  - [`internal/hwinfo/reading.go`](#key-files)
- [Polling Strategy](#polling-strategy)
- [Sensor & Reading Model](#sensor--reading-model)
- [Error Handling & Edge Cases](#error-handling--edge-cases)
- [Byte-Layout Reference](#byte-layout-reference-todo)

Deep dive into reading HWiNFO shared memory.

## Shared Memory Basics

HWiNFO exposes a `HWiNFO_SENSORS_SHARED_MEM2` structure in a named section of process memory. The plugin attaches to this region and copies it every update tick.

```text
+------------------------------+
| DWORD dwSignature "HWiS"     |
| DWORD dwVersion              |
| DWORD dwRevision             |
| time_t tPollTime             |
| ...                          |
+------------------------------+
```

### Memory Mapping Details

HWiNFO creates the following named kernel objects:

| Object | Win32 Name | Purpose |
| --- | --- | --- |
| File-mapping | `Global\HWiNFO_SENS_SM2` | Read-only view containing the `HWiNFO_SENSORS_SHARED_MEM2` structure followed by sensor & reading arrays. |
| Mutex | `Global\HWiNFO_SENS_SM2_MUTEX` | Guards writers while HWiNFO updates the shared memory. Readers should attempt a **non-blocking** `WaitForSingleObject` with short timeout to avoid UI stalls. |

### Recommended Access Sequence (from Go)

1. `OpenFileMappingW(FILE_MAP_READ, FALSE, L"Global\\HWiNFO_SENS_SM2")`
   Returns an `HANDLE` to the section object.
2. `MapViewOfFile(handle, FILE_MAP_READ, 0, 0, 0)`
   Yields a `void*` pointer to the beginning of `HWiNFO_SENSORS_SHARED_MEM2`.
3. Verify `dwSignature == 'HWiS'` (little-endian: `0x53574948`). If signature equals `'DEAD'` the provider is shutting down – treat as error but keep retrying.
4. Interpret the header fields:
   * `dwOffsetOfSensorSection` (bytes from base)
   * `dwNumSensorElements` / `dwSizeOfSensorElement`
   * `dwOffsetOfReadingSection`
   * `dwNumReadingElements` / `dwSizeOfReadingElement`
5. Iterate sensor or reading arrays by slicing the original byte slice using the offsets & element sizes – **no additional struct packing needed**.
6. Call `UnmapViewOfFile` when finished and `CloseHandle` on the section & mutex.

The Go implementation in this repository avoids CGO for mapping by using `golang.org/x/sys/windows` syscall wrappers; CGO is only required for the `C` type definitions to calculate field sizes reliably.

### Structure Alignment Notes

The `HWiNFO_SENSORS_SHARED_MEM2` header and element structs are compiled with Microsoft's default packing (`#pragma pack(push, 1)` in `hwisenssm2.h`). Therefore there is **no padding between fields** and all numeric types are little-endian. When slicing the raw byte array we rely on these exact sizes:

| Struct | Size (bytes) |
| --- | --- |
| `HWiNFO_SENSORS_SHARED_MEM2` | 68 |
| `HWiNFO_SENSORS_SENSOR_ELEMENT` | 128 |
| `HWiNFO_SENSORS_READING_ELEMENT` | 32 |

If HWiNFO upgrades the spec (e.g., v2) the `dwVersion` and `dwSizeOf*Element` fields will change—our code checks these at runtime and logs a warning.

## Key Files

| File | Role |
| --- | --- |
| `internal/hwinfo/hwinfo.go` | Top-level API for reading and streaming shared memory |
| `internal/hwinfo/shmem/shmem.go` | Windows syscall wrappers for opening and copying the memory-mapped file |
| `internal/hwinfo/mutex/mutex.go` | Ensures exclusive access while reading |
| `internal/hwinfo/sensor.go` | Converts raw bytes → Go structs for sensors |
| `internal/hwinfo/reading.go` | Same conversion for per-reading data |

### `internal/hwinfo/hwinfo.go`
High-level façade that other packages import. Key elements:

* **`SharedMemory` struct** – wraps a copied byte slice and provides typed accessor methods (signature, version, counts, iterators).
* **`ReadSharedMem()`** – One-shot helper that locks mutex, maps view, copies bytes, and returns `*SharedMemory`.
* **`StreamSharedMem()`** – Launches a goroutine that calls `ReadSharedMem` once per second and pipes the result through a channel; used by both the Stream Deck plugin and gRPC side-car.
* **Iterators** – `IterSensors()` / `IterReadings()` expose typed channels over sensors/readings without additional allocations.

### `internal/hwinfo/shmem/shmem.go`
Encapsulates the Win32 section mapping logic.

* Uses `OpenFileMapping` + `MapViewOfFile` from `golang.org/x/sys/windows`.
* Copies the full shared memory into a pre-allocated Go slice (`buf`) to avoid holding the mapping while parsing.
* Computes total byte length dynamically from header offsets.

### `internal/hwinfo/mutex/mutex.go`
Tiny package that provides **global mutual exclusion** between HWiNFO writers and our readers.

* `Lock()` – Opens the named mutex and blocks the caller until it can acquire read access. Returns a Go error with decoded Win32 error description if acquisition fails.
* `Unlock()` – Releases the mutex and closes the handle. Internal `sync.Mutex` protects multiple nested calls.

### `internal/hwinfo/sensor.go`
Parses **sensor descriptors** (motherboard, CPU, GPU…).

* Wraps a `PHWiNFO_SENSORS_SENSOR_ELEMENT` pointer without copying—zero-allocation wrapper.
* Methods such as `ID()`, `NameOrig()`, and `NameUser()` abstract the underlying C strings.

### `internal/hwinfo/reading.go`
Represents individual **sensor readings** (temperature, voltage, RPM…).

* Enum `ReadingType` mirrors HWiNFO's reading categories and includes a `.String()` method for human-readable output.
* `NewReading()` constructs a wrapper over raw bytes analogous to `Sensor`.
* Provides getters: `Value()`, `ValueMin()`, `ValueMax()`, `ValueAvg()`; each uses precise pointer arithmetic to avoid struct copying.

Together, these files form a lightweight, allocation-friendly bridge between low-level C structs and idiomatic Go objects – the foundation for all higher-level services.

## Polling Strategy

`StreamSharedMem()`—exported from `internal/hwinfo/hwinfo.go`—is the canonical way other packages receive incremental sensor updates without worrying about mutex/Win32 details.

### How it Works

1. **Channel Construction** – Function allocates `chan Result` where:
   * `Result.Shmem` is `*SharedMemory` (may be `nil` on error)
   * `Result.Err`   is `error` (non‐nil when mapping or copy fails)
2. **Goroutine Launch** – Immediately spawns an anonymous goroutine.
3. **Initial Read** – Calls internal helper `readAndSend(ch)` to push the first snapshot as soon as possible.
4. **Ticker Loop** – Uses `time.NewTicker(1 * time.Second)` to wait between reads. The one-second interval matches HWiNFO's typical polling cadence and keeps CPU overhead negligible (<0.2 %).
5. **Non-blocking Send** – Each iteration simply invokes `readAndSend` which performs:
   * `ReadSharedMem()` (includes mutex lock, map, copy, unlock)
   * Sends `Result` struct into the channel (synchronous send; if consumer is slow this goroutine blocks, back-pressure is intentional).
6. **Channel Lifetime** – The goroutine runs until the parent process exits; there is currently no cancellation context. Consumers that need shutdown semantics should wrap the call in their own goroutine and select on `ctx.Done()`.

### Consumer Example

```go
updates := hwinfo.StreamSharedMem()
for res := range updates {
    if res.Err != nil {
        log.Printf("hwinfo poll error: %v", res.Err)
        continue
    }
    sig := res.Shmem.Signature()
    if sig != "HWiS" {
        log.Println("provider offline")
        continue
    }
    for s := range res.Shmem.IterSensors() {
        fmt.Printf("sensor: %s\n", s.NameUser())
    }
}
```

### Performance Notes

* **Allocation** – The shared-memory buffer is copied into a reusable package-level slice; size adjustments are made only when capacity is insufficient, keeping GC pressure low.
* **Lock Duration** – Mutex is held only for the short `memcpy` (<1 ms on modern systems).
* **Error Handling** – If the memory mapping fails (e.g., HWiNFO closed), the error field is populated; the goroutine does NOT exit so that recovery is automatic once HWiNFO launches again.

Developers may tune the interval by copying the original code and adjusting the ticker, but 1 s strikes a good balance between UI responsiveness and CPU usage.

## Sensor & Reading Model

HWiNFO splits data into two parallel arrays stored in shared memory:

1. **Sensors** – high-level devices (CPU, GPU, motherboard, SSD, etc.)
2. **Readings** – individual metrics that belong to a sensor (temperature, clock, voltage…)

### Sensor Element (`HWiNFO_SENSORS_SENSOR_ELEMENT`)

| Field (C Name) | Go Accessor | Notes |
| --- | --- | --- |
| `dwSensorID` | `SensorID()` | Device-type identifier (e.g., all CPUs share same ID) |
| `dwSensorInst` | `SensorInst()` | Instance index to differentiate multiple devices of same type (GPU #1 vs #2) |
| `szSensorNameOrig` | `NameOrig()` | Original name from HWiNFO database |
| `szSensorNameUser` | `NameUser()` | User-renamed label (preferred for display) |

The plugin defines `(*Sensor).ID()` = `SensorID*100 + SensorInst` to create a **stable, string-based UID** that is easy to store in Stream Deck settings.

### Reading Element (`HWiNFO_SENSORS_READING_ELEMENT`)

| Field | Go Accessor | Description |
| --- | --- | --- |
| `dwReadingID` | `Reading.ID()` | Unique within its sensor |
| `tReading` | `Reading.Type()` | One of the enum constants (`Temp`, `Volt`, `Fan`, …) |
| `dwSensorIndex` | `Reading.SensorIndex()` | Index into sensor array indicating ownership |
| `szLabelOrig` / `szLabelUser` | `LabelOrig()` / `LabelUser()` | Metric label (e.g., "Core 0 Clock") |
| `szUnit` | `Unit()` | Unit string shown to the user |
| `dValue` | `Value()` | Current value (double) |
| `dValueMin` / `dValueMax` / `dValueAvg` | Min/Max/Avg getters | Aggregated stats since HWiNFO launch |

### Type Enumeration

`internal/hwinfo/reading.go` defines a Go enum that mirrors HWiNFO's `tReading` values:

```go
type ReadingType int

const (
    ReadingTypeNone ReadingType = iota
    ReadingTypeTemp   // °C
    ReadingTypeVolt   // V
    ReadingTypeFan    // RPM
    ReadingTypeCurrent // A
    ReadingTypePower  // W
    ReadingTypeClock  // MHz
    ReadingTypeUsage  // % / MB
    ReadingTypeOther
)
```

A helper `.String()` method converts the enum to human-readable form—useful for debugging and building dropdown lists.

### Mapping Readings to Sensors

Because readings reference their parent sensor **by index** (not pointer), the plugin must first build an array of sensor UIDs:

```go
sensorUIDs, _ := service.SensorIDByIdx()
for r := range shmem.IterReadings() {
    sid := sensorUIDs[int(r.SensorIndex())]
    readingsBySensor[sid] = append(readingsBySensor[sid], r)
}
```

This indirection keeps the shared-memory layout compact and allows variable-length arrays.

### Formatting for Display

The Stream Deck button needs a short, readable string. The application layer:

1. Retrieves the `Reading` for configured `SensorUID` + `ReadingID`.
2. Applies optional divisor (e.g., divide by 1000 for displaying volts as V instead of mV).
3. Uses user-supplied Go template or falls back to default formatting switch based on `ReadingType` (see `Plugin.applyDefaultFormat`).

Example default outputs:

| Type | Raw Value | Display |
| --- | --- | --- |
| Temp | `56.7` | `57 °C` |
| Clock | `4500` | `4500 MHz` |
| Usage | `74.2` | `74 %` |

### Memory/Performance Considerations

* Creating `Sensor`/`Reading` wrappers is **zero-copy**—we only keep pointers into the copied byte slice.
* Average allocations per polling tick: **O(numSensors + numReadings)** wrappers (on the order of hundreds). Garbage-collector impact is minimal given the 1 s polling interval.
* Strings decoded from C arrays are re-created each tick; heavy users may cache them by UID/ID.

## Error Handling & Edge Cases

## Byte-Layout Reference
