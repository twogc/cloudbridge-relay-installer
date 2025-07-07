# CI/CD Security Guide

## Обзор

CloudBridge Relay использует комплексную систему безопасности в CI/CD, которая включает:

- **Автоматические проверки безопасности** при каждом PR и push
- **Еженедельные аудиты безопасности** с детальными отчетами
- **Автоматическое обновление зависимостей** через Dependabot
- **Статический анализ кода** через CodeQL
- **Сканирование уязвимостей** через Trivy и gosec

## Workflows

### 1. Основной CI/CD Pipeline (`.github/workflows/ci-cd.yml`)

**Триггеры:**
- Push в `main` и `develop`
- Pull Request в `main`

**Проверки безопасности:**
- **Trivy**: Сканирование уязвимостей в коде и зависимостях
- **CodeQL**: Статический анализ безопасности
- **gosec**: Анализ безопасности Go кода
- **TruffleHog**: Поиск секретов в коде
- **Nancy**: Проверка уязвимостей в Go зависимостях
- **Safety**: Проверка уязвимостей в Python зависимостях

### 2. Еженедельный аудит безопасности (`.github/workflows/security-audit.yml`)

**Триггеры:**
- Каждый понедельник в 9:00 UTC
- Ручной запуск через workflow_dispatch

**Функции:**
- Комплексный аудит всех компонентов
- Генерация детальных отчетов
- Создание issues при обнаружении уязвимостей
- Уведомления команды через Slack

## Автоматическое обновление зависимостей

### Dependabot (`.github/dependabot.yml`)

**Настройки:**
- **Go модули**: Еженедельно по понедельникам
- **GitHub Actions**: Еженедельно по понедельникам
- **Python (ML)**: Еженедельно по понедельникам
- **Python (SIEM)**: Еженедельно по понедельникам
- **Docker**: Еженедельно по понедельникам

**Особенности:**
- Автоматическое создание PR для обновлений
- Назначение ревьюеров из security-team
- Игнорирование major версий для критических зависимостей

## Инструменты безопасности

### 1. gosec
```bash
# Локальный запуск
./scripts/security/check_go_security.sh

# Форматы отчетов
./scripts/security/check_go_security.sh --html
./scripts/security/check_go_security.sh --text
./scripts/security/check_go_security.sh --json
```

### 2. Safety (Python)
```bash
# Проверка ML зависимостей
./scripts/security/check_dependencies_security.sh --ml

# Проверка SIEM зависимостей
./scripts/security/check_dependencies_security.sh --siem
```

### 3. Комплексный аудит
```bash
# Полный аудит
./scripts/security/security_audit_report.sh

# Быстрая проверка
./scripts/security/security_audit_report.sh --quick
```

## Мониторинг и уведомления

### Slack уведомления
- Канал: `#security`
- Канал: `#deployments`
- Уведомления о статусе security checks
- Алерты о критических уязвимостях

### GitHub Issues
- Автоматическое создание issues при обнаружении уязвимостей
- Лейблы: `security`, `audit`, `high-priority`
- Назначение ответственных команд

## Best Practices

### 1. Code Review
- Все PR должны проходить security review
- Обязательная проверка security-team для зависимостей
- Проверка на наличие секретов в коде

### 2. Зависимости
- Регулярное обновление зависимостей
- Проверка уязвимостей перед обновлением
- Тестирование после обновления

### 3. Мониторинг
- Регулярный просмотр security reports
- Отслеживание новых уязвимостей
- Обновление security tools

## Troubleshooting

### Частые проблемы

1. **gosec false positives**
   - Используйте `// #nosec G304` для подавления
   - Добавьте комментарий с объяснением

2. **Dependabot конфликты**
   - Решите конфликты вручную
   - Обновите зависимости по одной

3. **Security check failures**
   - Проверьте логи в GitHub Actions
   - Исправьте уязвимости согласно отчетам

### Полезные команды

```bash
# Проверка статуса security checks
gh run list --workflow=security-audit.yml

# Скачивание артефактов
gh run download <run-id>

# Ручной запуск security audit
gh workflow run security-audit.yml
```

## Контакты

- **Security Team**: security-team@company.com
- **DevOps Team**: devops-team@company.com
- **Slack**: #security, #devops 