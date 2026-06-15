# Docker lab stack

Local Mosquitto + mock cert CDN for the optional MQTT path.

## Start

```bash
./deploy/docker/up.sh
```

Generates PKI via `../lepro-debug/gen-certs.sh` (into `deploy/docker/certs/`) and starts:

| Service | Port | Role |
|---------|------|------|
| mosquitto | 1883, 8883 | Plain + TLS MQTT |
| nginx | 443 | HTTPS cert CDN for bulb provision |

## DNS

Edit `lepro-lab.toml` with hostnames you control, then resolve them to this machine:

```
10.0.0.5  dvc-eu-iot.example.home mqtt.example.home
```

## Mosquitto auth

TLS listener uses relaxed auth (`require_certificate false`) because the bulb cannot complete mutual TLS with a static mock client bundle. See [docs/protocol.md](../../docs/protocol.md).

## Regenerate certs

```bash
FORCE=1 ./deploy/lepro-debug/gen-certs.sh
./deploy/docker/up.sh
```

Rebond and reprovision the bulb after rotation.
