# Satellite Communications Gateway

An independent educational prototype for exploring distributed systems and simulated satellite communications. It never communicates with real satellites, ground stations, or external space infrastructure.

This repository is the initial scaffold. The dashboard displays labeled placeholders, the Go gateway serves HTTP health JSON, and the Python simulator runs a signal-aware idle loop. gRPC, WebSockets, telemetry generation, command delivery, and outage simulation are planned, not implemented.

## Architecture and stack

The intended flow is Python 3.13 simulator → protobuf/gRPC → Go gateway → WebSocket → Next.js 16 / React / TypeScript dashboard. Commands will travel in the reverse direction. The dashboard uses plain CSS and reusable server-rendered components. Docker Compose supplies a shared local network. Bazel is a later enhancement; there is no authentication, database, cloud deployment, or external API integration.

```text
apps/dashboard/             Next.js App Router, components, CSS, Dockerfile
services/gateway/           Go module, cmd/gateway, internal configuration and transports
services/simulator/         Python package, pyproject.toml, unittest smoke tests
proto/satellite/v1/         Small shared proto3 contract
scripts/check.sh            Local verification commands
docs/                      Architecture and protobuf guidance
docker-compose.yml         Local service orchestration
.env.example               Compose configuration defaults
```

## Prerequisites

- Node.js 22 LTS and npm ([Next.js minimum](https://nextjs.org/docs/app/getting-started/installation) is Node.js 20.9).
- Go 1.24 or newer, Python 3.13 or newer.
- Docker Engine/Desktop running and Docker Compose v2 for containers.
- Git. The repository is already initialized; no remote is required.

## Local development

Run each service in a separate terminal from the repository root:

```sh
cd apps/dashboard
npm ci
npm run dev
```

```sh
cd services/gateway
go run ./cmd/gateway
```

```sh
cd services/simulator
python3.13 -m venv .venv
.venv/bin/python -m pip install -e '.[dev]'
.venv/bin/python -m satellite_simulator
```

Service configuration comes from environment variables. Per-service `.env.example` files document settings. Next.js loads `.env.local`; Go and Python read process environment only. For either service, export values in your shell (or `set -a; . ./.env.example; set +a` for the trusted checked-in example). No environment loader dependency is needed.

| Service | Default address | Current behavior |
| --- | --- | --- |
| Dashboard | http://localhost:3000 | Static placeholder dashboard |
| Gateway | http://localhost:8080/health | JSON health response |
| gRPC | localhost:50051 (reserved) | No listener yet |
| WebSocket | ws://localhost:8080/ws (reserved) | No handler yet |
| Simulator | No listening port | Idle process; no network traffic |

`GATEWAY_HTTP_PORT`, `GATEWAY_GRPC_PORT`, `GATEWAY_GRPC_ADDRESS`, `SATELLITE_ID`, and `SIMULATOR_INTERVAL_SECONDS` configure services. The dashboard's example WebSocket URL is reserved and currently unused. Dashboard local ports can be changed with `npm run dev -- --port 3001`.

## Docker Compose

```sh
cp .env.example .env
docker compose config --quiet
docker compose up --build -d
docker compose ps
curl --fail http://localhost:8080/health
curl --fail http://localhost:3000
docker compose logs -f
# When finished:
docker compose down
```

If port 8080 is occupied, set `GATEWAY_HTTP_PORT=18080` in `.env`; the health URL becomes http://localhost:18080/health. The scaffold verification used this override and left it in the local ignored `.env`.

Host ports bind to loopback. Compose creates a private default network with service DNS names. Dashboard and gateway have HTTP health checks; the idle simulator has no network readiness check. Services intentionally start independently without waiting for an unimplemented gRPC connection. `GATEWAY_GRPC_PORT` reserves the future internal port and is not published. `DASHBOARD_PORT` controls the dashboard's host port; its container listens on 3000.

## Verification and formatting

```sh
(cd apps/dashboard && npm run lint && npm run typecheck && npm run build)
(cd services/gateway && gofmt -w cmd internal && go vet ./... && go test ./... && go build -o /tmp/satellite-gateway ./cmd/gateway)
(cd services/simulator && .venv/bin/ruff check . && .venv/bin/ruff format --check . && .venv/bin/python -m unittest discover -s tests -v && .venv/bin/python -m satellite_simulator --once)
docker compose config --quiet
```

After installing prerequisites and dependencies, `sh scripts/check.sh` runs the core checks. `npm run build` uses webpack to avoid requiring Turbopack in restricted development environments. `npm start` serves the built dashboard. Python `--once` validates configuration and exits; it does not generate telemetry. See [verification results](docs/verification.md) for checks actually executed during scaffolding.

## Planned milestones

1. Generate bindings and implement one unary telemetry gRPC call with integration tests.
2. Add in-memory satellite registration and basic message validation.
3. Broadcast telemetry over WebSockets and replace dashboard placeholders.
4. Add simulated command delivery and acknowledgments.
5. Add deterministic outage scenarios and recovery behavior.
6. Evaluate Bazel for cross-language builds and code generation.

See [architecture](docs/architecture.md) and [protobuf guidance](proto/README.md).
