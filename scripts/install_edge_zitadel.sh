#!/usr/bin/env bash
set -euo pipefail

# Config (provide paths/passwords via env or edit below)
: "${DOMAIN:=edge.2gc.ru}"
: "${NAMESPACE:=zitadel}"
: "${TLS_SECRET:=edge-2gc-ru-tls}"
: "${FULLCHAIN_PATH:?Set FULLCHAIN_PATH to fullchain.pem path}"
: "${PRIVKEY_PATH:?Set PRIVKEY_PATH to privkey.pem path}"
: "${PG_SUPERPASS:?Set PG_SUPERPASS to Postgres superuser password}"
: "${ZITADEL_DB_USER:=zitadel}"
: "${ZITADEL_DB_NAME:=zitadel}"
: "${ZITADEL_DB_PASS:?Set ZITADEL_DB_PASS to ZITADEL DB user password}"
: "${MASTERKEY_SECRET:=zitadel-masterkey}"
: "${MASTERKEY_VALUE:?Set MASTERKEY_VALUE to strong base64/32+ string}"

echo "[1/8] Verifying tools..."
command -v kubectl >/dev/null || { echo "kubectl not found"; exit 1; }
command -v helm >/dev/null || { echo "helm not found"; exit 1; }
command -v minikube >/dev/null || { echo "minikube not found"; exit 1; }

echo "[2/8] Ensuring namespace ${NAMESPACE} exists..."
kubectl get ns "${NAMESPACE}" >/dev/null 2>&1 || kubectl create ns "${NAMESPACE}"

echo "[3/8] Creating/updating TLS secret ${TLS_SECRET} in ${NAMESPACE}..."
kubectl create secret tls "${TLS_SECRET}" -n "${NAMESPACE}" \
  --cert="${FULLCHAIN_PATH}" --key="${PRIVKEY_PATH}" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "[4/8] Installing/Upgrading PostgreSQL (Bitnami OCI) ..."
set +e
helm status pg -n "${NAMESPACE}" >/dev/null 2>&1
has_pg=$?
set -e
if [ "$has_pg" -ne 0 ]; then
  helm install pg oci://registry-1.docker.io/bitnamicharts/postgresql -n "${NAMESPACE}" \
    --set auth.postgresPassword="${PG_SUPERPASS}" \
    --set auth.username="${ZITADEL_DB_USER}" \
    --set auth.password="${ZITADEL_DB_PASS}" \
    --set auth.database="${ZITADEL_DB_NAME}"
else
  echo "PostgreSQL already installed, skipping install."
fi

echo "Waiting for PostgreSQL pod to be Ready..."
kubectl rollout status statefulset/pg-postgresql -n "${NAMESPACE}" --timeout=5m

echo "[5/8] Creating ZITADEL masterkey secret if missing..."
if ! kubectl get secret "${MASTERKEY_SECRET}" -n "${NAMESPACE}" >/dev/null 2>&1; then
  kubectl create secret generic "${MASTERKEY_SECRET}" -n "${NAMESPACE}" \
    --from-literal=masterkey="${MASTERKEY_VALUE}"
else
  echo "Secret ${MASTERKEY_SECRET} exists."
fi

echo "[6/8] Preparing values for ZITADEL..."
VALUES_FILE="/home/ubuntu/cloudbridge-relay-installer/k8s/zitadel/values.edge.yaml"
tmp_values=$(mktemp)
cp "${VALUES_FILE}" "${tmp_values}"

# Inject DB password
sed -i "s#Password: \"\"#Password: \"${ZITADEL_DB_PASS}\"#" "${tmp_values}"

echo "[7/8] Installing/Upgrading ZITADEL Helm chart..."
set +e
helm status zitadel -n "${NAMESPACE}" >/dev/null 2>&1
has_zit=$?
set -e
CHART_REF="oci://ghcr.io/zitadel/charts/zitadel"
if [ "$has_zit" -ne 0 ]; then
  helm install zitadel "${CHART_REF}" -n "${NAMESPACE}" -f "${tmp_values}" \
    --set zitadel.masterkeySecretName="${MASTERKEY_SECRET}" \
    --set zitadel.configmapConfig.ExternalDomain="${DOMAIN}"
else
  helm upgrade zitadel "${CHART_REF}" -n "${NAMESPACE}" -f "${tmp_values}" \
    --set zitadel.masterkeySecretName="${MASTERKEY_SECRET}" \
    --set zitadel.configmapConfig.ExternalDomain="${DOMAIN}"
fi

echo "Applying Ingress manifest..."
kubectl apply -f /home/ubuntu/cloudbridge-relay-installer/k8s/zitadel/ingress/ingress.yaml

echo "Waiting for ZITADEL setup jobs to complete..."
kubectl wait --for=condition=complete --timeout=10m -n "${NAMESPACE}" job/zitadel-init || true
kubectl wait --for=condition=complete --timeout=10m -n "${NAMESPACE}" job/zitadel-setup || true

echo "Waiting for ZITADEL deployment to be Ready..."
kubectl rollout status deploy/zitadel -n "${NAMESPACE}" --timeout=10m || true

echo "[8/8] Verifications"
echo "- Pods:"
kubectl get pods -n "${NAMESPACE}"
echo "- Jobs:"
kubectl get jobs -n "${NAMESPACE}"
echo "- Services:"
kubectl get svc -n "${NAMESPACE}"
echo "- Ingress:"
kubectl get ingress -n "${NAMESPACE}"

echo "- Healthz:" && curl -sk https://"${DOMAIN}"/debug/healthz || true
echo
echo "- Console HEAD:" && curl -skI https://"${DOMAIN}"/ui/console/ || true
echo
echo "- TLS cert (openssl):"
openssl s_client -connect "${DOMAIN}":443 -servername "${DOMAIN}" -showcerts </dev/null 2>/dev/null | openssl x509 -noout -subject -issuer -dates || true

echo "Done."


