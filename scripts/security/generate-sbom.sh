#!/bin/bash

# CloudBridge SBOM Generation Script
# Автоматическая генерация Software Bill of Materials для Docker образов

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Функции для вывода
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

security_info() {
    echo -e "${PURPLE}[SECURITY]${NC} $1"
}

# Параметры
IMAGE="${1:-ghcr.io/2gc/cloudbridge-relay:latest}"
OUTPUT_DIR="${2:-./sbom}"

# Баннер
echo -e "${PURPLE}"
cat << "EOF"
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║         CloudBridge SBOM Generator                        ║
║         Software Bill of Materials Creation               ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
EOF
echo -e "${NC}"

# Проверка зависимостей
check_dependencies() {
    info "Проверка зависимостей..."

    # Проверяем syft
    if ! command -v syft &> /dev/null; then
        error "syft не найден. Установите syft:"
        echo "  brew install syft  # macOS"
        echo "  или"
        echo "  curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin"
        exit 1
    fi

    # Проверяем grype (опционально)
    if ! command -v grype &> /dev/null; then
        warning "grype не найден. Сканирование уязвимостей будет пропущено."
        SKIP_GRYPE=true
    else
        SKIP_GRYPE=false
    fi

    # Проверяем docker/podman
    if ! command -v docker &> /dev/null && ! command -v podman &> /dev/null; then
        warning "docker/podman не найден. Работа только с remote образами."
    fi

    # Версии
    SYFT_VERSION=$(syft version 2>&1 | head -1)
    info "Версия syft: $SYFT_VERSION"

    if [ "$SKIP_GRYPE" = false ]; then
        GRYPE_VERSION=$(grype version 2>&1 | head -1)
        info "Версия grype: $GRYPE_VERSION"
    fi

    success "Все зависимости проверены"
}

# Создание output директории
create_output_directory() {
    info "Создание output директории..."

    if [ -d "$OUTPUT_DIR" ]; then
        warning "Директория $OUTPUT_DIR уже существует"
    else
        mkdir -p "$OUTPUT_DIR"
        success "Директория $OUTPUT_DIR создана"
    fi

    chmod 755 "$OUTPUT_DIR"
}

# Генерация SBOM в SPDX формате
generate_spdx_sbom() {
    info "Генерация SBOM в SPDX формате для образа: $IMAGE"

    syft "$IMAGE" -o spdx-json > "$OUTPUT_DIR/sbom-spdx.json"

    if [ $? -eq 0 ]; then
        FILE_SIZE=$(du -h "$OUTPUT_DIR/sbom-spdx.json" | cut -f1)
        success "SPDX SBOM создан: $OUTPUT_DIR/sbom-spdx.json ($FILE_SIZE)"
    else
        error "Ошибка при генерации SPDX SBOM"
        exit 1
    fi
}

# Генерация SBOM в CycloneDX формате
generate_cyclonedx_sbom() {
    info "Генерация SBOM в CycloneDX формате для образа: $IMAGE"

    syft "$IMAGE" -o cyclonedx-json > "$OUTPUT_DIR/sbom-cyclonedx.json"

    if [ $? -eq 0 ]; then
        FILE_SIZE=$(du -h "$OUTPUT_DIR/sbom-cyclonedx.json" | cut -f1)
        success "CycloneDX SBOM создан: $OUTPUT_DIR/sbom-cyclonedx.json ($FILE_SIZE)"
    else
        error "Ошибка при генерации CycloneDX SBOM"
        exit 1
    fi
}

# Генерация human-readable SBOM
generate_text_sbom() {
    info "Генерация text SBOM для образа: $IMAGE"

    syft "$IMAGE" -o table > "$OUTPUT_DIR/sbom.txt"

    if [ $? -eq 0 ]; then
        FILE_SIZE=$(du -h "$OUTPUT_DIR/sbom.txt" | cut -f1)
        success "Text SBOM создан: $OUTPUT_DIR/sbom.txt ($FILE_SIZE)"
    else
        error "Ошибка при генерации text SBOM"
        exit 1
    fi
}

# Генерация JSON summary
generate_json_summary() {
    info "Генерация JSON summary для образа: $IMAGE"

    syft "$IMAGE" -o syft-json > "$OUTPUT_DIR/sbom-syft.json"

    if [ $? -eq 0 ]; then
        FILE_SIZE=$(du -h "$OUTPUT_DIR/sbom-syft.json" | cut -f1)
        success "Syft JSON создан: $OUTPUT_DIR/sbom-syft.json ($FILE_SIZE)"
    else
        error "Ошибка при генерации Syft JSON"
        exit 1
    fi
}

# Анализ SBOM и создание summary
analyze_sbom() {
    info "Анализ SBOM..."

    # Подсчет пакетов
    TOTAL_PACKAGES=$(jq '.artifacts | length' "$OUTPUT_DIR/sbom-syft.json")
    info "Всего пакетов: $TOTAL_PACKAGES"

    # Подсчет по типам
    info "Пакеты по типам:"
    jq -r '.artifacts | group_by(.type) | .[] | "\(.[0].type): \(length)"' "$OUTPUT_DIR/sbom-syft.json" | while read line; do
        echo "  $line"
    done

    # Топ-10 пакетов по размеру (если доступно)
    if jq -e '.artifacts[0].metadata.size' "$OUTPUT_DIR/sbom-syft.json" &> /dev/null; then
        info "Топ-10 пакетов по размеру:"
        jq -r '.artifacts | sort_by(.metadata.size // 0) | reverse | .[0:10] | .[] | "\(.name) \(.version): \(.metadata.size // 0) bytes"' "$OUTPUT_DIR/sbom-syft.json" | while read line; do
            echo "  $line"
        done
    fi

    # Создание summary файла
    cat > "$OUTPUT_DIR/sbom-summary.txt" << EOF
# SBOM Summary for $IMAGE

**Generated**: $(date)
**Total Packages**: $TOTAL_PACKAGES

## Packages by Type
$(jq -r '.artifacts | group_by(.type) | .[] | "\(.[0].type): \(length)"' "$OUTPUT_DIR/sbom-syft.json")

## File Formats Generated
- SPDX JSON: sbom-spdx.json
- CycloneDX JSON: sbom-cyclonedx.json
- Text Table: sbom.txt
- Syft JSON: sbom-syft.json

EOF

    success "Summary создан: $OUTPUT_DIR/sbom-summary.txt"
}

# Сканирование уязвимостей с grype
scan_vulnerabilities() {
    if [ "$SKIP_GRYPE" = true ]; then
        warning "Пропуск сканирования уязвимостей (grype не найден)"
        return
    fi

    info "Сканирование уязвимостей с помощью grype..."

    # Сканирование SBOM
    grype sbom:"$OUTPUT_DIR/sbom-spdx.json" -o json > "$OUTPUT_DIR/vulnerabilities.json" 2>/dev/null || true

    if [ -f "$OUTPUT_DIR/vulnerabilities.json" ]; then
        FILE_SIZE=$(du -h "$OUTPUT_DIR/vulnerabilities.json" | cut -f1)
        success "Отчет об уязвимостях создан: $OUTPUT_DIR/vulnerabilities.json ($FILE_SIZE)"

        # Подсчет уязвимостей по severity
        info "Уязвимости по severity:"
        jq -r '.matches | group_by(.vulnerability.severity) | .[] | "\(.[0].vulnerability.severity): \(length)"' "$OUTPUT_DIR/vulnerabilities.json" 2>/dev/null | while read line; do
            echo "  $line"
        done || true

        # Генерация human-readable отчета
        grype sbom:"$OUTPUT_DIR/sbom-spdx.json" -o table > "$OUTPUT_DIR/vulnerabilities.txt" 2>/dev/null || true
        if [ -f "$OUTPUT_DIR/vulnerabilities.txt" ]; then
            success "Text отчет об уязвимостях: $OUTPUT_DIR/vulnerabilities.txt"
        fi
    else
        warning "Не удалось создать отчет об уязвимостях"
    fi
}

# Генерация metadata файла
generate_metadata() {
    info "Генерация metadata файла..."

    cat > "$OUTPUT_DIR/metadata.json" << EOF
{
  "image": "$IMAGE",
  "generated_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "generator": {
    "name": "syft",
    "version": "$(syft version 2>&1 | head -1 | awk '{print $2}')"
  },
  "formats": [
    "spdx-json",
    "cyclonedx-json",
    "syft-json",
    "text"
  ],
  "vulnerability_scan": $([ "$SKIP_GRYPE" = false ] && echo "true" || echo "false"),
  "total_packages": $(jq '.artifacts | length' "$OUTPUT_DIR/sbom-syft.json")
}
EOF

    success "Metadata создан: $OUTPUT_DIR/metadata.json"
}

# Создание README для SBOM
create_sbom_readme() {
    info "Создание README для SBOM..."

    cat > "$OUTPUT_DIR/README.md" << 'EOF'
# Software Bill of Materials (SBOM)

## 📋 О проекте

Этот SBOM содержит полный список компонентов и зависимостей Docker образа CloudBridge Relay.

## 📦 Файлы

### SBOM Форматы
- **sbom-spdx.json** - SPDX 2.3 формат (industry standard)
- **sbom-cyclonedx.json** - CycloneDX 1.4 формат (OWASP standard)
- **sbom-syft.json** - Syft native формат (полные метаданные)
- **sbom.txt** - Human-readable таблица

### Отчеты
- **sbom-summary.txt** - Краткая сводка по пакетам
- **vulnerabilities.json** - JSON отчет об уязвимостях
- **vulnerabilities.txt** - Text отчет об уязвимостях
- **metadata.json** - Метаданные генерации SBOM

## 🔍 Использование

### Просмотр пакетов
```bash
# Text формат
cat sbom.txt

# JSON с jq
jq '.artifacts[] | {name, version, type}' sbom-syft.json
```

### Анализ уязвимостей
```bash
# Сканирование SBOM
grype sbom:sbom-spdx.json

# Фильтр по severity
grype sbom:sbom-spdx.json --fail-on critical
```

### Сравнение версий
```bash
# Сравнить два SBOM
diff sbom-v1.txt sbom-v2.txt
```

## 🔐 Безопасность

### Critical Vulnerabilities
Если найдены критические уязвимости:
1. Проверьте `vulnerabilities.json`
2. Обновите зависимости
3. Пересоберите образ
4. Перегенерируйте SBOM

### Compliance
SBOM соответствует следующим стандартам:
- SPDX 2.3 (ISO/IEC 5962:2021)
- CycloneDX 1.4
- NTIA Minimum Elements

## 📊 Статистика

Смотрите `sbom-summary.txt` для:
- Общего количества пакетов
- Распределения по типам
- Топ пакетов по размеру

## 🔄 Обновление

Для обновления SBOM:
```bash
./scripts/security/generate-sbom.sh <image> <output-dir>
```

## 📚 Ссылки

- [SPDX Specification](https://spdx.dev/specifications/)
- [CycloneDX Specification](https://cyclonedx.org/specification/overview/)
- [Syft Documentation](https://github.com/anchore/syft)
- [Grype Documentation](https://github.com/anchore/grype)
EOF

    success "README создан: $OUTPUT_DIR/README.md"
}

# Вывод итоговой информации
print_summary() {
    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                           ║${NC}"
    echo -e "${GREEN}║              ✓ SBOM Generation Completed                  ║${NC}"
    echo -e "${GREEN}║                                                           ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
    echo ""

    info "📦 Образ: $IMAGE"
    info "📁 Output: $OUTPUT_DIR"
    echo ""

    info "📋 Созданные файлы:"
    ls -lh "$OUTPUT_DIR" | tail -n +2 | while read line; do
        echo "  $line"
    done
    echo ""

    if [ -f "$OUTPUT_DIR/vulnerabilities.json" ]; then
        security_info "🔍 Сканирование уязвимостей:"
        TOTAL_VULNS=$(jq '.matches | length' "$OUTPUT_DIR/vulnerabilities.json" 2>/dev/null || echo "0")
        if [ "$TOTAL_VULNS" -gt 0 ]; then
            warning "Найдено уязвимостей: $TOTAL_VULNS"
            echo "  Смотрите: $OUTPUT_DIR/vulnerabilities.txt"
        else
            success "Уязвимостей не найдено"
        fi
        echo ""
    fi

    info "📊 Статистика:"
    cat "$OUTPUT_DIR/sbom-summary.txt"
    echo ""

    info "📚 Следующие шаги:"
    echo "  1. Просмотрите $OUTPUT_DIR/sbom.txt"
    echo "  2. Проверьте $OUTPUT_DIR/vulnerabilities.txt"
    echo "  3. Прикрепите SBOM к образу с помощью cosign"
    echo "  4. Храните SBOM для аудита и compliance"
    echo ""

    success "✓ SBOM generation завершен успешно!"
}

# Основной процесс
main() {
    check_dependencies
    create_output_directory
    generate_spdx_sbom
    generate_cyclonedx_sbom
    generate_text_sbom
    generate_json_summary
    analyze_sbom
    scan_vulnerabilities
    generate_metadata
    create_sbom_readme
    print_summary
}

# Запуск
main
