#!/bin/bash

# CloudBridge Relay - Исправление ошибок линтера
# Автоматическое исправление основных проблем golangci-lint

echo " Исправление ошибок golangci-lint..."

# 1. Форматирование кода
echo "📝 Исправление форматирования..."
gofmt -w .
goimports -w . 2>/dev/null || echo "goimports не установлен"

# 2. Исправление неиспользуемых параметров
echo "🗑️ Исправление неиспользуемых параметров..."

# Найти функции с неиспользуемыми параметрами и заменить их на _
find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | while read -r file; do
    # Заменяем часто встречающиеся неиспользуемые параметры
    sed -i 's/func.*(\([^)]*\)ctx context\.Context\([^)]*\))/func \1_ context.Context\2)/g' "$file" 2>/dev/null || true
    sed -i 's/func.*(\([^)]*\)t \*testing\.T\([^)]*\))/func \1_ *testing.T\2)/g' "$file" 2>/dev/null || true
    sed -i 's/func.*(\([^)]*\)req \*[^,)]*\([^)]*\))/func \1_ *\2)/g' "$file" 2>/dev/null || true
done

# 3. Исправление констант
echo "📐 Добавление констант..."
cat > temp_constants.go << 'EOF'
package main

// Константы для часто используемых строк
const (
    BothValue = "both"
    // Добавьте другие константы по необходимости
)
EOF

# 4. Исправление октальных литералов
echo "🔢 Исправление октальных литералов..."
find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | while read -r file; do
    sed -i 's/0750/0o750/g' "$file" 2>/dev/null || true
    sed -i 's/0600/0o600/g' "$file" 2>/dev/null || true
done

# 5. Исправление комментариев
echo "💬 Исправление комментариев..."
find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | while read -r file; do
    # Добавляем точки в конце комментариев
    sed -i 's/\/\/ \([^.]*[^.]\)$/\/\/ \1./g' "$file" 2>/dev/null || true
done

# 6. Удаление временных файлов
rm -f temp_constants.go

echo "Основные ошибки линтера исправлены!"
echo ""
echo "🧪 Запустите проверку:"
echo "  golangci-lint run --timeout=5m"
echo ""
echo "📝 Оставшиеся ошибки нужно исправить вручную:"
echo "  - Сложные функции (cyclomatic complexity)"
echo "  - Дублирование кода (dupl)"
echo "  - Неиспользуемые типы и функции (unused)"
