# 📋 Техническое задание: Исправление форматирования Telegram сообщений

## 🎯 Цель
Исправить форматирование сообщений в Telegram, чтобы символы `\n` корректно отображались как переносы строк вместо текста.

## 🐛 Описание проблемы

### **Текущее состояние:**
Сообщения в Telegram отображаются с литеральными символами `\n` вместо переносов строк:

```
🚨 ALERT: Обнаружены подозрительные процессы\n\n📋 Процесс: python\n🔢 Количество: 4\n📝 Детали:\nroot         852  0.0  0.1 110144 13360 ?        Ssl  Jul08   0:00 /usr/bin/python3 /usr/share/unattended-upgrades/unattended-upgrade-shutdown --wait-for-signal
root      233324  0.0  0.2 548880 20072 ?        Ssl  Jul12  18:33 /usr/bin/python3 /usr/bin/fail2ban-server -xf start
root     2976624  0.0  0.2  33176 16324 ?        Ss   Jul25   0:00 /usr/bin/python3 /usr/bin/networkd-dispatcher --run-startup-triggers\n🎯 MITRE ATT
```

### **Ожидаемое состояние:**
Сообщения должны отображаться с правильными переносами строк:

```
🚨 ALERT: Обнаружены подозрительные процессы

📋 Процесс: python
🔢 Количество: 4
📝 Детали:
root         852  0.0  0.1 110144 13360 ?        Ssl  Jul08   0:00 /usr/bin/python3 /usr/share/unattended-upgrades/unattended-upgrade-shutdown --wait-for-signal
root      233324  0.0  0.2 548880 20072 ?        Ssl  Jul12  18:33 /usr/bin/python3 /usr/bin/fail2ban-server -xf start
root     2976624  0.0  0.2  33176 16324 ?        Ss   Jul25   0:00 /usr/bin/python3 /usr/bin/networkd-dispatcher --run-startup-triggers
🎯 MITRE ATT&CK: T1059 (Command and Scripting Interpreter)
⏰ Время: 2024-07-27 12:34:56
🖥️ Сервер: 2gc-relay-server
```

---

## 🔧 Технические требования

### **1. Файл для исправления:**
```
scripts/security/enhanced_security_monitor.sh
```

### **2. Функция для модификации:**
```bash
send_telegram_message() {
    local message="$1"
    local parse_mode="${2:-HTML}"
    
    # ТЕКУЩИЙ КОД НУЖДАЕТСЯ В ИСПРАВЛЕНИИ
    if [[ "$parse_mode" == "HTML" ]]; then
        message=$(echo "$message" | sed 's/&/\&amp;/g')
    fi
    
    local response=$(curl -s -X POST \
        "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
        -d "chat_id=${TELEGRAM_CHAT_ID}" \
        -d "text=${message}" \
        -d "parse_mode=${parse_mode}" \
        2>/dev/null)
    
    if echo "$response" | grep -q '"ok":true'; then
        log_info "Сообщение отправлено в Telegram"
    else
        log_error "Ошибка отправки в Telegram: $response"
    fi
}
```

### **3. Требуемое исправление:**
Заменить строку:
```bash
-d "text=${message}" \
```

На:
```bash
-d "text=${formatted_message}" \
```

И добавить перед curl запросом:
```bash
# Используем printf для правильной обработки \n
local formatted_message=$(printf '%b' "$message")
```

---

## 📝 Полное решение

### **Исправленная функция `send_telegram_message`:**
```bash
send_telegram_message() {
    local message="$1"
    local parse_mode="${2:-HTML}"
    
    if [[ "$parse_mode" == "HTML" ]]; then
        # Экранируем только амперсанд для HTML-форматирования
        message=$(echo "$message" | sed 's/&/\&amp;/g')
    fi
    
    # ИСПРАВЛЕНИЕ: Используем printf для правильной обработки \n
    local formatted_message=$(printf '%b' "$message")
    
    local response=$(curl -s -X POST \
        "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
        -d "chat_id=${TELEGRAM_CHAT_ID}" \
        -d "text=${formatted_message}" \
        -d "parse_mode=${parse_mode}" \
        2>/dev/null)
    
    if echo "$response" | grep -q '"ok":true'; then
        log_info "Сообщение отправлено в Telegram"
    else
        log_error "Ошибка отправки в Telegram: $response"
    fi
}
```

---

## 🧪 Тестирование

### **1. Команда для тестирования:**
```bash
./scripts/security/enhanced_security_monitor.sh --once
```

### **2. Ожидаемый результат:**
- ✅ Сообщения отправляются без ошибок
- ✅ Символы `\n` отображаются как переносы строк
- ✅ Форматирование читаемое и структурированное

### **3. Проверка в Telegram:**
Сообщения должны отображаться с правильными переносами строк, а не с литеральными символами `\n`.

---

## 📋 Критерии приемки

### **✅ Успешное выполнение:**
1. **Функциональность:** Все сообщения отправляются без ошибок
2. **Форматирование:** Символы `\n` корректно отображаются как переносы строк
3. **Читаемость:** Сообщения структурированы и легко читаются
4. **Совместимость:** Работает со всеми типами сообщений (процессы, сеть, файлы, целостность)

### **❌ Критерии отклонения:**
1. **Ошибки отправки:** Сообщения не отправляются или возвращают ошибки
2. **Неправильное форматирование:** Символы `\n` отображаются как текст
3. **Нарушение функциональности:** Другие функции перестали работать

---

## 🔍 Дополнительная информация

### **Почему это происходит:**
- Bash по умолчанию не интерпретирует escape-последовательности в переменных
- `printf '%b'` правильно обрабатывает escape-последовательности
- Telegram API ожидает уже обработанные символы переноса строк

### **Альтернативные решения:**
1. **Использование `echo -e`:** `formatted_message=$(echo -e "$message")`
2. **Использование `printf '%b'`:** `formatted_message=$(printf '%b' "$message")` ✅ (рекомендуется)
3. **Прямая замена:** `message=$(echo "$message" | sed 's/\\n/\n/g')`

### **Рекомендуемое решение:**
Использовать `printf '%b'` как наиболее надежный способ обработки escape-последовательностей в Bash.

---

## 📞 Контакты

**Ответственный:** Разработчик системы безопасности  
**Приоритет:** Высокий  
**Срок выполнения:** 1 день  
**Статус:** В работе

---

**📋 Статус:** Требует исправления  
**📅 Дата:** 27 июля 2024  
**👨‍💻 Приоритет:** Высокий 