CloudBridge Relay — платформа для построения «умного» глобального транспорта поверх интернета, которая объединяет современные сетевые протоколы (QUIC, MASQUE, WireGuard, gRPC) и автоматический NAT-traversal в единую экосистему «сервер + клиент».

────────────────────────────────────────
1. Главная идея
────────────────────────────────────────
• Дать приложениям надёжный, защищённый и быстрый туннель слой 3/4/7, который сам выбирает оптимальный маршрут и протокол.
• Избавить DevOps-команды от мучений с NAT, фаерволами, Anycast-сертификацией и мониторингом — всё делается внутри Relay-сети.

────────────────────────────────────────
2. Ключевые функции
────────────────────────────────────────
1. QUIC / MASQUE
   • CONNECT-UDP и CONNECT-IP для проксирования UDP-портов или целых подсетей.
   • HTTP Datagrams как основной формат трафика; Capsule fallback, если QUIC-датаграммы заблокированы.

2. WireGuard fallback
   • При блокировке UDP :8443 клиент автоматически переключается на L3-туннель WG.
   • Обратный переход при восстановлении качества.

3. P2P-mesh (ICE / STUN / TURN / DERP)
   • Клиенты пытаются выйти в прямой p2p-канал; Relay служит резервом.
   • Подробные NAT-метрики и auto-switch между режимами.

4. gRPC-контрольный канал
   • Hello / Auth / CreateTunnel / Heartbeat — поверх одного TLS 1.3-HTTP2 соединения.
   • JSON-API оставлен как fallback (deprecated).

5. Безопасность
   • TLS 1.3 + пост-квантовый KEM.
   • JWT (Zitadel) / mTLS, роль-based ACL.
   • Circuit-Breaker и Rate-Limiter на сервере.

6. Наблюдаемость
   • Prometheus-метрики на клиенте и сервере, опциональный Pushgateway.
   • Grafana dashboards, Loki логи, Thanos федерация.

7. Smart-Routing
   • Регистр PoP-ов с гео-тегами и RTT-пробами.
   • Клиент параллельно пингует 3–5 ближайших узлов (Happy Eyeballs) и выбирает лучший.
   • В планах — overlay-транзит через core-PoP и Anycast BGP.

────────────────────────────────────────
3. Архитектура
────────────────────────────────────────
• Edge PoP (Relay-узел) — HTTP/3 Ingress, MASQUE, STUN/TURN/DERP, Prometheus-exporter.
• Core PoP — облегчённый Relay для межрегионых «коротких» hop-ов.
• Control-Plane — Registry API, Auth (Zitadel), Metrics (Prometheus + Thanos), Logs (Loki), Workflow (Cadence/Temporal).

────────────────────────────────────────
4. Текущий статус (Q3 2025)
────────────────────────────────────────
 Сервер: gRPC-API на :8444, JSON-API на :5550, TLS 1.3, Prometheus-метрики.
 Клиент: тоннели, p2p-mesh, WireGuard fallback, hot-reload конфигурации, Pushgateway-метрики.
 Helm-чарт для Minikube / K8s, порт-forward работает.
В работе: Retry-policy gRPC, автоген CA-пиннинг, Anycast rollout.

 Добавлено в Q3-2025:
• **Полноценный gRPC-API** на сервере (:8444) и клиентский транспорт по умолчанию.
• **WireGuard Auto-Switch** — автоматический переход QUIC ↔ WG и корректный teardown.
• **Hot-Reload конфигурации** клиента (fsnotify) + Pushgateway-метрики.

────────────────────────────────────────
5. Типовой сценарий подключения
────────────────────────────────────────
1. Клиент делает STUN → узнаёт NAT-тип и ближайшие PoP.
2. Получает JWT от Zitadel и параллельно пробует QUIC-handshake к 3–5 PoP.
3. Выбирает лучший, проводит MASQUE CONNECT-UDP/IP.
4. Если сеть режет QUIC → авто-переключение на WireGuard.
5. Каждые 30 с heartbeat; при ухудшении качества — динамический reroute.
6. Метрики в Pushgateway; сервер Relay экспортирует свои в Prometheus.

────────────────────────────────────────
6. Использование
────────────────────────────────────────
• Сервер: `helm install cloudbridge-relay …` (порт :8444).
• Клиент:
```
cbctl init # создаёт YAML + токен
cbctl tunnel --transport grpc \
     --local-port 2222 \
     --remote-host 10.0.0.10 --remote-port 22
```

────────────────────────────────────────
7. Roadmap (следующие шаги)
────────────────────────────────────────
1. Перенос client-side signalling на gRPC.
2. Retry-/backoff-policy и TLS session resumption.
3. Публичный Anycast IP или GeoDNS (k8gb) для single entry-point.
4. BGP-lite overlay и динамическая маршрутизация между PoP.
5. Desktop-инсталляторы (Windows MSI, macOS PKG) и авто-update.

CloudBridge Relay уже закрывает подавляющее большинство use-case’ов VPN, туннелей и edge-прокси, предлагая современный QUIC-first стек с автоматической адаптацией к сетевым условиям и полноценным мониторингом.