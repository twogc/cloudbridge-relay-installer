/**
 * КОМПЛЕКСНЫЙ АУДИТ БЕЗОПАСНОСТИ
 * 
 * Этот инструмент объединяет все компоненты для проведения
 * полного аудита безопасности веб-приложения.
 * 
 * ⚠️ ВНИМАНИЕ: Используйте только на системах, которыми вы владеете
 * или имеете явное разрешение на тестирование.
 */

// Импорт всех компонентов
if (typeof SecurityAnalysisFramework !== 'undefined') {
    var SecurityFramework = SecurityAnalysisFramework;
} else {
    console.log("⚠️ SecurityAnalysisFramework не найден");
}

if (typeof APIContractAnalyzer !== 'undefined') {
    var APIAnalyzer = APIContractAnalyzer;
} else {
    console.log("⚠️ APIContractAnalyzer не найден");
}

if (typeof XSSPayloadGenerator !== 'undefined') {
    var XSSGenerator = XSSPayloadGenerator;
} else {
    console.log("⚠️ XSSPayloadGenerator не найден");
}

class ComprehensiveSecurityAudit {
    constructor() {
        this.config = {
            targetUrl: null,
            authToken: null,
            jsContent: null,
            enableDataExtraction: true,
            enableAPIAnalysis: true,
            enableXSSAnalysis: true,
            enableVulnerabilityScanning: true
        };
        
        this.results = {
            dataLeaks: [],
            apiVulnerabilities: [],
            xssVulnerabilities: [],
            recommendations: [],
            timestamp: new Date().toISOString()
        };
        
        this.frameworks = {
            security: null,
            api: null,
            xss: null
        };
    }

    /**
     * ИНИЦИАЛИЗАЦИЯ АУДИТА
     */
    init(targetUrl, authToken = null, jsContent = null) {
        console.log("🔒 ИНИЦИАЛИЗАЦИЯ КОМПЛЕКСНОГО АУДИТА БЕЗОПАСНОСТИ");
        console.log("Цель:", targetUrl);
        
        this.config.targetUrl = targetUrl;
        this.config.authToken = authToken;
        this.config.jsContent = jsContent;
        
        // Инициализация фреймворков
        if (SecurityFramework) {
            this.frameworks.security = new SecurityFramework();
            this.frameworks.security.init(targetUrl, authToken);
        }
        
        if (APIAnalyzer) {
            this.frameworks.api = new APIAnalyzer();
            this.frameworks.api.init(targetUrl, authToken);
        }
        
        if (XSSGenerator) {
            this.frameworks.xss = new XSSGenerator();
        }
        
        return this;
    }

    /**
     * АНАЛИЗ УТЕЧЕК ДАННЫХ
     */
    async analyzeDataLeaks() {
        console.log("\n🔍 ЭТАП 1: АНАЛИЗ УТЕЧЕК ДАННЫХ");
        
        if (!this.frameworks.security) {
            console.log("❌ SecurityFramework недоступен");
            return null;
        }
        
        try {
            const leaks = await this.frameworks.security.analyzeDataLeaks();
            this.results.dataLeaks.push({
                type: 'browser_data',
                data: leaks,
                timestamp: new Date().toISOString()
            });
            
            return leaks;
        } catch (error) {
            console.error("❌ Ошибка при анализе утечек данных:", error);
            return null;
        }
    }

    /**
     * АНАЛИЗ API УЯЗВИМОСТЕЙ
     */
    async analyzeAPIVulnerabilities() {
        console.log("\n🔌 ЭТАП 2: АНАЛИЗ API УЯЗВИМОСТЕЙ");
        
        if (!this.frameworks.api || !this.config.jsContent) {
            console.log("❌ APIAnalyzer или JS контент недоступен");
            return null;
        }
        
        try {
            const apiReport = await this.frameworks.api.runFullAPIAnalysis(this.config.jsContent);
            this.results.apiVulnerabilities.push({
                type: 'api_analysis',
                data: apiReport,
                timestamp: new Date().toISOString()
            });
            
            return apiReport;
        } catch (error) {
            console.error("❌ Ошибка при анализе API:", error);
            return null;
        }
    }

    /**
     * АНАЛИЗ XSS УЯЗВИМОСТЕЙ
     */
    async analyzeXSSVulnerabilities() {
        console.log("\n🎯 ЭТАП 3: АНАЛИЗ XSS УЯЗВИМОСТЕЙ");
        
        if (!this.frameworks.xss) {
            console.log("❌ XSSGenerator недоступен");
            return null;
        }
        
        try {
            const xssPayloads = this.frameworks.xss.getAllPayloads();
            const xssReport = this.frameworks.xss.generatePayloadReport();
            
            this.results.xssVulnerabilities.push({
                type: 'xss_analysis',
                payloads: xssPayloads,
                report: xssReport,
                timestamp: new Date().toISOString()
            });
            
            return xssReport;
        } catch (error) {
            console.error("❌ Ошибка при анализе XSS:", error);
            return null;
        }
    }

    /**
     * СКАНИРОВАНИЕ УЯЗВИМОСТЕЙ
     */
    async scanVulnerabilities() {
        console.log("\n🔍 ЭТАП 4: СКАНИРОВАНИЕ УЯЗВИМОСТЕЙ");
        
        const vulnerabilities = [];
        
        // Проверка на отсутствие HTTPS
        if (this.config.targetUrl && !this.config.targetUrl.startsWith('https://')) {
            vulnerabilities.push({
                type: 'INSECURE_TRANSPORT',
                severity: 'HIGH',
                description: 'Используется незащищенное HTTP соединение',
                recommendation: 'Перейдите на HTTPS'
            });
        }
        
        // Проверка на отсутствие заголовков безопасности
        try {
            const response = await fetch(this.config.targetUrl);
            const headers = response.headers;
            
            const securityHeaders = [
                'X-Content-Type-Options',
                'X-Frame-Options',
                'X-XSS-Protection',
                'Strict-Transport-Security',
                'Content-Security-Policy'
            ];
            
            securityHeaders.forEach(header => {
                if (!headers.get(header)) {
                    vulnerabilities.push({
                        type: 'MISSING_SECURITY_HEADER',
                        severity: 'MEDIUM',
                        description: `Отсутствует заголовок ${header}`,
                        recommendation: `Добавьте заголовок ${header}`
                    });
                }
            });
        } catch (error) {
            console.log("⚠️ Не удалось проверить заголовки безопасности");
        }
        
        // Проверка на утечку информации
        if (this.config.jsContent) {
            const infoLeakPatterns = [
                /password|passwd|pwd/i,
                /token|auth|bearer/i,
                /key|secret|private/i,
                /api_key|secret_key/i
            ];
            
            infoLeakPatterns.forEach(pattern => {
                if (pattern.test(this.config.jsContent)) {
                    vulnerabilities.push({
                        type: 'INFORMATION_DISCLOSURE',
                        severity: 'CRITICAL',
                        description: 'Обнаружены чувствительные данные в исходном коде',
                        recommendation: 'Удалите чувствительные данные из клиентского кода'
                    });
                }
            });
        }
        
        this.results.vulnerabilities = vulnerabilities;
        return vulnerabilities;
    }

    /**
     * ГЕНЕРАЦИЯ РЕКОМЕНДАЦИЙ
     */
    generateRecommendations() {
        console.log("\n📋 ЭТАП 5: ГЕНЕРАЦИЯ РЕКОМЕНДАЦИЙ");
        
        const recommendations = [];
        
        // Критические рекомендации
        if (this.results.dataLeaks.some(leak => leak.data && leak.data.critical && leak.data.critical.length > 0)) {
            recommendations.push({
                priority: 'CRITICAL',
                category: 'DATA_PROTECTION',
                title: 'Защита чувствительных данных',
                description: 'Обнаружены критические утечки данных',
                actions: [
                    'Не храните токены аутентификации в localStorage',
                    'Используйте httpOnly cookies для чувствительных данных',
                    'Шифруйте данные перед сохранением в браузере',
                    'Очищайте данные при выходе пользователя'
                ]
            });
        }
        
        // Рекомендации по API безопасности
        if (this.results.apiVulnerabilities.length > 0) {
            recommendations.push({
                priority: 'HIGH',
                category: 'API_SECURITY',
                title: 'Усиление API безопасности',
                description: 'Обнаружены уязвимости в API',
                actions: [
                    'Внедрите аутентификацию для всех API эндпоинтов',
                    'Добавьте валидацию входных данных',
                    'Используйте rate limiting',
                    'Логируйте все API запросы'
                ]
            });
        }
        
        // Рекомендации по защите от XSS
        recommendations.push({
            priority: 'HIGH',
            category: 'XSS_PROTECTION',
            title: 'Защита от XSS атак',
            description: 'Внедрите меры защиты от XSS',
            actions: [
                'Валидируйте и санитизируйте все пользовательские данные',
                'Используйте Content Security Policy (CSP)',
                'Применяйте HttpOnly cookies',
                'Используйте современные фреймворки с встроенной защитой'
            ]
        });
        
        // Рекомендации по инфраструктуре
        recommendations.push({
            priority: 'MEDIUM',
            category: 'INFRASTRUCTURE',
            title: 'Усиление инфраструктуры',
            description: 'Общие рекомендации по безопасности',
            actions: [
                'Используйте HTTPS для всех соединений',
                'Добавьте заголовки безопасности',
                'Внедрите WAF (Web Application Firewall)',
                'Регулярно обновляйте зависимости'
            ]
        });
        
        this.results.recommendations = recommendations;
        return recommendations;
    }

    /**
     * ГЕНЕРАЦИЯ ФИНАЛЬНОГО ОТЧЕТА
     */
    generateFinalReport() {
        console.log("\n📊 ГЕНЕРАЦИЯ ФИНАЛЬНОГО ОТЧЕТА");
        
        const report = {
            summary: {
                targetUrl: this.config.targetUrl,
                auditDate: new Date().toISOString(),
                totalVulnerabilities: this.results.vulnerabilities.length,
                criticalVulnerabilities: this.results.vulnerabilities.filter(v => v.severity === 'CRITICAL').length,
                highVulnerabilities: this.results.vulnerabilities.filter(v => v.severity === 'HIGH').length,
                dataLeaks: this.results.dataLeaks.length,
                apiVulnerabilities: this.results.apiVulnerabilities.length,
                xssVulnerabilities: this.results.xssVulnerabilities.length
            },
            vulnerabilities: this.results.vulnerabilities,
            dataLeaks: this.results.dataLeaks,
            apiVulnerabilities: this.results.apiVulnerabilities,
            xssVulnerabilities: this.results.xssVulnerabilities,
            recommendations: this.results.recommendations,
            riskAssessment: this.assessRiskLevel()
        };
        
        // Вывод сводки
        console.log("\n📋 СВОДКА АУДИТА:");
        console.log(`  Цель: ${report.summary.targetUrl}`);
        console.log(`  Дата: ${new Date(report.summary.auditDate).toLocaleString()}`);
        console.log(`  Всего уязвимостей: ${report.summary.totalVulnerabilities}`);
        console.log(`  Критических: ${report.summary.criticalVulnerabilities}`);
        console.log(`  Высокого уровня: ${report.summary.highVulnerabilities}`);
        console.log(`  Утечек данных: ${report.summary.dataLeaks}`);
        console.log(`  API уязвимостей: ${report.summary.apiVulnerabilities}`);
        console.log(`  XSS уязвимостей: ${report.summary.xssVulnerabilities}`);
        console.log(`  Уровень риска: ${report.riskAssessment.level}`);
        
        return report;
    }

    /**
     * ОЦЕНКА УРОВНЯ РИСКА
     */
    assessRiskLevel() {
        const criticalCount = this.results.vulnerabilities.filter(v => v.severity === 'CRITICAL').length;
        const highCount = this.results.vulnerabilities.filter(v => v.severity === 'HIGH').length;
        const mediumCount = this.results.vulnerabilities.filter(v => v.severity === 'MEDIUM').length;
        
        let level = 'LOW';
        let score = 0;
        
        score += criticalCount * 10;
        score += highCount * 5;
        score += mediumCount * 2;
        
        if (score >= 20) {
            level = 'CRITICAL';
        } else if (score >= 10) {
            level = 'HIGH';
        } else if (score >= 5) {
            level = 'MEDIUM';
        }
        
        return {
            level: level,
            score: score,
            criticalCount: criticalCount,
            highCount: highCount,
            mediumCount: mediumCount
        };
    }

    /**
     * СОХРАНЕНИЕ ОТЧЕТА
     */
    saveReport(report, filename = null) {
        if (!filename) {
            const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
            filename = `security_audit_${timestamp}.json`;
        }
        
        const reportData = JSON.stringify(report, null, 2);
        
        // В браузере - создание ссылки для скачивания
        if (typeof window !== 'undefined') {
            const blob = new Blob([reportData], { type: 'application/json' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename;
            a.click();
            URL.revokeObjectURL(url);
        }
        
        console.log(`📄 Отчет сохранен: ${filename}`);
        return filename;
    }

    /**
     * ЗАПУСК ПОЛНОГО АУДИТА
     */
    async runFullAudit() {
        console.log("🚀 ЗАПУСК КОМПЛЕКСНОГО АУДИТА БЕЗОПАСНОСТИ");
        console.log("=" * 60);
        
        try {
            // Этап 1: Анализ утечек данных
            if (this.config.enableDataExtraction) {
                await this.analyzeDataLeaks();
            }
            
            // Этап 2: Анализ API уязвимостей
            if (this.config.enableAPIAnalysis) {
                await this.analyzeAPIVulnerabilities();
            }
            
            // Этап 3: Анализ XSS уязвимостей
            if (this.config.enableXSSAnalysis) {
                await this.analyzeXSSVulnerabilities();
            }
            
            // Этап 4: Сканирование уязвимостей
            if (this.config.enableVulnerabilityScanning) {
                await this.scanVulnerabilities();
            }
            
            // Этап 5: Генерация рекомендаций
            this.generateRecommendations();
            
            // Генерация финального отчета
            const report = this.generateFinalReport();
            
            console.log("\n✅ КОМПЛЕКСНЫЙ АУДИТ ЗАВЕРШЕН");
            console.log("=" * 60);
            
            return report;
            
        } catch (error) {
            console.error("❌ ОШИБКА ПРИ ВЫПОЛНЕНИИ АУДИТА:", error);
            throw error;
        }
    }
}

// Экспорт для использования
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ComprehensiveSecurityAudit;
} else if (typeof window !== 'undefined') {
    window.ComprehensiveSecurityAudit = ComprehensiveSecurityAudit;
}

// Пример использования
if (typeof window !== 'undefined') {
    console.log("🔒 КОМПЛЕКСНЫЙ АУДИТ БЕЗОПАСНОСТИ ЗАГРУЖЕН");
    console.log("Используйте: new ComprehensiveSecurityAudit().init(url, token, jsContent).runFullAudit()");
    
    // Автоматический запуск при загрузке страницы (опционально)
    if (window.location.search.includes('audit=true')) {
        console.log("🔍 АВТОМАТИЧЕСКИЙ ЗАПУСК АУДИТА");
        
        const auditor = new ComprehensiveSecurityAudit();
        auditor.init(window.location.href)
            .runFullAudit()
            .then(report => {
                console.log("📊 АУДИТ ЗАВЕРШЕН:", report);
                auditor.saveReport(report);
            })
            .catch(error => {
                console.error("❌ ОШИБКА АУДИТА:", error);
            });
    }
}
