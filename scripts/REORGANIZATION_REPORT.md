# Отчет о реорганизации папки scripts

## Выполненные изменения

### 1. Создана новая структура папок

```
scripts/
├── setup/           # Скрипты установки и настройки (15 файлов)
├── security/        # Скрипты безопасности и аудита (7 файлов)
├── maintenance/     # Скрипты обслуживания и мониторинга (9 файлов)
├── testing/         # Скрипты тестирования (4 файла)
├── utils/           # Утилитарные скрипты (8 файлов)
├── siem/            # Скрипты SIEM интеграции (2 файла)
└── README.md        # Обновленная документация
```

### 2. Распределение файлов по категориям

#### 📦 Setup (15 файлов)
- `check_dependencies.sh` - Проверка зависимостей
- `generate_secrets.sh` - Генерация секретов
- `interactive_setup.sh` - Интерактивная установка
- `setup_certs.sh` - Настройка сертификатов
- `setup_dashboards.sh` - Установка дашбордов
- `setup_env.sh` - Настройка окружения
- `setup_fail2ban_logs.sh` - Настройка fail2ban
- `setup_firewall.sh` - Настройка файрвола
- `setup_keycloak.sh` - Настройка Keycloak
- `setup_ldap_integration.sh` - LDAP интеграция
- `setup_lynis.sh` - Установка Lynis
- `setup_moscow_region.sh` - Настройка для Москвы
- `setup_security_monitor.sh` - Мониторинг безопасности
- `setup_security_monitor_cron.sh` - Cron для мониторинга
- `setup_system.sh` - Базовая настройка системы
- `show_banner.sh` - Отображение баннера

#### 🔒 Security (7 файлов)
- `lynis_quick_test.sh` - Быстрый тест Lynis
- `lynis_with_telegram.sh` - Lynis с Telegram
- `quick_audit.sh` - Быстрый аудит
- `relay-security-audit` - Основной аудит безопасности
- `security_check.sh` - Проверка безопасности
- `security_hardening.sh` - Усиление безопасности
- `security_monitor.sh` - Мониторинг безопасности
- `security_monitor_telegram.sh` - Мониторинг с Telegram

#### 🔧 Maintenance (9 файлов)
- `backup_keycloak.sh` - Резервное копирование Keycloak
- `fix_blackbox_tls.sh` - Исправление TLS в Blackbox
- `manage_logs.sh` - Управление логами
- `renew_certs.sh` - Обновление сертификатов
- `restore_certs.sh` - Восстановление сертификатов
- `update_alertmanager_email.sh` - Обновление email Alertmanager
- `update_certs.sh` - Обновление сертификатов
- `update_keycloak_public_key.sh` - Обновление ключа Keycloak
- `update_keycloak_public_key_from_file.sh` - Обновление ключа из файла

#### 🧪 Testing (4 файла)
- `run_all_tests.sh` - Запуск всех тестов
- `test_integration.sh` - Интеграционные тесты
- `test_tls_connections.sh` - Тесты TLS соединений
- `test_vpn_client.sh` - Тесты VPN клиента

#### 🛠️ Utils (8 файлов)
- `create_real_connections.sh` - Создание реальных соединений
- `create_users.sh` - Создание пользователей
- `generate_business_metrics.sh` - Генерация бизнес-метрик
- `install_systemd.sh` - Установка systemd сервисов
- `post_install.sh` - Пост-установочные задачи
- `relay_start_wrapper.sh` - Обертка для запуска relay
- `start_production.sh` - Запуск production
- `start_services.sh` - Запуск сервисов

#### 📊 SIEM (2 файла)
- `deploy_siem_server.sh` - Развертывание SIEM сервера
- `integrate_siem.sh` - Интеграция с SIEM

### 3. Обновленные связи между скриптами

#### Основные изменения путей:
- `scripts/generate_secrets.sh` → `scripts/setup/generate_secrets.sh`
- `scripts/start_production.sh` → `scripts/utils/start_production.sh`
- `scripts/setup_dashboards.sh` → `scripts/setup/setup_dashboards.sh`
- `scripts/restore_certs.sh` → `scripts/maintenance/restore_certs.sh`
- `scripts/update_keycloak_public_key_from_file.sh` → `scripts/maintenance/update_keycloak_public_key_from_file.sh`
- `scripts/create_users.sh` → `scripts/utils/create_users.sh`
- `scripts/security_monitor.sh` → `scripts/security/security_monitor.sh`
- `scripts/lynis_with_telegram.sh` → `scripts/security/lynis_with_telegram.sh`

#### Обновленные файлы:
1. `setup_cloudbridge.sh` - основной скрипт установки
2. `start_production.sh` - скрипт запуска production
3. `scripts/utils/start_production.sh` - копия start_production.sh
4. `scripts/setup/interactive_setup.sh` - интерактивная установка
5. `scripts/setup/setup_env.sh` - настройка окружения
6. `scripts/setup/setup_certs.sh` - настройка сертификатов
7. `scripts/testing/run_all_tests.sh` - запуск тестов
8. `scripts/utils/post_install.sh` - пост-установка
9. `scripts/utils/create_users.sh` - создание пользователей
10. `scripts/utils/relay_start_wrapper.sh` - обертка запуска
11. `scripts/utils/install_systemd.sh` - установка systemd
12. `scripts/maintenance/renew_certs.sh` - обновление сертификатов
13. `scripts/maintenance/update_certs.sh` - обновление сертификатов
14. `scripts/setup/setup_security_monitor.sh` - настройка мониторинга
15. `scripts/setup/setup_security_monitor_cron.sh` - cron мониторинга
16. `scripts/setup/setup_lynis.sh` - установка Lynis

### 4. Созданная документация

#### README файлы:
- `scripts/README.md` - Общий обзор всех скриптов
- `scripts/setup/README.md` - Документация по скриптам установки
- `scripts/security/README.md` - Документация по скриптам безопасности
- `scripts/maintenance/README.md` - Документация по скриптам обслуживания
- `scripts/testing/README.md` - Документация по скриптам тестирования
- `scripts/utils/README.md` - Документация по утилитарным скриптам
- `scripts/siem/README.md` - Документация по SIEM интеграции

### 5. Преимущества новой структуры

1. **Логическая организация** - скрипты сгруппированы по функциональности
2. **Удобство навигации** - легко найти нужный скрипт
3. **Улучшенная документация** - каждый раздел имеет свой README
4. **Масштабируемость** - легко добавлять новые скрипты в соответствующие папки
5. **Поддержка** - четкое разделение ответственности между скриптами

### 6. Рекомендации по использованию

#### Для разработчиков:
- Добавляйте новые скрипты в соответствующие папки
- Обновляйте README файлы при добавлении новых скриптов
- Используйте относительные пути при ссылках между скриптами

#### Для администраторов:
- Используйте `./setup_cloudbridge.sh` для полной установки
- Используйте `./scripts/setup/interactive_setup.sh` для интерактивной установки
- Регулярно запускайте скрипты из папки `maintenance/`
- Настройте cron задачи для автоматизации

### 7. Автоматизация

Рекомендуемые cron задачи:
```bash
# Ежедневные задачи
0 1 * * * /path/to/scripts/maintenance/backup_keycloak.sh
0 2 * * * /path/to/scripts/security/relay-security-audit
0 4 * * * /path/to/scripts/testing/run_all_tests.sh
0 5 * * * /path/to/scripts/utils/generate_business_metrics.sh

# Еженедельные задачи
0 3 * * 0 /path/to/scripts/maintenance/manage_logs.sh

# Ежемесячные задачи
0 2 1 * * /path/to/scripts/maintenance/renew_certs.sh
```

## Заключение

Реорганизация папки `scripts` завершена успешно. Все скрипты логически сгруппированы, связи между ними обновлены, создана подробная документация. Новая структура обеспечивает лучшую организацию, удобство использования и масштабируемость проекта. 