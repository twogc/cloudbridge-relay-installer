# 🔧 ЗАДАНИЕ ДЛЯ DEVOPS: ИСПРАВЛЕНИЕ JWT И ПРОБЛЕМ С ХРАНИЛИЩЕМ

## 📋 ОБЗОР ПРОЕКТА

### Цель проекта CloudBridge Relay
CloudBridge Relay - это высокопроизводительный P2P (peer-to-peer) релейный сервер, предназначенный для:
- **P2P Mesh Networks**: Создание и управление mesh-сетями между устройствами
- **WireGuard Integration**: Интеграция с WireGuard для безопасных VPN-соединений
- **QUIC Protocol**: Поддержка современного протокола QUIC для быстрой передачи данных
- **Real-time Communication**: WebSocket соединения для real-time обновлений
- **Multi-tenant Architecture**: Поддержка множественных клиентов (tenants)
- **Kubernetes Deployment**: Развертывание в Kubernetes с Minikube для разработки

### Архитектура системы
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   P2P Clients   │◄──►│  CloudBridge    │◄──►│   Backend       │
│                 │    │     Relay       │    │   Services      │
│ - Mobile Apps   │    │                 │    │                 │
│ - Desktop Apps  │    │ - P2P API       │    │ - PostgreSQL    │
│ - IoT Devices   │    │ - WireGuard     │    │ - Redis         │
└─────────────────┘    │ - QUIC          │    │ - Cadence       │
                       │ - WebSocket     │    └─────────────────┘
                       └─────────────────┘
```

## 🎯 ОСНОВНЫЕ ЗАДАЧИ

### 1. 🔑 ИСПРАВЛЕНИЕ JWT АУТЕНТИФИКАЦИИ

#### Проблема
P2P API сервер не может валидировать JWT токены из-за неправильного использования секрета:
- **Текущее состояние**: Код использует `cfg.Django.JWTSecret` (Django конфигурация)
- **Проблема**: Django не используется в проекте, конфигурация пустая
- **Ожидаемое**: Использование переменной окружения `JWT_SECRET`

#### Текущий код (НЕПРАВИЛЬНЫЙ)
```go
// cmd/relay/main.go:342
jwtSecret := []byte(cfg.Django.JWTSecret)  // ❌ Django конфигурация
```

#### Требуемое исправление
```go
// cmd/relay/main.go:342
jwtSecret := []byte(os.Getenv("JWT_SECRET"))  // ✅ Переменная окружения
```

#### Файлы для изменения
- `cmd/relay/main.go` - основная логика инициализации JWT секрета
- Проверить другие места использования Django конфигурации

#### Проверка исправления
```bash
# После исправления в логах должно быть:
# "JWT secret prepared","secret_length":44  # вместо 18
```

### 2. 💾 РЕШЕНИЕ ПРОБЛЕМ С ХРАНИЛИЩЕМ

#### Проблема
Docker build падает с ошибкой "no space left on device":
```
failed to copy files: failed to copy directory: write /tools/cadence-web/node_modules/next/dist/compiled/next-server/app-page.runtime.dev.js: no space left on device
```

#### Анализ проблемы
1. **Размер проекта**: Проект содержит большие зависимости (node_modules, cadence-web)
2. **Docker context**: Весь проект копируется в Docker контекст
3. **Недостаток места**: На диске недостаточно места для сборки

#### Решения

##### Вариант 1: Очистка диска
```bash
# Проверить использование диска
df -h
du -sh /home/ubuntu/cloudbridge-relay-installer

# Очистить Docker
docker system prune -a
docker volume prune

# Очистить временные файлы
sudo apt autoremove
sudo apt autoclean
```

##### Вариант 2: Оптимизация Dockerfile
```dockerfile
# Использовать .dockerignore для исключения ненужных файлов
# Исключить node_modules, .git, документацию
```

##### Вариант 3: Увеличение места
```bash
# Если возможно, увеличить размер диска или добавить дополнительный диск
```

### 3. 🔍 ДОПОЛНИТЕЛЬНЫЕ ИССЛЕДОВАНИЯ

#### Изучить проект
1. **Прочитать документацию**:
   - `documentation/DEVOPS_RELAY_SERVER_TASK.md` - основное техническое задание
   - `README.md` - обзор проекта
   - `documentation/` - вся техническая документация

2. **Понять архитектуру**:
   - P2P API endpoints в `internal/api/`
   - WireGuard интеграция в `internal/wireguard/`
   - QUIC протокол в `internal/quic/`
   - WebSocket сервер в `internal/websocket/`

3. **Конфигурация**:
   - `config/config.go` - структура конфигурации
   - `test-config.json` - тестовая конфигурация
   - Kubernetes манифесты в `k8s/`

#### Проверить текущее состояние
```bash
# Статус Minikube
minikube status
kubectl get pods -n cloudbridge

# Логи P2P API
kubectl logs -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay

# Проверка JWT секрета
kubectl exec -n cloudbridge deployment/cloudbridge-relay -- env | grep JWT_SECRET
```

## 🚀 ПЛАН ВЫПОЛНЕНИЯ

### Этап 1: Изучение проекта (30 мин)
- [ ] Прочитать всю документацию в `documentation/`
- [ ] Изучить архитектуру и код
- [ ] Понять назначение P2P API и JWT аутентификации

### Этап 2: Исправление JWT (15 мин)
- [ ] Найти все места использования `cfg.Django.JWTSecret`
- [ ] Заменить на `os.Getenv("JWT_SECRET")`
- [ ] Проверить импорты (добавить `"os"` если нужно)

### Этап 3: Решение проблем с хранилищем (30 мин)
- [ ] Проверить использование диска
- [ ] Очистить Docker кэш и неиспользуемые образы
- [ ] Создать `.dockerignore` для исключения ненужных файлов
- [ ] При необходимости увеличить место на диске

### Этап 4: Тестирование (15 мин)
- [ ] Пересобрать Docker образ
- [ ] Загрузить в Minikube
- [ ] Перезапустить поды
- [ ] Протестировать JWT токен

## 🧪 ТЕСТИРОВАНИЕ

### JWT токен для тестирования
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsicmVsYXktc2VydmVyLWNsaWVudCJdLCJjb25uZWN0aW9uX3R5cGUiOiJwMnAtbWVzaCIsImRldmljZV9pZCI6ImRldmljZV8xNzU3NjM0NTEzIiwiaWF0IjoxNzU3NjM0NTEzLCJpc3MiOiJodHRwczovL2F1dGguMmdjLnJ1L3JlYWxtcy9yZWxheS1zZXJ2ZXJzIiwianRpIjoicDJwX3Rva2VuXzE3NTc2MzQ1MTMiLCJuYmYiOjE3NTc2MzQ1MTMsIm9yZ19pZCI6InRlc3Qtb3JnLTQ1NiIsInBlcm1pc3Npb25zIjpbInAycDpjb25uZWN0IiwicDJwOmRpc2NvdmVyIiwicDJwOm1lc2giXSwic3ViIjoicDJwLXNlcnZlci1jbGllbnQiLCJ0ZW5hbnRfaWQiOiJ0ZXN0LXRlbmFudC0xMjMifQ.AMcC-wlFXBddKf5mlfRlQfuS3PbBpecLN0vLgoQADwQ
```

### Команды для тестирования
```bash
# Health check
curl "http://95.163.250.190:30082/health"

# Peer discovery с JWT
curl -H "Authorization: <JWT_TOKEN>" \
     "http://95.163.250.190:30082/api/v1/peers/discover?tenant_id=test-tenant-123"
```

### Ожидаемый результат
```json
{
  "peers": [],
  "total": 0,
  "tenant_id": "test-tenant-123"
}
```

## 📊 КРИТЕРИИ УСПЕХА

### ✅ JWT исправлен
- [ ] В логах показывается `secret_length:44` (вместо 18)
- [ ] JWT токен валидируется успешно
- [ ] P2P API endpoints отвечают с кодом 200

### ✅ Проблемы с хранилищем решены
- [ ] Docker build проходит успешно
- [ ] Образ загружается в Minikube
- [ ] Поды запускаются без ошибок

### ✅ Система работает
- [ ] P2P API доступен на `http://95.163.250.190:30082`
- [ ] Health check возвращает статус "healthy"
- [ ] JWT аутентификация работает корректно

## 🔗 ПОЛЕЗНЫЕ ССЫЛКИ

- **JWT Secret в Kubernetes**: `kubectl get secret cloudbridge-secrets -n cloudbridge -o yaml`
- **Логи P2P API**: `kubectl logs -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay`
- **Статус подов**: `kubectl get pods -n cloudbridge`
- **Внешний доступ**: `http://95.163.250.190:30082`

## 📝 ЗАМЕТКИ

- Проект использует Go 1.23, Gin, WireGuard, QUIC
- P2P API работает на порту 8082 внутри Minikube, 30082 снаружи
- JWT секрет: `1Pc6bvrTUDtXMKm6mMF6+7UJi/EYAdOcGPkXCRYV4xI=`
- Nginx настроен как reverse proxy для внешнего доступа

---

**Приоритет**: 🔴 ВЫСОКИЙ  
**Время выполнения**: ~1.5 часа  
**Ответственный**: DevOps Engineer
