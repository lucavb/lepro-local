# lepro-bridge

Go MQTT bridge: translates Lepro-native topics (`le/<device_id>/prp/…`) to
Tasmota-style `cmnd/` / `stat/` topics for home automation stacks.

## Build

```bash
cd bridge
go test ./...
go build -o lepro-bridge ./cmd/lepro-bridge
./lepro-bridge -config ../lepro-lab.toml
```

## Bridge contract

1. Read [`docs/bridge-contract.md`](../docs/bridge-contract.md) — topic map and DP encoding rules.
2. Use [`tests/fixtures/dp/`](../tests/fixtures/dp/) — golden vectors derived from `lepro/commands.py`.
3. Read config from [`lepro-lab.toml`](../lepro-lab.toml.example) `[bridge]` section.

## Layout

```
internal/
  config/   # load lepro-lab.toml + TLS config
  lepro/    # native MQTT client (prp/set, prp/rpt)
  dp/       # DP JSON encode/decode
  tasmota/  # cmnd/stat topic mapping
```

## Prerequisites

- Bulb provisioned onto your MQTT broker (`lepro-provision`) — see [`docs/mqtt-setup.md`](../docs/mqtt-setup.md).
- Broker running with TLS (8883) for the bulb and plain MQTT (1883) for integrations — see [`deploy/docker/`](../deploy/docker/).

## Container image

```bash
docker build -t lepro-bridge:local -f bridge/Dockerfile bridge
```

Run it with a mounted config file:

```bash
docker run --rm \
  -v "$PWD/lepro-lab.toml:/lepro-lab.toml:ro" \
  lepro-bridge:local \
  -config /lepro-lab.toml
```

GitHub Actions publishes multi-arch images to GHCR as `ghcr.io/<owner>/lepro-bridge`.

## Environment variables

The bridge can also be configured without any TOML file. Supported variables:

- `LEPRO_DEVICE_ID`
- `LEPRO_DEVICE_FRIENDLY_NAME`
- `LEPRO_HOME_BROKER` or `LEPRO_BRIDGE_HOME_BROKER`
- `LEPRO_LEPRO_BROKER` or `LEPRO_BRIDGE_LEPRO_BROKER`
- `LEPRO_CA_FILE` or `LEPRO_BRIDGE_CA_FILE`
- `LEPRO_TLS_SERVER_NAME` or `LEPRO_BRIDGE_TLS_SERVER_NAME`
- `LEPRO_TLS_INSECURE` or `LEPRO_BRIDGE_TLS_INSECURE`
- `LEPRO_TLS_CERT_FILE` / `LEPRO_TLS_KEY_FILE`
- `LEPRO_LAB_CONFIG` for an optional config file path
