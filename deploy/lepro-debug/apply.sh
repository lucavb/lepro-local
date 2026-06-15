#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config-k3s-home}"

"$DIR/gen-certs.sh"

kubectl apply -f "$DIR/k8s/namespace.yaml"
kubectl apply -f "$DIR/k8s/configmap-nginx.yaml"
kubectl apply -f "$DIR/k8s/configmap-mosquitto.yaml"

kubectl -n lepro-debug create secret generic lepro-mock-pki \
  --from-file=ca.pem="$DIR/certs/ca.pem" \
  --from-file=server.pem="$DIR/certs/server.pem" \
  --from-file=server.key="$DIR/certs/server.key" \
  --from-file=client.pem="$DIR/certs/client.pem" \
  --from-file=client.key="$DIR/certs/client.key" \
  --from-file=client-bundle.pem="$DIR/certs/client-bundle.pem" \
  --from-file=server-fullchain.pem="$DIR/certs/server-fullchain.pem" \
  --from-file=cdn.pem="$DIR/certs/cdn.pem" \
  --from-file=cdn.key="$DIR/certs/cdn.key" \
  --dry-run=client -o yaml | kubectl apply -f -

# Traefik on .20 terminates TLS for dvc-eu-iot only (does not touch default wildcard).
kubectl -n lepro-debug create secret tls lepro-cdn-tls \
  --cert="$DIR/certs/cdn.pem" \
  --key="$DIR/certs/cdn.key" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl apply -f "$DIR/k8s/deployment-nginx.yaml"
kubectl apply -f "$DIR/k8s/service-nginx.yaml"
kubectl apply -f "$DIR/k8s/ingress.yaml"
kubectl apply -f "$DIR/k8s/deployment-mosquitto.yaml"
kubectl apply -f "$DIR/k8s/service-mosquitto.yaml"
kubectl apply -f "$DIR/k8s/deployment-lepro-bridge.yaml"
kubectl apply -f "$DIR/k8s/service-cert-cdn-https.yaml"

# Both consumers read their certs from the lepro-mock-pki secret at startup only,
# so they must be restarted after a cert (re)generation or they keep stale certs
# in memory. Mosquitto especially: a stale broker cert signed by a previous CA key
# makes the bulb reject the handshake (TLS alert 48 unknown_ca).
kubectl -n lepro-debug rollout restart deployment/lepro-cert-cdn
kubectl -n lepro-debug rollout restart deployment/mosquitto
kubectl -n lepro-debug rollout restart deployment/lepro-bridge
kubectl -n lepro-debug rollout status deployment/lepro-cert-cdn --timeout=120s
kubectl -n lepro-debug rollout status deployment/mosquitto --timeout=120s
kubectl -n lepro-debug rollout status deployment/lepro-bridge --timeout=120s

echo
echo "HTTPS cert CDN: https://dvc-eu-iot.example.home/pub/cert/AmazonRootCA13.pem"
echo "  TLS: self-signed Lepro mimic (CN=*.lepro.com, RSA-2048, CA:TRUE)"
echo "MQTT broker:    mqtts://mqtt.example.home:8883  (LB IP 10.0.0.5)"
echo
kubectl -n lepro-debug get ingress,svc,pods
