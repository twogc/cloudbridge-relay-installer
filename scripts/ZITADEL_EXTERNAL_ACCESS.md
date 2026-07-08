# **Zitadel External Access Setup - COMPLETED!**

## **НАСТРОЙКА ЗАВЕРШЕНА УСПЕШНО**

### ** Что было настроено:**

1. **UFW Firewall**: Открыты порты 80 и 443
2. **HAProxy Configuration**: Настроен reverse proxy для Zitadel
3. **NodePort Services**: Созданы Kubernetes NodePort сервисы
4. **Domain Routing**: Настроена маршрутизация по доменам
5. **Host Header**: Исправлен Host заголовок для Zitadel
6. **SSL Certificate**: Настроен валидный SSL сертификат для edge.2gc.ru

### **Внешний доступ к Zitadel:**

#### **HTTPS Access (Recommended)**
- **URL:** https://edge.2gc.ru
- **UI Login:** https://edge.2gc.ru/ui/login
- **OIDC Discovery:** https://edge.2gc.ru/.well-known/openid-configuration

#### **HTTP Access (Redirects to HTTPS)**
- **URL:** http://edge.2gc.ru
- **Auto-redirect:** HTTP → HTTPS (301)

### **Учетные данные для входа:**

- **Email:** `admin@local.test`
- **Password:** `jrcCaigzmqzHpRcxo61RDiS+xejxX4/K`

### **OIDC Endpoints:**

```json
{
  "issuer": "https://zitadel.local",
  "authorization_endpoint": "https://zitadel.local/oauth/v2/authorize",
  "token_endpoint": "https://zitadel.local/oauth/v2/token",
  "userinfo_endpoint": "https://zitadel.local/oidc/v1/userinfo",
  "jwks_uri": "https://zitadel.local/oauth/v2/keys",
  "scopes_supported": ["openid", "profile", "email", "phone", "address", "offline_access"],
  "grant_types_supported": ["authorization_code", "implicit", "refresh_token", "client_credentials"]
}
```

### **🛠️ Техническая конфигурация:**

#### **HAProxy Configuration:**
- **Frontend:** `fe_https_443_zitadel` (HTTPS на порту 443)
- **Backend:** `be_zitadel` (NodePort 192.168.49.2:30001)
- **Host Header:** Автоматически устанавливается `zitadel.local`
- **SSL Certificate:** `/etc/haproxy/certs/edge.2gc.ru.pem`

#### **SSL Certificate:**
- **Domain:** `edge.2gc.ru`
- **Issuer:** Let's Encrypt
- **Valid From:** Sep 8 07:45:48 2025 GMT
- **Valid Until:** Dec 7 07:45:47 2025 GMT
- **Certificate Path:** `/etc/haproxy/certs/edge.2gc.ru.pem`
- **Private Key:** Included in PEM file
- **Status:** Valid and working

#### **Kubernetes Services:**
- **Zitadel API:** `zitadel-nodeport` (NodePort 30001)
- **Zitadel Login:** `zitadel-login-nodeport` (NodePort 30000)
- **Selector:** `app.kubernetes.io/component=start`

#### **UFW Rules:**
```bash
80/tcp                     ALLOW       Anywhere
443/tcp                    ALLOW       Anywhere
```

### **Проверка работоспособности:**

#### **1. Health Check:**
```bash
curl https://edge.2gc.ru/healthz
# Expected: {"status":"SERVING"}
```

#### **2. OIDC Discovery:**
```bash
curl https://edge.2gc.ru/.well-known/openid-configuration
# Expected: JSON with OIDC configuration
```

#### **3. UI Access:**
```bash
curl -I https://edge.2gc.ru/ui/login
# Expected: HTTP 301 redirect to /ui/login/
```

#### **4. SSL Certificate:**
```bash
openssl s_client -connect edge.2gc.ru:443 -servername edge.2gc.ru < /dev/null 2>/dev/null | openssl x509 -noout -subject -dates
# Expected: subject=CN = edge.2gc.ru, valid dates
```

### **📝 Следующие шаги:**

#### **1. КРИТИЧНО: Создать проект в Zitadel UI**
1. Откройте браузер: https://edge.2gc.ru/ui/login
2. Войдите с учетными данными выше
3. Создайте проект `cloudbridge-project`
4. Создайте OIDC приложение `cloudbridge-relay`
5. Получите Client ID и Client Secret

#### **2.  ВЫСОКИЙ: Обновить конфигурацию CloudBridge**
После получения credentials обновить:
- `configs/config.json`
- `k8s/configmaps/cloudbridge-config.yaml`

#### **3. СРЕДНИЙ: Протестировать интеграцию**
- Перезапустить CloudBridge Relay
- Протестировать аутентификацию
- Проверить логи

### **🛠️ Управление сервисами:**

#### **HAProxy:**
```bash
# Перезагрузить конфигурацию
sudo systemctl reload haproxy

# Проверить статус
sudo systemctl status haproxy

# Проверить конфигурацию
sudo haproxy -c -f /etc/haproxy/haproxy.cfg
```

#### **Kubernetes Services:**
```bash
# Проверить NodePort сервисы
kubectl -n zitadel get services

# Проверить поды
kubectl -n zitadel get pods

# Проверить логи
kubectl -n zitadel logs -l app.kubernetes.io/name=zitadel
```

#### **UFW:**
```bash
# Проверить статус
sudo ufw status

# Добавить правило (если нужно)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
```

### ** Troubleshooting:**

#### **Проблема: "Instance not found"**
- Убедитесь, что Host заголовок установлен правильно
- Проверьте, что Zitadel настроен на домен `zitadel.local`

#### **Проблема: Connection refused**
- Проверьте, что NodePort сервисы работают
- Проверьте, что поды Zitadel готовы (Ready 1/1)

#### **Проблема: SSL Certificate errors**
- Проверьте, что сертификат существует: `/etc/haproxy/certs/edge.2gc.ru.pem`
- Проверьте срок действия сертификата: `openssl x509 -in /etc/haproxy/certs/edge.2gc.ru.pem -noout -dates`
- Убедитесь, что используете правильный домен: `edge.2gc.ru` (не `zitadel.2gc.ru`)

### **Мониторинг:**

#### **HAProxy Stats:**
```bash
# Проверить статистику HAProxy
curl -k https://zitadel.2gc.ru/stats
```

#### **Kubernetes Health:**
```bash
# Проверить здоровье подов
kubectl -n zitadel get pods -o wide

# Проверить события
kubectl -n zitadel get events
```

---

## **РЕЗУЛЬТАТ: Zitadel доступен извне через HTTPS!**

**Внешний доступ настроен и работает**
**OIDC endpoints доступны**
**UI доступен через браузер**
**Готов к созданию проекта и приложения**

**Доступ:** https://edge.2gc.ru/ui/login
