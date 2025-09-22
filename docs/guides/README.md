# 📖 **Руководства CloudBridge Relay**

## 📋 **Обзор**

Этот раздел содержит практические руководства по настройке, отладке и использованию CloudBridge Relay.

## 📚 **Документы**

### **Руководства по настройке:**
- [**TURN Server Implementation Report**](TURN_SERVER_IMPLEMENTATION_REPORT.md) - Реализация TURN сервера
- [**Reference Repos Setup Report**](REFERENCE_REPOS_SETUP_REPORT.md) - Настройка референсных репозиториев
- [**Reference Repos Battle Testing Ready**](REFERENCE_REPOS_BATTLE_TESTING_READY.md) - Готовность к боевому тестированию
- [**Developer Task MASQUE Debug**](DEVELOPER_TASK_MASQUE_DEBUG.md) - Задача разработчика по отладке MASQUE

## 🎯 **Категории руководств**

### **🔧 Настройка и конфигурация:**
- **TURN сервер** - Настройка для NAT traversal
- **Референсные репозитории** - Подготовка зависимостей
- **Конфигурация** - Настройка параметров
- **Среда разработки** - Подготовка окружения

### **🐛 Отладка и диагностика:**
- **MASQUE отладка** - Диагностика протокола
- **Сетевые проблемы** - Устранение проблем сети
- **Производительность** - Оптимизация и профилирование
- **Логирование** - Анализ логов

### **🧪 Тестирование:**
- **Боевое тестирование** - Нагрузочное тестирование
- **Интеграционные тесты** - Тестирование интеграций
- **Функциональные тесты** - Проверка функций
- **Тесты безопасности** - Проверка безопасности

## 🛠️ **Практические руководства**

### **Настройка TURN сервера:**
```bash
# Установка coturn
sudo apt-get install coturn

# Конфигурация
sudo nano /etc/turnserver.conf

# Запуск сервиса
sudo systemctl start coturn
sudo systemctl enable coturn
```

### **Отладка MASQUE:**
```bash
# Включение debug логов
export LOG_LEVEL=debug

# Проверка подключения
curl -k https://localhost:8443/.well-known/masque/udp/127.0.0.1:8080

# Анализ трафика
tcpdump -i any port 8443
```

### **Настройка референсных репозиториев:**
```bash
# Клонирование репозиториев
git clone https://github.com/quic-go/quic-go.git reference-repos/quic-go
git clone https://github.com/masque-go/masque-go.git reference-repos/masque-go

# Установка зависимостей
go mod download
go mod tidy
```

## 🔍 **Диагностика проблем**

### **Частые проблемы:**

1. **MASQUE не работает:**
   ```bash
   # Проверить порт 8443
   netstat -tlnp | grep 8443
   
   # Проверить сертификаты
   openssl s_client -connect localhost:8443
   ```

2. **TURN сервер недоступен:**
   ```bash
   # Проверить статус
   sudo systemctl status coturn
   
   # Проверить конфигурацию
   sudo turnserver --listening-port=3478 --external-ip=YOUR_IP
   ```

3. **QUIC соединения не устанавливаются:**
   ```bash
   # Проверить UDP порты
   netstat -ulnp | grep 5555
   
   # Проверить firewall
   sudo ufw status | grep 5555
   ```

## 📊 **Мониторинг и метрики**

### **Ключевые метрики для отслеживания:**
- **MASQUE connections** - Количество MASQUE соединений
- **TURN allocations** - TURN аллокации
- **QUIC streams** - QUIC потоки
- **NAT traversal success rate** - Успешность NAT traversal

### **Команды мониторинга:**
```bash
# Мониторинг MASQUE
curl -s http://localhost:5551/metrics | grep masque

# Мониторинг TURN
sudo journalctl -u coturn -f

# Мониторинг QUIC
ss -ulnp | grep 5555
```

## 🧪 **Тестирование**

### **Функциональные тесты:**
```bash
# Тест MASQUE
./test-masque.sh

# Тест TURN
./test-turn.sh

# Тест QUIC
./test-quic.sh
```

### **Нагрузочные тесты:**
```bash
# Нагрузочный тест
./load-test.sh --connections=1000 --duration=60s

# Тест производительности
./perf-test.sh --bandwidth=100Mbps
```

## 📝 **Чек-листы**

### **✅ Настройка TURN:**
- [ ] coturn установлен
- [ ] Конфигурация корректна
- [ ] Сервис запущен
- [ ] Порт 3478 открыт
- [ ] Тесты проходят

### **✅ Настройка MASQUE:**
- [ ] Сертификаты настроены
- [ ] Порт 8443 открыт
- [ ] Протокол работает
- [ ] Логи корректны
- [ ] Метрики собираются

### **✅ Референсные репозитории:**
- [ ] Репозитории клонированы
- [ ] Зависимости установлены
- [ ] Тесты проходят
- [ ] Документация изучена
- [ ] Интеграция работает

---

**📖 Руководства CloudBridge Relay помогут в настройке и отладке!**
