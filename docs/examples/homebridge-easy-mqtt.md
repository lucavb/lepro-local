# homebridge-easy-mqtt (optional example)

One way to attach a standardized MQTT side to HomeKit. **Not a project requirement** — any consumer of `cmnd/` / `stat/` topics works.

Prerequisites:

1. Bulb on your MQTT broker ([mqtt-setup.md](../mqtt-setup.md))
2. Your Go bridge running per [bridge-contract.md](../bridge-contract.md) (not included in this repo)
3. `home_broker` in `lepro-lab.toml` points to the plain MQTT listener (for example `mqtt://127.0.0.1:1883`)

## Example `config.json` fragment

Replace `patio-lights` with `friendly_name` from `lepro-lab.toml`.

```json
{
  "type": "lightbulb",
  "name": "Patio Lights",
  "url": "mqtt://127.0.0.1:1883",
  "topics": {
    "getOn": "stat/patio-lights/POWER",
    "setOn": "cmnd/patio-lights/POWER",
    "getBrightness": "stat/patio-lights/Dimmer",
    "setBrightness": "cmnd/patio-lights/Dimmer",
    "getRGB": "stat/patio-lights/Color",
    "setRGB": "cmnd/patio-lights/Color"
  },
  "onValue": "ON",
  "offValue": "OFF",
  "integerValue": true
}
```

Topic names must match [bridge-contract.md](../bridge-contract.md).
The bridge publishes retained `stat/<friendly_name>/POWER`, `Dimmer`, and `Color` topics from the Lepro state it receives on `le/<device_id>/prp/rpt`.
