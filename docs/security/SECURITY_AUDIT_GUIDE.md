# CloudBridge Relay - Security Audit Guide

## Обзор

CloudBridge Relay включает комплексную систему аудита безопасности, которая проверяет:

- **Python зависимости** - уязвимости в Python пакетах
- **Go код** - проблемы безопасности в исходном коде
- **Системная безопасность** - общие настройки безопасности

## Быстрый старт

### 1. Установка инструментов

```bash
# Установка safety для Python зависимостей
pip3 install safety==3.5.2

# Установка gosec для Go кода
sudo snap install gosec

# Установка jq для обработки JSON
sudo apt install jq
```

### 2. Запуск полного аудита

```bash
# Полный аудит безопасности
./scripts/security/security_audit_report.sh

# Быстрая проверка
./scripts/security/security_audit_report.sh --quick

# Подробный вывод
./scripts/security/security_audit_report.sh --verbose
```

## Индивидуальные проверки

### Проверка Python зависимостей

```bash
# Проверка всех Python зависимостей
./scripts/security/check_dependencies_security.sh

# Только JSON отчет
./scripts/security/check_dependencies_security.sh --json

# Подробный вывод
./scripts/security/check_dependencies_security.sh --verbose
```

**Проверяемые файлы:**
- `ml/requirements.txt`
- `siem/requirements.txt`

### Проверка Go кода

```bash
# Полный анализ Go кода
./scripts/security/check_go_security.sh

# Только HTML отчет
./scripts/security/check_go_security.sh --html

# Только текстовый отчет
./scripts/security/check_go_security.sh --text

# Только JSON отчет
./scripts/security/check_go_security.sh --json
```

**Проверяемые директории:**
- `cmd/`
- `internal/`
- `pkg/` (если существует)

## Результаты и отчеты

### Структура отчетов

```
security_reports/
├── security_audit_report.md          # Сводный отчет
├── gosec_report.html                 # HTML отчет Go анализа
├── gosec_report.txt                  # Текстовый отчет Go анализа
├── ml_requirements_security_report.json    # JSON отчет ML зависимостей
└── siem_requirements_security_report.json  # JSON отчет SIEM зависимостей
```

### Интерпретация результатов

#### Python зависимости (Safety)

- **Уязвимости найдены** - обновить пакеты до безопасных версий
- **Критические уязвимости** - немедленное обновление
- **Устаревшие пакеты** - рассмотреть альтернативы

#### Go код (Gosec)

- **G115** - Integer overflow - проверить границы значений
- **G402** - TLS MinVersion too low - обновить TLS настройки
- **G114** - Use of net/http serve function - добавить таймауты
- **G304** - Potential file inclusion - валидировать пути файлов
- **G302/G301** - File permissions - исправить права доступа

### Уровни риска

- 🔴 **CRITICAL** (>20 проблем) - Немедленное внимание
- 🟠 **HIGH** (10-20 проблем) - Срочное внимание
- 🟡 **MEDIUM** (5-10 проблем) - Внимание требуется
- 🟢 **LOW** (<5 проблем) - Мониторинг и поддержание

## Автоматизация

### CI/CD интеграция

```yaml
# .github/workflows/security.yml
name: Security Audit

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.10'
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install tools
        run: |
          pip install safety==3.5.2
          sudo snap install gosec
          sudo apt install jq
      
      - name: Run security audit
        run: ./scripts/security/security_audit_report.sh
      
      - name: Upload reports
        uses: actions/upload-artifact@v3
        with:
          name: security-reports
          path: security_reports/
```

### Cron job для регулярных проверок

```bash
# Добавить в crontab
# Еженедельная проверка безопасности
0 2 * * 1 /path/to/cloudbridge-relay-installer/scripts/security/security_audit_report.sh --quick
```

## Рекомендации по безопасности

### Немедленные действия

1. **Обновление зависимостей**
   ```bash
   # Обновить Python зависимости
   pip install --upgrade -r ml/requirements.txt
   pip install --upgrade -r siem/requirements.txt
   
   # Обновить Go зависимости
   go mod tidy
   go mod download
   ```

2. **Исправление проблем Go кода**
   - Добавить валидацию входных данных
   - Исправить права доступа к файлам
   - Обновить TLS настройки
   - Добавить таймауты для HTTP серверов

### Долгосрочные меры

1. **Автоматизация**
   - Настроить автоматические проверки в CI/CD
   - Регулярные аудиты безопасности
   - Мониторинг новых уязвимостей

2. **Обучение команды**
   - Безопасные практики программирования
   - Регулярные тренинги по безопасности
   - Code review с фокусом на безопасность

3. **Мониторинг**
   - Отслеживание CVE баз данных
   - Автоматические уведомления о новых уязвимостях
   - Регулярные penetration тесты

## Устранение неполадок

### Частые проблемы

#### Safety не найден
```bash
# Установить safety
pip3 install safety==3.5.2
export PATH="$HOME/.local/bin:$PATH"
```

#### Gosec не найден
```bash
# Установить gosec
sudo snap install gosec
```

#### Ошибки JSON парсинга
```bash
# Установить jq
sudo apt install jq
```

#### Проблемы с правами доступа
```bash
# Сделать скрипты исполняемыми
chmod +x scripts/security/*.sh
```

### Логи и отладка

```bash
# Подробный вывод
./scripts/security/security_audit_report.sh --verbose

# Проверка отдельных компонентов
safety check -r ml/requirements.txt
gosec ./...
```

## Контакты и поддержка

- **Документация:** [docs/security/](docs/security/)
- **Issues:** GitHub Issues
- **Security:** security@cloudbridge-relay.com

---

*Последнее обновление: $(date)* 