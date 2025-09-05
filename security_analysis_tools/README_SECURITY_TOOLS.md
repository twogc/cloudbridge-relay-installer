# 🔒 Инструменты анализа безопасности

Комплексный фреймворк для безопасного анализа уязвимостей веб-приложений в контролируемой среде.

## ⚠️ ВАЖНОЕ ПРЕДУПРЕЖДЕНИЕ

**Эти инструменты предназначены ТОЛЬКО для:**
- Обучения принципам безопасности
- Тестирования систем, которыми вы владеете
- Аудита систем с явного разрешения владельцев

**НЕ используйте эти инструменты для:**
- Атак на чужие системы
- Неавторизованного тестирования
- Любой незаконной деятельности

## 📁 Структура проекта

```
├── security_analysis_framework.js    # Основной фреймворк анализа
├── api_analyzer.js                   # Анализатор API контракта
├── xss_payload_generator.js          # Генератор XSS payload
├── comprehensive_security_audit.js   # Комплексный аудит
├── security_demo.html                # Демонстрационная страница
├── real_data_extractor.js           # Экстрактор реальных данных
├── extract_data.js                   # Базовый экстрактор данных
└── README_SECURITY_TOOLS.md         # Этот файл
```

## 🚀 Быстрый старт

### 1. Открытие демонстрационной страницы

```bash
# Откройте файл в браузере
open security_demo.html
```

### 2. Использование в консоли браузера

```javascript
// Инициализация аудитора
const auditor = new ComprehensiveSecurityAudit();
auditor.init('https://example.com', 'your-token', jsContent);

// Запуск полного аудита
auditor.runFullAudit().then(report => {
    console.log('Отчет:', report);
});
```

## 🛠️ Компоненты системы

### 1. SecurityAnalysisFramework

Основной фреймворк для анализа утечек данных в браузере.

**Возможности:**
- Анализ localStorage/sessionStorage
- Извлечение cookies
- Анализ форм и полей ввода
- Извлечение мета-данных
- Оценка чувствительности данных

**Пример использования:**
```javascript
const framework = new SecurityAnalysisFramework();
framework.init('https://example.com');
const leaks = await framework.analyzeDataLeaks();
```

### 2. APIContractAnalyzer

Анализатор API контракта и тестировщик эндпоинтов.

**Возможности:**
- Анализ JavaScript кода на предмет API эндпоинтов
- Автоматическое создание тестовых запросов
- Проверка заголовков безопасности
- Анализ уязвимостей API

**Пример использования:**
```javascript
const analyzer = new APIContractAnalyzer();
analyzer.init('https://api.example.com', 'bearer-token');
const report = await analyzer.runFullAPIAnalysis(jsContent);
```

### 3. XSSPayloadGenerator

Генератор XSS payload для демонстрации уязвимостей.

**Возможности:**
- Базовые XSS payload
- Продвинутые payload с извлечением данных
- Keylogger payload
- Payload для перенаправления
- Демонстрационные payload

**Пример использования:**
```javascript
const generator = new XSSPayloadGenerator();
const payloads = generator.getAllPayloads();
const customPayload = generator.createCustomPayload({
    action: 'alert',
    message: 'XSS Demo',
    dataExtraction: true
});
```

### 4. ComprehensiveSecurityAudit

Комплексный аудитор, объединяющий все компоненты.

**Возможности:**
- Полный анализ всех компонентов
- Генерация детальных отчетов
- Оценка уровня риска
- Рекомендации по устранению уязвимостей

**Пример использования:**
```javascript
const auditor = new ComprehensiveSecurityAudit();
auditor.init('https://example.com', 'token', jsContent);
const report = await auditor.runFullAudit();
auditor.saveReport(report);
```

## 📊 Типы анализируемых уязвимостей

### 1. Утечки данных
- **localStorage/sessionStorage**: Токены, персональные данные
- **Cookies**: Сессионные данные, настройки
- **Формы**: Чувствительные поля ввода
- **Мета-данные**: Информация о системе

### 2. API уязвимости
- **Отсутствие аутентификации**: Доступ без токенов
- **IDOR**: Недостаточная проверка авторизации
- **Отсутствие валидации**: Непроверенные входные данные
- **Отсутствие заголовков безопасности**: CSP, HSTS и др.

### 3. XSS уязвимости
- **Reflected XSS**: Отраженные атаки
- **Stored XSS**: Сохраненные атаки
- **DOM XSS**: Атаки на DOM
- **Data extraction**: Извлечение данных через XSS

### 4. Инфраструктурные уязвимости
- **HTTP вместо HTTPS**: Незащищенные соединения
- **Отсутствие WAF**: Нет защиты от атак
- **Устаревшие зависимости**: Уязвимые библиотеки

## 🎯 Демонстрационные сценарии

### 1. Анализ утечки данных
```javascript
// Демонстрация извлечения данных из браузера
demonstrateDataLeak();
```

### 2. Перехват токенов
```javascript
// Демонстрация перехвата аутентификационных данных
demonstrateTokenInterception();
```

### 3. Keylogger
```javascript
// Демонстрация перехвата пользовательского ввода
demonstrateKeylogger();
```

### 4. Фишинговое перенаправление
```javascript
// Демонстрация фишинговой атаки
demonstrateRedirect();
```

## 📋 Генерация отчетов

### Структура отчета
```json
{
  "summary": {
    "targetUrl": "https://example.com",
    "auditDate": "2024-01-01T00:00:00.000Z",
    "totalVulnerabilities": 5,
    "criticalVulnerabilities": 2,
    "highVulnerabilities": 2,
    "riskLevel": "HIGH"
  },
  "vulnerabilities": [...],
  "dataLeaks": [...],
  "apiVulnerabilities": [...],
  "xssVulnerabilities": [...],
  "recommendations": [...]
}
```

### Сохранение отчета
```javascript
// Автоматическое сохранение
auditor.saveReport(report);

// Ручное сохранение с именем файла
auditor.saveReport(report, 'my_audit_report.json');
```

## 🔧 Настройка и конфигурация

### Конфигурация аудитора
```javascript
const auditor = new ComprehensiveSecurityAudit();

// Настройка параметров
auditor.config = {
    enableDataExtraction: true,
    enableAPIAnalysis: true,
    enableXSSAnalysis: true,
    enableVulnerabilityScanning: true
};
```

### Кастомные payload
```javascript
const customPayload = xssGenerator.createCustomPayload({
    action: 'notification',
    message: 'Custom XSS Demo',
    dataExtraction: true,
    keylogger: false,
    redirect: false,
    targetUrl: 'https://attacker.com'
});
```

## 🛡️ Рекомендации по безопасности

### 1. Защита от утечек данных
- Используйте httpOnly cookies для токенов
- Не храните чувствительные данные в localStorage
- Шифруйте данные перед сохранением в браузере
- Очищайте данные при выходе пользователя

### 2. Защита API
- Внедрите аутентификацию для всех эндпоинтов
- Добавьте валидацию входных данных
- Используйте rate limiting
- Логируйте все API запросы

### 3. Защита от XSS
- Валидируйте и санитизируйте пользовательские данные
- Используйте Content Security Policy (CSP)
- Применяйте HttpOnly cookies
- Используйте современные фреймворки

### 4. Инфраструктурная безопасность
- Используйте HTTPS для всех соединений
- Добавьте заголовки безопасности
- Внедрите WAF (Web Application Firewall)
- Регулярно обновляйте зависимости

## 🚨 Обработка ошибок

### Типичные ошибки и решения

1. **CORS ошибки**
   ```javascript
   // Используйте прокси или настройте CORS на сервере
   ```

2. **Отсутствие разрешений**
   ```javascript
   // Убедитесь, что у вас есть разрешение на тестирование
   ```

3. **Блокировка WAF**
   ```javascript
   // Настройте WAF для разрешения легитимного тестирования
   ```

## 📚 Дополнительные ресурсы

### Документация
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Web Security Testing Guide](https://owasp.org/www-project-web-security-testing-guide/)
- [Content Security Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)

### Инструменты
- [Burp Suite](https://portswigger.net/burp)
- [OWASP ZAP](https://owasp.org/www-project-zap/)
- [Nikto](https://cirt.net/Nikto2)

## 🤝 Вклад в проект

Если вы хотите внести вклад в развитие инструментов:

1. Создайте форк проекта
2. Внесите изменения
3. Добавьте тесты
4. Создайте pull request

## 📄 Лицензия

Этот проект предназначен только для образовательных целей. Используйте ответственно и только на системах, которыми вы владеете или имеете явное разрешение на тестирование.

## 📞 Поддержка

Если у вас есть вопросы или предложения:

1. Создайте issue в репозитории
2. Опишите проблему или предложение
3. Приложите примеры кода, если необходимо

---

**🔒 Помните: Безопасность - это ответственность каждого!**
