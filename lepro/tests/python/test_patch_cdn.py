#!/usr/bin/env python3
"""Unit tests for firmware CDN pin and profile-driven byte patches."""

from __future__ import annotations

import tempfile
import unittest
from pathlib import Path

from lepro.firmware_profile import load_profiles, resolve_profile
from lepro.patch_cdn import apply_byte_patch, apply_byte_patches, check_byte_patches, find_bundle_layout
from lepro.paths import repo_path

MINI_PROFILE_TOML = """\
default_profile = "test"

[[profiles]]
id = "test"
description = "synthetic test profile"
chip = "esp32s3"
image_glob = "test_fw_*.bin"
stock_image = "firmware/test.bin"
patched_image = "firmware/test.patched.bin"
ota_url = "https://example.com/test.bin"

[profiles.bundle]
signature_hex = "deadbeef"
attach_error = "bundle-err"
lookback = 64
expected_cert_count = 1

[[profiles.patches]]
name = "gate_a"
offset = 43
stock_hex = "f6241c"
patched_hex = "f02000"
"""


class BytePatchTests(unittest.TestCase):
    def setUp(self) -> None:
        self.profile = load_profiles().by_id("zb1_v2_3_18")
        self.prov_patch = self.profile.patches[0]
        self.arm_patch = self.profile.patches[1]

    def test_apply_replaces_stock_with_patched(self) -> None:
        fw = bytearray(b"\x00" * 0x30000)
        off = self.prov_patch.offset
        fw[off : off + len(self.prov_patch.stock)] = self.prov_patch.stock
        result = apply_byte_patch(fw, self.prov_patch)
        self.assertTrue(result.changed)
        self.assertEqual(result.before, self.prov_patch.stock)
        self.assertEqual(result.after, self.prov_patch.patched)
        self.assertEqual(bytes(fw[off : off + len(self.prov_patch.patched)]), self.prov_patch.patched)

    def test_apply_is_idempotent(self) -> None:
        fw = bytearray(b"\x00" * 0x30000)
        off = self.prov_patch.offset
        fw[off : off + len(self.prov_patch.stock)] = self.prov_patch.stock
        apply_byte_patch(fw, self.prov_patch)
        result = apply_byte_patch(fw, self.prov_patch)
        self.assertFalse(result.changed)

    def test_apply_fails_on_unexpected_bytes(self) -> None:
        fw = bytearray(b"\x00" * 0x30000)
        off = self.prov_patch.offset
        fw[off : off + 3] = b"\xde\xad\xbe"
        with self.assertRaises(SystemExit):
            apply_byte_patch(fw, self.prov_patch)

    def test_check_byte_patch_requires_patched_bytes(self) -> None:
        fw = bytearray(b"\x00" * 0x30000)
        off = self.prov_patch.offset
        fw[off : off + len(self.prov_patch.stock)] = self.prov_patch.stock
        with self.assertRaises(SystemExit):
            check_byte_patches(fw, [self.prov_patch])

    def test_stock_firmware_patches(self) -> None:
        fw_path = self.profile.stock_image
        if not fw_path.is_file():
            self.skipTest("stock firmware bin not present")
        fw = bytearray(fw_path.read_bytes())
        results = apply_byte_patches(fw, self.profile.patches)
        self.assertEqual([r.name for r in results], ["prov_mode gate", "arm-flag gate"])
        self.assertTrue(all(r.changed for r in results))
        for patch, result in zip(self.profile.patches, results, strict=True):
            off = patch.offset
            self.assertEqual(bytes(fw[off : off + len(patch.patched)]), patch.patched)
        again = apply_byte_patches(fw, self.profile.patches)
        self.assertFalse(any(r.changed for r in again))


class ProfileConfigTests(unittest.TestCase):
    def test_default_profile_loads(self) -> None:
        profiles = load_profiles(repo_path("firmware/patch-profiles.toml"))
        self.assertEqual(profiles.default_profile, "zb1_v2_3_18")
        profile = profiles.by_id("zb1_v2_3_18")
        self.assertEqual(len(profile.patches), 2)
        self.assertEqual(profile.patches[0].offset, 0x2A1CA)
        self.assertEqual(profile.patches[1].offset, 0x2B21F)

    def test_resolve_by_filename_glob(self) -> None:
        profiles = load_profiles(repo_path("firmware/patch-profiles.toml"))
        profile = resolve_profile(
            profiles,
            firmware_path=Path("3_le_light_zb1_pid_55_v2.3.18.patched.bin"),
            profile_id=None,
        )
        self.assertEqual(profile.id, "zb1_v2_3_18")

    def test_resolve_by_id_override(self) -> None:
        with tempfile.NamedTemporaryFile("w", suffix=".toml", delete=False) as tmp:
            tmp.write(MINI_PROFILE_TOML)
            path = Path(tmp.name)
        try:
            profiles = load_profiles(path)
            profile = resolve_profile(profiles, profile_id="test")
            self.assertEqual(profile.id, "test")
            profile = resolve_profile(
                profiles, firmware_path=Path("test_fw_v9.bin"), profile_id=None
            )
            self.assertEqual(profile.id, "test")
        finally:
            path.unlink(missing_ok=True)

    def test_bundle_layout_on_stock_firmware(self) -> None:
        profile = load_profiles().by_id("zb1_v2_3_18")
        fw_path = profile.stock_image
        if not fw_path.is_file():
            self.skipTest("stock firmware bin not present")
        layout = find_bundle_layout(fw_path.read_bytes(), profile.bundle)
        self.assertEqual(layout.bundle_off, 0x9FD3)
        self.assertEqual(layout.entry_off, 0x9FDA)


if __name__ == "__main__":
    unittest.main()
