# Architecture

All infrastructure is simulated and runs locally. Services are independently runnable.

## Intended flow

1. Python simulator generates satellite telemetry.
2. Simulator sends telemetry to the Go gateway through gRPC using the shared protobuf contract.
3. Gateway validates satellite identity and message contents and processes messages in memory.
4. Gateway broadcasts updates to the Next.js dashboard through WebSockets.
5. Dashboard displays network activity and sends simulated commands through the future WebSocket connection.
6. Gateway forwards commands to the appropriate simulated satellite. A future simulator-initiated stream or command polling RPC must be designed for this reverse path.

The initial unary `SendTelemetry` contract only establishes the first planned transport boundary. Command and acknowledgment messages are reserved schema structures; no command RPC or reverse channel exists yet.

## What exists

- Responsive dashboard with reusable panels, illustrative fleet identifiers, empty telemetry log, disabled controls, and explicit scaffold labels. It makes no network requests to the gateway.
- Go HTTP `/health` endpoint, environment-backed ports, server timeouts, and signal-driven graceful shutdown. Health reports the HTTP service only, not end-to-end readiness.
- Python immutable configuration and satellite dataclasses, idle lifecycle loop, signal handling, startup validation, and no-op telemetry/connection/outage method placeholders.
- Versioned proto3 definitions and a service contract, without generated bindings.
- Independent Docker images and a Compose network; HTTP health checks for dashboard and gateway.

## Planned boundaries

Gateway `internal/httpserver` owns HTTP handlers. `internal/grpcserver` and `internal/websocket` currently contain package documentation only. Future transport handlers should delegate to small domain functions once behavior is needed. Simulator configuration and lifecycle remain separate from future telemetry and gRPC implementation. Browser state can remain simple until real event subscriptions exist.

No authentication, persistence, real orbital modeling, real satellite integration, cloud infrastructure, or Bazel pipeline is present.
