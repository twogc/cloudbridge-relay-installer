#!/bin/bash

# CloudBridge Relay - Создание релиза
# Локальный скрипт для создания релизов без GitHub Actions

set -e

VERSION=${1:-$(date +v%Y.%m.%d)}
DIST_DIR="dist"

echo "🚀 Создание релиза CloudBridge Relay $VERSION"
echo "============================================="

# Очистка предыдущих сборок
rm -rf $DIST_DIR
mkdir -p $DIST_DIR

# Проверка зависимостей
echo "📦 Проверка зависимостей..."
if ! command -v go &> /dev/null; then
    echo "❌ Go не установлен"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo "❌ Docker не установлен"
    exit 1
fi

echo "Зависимости проверены"

# Запуск тестов
echo "🧪 Запуск тестов..."
make test || {
    echo "❌ Тесты не прошли"
    exit 1
}
echo "Тесты прошли успешно"

# Сборка для разных платформ
echo "🔨 Сборка бинарных файлов..."

platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r os arch <<< "$platform"
    output_name="cloudbridge-relay-$os-$arch"
    
    if [ "$os" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo "  📦 Сборка для $os/$arch..."
    GOOS=$os GOARCH=$arch go build -ldflags "-X main.version=$VERSION" -o "$DIST_DIR/$output_name" ./cmd/relay
    
    if [ $? -ne 0 ]; then
        echo "❌ Ошибка сборки для $os/$arch"
        exit 1
    fi
done

echo "Бинарные файлы собраны"

# Создание архивов
echo "📦 Создание архивов..."
cd $DIST_DIR

for file in cloudbridge-relay-*; do
    if [[ $file == *.exe ]]; then
        # Windows - ZIP архив
        zip "${file%.exe}.zip" "$file"
        echo "  Создан ${file%.exe}.zip"
    else
        # Unix - TAR.GZ архив
        tar -czf "${file}.tar.gz" "$file"
        echo "  Создан ${file}.tar.gz"
    fi
done

cd ..

# Сборка Docker образа
echo "🐳 Сборка Docker образа..."
docker build -t "twogc/cloudbridge-relay:$VERSION" .
docker tag "twogc/cloudbridge-relay:$VERSION" "twogc/cloudbridge-relay:latest"
echo "Docker образ собран"

# Создание checksums
echo "🔐 Создание checksums..."
cd $DIST_DIR
sha256sum *.tar.gz *.zip > checksums.txt
echo "Checksums созданы"
cd ..

# Генерация changelog
echo "📝 Генерация changelog..."
git log --pretty=format:"- %s (%h)" $(git describe --tags --abbrev=0 2>/dev/null || echo "HEAD~10")..HEAD > $DIST_DIR/CHANGELOG.md
echo "Changelog сгенерирован"

# Показ результатов
echo ""
echo "🎉 Релиз $VERSION готов!"
echo "========================"
echo ""
echo "📁 Файлы релиза:"
ls -la $DIST_DIR/
echo ""
echo "🐳 Docker образы:"
docker images twogc/cloudbridge-relay

echo ""
echo "📤 Для публикации релиза:"
echo "  1. Создайте tag: git tag $VERSION && git push origin $VERSION"
echo "  2. Загрузите Docker образ: docker push twogc/cloudbridge-relay:$VERSION"
echo "  3. Создайте GitHub Release и приложите файлы из dist/"

echo ""
echo "Релиз $VERSION успешно создан!"
