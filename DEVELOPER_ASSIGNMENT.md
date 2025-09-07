# Задание для разработчика: Анализ и исправление ошибок CloudBridge Relay

## 📋 Обзор проекта

**CloudBridge Relay** - это система для ретрансляции данных между облачными сервисами, развернутая в Kubernetes. Проект включает в себя несколько компонентов: CloudBridge Relay, Cassandra, Cadence, Prometheus и Grafana.

## 🎯 Основная задача

Провести полный анализ проекта, ознакомиться с документацией и исправить критические ошибки, препятствующие корректной работе системы.

## 📚 Документация для изучения

### Обязательные файлы для ознакомления:
1. **`DEVELOPER_TASKS.md`** - Подробное техническое задание с описанием проблем
2. **`STATUS_REPORT.md`** - Отчет о текущем состоянии системы
3. **`CADENCE_PROBLEM_REPORT.md`** - Детальный анализ проблем с Cadence
4. **`FINAL_STATUS_REPORT.md`** - Итоговый статус системы

### Дополнительные файлы:
- **`README.md`** - Общая информация о проекте
- **`terraform/`** - Инфраструктурный код
- **`k8s/`** - Kubernetes манифесты

## 🔍 Текущее состояние системы

### ✅ Работающие компоненты (90% готовности):
- **Prometheus** - исправлен, работает стабильно
- **Grafana** - исправлен, работает стабильно  
- **Cassandra** - работает стабильно (1 под, 20Gi PV)
- **CloudBridge Relay** - работает стабильно (3 пода)

### ❌ Проблемные компоненты:
- **Cadence** - 2 пода в CrashLoopBackOff с ошибкой `Ringpop.Name: zero value`

## 🚨 Критические проблемы для исправления

### 1. Cadence - Основная проблема
**Ошибка**: `Config file corrupted.Ringpop.Name: zero value`

**Контекст**:
- Cadence использует утилиту `dockerize` для создания конфигурационного файла
- Реализовано решение с `emptyDir + subPath` для монтирования конфигурации
- Конфигурация загружается, но поле `ringpop.name` не читается корректно

**Файлы для анализа**:
- `k8s/cadence/configmap.yaml` - основная конфигурация
- `k8s/cadence/configmap-minimal.yaml` - упрощенная конфигурация
- `k8s/cadence/cadence-dynamicconfig.yaml` - динамическая конфигурация
- `k8s/deployments/cadence-deployment-final.yaml` - deployment

**Возможные причины**:
1. Проблема с YAML синтаксисом в ConfigMap
2. Неправильное монтирование файлов через subPath
3. Конфликт между переменными окружения и файлом конфигурации
4. Несовместимость версии образа Cadence с конфигурацией

## 🛠️ План действий для разработчика

### Этап 1: Анализ (1-2 часа)
1. **Изучить документацию**:
   - Прочитать все файлы из списка выше
   - Понять архитектуру системы
   - Изучить историю изменений и попыток исправления

2. **Проанализировать текущее состояние**:
   ```bash
   kubectl get pods -n cloudbridge
   kubectl logs -l app.kubernetes.io/name=cadence -n cloudbridge --tail=50
   kubectl describe deployment cadence -n cloudbridge
   ```

3. **Проверить конфигурацию**:
   ```bash
   kubectl get configmap cadence-config -n cloudbridge -o yaml
   kubectl get configmap cadence-config-minimal -n cloudbridge -o yaml
   ```

### Этап 2: Диагностика (2-3 часа)
1. **Проверить монтирование файлов**:
   - Убедиться, что `config_template.yaml` правильно монтируется через subPath
   - Проверить права доступа к файлам
   - Проверить содержимое файлов внутри пода

2. **Проанализировать логи Cadence**:
   - Найти точную причину ошибки `Ringpop.Name: zero value`
   - Проверить, создается ли файл `docker.yaml`
   - Убедиться, что `dockerize` работает корректно

3. **Проверить совместимость**:
   - Изучить документацию Cadence для версии `ubercadence/server:latest`
   - Проверить формат конфигурации
   - Убедиться в корректности переменных окружения

### Этап 3: Исправление (3-4 часа)
1. **Исправить конфигурацию Cadence**:
   - Создать корректный ConfigMap с правильным YAML синтаксисом
   - Убедиться, что все обязательные поля присутствуют
   - Проверить совместимость с версией образа

2. **Оптимизировать deployment**:
   - Убрать конфликтующие переменные окружения
   - Настроить правильное монтирование volumes
   - Добавить необходимые health checks

3. **Протестировать решение**:
   - Применить изменения
   - Проверить запуск подов
   - Убедиться в стабильной работе

### Этап 4: Валидация (1 час)
1. **Проверить все сервисы**:
   ```bash
   kubectl get pods -n cloudbridge
   kubectl get services -n cloudbridge
   ```

2. **Протестировать функциональность**:
   - Проверить доступность Prometheus
   - Проверить доступность Grafana
   - Проверить работу Cadence
   - Проверить подключение к Cassandra

## 🔧 Команды для работы

### Основные команды диагностики:
```bash
# Статус всех подов
kubectl get pods -n cloudbridge -o wide

# Логи Cadence
kubectl logs -l app.kubernetes.io/name=cadence -n cloudbridge --tail=100 -f

# Описание deployment
kubectl describe deployment cadence -n cloudbridge

# Проверка ConfigMap
kubectl get configmap -n cloudbridge | grep cadence

# Проверка volumes
kubectl describe pod <pod-name> -n cloudbridge | grep -A 20 "Volumes:"
```

### Команды для применения изменений:
```bash
# Применить ConfigMap
kubectl apply -f k8s/cadence/configmap.yaml

# Применить deployment
kubectl apply -f k8s/deployments/cadence-deployment-final.yaml

# Перезапустить deployment
kubectl rollout restart deploy/cadence -n cloudbridge

# Проверить статус rollout
kubectl rollout status deploy/cadence -n cloudbridge
```

## 📝 Ожидаемый результат

После выполнения задания система должна достичь **100% готовности**:

- ✅ **Prometheus** - работает
- ✅ **Grafana** - работает  
- ✅ **Cassandra** - работает
- ✅ **CloudBridge Relay** - работает
- ✅ **Cadence** - работает стабильно (2 пода в состоянии Running)

## 🚀 Дополнительные рекомендации

1. **Версионирование**: Зафиксировать версии образов вместо `:latest`
2. **Мониторинг**: Настроить алерты для всех компонентов
3. **Безопасность**: Проверить RBAC и network policies
4. **Документация**: Обновить README с инструкциями по развертыванию

## 📞 Поддержка

При возникновении вопросов или проблем:
1. Изучите существующую документацию
2. Проверьте логи и статус ресурсов
3. Обратитесь к команде DevOps за консультацией

---

**Время выполнения**: 6-10 часов  
**Приоритет**: Высокий  
**Статус**: В работе
