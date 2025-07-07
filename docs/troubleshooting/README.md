# Troubleshooting - Устранение неполадок

Документация по устранению неполадок в CloudBridge Relay.

## 📋 Содержание

### 🔧 Основная отладка
- **[Debugging](debugging.md)** - Основное руководство по отладке системы

### 🛡️ Проблемы безопасности
- **[Fail2ban Troubleshooting](fail2ban_troubleshooting.md)** - Устранение проблем с Fail2ban
- **[Blackbox TLS Fix](blackbox_tls_fix.md)** - Исправление проблем с TLS в Blackbox Exporter

## 🎯 Рекомендуемый порядок изучения

### При возникновении проблем:
1. [Debugging](debugging.md) - Начните с общего руководства по отладке
2. [Fail2ban Troubleshooting](fail2ban_troubleshooting.md) - Если проблемы с блокировкой
3. [Blackbox TLS Fix](blackbox_tls_fix.md) - Если проблемы с TLS мониторингом

## 🔍 Типичные проблемы

### Проблемы с подключением:
- Проверьте [Debugging](debugging.md) - раздел "Network Issues"
- Убедитесь в правильности конфигурации портов

### Проблемы с безопасностью:
- Изучите [Fail2ban Troubleshooting](fail2ban_troubleshooting.md)
- Проверьте логи безопасности

### Проблемы с мониторингом:
- См. [Blackbox TLS Fix](blackbox_tls_fix.md)
- Проверьте конфигурацию TLS сертификатов

## 🔗 Связанные разделы

- **[Security](../security/)** - Проблемы безопасности
- **[Integration](../integration/)** - Проблемы интеграции
- **[Reports](../reports/)** - Отчеты о проблемах 