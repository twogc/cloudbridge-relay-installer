#!/usr/bin/env node

/**
 * Тестовый скрипт для CloudBridge AI Monitoring Agent
 * Проверяет основные функции агента
 */

const https = require('https');
const http = require('http');

// Конфигурация
const CONFIG = {
  agentUrl: process.env.AGENT_URL || 'http://localhost:8787',
  timeout: 10000,
  retries: 3
};

// Цвета для вывода
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m'
};

// Функции для вывода
function log(message, color = 'reset') {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

function logSuccess(message) {
  log(`✅ ${message}`, 'green');
}

function logError(message) {
  log(`❌ ${message}`, 'red');
}

function logWarning(message) {
  log(`⚠️  ${message}`, 'yellow');
}

function logInfo(message) {
  log(`ℹ️  ${message}`, 'blue');
}

// HTTP клиент
function makeRequest(url, options = {}) {
  return new Promise((resolve, reject) => {
    const urlObj = new URL(url);
    const isHttps = urlObj.protocol === 'https:';
    const client = isHttps ? https : http;
    
    const requestOptions = {
      hostname: urlObj.hostname,
      port: urlObj.port || (isHttps ? 443 : 80),
      path: urlObj.pathname + urlObj.search,
      method: options.method || 'GET',
      headers: {
        'Content-Type': 'application/json',
        'User-Agent': 'CloudBridge-AI-Agent-Test/1.0',
        ...options.headers
      },
      timeout: CONFIG.timeout
    };

    const req = client.request(requestOptions, (res) => {
      let data = '';
      
      res.on('data', (chunk) => {
        data += chunk;
      });
      
      res.on('end', () => {
        try {
          const jsonData = JSON.parse(data);
          resolve({
            status: res.statusCode,
            headers: res.headers,
            data: jsonData
          });
        } catch (error) {
          resolve({
            status: res.statusCode,
            headers: res.headers,
            data: data
          });
        }
      });
    });

    req.on('error', (error) => {
      reject(error);
    });

    req.on('timeout', () => {
      req.destroy();
      reject(new Error('Request timeout'));
    });

    if (options.body) {
      req.write(JSON.stringify(options.body));
    }
    
    req.end();
  });
}

// Тесты
async function testHealthCheck() {
  logInfo('Тестирование health check...');
  
  try {
    const response = await makeRequest(`${CONFIG.agentUrl}/health`);
    
    if (response.status === 200) {
      logSuccess('Health check прошел успешно');
      console.log('  Ответ:', JSON.stringify(response.data, null, 2));
      return true;
    } else {
      logError(`Health check вернул статус ${response.status}`);
      return false;
    }
  } catch (error) {
    logError(`Health check не удался: ${error.message}`);
    return false;
  }
}

async function testMetrics() {
  logInfo('Тестирование metrics endpoint...');
  
  try {
    const response = await makeRequest(`${CONFIG.agentUrl}/metrics`);
    
    if (response.status === 200) {
      logSuccess('Metrics endpoint работает');
      console.log('  Метрики:', JSON.stringify(response.data, null, 2));
      return true;
    } else {
      logError(`Metrics endpoint вернул статус ${response.status}`);
      return false;
    }
  } catch (error) {
    logError(`Metrics endpoint не удался: ${error.message}`);
    return false;
  }
}

async function testSecurityCheck() {
  logInfo('Тестирование security check...');
  
  try {
    const response = await makeRequest(`${CONFIG.agentUrl}/check`, {
      method: 'POST'
    });
    
    if (response.status === 200) {
      logSuccess('Security check выполнен');
      console.log('  Результат:', response.data.result);
      return true;
    } else {
      logError(`Security check вернул статус ${response.status}`);
      return false;
    }
  } catch (error) {
    logError(`Security check не удался: ${error.message}`);
    return false;
  }
}

async function testChat() {
  logInfo('Тестирование chat endpoint...');
  
  const testMessages = [
    '/status',
    '/history',
    'Привет, как дела?'
  ];
  
  let successCount = 0;
  
  for (const message of testMessages) {
    try {
      logInfo(`  Отправка сообщения: "${message}"`);
      
      const response = await makeRequest(`${CONFIG.agentUrl}/chat`, {
        method: 'POST',
        body: { message }
      });
      
      if (response.status === 200) {
        logSuccess(`  Ответ получен: ${response.data.response.substring(0, 100)}...`);
        successCount++;
      } else {
        logError(`  Ошибка: статус ${response.status}`);
      }
    } catch (error) {
      logError(`  Ошибка: ${error.message}`);
    }
  }
  
  return successCount === testMessages.length;
}

async function testWebSocket() {
  logInfo('Тестирование WebSocket соединения...');
  
  return new Promise((resolve) => {
    const WebSocket = require('ws');
    const wsUrl = CONFIG.agentUrl.replace('http', 'ws') + '/ws';
    
    try {
      const ws = new WebSocket(wsUrl);
      
      ws.on('open', () => {
        logSuccess('WebSocket соединение установлено');
        
        // Отправляем тестовое сообщение
        ws.send(JSON.stringify({
          message: '/status'
        }));
      });
      
      ws.on('message', (data) => {
        try {
          const message = JSON.parse(data);
          logSuccess('WebSocket сообщение получено');
          console.log('  Данные:', JSON.stringify(message, null, 2));
          
          ws.close();
          resolve(true);
        } catch (error) {
          logError(`Ошибка парсинга WebSocket сообщения: ${error.message}`);
          ws.close();
          resolve(false);
        }
      });
      
      ws.on('error', (error) => {
        logError(`WebSocket ошибка: ${error.message}`);
        resolve(false);
      });
      
      ws.on('close', () => {
        logInfo('WebSocket соединение закрыто');
      });
      
      // Таймаут
      setTimeout(() => {
        logWarning('WebSocket тест превысил таймаут');
        ws.close();
        resolve(false);
      }, 5000);
      
    } catch (error) {
      logError(`Ошибка создания WebSocket: ${error.message}`);
      resolve(false);
    }
  });
}

async function testLoad() {
  logInfo('Тестирование нагрузки...');
  
  const concurrentRequests = 5;
  const promises = [];
  
  for (let i = 0; i < concurrentRequests; i++) {
    promises.push(
      makeRequest(`${CONFIG.agentUrl}/health`)
        .then(response => ({ success: response.status === 200, id: i }))
        .catch(error => ({ success: false, id: i, error: error.message }))
    );
  }
  
  const results = await Promise.all(promises);
  const successCount = results.filter(r => r.success).length;
  
  if (successCount === concurrentRequests) {
    logSuccess(`Все ${concurrentRequests} запросов выполнены успешно`);
    return true;
  } else {
    logError(`${concurrentRequests - successCount} из ${concurrentRequests} запросов не удались`);
    return false;
  }
}

// Основная функция тестирования
async function runTests() {
  log('🚀 Запуск тестов CloudBridge AI Monitoring Agent', 'bright');
  log(`📍 Агент: ${CONFIG.agentUrl}`, 'cyan');
  log('');
  
  const tests = [
    { name: 'Health Check', fn: testHealthCheck },
    { name: 'Metrics', fn: testMetrics },
    { name: 'Security Check', fn: testSecurityCheck },
    { name: 'Chat', fn: testChat },
    { name: 'WebSocket', fn: testWebSocket },
    { name: 'Load Test', fn: testLoad }
  ];
  
  let passedTests = 0;
  let totalTests = tests.length;
  
  for (const test of tests) {
    log(`\n📋 Тест: ${test.name}`, 'magenta');
    log('─'.repeat(50));
    
    try {
      const result = await test.fn();
      if (result) {
        passedTests++;
      }
    } catch (error) {
      logError(`Неожиданная ошибка в тесте ${test.name}: ${error.message}`);
    }
  }
  
  // Результаты
  log('\n' + '='.repeat(50));
  log('📊 РЕЗУЛЬТАТЫ ТЕСТИРОВАНИЯ', 'bright');
  log('='.repeat(50));
  
  if (passedTests === totalTests) {
    logSuccess(`Все тесты прошли успешно! (${passedTests}/${totalTests})`);
  } else {
    logError(`Провалено тестов: ${totalTests - passedTests}/${totalTests}`);
  }
  
  log('');
  log('💡 Рекомендации:', 'yellow');
  
  if (passedTests < totalTests) {
    log('  • Проверьте, что агент запущен и доступен');
    log('  • Убедитесь в правильности URL агента');
    log('  • Проверьте логи агента на наличие ошибок');
    log('  • Убедитесь в наличии всех необходимых переменных окружения');
  } else {
    log('  • Агент работает корректно');
    log('  • Можно переходить к интеграции');
    log('  • Рекомендуется настроить мониторинг');
  }
  
  return passedTests === totalTests;
}

// Обработка аргументов командной строки
function parseArgs() {
  const args = process.argv.slice(2);
  
  for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
      case '--url':
      case '-u':
        if (i + 1 < args.length) {
          CONFIG.agentUrl = args[i + 1];
          i++;
        }
        break;
      case '--timeout':
      case '-t':
        if (i + 1 < args.length) {
          CONFIG.timeout = parseInt(args[i + 1]);
          i++;
        }
        break;
      case '--help':
      case '-h':
        showHelp();
        process.exit(0);
        break;
    }
  }
}

function showHelp() {
  console.log(`
CloudBridge AI Monitoring Agent - Test Script

Usage: node test-agent.js [options]

Options:
  -u, --url <url>      URL агента (default: http://localhost:8787)
  -t, --timeout <ms>   Таймаут запросов в миллисекундах (default: 10000)
  -h, --help           Показать эту справку

Examples:
  node test-agent.js
  node test-agent.js --url https://my-agent.example.com
  node test-agent.js --timeout 5000
`);
}

// Запуск
if (require.main === module) {
  parseArgs();
  runTests()
    .then(success => {
      process.exit(success ? 0 : 1);
    })
    .catch(error => {
      logError(`Критическая ошибка: ${error.message}`);
      process.exit(1);
    });
}

module.exports = {
  runTests,
  testHealthCheck,
  testMetrics,
  testSecurityCheck,
  testChat,
  testWebSocket,
  testLoad
}; 