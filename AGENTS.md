# AGENTS.md

This repo contains Python tooling for Lepro BLE control plus an optional Go MQTT bridge. Keep changes scoped to the relevant subproject and prefer the existing docs over inventing new workflows.

## Core commands

```bash
# Python tools (pyproject.toml is under lepro/)
cd lepro
uv sync
uv run lepro-status --help
uv run pytest

# Go bridge
cd bridge
go test ./...
go build -o lepro-bridge ./cmd/lepro-bridge

# Optional local lab stack (from repo root)
./deploy/docker/up.sh
```

## Project layout

- `lepro/` is the Python package, CLI entrypoints, and Python tests (`lepro/tests/python/`).
- `bridge/` is the Go MQTT bridge and its tests.
- `docs/` is the source of truth for setup, protocol details, and bridge behavior.
- `tests/fixtures/` contains golden DP payloads used by bridge-related tests.
- `deploy/docker/` is the default local Mosquitto + CDN stack; `deploy/lepro-debug/` is an optional k8s lab.

## What to read first

- `README.md` for the high-level repo shape.
- `docs/setup.md` for the correct path depending on BLE-only or MQTT work.
- `docs/bridge-contract.md` before changing anything in `bridge/`.
- `bridge/README.md` and `deploy/docker/README.md` for bridge and lab-stack specifics.

## Conventions

- Keep Python code under `lepro/`; do not add new root-level modules.
- Run `uv sync` and `uv run pytest` from `lepro/` (there is no root `pyproject.toml`).
- Treat `tests/fixtures/dp/` as canonical for DP encoding and decoding behavior.
- Keep bridge changes aligned with `docs/bridge-contract.md` and the golden fixtures.
- Update docs when behavior or commands change.
- Use the verified device and firmware wording from the docs; avoid overstating support for unverified hardware.

## Safety boundaries

- Do not commit secrets, generated certs, firmware binaries, or other build outputs.
- `lepro-lab.toml` is gitignored (copy from `lepro-lab.toml.example`).
- If a change affects provisioning, firmware patching, or broker auth, double-check the relevant docs before editing code.
