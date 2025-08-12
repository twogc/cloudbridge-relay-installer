# Threat Detection Integration Guide

## Обзор

Данная документация описывает интеграцию CloudBridge Relay Security с современными инструментами Threat Detection и Hunting из списка Awesome Threat Detection and Hunting.

## 🎯 Интегрированные инструменты

### 1. Sigma Rules
**Описание:** Стандартизированный формат правил детекции для SIEM систем.

**Файл:** `scripts/security/sigma_rules.yaml`

**Правила:**
- SSH Brute Force Attack (T1078)
- PostgreSQL Connection Attempt (T1133)
- Relay Server Port Scan (T1046)
- Data Exfiltration Detection (T1020)
- Privilege Escalation (T1068)
- Suspicious Process Creation (T1059)
- Network Anomaly Detection (T1071)
- File Integrity Monitoring (T1005)
- Docker Container Escape (T1610)
- API Rate Limiting Violation (T1078)

**Использование:**
```bash
# Проверка Sigma Rules
./scripts/security/enhanced_security_monitor.sh --once
```

### 2. YARA Rules
**Описание:** Инструмент для детекции вредоносного ПО и подозрительных файлов.

**Файл:** `scripts/security/yara_rules.yar`

**Категории правил:**
- **Suspicious Scripts** - подозрительные скрипты
- **Cryptominer Detection** - криптомайнеры
- **Backdoor Detection** - бэкдоры
- **Keylogger Detection** - кейлоггеры
- **RAT Detection** - Remote Access Trojans
- **Suspicious Filenames** - подозрительные имена файлов
- **Obfuscated Code** - обфусцированный код
- **Persistence Detection** - механизмы persistence
- **Network Scanner** - сетевые сканеры
- **Data Exfiltration** - утечка данных
- **Anti Analysis** - техники противодействия анализу

**Установка YARA:**
```bash
sudo apt update
sudo apt install -y yara
```

**Использование:**
```bash
# Сканирование файла
yara scripts/security/yara_rules.yar /path/to/file

# Сканирование директории
find /tmp -type f -exec yara scripts/security/yara_rules.yar {} \;
```

### 3. MITRE ATT&CK Framework
**Описание:** Стандартизированная таксономия тактик и техник атак.

**Файл:** `scripts/security/mitre_attack_mapping.yaml`

**Тактики:**
- **Initial Access (TA0001)** - получение доступа
- **Execution (TA0002)** - выполнение
- **Persistence (TA0003)** - закрепление
- **Privilege Escalation (TA0004)** - повышение привилегий
- **Defense Evasion (TA0005)** - уклонение от защиты
- **Credential Access (TA0006)** - доступ к учетным данным
- **Discovery (TA0007)** - разведка
- **Lateral Movement (TA0008)** - горизонтальное перемещение
- **Collection (TA0009)** - сбор данных
- **Command and Control (TA0011)** - командование и управление
- **Exfiltration (TA0010)** - утечка данных
- **Impact (TA0040)** - воздействие

**Корреляционные правила:**
1. **Attack Chain 1:** Initial Access → Execution → Persistence
2. **Attack Chain 2:** Privilege Escalation → Discovery → Lateral Movement
3. **Attack Chain 3:** Collection → Exfiltration → Impact

### 4. Threat Intelligence
**Описание:** Интеграция с внешними сервисами Threat Intelligence.

**Файл:** `scripts/security/threat_intelligence.py`

**Поддерживаемые сервисы:**
- **VirusTotal** - проверка файлов и доменов
- **AbuseIPDB** - репутация IP адресов
- **AlienVault OTX** - открытая база угроз

**Настройка API ключей:**
```python
config = {
    'virustotal_api_key': 'YOUR_VIRUSTOTAL_API_KEY',
    'abuseipdb_api_key': 'YOUR_ABUSEIPDB_API_KEY',
    'alienvault_api_key': 'YOUR_ALIENVAULT_API_KEY',
    'cache_ttl': 3600,  # 1 час
    'timeout': 10,
    'max_retries': 3
}
```

**Использование:**
```bash
# Проверка IP адреса
python3 scripts/security/threat_intelligence.py --check-ip 8.8.8.8

# Проверка домена
python3 scripts/security/threat_intelligence.py --check-domain google.com

# Проверка хеша файла
python3 scripts/security/threat_intelligence.py --check-hash d41d8cd98f00b204e9800998ecf8427e
```

## 🔧 Установка и настройка

### 1. Установка зависимостей
```bash
# Основные инструменты
sudo apt update
sudo apt install -y yara python3-pip curl jq

# Python зависимости
pip3 install requests dataclasses
```

### 2. Настройка переменных окружения
```bash
# Telegram настройки
export TELEGRAM_BOT_TOKEN="YOUR_BOT_TOKEN"
export TELEGRAM_CHAT_ID="YOUR_CHAT_ID"

# Threat Intelligence API ключи
export VIRUSTOTAL_API_KEY="YOUR_VIRUSTOTAL_API_KEY"
export ABUSEIPDB_API_KEY="YOUR_ABUSEIPDB_API_KEY"
export ALIENVAULT_API_KEY="YOUR_ALIENVAULT_API_KEY"
```

### 3. Настройка прав доступа
```bash
# Права на выполнение скриптов
chmod +x scripts/security/enhanced_security_monitor.sh
chmod +x scripts/security/threat_intelligence.py

# Права на логи
mkdir -p logs
chown ubuntu:ubuntu logs/
```

## 🚀 Использование

### Базовая проверка
```bash
# Разовая проверка
./scripts/security/enhanced_security_monitor.sh --once

# Справка
./scripts/security/enhanced_security_monitor.sh --help
```

### Интеграция с cron
```bash
# Добавить в crontab для автоматических проверок каждые 5 минут
*/5 * * * * /home/ubuntu/cloudbridge-relay-installer/scripts/security/enhanced_security_monitor.sh --once
```

### Интеграция с существующей системой
```bash
# Запуск вместе с основным security_monitor.sh
./scripts/security/security_monitor.sh --once
./scripts/security/enhanced_security_monitor.sh --once
```

## 📊 Мониторинг и логирование

### Логи
- **Основной лог:** `logs/enhanced-security-monitor.log`
- **Алерты:** `logs/enhanced-security-alerts.log`

### Уведомления в Telegram
Система отправляет структурированные уведомления с:
- **MITRE ATT&CK техниками**
- **Уровнем угрозы**
- **Детальной информацией**
- **Временными метками**

### Примеры уведомлений

**YARA обнаружение:**
```
🚨 [ALERT] YARA обнаружил угрозу

[FILE] Файл: /tmp/suspicious_file
[RULE] Правило: Suspicious_Scripts
[TIME] Время: 2024-01-01 12:00:00
[SERVER] Сервер: relay-server
```

**MITRE ATT&CK:**
```
🚨 [PRIVILEGE ESCALATION] Обнаружено повышение привилегий

Техника: Process Injection (T1055)
Пользователь: attacker
Процесс: /tmp/backdoor
Время: 2024-01-01 12:00:00
Уровень угрозы: critical

КРИТИЧЕСКАЯ УГРОЗА!
```

**Threat Intelligence:**
```
🚨 [ALERT] Threat Intelligence обнаружил подозрительный IP

[IP] Адрес: 192.168.1.100
[INFO] Информация: Found in 5 threat reports
[TIME] Время: 2024-01-01 12:00:00
[SERVER] Сервер: relay-server
```

## 🔍 Расширенная настройка

### Добавление новых Sigma Rules
1. Создайте новое правило в `scripts/security/sigma_rules.yaml`
2. Добавьте MITRE ATT&CK теги
3. Настройте условия детекции
4. Протестируйте правило

### Добавление новых YARA Rules
1. Создайте новое правило в `scripts/security/yara_rules.yar`
2. Определите строки для поиска
3. Настройте условия
4. Протестируйте на образцах

### Интеграция с дополнительными Threat Intelligence сервисами
1. Добавьте новый метод в `ThreatIntelligence` класс
2. Настройте API ключи
3. Добавьте обработку результатов
4. Протестируйте интеграцию

## 🛡️ Безопасность

### Рекомендации
- **Регулярно обновляйте правила** Sigma и YARA
- **Мониторьте API лимиты** Threat Intelligence сервисов
- **Проверяйте логи** на предмет ошибок
- **Настройте алерты** на критические события
- **Тестируйте систему** на регулярной основе

### Ограничения
- **API лимиты** внешних сервисов
- **Ложные срабатывания** YARA правил
- **Задержки** в получении Threat Intelligence данных
- **Зависимость** от внешних сервисов

## 📈 Производительность

### Оптимизация
- **Кэширование** результатов Threat Intelligence
- **Параллельная обработка** множественных проверок
- **Фильтрация** ложных срабатываний
- **Оптимизация** регулярных выражений

### Мониторинг
- **Время выполнения** проверок
- **Использование памяти** и CPU
- **Количество алертов** и их типы
- **Статистика** обнаружений

## 🔗 Полезные ссылки

- [Awesome Threat Detection and Hunting](https://github.com/0x4D31/awesome-threat-detection)
- [Sigma Rules Repository](https://github.com/SigmaHQ/sigma)
- [YARA Rules Repository](https://github.com/Yara-Rules/rules)
- [MITRE ATT&CK Framework](https://attack.mitre.org/)
- [VirusTotal API](https://developers.virustotal.com/)
- [AbuseIPDB API](https://docs.abuseipdb.com/)
- [AlienVault OTX API](https://otx.alienvault.com/api)

## 📝 Заключение

Интеграция с современными инструментами Threat Detection значительно повышает возможности системы безопасности CloudBridge Relay. Использование стандартизированных форматов (Sigma, YARA) и фреймворков (MITRE ATT&CK) обеспечивает совместимость и масштабируемость решения.

Регулярное обновление правил и мониторинг производительности помогут поддерживать высокий уровень защиты системы. 