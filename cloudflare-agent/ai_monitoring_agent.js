import { AIChatAgent, agentContext, createDataStreamResponse, processToolCalls, createOpenAI, streamText } from "agents-sdk";

// Конфигурация
const CONFIG = {
  // API endpoints
  RELAY_METRICS_URL: 'http://localhost:9090/metrics',
  ML_API_URL: 'http://localhost:8000/detect',
  SIEM_API_URL: 'http://localhost:8081/api/v1/alerts',
  SIEM_WEBHOOK_URL: 'http://localhost:8081/api/v1/webhook',
  
  // Настройки мониторинга
  CHECK_INTERVAL: 30000, // 30 секунд
  ANOMALY_THRESHOLD: 0.7,
  
  // Аутентификация
  API_KEYS: {
    ML_API: 'your-ml-api-key',
    SIEM_API: 'your-siem-api-key'
  }
};

// Класс для работы с метриками Relay
class RelayMetricsCollector {
  constructor(metricsUrl) {
    this.metricsUrl = metricsUrl;
  }

  async collectMetrics() {
    try {
      const response = await fetch(this.metricsUrl);
      const text = await response.text();
      return this.parsePrometheusMetrics(text);
    } catch (error) {
      console.error('Ошибка сбора метрик:', error);
      return null;
    }
  }

  parsePrometheusMetrics(metricsText) {
    const metrics = {};
    const lines = metricsText.split('\n');
    
    for (const line of lines) {
      if (line.startsWith('#') || !line.trim()) continue;
      
      const match = line.match(/^([a-zA-Z_:][a-zA-Z0-9_:]*)\{?([^}]*)\}?\s+([0-9.-]+)/);
      if (match) {
        const [, name, labels, value] = match;
        const parsedLabels = this.parseLabels(labels);
        metrics[name] = {
          value: parseFloat(value),
          labels: parsedLabels
        };
      }
    }
    
    return metrics;
  }

  parseLabels(labelsStr) {
    if (!labelsStr) return {};
    
    const labels = {};
    const matches = labelsStr.matchAll(/([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*"([^"]*)"/g);
    
    for (const match of matches) {
      const [, key, value] = match;
      labels[key] = value;
    }
    
    return labels;
  }

  // Преобразование метрик в формат для ML API
  transformForML(metrics) {
    const now = new Date().toISOString();
    
    return {
      timestamp: now,
      bytes_sent: metrics.relay_bytes_sent?.value || 0,
      bytes_received: metrics.relay_bytes_received?.value || 0,
      packets_sent: metrics.relay_packets_sent?.value || 0,
      packets_received: metrics.relay_packets_received?.value || 0,
      connection_duration: metrics.relay_connection_duration_seconds?.value || 0,
      requests_per_second: metrics.relay_requests_per_second?.value || 0,
      error_rate: metrics.relay_error_rate?.value || 0,
      user_id_hash: metrics.relay_user_id_hash?.value || 0,
      session_duration: metrics.relay_session_duration_seconds?.value || 0,
      requests_in_session: metrics.relay_requests_in_session?.value || 0,
      unique_endpoints: metrics.relay_unique_endpoints?.value || 0,
      failed_auth_attempts: metrics.relay_auth_failures_total?.value || 0,
      suspicious_ips: metrics.relay_suspicious_ips_total?.value || 0,
      rate_limit_violations: metrics.relay_rate_limit_violations_total?.value || 0,
      ssl_errors: metrics.relay_ssl_errors_total?.value || 0
    };
  }
}

// Класс для работы с ML API
class MLAnomalyDetector {
  constructor(apiUrl, apiKey) {
    this.apiUrl = apiUrl;
    this.apiKey = apiKey;
  }

  async detectAnomaly(data) {
    try {
      const response = await fetch(this.apiUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.apiKey}`
        },
        body: JSON.stringify(data)
      });

      if (!response.ok) {
        throw new Error(`ML API error: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Ошибка детекции аномалий:', error);
      return {
        anomaly_score: 0,
        is_anomaly: false,
        confidence: 0,
        anomaly_type: 'error',
        recommendations: [`Ошибка ML API: ${error.message}`]
      };
    }
  }
}

// Класс для работы с SIEM
class SIEMAlertManager {
  constructor(apiUrl, webhookUrl, apiKey) {
    this.apiUrl = apiUrl;
    this.webhookUrl = webhookUrl;
    this.apiKey = apiKey;
  }

  async sendAlert(alert) {
    try {
      const response = await fetch(this.apiUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.apiKey}`
        },
        body: JSON.stringify(alert)
      });

      if (!response.ok) {
        throw new Error(`SIEM API error: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Ошибка отправки алерта в SIEM:', error);
      return null;
    }
  }

  async sendWebhook(webhookData) {
    try {
      const response = await fetch(this.webhookUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.apiKey}`
        },
        body: JSON.stringify(webhookData)
      });

      return response.ok;
    } catch (error) {
      console.error('Ошибка отправки webhook:', error);
      return false;
    }
  }

  createAlert(anomalyResult, metrics, context = {}) {
    const severity = this.mapAnomalyScoreToSeverity(anomalyResult.anomaly_score);
    
    return {
      id: `ai_alert_${Date.now()}`,
      timestamp: new Date().toISOString(),
      severity: severity,
      title: `AI Detected Anomaly: ${anomalyResult.anomaly_type}`,
      description: `AI agent detected ${anomalyResult.anomaly_type} anomaly with score ${anomalyResult.anomaly_score.toFixed(3)}`,
      source: 'ai_monitoring_agent',
      event_type: 'ai_anomaly_detection',
      ip_address: context.source_ip,
      user_id: context.user_id,
      indicators: anomalyResult.recommendations,
      tags: ['ai', 'anomaly', 'automated'],
      status: 'new',
      metadata: {
        anomaly_score: anomalyResult.anomaly_score,
        confidence: anomalyResult.confidence,
        anomaly_type: anomalyResult.anomaly_type,
        recommendations: anomalyResult.recommendations,
        metrics_snapshot: metrics
      }
    };
  }

  mapAnomalyScoreToSeverity(score) {
    if (score > 0.9) return 'critical';
    if (score > 0.8) return 'high';
    if (score > 0.7) return 'medium';
    return 'low';
  }
}

// Основной AI-агент мониторинга
export class AIMonitoringAgent extends AIChatAgent {
  constructor() {
    super();
    
    // Инициализация компонентов
    this.metricsCollector = new RelayMetricsCollector(CONFIG.RELAY_METRICS_URL);
    this.anomalyDetector = new MLAnomalyDetector(CONFIG.ML_API_URL, CONFIG.API_KEYS.ML_API);
    this.alertManager = new SIEMAlertManager(CONFIG.SIEM_API_URL, CONFIG.SIEM_WEBHOOK_URL, CONFIG.API_KEYS.SIEM_API);
    
    // Состояние агента
    this.state = {
      lastCheck: null,
      anomaliesDetected: 0,
      totalChecks: 0,
      isRunning: false
    };
    
    // История действий
    this.actionHistory = [];
  }

  async onChatMessage(onFinish) {
    return agentContext.run(this, async () => {
      const message = await this.getLatestMessage();
      
      // Обработка команд
      if (message.text.startsWith('/')) {
        return await this.handleCommand(message.text);
      }
      
      // Обычный диалог
      return await this.handleDialog(message.text);
    });
  }

  async handleCommand(command) {
    const [cmd, ...args] = command.split(' ');
    
    switch (cmd.toLowerCase()) {
      case '/status':
        return await this.getStatus();
      
      case '/check':
        return await this.performCheck();
      
      case '/history':
        return await this.getActionHistory();
      
      case '/config':
        return await this.getConfiguration();
      
      case '/stop':
        return await this.stopMonitoring();
      
      case '/start':
        return await this.startMonitoring();
      
      default:
        return new Response(`Неизвестная команда: ${cmd}. Доступные команды: /status, /check, /history, /config, /start, /stop`);
    }
  }

  async handleDialog(message) {
    // Анализ сообщения пользователя и формирование ответа
    const response = await this.generateResponse(message);
    return new Response(response);
  }

  async performCheck() {
    try {
      this.state.totalChecks++;
      
      // Сбор метрик
      const metrics = await this.metricsCollector.collectMetrics();
      if (!metrics) {
        return new Response('❌ Ошибка сбора метрик');
      }
      
      // Преобразование для ML API
      const mlData = this.metricsCollector.transformForML(metrics);
      
      // Детекция аномалий
      const anomalyResult = await this.anomalyDetector.detectAnomaly(mlData);
      
      // Обработка результата
      if (anomalyResult.is_anomaly && anomalyResult.anomaly_score > CONFIG.ANOMALY_THRESHOLD) {
        this.state.anomaliesDetected++;
        
        // Создание и отправка алерта
        const alert = this.alertManager.createAlert(anomalyResult, metrics);
        await this.alertManager.sendAlert(alert);
        
        // Логирование действия
        this.logAction('anomaly_detected', {
          score: anomalyResult.anomaly_score,
          type: anomalyResult.anomaly_type,
          alert_id: alert.id
        });
        
        return new Response(`🚨 Аномалия обнаружена!
        
Тип: ${anomalyResult.anomaly_type}
Счет: ${anomalyResult.anomaly_score.toFixed(3)}
Уверенность: ${anomalyResult.confidence.toFixed(3)}

Рекомендации:
${anomalyResult.recommendations.map(r => `• ${r}`).join('\n')}

Алерт отправлен в SIEM (ID: ${alert.id})`);
      }
      
      this.state.lastCheck = new Date().toISOString();
      
      return new Response(`✅ Проверка завершена
      
Метрики собраны: ${Object.keys(metrics).length}
Аномалий не обнаружено
Последняя проверка: ${this.state.lastCheck}`);
      
    } catch (error) {
      console.error('Ошибка выполнения проверки:', error);
      return new Response(`❌ Ошибка проверки: ${error.message}`);
    }
  }

  async getStatus() {
    const uptime = this.state.lastCheck ? 
      Math.floor((Date.now() - new Date(this.state.lastCheck).getTime()) / 1000) : 0;
    
    return new Response(`🤖 Статус AI-агента мониторинга

Состояние: ${this.state.isRunning ? '🟢 Активен' : '🔴 Остановлен'}
Всего проверок: ${this.state.totalChecks}
Обнаружено аномалий: ${this.state.anomaliesDetected}
Последняя проверка: ${this.state.lastCheck || 'Нет'}
Время работы: ${uptime} сек

Конфигурация:
• Интервал проверки: ${CONFIG.CHECK_INTERVAL / 1000} сек
• Порог аномалий: ${CONFIG.ANOMALY_THRESHOLD}
• ML API: ${CONFIG.ML_API_URL}
• SIEM API: ${CONFIG.SIEM_API_URL}`);
  }

  async getActionHistory() {
    const recentActions = this.actionHistory.slice(-10);
    
    if (recentActions.length === 0) {
      return new Response('📝 История действий пуста');
    }
    
    const historyText = recentActions.map(action => 
      `[${action.timestamp}] ${action.type}: ${action.details}`
    ).join('\n');
    
    return new Response(`📝 Последние действия:

${historyText}`);
  }

  async getConfiguration() {
    return new Response(`⚙️ Конфигурация агента:

API Endpoints:
• Relay Metrics: ${CONFIG.RELAY_METRICS_URL}
• ML API: ${CONFIG.ML_API_URL}
• SIEM API: ${CONFIG.SIEM_API_URL}
• SIEM Webhook: ${CONFIG.SIEM_WEBHOOK_URL}

Настройки:
• Интервал проверки: ${CONFIG.CHECK_INTERVAL}ms
• Порог аномалий: ${CONFIG.ANOMALY_THRESHOLD}

Аутентификация:
• ML API Key: ${CONFIG.API_KEYS.ML_API ? '✅ Настроен' : '❌ Не настроен'}
• SIEM API Key: ${CONFIG.API_KEYS.SIEM_API ? '✅ Настроен' : '❌ Не настроен'}`);
  }

  async startMonitoring() {
    if (this.state.isRunning) {
      return new Response('⚠️ Мониторинг уже запущен');
    }
    
    this.state.isRunning = true;
    this.logAction('monitoring_started', {});
    
    // Запуск периодических проверок
    this.monitoringInterval = setInterval(async () => {
      if (this.state.isRunning) {
        await this.performCheck();
      }
    }, CONFIG.CHECK_INTERVAL);
    
    return new Response('🟢 Мониторинг запущен');
  }

  async stopMonitoring() {
    if (!this.state.isRunning) {
      return new Response('⚠️ Мониторинг уже остановлен');
    }
    
    this.state.isRunning = false;
    if (this.monitoringInterval) {
      clearInterval(this.monitoringInterval);
    }
    
    this.logAction('monitoring_stopped', {});
    return new Response('🔴 Мониторинг остановлен');
  }

  async generateResponse(message) {
    // Простая логика генерации ответов
    const responses = {
      'привет': 'Привет! Я AI-агент мониторинга CloudBridge Relay. Чем могу помочь?',
      'помощь': 'Доступные команды:\n/status - статус агента\n/check - выполнить проверку\n/history - история действий\n/config - конфигурация\n/start - запустить мониторинг\n/stop - остановить мониторинг',
      'спасибо': 'Пожалуйста! Всегда готов помочь с мониторингом системы.',
      'default': 'Я AI-агент мониторинга. Используйте команды для управления или задайте вопрос о состоянии системы.'
    };
    
    const lowerMessage = message.toLowerCase();
    
    for (const [key, response] of Object.entries(responses)) {
      if (lowerMessage.includes(key)) {
        return response;
      }
    }
    
    return responses.default;
  }

  logAction(type, details) {
    this.actionHistory.push({
      timestamp: new Date().toISOString(),
      type,
      details: typeof details === 'string' ? details : JSON.stringify(details)
    });
    
    // Ограничиваем историю 100 записями
    if (this.actionHistory.length > 100) {
      this.actionHistory = this.actionHistory.slice(-100);
    }
  }
}

// Экспорт для использования в Cloudflare Worker
export default {
  async fetch(request, env, ctx) {
    const agent = new AIMonitoringAgent();
    
    // Обработка WebSocket для real-time взаимодействия
    if (request.headers.get('Upgrade') === 'websocket') {
      return handleWebSocket(request, agent);
    }
    
    // Обработка HTTP запросов
    return handleHTTP(request, agent);
  }
};

// Обработка WebSocket соединений
async function handleWebSocket(request, agent) {
  const { 0: client, 1: server } = new WebSocketPair();
  
  server.accept();
  
  server.addEventListener('message', async (event) => {
    try {
      const message = JSON.parse(event.data);
      const response = await agent.handleCommand(message.command || message.text);
      server.send(JSON.stringify({
        type: 'response',
        data: await response.text()
      }));
    } catch (error) {
      server.send(JSON.stringify({
        type: 'error',
        error: error.message
      }));
    }
  });
  
  server.addEventListener('close', () => {
    console.log('WebSocket соединение закрыто');
  });
  
  return new Response(null, {
    status: 101,
    webSocket: client
  });
}

// Обработка HTTP запросов
async function handleHTTP(request, agent) {
  const url = new URL(request.url);
  
  if (url.pathname === '/health') {
    return new Response(JSON.stringify({ status: 'ok', timestamp: new Date().toISOString() }), {
      headers: { 'Content-Type': 'application/json' }
    });
  }
  
  if (url.pathname === '/metrics') {
    const status = await agent.getStatus();
    return new Response(await status.text(), {
      headers: { 'Content-Type': 'text/plain' }
    });
  }
  
  if (url.pathname === '/check' && request.method === 'POST') {
    const response = await agent.performCheck();
    return new Response(await response.text(), {
      headers: { 'Content-Type': 'text/plain' }
    });
  }
  
  // Обработка чат-сообщений
  if (url.pathname === '/chat' && request.method === 'POST') {
    const body = await request.json();
    const response = await agent.handleDialog(body.message || '');
    return new Response(await response.text(), {
      headers: { 'Content-Type': 'text/plain' }
    });
  }
  
  return new Response('AI Monitoring Agent - CloudBridge Relay\n\nEndpoints:\n/health - Health check\n/metrics - Agent metrics\n/check - Perform check (POST)\n/chat - Chat with agent (POST)', {
    headers: { 'Content-Type': 'text/plain' }
  });
} 