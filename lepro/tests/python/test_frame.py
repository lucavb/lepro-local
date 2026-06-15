#!/usr/bin/env python3
"""Offline unit tests for lepro.frame against captures.log."""

from __future__ import annotations

import struct
import unittest

from lepro.frame import (
    DecryptedFrame,
    RxAssembler,
    load_capture_packets,
    parse_bond_result,
    parse_ota_data_resp,
    parse_ota_start_resp,
)
from lepro.paths import repo_path
from lepro.protocol import OP_OTA_DATA_RESP, OP_OTA_START_RESP
from lepro.protocol import OP_BOND_RESP, OP_DP_RESP, OP_SEARCH, OP_SEARCH_RESP, parse_packet, verify_packet


MAC = "10:20:BA:31:B2:BA"


class FrameTests(unittest.TestCase):
    def test_crc_majority_pass(self) -> None:
        cap = repo_path("captures.log")
        if not cap.is_file():
            self.skipTest("captures.log not present")
        packets = load_capture_packets(cap)
        ok = sum(1 for p in packets if verify_packet(p))
        self.assertGreater(ok, len(packets) * 0.9)

    def test_parse_hello(self) -> None:
        hello = bytes.fromhex(
            "8a015a50000010000020c4012a79a0dd031c727eea6cf7c736dc45a764086a2bd4ce293aea2f0264a166"
        )
        hdr = parse_packet(hello)
        assert hdr is not None
        self.assertEqual(hdr.opcode, OP_SEARCH)
        self.assertEqual(hdr.length, 32)

    def test_assembler_device_info(self) -> None:
        raw = bytes.fromhex(
            "274b5a50000010010050634c835b5edd9103db61d6cda526c02e701babc664e158c05a9a8a7233"
            "ebb703838a31fcd170bee566b6a62d27f806cc3fa4beb896f0a22807125cbe236cc6b6be7674fe079fd63f6562c782cc749c7b"
        )
        asm = RxAssembler(MAC, session_rand=0x67458B6B)
        frames = asm.feed(raw)
        self.assertEqual(len(frames), 1)
        self.assertEqual(frames[0].opcode, OP_SEARCH_RESP)
        self.assertTrue(frames[0].decrypted)
        self.assertEqual(len(frames[0].payload), 72)

    def test_assembler_light_ack(self) -> None:
        # After CMD_HS, light ACK uses wire 0x10/0x03 (OP_BOND_RESP), not 0x1103
        raw = bytes.fromhex(
            "19035a50000110030010a3a27817fb8b866af077db02350603f0"
        )
        asm = RxAssembler(MAC)
        frames = asm.feed(raw)
        self.assertEqual(frames[0].opcode, OP_BOND_RESP)
        self.assertEqual(len(frames[0].payload), 16)

    def test_cert_fragment_reassembly(self) -> None:
        cert = bytes.fromhex(
            "09905a500003200000b0389cb28d7f2d6b04a17b4992039c506b13890a2d2155cd47480a663cbce79a8511d1d93c07f488db1688a6895eb41fba3c87d64de2c017b033d2295f541b99baa25f947bcaa0704434c033b3a74db323a1397072c8055e0b42c77be06ade1954fdc3877b8c786eaa1ba79709ad037d43d1a8c1cda712c1f8e081a8276267a65dcceac4591ce8204caca4d90839de321807b425981bc11e71099073c605c9bdf502c93b66366be8e663734a165bcf1786"
        )
        asm = RxAssembler(MAC)
        frames = asm.feed(cert)
        self.assertEqual(len(frames), 1)
        self.assertEqual(frames[0].opcode, 0x2000)
        self.assertEqual(len(frames[0].payload), 176)


class OtaParseTests(unittest.TestCase):
    def test_ota_start_resp(self) -> None:
        payload = b"\x00" * 4 + b"\x00\x05\x00\x00" + b"\x00" * 6 + struct.pack("<I", 8192)
        f = DecryptedFrame(seq=1, opcode=OP_OTA_START_RESP, payload=payload, raw=b"")
        p = parse_ota_start_resp(f)
        assert p is not None
        self.assertEqual(p.ack_sn, 5)
        self.assertEqual(p.code, 0)
        self.assertEqual(p.offset, 8192)

    def test_ota_start_resp_compact_error(self) -> None:
        payload = b"\x00\x00\x01\x12" + b"\x00" * 4
        f = DecryptedFrame(seq=1, opcode=OP_OTA_START_RESP, payload=payload, raw=b"")
        p = parse_ota_start_resp(f)
        assert p is not None
        self.assertEqual(p.ack_sn, 0)
        self.assertEqual(p.code, 0x112)
        self.assertEqual(p.offset, 0)

    def test_ota_data_resp(self) -> None:
        payload = b"\x00" * 4 + b"\x00\x0a\x00\x00"
        f = DecryptedFrame(seq=2, opcode=OP_OTA_DATA_RESP, payload=payload, raw=b"")
        p = parse_ota_data_resp(f)
        assert p is not None
        self.assertEqual(p.ack_sn, 10)
        self.assertEqual(p.result, 0)

    def test_ota_data_resp_compact_result(self) -> None:
        payload = b"\x00\x00\x01\x15"
        f = DecryptedFrame(seq=2, opcode=OP_OTA_DATA_RESP, payload=payload, raw=b"")
        p = parse_ota_data_resp(f)
        assert p is not None
        self.assertEqual(p.ack_sn, 0)
        self.assertEqual(p.result, 0x115)


class BondParseTests(unittest.TestCase):
    def test_bond_result_none_for_wrong_opcode(self) -> None:
        f = DecryptedFrame(seq=0, opcode=OP_SEARCH_RESP, payload=b"\x00" * 4, raw=b"")
        self.assertIsNone(parse_bond_result(f))

    def test_bond_result_big_endian(self) -> None:
        f = DecryptedFrame(seq=0, opcode=OP_BOND_RESP, payload=b"\x00\x00\x00\x07", raw=b"")
        self.assertEqual(parse_bond_result(f), 7)


if __name__ == "__main__":
    unittest.main()
