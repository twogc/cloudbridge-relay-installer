

// Типы для окружения
interface Env {
  OPENAI_API_KEY?: string; // Опционально, если используем OpenAI
  CLOUDFLARE_API_TOKEN?: string; // Для Cloudflare AI
  RELAY_METRICS_URL: string;
  ML_API_URL: string;
  SIEM_API_URL: string;
  SIEM_WEBHOOK_URL: string;
  CHECK_INTERVAL: string;
  ANOMALY_THRESHOLD: string;
  SIEM_API_KEY: string;
  AI_MODEL?: string; // Модель Cloudflare AI
  AI: CloudflareAI; // Cloudflare AI binding
}

// Интерфейсы для данных
interface MetricsData {
  [key: string]: {
    value: number;
    labels: Record<string, string>;
  };
}

interface MLRequestData {
  timestamp: string;
  bytes_sent: number;
  bytes_received: number;
  packets_sent: number;
  packets_received: number;
  connection_duration: number;
  requests_per_second: number;
  error_rate: number;
  user_id_hash: number;
  session_duration: number;
  requests_in_session: number;
  unique_endpoints: number;
  failed_auth_attempts: number;
  suspicious_ips: number;
  rate_limit_violations: number;
  ssl_errors: number;
}

interface AnomalyResult {
  anomaly_score: number;
  is_anomaly: boolean;
  confidence: number;
  anomaly_type: string;
  recommendations: string[];
  timestamp: string;
}

interface SIEMAlert {
  id: string;
  timestamp: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  title: string;
  description: string;
  source: string;
  event_type: string;
  ip_address?: string;
  user_id?: string;
  indicators: string[];
  tags: string[];
  status: string;
  metadata: Record<string, any>;
}

// Утилиты для валидации и обработки ошибок
class ValidationUtils {
  static validateMetricsData(data: any): data is MetricsData {
    return data && typeof data === 'object' && Object.keys(data).length > 0;
  }

  static validateMLRequestData(data: any): data is MLRequestData {
    return data && 
           typeof data.timestamp === 'string' &&
           typeof data.bytes_sent === 'number' &&
           typeof data.bytes_received === 'number';
  }

  static validateAnomalyResult(data: any): data is AnomalyResult {
    return data && 
           typeof data.anomaly_score === 'number' &&
           typeof data.is_anomaly === 'boolean';
  }
}

// Поддержка Cloudflare AI
interface CloudflareAI {
  run: (model: string, inputs: any) => Promise<any>;
}

// Retry утилита для API вызовов
async function fetchWithRetry(
  url: string, 
  options: RequestInit, 
  retries = 3,
  delay = 1000
): Promise<Response> {
  for (let i = 0; i < retries; i++) {
    try {
      const response = await fetch(url, options);
      if (response.ok) {
        return response;
      }
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    } catch (error) {
      if (i === retries - 1) {
        throw error;
      }
      console.warn(`Попытка ${i + 1} не удалась, повтор через ${delay}ms:`, error);
      await new Promise(resolve => setTimeout(resolve, delay * (i + 1)));
    }
  }
  throw new Error('Все попытки исчерпаны');
}

// Мониторинг производительности
class PerformanceMonitor {
  private metrics: Map<string, number[]> = new Map();
  private startTimes: Map<string, number> = new Map();

  startTimer(name: string): void {
    this.startTimes.set(name, Date.now());
  }

  endTimer(name: string): number {
    const startTime = this.startTimes.get(name);
    if (!startTime) {
      throw new Error(`Таймер ${name} не был запущен`);
    }
    
    const duration = Date.now() - startTime;
    this.recordMetric(name, duration);
    this.startTimes.delete(name);
    return duration;
  }

  recordMetric(name: string, value: number): void {
    if (!this.metrics.has(name)) {
      this.metrics.set(name, []);
    }
    this.metrics.get(name)!.push(value);
  }

  getMetrics(): Record<string, { avg: number; min: number; max: number; count: number }> {
    const result: Record<string, { avg: number; min: number; max: number; count: number }> = {};
    
    for (const [name, values] of this.metrics.entries()) {
      const avg = values.reduce((a, b) => a + b, 0) / values.length;
      const min = Math.min(...values);
      const max = Math.max(...values);
      
      result[name] = {
        avg: Math.round(avg * 100) / 100,
        min,
        max,
        count: values.length
      };
    }
    
    return result;
  }

  reset(): void {
    this.metrics.clear();
    this.startTimes.clear();
  }
}

// Глобальный экземпляр мониторинга производительности
const performanceMonitor = new PerformanceMonitor();

// Класс для работы с Cloudflare AI
class CloudflareAIService {
  constructor(private ai: CloudflareAI, private model: string = '@cf/meta/llama-2-7b-chat-int8') {}

  async generateResponse(prompt: string, context: string = ''): Promise<string> {
    performanceMonitor.startTimer('ai_generation');
    
    try {
      const fullPrompt = `Ты - AI агент безопасности для системы CloudBridge Relay.

Контекст: ${context}

Запрос пользователя: ${prompt}

Ответь кратко и по делу. Если это команда - выполни её. Если это вопрос - дай полезный ответ.`;

      const response = await this.ai.run(this.model, {
        messages: [
          {
            role: 'system',
            content: 'Ты - AI агент безопасности. Отвечай кратко и по делу.'
          },
          {
            role: 'user',
            content: fullPrompt
          }
        ]
      });

      performanceMonitor.endTimer('ai_generation');
      return response.response || 'Извините, не удалось сгенерировать ответ.';
    } catch (error) {
      performanceMonitor.recordMetric('ai_generation_error', 1);
      console.error('Ошибка генерации AI:', error);
      return 'Извините, произошла ошибка при обработке запроса.';
    }
  }

  async analyzeSecurityData(metrics: MetricsData, anomalyResult: AnomalyResult): Promise<string> {
    const prompt = `Проанализируй данные безопасности:

Метрики: ${JSON.stringify(metrics, null, 2)}
Результат аномалии: ${JSON.stringify(anomalyResult, null, 2)}

Дай краткий анализ и рекомендации по безопасности.`;

    return await this.generateResponse(prompt);
  }
}

// Класс для сбора метрик Relay
class RelayMetricsCollector {
  constructor(private metricsUrl: string) {}

  async collectMetrics(): Promise<MetricsData | null> {
    performanceMonitor.startTimer('metrics_collection');
    
    try {
      const response = await fetchWithRetry(this.metricsUrl, {}, 3, 1000);
      const text = await response.text();
      
      if (!text.trim()) {
        throw new Error('Пустой ответ от метрик');
      }
      
      const metrics = this.parsePrometheusMetrics(text);
      
      if (!ValidationUtils.validateMetricsData(metrics)) {
        throw new Error('Невалидные данные метрик');
      }
      
      performanceMonitor.endTimer('metrics_collection');
      return metrics;
    } catch (error) {
      performanceMonitor.recordMetric('metrics_collection_error', 1);
      console.error('Ошибка сбора метрик:', error);
      return null;
    }
  }

  private parsePrometheusMetrics(metricsText: string): MetricsData {
    const metrics: MetricsData = {};
    const lines = metricsText.split('\n');
    
    for (const line of lines) {
      if (line.startsWith('#') || !line.trim()) continue;
      
      const match = line.match(/^([a-zA-Z_:][a-zA-Z0-9_:]*)\{?([^}]*)\}?\s+([0-9.-]+)/);
      if (match) {
        const [, name, labels, value] = match;
        if (name && value) {
          const parsedLabels = this.parseLabels(labels || '');
          metrics[name] = {
            value: parseFloat(value),
            labels: parsedLabels
          };
        }
      }
    }
    
    return metrics;
  }

  private parseLabels(labelsStr: string): Record<string, string> {
    if (!labelsStr) return {};
    
    const labels: Record<string, string> = {};
    const matches = labelsStr.matchAll(/([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*"([^"]*)"/g);
    
    for (const match of matches) {
      const [, key, value] = match;
      if (key && value) {
        labels[key] = value;
      }
    }
    
    return labels;
  }

  transformForML(metrics: MetricsData): MLRequestData {
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
  constructor(
    private apiUrl: string
  ) {}

  async detectAnomaly(data: MLRequestData): Promise<AnomalyResult> {
    performanceMonitor.startTimer('anomaly_detection');
    
    try {
      if (!ValidationUtils.validateMLRequestData(data)) {
        throw new Error('Невалидные данные для ML API');
      }

      const response = await fetchWithRetry(this.apiUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
      }, 3, 2000);

      const result = await response.json();
      
      if (!ValidationUtils.validateAnomalyResult(result)) {
        throw new Error('Невалидный ответ от ML API');
      }
      
      performanceMonitor.endTimer('anomaly_detection');
      return result;
    } catch (error) {
      performanceMonitor.recordMetric('anomaly_detection_error', 1);
      console.error('Ошибка детекции аномалий:', error);
      return {
        anomaly_score: 0,
        is_anomaly: false,
        confidence: 0,
        anomaly_type: 'error',
        recommendations: [`Ошибка ML API: ${error instanceof Error ? error.message : 'Unknown error'}`],
        timestamp: new Date().toISOString()
      };
    }
  }
}

// Класс для работы с SIEM
class SIEMAlertManager {
  constructor(
    private apiUrl: string, 
    private webhookUrl: string, 
    private apiKey: string
  ) {}

  async sendAlert(alert: SIEMAlert): Promise<any> {
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

  async sendWebhook(webhookData: any): Promise<boolean> {
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

  createAlert(anomalyResult: AnomalyResult, metrics: MetricsData, context: Record<string, any> = {}): SIEMAlert {
    const severity = this.mapAnomalyScoreToSeverity(anomalyResult.anomaly_score);
    
    return {
      id: `ai_alert_${Date.now()}`,
      timestamp: new Date().toISOString(),
      severity,
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

  private mapAnomalyScoreToSeverity(score: number): 'low' | 'medium' | 'high' | 'critical' {
    if (score > 0.9) return 'critical';
    if (score > 0.8) return 'high';
    if (score > 0.7) return 'medium';
    return 'low';
  }
}

// Основной класс AI агента мониторинга
class SecurityMonitoringAgent {
  private metricsCollector: RelayMetricsCollector;
  private anomalyDetector: MLAnomalyDetector;
  private alertManager: SIEMAlertManager;
  private aiService?: CloudflareAIService;
  private state: {
    lastCheck: string | null;
    anomaliesDetected: number;
    totalChecks: number;
    isRunning: boolean;
  };

  constructor(env: Env) {
    this.metricsCollector = new RelayMetricsCollector(env.RELAY_METRICS_URL);
    this.anomalyDetector = new MLAnomalyDetector(env.ML_API_URL);
    this.alertManager = new SIEMAlertManager(env.SIEM_API_URL, env.SIEM_WEBHOOK_URL, env.SIEM_API_KEY);
    
    // Инициализация AI сервиса
    if (env.AI) {
      this.aiService = new CloudflareAIService(env.AI, env.AI_MODEL || '@cf/meta/llama-2-7b-chat-int8');
    }
    
    this.state = {
      lastCheck: null,
      anomaliesDetected: 0,
      totalChecks: 0,
      isRunning: false
    };
  }

  // Метод для обработки чат-сообщений с использованием AI
  async processChatMessage(message: string): Promise<string> {
    try {
      // Обработка команд
      if (message.startsWith('/')) {
        return await this.handleCommand(message, '');
      }

      // Использование AI для генерации ответа
      if (this.aiService) {
        const context = `Текущее время: ${new Date().toISOString()}
Состояние системы: ${this.state.isRunning ? 'активна' : 'остановлена'}
Всего проверок: ${this.state.totalChecks}
Обнаружено аномалий: ${this.state.anomaliesDetected}

Доступные команды:
- /status - получить статус системы
- /check - выполнить проверку безопасности
- /start - запустить мониторинг
- /stop - остановить мониторинг
- /history - показать историю действий`;

        return await this.aiService.generateResponse(message, context);
      }

      // Fallback ответ без AI
      return `Привет! Я AI агент безопасности для системы CloudBridge Relay.

Контекст:
- Текущее время: ${new Date().toISOString()}
- Состояние системы: ${this.state.isRunning ? 'активна' : 'остановлена'}
- Всего проверок: ${this.state.totalChecks}
- Обнаружено аномалий: ${this.state.anomaliesDetected}

Доступные команды:
- /status - получить статус системы
- /check - выполнить проверку безопасности
- /start - запустить мониторинг
- /stop - остановить мониторинг
- /history - показать историю действий

Используйте команды для взаимодействия со мной.`;
    } catch (error) {
      console.error('Ошибка обработки сообщения:', error);
      return 'Извините, произошла ошибка при обработке вашего сообщения.';
    }
  }

  // Обработка команд
  private async handleCommand(command: string, aiResponse: string): Promise<string> {
    const [cmd, ...args] = command.split(' ');
    
    switch (cmd?.toLowerCase()) {
      case '/status':
        return await this.getStatus();
      
      case '/check':
        return await this.performCheck();
      
      case '/start':
        return await this.startMonitoring();
      
      case '/stop':
        return await this.stopMonitoring();
      
      case '/history':
        return await this.getActionHistory();
      
      default:
        return `Неизвестная команда: ${cmd}. ${aiResponse}`;
    }
  }

  // Выполнение проверки безопасности
  async performCheck(): Promise<string> {
    try {
      this.state.totalChecks++;
      
      // Сбор метрик
      const metrics = await this.metricsCollector.collectMetrics();
      if (!metrics) {
        return '❌ Ошибка сбора метрик';
      }
      
      // Преобразование для ML API
      const mlData = this.metricsCollector.transformForML(metrics);
      
      // Детекция аномалий
      const anomalyResult = await this.anomalyDetector.detectAnomaly(mlData);
      
      // Обработка результата
      if (anomalyResult.is_anomaly && anomalyResult.anomaly_score > 0.7) {
        this.state.anomaliesDetected++;
        
        // Создание и отправка алерта
        const alert = this.alertManager.createAlert(anomalyResult, metrics);
        await this.alertManager.sendAlert(alert);
        
        return `🚨 Аномалия обнаружена!

Тип: ${anomalyResult.anomaly_type}
Счет: ${anomalyResult.anomaly_score.toFixed(3)}
Уверенность: ${anomalyResult.confidence.toFixed(3)}

Рекомендации:
${anomalyResult.recommendations.map(r => `• ${r}`).join('\n')}

Алерт отправлен в SIEM (ID: ${alert.id})`;
      }
      
      this.state.lastCheck = new Date().toISOString();
      
      return `✅ Проверка завершена

Метрики собраны: ${Object.keys(metrics).length}
Аномалий не обнаружено
Последняя проверка: ${this.state.lastCheck}`;
      
    } catch (error) {
      console.error('Ошибка выполнения проверки:', error);
      return `❌ Ошибка проверки: ${error instanceof Error ? error.message : 'Unknown error'}`;
    }
  }

  // Получение статуса системы
  async getStatus(): Promise<string> {
    const uptime = this.state.lastCheck ? 
      Math.floor((Date.now() - new Date(this.state.lastCheck).getTime()) / 1000) : 0;
    
    const performanceMetrics = performanceMonitor.getMetrics();
    
    return `🤖 Статус AI-агента мониторинга

Состояние: ${this.state.isRunning ? '🟢 Активен' : '🔴 Остановлен'}
Всего проверок: ${this.state.totalChecks}
Обнаружено аномалий: ${this.state.anomaliesDetected}
Последняя проверка: ${this.state.lastCheck || 'Нет'}
Время работы: ${uptime} сек

📊 Метрики производительности:
${Object.entries(performanceMetrics).map(([name, metrics]) => 
  `• ${name}: avg=${metrics.avg}ms, min=${metrics.min}ms, max=${metrics.max}ms (${metrics.count} calls)`
).join('\n')}`;
  }

  // Запуск мониторинга
  async startMonitoring(): Promise<string> {
    if (this.state.isRunning) {
      return '⚠️ Мониторинг уже запущен';
    }
    
    this.state.isRunning = true;
    return '🟢 Мониторинг запущен';
  }

  // Остановка мониторинга
  async stopMonitoring(): Promise<string> {
    if (!this.state.isRunning) {
      return '⚠️ Мониторинг уже остановлен';
    }
    
    this.state.isRunning = false;
    return '🔴 Мониторинг остановлен';
  }

  // Получение истории действий
  async getActionHistory(): Promise<string> {
    return `📝 История действий:

Последние действия:
• Проверок выполнено: ${this.state.totalChecks}
• Аномалий обнаружено: ${this.state.anomaliesDetected}
• Последняя проверка: ${this.state.lastCheck || 'Нет'}`;
  }

  // Получение состояния агента
  getState() {
    return { ...this.state };
  }
}

// Глобальный экземпляр агента
let agent: SecurityMonitoringAgent | null = null;

// Основной обработчик запросов
export default {
  async fetch(request: Request, env: Env) {
    // Проверяем наличие AI (Cloudflare AI или OpenAI)
    if (!env.AI && !env.OPENAI_API_KEY && !env.CLOUDFLARE_API_TOKEN) {
      console.error("Neither AI binding, OPENAI_API_KEY, nor CLOUDFLARE_API_TOKEN is set");
      return new Response("AI binding, OPENAI_API_KEY, or CLOUDFLARE_API_TOKEN must be set", { status: 500 });
    }

    // Инициализация агента
    if (!agent) {
      agent = new SecurityMonitoringAgent(env);
    }

    const url = new URL(request.url);
    
    // Health check
    if (url.pathname === '/health') {
      return new Response(JSON.stringify({ 
        status: 'ok', 
        timestamp: new Date().toISOString(),
        agent: 'SecurityMonitoringAgent'
      }), {
        headers: { 'Content-Type': 'application/json' }
      });
    }

    // Performance metrics endpoint
    if (url.pathname === '/performance') {
      return new Response(JSON.stringify({
        timestamp: new Date().toISOString(),
        metrics: performanceMonitor.getMetrics(),
        agent_state: agent.getState()
      }, null, 2), {
        headers: { 'Content-Type': 'application/json' }
      });
    }

    // Metrics endpoint
    if (url.pathname === '/metrics' && request.method === 'GET') {
      const state = agent.getState();
      return new Response(JSON.stringify(state), {
        headers: { 'Content-Type': 'application/json' }
      });
    }

    // Security check endpoint
    if (url.pathname === '/check' && request.method === 'POST') {
      const result = await agent.performCheck();
      return new Response(JSON.stringify({ 
        result,
        timestamp: new Date().toISOString()
      }), {
        headers: { 'Content-Type': 'application/json' }
      });
    }

    // Chat endpoint
    if (url.pathname === '/chat' && request.method === 'POST') {
      try {
        const body = await request.json();
        const message = body.message || '';
        
        const response = await agent.processChatMessage(message);
        
        return new Response(JSON.stringify({ 
          response,
          timestamp: new Date().toISOString()
        }), {
          headers: { 'Content-Type': 'application/json' }
        });
      } catch (error) {
        return new Response(JSON.stringify({ 
          error: 'Invalid request body' 
        }), { 
          status: 400,
          headers: { 'Content-Type': 'application/json' }
        });
      }
    }

    // Главная страница с информацией
    return new Response(`
AI Security Monitoring Agent - CloudBridge Relay

Endpoints:
- /health - Health check
- /performance - Performance metrics (GET)
- /metrics - Agent metrics (GET)
- /check - Perform security check (POST)
- /chat - Chat with agent (POST)

Example usage:
curl -X POST http://localhost:8787/chat \\
  -H "Content-Type: application/json" \\
  -d '{"message": "/status"}'

curl http://localhost:8787/performance
    `, {
      headers: { 'Content-Type': 'text/plain' }
    });
  },
}; 