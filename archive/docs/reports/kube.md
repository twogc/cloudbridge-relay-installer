## План разработки: kubectl proxy integration (специфическая логика)

Цель: дать пользователю kubectl доступ к кластерам через CloudBridge Relay на 443/HTTPS (QUIC/MASQUE-friendly), с RBAC-контролем, аудитом и без необходимости прямой видимости API-сервера.

### 1) Архитектура (v1 — прозрачный reverse-proxy)
- Точка входа: `https://relay/kube/{clusterId}/...` (в kubeconfig используется этот URL).
- Reverse proxy до API-сервера кластера (in-cluster svc или внешний endpoint).
- Аутентификация на Relay: JWT/mTLS → маппинг на роли/политику.
- Авторизация: policy-движок (allow-list verbs/namespaces/resources per tenant).
- Upstream к API Server: мTLS/сертификаты, пер-кластерные creds.
- Прозрачная поддержка WebSocket Upgrade (основной; субпротоколы `v*.channel.k8s.io`) и потоков (exec/attach/logs/port-forward); SPDY — только как legacy.  
+ Для отладки kubectl можно принудительно включить WS переменными окружения:  
+ `KUBECTL_REMOTE_COMMAND_WEBSOCKETS=1` (exec/attach/cp) и `KUBECTL_PORT_FORWARD_WEBSOCKETS=1` (port-forward).

Альтернативные режимы (v2/v3):
- v2 (ExecCredential): Relay выдает short-lived токены (OIDC или подпись Relay), kubeconfig использует ExecCredential плагин.
- v3 (Impersonation): при наличии `impersonate`-прав — прокси с `Impersonate-User/Group`, или per-user SA-tokens.

### 2) Компоненты и изменения в коде
- internal/kube/
  - cluster_registry: хранение кластеров, CA, endpoints, креды.
  - authz_policy: матрица прав на ресурсы (verbs/namespaces/apiGroups).
  - proxy: `httputil.ReverseProxy` + кастомный `Transport` (timeouts, mTLS, h2, keep-alive).
  - streams: WebSocket handler (exec/attach/logs/port-forward) + поддержка субпротоколов `v*.channel.k8s.io`; SPDY — только для legacy кластеров.
- cmd/relay/
  - маршрутизация `/kube/{clusterId}/...` до kube-proxy handler.
  - метрики Prometheus: `kube_proxy_requests_total`, `kube_proxy_request_duration_seconds`, `kube_proxy_upstream_errors_total`.
- config
  - clusters[]: {id, apiServerURL, ca, clientCert/key или serviceAccount, skipTLSVerify?}
  - policy: правила на уровень tenant/role (resources/verbs/namespaces).
- internal/transport/grpc/
  - kube_service (v2): IssueExecCredential, (опционально) PortForward gRPC bridge.
- client (позже)
  - `cloudbridge-client kube login` → генерит kubeconfig/ExecCredential.
  - `cloudbridge-client kube configure --cluster <id>`.

### 3) Спецлогика для kubectl (тонкости протокола)
- Exec/attach/logs/port-forward:
  - Поддержать WebSocket Upgrade (101 Switching Protocols) сквозь Relay; прокидывать `Sec-WebSocket-Protocol: v*.channel.k8s.io`, `Connection: Upgrade`, `Upgrade: websocket`. SPDY — только как legacy (флаг совместимости).
  - Сохранять заголовки: `Sec-WebSocket-Protocol`, `X-Stream-Protocol-Version` (legacy), `Connection`, `Upgrade`.
  - Не буферизовать тело; передавать bidirectional потоки в режиме raw.
  - Таймауты: длинные (watch/logs -f, exec), keep-alive ping.
  - Фичегейты: учитывать `TranslateStreamCloseWebsocketRequests` (>=1.29) и `PortForwardWebsockets` (>=1.30).
- Port-forward (TCP-only):
  - Прокидывать подпотоки порта; следить за backpressure.
  - Лимиты: макс. одновременные форварды per client/tenant.
- Watch:
  - Не сжимать/не мутировать события; передача chunked/streamed как есть.

### 4) Аутентификация/авторизация
- AuthN:
  - Relay проверяет JWT/mTLS → извлекает userId, tenantId, groups/roles.
- AuthZ (на Relay):
  - Политики на уровне tenant/role для verbs/resources/namespaces.
  - В v1 Relay сам контролирует доступ; upstream SA/cluster creds — технические.
- Upstream creds:
  - mTLS к API server (пер-кластерный client cert) или SA token с ограниченным RBAC.
  - Пер-кластерный service account с минимальными правами + строгая Relay-политика.
- v2 (после v1): ExecCredential OIDC (если API server доверяет IdP).  
+ • **Дефолт** — `client.authentication.k8s.io/v1beta1` (максимальная совместимость).  
+ • **При поддержке кластера** — переключение на `v1`; предусмотреть fallback, так как несовпадение версий частая причина ошибок.
- v3 (опционально): Impersonation — выставлять `Impersonate-User`, многоразовые `Impersonate-Group`, при необходимости `Impersonate-Extra-*`; требуется RBAC‑право `impersonate` у SA Relay на соответствующие сущности.

### 5) Безопасность
- mTLS к API server, проверка SAN/CA.
- Политики запретов (например, отказ от /proxy/ узлов, от `pods/exec` без явного разрешения).
- Rate limit per client/tenant, circuit breaker на upstream.
- Аудит: логировать user, tenant, verb, resource, namespace, resultCode, latency.

### 6) Наблюдаемость
- Метрики:
  - kube_proxy_requests_total{verb,resource,code}
  - kube_proxy_request_duration_seconds{verb,resource}
  - kube_ws_handshakes_total{code,subprotocol}
  - kube_ws_upgrades_total{subprotocol}
  - kube_ws_active_streams
  - kube_ws_errors_total{reason}
  - kube_upgrade_fallbacks_total{from} (например, spdy→ws)
  - kube_upstream_errors_total{code}
- Логи/аудит: структурированные события (request_id, user_id, tenant_id, cluster_id, protocol=ws|spdy, subprotocol=v*.channel.k8s.io).

### 7) Тесты (E2E)
- kind/minikube стенд.
- Сценарии:
  - kubectl get/describe/list/watch по namespaces.
  - kubectl logs -f.
  - kubectl exec -it sh (stdin/stdout/resize) — проверка WebSocket 101 и `Sec-WebSocket-Protocol`.
  - kubectl port-forward (TCP; проверка доступности порта).
- Негатив:
  - запрет по политике; проверка корректных кодов.
  - impersonation без RBAC `impersonate` — отказ с кодом **403** (Forbidden).
  - разрыв аплинка; восстановление стрима/корректное закрытие.

### 8) Доставка и конфигурация
- Helm values: clusters[], policies[], включение `/kube` прокси.
- Kubeconfig генератор (v1): готовый kubeconfig на Relay endpoint.
- Документация: “как подключить kubectl”, примеры политик, troubleshooting.  
  - **Важно:** `port-forward` — **TCP-only** (см. [issue #47862](https://github.com/kubernetes/kubernetes/issues/47862)).

### 9) План-этапы и оценки
- Неделя 1:
  - Cluster Registry + конфиг (0.5д)
  - Базовый Reverse Proxy до API server (1д)
  - AuthZ Policy (минимал) + метрики (1д)
  - E2E: `kubectl get pods`, `watch` (0.5д)
- Неделя 2:
  - WS‑streams: exec/logs/attach/port-forward (2д); SPDY legacy — опционально флагом совместимости
  - Port-forward (TCP‑only) (1д)
  - Rate limit + аудит (0.5д)
  - E2E портфель (0.5д)
- Неделя 3:
  - Политики на ресурсы/нэймспейсы (1д)
  - CLI kubeconfig генератор (0.5д)
  - Документация + Helm (0.5д)
  - Нагрузочное и стабильность (1д)
- Неделя 4 (v2 — опционально):
  - ExecCredential OIDC bridge (2д) — если API server доверяет IdP.
  - Multi-cluster routing + RBAC матрицы (2д).

### 10) Критические решения (до старта)
- Режим v1: прозрачный прокси с SA-токеном upstream и строгим AuthZ в Relay (быстро и надежно).
- Включаем ли Impersonation? Только если кластер настроен и есть `impersonate` права.
- Выбираем IdP-стратегию (Zitadel/Keycloak) для v2 (ExecCredential).

### 11) Критерии готовности (DoD)
- kubectl без патчей работает через Relay: `get/describe/list/watch/logs -f/exec/port-forward`.
- Политики применяются корректно, запреты отрабатывают.
- Метрики и аудит покрывают все типы запросов.
- Документация позволяет подключить кластер за ≤15 минут.

### 12) Интеграция с внешним LoadBalancer

**Балансировщик:** `LoadBalancer_3398`

| Параметр | Значение |
|----------|----------|
| Внешний IP | **109.120.180.160** |
| Внутренний IP | 10.0.0.242 |
| DNS внутри VPC | `loadbalancer-3398.mcs.local` |

#### Service c фиксированным EIP
```yaml
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay-external
  namespace: cloudbridge
spec:
  type: LoadBalancer
  loadBalancerIP: 109.120.180.160  # фиксируем EIP
  ports:
    - name: https-grpc
      protocol: TCP
      port: 443
      targetPort: 8444  # gRPC + kubectl WS
    - name: metrics
      protocol: TCP
      port: 5551
      targetPort: 5551  # опционально наружу
    - name: quic-udp   # если нужен MASQUE UDP (опц.)
      protocol: UDP
      port: 443
      targetPort: 8443
  selector:
    app: cloudbridge-relay
```

#### DNS-записи (публичная зона)
| Тип | Запись | Значение |
|-----|--------|----------|
| A   | `b1.2gc.space` | 109.120.180.160 |
| CNAME | `relay.2gc.space` | `b1.2gc.space.` |
| CNAME | `masque.2gc.space` (опц.) | `b1.2gc.space.` |

*Wildcard (multi-tenant):*
```text
*.relay.2gc.space.  IN  CNAME  b1.2gc.space.
```

> `b1.2gc.space` — основной DNS у внешнего LoadBalancer `LoadBalancer_3398`.


#### kubeconfig пример
```yaml
clusters:
- name: relay
  cluster:
    server: https://relay.2gc.space/kube/minikube   # или напрямую https://b1.2gc.space/...
    insecure-skip-tls-verify: true                  # либо ваш CA
```

> При использовании UDP-QUIC/MASQUE убедитесь, что LB пропускает **UDP 443**.
