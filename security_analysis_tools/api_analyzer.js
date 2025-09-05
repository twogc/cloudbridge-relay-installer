/**
 * АНАЛИЗАТОР API КОНТРАКТА
 * 
 * Этот инструмент анализирует API контракт и создает безопасные запросы
 * для тестирования в контролируемой среде.
 */

class APIContractAnalyzer {
    constructor() {
        this.endpoints = [];
        this.authToken = null;
        this.baseUrl = null;
        this.results = {
            discoveredEndpoints: [],
            testResults: [],
            vulnerabilities: []
        };
    }

    /**
     * ИНИЦИАЛИЗАЦИЯ АНАЛИЗАТОРА
     */
    init(baseUrl, authToken = null) {
        console.log("🔌 ИНИЦИАЛИЗАЦИЯ API АНАЛИЗАТОРА");
        console.log("Базовый URL:", baseUrl);
        
        this.baseUrl = baseUrl;
        this.authToken = authToken;
        
        return this;
    }

    /**
     * АНАЛИЗ API КОНТРАКТА ИЗ JS ФАЙЛА
     */
    analyzeContractFromJS(jsContent) {
        console.log("\n📋 АНАЛИЗ API КОНТРАКТА");
        
        // Поиск API эндпоинтов в коде
        const endpointPatterns = [
            /\/api\/[a-zA-Z0-9\/\-_]+/g,
            /fetch\(['"`]([^'"`]+)['"`]/g,
            /axios\.(get|post|put|delete)\(['"`]([^'"`]+)['"`]/g
        ];
        
        const discoveredEndpoints = new Set();
        
        endpointPatterns.forEach(pattern => {
            const matches = jsContent.match(pattern);
            if (matches) {
                matches.forEach(match => {
                    if (match.startsWith('/api/')) {
                        discoveredEndpoints.add(match);
                    }
                });
            }
        });
        
        // Анализ найденных эндпоинтов
        this.endpoints = Array.from(discoveredEndpoints).map(endpoint => {
            return this.analyzeEndpoint(endpoint);
        });
        
        console.log(`Найдено эндпоинтов: ${this.endpoints.length}`);
        this.endpoints.forEach(endpoint => {
            console.log(`  ${endpoint.method} ${endpoint.path} - ${endpoint.description}`);
        });
        
        return this.endpoints;
    }

    /**
     * АНАЛИЗ ОДНОГО ЭНДПОИНТА
     */
    analyzeEndpoint(path) {
        const analysis = {
            path: path,
            method: this.detectMethod(path),
            description: this.generateDescription(path),
            parameters: this.detectParameters(path),
            security: this.analyzeSecurity(path)
        };
        
        return analysis;
    }

    /**
     * ОПРЕДЕЛЕНИЕ HTTP МЕТОДА
     */
    detectMethod(path) {
        const methodPatterns = {
            'GET': /retrieve|get|list|fetch|read/i,
            'POST': /create|add|submit|send|upload/i,
            'PUT': /update|modify|edit|change/i,
            'DELETE': /delete|remove|destroy/i
        };
        
        for (const [method, pattern] of Object.entries(methodPatterns)) {
            if (pattern.test(path)) {
                return method;
            }
        }
        
        return 'GET'; // По умолчанию
    }

    /**
     * ГЕНЕРАЦИЯ ОПИСАНИЯ ЭНДПОИНТА
     */
    generateDescription(path) {
        const descriptions = {
            '/api/application/retrieve': 'Получение данных заявки',
            '/api/application/files': 'Получение файлов заявки',
            '/api/application/urls': 'Получение URL заявки',
            '/api/application/update': 'Обновление заявки',
            '/api/application/submit': 'Отправка заявки',
            '/api/user/profile': 'Профиль пользователя',
            '/api/auth/login': 'Аутентификация',
            '/api/auth/logout': 'Выход из системы'
        };
        
        return descriptions[path] || `API эндпоинт: ${path}`;
    }

    /**
     * ОПРЕДЕЛЕНИЕ ПАРАМЕТРОВ
     */
    detectParameters(path) {
        const parameters = [];
        
        // Поиск ID параметров
        const idMatches = path.match(/\/(\d+)/g);
        if (idMatches) {
            idMatches.forEach(match => {
                parameters.push({
                    name: 'id',
                    type: 'number',
                    required: true,
                    example: match.replace('/', '')
                });
            });
        }
        
        // Поиск других параметров
        const paramMatches = path.match(/\{([^}]+)\}/g);
        if (paramMatches) {
            paramMatches.forEach(match => {
                const paramName = match.replace(/[{}]/g, '');
                parameters.push({
                    name: paramName,
                    type: 'string',
                    required: true,
                    example: 'example_value'
                });
            });
        }
        
        return parameters;
    }

    /**
     * АНАЛИЗ БЕЗОПАСНОСТИ ЭНДПОИНТА
     */
    analyzeSecurity(path) {
        const security = {
            requiresAuth: false,
            sensitiveData: false,
            vulnerabilities: []
        };
        
        // Проверка на чувствительные данные
        const sensitivePatterns = [
            /user|profile|personal/i,
            /password|auth|login/i,
            /token|key|secret/i,
            /admin|management/i
        ];
        
        sensitivePatterns.forEach(pattern => {
            if (pattern.test(path)) {
                security.sensitiveData = true;
                security.requiresAuth = true;
            }
        });
        
        // Проверка на потенциальные уязвимости
        if (path.includes('id') && !path.includes('user')) {
            security.vulnerabilities.push({
                type: 'IDOR',
                description: 'Потенциальная уязвимость IDOR - проверка авторизации'
            });
        }
        
        return security;
    }

    /**
     * СОЗДАНИЕ ТЕСТОВЫХ ЗАПРОСОВ
     */
    createTestRequests() {
        console.log("\n🧪 СОЗДАНИЕ ТЕСТОВЫХ ЗАПРОСОВ");
        
        const requests = [];
        
        this.endpoints.forEach(endpoint => {
            const request = this.createRequest(endpoint);
            if (request) {
                requests.push(request);
            }
        });
        
        return requests;
    }

    /**
     * СОЗДАНИЕ ОДНОГО ЗАПРОСА
     */
    createRequest(endpoint) {
        const url = this.buildUrl(endpoint);
        const headers = this.buildHeaders(endpoint);
        const body = this.buildBody(endpoint);
        
        return {
            url: url,
            method: endpoint.method,
            headers: headers,
            body: body,
            description: endpoint.description,
            security: endpoint.security
        };
    }

    /**
     * ПОСТРОЕНИЕ URL
     */
    buildUrl(endpoint) {
        let url = this.baseUrl + endpoint.path;
        
        // Замена параметров примерами
        endpoint.parameters.forEach(param => {
            url = url.replace(`{${param.name}}`, param.example);
            url = url.replace(`/${param.example}`, `/${param.example}`);
        });
        
        return url;
    }

    /**
     * ПОСТРОЕНИЕ ЗАГОЛОВКОВ
     */
    buildHeaders(endpoint) {
        const headers = {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            'User-Agent': 'SecurityAnalyzer/1.0'
        };
        
        if (this.authToken && endpoint.security.requiresAuth) {
            headers['Authorization'] = `Bearer ${this.authToken}`;
        }
        
        return headers;
    }

    /**
     * ПОСТРОЕНИЕ ТЕЛА ЗАПРОСА
     */
    buildBody(endpoint) {
        if (endpoint.method === 'GET' || endpoint.method === 'DELETE') {
            return null;
        }
        
        // Создание примера тела запроса
        const body = {};
        
        if (endpoint.path.includes('application')) {
            body.applicationId = 60526;
            body.status = 'draft';
            body.timestamp = new Date().toISOString();
        }
        
        if (endpoint.path.includes('user')) {
            body.userId = '2045597475330327839';
            body.email = 'test@example.com';
        }
        
        return Object.keys(body).length > 0 ? JSON.stringify(body) : null;
    }

    /**
     * ВЫПОЛНЕНИЕ ТЕСТОВЫХ ЗАПРОСОВ
     */
    async executeTestRequests() {
        console.log("\n🚀 ВЫПОЛНЕНИЕ ТЕСТОВЫХ ЗАПРОСОВ");
        
        const requests = this.createTestRequests();
        const results = [];
        
        for (const request of requests) {
            try {
                console.log(`Тестирование: ${request.method} ${request.url}`);
                
                const response = await this.executeRequest(request);
                results.push({
                    request: request,
                    response: response,
                    success: response.status < 400,
                    timestamp: new Date().toISOString()
                });
                
                // Анализ ответа на уязвимости
                const vulnerabilities = this.analyzeResponse(response);
                if (vulnerabilities.length > 0) {
                    console.log(`  ⚠️  УЯЗВИМОСТИ В ОТВЕТЕ:`);
                    vulnerabilities.forEach(vuln => {
                        console.log(`    - ${vuln.type}: ${vuln.description}`);
                    });
                }
                
            } catch (error) {
                console.log(`  ❌ ОШИБКА: ${error.message}`);
                results.push({
                    request: request,
                    error: error.message,
                    success: false,
                    timestamp: new Date().toISOString()
                });
            }
        }
        
        this.results.testResults = results;
        return results;
    }

    /**
     * ВЫПОЛНЕНИЕ ОДНОГО ЗАПРОСА
     */
    async executeRequest(request) {
        const options = {
            method: request.method,
            headers: request.headers
        };
        
        if (request.body) {
            options.body = request.body;
        }
        
        const response = await fetch(request.url, options);
        
        return {
            status: response.status,
            statusText: response.statusText,
            headers: Object.fromEntries(response.headers.entries()),
            body: await response.text()
        };
    }

    /**
     * АНАЛИЗ ОТВЕТА НА УЯЗВИМОСТИ
     */
    analyzeResponse(response) {
        const vulnerabilities = [];
        
        // Проверка на отсутствие заголовков безопасности
        const securityHeaders = [
            'X-Content-Type-Options',
            'X-Frame-Options',
            'X-XSS-Protection',
            'Strict-Transport-Security'
        ];
        
        securityHeaders.forEach(header => {
            if (!response.headers[header]) {
                vulnerabilities.push({
                    type: 'MISSING_SECURITY_HEADER',
                    description: `Отсутствует заголовок ${header}`
                });
            }
        });
        
        // Проверка на чувствительные данные в ответе
        if (response.body) {
            const sensitivePatterns = [
                /password|passwd|pwd/i,
                /token|auth|bearer/i,
                /key|secret|private/i,
                /ssn|passport|credit_card/i
            ];
            
            sensitivePatterns.forEach(pattern => {
                if (pattern.test(response.body)) {
                    vulnerabilities.push({
                        type: 'SENSITIVE_DATA_EXPOSURE',
                        description: `Обнаружены чувствительные данные в ответе`
                    });
                }
            });
        }
        
        return vulnerabilities;
    }

    /**
     * ГЕНЕРАЦИЯ ОТЧЕТА ПО API
     */
    generateAPIReport() {
        console.log("\n📊 ОТЧЕТ ПО АНАЛИЗУ API");
        
        const report = {
            summary: {
                totalEndpoints: this.endpoints.length,
                testedEndpoints: this.results.testResults.length,
                successfulRequests: this.results.testResults.filter(r => r.success).length,
                vulnerabilities: this.results.vulnerabilities.length
            },
            endpoints: this.endpoints,
            testResults: this.results.testResults,
            recommendations: this.generateAPIRecommendations()
        };
        
        console.log("📋 СВОДКА API:");
        console.log(`  Всего эндпоинтов: ${report.summary.totalEndpoints}`);
        console.log(`  Протестировано: ${report.summary.testedEndpoints}`);
        console.log(`  Успешных запросов: ${report.summary.successfulRequests}`);
        console.log(`  Уязвимостей: ${report.summary.vulnerabilities}`);
        
        return report;
    }

    /**
     * ГЕНЕРАЦИЯ РЕКОМЕНДАЦИЙ ПО API
     */
    generateAPIRecommendations() {
        const recommendations = [];
        
        // Рекомендации по безопасности
        recommendations.push({
            priority: 'HIGH',
            category: 'API_SECURITY',
            title: 'Усиление безопасности API',
            description: 'Внедрите дополнительные меры защиты API',
            actions: [
                'Добавьте заголовки безопасности',
                'Внедрите rate limiting',
                'Валидируйте все входные данные',
                'Используйте HTTPS для всех запросов'
            ]
        });
        
        // Рекомендации по аутентификации
        if (this.endpoints.some(e => e.security.requiresAuth)) {
            recommendations.push({
                priority: 'CRITICAL',
                category: 'AUTHENTICATION',
                title: 'Усиление аутентификации',
                description: 'Улучшите систему аутентификации',
                actions: [
                    'Используйте JWT токены с коротким временем жизни',
                    'Внедрите refresh токены',
                    'Добавьте двухфакторную аутентификацию',
                    'Логируйте все попытки доступа'
                ]
            });
        }
        
        return recommendations;
    }

    /**
     * ЗАПУСК ПОЛНОГО АНАЛИЗА API
     */
    async runFullAPIAnalysis(jsContent) {
        console.log("🚀 ЗАПУСК ПОЛНОГО АНАЛИЗА API");
        
        try {
            // Анализ контракта
            const endpoints = this.analyzeContractFromJS(jsContent);
            
            // Выполнение тестовых запросов
            const testResults = await this.executeTestRequests();
            
            // Генерация отчета
            const report = this.generateAPIReport();
            
            console.log("\n✅ АНАЛИЗ API ЗАВЕРШЕН");
            return report;
            
        } catch (error) {
            console.error("❌ ОШИБКА ПРИ АНАЛИЗЕ API:", error);
            throw error;
        }
    }
}

// Экспорт для использования
if (typeof module !== 'undefined' && module.exports) {
    module.exports = APIContractAnalyzer;
} else if (typeof window !== 'undefined') {
    window.APIContractAnalyzer = APIContractAnalyzer;
}

// Пример использования
if (typeof window !== 'undefined') {
    console.log("🔌 API АНАЛИЗАТОР ЗАГРУЖЕН");
    console.log("Используйте: new APIContractAnalyzer().init(url, token).runFullAPIAnalysis(jsContent)");
}
