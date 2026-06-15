#!/usr/bin/env bash
# One-time: register ZeroSSL EAB with cert-manager (free ACME, no Gen-Y chain).
#
# 1. Create account at https://zerossl.com
# 2. Developer → "EAB Credentials for ACME Clients" → Generate
# 3. Run:
#      ZEROSSL_EAB_KID='...' ZEROSSL_EAB_HMAC='...' ./setup-zerossl-eab.sh
#
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config-k3s-home}"

: "${ZEROSSL_EAB_KID:?Set ZEROSSL_EAB_KID from ZeroSSL Developer → EAB}"
: "${ZEROSSL_EAB_HMAC:?Set ZEROSSL_EAB_HMAC from ZeroSSL Developer → EAB}"

kubectl -n cert-manager create secret generic zerossl-eab-lepro \
  --from-literal=secret="$ZEROSSL_EAB_HMAC" \
  --dry-run=client -o yaml | kubectl apply -f -

# keyID is not secret; substitute into issuer manifest.
sed "s/__ZEROSSL_EAB_KID__/${ZEROSSL_EAB_KID}/g" "$DIR/k8s/clusterissuer-zerossl.yaml" | kubectl apply -f -

echo "Waiting for ClusterIssuer zerossl-prod-lepro..."
kubectl wait --for=condition=Ready clusterissuer/zerossl-prod-lepro --timeout=120s
kubectl get clusterissuer zerossl-prod-lepro
