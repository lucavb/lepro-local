#!/usr/bin/env python3
"""Unit tests for APK-aligned DP JSON builders."""

from lepro.commands import (
    color_hsv,
    dp_json_text,
    power_off,
    power_on,
    status_query,
)


def test_power_on_off_json() -> None:
    assert dp_json_text(power_on()) == '{"d1":1}'
    assert dp_json_text(power_off()) == '{"d1":0}'


def test_power_on_with_brightness() -> None:
    assert dp_json_text(power_on(500)) == '{"d1":1,"d2":1,"d3":500}'


def test_status_query_array() -> None:
    assert dp_json_text(status_query()) == '["d1","d2","d3","d4","d5","d50","d52"]'
    assert dp_json_text(status_query(("d1", "d5"))) == '["d1","d5"]'


def test_color_hsv_flat_keys() -> None:
    payload = color_hsv(120, 500, 800, 1000)
    text = dp_json_text(payload)
    assert text.startswith('{"d1":1,"d2":1,"d3":1000,"d5":"')
    assert '"d":' not in text


if __name__ == "__main__":
    test_power_on_off_json()
    test_power_on_with_brightness()
    test_status_query_array()
    test_color_hsv_flat_keys()
    print("test_commands: PASS")
