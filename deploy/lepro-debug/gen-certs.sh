#!/usr/bin/env bash
# Generate Lepro mock PKI for local MQTT hijack (mTLS + HTTPS cert CDN).
#
# HTTPS CDN cert mimics ota-dvc-eu-iot.lepro.com (self-signed, firmware embeds *.lepro.com):
#   RSA 2048, sha256, CN=*.lepro.com, O=LEPRO INNOVATION INC, CA:TRUE, ~30y validity
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
OUT="$DIR/certs"
CN_MQTT="${CN_MQTT:-mqtt.example.home}"
DID="${DID:-debug}"

mkdir -p "$OUT"

if [[ -f "$OUT/ca.key" && "${FORCE:-0}" != "1" ]]; then
  echo "Certs already exist in $OUT (set FORCE=1 to regenerate)"
  exit 0
fi

# Mock Amazon IoT root — bulb downloads this for MQTT mTLS trust.
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout "$OUT/ca.key" -out "$OUT/ca.pem" -days 3650 \
  -subj "/CN=Lepro Mock CA/O=lepro-debug"

openssl req -newkey rsa:2048 -nodes \
  -keyout "$OUT/server.key" -out "$OUT/server.csr" \
  -subj "/CN=$CN_MQTT/O=lepro-debug"

openssl x509 -req -in "$OUT/server.csr" \
  -CA "$OUT/ca.pem" -CAkey "$OUT/ca.key" -CAcreateserial \
  -out "$OUT/server.pem" -days 825 \
  -extfile <(printf "subjectAltName=DNS:%s\n" "$CN_MQTT")

openssl req -newkey rsa:2048 -nodes \
  -keyout "$OUT/client.key" -out "$OUT/client.csr" \
  -subj "/CN=lepro-device-$DID/O=lepro-debug"

openssl x509 -req -in "$OUT/client.csr" \
  -CA "$OUT/ca.pem" -CAkey "$OUT/ca.key" -CAcreateserial \
  -out "$OUT/client.pem" -days 825

cat "$OUT/client.pem" "$OUT/client.key" > "$OUT/client-bundle.pem"
cat "$OUT/server.pem" "$OUT/ca.pem" > "$OUT/server-fullchain.pem"

# HTTPS CDN — match real Lepro OTA server cert parameters closely.
# The live cert is a long-lived self-signed CA-style cert used directly as the
# server leaf. Keep the fake host in SAN, but avoid extra key-usage constraints
# that the real cert does not carry.
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout "$OUT/cdn.key" -out "$OUT/cdn.pem" -days 10950 -sha256 \
  -subj "/C=US/ST=NV/L=Default City/O=LEPRO INNOVATION INC/CN=*.lepro.com" \
  -addext "basicConstraints=critical,CA:TRUE" \
  -addext "subjectKeyIdentifier=hash" \
  -addext "authorityKeyIdentifier=keyid,issuer" \
  -addext "subjectAltName=DNS:dvc-eu-iot.example.home,DNS:*.lepro.com"

echo "Wrote mock PKI to $OUT"
echo "  cdn.pem  — HTTPS (Lepro *.lepro.com mimic, RSA-2048/sha256/CA:TRUE)"
echo "  ca.pem   — mock AmazonRootCA13 for MQTT"
