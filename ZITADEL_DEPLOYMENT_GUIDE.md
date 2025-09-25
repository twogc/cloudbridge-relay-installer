# Руководство по развертыванию ZITADEL в Minikube

## Содержание
1. [Предварительные требования](#предварительные-требования)
2. [Установка PostgreSQL](#установка-postgresql)
3. [Установка ZITADEL](#установка-zitadel)
4. [Настройка Ingress](#настройка-ingress)
5. [Настройка внешнего доступа](#настройка-внешнего-доступа)
6. [Типичные ошибки и решения](#типичные-ошибки-и-решения)
7. [Проверка работоспособности](#проверка-работоспособности)

## Предварительные требования

### Системные требования
- Ubuntu 20.04+ или аналогичная ОС
- Minikube установлен и запущен
- kubectl настроен
- Helm v3 установлен
- Docker установлен

### Проверка окружения
```bash
# Проверка Minikube
minikube status

# Проверка kubectl
kubectl cluster-info

# Проверка Helm
helm version
```

## Установка PostgreSQL

### 1. Создание namespace
```bash
kubectl create namespace zitadel
```

### 2. Установка PostgreSQL через Helm
```bash
# Добавление репозитория Bitnami (если недоступен, используйте OCI)
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# Установка PostgreSQL через OCI (рекомендуется)
helm install pg oci://registry-1.docker.io/bitnamicharts/postgresql \
  -n zitadel \
  --set auth.postgresPassword="bd94d943179bac151890531e0631cbc7" \
  --set auth.username="zitadel" \
  --set auth.password="1e7b5f2371b958405db20b1c25a9fd86" \
  --set auth.database="zitadel"
```

### 3. Проверка установки PostgreSQL
```bash
# Проверка статуса подов
kubectl get pods -n zitadel

# Проверка подключения к базе данных
kubectl exec -n zitadel pg-postgresql-0 -- bash -c 'PGPASSWORD=bd94d943179bac151890531e0631cbc7 psql -U postgres -c "\l"'
```

## Установка ZITADEL

### 1. Создание masterkey
```bash
# Создание секрета с masterkey (32 байта)
kubectl create secret generic zitadel-masterkey \
  -n zitadel \
  --from-literal=masterkey="$(openssl rand -hex 16)"
```

### 2. Создание values файла
```yaml
# zitadel-values.yaml
zitadel:
  configmapConfig:
    ExternalDomain: "edge.2gc.ru"
    ExternalSecure: true
    Machine:
      Identification:
        Hostname:
          Enabled: true
        Webhook:
          Enabled: false
    FirstInstance:
      Org:
        LoginClient:
          Machine:
            Username: 'login-client'
            Name: 'Automatically Initialized IAM Login Client'
          Pat.ExpirationDate: '2029-01-01T00:00:00Z'
  
  secretConfig:
    Database:
      Postgres:
        Database: zitadel
        Host: pg-postgresql.zitadel.svc.cluster.local
        Port: 5432
        SSL:
          Mode: disable
        User:
          Password: "1e7b5f2371b958405db20b1c25a9fd86"
          Username: zitadel
        Admin:
          Password: "bd94d943179bac151890531e0631cbc7"
          Username: postgres
    TLS:
      Enabled: false
      Mode: external
  
  masterkeySecretName: "zitadel-masterkey"
  
  # Добавление переменной окружения для таймзоны
  extraEnv:
    - name: TZ
      value: "Europe/Moscow"
```

### 3. Установка ZITADEL
```bash
# Добавление репозитория ZITADEL
helm repo add zitadel https://charts.zitadel.com
helm repo update

# Установка ZITADEL
helm install zitadel zitadel/zitadel \
  -n zitadel \
  -f zitadel-values.yaml
```

### 4. Проверка установки
```bash
# Проверка статуса jobs
kubectl get jobs -n zitadel

# Проверка логов init job
kubectl logs -n zitadel job/zitadel-init

# Проверка логов setup job
kubectl logs -n zitadel job/zitadel-setup

# Проверка статуса подов
kubectl get pods -n zitadel
```

## Настройка Ingress

### 1. Создание TLS сертификата
```bash
# Создание самоподписанного сертификата (для тестирования)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /tmp/edge-2gc-ru.key \
  -out /tmp/edge-2gc-ru.crt \
  -subj "/CN=edge.2gc.ru"

# Создание Kubernetes secret
kubectl create secret tls edge-2gc-ru-tls \
  -n zitadel \
  --cert=/tmp/edge-2gc-ru.crt \
  --key=/tmp/edge-2gc-ru.key
```

### 2. Создание Ingress
```yaml
# zitadel-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: zitadel-ingress
  namespace: zitadel
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - edge.2gc.ru
    secretName: edge-2gc-ru-tls
  rules:
  - host: edge.2gc.ru
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: zitadel
            port:
              number: 8080
```

### 3. Применение Ingress
```bash
kubectl apply -f zitadel-ingress.yaml
```

## Настройка внешнего доступа

### 1. Проверка портов Minikube
```bash
# Получение портов Ingress контроллера
kubectl get service -n ingress-nginx ingress-nginx-controller
```

### 2. Настройка HAProxy
```bash
# Обновление конфигурации HAProxy
sudo sed -i 's/30502/30267/g' /etc/haproxy/haproxy.cfg

# Изменение протокола на HTTPS
sudo sed -i 's/server zitadel_ingress 192.168.49.2:30267 check/server zitadel_ingress 192.168.49.2:30267 check ssl verify none/' /etc/haproxy/haproxy.cfg

# Перезагрузка HAProxy
sudo systemctl reload haproxy
```

### 3. Настройка nginx (альтернатива)
```bash
# Создание конфигурации nginx
sudo tee /etc/nginx/sites-enabled/edge-2gc-ru-https.conf << 'EOF'
server {
    listen 443 ssl;
    server_name edge.2gc.ru;
    
    ssl_certificate /etc/nginx/ssl/edge.2gc.ru.crt;
    ssl_certificate_key /etc/nginx/ssl/edge.2gc.ru.key;
    
    location / {
        proxy_pass https://192.168.49.2:30267;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_ssl_verify off;
        proxy_ssl_server_name on;
    }
}
EOF

# Перезагрузка nginx
sudo systemctl reload nginx
```

## Типичные ошибки и решения

### 1. Ошибка подключения к базе данных

**Ошибка:**
```
FATAL: password authentication failed for user "zitadel"
```

**Решение:**
```bash
# Проверка паролей в секретах
kubectl get secret -n zitadel pg-postgresql -o yaml

# Проверка подключения к PostgreSQL
kubectl exec -n zitadel pg-postgresql-0 -- bash -c 'PGPASSWORD=bd94d943179bac151890531e0631cbc7 psql -U postgres -c "\du"'
```

### 2. Ошибка masterkey

**Ошибка:**
```
masterkey must be 32 bytes, but is 48
```

**Решение:**
```bash
# Удаление старого masterkey
kubectl delete secret zitadel-masterkey -n zitadel

# Создание нового masterkey правильной длины
kubectl create secret generic zitadel-masterkey \
  -n zitadel \
  --from-literal=masterkey="$(openssl rand -hex 16)"
```

### 3. Ошибка инициализации базы данных

**Ошибка:**
```
Did not find any relations.
```

**Решение:**
```bash
# Проверка статуса jobs
kubectl get jobs -n zitadel

# Перезапуск jobs
kubectl delete jobs --all -n zitadel
helm upgrade zitadel zitadel/zitadel -n zitadel -f zitadel-values.yaml
```

### 4. Ошибка 503 Service Unavailable

**Ошибка:**
```
503 Service Unavailable
```

**Решение:**
```bash
# Проверка статуса подов
kubectl get pods -n zitadel

# Проверка логов Ingress контроллера
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx

# Проверка портов Minikube
kubectl get service -n ingress-nginx ingress-nginx-controller
```

### 5. Ошибка "The plain HTTP request was sent to HTTPS port"

**Ошибка:**
```
400 Bad Request: The plain HTTP request was sent to HTTPS port
```

**Решение:**
```bash
# Изменение конфигурации HAProxy на HTTPS
sudo sed -i 's/server zitadel_ingress 192.168.49.2:30267 check/server zitadel_ingress 192.168.49.2:30267 check ssl verify none/' /etc/haproxy/haproxy.cfg

# Перезагрузка HAProxy
sudo systemctl reload haproxy
```

### 6. Ошибка постоянного редиректа

**Ошибка:**
```
Постоянный редирект между /ui/console/ и /ui/login/
```

**Решение:**
```bash
# Удаление конфликтующих Ingress
kubectl delete ingress zitadel-login-ingress -n zitadel

# Перезапуск подов ZITADEL
kubectl rollout restart deployment/zitadel -n zitadel
kubectl rollout restart deployment/zitadel-login -n zitadel

# Очистка кэша браузера (Ctrl+F5)
```

### 7. Ошибка доступа к репозиториям Helm

**Ошибка:**
```
Error: INSTALLATION FAILED: failed to fetch https://github.com/zitadel/zitadel-charts/releases/download/zitadel-9.5.1/zitadel-9.5.1.tgz : 503 Service Unavailable
```

**Решение:**
```bash
# Использование OCI репозитория
helm install zitadel oci://ghcr.io/zitadel/charts/zitadel \
  -n zitadel \
  -f zitadel-values.yaml

# Или использование локального chart
helm pull zitadel/zitadel --untar
helm install zitadel ./zitadel -n zitadel -f zitadel-values.yaml
```

### 8. Ошибка портов в Minikube

**Ошибка:**
```
Порты изменились после перезапуска Minikube
```

**Решение:**
```bash
# Проверка текущих портов
kubectl get service -n ingress-nginx ingress-nginx-controller

# Обновление конфигурации HAProxy/nginx
sudo sed -i 's/старый_порт/новый_порт/g' /etc/haproxy/haproxy.cfg
sudo systemctl reload haproxy
```

## Проверка работоспособности

### 1. Проверка здоровья ZITADEL
```bash
# Проверка health check
curl -k https://edge.2gc.ru/debug/healthz

# Ожидаемый результат: "ok"
```

### 2. Проверка веб-интерфейса
```bash
# Проверка консоли
curl -k https://edge.2gc.ru/ui/console/ -I

# Ожидаемый результат: HTTP/1.1 200 OK
```

### 3. Проверка базы данных
```bash
# Проверка таблиц в базе данных
kubectl exec -n zitadel pg-postgresql-0 -- bash -c 'PGPASSWORD=bd94d943179bac151890531e0631cbc7 psql -U postgres -d zitadel -c "\dt"'
```

### 4. Проверка домена в базе данных
```bash
# Проверка настроек домена
kubectl exec -n zitadel pg-postgresql-0 -- bash -c 'PGPASSWORD=bd94d943179bac151890531e0631cbc7 psql -U postgres -d zitadel -c "SELECT instance_id, domain, is_primary FROM projections.instance_domains;"'
```

## Полезные команды

### Мониторинг
```bash
# Просмотр логов ZITADEL
kubectl logs -n zitadel deployment/zitadel -f

# Просмотр логов PostgreSQL
kubectl logs -n zitadel pg-postgresql-0 -f

# Просмотр событий
kubectl get events -n zitadel --sort-by='.lastTimestamp'
```

### Отладка
```bash
# Проверка конфигурации ZITADEL
kubectl get configmap -n zitadel zitadel-config -o yaml

# Проверка секретов
kubectl get secret -n zitadel -o yaml

# Проверка Ingress
kubectl describe ingress zitadel-ingress -n zitadel
```

### Очистка
```bash
# Удаление ZITADEL
helm uninstall zitadel -n zitadel

# Удаление PostgreSQL
helm uninstall pg -n zitadel

# Удаление namespace
kubectl delete namespace zitadel
```

## Заключение

Этот документ покрывает основные аспекты развертывания ZITADEL в Minikube. При возникновении проблем:

1. Проверьте логи подов и jobs
2. Убедитесь в правильности конфигурации
3. Проверьте подключение к базе данных
4. Убедитесь в правильности настроек Ingress и внешнего доступа

Для получения дополнительной помощи обратитесь к [официальной документации ZITADEL](https://zitadel.com/docs/self-hosting/deploy/kubernetes).
