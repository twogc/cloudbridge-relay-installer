# Информация для DevOps: Настройка Zabbix Server для CloudBridge Relay

## 📋 Обзор

Данный документ содержит всю необходимую информацию для настройки Zabbix Server для мониторинга CloudBridge Relay системы.

## 🎯 Цель

Настроить мониторинг CloudBridge Relay Server с AI интеграцией через Zabbix Server на 10.0.0.241.

## 📊 Информация о мониторируемом сервере

### Основные данные:
- **Hostname**: `CloudBridge-Relay-Server`
- **IP адрес**: 10.0.0.249
- **Операционная система**: Ubuntu 22.04 LTS
- **Zabbix Agent версия**: 6.0.41
- **Порт агента**: 10050

### Безопасность:
- **Шифрование**: TLS ОТКЛЮЧЕН (unencrypted)
- **Режим**: Открытое соединение без шифрования

## 🔧 Настройка хоста в Zabbix

### 1. Создание хоста

**В веб-интерфейсе Zabbix:**
- Перейти в `Configuration → Hosts → Create host`
- **Host name**: `CloudBridge-Relay-Server`
- **Groups**: `CloudBridge` (создать группу если не существует)
- **IP address**: 10.0.0.249
- **Port**: `10050`

### 2. Настройка безопасности

**В разделе Encryption:**
- **Connections to host**: `No encryption`
- **Connections from host**: `No encryption`
- **Примечание**: TLS отключен на агенте

### 3. Настройка интерфейсов

**Agent interface:**
- **IP address**: 10.0.0.249
- **DNS name**: `CloudBridge-Relay-Server`
- **Port**: `10050`
- **Default**: ✓

## 📈 Доступные метрики

### Relay Server метрики:
```
cloudbridge.relay.connections     - активные соединения
cloudbridge.relay.uptime          - время работы в секундах
cloudbridge.relay.memory          - использование памяти в MB
cloudbridge.relay.goroutines      - количество горутин
```

### AI Service метрики:
```
cloudbridge.ai.status             - статус AI сервиса
cloudbridge.ai.detections         - количество обнаруженных аномалий
cloudbridge.ai.predictions        - количество предсказаний нагрузки
cloudbridge.ai.optimizations      - количество оптимизаций маршрутов
```

### Системные метрики:
```
cloudbridge.system.cpu.usage      - загрузка CPU
cloudbridge.system.memory.usage   - использование памяти в %
cloudbridge.system.disk.usage     - использование диска в %
cloudbridge.system.network.connections - количество сетевых соединений
```

### Docker метрики:
```
cloudbridge.docker.containers.running - запущенные контейнеры
cloudbridge.docker.containers.total   - общее количество контейнеров
cloudbridge.docker.images.total       - количество Docker образов
```

### Kubernetes метрики:
```
cloudbridge.k8s.pods.running      - запущенные поды
cloudbridge.k8s.pods.total        - общее количество подов
cloudbridge.k8s.nodes.ready       - готовые узлы
```

## 🚨 Рекомендуемые алерты

### Критические алерты:
1. **CloudBridge Relay Server недоступен**
   - Условие: `cloudbridge.relay.connections` = 0 в течение 5 минут
   - Приоритет: High

2. **AI Service недоступен**
   - Условие: `cloudbridge.ai.status` ≠ "running"
   - Приоритет: High

3. **Высокая загрузка CPU**
   - Условие: `cloudbridge.system.cpu.usage` > 80%
   - Приоритет: High

4. **Нехватка памяти**
   - Условие: `cloudbridge.system.memory.usage` > 90%
   - Приоритет: High

5. **Заполнение диска**
   - Условие: `cloudbridge.system.disk.usage` > 85%
   - Приоритет: High

### Предупреждающие алерты:
1. **Средняя загрузка CPU**
   - Условие: `cloudbridge.system.cpu.usage` > 60%
   - Приоритет: Warning

2. **Среднее использование памяти**
   - Условие: `cloudbridge.system.memory.usage` > 70%
   - Приоритет: Warning

3. **Среднее заполнение диска**
   - Условие: `cloudbridge.system.disk.usage` > 70%
   - Приоритет: Warning

## 📊 Рекомендуемые дашборды

### 1. CloudBridge Overview
- Общий статус системы
- Активные соединения
- Время работы
- Статус AI сервиса

### 2. AI Performance
- Количество обнаруженных аномалий
- Количество предсказаний
- Количество оптимизаций
- Статус AI моделей

### 3. System Resources
- Загрузка CPU
- Использование памяти
- Использование диска
- Сетевые соединения

### 4. Container Status
- Запущенные контейнеры
- Docker образы
- Kubernetes поды
- Статус узлов

## 🔍 Тестирование подключения

### Команды для проверки:

```bash
# Проверка доступности агента
telnet [IP_CloudBridge_Server] 10050

# Проверка метрик (с CloudBridge сервера)
sudo zabbix_agent2 -t cloudbridge.relay.connections
sudo zabbix_agent2 -t cloudbridge.ai.status
sudo zabbix_agent2 -t cloudbridge.system.cpu.usage
```

### Ожидаемые результаты:
- `cloudbridge.relay.connections`: числовое значение
- `cloudbridge.ai.status`: "running" или "unknown"
- `cloudbridge.system.cpu.usage`: значение от 0 до 100

## 📝 Шаблоны для импорта

### Создать шаблон "CloudBridge Relay Template":

```xml
<?xml version="1.0" encoding="UTF-8"?>
<zabbix_export>
    <version>6.0</version>
    <date>2025-09-14T00:00:00Z</date>
    <groups>
        <group>
            <uuid>cloudbridge-group-uuid</uuid>
            <name>CloudBridge</name>
        </group>
    </groups>
    <templates>
        <template>
            <uuid>cloudbridge-template-uuid</uuid>
            <template>CloudBridge Relay Template</template>
            <name>CloudBridge Relay Template</name>
            <description>Template for monitoring CloudBridge Relay Server with AI integration</description>
            <groups>
                <group>
                    <name>CloudBridge</name>
                </group>
            </groups>
            <items>
                <!-- Relay Server Items -->
                <item>
                    <uuid>relay-connections-uuid</uuid>
                    <name>Relay Active Connections</name>
                    <type>ZABBIX_AGENT</type>
                    <key>cloudbridge.relay.connections</key>
                    <history>7d</history>
                    <trends>365d</trends>
                    <value_type>UINT64</value_type>
                    <units>connections</units>
                </item>
                <item>
                    <uuid>relay-uptime-uuid</uuid>
                    <name>Relay Uptime</name>
                    <type>ZABBIX_AGENT</type>
                    <key>cloudbridge.relay.uptime</key>
                    <history>7d</history>
                    <trends>365d</trends>
                    <value_type>FLOAT</value_type>
                    <units>seconds</units>
                </item>
                <!-- AI Service Items -->
                <item>
                    <uuid>ai-status-uuid</uuid>
                    <name>AI Service Status</name>
                    <type>ZABBIX_AGENT</type>
                    <key>cloudbridge.ai.status</key>
                    <history>7d</history>
                    <trends>365d</trends>
                    <value_type>TEXT</value_type>
                </item>
                <!-- System Items -->
                <item>
                    <uuid>system-cpu-uuid</uuid>
                    <name>System CPU Usage</name>
                    <type>ZABBIX_AGENT</type>
                    <key>cloudbridge.system.cpu.usage</key>
                    <history>7d</history>
                    <trends>365d</trends>
                    <value_type>FLOAT</value_type>
                    <units>%</units>
                </item>
            </items>
            <triggers>
                <!-- Critical Triggers -->
                <trigger>
                    <uuid>relay-down-uuid</uuid>
                    <expression>last(/CloudBridge Relay Template/cloudbridge.relay.connections)=0</expression>
                    <name>CloudBridge Relay Server is down</name>
                    <priority>HIGH</priority>
                    <description>CloudBridge Relay Server has no active connections</description>
                </trigger>
                <trigger>
                    <uuid>ai-down-uuid</uuid>
                    <expression>last(/CloudBridge Relay Template/cloudbridge.ai.status)<>"running"</expression>
                    <name>AI Service is down</name>
                    <priority>HIGH</priority>
                    <description>AI Service is not running</description>
                </trigger>
                <trigger>
                    <uuid>high-cpu-uuid</uuid>
                    <expression>last(/CloudBridge Relay Template/cloudbridge.system.cpu.usage)>80</expression>
                    <name>High CPU usage</name>
                    <priority>HIGH</priority>
                    <description>CPU usage is above 80%</description>
                </trigger>
            </triggers>
        </template>
    </templates>
</zabbix_export>
```

## 🔧 Дополнительные настройки

### 1. Настройка уведомлений
- Email: настроить SMTP сервер
- Slack: добавить webhook URL
- Telegram: настроить bot token
- SMS: настроить SMS gateway

### 2. Настройка резервного копирования
- Ежедневные бэкапы базы данных
- Резервное копирование конфигурации
- Тестирование восстановления

### 3. Настройка производительности
- Оптимизация базы данных
- Настройка кэширования
- Мониторинг производительности Zabbix

## 📞 Контакты и поддержка

### Ответственные:
- **CloudBridge Team**: cloudbridge@company.com
- **DevOps Team**: devops@company.com
- **AI Team**: ai@company.com

### Документация:
- [Zabbix Official Documentation](https://www.zabbix.com/documentation)
- [CloudBridge AI Integration Guide](AI_IMPLEMENTATION_GUIDE.md)
- [Zabbix Installation Task](DEVELOPER_TASK_ZABBIX_INSTALLATION.md)

## 🚀 Заключение

После выполнения всех настроек у вас будет полностью функциональный мониторинг CloudBridge Relay системы с AI интеграцией. Система будет автоматически отслеживать:

1. **Состояние Relay Server** - доступность и производительность
2. **AI Service** - статус и метрики AI компонентов
3. **Системные ресурсы** - CPU, память, диск, сеть
4. **Контейнеры** - Docker и Kubernetes метрики
5. **Автоматические алерты** - уведомления о проблемах

Все метрики будут собираться в реальном времени и доступны через веб-интерфейс Zabbix и Grafana дашборды.

---
**Дата создания**: 2025-09-14
**Версия**: 1.0
**Статус**: Готов к настройке
