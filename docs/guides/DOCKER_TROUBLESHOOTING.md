# Docker Troubleshooting Guide

## Проблема: Ошибка при сборке Docker образа

### Ошибка:
```
fatal: unable to access 'https://git.zx2c4.com/wireguard-go.git/': Could not resolve host: git.zx2c4.com
```

### Причина:
Хост `git.zx2c4.com` недоступен или заблокирован в вашей сети (часто встречается в РФ).

## Решения:

### Решение 1: Использовать исправленный Dockerfile (РЕКОМЕНДУЕТСЯ)
Исправленный Dockerfile с альтернативными источниками WireGuard-go:
```bash
make docker-build-fixed
```

### Решение 2: Использовать упрощенный Dockerfile
Упрощенная версия без внешних зависимостей:
```bash
make docker-build-simple
```

### Решение 3: Использовать стандартную сборку
Если DNS проблемы решены:
```bash
make docker-build
```

### Решение 4: Ручная сборка с альтернативным Dockerfile
```bash
# С исправленным Dockerfile
docker build -f Dockerfile.relay.fixed -t cloudbridge-relay:fixed .

# С упрощенным Dockerfile
docker build -f Dockerfile.relay.simple -t cloudbridge-relay:simple .
```

### Решение 5: Проверить DNS настройки
Если проблема в DNS, добавьте в `/etc/hosts`:
```
8.8.8.8 git.zx2c4.com
```

### Решение 6: Временно сменить DNS
```bash
# Разово (Linux), если используется systemd-resolved
sudo resolvectl dns eth0 1.1.1.1 1.0.0.1
sudo resolvectl flush-caches

# Или отредактировать /etc/resolv.conf:
# nameserver 1.1.1.1
# nameserver 8.8.8.8
```

## Команды Makefile:

- `make docker-build` - стандартная сборка
- `make docker-build-fixed` - сборка с исправленным WireGuard-go
- `make docker-build-simple` - упрощенная сборка без внешних зависимостей
- `make docker-clean` - очистка Docker образов
- `make docker-run` - запуск контейнера
- `make docker-stop` - остановка контейнера

## Проверка сборки:

```bash
# Проверить собранные образы
docker images | grep cloudbridge

# Запустить тестовый контейнер
docker run --rm -it cloudbridge-relay:fixed --help
docker run --rm -it cloudbridge-relay:simple --help
```

## Логи сборки:

Если сборка все еще не работает, проверьте логи:
```bash
make docker-build-fixed 2>&1 | tee build-fixed.log
make docker-build-simple 2>&1 | tee build-simple.log
```

## Альтернативные источники WireGuard-go:

1. **GitHub Mirror** (основной): `https://github.com/WireGuard/wireguard-go.git`
2. **Официальный**: `https://git.zx2c4.com/wireguard-go.git`
3. **Системные пакеты**: `wireguard-tools` из дистрибутива

## Примечания:

- GitHub mirror является официальным и поддерживается авторами WireGuard
- Исправленный Dockerfile автоматически пробует несколько источников
- Упрощенный Dockerfile не требует внешних зависимостей
- Все образы совместимы с основным приложением
