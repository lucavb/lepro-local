# Setup guide

Pick the path that matches what you want.

## BLE only (no broker)

1. [ble-control.md](ble-control.md) — bond, status, control with `lepro-ctl`
2. Optional: [esp32.md](esp32.md) — export creds for an ESP32 BLE client

## Local MQTT (optional)

1. Complete BLE bond ([ble-control.md](ble-control.md))
2. [firmware.md](firmware.md) — download stock image, patch CDN pin + cert-refetch gates
3. [mqtt-setup.md](mqtt-setup.md) — lab deploy, OTA, provision
4. [protocol.md](protocol.md) — native MQTT topics and DP payloads
5. Optional: [bridge-contract.md](bridge-contract.md) + [../bridge/README.md](../bridge/README.md) — Go MQTT bridge
6. Optional: [examples/homebridge-easy-mqtt.md](examples/homebridge-easy-mqtt.md) — one HomeKit integration example

## Hardware

- **Verified:** Lepro ZB1 RGB-IC string light, firmware `3_le_light_zb1_pid_55_v2.3.18`
- **May work:** other Lepro app lights/plugs using the same BLE + `dvc-eu-iot` cloud (unverified)
