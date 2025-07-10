require('dotenv').config();
const fetch = require('node-fetch');
const { Telegraf } = require('telegraf');

// Отладочная информация
console.log('TELEGRAM_BOT_TOKEN:', process.env.TELEGRAM_BOT_TOKEN);
console.log('TELEGRAM_ADMIN_CHAT_ID:', process.env.TELEGRAM_ADMIN_CHAT_ID);
console.log('RELAY_METRICS_URL:', process.env.RELAY_METRICS_URL);

const bot = new Telegraf(process.env.TELEGRAM_BOT_TOKEN);
const ADMIN_CHAT_ID = process.env.TELEGRAM_ADMIN_CHAT_ID;
const METRICS_URL = process.env.RELAY_METRICS_URL || 'http://localhost:9090/metrics';

console.log('Бот инициализирован, запуск...');

// Мониторинг метрик и алерты
async function pollMetrics() {
  try {
    console.log('Проверка метрик...');
    const res = await fetch(METRICS_URL);
    const text = await res.text();
    
    // Пример: алерт при большом числе auth failures
    const match = text.match(/relay_auth_failures_total\{.*\} (\d+)/);
    if (match && parseInt(match[1]) > 10) {
      console.log('Обнаружено много ошибок аутентификации!');
      await bot.telegram.sendMessage(ADMIN_CHAT_ID, '⚠️ Много ошибок аутентификации!');
    }
  } catch (e) {
    console.error('Ошибка мониторинга метрик:', e);
  }
}

// Запуск периодического мониторинга
setInterval(pollMetrics, 10000); // каждые 10 секунд

// /start — приветствие и справка
bot.start((ctx) => {
  console.log('Получена команда /start');
  ctx.reply(
    'Привет! Я бот мониторинга CloudBridge Relay.\n' +
    'Доступные команды:\n' +
    '/status — фрагмент метрик\n' +
    '/test — тестовое сообщение\n' +
    '/block <ip> — заблокировать IP (заглушка)\n' +
    '/echoenv — показать переменные окружения\n' +
    '/alert <сообщение> — отправить алерт (админ)\n' +
    '/help — список команд'
  );
});

// /help — справка по командам
bot.command('help', (ctx) => {
  console.log('Получена команда /help');
  ctx.reply(
    'Доступные команды:\n' +
    '/status — фрагмент метрик\n' +
    '/test — тестовое сообщение\n' +
    '/block <ip> — заблокировать IP (заглушка)\n' +
    '/echoenv — показать переменные окружения\n' +
    '/alert <сообщение> — отправить алерт (админ)\n' +
    '/help — список команд'
  );
});

// /status — фрагмент метрик
bot.command('status', async (ctx) => {
  console.log('Получена команда /status');
  try {
    const res = await fetch(METRICS_URL);
    const text = await res.text();
    ctx.reply('Текущие метрики (фрагмент):\n' + text.slice(0, 1000));
  } catch (e) {
    console.error('Ошибка получения метрик:', e);
    ctx.reply('Ошибка получения метрик: ' + e.message);
  }
});

// /test — тестовое сообщение
bot.command('test', async (ctx) => {
  console.log('Получена команда /test');
  try {
    await ctx.reply('🧪 Тестовое сообщение: Cloudflare Agent Telegram-бот работает!');
    console.log('Тестовая команда /test выполнена.');
  } catch (e) {
    console.error('Ошибка при отправке тестового сообщения:', e);
  }
});

// /block <ip> — заглушка для блокировки IP
bot.command('block', (ctx) => {
  console.log('Получена команда /block');
  const ip = ctx.message.text.split(' ')[1];
  if (!ip) return ctx.reply('Используйте: /block <ip>');
  ctx.reply(`IP ${ip} заблокирован (заглушка)`);
});

// /echoenv — показать переменные окружения (только для админа)
bot.command('echoenv', async (ctx) => {
  console.log('Получена команда /echoenv');
  if (ctx.chat.id.toString() !== process.env.TELEGRAM_ADMIN_CHAT_ID) {
    return ctx.reply('⛔️ Нет доступа');
  }
  const msg = `TELEGRAM_BOT_TOKEN: ${process.env.TELEGRAM_BOT_TOKEN}\nTELEGRAM_ADMIN_CHAT_ID: ${process.env.TELEGRAM_ADMIN_CHAT_ID}\nRELAY_METRICS_URL: ${process.env.RELAY_METRICS_URL}`;
  await ctx.reply(msg);
  console.log('Выполнена команда /echoenv');
});

// /alert <сообщение> — отправить алерт (только для админа)
bot.command('alert', async (ctx) => {
  console.log('Получена команда /alert');
  if (ctx.chat.id.toString() !== process.env.TELEGRAM_ADMIN_CHAT_ID) {
    return ctx.reply('⛔️ Нет доступа');
  }
  const msg = ctx.message.text.replace('/alert', '').trim();
  if (!msg) return ctx.reply('Используйте: /alert <сообщение>');
  await ctx.reply('🚨 АЛЕРТ: ' + msg);
  console.log('Выполнена команда /alert');
});

// Обработка ошибок
bot.catch((err, ctx) => {
  console.error('Ошибка бота:', err);
  ctx.reply('Произошла ошибка: ' + err.message);
});

// Запуск бота
console.log('Запуск бота...');
bot.launch().then(() => {
  console.log('Бот успешно запущен!');
  // Отправляем тестовое сообщение при запуске
  bot.telegram.sendMessage(ADMIN_CHAT_ID, '🤖 Cloudflare Agent Telegram-бот запущен и готов к работе!');
}).catch((err) => {
  console.error('Ошибка запуска бота:', err);
});

// Graceful stop
process.once('SIGINT', () => {
  console.log('Получен SIGINT, останавливаю бота...');
  bot.stop('SIGINT');
});
process.once('SIGTERM', () => {
  console.log('Получен SIGTERM, останавливаю бота...');
  bot.stop('SIGTERM');
}); 