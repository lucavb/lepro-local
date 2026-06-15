# Research & debug tools

Optional utilities under `lepro/lepro/tools/`. Run from `lepro/` after `uv sync`:

```bash
cd lepro
uv run python -m lepro.tools.<module> [args]
```

| Module | Purpose |
|--------|---------|
| `decrypt_capture` | Decrypt DP payloads from `captures.log` |
| `debug_hijack` | Diagnose CDN pin + MQTT mTLS lab setup from a pcap |
| `mqtt_ca_dump` | Dump MQTT/CDN trust material for provisioning |
| `replay` | Replay captured BLE session to a light |
| `esp_idf_bin_to_elf` | Wrap a raw ESP32 image as a minimal ELF for Ghidra |
| `validate_ota_parser` | Cross-check OTA JSON parser against firmware ELF |

User-facing workflows use the `lepro-*` CLIs (`lepro-bond`, `lepro-ota`, etc.) — see [setup.md](setup.md).
