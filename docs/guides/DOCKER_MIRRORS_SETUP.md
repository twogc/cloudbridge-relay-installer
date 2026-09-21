# Настройка зеркал Docker Hub для России

## Проблема

В ночь на 30 мая 2024 г. сервис Docker Hub закрыл доступ для пользователей из России.

> Since Docker is a US company, we must comply with US export control regulations. In an effort to comply with these, we now block all IP addresses that are located in Cuba, Iran, North Korea, Republic of Crimea, Sudan, and Syria.

Docker, как американская компания, обязана соблюдать экспортные контрольные регуляции США, которые включают блокировку IP-адресов из указанных стран.

## Решение

### Список актуальных зеркал Docker Hub

| Зеркало | Описание | Статус |
|---------|----------|--------|
| `https://dockerhub.timeweb.cloud` | Зеркало Timeweb | Активно |
| `https://huecker.io` | Зеркало Huecker | ктивно |
| `https://mirror.gcr.io` | Зеркало Google | ктивно |
| `https://c.163.com` | Зеркало Китай | ктивно |
| `https://registry.docker-cn.com` | Зеркало Китай | ктивно |
| `https://daocloud.io` | Зеркало Китай | ктивно |
| `https://cr.yandex/mirror` | Зеркало Яндекс | ктивно |
| `https://noohub.ru` | Зеркало noosoft | ктивно |
| `https://quay.io` | Зеркало Redhat | ктивно |
| `https://registry.access.redhat.com` | Зеркало Redhat | ктивно |
| `https://registry.redhat.io` | Зеркало Redhat | ктивно |
| `https://public.ecr.aws` | Зеркало Amazon | ктивно |

## Настройка для Linux

### Обычная установка

```bash
# Создание/обновление конфигурации Docker
cat << EOF | sudo tee -a /etc/docker/daemon.json
{
  "registry-mirrors": [
    "https://dockerhub.timeweb.cloud",
    "https://huecker.io",
    "https://mirror.gcr.io",
    "https://c.163.com",
    "https://registry.docker-cn.com",
    "https://daocloud.io",
    "https://cr.yandex/mirror",
    "https://noohub.ru"
  ]
}
EOF

# Перезапуск Docker
sudo systemctl restart docker

# Проверка настройки
docker info | grep -A 10 "Registry Mirrors"
```

### Rootless режим

```bash
# Создание конфигурации для rootless Docker
mkdir -p ~/.config/docker
cat > ~/.config/docker/daemon.json <<EOF
{
  "registry-mirrors": [
    "https://dockerhub.timeweb.cloud",
    "https://huecker.io",
    "https://mirror.gcr.io",
    "https://c.163.com",
    "https://registry.docker-cn.com",
    "https://daocloud.io",
    "https://cr.yandex/mirror",
    "https://noohub.ru"
  ]
}
EOF

# Перезапуск Docker
systemctl --user restart docker
```

## Настройка для macOS

```bash
# Создание/редактирование конфигурации
mkdir -p $HOME/.docker
cat > $HOME/.docker/daemon.json <<EOF
{
  "registry-mirrors": [
    "https://dockerhub.timeweb.cloud",
    "https://huecker.io",
    "https://mirror.gcr.io",
    "https://c.163.com",
    "https://registry.docker-cn.com",
    "https://daocloud.io",
    "https://cr.yandex/mirror",
    "https://noohub.ru"
  ],
  "experimental": true
}
EOF

# Перезапуск Docker Desktop
# Через GUI: Docker Desktop -> Settings -> Restart
# Или через командную строку:
killall Docker && open /Applications/Docker.app
```

## Настройка для Windows

### Через Docker Desktop

1. Откройте Docker Desktop
2. Перейдите в Settings → Docker Engine
3. Добавьте конфигурацию:

```json
{
  "registry-mirrors": [
    "https://dockerhub.timeweb.cloud",
    "https://huecker.io",
    "https://mirror.gcr.io",
    "https://c.163.com",
    "https://registry.docker-cn.com",
    "https://daocloud.io",
    "https://cr.yandex/mirror",
    "https://noohub.ru"
  ]
}
```

4. Нажмите "Apply & Restart"

### Через файл конфигурации

```bash
# Создание файла конфигурации
mkdir -p C:\ProgramData\docker\config
cat > C:\ProgramData\docker\config\daemon.json <<EOF
{
  "registry-mirrors": [
    "https://dockerhub.timeweb.cloud",
    "https://huecker.io",
    "https://mirror.gcr.io",
    "https://c.163.com",
    "https://registry.docker-cn.com",
    "https://daocloud.io",
    "https://cr.yandex/mirror",
    "https://noohub.ru"
  ]
}
EOF
```

## Автоматическая настройка

### Скрипт для Linux

```bash
#!/bin/bash
# Автоматическая настройка зеркал Docker Hub

echo " Настройка зеркал Docker Hub..."

# Создание резервной копии
if [ -f /etc/docker/daemon.json ]; then
    sudo cp /etc/docker/daemon.json /etc/docker/daemon.json.backup
    echo " Создана резервная копия /etc/docker/daemon.json.backup"
fi

# Настройка зеркал
cat << EOF | sudo tee /etc/docker/daemon.json
{
  "registry-mirrors": [
    "https://dockerhub.timeweb.cloud",
    "https://huecker.io",
    "https://mirror.gcr.io",
    "https://c.163.com",
    "https://registry.docker-cn.com",
    "https://daocloud.io",
    "https://cr.yandex/mirror",
    "https://noohub.ru"
  ]
}
EOF

# Перезапуск Docker
echo " Перезапуск Docker..."
sudo systemctl restart docker

# Проверка настройки
echo " Проверка настройки..."
docker info | grep -A 10 "Registry Mirrors"

echo "Настройка завершена!"
```

### Использование готового скрипта

```bash
# Скачивание и запуск автоматического скрипта
wget -O unlock.sh https://raw.githubusercontent.com/StasPlov/docker-unlock/main/unlock.sh
chmod +x unlock.sh
./unlock.sh
```

## Тестирование

### Проверка настройки

```bash
# Проверка зеркал
docker info | grep -A 10 "Registry Mirrors"

# Тестирование загрузки образов
docker pull alpine:latest
docker pull nginx:latest
docker pull redis:latest
```

### Проверка скорости

```bash
# Тестирование скорости загрузки
time docker pull alpine:latest
time docker pull nginx:latest
time docker pull redis:latest
```

## Альтернативные методы

### Прямое указание зеркала

```bash
# Загрузка с конкретного зеркала
docker pull cr.yandex/mirror/alpine:latest
docker pull huecker.io/library/nginx:latest
docker pull dockerhub.timeweb.cloud/library/redis:latest
```

### Использование прокси

```bash
# Настройка прокси для Docker
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
export NO_PROXY=localhost,127.0.0.1

# Запуск Docker с прокси
docker run --rm alpine:latest echo "Hello from mirror!"
```

## Устранение неполадок

### Проблема: Зеркала не работают

```bash
# Проверка доступности зеркал
curl -I https://dockerhub.timeweb.cloud/v2/
curl -I https://huecker.io/v2/
curl -I https://mirror.gcr.io/v2/

# Проверка DNS
nslookup dockerhub.timeweb.cloud
nslookup huecker.io
```

### Проблема: Медленная загрузка

```bash
# Тестирование скорости зеркал
for mirror in dockerhub.timeweb.cloud huecker.io mirror.gcr.io; do
    echo "Тестирование $mirror..."
    time curl -s https://$mirror/v2/ > /dev/null
done
```

### Проблема: Ошибки аутентификации

```bash
# Очистка кэша Docker
docker system prune -a

# Перезапуск Docker
sudo systemctl restart docker

# Проверка логов
journalctl -u docker.service -f
```

## Мониторинг зеркал

### Скрипт проверки статуса

```bash
#!/bin/bash
# Проверка статуса зеркал Docker Hub

MIRRORS=(
    "dockerhub.timeweb.cloud"
    "huecker.io"
    "mirror.gcr.io"
    "c.163.com"
    "registry.docker-cn.com"
    "daocloud.io"
    "cr.yandex"
    "noohub.ru"
)

echo "Проверка статуса зеркал Docker Hub..."
echo "=========================================="

for mirror in "${MIRRORS[@]}"; do
    echo -n "Проверка $mirror... "
    if curl -s --connect-timeout 5 https://$mirror/v2/ > /dev/null 2>&1; then
        echo " Доступно"
    else
        echo " Недоступно"
    fi
done
```

## Рекомендации

### Лучшие зеркала для России

1. **dockerhub.timeweb.cloud** - стабильное, быстрое
2. **huecker.io** - хорошая скорость
3. **mirror.gcr.io** - надежное Google зеркало
4. **cr.yandex/mirror** - российское зеркало

### Оптимальная конфигурация

```json
{
  "registry-mirrors": [
    "https://dockerhub.timeweb.cloud",
    "https://huecker.io",
    "https://mirror.gcr.io",
    "https://cr.yandex/mirror"
  ]
}
```

## Полезные ссылки

- [Официальная документация Docker](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-configuration-file)
- [Docker Unlock Script](https://github.com/StasPlov/docker-unlock)
- [Список зеркал Docker Hub](https://github.com/docker/docker.github.io/issues/6910)

---

**Удачной работы с Docker! **
