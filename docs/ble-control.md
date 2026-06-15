# BLE control

Control Lepro lights over Bluetooth — no MQTT, no `lepro-lab.toml` required.

Install the Python tools once (`cd lepro && uv sync`). Commands below use `uv run` from
`lepro/`; you can omit `uv run` if the venv is activated.

## Bond once

```bash
cd lepro
uv run lepro-bond --mac AA:BB:CC:DD:EE:FF
```

Credentials are saved to `~/.lepro/<mac>.json`. Re-bond with `--force-bond`.

Probe discovery only (no bond TX):

```bash
uv run lepro-bond --mac AA:BB:CC:DD:EE:FF --probe
```

## Status

```bash
uv run lepro-status --mac AA:BB:CC:DD:EE:FF
```

Returns decrypted DP JSON from the light (when the firmware responds to getDpState).

## Control

```bash
uv run lepro-ctl on --mac AA:BB:CC:DD:EE:FF
uv run lepro-ctl off --mac AA:BB:CC:DD:EE:FF
uv run lepro-ctl color --mac AA:BB:CC:DD:EE:FF --hue 0 --sat 1000 --val 1000
```

### Modes and presets

```bash
uv run lepro-ctl mode list
uv run lepro-ctl mode play diy --mac AA:BB:CC:DD:EE:FF --color FFAA00 --effect Steady
uv run lepro-ctl mode debug gradient --mac AA:BB:CC:DD:EE:FF   # print DP JSON only
```

Use `--dry-run` to print TX packets without sending BLE.

## ESP32 cred export

```bash
uv run lepro-bond --mac AA:BB:CC:DD:EE:FF --export-provision -o esp32/include/lepro_creds.h
```

See [esp32.md](esp32.md).

## Troubleshooting

- **Light not found:** ensure the light is powered and in range; try onboarding mode if the app required it for first pairing.
- **No credentials:** run `lepro-bond` first.
- **macOS:** BLE can be flaky; retry scan or move closer.
