#!/bin/bash

# CloudBridge Cosign Keys Setup Script
# Автоматическая генерация и настройка ключей для cosign

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

# Конфигурация
NAMESPACE="cloudbridge"
KEYS_DIR="./cosign-keys"
SECRET_NAME="cosign-keys"

# Баннер
echo -e "${PURPLE}"
cat << "EOF"
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║         CloudBridge Cosign Keys Setup                     ║
║         Secure Image Signing Infrastructure               ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
EOF
echo -e "${NC}"

# Проверка зависимостей
check_dependencies() {
    info "Проверка зависимостей..."

    # Проверяем cosign
    if ! command -v cosign &> /dev/null; then
        error "cosign не найден. Установите cosign:"
        echo "  brew install cosign  # macOS"
        echo "  или"
        echo "  go install github.com/sigstore/cosign/v2/cmd/cosign@latest"
        exit 1
    fi

    # Проверяем kubectl
    if ! command -v kubectl &> /dev/null; then
        error "kubectl не найден. Установите kubectl и настройте доступ к кластеру."
        exit 1
    fi

    # Проверяем доступ к кластеру
    if ! kubectl cluster-info &> /dev/null; then
        warning "Нет доступа к Kubernetes кластеру. Ключи будут созданы локально."
        SKIP_K8S=true
    else
        SKIP_K8S=false
        success "Доступ к Kubernetes кластеру подтвержден"
    fi

    # Версия cosign
    COSIGN_VERSION=$(cosign version 2>&1 | head -1)
    info "Версия cosign: $COSIGN_VERSION"

    success "Все зависимости проверены"
}

# Создание директории для ключей
create_keys_directory() {
    info "Создание директории для ключей..."

    if [ -d "$KEYS_DIR" ]; then
        warning "Директория $KEYS_DIR уже существует"
        read -p "Перезаписать существующие ключи? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            error "Отменено пользователем"
            exit 1
        fi
        rm -rf "$KEYS_DIR"
    fi

    mkdir -p "$KEYS_DIR"
    chmod 700 "$KEYS_DIR"
    success "Директория $KEYS_DIR создана"
}

# Генерация ключей cosign
generate_keys() {
    info "Генерация ключей cosign..."

    cd "$KEYS_DIR"

    # Генерация ключей (пустой пароль для автоматизации, в production используйте пароль!)
    security_info "Генерируем ключи без пароля для CI/CD автоматизации"
    warning "В production среде используйте пароль и храните его в Secret Manager!"

    COSIGN_PASSWORD="" cosign generate-key-pair

    if [ -f "cosign.key" ] && [ -f "cosign.pub" ]; then
        success "Ключи cosign успешно созданы"
        info "Private key: $(pwd)/cosign.key"
        info "Public key: $(pwd)/cosign.pub"
    else
        error "Ошибка при генерации ключей"
        exit 1
    fi

    cd ..
}

# Создание README для ключей
create_keys_readme() {
    info "Создание README для ключей..."

    cat > "$KEYS_DIR/README.md" << 'EOF'
# Cosign Keys

## 🔐 Безопасность

**ВАЖНО**: Эти ключи предназначены для подписи Docker образов.

### Private Key (cosign.key)
- ❌ **НЕ КОММИТИТЬ** в Git
- ✅ Хранить в Kubernetes Secret
- ✅ Хранить в Secret Manager (Vault/AWS/GCP)
- ✅ Использовать для подписи образов в CI/CD

### Public Key (cosign.pub)
- ✅ Можно коммитить в Git
- ✅ Использовать для верификации образов
- ✅ Публиковать для transparency

## 📦 Использование

### Подпись образа
```bash
cosign sign --key cosign.key <image>
```

### Верификация подписи
```bash
cosign verify --key cosign.pub <image>
```

## 🔄 Keyless Signing (рекомендуется для production)

Вместо хранения ключей используйте keyless signing через OIDC:

```bash
# Подпись (автоматически использует GitHub OIDC)
cosign sign --yes <image>

# Верификация
cosign verify \
  --certificate-identity-regexp="^https://github.com/YOUR_ORG/YOUR_REPO/" \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
  <image>
```

## 📝 Rotation Policy

Рекомендуется ротация ключей каждые:
- 90 дней для production
- 180 дней для staging
- 365 дней для development

## 🗑️ Revocation

Если ключи скомпрометированы:
1. Удалить из Kubernetes Secrets
2. Удалить из Secret Manager
3. Сгенерировать новые ключи
4. Перевыпустить все подписанные образы
5. Обновить публичный ключ
EOF

    success "README создан: $KEYS_DIR/README.md"
}

# Создание .gitignore для ключей
create_gitignore() {
    info "Создание .gitignore для защиты приватного ключа..."

    cat > "$KEYS_DIR/.gitignore" << 'EOF'
# Cosign private key - НЕ КОММИТИТЬ В GIT!
cosign.key

# Опционально: публичный ключ можно коммитить
# cosign.pub
EOF

    success ".gitignore создан: $KEYS_DIR/.gitignore"
}

# Создание Kubernetes Secret
create_kubernetes_secret() {
    if [ "$SKIP_K8S" = true ]; then
        warning "Пропуск создания Kubernetes Secret (нет доступа к кластеру)"
        return
    fi

    info "Создание Kubernetes Secret для cosign ключей..."

    # Создание namespace если не существует
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        info "Создание namespace $NAMESPACE..."
        kubectl create namespace "$NAMESPACE"
        success "Namespace $NAMESPACE создан"
    fi

    # Удаление существующего Secret если есть
    if kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" &> /dev/null; then
        warning "Secret $SECRET_NAME уже существует, удаляем..."
        kubectl delete secret "$SECRET_NAME" -n "$NAMESPACE"
    fi

    # Создание Secret
    kubectl create secret generic "$SECRET_NAME" \
        --from-file=cosign.key="$KEYS_DIR/cosign.key" \
        --from-file=cosign.pub="$KEYS_DIR/cosign.pub" \
        -n "$NAMESPACE"

    if [ $? -eq 0 ]; then
        success "Kubernetes Secret $SECRET_NAME создан в namespace $NAMESPACE"
    else
        error "Ошибка при создании Kubernetes Secret"
        exit 1
    fi

    # Проверка Secret
    info "Проверка содержимого Secret..."
    kubectl describe secret "$SECRET_NAME" -n "$NAMESPACE"
}

# Тестирование ключей
test_keys() {
    info "Тестирование ключей cosign..."

    # Создаем тестовый файл
    echo "Test data for cosign" > "$KEYS_DIR/test.txt"

    # Подписываем файл
    info "Подпись тестового файла..."
    COSIGN_PASSWORD="" cosign sign-blob --key "$KEYS_DIR/cosign.key" "$KEYS_DIR/test.txt" > "$KEYS_DIR/test.txt.sig"

    if [ $? -eq 0 ]; then
        success "Тестовый файл подписан"
    else
        error "Ошибка при подписи тестового файла"
        exit 1
    fi

    # Верификация подписи
    info "Верификация подписи тестового файла..."
    cosign verify-blob --key "$KEYS_DIR/cosign.pub" --signature "$KEYS_DIR/test.txt.sig" "$KEYS_DIR/test.txt"

    if [ $? -eq 0 ]; then
        success "Подпись верифицирована успешно!"
    else
        error "Ошибка верификации подписи"
        exit 1
    fi

    # Очистка тестовых файлов
    rm -f "$KEYS_DIR/test.txt" "$KEYS_DIR/test.txt.sig"
}

# Создание инструкций
create_usage_instructions() {
    info "Создание инструкций по использованию..."

    cat > "$KEYS_DIR/USAGE.md" << 'EOF'
# Cosign Usage Instructions

## 🚀 Quick Start

### 1. Подпись Docker образа

```bash
# Экспорт ключа (если необходимо)
export COSIGN_PASSWORD=""

# Подпись образа
cosign sign --key cosign.key ghcr.io/2gc/cloudbridge-relay:latest

# Подпись с keyless (рекомендуется)
cosign sign --yes ghcr.io/2gc/cloudbridge-relay:latest
```

### 2. Верификация подписи

```bash
# С ключом
cosign verify --key cosign.pub ghcr.io/2gc/cloudbridge-relay:latest

# Keyless
cosign verify \
  --certificate-identity-regexp="^https://github.com/2gc/cloudbridge-relay-installer/" \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
  ghcr.io/2gc/cloudbridge-relay:latest
```

### 3. Генерация SBOM

```bash
# Генерация SBOM с syft
syft ghcr.io/2gc/cloudbridge-relay:latest -o spdx-json > sbom.spdx.json

# Прикрепление SBOM к образу
cosign attest --yes --predicate sbom.spdx.json --type spdxjson \
  ghcr.io/2gc/cloudbridge-relay:latest
```

### 4. Верификация SBOM attestation

```bash
cosign verify-attestation --type spdxjson --key cosign.pub \
  ghcr.io/2gc/cloudbridge-relay:latest
```

## 🔐 GitHub Actions Integration

### Keyless Signing в CI/CD

```yaml
name: Sign Images
on:
  push:
    branches: [main]

jobs:
  sign:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      packages: write
      contents: read

    steps:
      - name: Install cosign
        uses: sigstore/cosign-installer@v3

      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and Push
        run: |
          docker build -t ghcr.io/2gc/cloudbridge-relay:${{ github.sha }} .
          docker push ghcr.io/2gc/cloudbridge-relay:${{ github.sha }}

      - name: Sign image
        run: cosign sign --yes ghcr.io/2gc/cloudbridge-relay:${{ github.sha }}
```

## 📊 Kubernetes Policy Enforcement

### Kyverno Policy

```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: verify-cloudbridge-images
spec:
  validationFailureAction: enforce
  rules:
    - name: verify-signature
      match:
        any:
        - resources:
            kinds: [Pod]
      verifyImages:
      - imageReferences:
        - "ghcr.io/2gc/cloudbridge-relay*"
        attestors:
        - entries:
          - keyless:
              subject: "https://github.com/2gc/*"
              issuer: "https://token.actions.githubusercontent.com"
              rekor:
                url: https://rekor.sigstore.dev
```

## 🔄 Best Practices

1. **Используйте keyless signing** для production
2. **Храните private keys** в Secret Manager
3. **Ротируйте ключи** каждые 90 дней
4. **Используйте Rekor** для transparency log
5. **Применяйте policy enforcement** в Kubernetes
6. **Генерируйте SBOM** для каждого образа
7. **Верифицируйте подписи** перед deployment

## 🆘 Troubleshooting

### Ошибка: "Error: signing [ghcr.io/...]: PUT request failed"

**Решение**: Проверьте права доступа к registry:
```bash
docker login ghcr.io
```

### Ошибка: "Error: failed to verify signature"

**Решение**: Проверьте используемый публичный ключ:
```bash
cosign verify --key cosign.pub <image>
```

### Ошибка: "context deadline exceeded"

**Решение**: Увеличьте timeout:
```bash
cosign sign --timeout=5m --key cosign.key <image>
```

## 📚 Полезные ссылки

- [Cosign Documentation](https://docs.sigstore.dev/cosign/overview/)
- [Keyless Signing](https://docs.sigstore.dev/cosign/keyless/)
- [Rekor Transparency Log](https://docs.sigstore.dev/rekor/overview/)
- [Kyverno Image Verification](https://kyverno.io/docs/writing-policies/verify-images/)
EOF

    success "Инструкции созданы: $KEYS_DIR/USAGE.md"
}

# Вывод итоговой информации
print_summary() {
    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                           ║${NC}"
    echo -e "${GREEN}║              ✓ Setup Completed Successfully               ║${NC}"
    echo -e "${GREEN}║                                                           ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
    echo ""

    info "Созданные файлы:"
    echo "  📁 $KEYS_DIR/"
    echo "    🔐 cosign.key (private key - НЕ КОММИТИТЬ!)"
    echo "    🔓 cosign.pub (public key)"
    echo "    📝 README.md"
    echo "    📝 USAGE.md"
    echo "    🚫 .gitignore"
    echo ""

    if [ "$SKIP_K8S" = false ]; then
        info "Kubernetes Secret:"
        echo "  📦 Namespace: $NAMESPACE"
        echo "  🔑 Secret: $SECRET_NAME"
        echo ""
        echo "  Проверить Secret:"
        echo "  kubectl get secret $SECRET_NAME -n $NAMESPACE"
        echo ""
    fi

    security_info "⚠️  ВАЖНО: Безопасность ключей"
    echo "  1. НЕ коммитьте cosign.key в Git"
    echo "  2. Храните private key в Secret Manager"
    echo "  3. Используйте keyless signing для production"
    echo "  4. Ротируйте ключи каждые 90 дней"
    echo ""

    info "📚 Следующие шаги:"
    echo "  1. Прочитайте $KEYS_DIR/USAGE.md"
    echo "  2. Настройте GitHub Actions для keyless signing"
    echo "  3. Настройте Kyverno policy для image verification"
    echo "  4. Генерируйте SBOM для образов с помощью syft"
    echo ""

    success "✓ Cosign keys setup завершен успешно!"
}

# Основной процесс
main() {
    check_dependencies
    create_keys_directory
    generate_keys
    create_keys_readme
    create_gitignore
    test_keys
    create_usage_instructions
    create_kubernetes_secret
    print_summary
}

# Запуск
main
