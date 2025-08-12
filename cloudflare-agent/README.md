# CloudBridge AI Monitoring Agent

AI-агент мониторинга безопасности для CloudBridge Relay, построенный с использованием [Vercel AI SDK](https://ai-sdk.dev/docs/introduction) и развернутый на Cloudflare Workers.

## 🚀 Возможности

- **AI-мониторинг безопасности**: Автоматический сбор и анализ метрик CloudBridge Relay
- **Детекция аномалий**: Интеграция с ML API для выявления подозрительной активности
- **SIEM интеграция**: Автоматическая отправка алертов в системы безопасности
- **Real-time чат**: WebSocket интерфейс для взаимодействия с AI-агентом
- **Автоматические действия**: Блокировка подозрительных IP, настройка правил безопасности
- **Мультитенантность**: Поддержка мониторинга нескольких клиентов

## 🏗️ Архитектура

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CloudBridge   │    │   AI Monitoring  │    │   ML Anomaly    │
│     Relay       │───▶│      Agent       │───▶│   Detection     │
│                 │    │                  │    │      API        │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   SIEM System   │
                       │   (Alerts)      │
                       └─────────────────┘
```

## 🛠️ Технологии

- **[Vercel AI SDK](https://ai-sdk.dev/docs/introduction)** - Универсальный SDK для работы с AI моделями
- **Cloudflare Workers** - Serverless платформа для развертывания
- **OpenAI GPT-4** - Модель для обработки естественного языка
- **TypeScript** - Типизированный JavaScript
- **React** - UI компоненты для веб-интерфейса
- **WebSocket** - Real-time коммуникация

## 📦 Установка

### Предварительные требования

- Node.js 18+
- npm или yarn
- Cloudflare аккаунт
- OpenAI API ключ

### Быстрый старт

```bash
# Клонирование репозитория
git clone <repository-url>
cd cloudbridge-relay-installer/cloudflare-agent

# Автоматическая настройка (рекомендуется)
./setup.sh

# Или ручная настройка:
npm install
cp env.example .env
# Отредактируйте .env файл с вашими API ключами

# Проверка типов (опционально)
npm run type-check

# Запуск в режиме разработки
npm run dev
```

## ⚙️ Конфигурация

### Переменные окружения

Создайте файл `.env` со следующими переменными:

```env
# OpenAI API Key (обязательно)
OPENAI_API_KEY=your_openai_api_key_here

# CloudBridge Relay Metrics URL
RELAY_METRICS_URL=http://localhost:8080/metrics

# ML Anomaly Detection API
ML_API_URL=http://localhost:5000/detect_anomaly
ML_API_KEY=your_ml_api_key_here

# SIEM Integration
SIEM_API_URL=http://localhost:9000/api/alerts
SIEM_WEBHOOK_URL=http://localhost:9000/webhook
SIEM_API_KEY=your_siem_api_key_here

# Monitoring Configuration
CHECK_INTERVAL=300
ANOMALY_THRESHOLD=0.7
```

### Cloudflare Workers конфигурация

Файл `wrangler.toml`:

```toml
name = "cloudbridge-ai-agent"
main = "src/index.ts"
compatibility_date = "2024-01-01"

[env.staging]
name = "cloudbridge-ai-agent-staging"

[env.production]
name = "cloudbridge-ai-agent-prod"

[[env.production.vars]]
key = "ENVIRONMENT"
value = "production"
```

## 🚀 Развертывание

### Автоматическое развертывание

Используйте скрипт развертывания:

```bash
# Установка и настройка
./deploy.sh install
./deploy.sh setup

# Развертывание в staging
./deploy.sh deploy

# Развертывание в production
./deploy.sh deploy:prod

# Просмотр логов
./deploy.sh logs production

# Очистка
./deploy.sh cleanup staging
```

### Ручное развертывание

```bash
# Установка зависимостей
npm install

# Проверка типов
npm run type-check

# Сборка
npm run build

# Развертывание
wrangler deploy
```

## 📡 API Endpoints

### Health Check
```http
GET /health
```

### Метрики агента
```http
GET /metrics
```

### Выполнение проверки безопасности
```http
POST /check
```

### Чат с AI-агентом
```http
POST /chat
Content-Type: application/json

{
  "message": "/status"
}
```

### WebSocket для real-time чата
```http
WebSocket /ws
```

## 💬 Команды AI-агента

- `/status` - Получить статус системы
- `/check` - Выполнить проверку безопасности
- `/start` - Запустить мониторинг
- `/stop` - Остановить мониторинг
- `/history` - Показать историю действий

## 🎨 UI Компоненты

### React UI

```tsx
import { AgentUI } from './src/components/AgentUI';

function App() {
  return (
    <AgentUI 
      agentUrl="http://localhost:8787"
      title="CloudBridge AI Security Agent"
    />
  );
}
```

### Использование AI SDK

```tsx
import { useChat } from 'ai/react';

function ChatComponent() {
  const { messages, input, handleInputChange, handleSubmit } = useChat({
    api: '/chat',
    onError: (error) => console.error('Chat error:', error)
  });

  return (
    <form onSubmit={handleSubmit}>
      <input value={input} onChange={handleInputChange} />
      <button type="submit">Отправить</button>
    </form>
  );
}
```

## 🔧 Разработка

### Структура проекта

```
cloudflare-agent/
├── src/
│   ├── index.ts              # Основной Worker
│   └── components/
│       └── AgentUI.tsx       # React UI компонент
├── test/
│   └── test-agent.js         # Тестовый скрипт
├── package.json              # Зависимости
├── tsconfig.json             # TypeScript конфигурация (общая)
├── tsconfig.workers.json     # TypeScript конфигурация для Workers
├── wrangler.toml             # Cloudflare Workers конфигурация
├── setup.sh                  # Скрипт автоматической настройки
├── deploy.sh                 # Скрипт развертывания
├── env.example               # Пример переменных окружения
└── README.md                 # Документация
```

### TypeScript Конфигурация

Проект использует две конфигурации TypeScript:

- `tsconfig.json` - основная конфигурация для React компонентов
- `tsconfig.workers.json` - конфигурация для Cloudflare Workers с типами

Для проверки типов используйте:
```bash
npm run type-check        # Общая проверка
npm run type-check:workers # С типами Workers
```

### Команды разработки

```bash
# Запуск в режиме разработки
npm run dev

# Проверка типов
npm run type-check
npm run type-check:workers  # С типами Cloudflare Workers

# Сборка
npm run build
npm run build:workers       # Сборка для Workers

# Тестирование
npm test
npm run test:agent          # Тестирование агента

# Линтинг
npm run lint
```

## 🔒 Безопасность

### API Ключи

- Все API ключи хранятся как секреты Cloudflare Workers
- Никогда не коммитьте API ключи в репозиторий
- Используйте переменные окружения для локальной разработки

### Аутентификация

- WebSocket соединения проверяются на валидность
- API endpoints защищены CORS политиками
- Все запросы логируются для аудита

## 📊 Мониторинг

### Метрики

Агент собирает следующие метрики:

- Количество проверок безопасности
- Обнаруженные аномалии
- Время отклика API
- Статус подключений
- Ошибки и исключения

### Логирование

```bash
# Просмотр логов в реальном времени
wrangler tail

# Просмотр логов конкретной среды
wrangler tail --env production
```

## 🐛 Устранение неполадок

### Частые проблемы

1. **Ошибка подключения к OpenAI API**
   - Проверьте правильность API ключа
   - Убедитесь в достаточности кредитов

2. **Ошибки WebSocket соединения**
   - Проверьте настройки CORS
   - Убедитесь в доступности порта

3. **Ошибки ML API**
   - Проверьте доступность ML сервиса
   - Убедитесь в правильности формата данных

### Отладка

```bash
# Запуск с подробными логами
DEBUG=* npm run dev

# Проверка конфигурации
wrangler whoami
wrangler config list
```

## 🤝 Вклад в проект

1. Форкните репозиторий
2. Создайте ветку для новой функции
3. Внесите изменения
4. Добавьте тесты
5. Создайте Pull Request

## 📄 Лицензия

MIT License - см. файл [LICENSE](LICENSE) для деталей.

## 🔗 Ссылки

- [Vercel AI SDK Documentation](https://ai-sdk.dev/docs/introduction)
- [Cloudflare Workers Documentation](https://developers.cloudflare.com/workers/)
- [OpenAI API Documentation](https://platform.openai.com/docs)
- [CloudBridge Relay Documentation](docs/)

## 📞 Поддержка

- Создайте Issue в GitHub для багов
- Используйте Discussions для вопросов
- Обратитесь к команде CloudBridge для коммерческой поддержки 