/**
 * ФРЕЙМВОРК ДЛЯ АНАЛИЗА БЕЗОПАСНОСТИ ВЕБ-ПРИЛОЖЕНИЙ
 * 
 * Этот инструмент предназначен для:
 * - Анализа уязвимостей в контролируемой среде
 * - Демонстрации методов защиты
 * - Обучения принципам безопасности
 * 
 * ⚠️ ВНИМАНИЕ: Используйте только на системах, которыми вы владеете
 * или имеете явное разрешение на тестирование.
 */

class SecurityAnalysisFramework {
    constructor() {
        this.results = {
            vulnerabilities: [],
            dataLeaks: [],
            recommendations: [],
            timestamp: new Date().toISOString()
        };
        
        this.config = {
            targetUrl: null,
            authToken: null,
            enableXSSDetection: true,
            enableDataExtraction: true,
            enableAPIAnalysis: true
        };
    }

    /**
     * ИНИЦИАЛИЗАЦИЯ АНАЛИЗА
     */
    init(targetUrl, authToken = null) {
        console.log("🔒 ИНИЦИАЛИЗАЦИЯ АНАЛИЗА БЕЗОПАСНОСТИ");
        console.log("Цель:", targetUrl);
        
        this.config.targetUrl = targetUrl;
        this.config.authToken = authToken;
        
        return this;
    }

    /**
     * АНАЛИЗ УТЕЧЕК ДАННЫХ В БРАУЗЕРЕ
     */
    analyzeDataLeaks() {
        console.log("\n🔍 АНАЛИЗ УТЕЧЕК ДАННЫХ");
        
        const leaks = {
            localStorage: this.extractLocalStorage(),
            sessionStorage: this.extractSessionStorage(),
            cookies: this.extractCookies(),
            forms: this.extractFormData(),
            environment: this.extractEnvironmentData(),
            meta: this.extractMetaData()
        };

        // Анализ критичности утечек
        const criticalLeaks = this.analyzeCriticalLeaks(leaks);
        
        this.results.dataLeaks.push({
            type: 'browser_data',
            data: leaks,
            critical: criticalLeaks,
            timestamp: new Date().toISOString()
        });

        return leaks;
    }

    /**
     * ИЗВЛЕЧЕНИЕ ДАННЫХ ИЗ LOCALSTORAGE
     */
    extractLocalStorage() {
        const data = {};
        console.log("📦 Анализ localStorage:");
        
        for (let i = 0; i < localStorage.length; i++) {
            const key = localStorage.key(i);
            const value = localStorage.getItem(key);
            data[key] = value;
            
            // Анализ чувствительности данных
            const sensitivity = this.analyzeDataSensitivity(key, value);
            if (sensitivity.critical) {
                console.log(`  ⚠️  КРИТИЧНО: ${key} - ${sensitivity.reason}`);
            } else if (sensitivity.sensitive) {
                console.log(`  🔸 ЧУВСТВИТЕЛЬНО: ${key}`);
            } else {
                console.log(`  ✅ БЕЗОПАСНО: ${key}`);
            }
        }
        
        return data;
    }

    /**
     * ИЗВЛЕЧЕНИЕ ДАННЫХ ИЗ SESSIONSTORAGE
     */
    extractSessionStorage() {
        const data = {};
        console.log("📦 Анализ sessionStorage:");
        
        for (let i = 0; i < sessionStorage.length; i++) {
            const key = sessionStorage.key(i);
            const value = sessionStorage.getItem(key);
            data[key] = value;
            
            const sensitivity = this.analyzeDataSensitivity(key, value);
            if (sensitivity.critical) {
                console.log(`  ⚠️  КРИТИЧНО: ${key} - ${sensitivity.reason}`);
            }
        }
        
        return data;
    }

    /**
     * ИЗВЛЕЧЕНИЕ COOKIES
     */
    extractCookies() {
        const cookies = {};
        console.log("🍪 Анализ cookies:");
        
        document.cookie.split(';').forEach(cookie => {
            const [name, value] = cookie.trim().split('=');
            cookies[name] = value;
            
            const sensitivity = this.analyzeDataSensitivity(name, value);
            if (sensitivity.critical) {
                console.log(`  ⚠️  КРИТИЧНО: ${name} - ${sensitivity.reason}`);
            }
        });
        
        return cookies;
    }

    /**
     * ИЗВЛЕЧЕНИЕ ДАННЫХ ИЗ ФОРМ
     */
    extractFormData() {
        const forms = document.querySelectorAll('form');
        const formData = {};
        
        console.log("📝 Анализ форм:");
        forms.forEach((form, index) => {
            formData[`form_${index}`] = {
                action: form.action,
                method: form.method,
                inputs: {}
            };
            
            // Анализ input полей
            const inputs = form.querySelectorAll('input, textarea, select');
            inputs.forEach(input => {
                const key = input.name || input.id || `input_${index}`;
                formData[`form_${index}`].inputs[key] = {
                    type: input.type,
                    value: input.value,
                    required: input.required,
                    placeholder: input.placeholder
                };
                
                // Проверка на чувствительные поля
                if (this.isSensitiveField(key, input.type)) {
                    console.log(`  ⚠️  ЧУВСТВИТЕЛЬНОЕ ПОЛЕ: ${key} (${input.type})`);
                }
            });
        });
        
        return formData;
    }

    /**
     * ИЗВЛЕЧЕНИЕ ДАННЫХ ОКРУЖЕНИЯ
     */
    extractEnvironmentData() {
        return {
            userAgent: navigator.userAgent,
            language: navigator.language,
            platform: navigator.platform,
            cookieEnabled: navigator.cookieEnabled,
            online: navigator.onLine,
            url: window.location.href,
            referrer: document.referrer,
            screen: {
                width: screen.width,
                height: screen.height,
                colorDepth: screen.colorDepth
            },
            timezone: Intl.DateTimeFormat().resolvedOptions().timeZone
        };
    }

    /**
     * ИЗВЛЕЧЕНИЕ МЕТА-ДАННЫХ
     */
    extractMetaData() {
        const meta = {};
        const metaTags = document.querySelectorAll('meta');
        
        metaTags.forEach(tag => {
            const name = tag.name || tag.property || tag.httpEquiv;
            if (name) {
                meta[name] = tag.content;
            }
        });
        
        return meta;
    }

    /**
     * АНАЛИЗ ЧУВСТВИТЕЛЬНОСТИ ДАННЫХ
     */
    analyzeDataSensitivity(key, value) {
        const sensitivePatterns = {
            token: /token|auth|bearer|jwt|session/i,
            password: /password|passwd|pwd|secret/i,
            key: /key|api_key|secret_key|private_key/i,
            id: /user_id|id|guid|uuid/i,
            personal: /email|phone|address|name|ssn|passport/i
        };

        for (const [type, pattern] of Object.entries(sensitivePatterns)) {
            if (pattern.test(key)) {
                return {
                    sensitive: true,
                    critical: type === 'token' || type === 'password' || type === 'key',
                    reason: `Содержит ${type} данные`,
                    type: type
                };
            }
        }

        return { sensitive: false, critical: false };
    }

    /**
     * ПРОВЕРКА ЧУВСТВИТЕЛЬНЫХ ПОЛЕЙ
     */
    isSensitiveField(name, type) {
        const sensitiveFields = [
            'password', 'passwd', 'pwd', 'secret',
            'email', 'phone', 'ssn', 'passport',
            'credit_card', 'card_number', 'cvv'
        ];
        
        return sensitiveFields.some(field => 
            name.toLowerCase().includes(field)
        ) || type === 'password';
    }

    /**
     * АНАЛИЗ КРИТИЧЕСКИХ УТЕЧЕК
     */
    analyzeCriticalLeaks(leaks) {
        const critical = [];
        
        // Проверка токенов аутентификации
        Object.entries(leaks.localStorage).forEach(([key, value]) => {
            if (this.analyzeDataSensitivity(key, value).critical) {
                critical.push({
                    type: 'auth_token',
                    location: 'localStorage',
                    key: key,
                    severity: 'CRITICAL',
                    description: 'Обнаружен токен аутентификации в localStorage'
                });
            }
        });

        // Проверка персональных данных
        Object.entries(leaks.cookies).forEach(([key, value]) => {
            if (this.analyzeDataSensitivity(key, value).sensitive) {
                critical.push({
                    type: 'personal_data',
                    location: 'cookies',
                    key: key,
                    severity: 'HIGH',
                    description: 'Обнаружены персональные данные в cookies'
                });
            }
        });

        return critical;
    }

    /**
     * АНАЛИЗ API УЯЗВИМОСТЕЙ
     */
    analyzeAPIVulnerabilities() {
        console.log("\n🔌 АНАЛИЗ API УЯЗВИМОСТЕЙ");
        
        // Перехват fetch запросов
        const originalFetch = window.fetch;
        const apiCalls = [];
        
        window.fetch = function(...args) {
            const url = args[0];
            const options = args[1] || {};
            
            const call = {
                url: url,
                method: options.method || 'GET',
                headers: options.headers,
                body: options.body,
                timestamp: new Date().toISOString()
            };
            
            apiCalls.push(call);
            
            // Анализ уязвимостей в запросе
            const vulnerabilities = this.analyzeAPIRequest(call);
            if (vulnerabilities.length > 0) {
                console.log(`  ⚠️  УЯЗВИМОСТИ В ЗАПРОСЕ: ${url}`);
                vulnerabilities.forEach(vuln => {
                    console.log(`    - ${vuln.type}: ${vuln.description}`);
                });
            }
            
            return originalFetch.apply(this, args);
        }.bind(this);

        return apiCalls;
    }

    /**
     * АНАЛИЗ ОДНОГО API ЗАПРОСА
     */
    analyzeAPIRequest(call) {
        const vulnerabilities = [];
        
        // Проверка на отсутствие HTTPS
        if (call.url.startsWith('http://')) {
            vulnerabilities.push({
                type: 'INSECURE_TRANSPORT',
                description: 'Используется незащищенное HTTP соединение'
            });
        }
        
        // Проверка на отсутствие заголовков безопасности
        if (!call.headers || !call.headers['X-Content-Type-Options']) {
            vulnerabilities.push({
                type: 'MISSING_SECURITY_HEADERS',
                description: 'Отсутствуют заголовки безопасности'
            });
        }
        
        // Проверка на чувствительные данные в URL
        const sensitiveParams = ['token', 'auth', 'password', 'key', 'secret'];
        sensitiveParams.forEach(param => {
            if (call.url.includes(param + '=')) {
                vulnerabilities.push({
                    type: 'SENSITIVE_DATA_IN_URL',
                    description: `Чувствительные данные в URL: ${param}`
                });
            }
        });
        
        return vulnerabilities;
    }

    /**
     * СОЗДАНИЕ XSS PAYLOAD ДЛЯ ТЕСТИРОВАНИЯ
     */
    createXSSPayload() {
        console.log("\n🎯 СОЗДАНИЕ XSS PAYLOAD");
        
        const payloads = {
            basic: '<script>alert("XSS")</script>',
            img: '<img src=x onerror=alert("XSS")>',
            svg: '<svg onload=alert("XSS")>',
            data: 'javascript:alert("XSS")',
            encoded: '&#60;script&#62;alert("XSS")&#60;/script&#62;'
        };
        
        console.log("Доступные payloads:");
        Object.entries(payloads).forEach(([name, payload]) => {
            console.log(`  ${name}: ${payload}`);
        });
        
        return payloads;
    }

    /**
     * ГЕНЕРАЦИЯ ОТЧЕТА
     */
    generateReport() {
        console.log("\n📊 ГЕНЕРАЦИЯ ОТЧЕТА ПО БЕЗОПАСНОСТИ");
        
        const report = {
            summary: {
                totalVulnerabilities: this.results.vulnerabilities.length,
                criticalVulnerabilities: this.results.vulnerabilities.filter(v => v.severity === 'CRITICAL').length,
                dataLeaks: this.results.dataLeaks.length,
                timestamp: new Date().toISOString()
            },
            vulnerabilities: this.results.vulnerabilities,
            dataLeaks: this.results.dataLeaks,
            recommendations: this.generateRecommendations()
        };
        
        console.log("📋 СВОДКА:");
        console.log(`  Всего уязвимостей: ${report.summary.totalVulnerabilities}`);
        console.log(`  Критических уязвимостей: ${report.summary.criticalVulnerabilities}`);
        console.log(`  Утечек данных: ${report.summary.dataLeaks}`);
        
        return report;
    }

    /**
     * ГЕНЕРАЦИЯ РЕКОМЕНДАЦИЙ
     */
    generateRecommendations() {
        const recommendations = [];
        
        // Рекомендации по защите данных
        if (this.results.dataLeaks.some(leak => leak.critical.length > 0)) {
            recommendations.push({
                priority: 'CRITICAL',
                category: 'DATA_PROTECTION',
                title: 'Защита чувствительных данных',
                description: 'Не храните токены аутентификации и персональные данные в localStorage/cookies',
                actions: [
                    'Используйте httpOnly cookies для токенов',
                    'Очищайте localStorage при выходе',
                    'Шифруйте чувствительные данные'
                ]
            });
        }
        
        // Рекомендации по API безопасности
        recommendations.push({
            priority: 'HIGH',
            category: 'API_SECURITY',
            title: 'Усиление API безопасности',
            description: 'Внедрите дополнительные меры защиты API',
            actions: [
                'Используйте HTTPS для всех запросов',
                'Добавьте заголовки безопасности',
                'Внедрите rate limiting',
                'Валидируйте все входные данные'
            ]
        });
        
        return recommendations;
    }

    /**
     * ЗАПУСК ПОЛНОГО АНАЛИЗА
     */
    async runFullAnalysis() {
        console.log("🚀 ЗАПУСК ПОЛНОГО АНАЛИЗА БЕЗОПАСНОСТИ");
        
        try {
            // Анализ утечек данных
            const dataLeaks = this.analyzeDataLeaks();
            
            // Анализ API уязвимостей
            const apiVulnerabilities = this.analyzeAPIVulnerabilities();
            
            // Создание XSS payload
            const xssPayloads = this.createXSSPayload();
            
            // Генерация отчета
            const report = this.generateReport();
            
            console.log("\n✅ АНАЛИЗ ЗАВЕРШЕН");
            return report;
            
        } catch (error) {
            console.error("❌ ОШИБКА ПРИ АНАЛИЗЕ:", error);
            throw error;
        }
    }
}

// Экспорт для использования
if (typeof module !== 'undefined' && module.exports) {
    module.exports = SecurityAnalysisFramework;
} else if (typeof window !== 'undefined') {
    window.SecurityAnalysisFramework = SecurityAnalysisFramework;
}

// Пример использования
if (typeof window !== 'undefined') {
    console.log("🔒 ФРЕЙМВОРК АНАЛИЗА БЕЗОПАСНОСТИ ЗАГРУЖЕН");
    console.log("Используйте: new SecurityAnalysisFramework().init(url).runFullAnalysis()");
}
