# CloudBridge Appliance

Enterprise-grade коробочное решение для создания безопасных виртуальных локальных сетей с полным enterprise контролем, аудитом и multi-tenancy изоляцией.

## 🏗️ Структура проекта

```
appliance/
├── docker/                 # Docker образы и конфигурации
│   ├── Dockerfile.appliance
│   ├── docker-compose.appliance.yml
│   └── build-appliance.sh
├── api/                    # Appliance Management API
│   ├── appliance_manager.go
│   ├── types.go
│   └── server.go
├── pilot/                  # Пилотная программа
│   ├── pilot-program.md
│   ├── pilot-application.json
│   └── pilot-metrics.json
├── configs/                # Конфигурационные файлы
│   └── appliance-config.yaml
├── scripts/                # Скрипты управления
│   ├── install-appliance.sh
│   └── manage-appliance.sh
└── docs/                   # Документация
    ├── README.md
    └── pricing-guide.md
```

## 🚀 Быстрый старт

### Установка

```bash
# Клонировать репозиторий
git clone https://github.com/cloudbridge/relay-installer.git
cd relay-installer/appliance

# Запустить установку
sudo ./scripts/install-appliance.sh
```

### Управление

```bash
# Управление сервисом
./scripts/manage-appliance.sh start
./scripts/manage-appliance.sh stop
./scripts/manage-appliance.sh restart
./scripts/manage-appliance.sh status

# Мониторинг
./scripts/manage-appliance.sh health
./scripts/manage-appliance.sh metrics
./scripts/manage-appliance.sh logs
```

## 📦 Docker образы

### Сборка образа

```bash
cd docker
./build-appliance.sh
```

### Запуск с Docker Compose

```bash
docker-compose -f docker-compose.appliance.yml up -d
```

## 🔧 API управления

### Статус системы

```bash
curl -H "X-API-Key: your-key" \
  http://localhost:8080/api/v1/appliance/status
```

### Метрики

```bash
curl -H "X-API-Key: your-key" \
  http://localhost:8080/api/v1/appliance/metrics
```

### Конфигурация

```bash
curl -H "X-API-Key: your-key" \
  http://localhost:8080/api/v1/appliance/config
```

## 🎯 Пилотная программа

### Подача заявки

```bash
# Заполнить заявку
cp pilot/pilot-application.json pilot/my-application.json
# Отредактировать my-application.json
# Отправить на pilot@cloudbridge.ru
```

### Метрики пилота

```bash
# Просмотр метрик
cat pilot/pilot-metrics.json | jq '.pilot_metrics'
```

## 💰 Ценообразование

### Конфигурации

- **CloudBridge Gateway Basic**: 6 000 000 ₽
- **CloudBridge Gateway Secure**: 12 000 000 ₽  
- **CloudBridge Gateway Enterprise**: 25 000 000 ₽

### Тарифные планы

- **Startup**: 0 ₽ (тестовый период)
- **Start**: 150 ₽/пользователь
- **Business**: 100 ₽/пользователь + 5 000 ₽ база
- **Enterprise**: 50 000 ₽ база + 50 ₽/пользователь свыше 200

## 📚 Документация

- **Руководство администратора**: [docs/README.md](docs/README.md)
- **Ценообразование**: [docs/pricing-guide.md](docs/pricing-guide.md)
- **Пилотная программа**: [pilot/pilot-program.md](pilot/pilot-program.md)

## 🔐 Безопасность

### Встроенные функции

- **WAF**: Web Application Firewall
- **DDoS Protection**: Защита от атак
- **Zero Trust**: Модель нулевого доверия
- **Audit Logging**: Аудит всех действий
- **Encryption**: Шифрование трафика

### Интеграция

- **Active Directory**: Аутентификация пользователей
- **LDAP**: Централизованное управление
- **RADIUS**: Аутентификация сетевых устройств
- **SIEM**: Централизованное логирование

## 📊 Мониторинг

### Встроенные метрики

- **Производительность**: Throughput, latency, packet loss
- **Ресурсы**: CPU, memory, disk, network
- **Безопасность**: Failed logins, blocked attacks
- **Соединения**: Active connections, peak usage

### Интеграция

- **Prometheus**: Экспорт метрик
- **Grafana**: Дашборды и алерты
- **Zabbix**: Мониторинг инфраструктуры
- **SIEM**: Централизованное логирование

## 🛠️ Troubleshooting

### Частые проблемы

```bash
# Проверить статус
./scripts/manage-appliance.sh status

# Проверить логи
./scripts/manage-appliance.sh logs

# Проверить здоровье
./scripts/manage-appliance.sh health
```

### Резервное копирование

```bash
# Создать резервную копию
./scripts/manage-appliance.sh backup

# Восстановить из резервной копии
./scripts/manage-appliance.sh restore backup-file.tar.gz
```

## 📞 Поддержка

### Техническая поддержка

- **Email**: support@cloudbridge.ru
- **Phone**: +7 (495) 123-45-67
- **Telegram**: @cloudbridge_support

### Коммерческие вопросы

- **Email**: sales@cloudbridge.ru
- **Phone**: +7 (495) 123-45-68
- **Telegram**: @cloudbridge_sales

## 📄 Лицензия

Copyright (c) 2024 CloudBridge. Все права защищены.

## 🤝 Участие в разработке

1. Fork репозитория
2. Создать feature branch
3. Commit изменения
4. Push в branch
5. Создать Pull Request

## 📈 Roadmap

- [ ] Поддержка Kubernetes
- [ ] Интеграция с AWS/Azure
- [ ] Мобильные приложения
- [ ] AI-powered мониторинг
- [ ] Автоматическое масштабирование
