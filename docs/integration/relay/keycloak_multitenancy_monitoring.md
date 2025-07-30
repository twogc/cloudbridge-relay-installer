# Мониторинг и аудит для Keycloak Multi-Tenancy

## 1. Метрики (Prometheus)
- Количество успешных/неуспешных аутентификаций
- Количество созданных/удалённых арендаторов
- Количество приглашений, членств, ролей
- Время отклика Keycloak и Relay
- Ошибки синхронизации с внешними IDP
- Использование памяти, CPU, задержки GC

### Пример Prometheus-метрик
```
relay_keycloak_auth_success_total{tenant="company-a"}
relay_keycloak_auth_failure_total{tenant="company-a"}
relay_tenant_created_total
relay_tenant_deleted_total
relay_ldap_sync_total{company="company-a",result="success"}
relay_ldap_sync_total{company="company-a",result="failure"}
```

## 2. Алерты (Prometheus)
- Неудачные попытки входа (brute force)
- Ошибки синхронизации LDAP/IDP
- Высокое время отклика Keycloak
- Массовое удаление пользователей

### Пример алерта
```yaml
groups:
  - name: keycloak_tenants
    rules:
      - alert: KeycloakAuthFailures
        expr: increase(relay_keycloak_auth_failure_total[5m]) > 10
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Много неудачных попыток входа в Keycloak"
```

## 3. Grafana
- Дашборды по аутентификации, арендаторам, IDP, ошибкам
- Визуализация метрик по компаниям/tenant
- Дашборды для DevOps и безопасности

## 4. Аудит-логи
- Логируйте все операции с арендаторами, пользователями, ролями, интеграциями
- Храните логи в централизованном хранилище (ELK, Loki, Cloud)
- Настройте ротацию и защиту логов

## 5. Рекомендации по мониторингу
- Включите health checks для Keycloak и Relay
- Настройте алерты на критические события
- Регулярно проверяйте дашборды и логи

## 6. Чеклист мониторинга
- [ ] Метрики Keycloak и Relay доступны в Prometheus
- [ ] Дашборды Grafana настроены
- [ ] Алерты работают и тестируются
- [ ] Аудит-логи собираются и защищены
- [ ] Health checks настроены

---

**Регулярно проверяйте мониторинг и реагируйте на алерты!** 