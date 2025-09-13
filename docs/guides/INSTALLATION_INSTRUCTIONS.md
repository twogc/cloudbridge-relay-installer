# Инструкция по установке CloudBridge Relay

## Быстрый старт

### Автоматическая установка (рекомендуется)

```bash
# Скачивание и запуск установщика
curl -sSL https://raw.githubusercontent.com/twogc/cloudbridge-relay-installer/main/ubuntu_22_04_installer.sh | sudo bash
```

### Ручная установка

```bash
# 1. Скачать установщик
wget https://raw.githubusercontent.com/twogc/cloudbridge-relay-installer/main/ubuntu_22_04_installer.sh

# 2. Сделать исполняемым
chmod +x ubuntu_22_04_installer.sh

# 3. Запустить установку
sudo ./ubuntu_22_04_installer.sh
```

## Что устанавливается

### Модуль 01: System Update
- Обновление системы Ubuntu 22.04
- Установка базовых пакетов
- Настройка автоматических обновлений
- Создание системных пользователей

### Модуль 02: Security Basics
- Hardening SSH конфигурации
- Настройка параметров безопасности ядра
- Установка и настройка Fail2Ban
- Настройка AppArmor и rkhunter
- Система аудита

### Модуль 03: Docker Installation
- Установка Docker Engine
- Установка Docker Compose
- Настройка безопасности Docker
- Дополнительные инструменты

### Модуль 04: Firewall Setup
- Настройка UFW файрвола
- Открытие портов для сервисов
- Rate limiting для SSH

### Модуль 05: Keycloak Setup
- Развертывание Keycloak в Docker
- Настройка PostgreSQL
- Создание realm'ов и пользователей

### Модуль 06: Lynis Setup
- Установка Lynis для аудита безопасности
- Настройка автоматических проверок
- Первоначальный аудит системы

### Модуль 07: Monitoring Setup
- Установка Prometheus
- Установка Grafana
- Настройка дашбордов мониторинга

### Модуль 08: Final Configuration
- Создание системных сервисов
- Настройка резервного копирования
- Утилиты управления

## Доступ к сервисам

После установки будут доступны:

- **Keycloak**: http://localhost:8080 (admin/admin123)
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin123)

## Утилиты управления

```bash
# Проверка статуса всех сервисов
sudo cloudbridge-status

# Проверка безопасности
sudo cloudbridge-security-check

# Аудит системы
sudo lynis audit system

# Резервное копирование
sudo cloudbridge-backup.sh
```

## Требования

- Ubuntu 22.04 LTS
- Минимум 2GB RAM
- Минимум 10GB свободного места
- Интернет соединение

## Логи установки

```bash
# Основной лог установки
tail -f /var/log/cloudbridge-install.log

# Логи отдельных модулей
tail -f /var/log/cloudbridge-module-*.log
```

## Устранение неполадок

### Проблемы с Docker
```bash
sudo systemctl restart docker
sudo systemctl status docker
```

### Проблемы с файрволом
```bash
sudo ufw status
sudo ufw --force reset
```

### Проблемы с Keycloak
```bash
cd /opt/cloudbridge/keycloak
sudo docker-compose logs keycloak
```

## Поддержка

- Документация: [SCRIPTS_ANALYSIS.md](SCRIPTS_ANALYSIS.md)
- Руководство по модульной установке: [MODULAR_INSTALLER_GUIDE.md](MODULAR_INSTALLER_GUIDE.md)
- Анализ скриптов: [SCRIPTS_ANALYSIS.md](SCRIPTS_ANALYSIS.md)
