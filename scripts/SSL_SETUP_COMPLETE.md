# **SSL Certificate Setup - COMPLETED!**

## **SSL СЕРТИФИКАТ НАСТРОЕН УСПЕШНО**

### ** Что было сделано:**

1. **Найден существующий сертификат** для домена `edge.2gc.ru`
2. **Обновлена HAProxy конфигурация** для использования правильного домена
3. **Настроена маршрутизация** через `edge.2gc.ru` вместо `zitadel.2gc.ru`
4. **Обновлены записи в /etc/hosts** для правильного разрешения домена
5. **Проверена работоспособность** SSL соединения

### **SSL Сертификат:**

- **Домен:** `edge.2gc.ru`
- **Issuer:** Let's Encrypt
- **Valid From:** Sep 8 07:45:48 2025 GMT
- **Valid Until:** Dec 7 07:45:47 2025 GMT
- **Status:** Valid and working
- **Certificate Path:** `/etc/haproxy/certs/edge.2gc.ru.pem`

### **Обновленный доступ к Zitadel:**

#### **HTTPS Access (SSL Enabled)**
- **URL:** https://edge.2gc.ru
- **UI Login:** https://edge.2gc.ru/ui/login
- **OIDC Discovery:** https://edge.2gc.ru/.well-known/openid-configuration

#### **HTTP Access (Redirects to HTTPS)**
- **URL:** http://edge.2gc.ru
- **Auto-redirect:** HTTP → HTTPS (301)

### **Учетные данные для входа:**

- **Email:** `admin@local.test`
- **Password:** `jrcCaigzmqzHpRcxo61RDiS+xejxX4/K`

### **Проверка SSL:**

#### **1. Health Check (SSL):**
```bash
curl https://edge.2gc.ru/healthz
# Expected: {"status":"SERVING"}
```

#### **2. OIDC Discovery (SSL):**
```bash
curl https://edge.2gc.ru/.well-known/openid-configuration
# Expected: JSON with OIDC configuration
```

#### **3. UI Access (SSL):**
```bash
curl -I https://edge.2gc.ru/ui/login
# Expected: HTTP 301 redirect to /ui/login/
```

#### **4. SSL Certificate Verification:**
```bash
openssl s_client -connect edge.2gc.ru:443 -servername edge.2gc.ru < /dev/null 2>/dev/null | openssl x509 -noout -subject -dates
# Expected: subject=CN = edge.2gc.ru, valid dates
```

### ** Техническая конфигурация:**

#### **HAProxy Configuration:**
- **Frontend:** `fe_https_443_zitadel` (HTTPS на порту 443)
- **Backend:** `be_zitadel` (NodePort 192.168.49.2:30001)
- **Host Header:** Автоматически устанавливается `zitadel.local`
- **SSL Certificate:** `/etc/haproxy/certs/edge.2gc.ru.pem`

#### **Domain Routing:**
- **External Domain:** `edge.2gc.ru` (SSL enabled)
- **Internal Domain:** `zitadel.local` (for Zitadel instance)
- **Host Header:** Automatically set to `zitadel.local` by HAProxy

### ** Следующие шаги:**

#### **1. КРИТИЧНО: Создать проект в Zitadel UI**
1. Откройте браузер: https://edge.2gc.ru/ui/login
2. Войдите с учетными данными выше
3. Создайте проект `cloudbridge-project`
4. Создайте OIDC приложение `cloudbridge-relay`
5. Получите Client ID и Client Secret

#### **2. ВЫСОКИЙ: Обновить конфигурацию CloudBridge**
После получения credentials обновить:
- `configs/config.json`
- `k8s/configmaps/cloudbridge-config.yaml`

#### **3. СРЕДНИЙ: Протестировать интеграцию**
- Перезапустить CloudBridge Relay
- Протестировать аутентификацию
- Проверить логи

### **Troubleshooting:**

#### **Проблема: "SSL: no alternative certificate subject name matches"**
- Убедитесь, что используете `edge.2gc.ru`, а не `zitadel.2gc.ru`
- Проверьте, что запись в `/etc/hosts` правильная

#### **Проблема: "Connection refused"**
- Проверьте, что HAProxy работает: `sudo systemctl status haproxy`
- Проверьте, что NodePort сервисы работают: `kubectl -n zitadel get services`

#### **Проблема: "Certificate expired"**
- Проверьте срок действия: `openssl x509 -in /etc/haproxy/certs/edge.2gc.ru.pem -noout -dates`
- Обновите сертификат через Let's Encrypt

### ** Управление:**

#### **HAProxy:**
```bash
# Перезагрузить конфигурацию
sudo systemctl reload haproxy

# Проверить статус
sudo systemctl status haproxy

# Проверить конфигурацию
sudo haproxy -c -f /etc/haproxy/haproxy.cfg
```

#### **SSL Certificate:**
```bash
# Проверить срок действия
openssl x509 -in /etc/haproxy/certs/edge.2gc.ru.pem -noout -dates

# Проверить домены
openssl x509 -in /etc/haproxy/certs/edge.2gc.ru.pem -text -noout | grep -E "(Subject:|DNS:)"
```

---

## **РЕЗУЛЬТАТ: SSL сертификат настроен и работает!**

**SSL соединение работает**
**Валидный сертификат Let's Encrypt**
**HTTPS доступ к Zitadel**
**Автоматический redirect HTTP → HTTPS**
**Готов к созданию проекта и приложения**

**Доступ:** https://edge.2gc.ru/ui/login



