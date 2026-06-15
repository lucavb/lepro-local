# FUN_42008eb0 — return-3 branch map (v2.3.18)

Static trace for why BLE `0x1010` maps parser failure to **`0x112`**:

```asm
4200958d: call8 0x42008eb0
42009590: mov.n a6, a10          ; saves pre-parser a10 (not parser ret)
42009597: movi  a2, 0x112
4200959a: beqi  a6, 0x3, 0x420095c0
```

Parser contract: **`a2` = cJSON object**, **`a4` = OTA ctx** (global; see caller `@ 0x4200956d`).
Returns **`a2 = 0`** success, **`a2 = 3`** failure.

---

## Caller: flat BLE vs envelope

| Path | Call site | When |
|------|-----------|------|
| **Flat (BLE `0x1010`)** | `0x4200958d` | `FUN_4200886c` → tail `0x4200956c` after NUL-length check |
| **Envelope** | `0x4200913c`, `0x42009152` | `FUN_420090e0` only; reached from `FUN_42009668` (`0x42000a5c`) — cloud/MQTT style, **not** the BLE opcode path |
| Alt preload | `FUN_42009520` | Copies parsed string fields into ctx, calls `0x42009054` — different entry |

**BLE pre-check before parser** (`FUN_4200886c`):

```asm
4200886f: l32i.n a4, a2, 0x4      ; payload char*
42008873: callx8 strlen
42008879: l32i.n a3, a2, 0x8      ; declared length
4200887b: addi.n a10, a10, 1
4200887d: beq a3, a10, ok         ; require len == strlen(json)+1 (NUL included)
```

If this fails → early error (`a2=1` response path), **never reaches** `FUN_42008eb0`.

**Ctx byte 0** written once before parse (`0x42009573`: `s32i.n a5, a4, 0`) from the register **`a5` on entry to `0x4200956c`** (tail-call from `0x420088b4`). That byte feeds the version gate below.

---

## Decision tree (in order)

```
a2 == NULL?  ──yes──► 0x42008f60 → return 3
        │
        ▼
fwType GetObjectItem
        │
   missing ──► 0x42009728 (ctx singleton) ──null──► 0x42008f60
        │                              └──ok──► continue @ do_check
        │
   present ──► IsString? ──no──► ec8 (singleton path)
        │                      └──yes──► validate_string_len(max=16)
        │                               ret==0 ──► ec8 (singleton) ──► continue
        │                               ret!=0 ──► 0x42008ef2 → 0x42008f60
        │
do_check (optional number; default 1 if absent / not number)
        │
version (required string)
        │
   missing / not string ──► a5=0 ──► (later) counter fail @ 0x42009045
        │
   present ──► store ctx+0x24
        │
   do_check == 0? ──yes──► a5=1, skip version gate
        │
   do_check != 0:
        ctx[0] == 0? ──yes──► 0x42008f58 → global err=4 → return 3
        strcmp(ctx, version) == 0? ──yes──► same (return 3)   ; "same version"
        strcmp != 0 ──► continue, a5=1
        │
hash (required string) ──► counter++
path (required string) ──► counter++
size (required string, max 10) ──► counter++   ; JSON number fails IsString
secret (required string; "" OK)
        │
   missing ──► log "ota info error" / "le_ota" @ 0x42009008 → return 3
   not string ──► 0x42009008 → return 3
        │
a5 == 4? ──no──► 0x42009008 → return 3
        │
yes ──► return 0
```

**Counter model:** baseline `1` (from `do_check` path or forced `1` when `do_check==0`) plus **`hash`**, **`path`**, **`size`** only. **`version` does not increment.**

---

## Per-field validators (fwType / version / secret)

### fwType (`0x420009ac`)

1. `cJSON_GetObjectItem` — may be absent; absent uses singleton ctx (`0x42009728` → `0x3fca3b08`).
2. If present: `cJSON_IsString` (`0x420b855c`).
3. `validate_string_len` (`0x400014f4`, **max 16**, `a11=0`).
4. Return handling is inverted-looking but correct:
   - **`extui` low byte == 0** → take **`0x42008ec8`** singleton detour, then continue.
   - **low byte != 0** → **`0x42008ef2` → return 3**.

### version (`0x420009b4`)

1. Required string (`IsString`).
2. Stored at **ctx+0x24** (replaces prior ptr; old freed).
3. **Hidden gate when `do_check != 0`** (`0x42008f43`–`0x42008f5d`):

```asm
42008f43: beqz.n a5, 0x42008f6a        ; do_check==0 → skip gate, a5=1
42008f45: l8ui   a5, a4, 0x0           ; ctx[0] (fw_version prefix / flag)
42008f48: beqz.n a5, 0x42008f58        ; FAIL if zero
42008f4a: mov.n  a11, a10              ; JSON version string
42008f4c: mov.n  a10, a4               ; ctx base as C string
42008f4e: callx8 strcmp (0x40001230)
42008f56: bnez.n a10, 0x42008f6c       ; continue only if != 0
42008f58: global[0]=4; return 3
```

So with **`do_check: 1`**: need **non-zero ctx[0]** and **JSON `version` must differ** from the running version string at ctx base (same-version flash rejected here — separate from ESP app-descriptor check in `FUN_420097a8` which logs `"running version same as new"`).

With **`do_check: 0`**: entire strcmp block skipped.

### secret (`0x420009c4`)

1. `GetObjectItem` — **must exist** (key present).
2. `IsString` — required; non-string → `0x42009008`.
3. Value may be **`""`**.
4. Copied via `0x4200cf38` (strdup-style).

---

## Most likely `0x0112` causes for current experiment JSON

Given `tmp/ota-info-v2-experiment.json` (all fields present, `size` string, `secret: ""`):

| Priority | Branch | Condition |
|----------|--------|-----------|
| 1 | `0x42008f48` / `0x42008f58` | `do_check: 1` and **ctx[0]==0** (nothing valid in ctx base before parse) |
| 2 | `0x42008f56` → `0x42008f58` | `strcmp(ctx, "2.3.19")==0` (version matches running — unlikely if bumping) |
| 3 | `0x42009045` | counter `a5 != 4` (e.g. missing/invalid hash/path/size typing) |
| 4 | `0x4200887d` | BLE length metadata ≠ `strlen(JSON)+1` (outside parser) |
| 5 | `0x42009005` | `secret` key absent (not the case for `""`) |

**Not the BLE path:** envelope `[{"action":"fwUpgrade","params":{...}}]` (`FUN_420090e0`).

---

## Static “which branch” without UART

Ghidra MCP: set breakpoint on `0x42008f60`, `0x42008ef2`, `0x42008f58`, `0x42009008`, `0x42009045` and run under QEMU / chip debugger.

Without hardware, emulate the counter + gates in Python from parsed JSON + assumed ctx prefix (see `uv run python -m lepro.tools.validate_ota_parser --simulate`).
