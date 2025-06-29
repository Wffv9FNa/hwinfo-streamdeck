# gRPC Service
## Table of Contents
- [Proto Definition](#proto-definition)
- [Handshake & `go-plugin` Framework](#handshake--go-plugin-framework)
- [Data Types](#data-types)
- [Running the Service](#running-the-service)
- [Auth & Streaming Modes](#authentication--streaming-modes-todo)

Optional plugin side-car exposing real-time sensor data to other processes.

## Proto Definition

Located at `pkg/service/proto/hwservice.proto`.

```proto
service HWService {
  rpc PollTime (google.protobuf.Empty) returns (PollTimeResponse);
  rpc Sensors (google.protobuf.Empty) returns (SensorList);
  rpc ReadingsForSensorID (SensorRequest) returns (ReadingList);
}
```

## Handshake & Plugin Framework

* Uses `hashicorp/go-plugin` for host ↔ plugin comms.
* Shared `Handshake` config in `pkg/service/interface.go` ensures version compatibility.

## Data Types

| Proto Message | Go Interface |
| --- | --- |
| `Sensor` | `hwsensorsservice.Sensor` |
| `Reading` | `hwsensorsservice.Reading` |

## Running the Service

```bash
hwinfo-plugin --port 50051 --log-level debug
```

Clients can then connect via standard gRPC tooling.

---

> **TODO**: Add auth considerations, streaming mode, and performance benchmarks.
