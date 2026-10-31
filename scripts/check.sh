#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/.."
(cd apps/dashboard && npm run lint && npm run typecheck && npm run build)
(cd services/gateway && go vet ./... && go test ./... && go build -o /tmp/satellite-gateway ./cmd/gateway)
(cd services/simulator && .venv/bin/ruff check . && .venv/bin/ruff format --check . && .venv/bin/python -m unittest discover -s tests -v && .venv/bin/python -m satellite_simulator --once)
docker compose config --quiet
