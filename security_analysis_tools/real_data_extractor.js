// РЕАЛЬНЫЙ ЭКСТРАКТОР ДАННЫХ ИЗ БРАУЗЕРА
// Этот скрипт демонстрирует реальный доступ к данным пользователя

console.log("=== РЕАЛЬНЫЙ ДОСТУП К ДАННЫМ ПОЛЬЗОВАТЕЛЯ ===");

// 1. Извлечение всех данных из localStorage
function extractAllLocalStorage() {
    console.log("🔍 ВСЕ ДАННЫЕ ИЗ LOCALSTORAGE:");
    const data = {};
    for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        const value = localStorage.getItem(key);
        data[key] = value;
        console.log(`  ${key}: ${value}`);
    }
    return data;
}

// 2. Извлечение всех данных из sessionStorage
function extractAllSessionStorage() {
    console.log("🔍 ВСЕ ДАННЫЕ ИЗ SESSIONSTORAGE:");
    const data = {};
    for (let i = 0; i < sessionStorage.length; i++) {
        const key = sessionStorage.key(i);
        const value = sessionStorage.getItem(key);
        data[key] = value;
        console.log(`  ${key}: ${value}`);
    }
    return data;
}

// 3. Извлечение всех cookies
function extractAllCookies() {
    console.log("🔍 ВСЕ COOKIES:");
    const cookies = {};
    document.cookie.split(';').forEach(cookie => {
        const [name, value] = cookie.trim().split('=');
        cookies[name] = value;
        console.log(`  ${name}: ${value}`);
    });
    return cookies;
}

// 4. Извлечение данных из форм (включая скрытые поля)
function extractAllFormData() {
    console.log("🔍 ВСЕ ДАННЫЕ ИЗ ФОРМ:");
    const forms = document.querySelectorAll('form');
    const formData = {};
    
    forms.forEach((form, index) => {
        formData[`form_${index}`] = {};
        console.log(`  Форма ${index + 1}:`);
        
        // Все input поля
        const inputs = form.querySelectorAll('input');
        inputs.forEach(input => {
            formData[`form_${index}`][input.name || input.id] = input.value;
            console.log(`    ${input.name || input.id}: ${input.value} (тип: ${input.type})`);
        });
        
        // Все textarea поля
        const textareas = form.querySelectorAll('textarea');
        textareas.forEach(textarea => {
            formData[`form_${index}`][textarea.name || textarea.id] = textarea.value;
            console.log(`    ${textarea.name || textarea.id}: ${textarea.value}`);
        });
        
        // Все select поля
        const selects = form.querySelectorAll('select');
        selects.forEach(select => {
            formData[`form_${index}`][select.name || select.id] = select.value;
            console.log(`    ${select.name || select.id}: ${select.value}`);
        });
    });
    
    return formData;
}

// 5. Перехват всех API запросов
function interceptAllAPIRequests() {
    console.log("🔍 ПЕРЕХВАТ ВСЕХ API ЗАПРОСОВ:");
    
    // Перехват fetch
    const originalFetch = window.fetch;
    window.fetch = function(...args) {
        const url = args[0];
        const options = args[1] || {};
        
        console.log("  FETCH запрос:", url);
        console.log("  Метод:", options.method || 'GET');
        console.log("  Заголовки:", options.headers);
        console.log("  Тело запроса:", options.body);
        
        return originalFetch.apply(this, args).then(response => {
            response.clone().json().then(data => {
                console.log("  Ответ:", data);
            }).catch(() => {
                response.clone().text().then(text => {
                    console.log("  Ответ (текст):", text.substring(0, 200));
                });
            });
            return response;
        });
    };
    
    // Перехват XMLHttpRequest
    const originalXHROpen = XMLHttpRequest.prototype.open;
    const originalXHRSend = XMLHttpRequest.prototype.send;
    
    XMLHttpRequest.prototype.open = function(method, url, ...args) {
        console.log("  XHR запрос:", method, url);
        return originalXHROpen.apply(this, [method, url, ...args]);
    };
    
    XMLHttpRequest.prototype.send = function(data) {
        console.log("  XHR данные:", data);
        return originalXHRSend.apply(this, [data]);
    };
}

// 6. Извлечение данных из React/Redux состояния
function extractReactReduxState() {
    console.log("🔍 ПОПЫТКА ДОСТУПА К REACT/REDUX СОСТОЯНИЮ:");
    
    // Попытка доступа к React DevTools
    if (window.__REACT_DEVTOOLS_GLOBAL_HOOK__) {
        console.log("  React DevTools доступны");
        const hook = window.__REACT_DEVTOOLS_GLOBAL_HOOK__;
        console.log("  React версия:", hook.reactVersion);
    }
    
    // Попытка доступа к Redux DevTools
    if (window.__REDUX_DEVTOOLS_EXTENSION__) {
        console.log("  Redux DevTools доступны");
    }
    
    // Попытка найти Redux store
    if (window.store) {
        console.log("  Redux store найден:", window.store.getState());
    }
    
    // Поиск React компонентов
    const reactRoots = document.querySelectorAll('[data-reactroot], [data-reactid]');
    console.log("  Найдено React корней:", reactRoots.length);
}

// 7. Извлечение всех мета-данных
function extractAllMetaData() {
    console.log("🔍 ВСЕ МЕТА-ДАННЫЕ:");
    const metaData = {};
    
    // Мета-теги
    const metaTags = document.querySelectorAll('meta');
    metaTags.forEach(meta => {
        const name = meta.name || meta.property;
        const content = meta.content;
        if (name && content) {
            metaData[name] = content;
            console.log(`  ${name}: ${content}`);
        }
    });
    
    // Заголовки страницы
    metaData.title = document.title;
    console.log("  Title:", document.title);
    
    // Canonical URL
    const canonical = document.querySelector('link[rel="canonical"]');
    if (canonical) {
        metaData.canonical = canonical.href;
        console.log("  Canonical:", canonical.href);
    }
    
    return metaData;
}

// 8. Извлечение данных из скриптов
function extractAllScriptData() {
    console.log("🔍 ВСЕ ДАННЫЕ ИЗ СКРИПТОВ:");
    const scriptData = {};
    
    const scripts = document.querySelectorAll('script');
    scripts.forEach((script, index) => {
        if (script.src) {
            scriptData[`script_${index}_src`] = script.src;
            console.log(`  Скрипт ${index + 1}: ${script.src}`);
        }
        if (script.textContent && script.textContent.length > 0) {
            scriptData[`script_${index}_content`] = script.textContent.substring(0, 500);
            console.log(`  Встроенный скрипт ${index + 1}: ${script.textContent.substring(0, 100)}...`);
        }
    });
    
    return scriptData;
}

// 9. Поиск всех токенов и ключей
function extractAllTokens() {
    console.log("🔍 ПОИСК ВСЕХ ТОКЕНОВ И КЛЮЧЕЙ:");
    const tokens = {};
    
    const tokenPatterns = [
        /Bearer\s+([a-zA-Z0-9._-]+)/g,
        /token["\s]*[:=]["\s]*([a-zA-Z0-9._-]+)/g,
        /api[_-]?key["\s]*[:=]["\s]*([a-zA-Z0-9._-]+)/g,
        /auth[_-]?token["\s]*[:=]["\s]*([a-zA-Z0-9._-]+)/g,
        /secret["\s]*[:=]["\s]*([a-zA-Z0-9._-]+)/g,
        /password["\s]*[:=]["\s]*([a-zA-Z0-9._-]+)/g,
        /key["\s]*[:=]["\s]*([a-zA-Z0-9._-]+)/g
    ];
    
    const pageContent = document.documentElement.outerHTML;
    tokenPatterns.forEach((pattern, index) => {
        const matches = pageContent.match(pattern);
        if (matches) {
            tokens[`pattern_${index}`] = matches[1];
            console.log("  Найден токен:", matches[1]);
        }
    });
    
    return tokens;
}

// 10. Извлечение данных окружения
function extractEnvironmentData() {
    console.log("🔍 ДАННЫЕ ОКРУЖЕНИЯ:");
    const envData = {
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
            availWidth: screen.availWidth,
            availHeight: screen.availHeight
        },
        viewport: {
            width: window.innerWidth,
            height: window.innerHeight
        },
        timezone: Intl.DateTimeFormat().resolvedOptions().timeZone
    };
    
    Object.entries(envData).forEach(([key, value]) => {
        console.log(`  ${key}:`, value);
    });
    
    return envData;
}

// 11. Извлечение данных из Yandex Metrica
function extractYandexMetricaData() {
    console.log("🔍 ДАННЫЕ YANDEX METRICA:");
    const ymData = {};
    
    // Поиск Yandex Metrica cookies
    const ymCookies = document.cookie.split(';').filter(cookie => 
        cookie.includes('_ym_') || cookie.includes('yandex')
    );
    
    ymCookies.forEach(cookie => {
        const [name, value] = cookie.trim().split('=');
        ymData[name] = value;
        console.log(`  ${name}: ${value}`);
    });
    
    // Поиск Yandex Metrica в localStorage
    for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (key && (key.includes('_ym_') || key.includes('yandex'))) {
            const value = localStorage.getItem(key);
            ymData[key] = value;
            console.log(`  ${key}: ${value}`);
        }
    }
    
    return ymData;
}

// 12. Извлечение данных из Google Analytics
function extractGoogleAnalyticsData() {
    console.log("🔍 ДАННЫЕ GOOGLE ANALYTICS:");
    const gaData = {};
    
    // Поиск Google Analytics cookies
    const gaCookies = document.cookie.split(';').filter(cookie => 
        cookie.includes('_ga') || cookie.includes('_gid') || cookie.includes('_gat')
    );
    
    gaCookies.forEach(cookie => {
        const [name, value] = cookie.trim().split('=');
        gaData[name] = value;
        console.log(`  ${name}: ${value}`);
    });
    
    return gaData;
}

// Главная функция извлечения всех данных
function extractAllUserData() {
    console.log("🚀 ЗАПУСК ПОЛНОГО ИЗВЛЕЧЕНИЯ ДАННЫХ ПОЛЬЗОВАТЕЛЯ...");
    
    const allData = {
        localStorage: extractAllLocalStorage(),
        sessionStorage: extractAllSessionStorage(),
        cookies: extractAllCookies(),
        forms: extractAllFormData(),
        metaData: extractAllMetaData(),
        scriptData: extractAllScriptData(),
        tokens: extractAllTokens(),
        environment: extractEnvironmentData(),
        yandexMetrica: extractYandexMetricaData(),
        googleAnalytics: extractGoogleAnalyticsData()
    };
    
    // Настройка перехвата API запросов
    interceptAllAPIRequests();
    
    // Попытка доступа к React/Redux
    extractReactReduxState();
    
    console.log("✅ ПОЛНОЕ ИЗВЛЕЧЕНИЕ ДАННЫХ ЗАВЕРШЕНО");
    console.log("📊 СВОДКА ИЗВЛЕЧЕННЫХ ДАННЫХ:");
    console.log("  - localStorage:", Object.keys(allData.localStorage).length, "записей");
    console.log("  - sessionStorage:", Object.keys(allData.sessionStorage).length, "записей");
    console.log("  - cookies:", Object.keys(allData.cookies).length, "записей");
    console.log("  - формы:", Object.keys(allData.forms).length, "форм");
    console.log("  - токены:", Object.keys(allData.tokens).length, "токенов");
    
    return allData;
}

// Автоматический запуск
const extractedData = extractAllUserData();

// Экспорт для использования в консоли
window.extractAllUserData = extractAllUserData;
window.extractedData = extractedData;

console.log("🎯 ДАННЫЕ ЭКСПОРТИРОВАНЫ В window.extractedData");
