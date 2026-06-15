# Firmware

Run `lepro-firmware` from `lepro/` (`cd lepro && uv sync`). Examples below use `uv run`.

## Stock download

ZB1 v2.3.18 (tested):

```bash
curl -k -o firmware/3_le_light_zb1_pid_55_v2.3.18.bin \
  "https://ota-dvc-eu-iot.lepro.com/pub/ota/3_le_light_zb1_pid_55_v2.3.18.bin"
```

Or:

```bash
cd lepro
uv run lepro-firmware fetch
```

URL pattern: `https://ota-dvc-eu-iot.lepro.com/pub/ota/<image_name>.bin`

`-k` is needed because Lepro's OTA host certificate may not verify against your system CA store.

Binaries are **not** committed to git (`firmware/.gitignore`).

## Patch CDN pin

The stock firmware embeds a pinned HTTPS trust anchor for Lepro's CDN. Patch it to trust **your** mock CDN cert:

```bash
./deploy/docker/up.sh          # from repo root; generates deploy/docker/certs/cdn.pem
cd lepro
uv run lepro-firmware patch
```

Default output: `../firmware/3_le_light_zb1_pid_55_v2.3.18.patched.bin` (repo-root `firmware/`)

`lepro-firmware patch` always applies **three** edits to the stock image (v2.3.18):

1. **CDN pin** — replace the embedded crt-bundle trust anchor with your lab `cdn.pem`.
2. **prov_mode gate** — NOP the `bgeui a4,0x2` branch in `ota_state_machine_advance`
   state `0x2d` so the cert fetch is no longer gated on onboarding mode.
3. **arm-flag gate** — rewrite the `beq a10,a4` arm-flag compare in
   `prov_cdn_https_fetch_resource` to `beq a10,a10` (always taken) so the HTTPS
   download runs regardless of the per-resource arm sentinel.

Edits 2 and 3 together force root/client certs to be re-downloaded from the CDN on
every provision (and on every WiFi-up), which is what we always want for cert rotation.

Verify without writing:

```bash
cd lepro
uv run lepro-firmware verify
```

Reference manifest: `firmware/3_le_light_zb1_pid_55_v2.3.18.patched.manifest.json`

### Firmware profiles

Version-specific patch data (offsets, stock/patched bytes, bundle detection) lives in
[`firmware/patch-profiles.toml`](../firmware/patch-profiles.toml). The patcher
auto-selects a profile from the firmware filename (glob match) or falls back to
`default_profile`. Override with `--profile zb1_v2_3_18`.

Each profile defines:

- **Metadata** — chip type, stock/patched paths, OTA download URL
- **`[bundle]`** — crt-bundle auto-detection (signature, error string, cert count).
  Replacement bytes come from your lab CDN cert at patch time, not from the config.
- **`[[patches]]`** — static byte edits: `offset`, `stock_hex`, `patched_hex`

#### Adding a new firmware version

1. Fetch the stock `.bin` (`lepro-firmware fetch --profile <new_id>` once the profile exists, or `curl`).
2. RE in Ghidra — locate the crt-bundle site, prov_mode gate, and arm-flag gate (or equivalent).
3. Add a new `[[profiles]]` block to `firmware/patch-profiles.toml` with offsets and hex bytes.
4. Patch and verify:

```bash
cd lepro
uv run lepro-firmware patch -f ../firmware/<new_image>.bin
uv run lepro-firmware verify -f ../firmware/<new_image>.patched.bin
```

5. **Cert geometry** — the lab CDN cert's issuer Name TLV and SPKI DER lengths must match
   the stock bundle entry (`name_len` + `key_len`). If they differ, generate a CDN cert
   with matching geometry or adjust the bundle layout support separately.

### Force cert re-fetch on provision

Stock ZB1 v2.3.18 suppresses the CDN cert download behind **two** independent gates:

1. **prov_mode gate** — `ota_state_machine_advance` only calls
   `prov_cdn_https_fetch_resource` for root/client when `prov_mode` is 2 or 3
   (onboarding). On a normal provisioned bulb (`prov_mode` 1), WiFi-up state `0x2d`
   skips the download even though provision JSON was accepted.
   Patch: NOP the 3-byte `bgeui a4, 2` branch at VA `0x4200a1ca` / file offset
   `0x2a1ca` (`f6 24 1c` → `f0 20 00`).
2. **arm-flag gate** — inside `prov_cdn_https_fetch_resource` each resource is only
   fetched when its retained "armed" sentinel reads `0x5a5a5a5a`. The sentinel is set
   to armed during JSON ingest (state `0x2b`) only when the NVS key is missing or the
   sent path is byte-identical to the stored path, and is cleared to `0xa5a5a5a5`
   after a successful download. So NOPing gate 1 alone is not enough — a re-provision
   can still be skipped when the flag is disarmed.
   Patch: rewrite the `beq a10, a4, download` compare at VA `0x4200b21f` / file offset
   `0x2b21f` to `beq a10, a10, download` (`47 1a 12` → `a7 1a 12`), which always takes
   the download branch regardless of the arm sentinel. One edit covers both root and
   client cert (shared code path).

With both gates neutered the bulb re-downloads root + client certs on every provision
(and on every WiFi-up / boot) — exactly what cert rotation needs.

Patch idempotency: re-running `lepro-firmware patch` detects already-patched gates and
skips the byte edits. ESP image checksum and SHA256 footer are recalculated after any
change.

## BLE OTA flash

Requires prior bond:

```bash
lepro-bond --mac AA:BB:CC:DD:EE:FF
lepro-ota --mac AA:BB:CC:DD:EE:FF \
  --firmware firmware/3_le_light_zb1_pid_55_v2.3.18.patched.bin
```

By default the OTA JSON keeps the firmware filename version (e.g. `2.3.18`).
BLE OTA sends `do_check:0`, which skips the firmware version gate, so same-version
and repeated reflashes work without bumping. Pass `--reflash` to opt into bumping
the patch version (e.g. for cloud/`do_check:1` flashing).

## RE artifacts

- `firmware/asm/` — Ghidra notes for OTA parser
- `uv run python -m lepro.tools.validate_ota_parser` — offline parser checks (requires local `.elf`; see [tools.md](tools.md))
