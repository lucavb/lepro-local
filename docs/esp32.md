# ESP32 BLE client (skeleton)

Export bond credentials from a PC and flash them on an ESP32 for **bond-less** BLE control (the ESP never calls `requestBond`).

Full ESP-IDF port is **not** included — only a skeleton header and export path.

## Export creds

After `lepro-bond` on a factory-reset or bonded light:

```bash
lepro-bond --mac AA:BB:CC:DD:EE:FF --export-provision \
  --format header -o esp32/include/lepro_creds.h
```

NVS CSV format:

```bash
lepro-bond --mac AA:BB:CC:DD:EE:FF --export-provision \
  --format nvs -o esp32/nvs/lepro_creds.csv
```

## Session flow

Same as the Lepro app: discovery → auth → dpValue → commit. See `lepro/session.py` and `esp32/include/lepro_session.h`.

## Suggested port order

1. `crypto.c` ← `lepro/crypto.py`
2. `protocol.c` ← `lepro/protocol.py`
3. `frame.c` ← `lepro/frame.py`
4. `commands.c` ← `lepro/commands.py`
5. `session.c` ← `lepro/session.py`
6. `main.c` — NimBLE GATT client, load creds from NVS

See also [esp32/README.md](../esp32/README.md).
