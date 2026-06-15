# lepro-local

Knowledge base and tools for running **Lepro WiFi/BLE devices** on your own terms: bond over Bluetooth, control lights locally, and optionally point firmware at **your** MQTT broker instead of Lepro cloud.

**Tested on:** Lepro ZB1 RGB-IC outdoor string lights (firmware `3_le_light_zb1_pid_55_v2.3.18`). Other Lepro app devices (B1 bulbs, S1 strips, P1 plug, etc.) may share the same BLE/MQTT stack — unverified.

**MQTT is not required.** Many users only need the BLE tools below.

## Prerequisites

- [uv](https://docs.astral.sh/uv/) (Python 3.13+)
- Bluetooth adapter (for BLE commands)
- Optional: [Go](https://go.dev/) 1.26+ (for the MQTT bridge)

All Python CLIs live under `lepro/`. Install once:

```bash
cd lepro
uv sync
```

Run commands with `uv run lepro-<cmd> …` from `lepro/`, or activate `.venv/bin/activate` and call `lepro-<cmd>` directly.

## Quick start — BLE only

```bash
cd lepro
uv run lepro-bond --mac AA:BB:CC:DD:EE:FF
uv run lepro-status --mac AA:BB:CC:DD:EE:FF
uv run lepro-ctl on --mac AA:BB:CC:DD:EE:FF
uv run lepro-ctl color --mac AA:BB:CC:DD:EE:FF --hue 120 --sat 1000 --val 1000
```

See [docs/ble-control.md](docs/ble-control.md).

## Quick start — local MQTT (optional)

```bash
cp lepro-lab.toml.example lepro-lab.toml   # edit WiFi, hostnames, device id
./deploy/docker/up.sh
cd lepro
uv run lepro-firmware fetch
uv run lepro-firmware patch
uv run lepro-ota --mac AA:BB:CC:DD:EE:FF --firmware ../firmware/3_le_light_zb1_pid_55_v2.3.18.patched.bin
uv run lepro-bond --mac AA:BB:CC:DD:EE:FF --force-bond
uv run lepro-provision --mac AA:BB:CC:DD:EE:FF --config ../lepro-lab.toml
```

`lepro-firmware patch` pins your CDN trust anchor and neuters two stock gates so root/client certs are re-fetched on every provision — see [docs/firmware.md](docs/firmware.md).

Then control via native MQTT topics (`le/<device_id>/prp/set`) or run the Go bridge — see [docs/mqtt-setup.md](docs/mqtt-setup.md).

## Tools

| Command | Purpose |
|---------|---------|
| `lepro-bond` | Bond once; `--export-provision` for ESP32 creds |
| `lepro-status` | Query DP state over BLE |
| `lepro-ctl` | `on` / `off` / `color` / `mode` over BLE |
| `lepro-firmware` | `fetch` / `patch` / `verify` stock + patched images |
| `lepro-ota` | BLE firmware flash |
| `lepro-provision` | Point bulb at your WiFi + MQTT/CDN |

## Repository layout

```
lepro/              Python library + CLIs (cd here for uv sync / pytest)
bridge/             Go MQTT bridge (Tasmota-style cmnd/stat ↔ Lepro prp topics)
docs/               Protocol notes, setup guides
deploy/docker/      Local Mosquitto + cert CDN (docker compose)
deploy/lepro-debug/ Optional k8s lab stack (advanced)
firmware/           RE notes + manifest (binaries not in git)
tests/fixtures/     Golden DP payloads for bridge tests
lepro-lab.toml.example   Lab WiFi/MQTT config template (copy to lepro-lab.toml)
```

No Python modules live at the repo root — use `lepro-*` commands from `lepro/` or `python -m lepro.tools.*` (see [docs/tools.md](docs/tools.md)).

## Documentation

- [docs/setup.md](docs/setup.md) — index of both paths
- [docs/ble-control.md](docs/ble-control.md) — bond, status, control
- [docs/mqtt-setup.md](docs/mqtt-setup.md) — firmware, OTA, lab, provision
- [docs/protocol.md](docs/protocol.md) — BLE + native MQTT + DP schema
- [docs/firmware.md](docs/firmware.md) — download, patch, flash
- [docs/bridge-contract.md](docs/bridge-contract.md) — bridge topic map and DP rules
- [docs/esp32.md](docs/esp32.md) — export bond creds for ESP32
- [docs/tools.md](docs/tools.md) — optional research/debug scripts
- [docs/examples/homebridge-easy-mqtt.md](docs/examples/homebridge-easy-mqtt.md) — optional HomeKit example

## License

Use at your own risk. Not affiliated with Lepro Innovation Inc.
