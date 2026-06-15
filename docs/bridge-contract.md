# Bridge contract

Specification for the **Go MQTT bridge** you implement in [`bridge/`](../bridge/). Integration-agnostic: any stack that speaks Tasmota-style `cmnd/` / `stat/` topics can consume the home side.

Golden test vectors: [`tests/fixtures/dp/`](../tests/fixtures/dp/)

## Lepro side (bulb)

| Topic | Direction | Payload |
|-------|-----------|---------|
| `le/<device_id>/prp/set` | bridge → bulb | `{"d":{…}}` |
| `le/<device_id>/prp/rpt` | bulb → bridge | `{"d":{…}}` |

Subscribe to `prp/rpt`, publish commands to `prp/set`. Reference implementation for DP encoding: `lepro/commands.py`.

## Standardized side (integrations)

`<name>` = `friendly_name` from `lepro-lab.toml` (e.g. `patio-lights`).

| Direction | Topic | Payload | Lepro mapping |
|-----------|-------|---------|---------------|
| cmd in | `cmnd/<name>/POWER` | `ON` / `OFF` | `d1` |
| state out | `stat/<name>/POWER` | `ON` / `OFF` | `d1` |
| cmd in | `cmnd/<name>/Dimmer` | `0`–`100` | `d3` = value × 10 (clamp 10–1000) |
| state out | `stat/<name>/Dimmer` | `0`–`100` | `d3` ÷ 10 |
| cmd in | `cmnd/<name>/Color` | `R,G,B` decimals | `d2=1`, `d5` HSV hex |
| state out | `stat/<name>/Color` | `R,G,B` | decode `d5` when `d2=1` |
| state out | `tele/<name>/LWT` | `Online` / `Offline` | optional bridge presence |

### DP encoding (v1)

- Power on: `{"d1":1}` inside `{"d":{…}}`
- Power off: `{"d1":0}`
- Solid color: `{"d1":1,"d2":1,"d3":<brightness>,"d5":"<HHHHSSSSVVVV>"}`
- HSV hex: `f"{hue:04X}{sat:04X}{val:04X}"` — see `color_hsv()` in `commands.py`
- Ignore or pass through read-only `d30` on set

## Broker

- **8883 TLS** — bulb; bridge must publish/subscribe native `le/…` topics (may need same CA as bulb)
- **1883 plain** — home integrations + bridge standardized topics

Config: `[bridge]` section in `lepro-lab.toml.example`.

## Out of scope for v1

Scene modes (`d50`), white CCT (`d2=0`, `d4`), music mode.
