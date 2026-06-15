#!/usr/bin/env python3
"""Tests for fancy mode builders."""

from lepro.modes import (
    APK_SCENE_GRADIENT,
    DIY_EFFECTS,
    build_d50_center_solid,
    build_d50_flash,
    build_d50_solid,
    debug_payload,
    diy_solid,
    effect_tail,
    load_preset,
    preset_frames,
    resolve_mode,
    scene_gradient,
)
from lepro.paths import repo_path

PRESET_FIXTURES = repo_path("tests/fixtures/presets")


def test_diy_effects_tails() -> None:
    assert effect_tail("Steady", 50) == "000640000E1"
    tail = effect_tail("Breathe", 50)
    assert tail.startswith("000640000E4")
    assert effect_tail("Leftward", 50).startswith("00164")
    assert effect_tail("Leftward", 50).endswith("E1")


def test_d50_solid_orange() -> None:
    d50 = build_d50_solid("FFAA00", "Steady")
    assert d50 == "N01:P10001FFAA00F210001000FU3V3" + effect_tail("Steady") + ";"


def test_d50_solid_tb1_length() -> None:
    d50 = build_d50_solid("FFAA00", "Steady", bulb_count=196)
    assert "00C4" in d50


def test_flash_d50() -> None:
    d50 = build_d50_flash("FFAA00", speed=50)
    assert d50 == "N01:P10001FFAA00U3F300101E20088;"
    assert "F210001" not in d50
    assert "U3F3" in d50


def test_center_out_d50() -> None:
    d50 = build_d50_center_solid("FFAA00", "CenterOut", speed=50, bulb_count=15)
    assert d50 == (
        "#V:02080F000000000700000000;"
        "#I00:N01:P10001FFAA00F2100010008U3V3002640088E1;"
        "#I01:N01:P10001FFAA00F2100010007U3V3001640088E1;"
    )


def test_center_in_d50() -> None:
    d50 = build_d50_center_solid("FFAA00", "CenterIn", speed=50, bulb_count=15)
    assert "U3V3001640088E1" in d50.split(";")[1]
    assert "U3V3002640088E1" in d50.split(";")[2]


def test_diy_payload_flash() -> None:
    p = diy_solid("FFAA00", "Flash")
    assert p["d1"] == 1
    assert p["d2"] == 2
    assert p["d50"].startswith("N01:")
    assert "U3F300101E2" in p["d50"]


def test_diy_payload_keys() -> None:
    p = diy_solid("FF0000", "Gradient")
    assert p["d1"] == 1
    assert p["d2"] == 2
    assert p["d50"].startswith("N01:")
    assert p["d52"] == 1000


def test_apk_gradient_d6() -> None:
    p = scene_gradient()
    assert p["d6"] == APK_SCENE_GRADIENT
    assert p["d2"] == 2


def test_preset_hulk_frames() -> None:
    preset = load_preset("hulk", presets_dir=PRESET_FIXTURES)
    frames = preset_frames(preset)
    assert len(frames) >= 20
    assert all("d50" in f for f in frames)


def test_debug_payload_diy() -> None:
    mac = "10:20:BA:31:B2:BA"
    info = debug_payload(mac, resolve_mode("diy", color="FFAA00"))
    assert info["opcode"] == "0x1100"
    assert info["json_len"] > 50
    assert sum(info["ciphertext_bytes"]) >= 16


if __name__ == "__main__":
    test_diy_effects_tails()
    test_d50_solid_orange()
    test_d50_solid_tb1_length()
    test_flash_d50()
    test_center_out_d50()
    test_center_in_d50()
    test_diy_payload_flash()
    test_diy_payload_keys()
    test_apk_gradient_d6()
    test_preset_hulk_frames()
    test_debug_payload_diy()
    print("test_modes: PASS")
