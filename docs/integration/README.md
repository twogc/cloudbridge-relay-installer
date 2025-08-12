# Integration - Интеграции

Документация по интеграции CloudBridge Relay с внешними системами.

## 📋 Содержание

### 🔐 Аутентификация и авторизация
- **[Keycloak Integration](keycloak_integration.md)** - Интеграция с Keycloak для аутентификации
- **[Keycloak Multi-Tenancy: внедрение 2025](../reports/tech_task_keycloak_multitenancy.md)** - Практические шаги и API Multi-Tenancy
- **[Keycloak Technical Specification](keycloak_specialist_technical_specification.md)** - Техническая спецификация интеграции с Keycloak
- **[Keycloak Technical Specification Improved](keycloak_specialist_technical_specification_improved.md)** - Улучшенная техническая спецификация Keycloak
- **[LDAP/AD Integration](ldap_ad_integration_technical_specification.md)** - Интеграция с LDAP/Active Directory

### 📊 SIEM Интеграции
- **[SIEM Integration Guide](siem_integration_guide.md)** - Руководство по интеграции с SIEM системами
- **[SIEM Integration Analysis](siem_integration_analysis.md)** - Анализ интеграции с SIEM
- **[SIEM Integration Summary](siem_integration_summary.md)** - Краткое резюме SIEM интеграции
- **[SIEM Migration Guide](siem_migration_guide.md)** - Руководство по миграции SIEM

### 🔗 Клиентские интеграции
- **[Client Integration](client_integration.md)** - Интеграция клиентов с системой
- **[Server Client Implementation](server_client_implementation.md)** - Реализация сервер-клиент архитектуры
- **[Server Client Technical Specification](server_client_technical_specification.md)** - Техническая спецификация сервер-клиент
- **[Server Client Testing](server_client_testing.md)** - Тестирование сервер-клиент интеграции
- **[Server Client Token Guide](server_client_token_guide.md)** - Руководство по работе с токенами

## 🎯 Рекомендуемый порядок изучения

### Для интеграции аутентификации:
1. [Keycloak Integration](keycloak_integration.md) - Основная интеграция
2. [Keycloak Technical Specification](keycloak_specialist_technical_specification.md) - Детали
3. [LDAP/AD Integration](ldap_ad_integration_technical_specification.md) - Если нужен LDAP

### Для SIEM интеграции:
1. [SIEM Integration Guide](siem_integration_guide.md) - Основное руководство
2. [SIEM Integration Analysis](siem_integration_analysis.md) - Анализ
3. [SIEM Migration Guide](siem_migration_guide.md) - Если миграция

### Для разработчиков клиентов:
1. [Client Integration](client_integration.md) - Общая интеграция
2. [Server Client Technical Specification](server_client_technical_specification.md) - Спецификация
3. [Server Client Implementation](server_client_implementation.md) - Реализация
4. [Server Client Testing](server_client_testing.md) - Тестирование

## 🔗 Связанные разделы

- **[Deployment](../deployment/)** - Развертывание интегрированных систем
- **[Security](../security/)** - Безопасность интеграций
- **[Troubleshooting](../troubleshooting/)** - Устранение проблем интеграции 