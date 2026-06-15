from __future__ import annotations

import json
from pathlib import Path

from lepro.paths import repo_path
from lepro.provision import resolve_cert_paths, save_cert_paths


def _store_path(tmp_path: Path, mac: str) -> Path:
    slug = mac.replace(":", "-").lower()
    path = tmp_path / ".lepro" / f"{slug}-prov-paths.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    return path


def test_resolve_cert_paths_uses_stored_when_paths_differ(tmp_path: Path, monkeypatch) -> None:
    mac = "10:20:BA:31:B2:BA"
    store = _store_path(tmp_path, mac)
    store.write_text(
        json.dumps(
            {
                "root": "pub/cert/AmazonRootCA13.pem",
                "cert": "device/cert/did/debug/ver/1/sign/lepro",
            }
        )
    )

    def fake_home() -> Path:
        return tmp_path

    monkeypatch.setattr("lepro.provision.Path.home", fake_home)

    root, cert = resolve_cert_paths(
        mac,
        "pub/cert/AmazonRootCA13.pem",
        "device/cert/did/debug/ver/2/sign/lepro",
    )
    assert root == "pub/cert/AmazonRootCA13.pem"
    assert cert == "device/cert/did/debug/ver/1/sign/lepro"


def test_resolve_cert_paths_update_flag_uses_requested(tmp_path: Path, monkeypatch) -> None:
    mac = "10:20:BA:31:B2:BA"
    store = _store_path(tmp_path, mac)
    store.write_text(
        json.dumps(
            {
                "root": "pub/cert/AmazonRootCA13.pem",
                "cert": "device/cert/did/debug/ver/1/sign/lepro",
            }
        )
    )

    monkeypatch.setattr("lepro.provision.Path.home", lambda: tmp_path)

    new_cert = "device/cert/did/debug/ver/2/sign/lepro"
    root, cert = resolve_cert_paths(
        mac,
        "pub/cert/AmazonRootCA13.pem",
        new_cert,
        update_cert_paths=True,
    )
    assert cert == new_cert


def test_save_cert_paths_round_trip(tmp_path: Path, monkeypatch) -> None:
    mac = "10:20:BA:31:B2:BA"
    monkeypatch.setattr("lepro.provision.Path.home", lambda: tmp_path)

    save_cert_paths(mac, "pub/root.pem", "device/cert/path")
    root, cert = resolve_cert_paths(mac, "other", "other")
    assert (root, cert) == ("pub/root.pem", "device/cert/path")
