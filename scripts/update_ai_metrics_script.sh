#!/bin/bash

# Скрипт для обновления ai_metrics_collector.sh

set -e

echo "🔄 Обновление ai_metrics_collector.sh..."

# Проверяем, что файл существует
if [ ! -f "scripts/ai_metrics_collector.sh" ]; then
    echo "❌ Файл scripts/ai_metrics_collector.sh не найден"
    exit 1
fi

# Копируем файл в /usr/local/bin/
echo "Копирование файла в /usr/local/bin/..."
sudo cp scripts/ai_metrics_collector.sh /usr/local/bin/ai_metrics_collector.sh

# Устанавливаем правильные права
echo "🔐 Установка прав доступа..."
sudo chown root:root /usr/local/bin/ai_metrics_collector.sh
sudo chmod +x /usr/local/bin/ai_metrics_collector.sh

# Проверяем, что файл обновился
echo "Проверка обновления..."
if [ -f "/usr/local/bin/ai_metrics_collector.sh" ]; then
    echo "Файл успешно обновлен в /usr/local/bin/"
    ls -la /usr/local/bin/ai_metrics_collector.sh
else
    echo "❌ Ошибка обновления файла"
    exit 1
fi

# Тестируем обновленный скрипт
echo "🧪 Тестирование обновленного скрипта..."
echo "   AI Status: $(/usr/local/bin/ai_metrics_collector.sh status)"
echo "   AI Health: $(/usr/local/bin/ai_metrics_collector.sh health)"
echo "   Anomaly Detections: $(/usr/local/bin/ai_metrics_collector.sh anomaly_detections)"

echo ""
echo "🎉 Скрипт ai_metrics_collector.sh успешно обновлен!"
echo ""
echo "Информация:"
echo "   Исходный файл: scripts/ai_metrics_collector.sh"
echo "   Целевой файл: /usr/local/bin/ai_metrics_collector.sh"
echo "   Права: root:root, исполняемый"
echo ""
echo "Для редактирования используйте файл в scripts/ai_metrics_collector.sh"
echo "   Затем запустите этот скрипт для обновления"
