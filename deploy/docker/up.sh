#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
CERT_SRC="$DIR/../lepro-debug"
CERT_DST="$DIR/certs"

mkdir -p "$CERT_DST"
CN_MQTT="${CN_MQTT:-mqtt.example.home}" \
CDN_SAN="${CDN_SAN:-dvc-eu-iot.example.home}" \
FORCE="${FORCE:-0}" \
  "$CERT_SRC/gen-certs.sh"

# gen-certs writes to lepro-debug/certs; symlink-copy for compose mounts
if [[ "$CERT_SRC/certs" != "$CERT_DST" ]]; then
  rm -rf "$CERT_DST"
  mkdir -p "$CERT_DST"
  cp -a "$CERT_SRC/certs/." "$CERT_DST/"
fi

docker compose -f "$DIR/docker-compose.yml" up -d

echo
echo "Lab stack running:"
echo "  MQTT TLS:  mqtts://\${mqtt_host}:8883  (set in lepro-lab.toml)"
echo "  MQTT plain: mqtt://127.0.0.1:1883"
echo "  CDN HTTPS: https://\${cdn_host}/pub/cert/AmazonRootCA13.pem"
echo
echo "Point DNS or /etc/hosts for cdn_host and mqtt_host to this host."
echo "Then: lepro-firmware patch && lepro-ota ... && lepro-provision ..."
