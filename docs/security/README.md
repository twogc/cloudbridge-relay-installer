# Безопасность CloudBridge Relay

## Обзор

Документы по безопасности: аутентификация/авторизация, TLS, zero-trust, compliance.  
**Canonical IAM: [Zitadel](../guides/ZITADEL_SETUP.md)** (Keycloak удалён — не восстанавливать).

> Порты: [openwiki/port-scheme.md](../../openwiki/port-scheme.md).  
> Deploy: [openwiki/deploy.md](../../openwiki/deploy.md).

## Документы

| Документ | Описание |
|----------|----------|
| [SECURITY.md](SECURITY.md) | Security policy, reporting, supported versions |
| [ENTERPRISE_SECURITY_COMPLETE.md](ENTERPRISE_SECURITY_COMPLETE.md) | Enterprise security features |
| [guides/ZITADEL_SETUP.md](../guides/ZITADEL_SETUP.md) | Zitadel integration setup |
| [guides/ZITADEL_DEPLOYMENT_GUIDE.md](../guides/ZITADEL_DEPLOYMENT_GUIDE.md) | Zitadel K8s deployment |
| [guides/ZITADEL_SERVICE_USER_PERMISSIONS.md](../guides/ZITADEL_SERVICE_USER_PERMISSIONS.md) | Service user permissions |

Устаревшие отчёты Keycloak (если встречались во внешних копиях) — архив; в этом репозитории **нет** `KEYCLOAK_*.md` в `docs/security/`. Исторический контекст: `docs/legacy/`, миссия `cloudbridge-relay-fixes`.

## Модель безопасности

### Zero-Trust

- Проверка каждого соединения
- Минимальные привилегии
- Постоянная верификация (JWT / mTLS где применимо)

### Аутентификация (canonical)

- **Zitadel** — OIDC IdP, JWKS (RS256)
- **JWT** — access tokens для API / clients
- **OAuth 2.0 / OIDC** — standard flows

### Авторизация

- **RBAC** — Role-Based Access Control (Zitadel roles / project grants)
- Tenant isolation — network + policy controller
- Политики доступа на уровне tenant / peer

## Компоненты

### Zitadel

- Server / instance: per-environment (см. `k8s-manifests/09–15`, guides)
- JWT validation via JWKS in `internal/auth/`, `internal/zitadel/`
- Не использовать Keycloak env (`KEYCLOAK_*`) — только Zitadel config

### JWT валидация

- Issuer / audience checks
- Clock skew tolerance
- Signature verification (JWKS)

### TLS / QUIC

- TLS 1.2+ / 1.3 where configured
- QUIC built-in crypto (UDP 9090 / P2P 5553)
- Certificate validation for HTTPS API (5553) and MASQUE (8443)

## Защита от угроз

- Rate limiting, circuit breaker, connection limits
- STUN / TURN / ICE для NAT traversal
- Audit logs, security metrics, anomaly detection (AI optional)

## Конфигурация (пример env)

```bash
# Zitadel (names vary by config key — prefer configs/config.json zitadel section)
ZITADEL_DOMAIN=https://auth.example.com
ZITADEL_PROJECT_ID=...
# Client credentials / machine user — see ZITADEL_SERVICE_USER_PERMISSIONS.md
```

Не задавать `KEYCLOAK_*` в production.

## Диагностика

```bash
# OIDC discovery (подставьте ваш Zitadel domain)
curl -sS "https://auth.example.com/.well-known/openid-configuration" | head

# JWT decode (локально, без отправки секретов)
jwt decode "$ACCESS_TOKEN"

# Auth logs on relay
kubectl logs -f deployment/cloudbridge-relay -n cloudbridge | grep -i auth

# Zitadel (not keycloak)
kubectl logs -f deployment/zitadel -n cloudbridge
```

## Чек-лист

### Аутентификация

- [ ] Zitadel доступен (OIDC discovery)
- [ ] JWKS reachable from relay
- [ ] JWT audience/issuer совпадают с config
- [ ] Clients могут получить токен

### Авторизация

- [ ] Роли / project grants назначены
- [ ] Tenant isolation проверена
- [ ] Admin paths защищены

### Мониторинг

- [ ] Audit logs
- [ ] Metrics scrape (`metrics.port` **5551** на relay)
- [ ] Alerts on auth failures

---

**SoT IAM:** Zitadel. Security policy details: [SECURITY.md](SECURITY.md).
