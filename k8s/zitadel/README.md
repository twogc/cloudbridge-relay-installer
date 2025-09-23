# ZITADEL в Minikube (PostgreSQL, Ingress, TLS)

## Предусловия
- Minikube с включённым ingress: `minikube addons enable ingress`
- CPU=8, RAM≈20Gi, Disk≥80Gi
- kubectl, helm

## Установка
```bash
# Namespace
kubectl create namespace zitadel || true

# TLS (self-signed)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout zitadel.local.key -out zitadel.local.crt \
  -subj "/CN=zitadel.local/O=Local"
kubectl -n zitadel create secret tls zitadel-tls \
  --cert=zitadel.local.crt --key=zitadel.local.key --dry-run=client -o yaml | kubectl apply -f -

# PostgreSQL
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
kubectl -n zitadel create secret generic pg-auth \
  --from-literal=postgres-password=$(openssl rand -hex 16) \
  --from-literal=password=$(openssl rand -hex 16) \
  --dry-run=client -o yaml | kubectl apply -f -
helm upgrade --install pg bitnami/postgresql -n zitadel -f k8s/zitadel/pg-values.yaml

# ZITADEL
helm repo add zitadel https://charts.zitadel.com
helm repo update
kubectl -n zitadel create secret generic zitadel-admin \
  --from-literal=email=admin@local.test \
  --from-literal=password=$(openssl rand -base64 24) \
  --dry-run=client -o yaml | kubectl apply -f -
helm upgrade --install zitadel zitadel/zitadel -n zitadel -f k8s/zitadel/values.yaml

# Ingress
kubectl apply -f k8s/zitadel/ingress/ingress.yaml

# hosts (на хост-машине)
echo "$(minikube ip)  zitadel.local" | sudo tee -a /etc/hosts
```

## Проверка
```bash
kubectl -n zitadel get pods
curl -k https://zitadel.local/.well-known/openid-configuration | jq .
```

## Экспорт OIDC параметров
```bash
bash k8s/zitadel/export-oidc.sh
```

## Бэкап (CronJob)
```bash
kubectl apply -f k8s/zitadel/backup/pvc.yaml
kubectl apply -f k8s/zitadel/backup/cronjob.yaml
```

## Ресурсы
- ZITADEL: requests cpu=500m mem=2Gi; limits cpu=2 mem=4Gi
- PostgreSQL: requests cpu=500m mem=2Gi; limits cpu=2 mem=4Gi


