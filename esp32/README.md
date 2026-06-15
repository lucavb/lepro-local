# ESP32 Lepro LP BLE Client (Path B port)

Python reference implementation lives in `lepro/`. Port in this order:

1. `crypto.c` ← `lepro/crypto.py` (derive_aes_key, encrypt_search_hello, encrypt_dp_json, Lepro padding)
2. `protocol.c` ← `lepro/protocol.py` (CRC table, build_packet_opcode, build_packets_fragmented)
3. `frame.c` ← `lepro/frame.py` (RX CRC, fragment reassembly)
4. `commands.c` ← `lepro/commands.py` (DP JSON builders)
5. `session.c` ← `lepro/session.py` (discovery → auth → dpValue → commit; **no bond**)
6. `main.c` ← NimBLE GATT client, load creds from NVS

See also [`docs/esp32.md`](../docs/esp32.md).

## Provisioning

After bonding a factory-reset light:

```bash
uv run lepro-bond --mac 10:20:BA:31:B2:BA --export-provision --format header -o esp32/include/lepro_creds.h
uv run lepro-bond --mac 10:20:BA:31:B2:BA --export-provision --format nvs -o esp32/nvs/lepro_creds.csv
```

Copy `esp32/include/lepro_creds.h.example` for the expected layout, then replace placeholders with output from `lepro-bond`.

Flash `lepro_creds.h` or import the NVS CSV via `idf.py nvs-partition-gen`.

## Session flow (every control command)

Same as the Lepro app — see `lepro/apk_analysis.py` `CONTROL_SESSION_ORDER`.

Bond tokens are pre-flashed; ESP32 never calls `requestBond`.
