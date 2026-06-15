#!/usr/bin/env python3
"""Unit tests for BLE OTA inner buffer layout and packet builders."""

from __future__ import annotations

import json
import struct
import unittest

from lepro.crypto import decrypt_cbc, derive_session_key, MAIN_IV
from lepro.ota import (
    OTA_ERR_MD5,
    OTA_ERR_NO_CTX,
    OTA_ERR_OTA_END,
    OTA_ERR_SET_BOOT,
    OTA_INNER_SIZE,
    OTA_MAX_CHUNK,
    OTA_SUCCESS,
    _device_matches,
    build_ota_chunk_inner,
    build_ota_chunk_plaintext,
    build_ota_chunk_packets,
    build_ota_start_inner,
    build_ota_start_json,
    build_ota_start_packets,
    build_ota_start_plaintext,
    bump_ota_version,
    classify_ota_verdict,
    describe_ota_data_error,
    encrypt_ota_inner,
    prepare_ota_image,
    firmware_cdn_path,
    md5_hex,
    patch_discovery_version,
    resolve_ota_version,
    version_from_discovery,
)
from lepro.protocol import (
    MAGIC_SINGLE_ENC,
    OP_OTA_DATA,
    OP_OTA_START,
    crc16_lepro,
    opcode_to_wire,
    parse_packet,
    verify_packet,
)

MAC = "10:20:BA:31:B2:BA"
SESSION_RAND = 0x67458B6B


class OtaInnerBufferTests(unittest.TestCase):
    def test_patch_discovery_version_updates_hardcoded_semver(self) -> None:
        image = bytearray(
            b"\x00\x00\x00\x00"
            + bytes.fromhex("0c288244420c388244431c28624441824444624445")
            + b"\x00\x00\x00\x00"
        )
        changed = patch_discovery_version(image, "2.3.19")
        self.assertEqual(changed, 1)
        self.assertIn(bytes.fromhex("1c38624441824444624445"), bytes(image))

    def test_start_inner_layout(self) -> None:
        js = b'{"version":"2.3.18"}'
        inner = build_ota_start_inner(js)
        self.assertEqual(len(inner), OTA_INNER_SIZE)
        self.assertEqual(inner[:8], b"\x00" * 8)
        self.assertEqual(inner[8 : 8 + len(js)], js)
        self.assertEqual(inner[8 + len(js) :], b"\x00" * (OTA_INNER_SIZE - 8 - len(js)))

    def test_start_plaintext_is_nul_terminated_json(self) -> None:
        js = b'{"version":"2.3.18"}'
        self.assertEqual(build_ota_start_plaintext(js), js + b"\x00")

    def test_chunk_inner_layout(self) -> None:
        chunk = b"\xe9\x06" + b"\xab" * 100
        inner = build_ota_chunk_inner(4096, chunk)
        self.assertEqual(len(inner), OTA_INNER_SIZE)
        self.assertEqual(struct.unpack(">I", inner[0:4])[0], 4096)
        self.assertEqual(struct.unpack(">H", inner[4:6])[0], crc16_lepro(chunk))
        self.assertEqual(struct.unpack(">H", inner[6:8])[0], len(chunk))
        self.assertEqual(inner[8 : 8 + len(chunk)], chunk)

    def test_chunk_plaintext_is_compact_header_plus_data(self) -> None:
        chunk = b"\xe9\x06" + b"\xab" * 100
        plain = build_ota_chunk_plaintext(4096, chunk)
        self.assertEqual(len(plain), 8 + len(chunk))
        self.assertEqual(struct.unpack(">I", plain[0:4])[0], 4096)
        self.assertEqual(struct.unpack(">H", plain[4:6])[0], crc16_lepro(chunk))
        self.assertEqual(struct.unpack(">H", plain[6:8])[0], len(chunk))
        self.assertEqual(plain[8:], chunk)

    def test_encrypt_round_trip(self) -> None:
        plain = build_ota_chunk_plaintext(0, b"\xe9" * 64)
        ct = encrypt_ota_inner(MAC, SESSION_RAND, plain)
        key = derive_session_key(MAC, SESSION_RAND)
        self.assertEqual(decrypt_cbc(ct, key, iv=MAIN_IV), plain)

    def test_start_packets_crc_and_opcode(self) -> None:
        js = b'{"x":1}'
        packets, next_seq = build_ota_start_packets(
            MAC, 1, js, session_rand=SESSION_RAND, max_payload=128
        )
        self.assertEqual(len(packets), 1)
        self.assertEqual(next_seq, 2)
        self.assertEqual(packets[0][3], MAGIC_SINGLE_ENC)
        ptype, subtype = opcode_to_wire(OP_OTA_START)
        self.assertEqual((packets[-1][6], packets[-1][7]), (ptype, subtype))
        for pkt in packets:
            self.assertTrue(verify_packet(pkt))
        header = parse_packet(packets[0])
        self.assertIsNotNone(header)
        key = derive_session_key(MAC, SESSION_RAND)
        self.assertEqual(decrypt_cbc(header.payload, key, iv=MAIN_IV), js + b"\x00")

    def test_chunk_packets_opcode(self) -> None:
        chunk = b"\xe9" * 200
        packets, _ = build_ota_chunk_packets(
            MAC, 0, 0, chunk, session_rand=SESSION_RAND
        )
        ptype, subtype = opcode_to_wire(OP_OTA_DATA)
        self.assertEqual((packets[-1][6], packets[-1][7]), (ptype, subtype))
        for pkt in packets:
            self.assertTrue(verify_packet(pkt))
        self.assertEqual(len(build_ota_chunk_plaintext(0, chunk)), 8 + len(chunk))

    def test_bump_ota_version(self) -> None:
        self.assertEqual(bump_ota_version("2.3.18"), "2.3.19")
        self.assertEqual(bump_ota_version("2.2.13"), "2.2.14")

    def test_explicit_ota_version_is_not_bumped_again(self) -> None:
        from pathlib import Path

        self.assertEqual(
            resolve_ota_version(
                Path("firmware/3_le_light_zb1_pid_55_v2.3.18.bin"),
                version="2.3.20",
                reflash=True,
            ),
            "2.3.20",
        )

    def test_json_builder(self) -> None:
        from pathlib import Path

        fw = Path("firmware/3_le_light_zb1_pid_55_v2.3.18.bin")
        if not fw.is_file():
            self.skipTest("firmware bin not present")
        image = fw.read_bytes()
        raw = build_ota_start_json(fw, image, version="2.3.18")
        obj = json.loads(raw)
        self.assertEqual(obj["version"], "2.3.18")
        self.assertEqual(obj["size"], str(fw.stat().st_size))
        self.assertEqual(obj["hash"], md5_hex(fw.read_bytes()))
        self.assertEqual(obj["path"], firmware_cdn_path(fw))
        self.assertEqual(obj["secret"], "")
        self.assertNotIn("fwType", obj)
        self.assertEqual(obj["do_check"], 0)

    def test_prepare_ota_image_default_keeps_version(self) -> None:
        from pathlib import Path

        fw = Path("firmware/3_le_light_zb1_pid_55_v2.3.18.bin")
        if not fw.is_file():
            self.skipTest("firmware bin not present")
        on_disk = fw.read_bytes()
        image, ver = prepare_ota_image(fw, reflash=False)
        self.assertEqual(ver, "2.3.18")
        self.assertEqual(image, on_disk)

    def test_json_builder_reflash_bump(self) -> None:
        from pathlib import Path

        fw = Path("firmware/3_le_light_zb1_pid_55_v2.3.18.bin")
        if not fw.is_file():
            self.skipTest("firmware bin not present")
        image, ver = prepare_ota_image(fw, reflash=True)
        obj = json.loads(build_ota_start_json(fw, image, version=ver))
        self.assertEqual(obj["version"], "2.3.19")
        self.assertEqual(
            obj["path"],
            "pub/ota/3_le_light_zb1_pid_55_v2.3.18.bin",
        )
        self.assertEqual(obj["hash"], md5_hex(image))
        self.assertEqual(obj["size"], str(len(image)))
        anchor = bytes.fromhex("624441824444624445")
        anchor_off = image.find(anchor)
        self.assertNotEqual(anchor_off, -1)
        self.assertEqual(image.find(anchor, anchor_off + 1), -1)
        self.assertEqual(image[anchor_off - 2 : anchor_off], bytes.fromhex("1c38"))

    def test_cdn_path_can_follow_bumped_ota_version(self) -> None:
        from pathlib import Path

        fw = Path("firmware/3_le_light_zb1_pid_55_v2.3.18.bin")
        self.assertEqual(
            firmware_cdn_path(fw, version="2.3.20"),
            "pub/ota/3_le_light_zb1_pid_55_v2.3.20.bin",
        )

    def test_json_builder_explicit_path_version(self) -> None:
        from pathlib import Path

        fw = Path("firmware/3_le_light_zb1_pid_55_v2.3.18.bin")
        image = b"\xe9test"
        obj = json.loads(
            build_ota_start_json(fw, image, version="2.3.20", path_version="2.3.20")
        )
        self.assertEqual(
            obj["path"],
            "pub/ota/3_le_light_zb1_pid_55_v2.3.20.bin",
        )

    def test_max_chunk(self) -> None:
        big = b"x" * (OTA_MAX_CHUNK + 1)
        with self.assertRaises(ValueError):
            build_ota_chunk_inner(0, big)
        with self.assertRaises(ValueError):
            build_ota_chunk_plaintext(0, big)


class OtaErrorDescriptionTests(unittest.TestCase):
    def test_describe_finalize_errors(self) -> None:
        self.assertEqual(describe_ota_data_error(OTA_SUCCESS), "success")
        self.assertIn("MD5", describe_ota_data_error(OTA_ERR_MD5))
        self.assertIn("esp_ota_end", describe_ota_data_error(OTA_ERR_OTA_END))
        self.assertIn("set_boot", describe_ota_data_error(OTA_ERR_SET_BOOT))
        self.assertIn("context", describe_ota_data_error(OTA_ERR_NO_CTX))


class OtaVerdictTests(unittest.TestCase):
    def test_device_matches_lp_name(self) -> None:
        class Dev:
            address = "unknown-uuid"
            name = "LP"

        self.assertTrue(_device_matches("10:20:BA:31:B2:BA", Dev()))

    def test_device_matches_mac(self) -> None:
        class Dev:
            address = "10:20:BA:31:B2:BA"
            name = ""

        self.assertTrue(_device_matches("10:20:ba:31:b2:ba", Dev()))

    def test_version_from_discovery_patch_byte(self) -> None:
        # ZB1 discovery tail: 01 02 03 <patch>
        raw = bytes.fromhex(
            "0006070600000037d7ba0d5b00000296000001fdaf"
            "000000000000000000000000000000000000000000"
            "000000000000000000000000000000000001020312010000"
        )
        self.assertEqual(version_from_discovery(raw), "2.3.18")

    def test_classify_stuck(self) -> None:
        v = classify_ota_verdict("2.3.19", "2.3.19", reconnect_ok=True)
        self.assertEqual(v.status, "stuck")

    def test_classify_rolled_back(self) -> None:
        v = classify_ota_verdict("2.3.19", "2.3.18", reconnect_ok=True)
        self.assertEqual(v.status, "rolled_back")

    def test_classify_reconnect_failed(self) -> None:
        v = classify_ota_verdict("2.3.19", None, reconnect_ok=False)
        self.assertEqual(v.status, "reconnect_failed")

    def test_classify_unknown_version(self) -> None:
        v = classify_ota_verdict("2.3.19", None, reconnect_ok=True)
        self.assertEqual(v.status, "unknown_version")


if __name__ == "__main__":
    unittest.main()
