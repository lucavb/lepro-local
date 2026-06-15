# Protocol reference

## BLE

- GATT service and opcodes: `lepro/protocol.py`
- Session flow (discovery → auth → dpValue): `lepro/session.py`
- Bond tokens: `~/.lepro/<mac>.json` via `lepro-bond`
- DP builders: `lepro/commands.py`, `lepro/modes.py`

## Native MQTT (after provision)

Bulb **subscribes** (publish commands here):

| Topic | Purpose |
|-------|---------|
| `le/<id>/prp/set` | Set datapoints (primary control) |
| `le/<id>/prp/get` | Request state (request format unknown) |
| `le/<id>/act/exe` | Actions (format unknown) |
| `le/<id>/prf/get` | Profile |

Bulb **publishes**:

| Topic | Purpose |
|-------|---------|
| `le/<id>/prp/rpt` | State reports |
| `le/<id>/prf/rpt` | Profile |
| `le/<id>/log/rpt` | Logs |

### Payload envelope

```json
{"d":{"d1":1,"d2":2,"d3":1000}}
```

BLE sends the inner object encrypted; MQTT wraps it in `{"d":{…}}` over TLS.

### Datapoints (ZB1)

| DP | Meaning | Values |
|----|---------|--------|
| `d1` | power | `0` off / `1` on |
| `d2` | work mode | `0` white CCT, `1` solid, `2` scene/RGB-IC, `3` music |
| `d3` | brightness | `10`–`1000` |
| `d4` | color temp (white) | `0`–`1000` |
| `d5` | solid color HSV hex | `"HHHHSSSSVVVV"` |
| `d50` | RGB-IC scene string | per-segment; see `modes.py` |
| `d52` | scene brightness | `100`–`1000` |
| `d53` | bulb count | ZB1 default `15` |
| `d30` | constant in reports | device-specific; read-only |

### Examples

```bash
ID=3619294555
CA=deploy/docker/certs/ca.pem

mosquitto_pub -h mqtt.example.home -p 8883 --cafile $CA --insecure \
  -t "le/$ID/prp/set" -m '{"d":{"d1":1}}'

mosquitto_pub ... -m '{"d":{"d1":1,"d2":1,"d3":1000,"d5":"000003E803E8"}}'   # red
```

## mTLS limitation

The bulb stores the downloaded client **cert** but signs TLS with a factory `mqtt_key_client` that is never overwritten by the CDN bundle. Mock mTLS therefore fails with alert 51; lab Mosquitto allows anonymous after server-auth TLS. See [mqtt-setup.md](mqtt-setup.md).

## Latency

MQTT command → bulb state report is ~2s (firmware loop), not limited by tooling.
