# CloudBridge Relay - Project Structure

## 📁 Root Files
- `MINIKUBE_BUILD_GUIDE.md` - Основное руководство по сборке в Minikube
- `DEVELOPER_BRIEFING.md` - Техническое описание проекта для разработчиков

## 📁 Directory Structure

### `/docs/` - Документация
- `reports/` - Отчеты по тестированию, диагностике и статусу
- `guides/` - Руководства пользователя и инструкции
- `analysis/` - Технические анализы и исследования
- `security/` - Документация по безопасности и аудиты
- `deployment/` - Инструкции по развертыванию
- `troubleshooting/` - Руководства по устранению неполадок

### `/archive/` - Архивные файлы
- `scripts/` - Старые скрипты и утилиты
- `configs/` - Старые конфигурационные файлы
- `old-reports/` - Устаревшие отчеты и логи

### `/docker/` - Docker конфигурации
- `dockerfiles/` - Все Dockerfile'ы проекта

### Основные папки проекта
- `cmd/` - Исполняемые файлы
- `internal/` - Внутренняя логика приложения
- `pkg/` - Публичные пакеты
- `k8s/` - Kubernetes манифесты
- `test/` - Тесты и тестовые клиенты
- `appliance/` - Appliance решения

## 🚀 Quick Start
Для начала работы см. `MINIKUBE_BUILD_GUIDE.md`

## 📖 Developer Info
Для технических деталей см. `DEVELOPER_BRIEFING.md`
