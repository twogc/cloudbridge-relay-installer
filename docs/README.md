# CloudBridge Relay Documentation

Полная документация по CloudBridge Relay - системе для создания защищенных VPN туннелей с поддержкой мультитенантности.

## 📚 Структура документации

```
docs/
├── guides/           # Руководства и инструкции
├── deployment/       # Документация по развертыванию
├── security/         # Документация по безопасности
├── integration/      # Документация по интеграциям
├── troubleshooting/  # Устранение неполадок
├── reports/          # Отчеты и технические задания
├── changelog/        # Журналы изменений
└── README.md         # Этот файл
```

## 🚀 Быстрый старт

### Для новых пользователей:
1. [Руководство по быстрому старту с Telegram](guides/quick_start_telegram.md)
2. [Руководство по развертыванию](deployment/deployment_guide.md)
3. [Настройка уведомлений Telegram](guides/telegram_notifications_setup.md)

### Для администраторов:
1. [Руководство по безопасности](security/security.md)
2. [Интеграция с Keycloak](integration/keycloak_integration.md)
3. [Мониторинг безопасности](security/security_monitoring.md)

## 📖 Основные разделы

### 📋 Guides - Руководства
- **API Reference** - Справочник по API
- **Dashboard Guide** - Руководство по дашбордам
- **Metrics** - Метрики и мониторинг
- **Multi-tenancy Guide** - Руководство по мультитенантности
- **Ports Configuration** - Конфигурация портов
- **Quick Start Telegram** - Быстрый старт с Telegram
- **Quick Token Guide** - Руководство по токенам
- **TCP Proxy Guide** - Руководство по TCP прокси

### 🚀 Deployment - Развертывание
- **Deployment Guide** - Основное руководство по развертыванию
- **Moscow Region Quickstart** - Быстрый старт для московского региона
- **SIEM Quick Deployment** - Быстрое развертывание SIEM
- **SIEM Separate Server Deployment** - Развертывание SIEM на отдельном сервере
- **SIEM Deployment Complete Guide** - Полное руководство по развертыванию SIEM

### 🔒 Security - Безопасность
- **Security** - Основная документация по безопасности
- **Security Hardening Checklist** - Чек-лист усиления безопасности
- **Security Monitoring** - Мониторинг безопасности
- **Security Monitoring Telegram** - Мониторинг безопасности с Telegram
- **Lynis Integration** - Интеграция с Lynis
- **Lynis Quick Reference** - Быстрая справка по Lynis

### 🔗 Integration - Интеграции
- **Keycloak Integration** - Интеграция с Keycloak
- **Keycloak Technical Specification** - Техническая спецификация Keycloak
- **LDAP/AD Integration** - Интеграция с LDAP/Active Directory
- **SIEM Integration Guide** - Руководство по интеграции с SIEM
- **Client Integration** - Интеграция клиентов
- **Server Client Implementation** - Реализация сервер-клиент
- **Server Client Testing** - Тестирование сервер-клиент

### 🔧 Troubleshooting - Устранение неполадок
- **Debugging** - Отладка системы
- **Fail2ban Troubleshooting** - Устранение проблем с Fail2ban
- **Blackbox TLS Fix** - Исправление проблем с TLS в Blackbox

### 📊 Reports - Отчеты
- **Tech Task Monitoring Security** - Техническое задание по мониторингу безопасности
- **Migration Guide v2** - Руководство по миграции v2
- **Alertmanager Status Report** - Отчет о статусе Alertmanager
- **Blackbox TLS Fix Report** - Отчет об исправлении TLS в Blackbox
- **Firewall Ports Report** - Отчет о портах файрвола
- **Firewall Setup Complete** - Полная настройка файрвола

### 📝 Changelog - Журналы изменений
- **CHANGELOG** - Основной журнал изменений
- **CHANGELOG v2** - Журнал изменений версии 2

## 🎯 Рекомендуемые маршруты чтения

### Для разработчиков:
1. [API Reference](guides/api_reference.md)
2. [Server Client Technical Specification](integration/server_client_technical_specification.md)
3. [Client Integration](integration/client_integration.md)
4. [Debugging](troubleshooting/debugging.md)

### Для DevOps инженеров:
1. [Deployment Guide](deployment/deployment_guide.md)
2. [Security](security/security.md)
3. [Metrics](guides/metrics.md)
4. [Troubleshooting](troubleshooting/debugging.md)

### Для администраторов безопасности:
1. [Security](security/security.md)
2. [Security Hardening Checklist](security/security_hardening_checklist.md)
3. [Lynis Integration](security/lynis_integration.md)
4. [SIEM Integration Guide](integration/siem_integration_guide.md)

## 🔍 Поиск документации

### По ключевым словам:
- **Установка**: [Deployment Guide](deployment/deployment_guide.md)
- **Безопасность**: [Security](security/security.md)
- **Keycloak**: [Keycloak Integration](integration/keycloak_integration.md)
- **SIEM**: [SIEM Integration Guide](integration/siem_integration_guide.md)
- **Мониторинг**: [Metrics](guides/metrics.md)
- **Отладка**: [Debugging](troubleshooting/debugging.md)
- **Telegram**: [Quick Start Telegram](guides/quick_start_telegram.md)

## 📞 Поддержка

При возникновении вопросов:
1. Проверьте раздел [Troubleshooting](troubleshooting/debugging.md)
2. Изучите [Debugging](troubleshooting/debugging.md)
3. Обратитесь к [Security](security/security.md) для проблем безопасности
4. Создайте issue в репозитории проекта

## 🔄 Обновления

Последние изменения в документации:
- [CHANGELOG](changelog/CHANGELOG.md) - основной журнал изменений
- [CHANGELOG v2](changelog/CHANGELOG_v2.md) - изменения версии 2

---

*Документация обновляется регулярно. Для получения последней версии обратитесь к репозиторию проекта.* 