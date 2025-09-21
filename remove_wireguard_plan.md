# План удаления WireGuard из проекта

## 1. Основные файлы для удаления
- `internal/wireguard/` - вся папка
- `internal/cadence/workflows/wireguard.go` - workflow файл
- `internal/metrics/wireguard_test.go` - тесты метрик
- `internal/token/wireguard_config.go` - конфигурация токенов
- `internal/token/wireguard_config_test.go` - тесты конфигурации

## 2. Зависимости в go.mod
- `golang.zx2c4.com/wireguard/wgctrl`
- `golang.zx2c4.com/wireguard`

## 3. Файлы для обновления
- Удалить импорты wireguard
- Удалить упоминания в тестах
- Обновить конфигурации
- Обновить документацию

## 4. Тестовые файлы
- Удалить тесты, связанные с WireGuard
- Обновить unit тесты в workflows

## 5. Конфигурационные файлы
- Обновить config.json
- Обновить Docker файлы
- Обновить Kubernetes манифесты
