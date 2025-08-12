# Настройка Cloudflare AI для AI Monitoring Agent

## 🚀 Преимущества Cloudflare AI

- **Бесплатно**: Нет платы за API вызовы
- **Быстро**: Модели работают на edge-серверах Cloudflare
- **Просто**: Встроенная интеграция в Workers
- **Надежно**: Высокая доступность и производительность

## 📋 Доступные модели

### Рекомендуемые модели:
- `@cf/meta/llama-2-7b-chat-int8` - Llama 2 7B (быстрая)
- `@cf/mistral/mistral-7b-instruct-v0.2` - Mistral 7B (качественная)
- `@cf/meta/llama-3.1-8b-instruct` - Llama 3.1 8B (новая)

### Полный список:
```bash
wrangler ai list
```

## ⚙️ Настройка

### 1. Получение Cloudflare API токена

1. **Войдите в [Cloudflare Dashboard](https://dash.cloudflare.com/)**
2. **Перейдите в "My Profile" → "API Tokens"**
3. **Создайте новый токен:**
   - **Permissions**: `Cloudflare Workers AI` → `Edit`
   - **Account Resources**: `All accounts`
   - **Zone Resources**: `All zones`
4. **Скопируйте токен** (начинается с `CF_`)

### 2. Обновленный wrangler.toml
```toml
# Cloudflare AI binding
[[ai]]
binding = "AI"

[vars]
AI_MODEL = "@cf/meta/llama-2-7b-chat-int8"
```

### 3. Переменные окружения
```bash
# Обязательные
ML_API_KEY=your-ml-api-key
SIEM_API_KEY=your-siem-api-key

# AI токены (один из двух)
CLOUDFLARE_API_TOKEN=CF_your-cloudflare-token  # Для Cloudflare AI
OPENAI_API_KEY=sk-your-openai-key              # Для OpenAI (опционально)
```

### 4. Деплой
```bash
# Добавьте секреты
wrangler secret put CLOUDFLARE_API_TOKEN --env production
wrangler secret put ML_API_KEY --env production
wrangler secret put SIEM_API_KEY --env production

# Запустите деплой
./start_production.sh
```

## 🔄 Миграция с OpenAI

### Если у вас уже есть OpenAI:
1. Оставьте `OPENAI_API_KEY` - код будет работать с обеими системами
2. Добавьте `CLOUDFLARE_API_TOKEN` для Cloudflare AI
3. Добавьте AI binding в wrangler.toml
4. Установите `AI_MODEL` в переменных

### Если переходите полностью на Cloudflare AI:
1. Удалите `OPENAI_API_KEY` из секретов
2. Добавьте `CLOUDFLARE_API_TOKEN`
3. Обновите wrangler.toml с AI binding
4. Перезапустите агента

## 📊 Сравнение производительности

| Метрика | Cloudflare AI | OpenAI |
|---------|---------------|--------|
| Стоимость | Бесплатно | $0.01-0.03 за 1K токенов |
| Скорость | ~100-500ms | ~500-2000ms |
| Доступность | 99.9% | 99.5% |
| Локализация | Edge-серверы | Центральные дата-центры |
| Токен | CLOUDFLARE_API_TOKEN | OPENAI_API_KEY |

## 🛠️ Отладка

### Проверка AI binding:
```bash
curl -X POST http://localhost:8787/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Привет, как дела?"}'
```

### Проверка токенов:
```bash
# Проверить наличие токенов
wrangler secret list --env production

# Добавить Cloudflare токен
wrangler secret put CLOUDFLARE_API_TOKEN --env production
```

### Логи AI:
- Метрики производительности в `/performance`
- Ошибки в консоли wrangler

## 🔧 Кастомизация

### Изменение модели:
```toml
[vars]
AI_MODEL = "@cf/mistral/mistral-7b-instruct-v0.2"
```

### Настройка промптов:
Отредактируйте `CloudflareAIService.generateResponse()` в `src/index.ts`

## 📈 Мониторинг

### Метрики AI:
- `ai_generation` - время генерации ответа
- `ai_generation_error` - количество ошибок

### Endpoints:
- `/performance` - метрики производительности
- `/health` - статус системы
- `/chat` - взаимодействие с AI

## ⚠️ Важные замечания

1. **CLOUDFLARE_API_TOKEN обязателен** для работы с Cloudflare AI
2. **Токен должен иметь права** `Cloudflare Workers AI:Edit`
3. **AI binding автоматически** подключается при наличии токена
4. **Можно использовать оба токена** одновременно для fallback 