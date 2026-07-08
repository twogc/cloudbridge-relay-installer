# Supply Chain Security Implementation Plan

## Цель
Реализовать полный Supply Chain Security стек с подписью Docker образов и генерацией SBOM для CloudBridge Relay.

## Scope

### Этап 6.1: Подпись образов (cosign)
**Статус**:  НЕ РЕАЛИЗОВАНО (0%)
**Приоритет**: КРИТИЧНО
**Время**: 1-1.5 недели

### Этап 6.2: SBOM генерация (syft)
**Статус**:  НЕ РЕАЛИЗОВАНО (0%)
**Приоритет**: КРИТИЧНО
**Время**: 1-1.5 недели

**Общее время**: 2-3 недели

---

## Этап 6.1: Подпись образов (Неделя 1-1.5)

### 6.1.1 Подготовка окружения (День 1)

**Действия**:
- [ ] Клонировать `sigstore/cosign` в `supply-chain/signing/cosign/`
- [ ] Установить cosign локально для тестирования
- [ ] Создать документацию по cosign в `docs/supply-chain/cosign-setup.md`
- [ ] Изучить best practices для keyless signing

**Команды**:
```bash
# Клонирование cosign
cd supply-chain/signing/
git clone https://github.com/sigstore/cosign.git cosign/
cd cosign/
git checkout v2.2.2  # Latest stable version

# Установка cosign
go install github.com/sigstore/cosign/v2/cmd/cosign@latest

# Проверка установки
cosign version
```

**Ожидаемый результат**: cosign установлен и готов к использованию

---

### 6.1.2 Настройка подписи образов (День 2-3)

**Действия**:
- [ ] Создать скрипт для генерации ключей (`scripts/setup/setup-cosign-keys.sh`)
- [ ] Настроить keyless signing через OIDC (GitHub Actions)
- [ ] Создать Kubernetes Secret для хранения ключей
- [ ] Протестировать подпись локального образа

**Скрипт генерации ключей**:
```bash
#!/bin/bash
# scripts/setup/setup-cosign-keys.sh

# Генерация ключей для cosign
cosign generate-key-pair

# Создание Kubernetes Secret
kubectl create secret generic cosign-keys \
  --from-file=cosign.key=cosign.key \
  --from-file=cosign.pub=cosign.pub \
  -n cloudbridge

# Верификация
kubectl get secret cosign-keys -n cloudbridge
```

**Ожидаемый результат**: Ключи созданы и сохранены в Kubernetes Secret

---

### 6.1.3 Интеграция с CI/CD (День 4-5)

**Действия**:
- [ ] Создать GitHub Actions workflow для подписи образов
- [ ] Интегрировать cosign в существующий build pipeline
- [ ] Добавить автоматическую подпись при push в main
- [ ] Настроить верификацию подписей в CI

**GitHub Actions Workflow**:
```yaml
# .github/workflows/sign-images.yaml
name: Sign Docker Images

on:
  push:
    branches: [main]
    paths:
      - 'cmd/**'
      - 'internal/**'
      - 'Dockerfile*'

jobs:
  build-and-sign:
    runs-on: ubuntu-latest
    permissions:
      id-token: write  # For OIDC
      packages: write  # For GitHub Container Registry
      contents: read

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Install cosign
        uses: sigstore/cosign-installer@v3
        with:
          cosign-release: 'v2.2.2'

      - name: Login to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build Docker image
        run: |
          docker build -t ghcr.io/${{ github.repository }}/cloudbridge-relay:${{ github.sha }} .
          docker push ghcr.io/${{ github.repository }}/cloudbridge-relay:${{ github.sha }}

      - name: Sign Docker image (keyless)
        run: |
          cosign sign --yes ghcr.io/${{ github.repository }}/cloudbridge-relay:${{ github.sha }}

      - name: Verify signature
        run: |
          cosign verify \
            --certificate-identity-regexp="^https://github.com/${{ github.repository }}/" \
            --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
            ghcr.io/${{ github.repository }}/cloudbridge-relay:${{ github.sha }}
```

**Ожидаемый результат**: Автоматическая подпись образов в CI/CD

---

### 6.1.4 Верификация подписей в Kubernetes (День 6-7)

**Действия**:
- [ ] Создать Admission Controller для верификации подписей
- [ ] Настроить policy enforcement через Kyverno или OPA
- [ ] Протестировать запрет непроверенных образов
- [ ] Создать документацию по верификации

**Kyverno Policy**:
```yaml
# k8s/policies/verify-image-signatures.yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: verify-image-signatures
spec:
  validationFailureAction: enforce
  background: false
  rules:
    - name: verify-cloudbridge-images
      match:
        any:
        - resources:
            kinds:
              - Pod
      verifyImages:
      - imageReferences:
        - "ghcr.io/*/cloudbridge-relay*"
        attestors:
        - entries:
          - keyless:
              subject: "https://github.com/*"
              issuer: "https://token.actions.githubusercontent.com"
              rekor:
                url: https://rekor.sigstore.dev
```

**Ожидаемый результат**: Kubernetes кластер принимает только подписанные образы

---

## Этап 6.2: SBOM генерация (Неделя 1.5-3)

### 6.2.1 Подготовка окружения (День 8)

**Действия**:
- [ ] Клонировать `anchore/syft` в `supply-chain/signing/syft/`
- [ ] Установить syft локально для тестирования
- [ ] Создать документацию по syft в `docs/supply-chain/syft-setup.md`
- [ ] Изучить форматы SBOM (SPDX, CycloneDX)

**Команды**:
```bash
# Клонирование syft
cd supply-chain/signing/
git clone https://github.com/anchore/syft.git syft/
cd syft/
git checkout v1.0.1  # Latest stable version

# Установка syft
curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin

# Проверка установки
syft version
```

**Ожидаемый результат**: syft установлен и готов к использованию

---

### 6.2.2 Генерация SBOM для образов (День 9-10)

**Действия**:
- [ ] Создать скрипт для генерации SBOM (`scripts/security/generate-sbom.sh`)
- [ ] Генерировать SBOM в форматах SPDX и CycloneDX
- [ ] Сохранять SBOM в artifacts
- [ ] Протестировать генерацию для всех образов

**Скрипт генерации SBOM**:
```bash
#!/bin/bash
# scripts/security/generate-sbom.sh

set -e

IMAGE="${1:-ghcr.io/2gc/cloudbridge-relay:latest}"
OUTPUT_DIR="${2:-./sbom}"

info() {
    echo "[INFO] $1"
}

# Создаем директорию для SBOM
mkdir -p "$OUTPUT_DIR"

# Генерация SBOM в SPDX формате
info "Generating SPDX SBOM for $IMAGE..."
syft "$IMAGE" -o spdx-json > "$OUTPUT_DIR/sbom-spdx.json"

# Генерация SBOM в CycloneDX формате
info "Generating CycloneDX SBOM for $IMAGE..."
syft "$IMAGE" -o cyclonedx-json > "$OUTPUT_DIR/sbom-cyclonedx.json"

# Генерация human-readable SBOM
info "Generating text SBOM for $IMAGE..."
syft "$IMAGE" -o table > "$OUTPUT_DIR/sbom.txt"

info "SBOM generation completed successfully"
ls -lh "$OUTPUT_DIR"
```

**Ожидаемый результат**: SBOM генерируется в нескольких форматах

---

### 6.2.3 Интеграция с CI/CD (День 11-12)

**Действия**:
- [ ] Добавить генерацию SBOM в GitHub Actions
- [ ] Загружать SBOM как artifacts
- [ ] Прикреплять SBOM к Docker образу как attestation
- [ ] Настроить хранение SBOM в S3/GCS

**GitHub Actions расширение**:
```yaml
# Добавить в .github/workflows/sign-images.yaml

      - name: Generate SBOM with syft
        run: |
          syft ghcr.io/${{ github.repository }}/cloudbridge-relay:${{ github.sha }} \
            -o spdx-json > sbom-spdx.json
          syft ghcr.io/${{ github.repository }}/cloudbridge-relay:${{ github.sha }} \
            -o cyclonedx-json > sbom-cyclonedx.json

      - name: Attest SBOM
        run: |
          cosign attest --yes \
            --predicate sbom-spdx.json \
            --type spdxjson \
            ghcr.io/${{ github.repository }}/cloudbridge-relay:${{ github.sha }}

      - name: Upload SBOM artifacts
        uses: actions/upload-artifact@v4
        with:
          name: sbom-${{ github.sha }}
          path: |
            sbom-spdx.json
            sbom-cyclonedx.json

      - name: Verify SBOM attestation
        run: |
          cosign verify-attestation \
            --type spdxjson \
            --certificate-identity-regexp="^https://github.com/${{ github.repository }}/" \
            --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
            ghcr.io/${{ github.repository }}/cloudbridge-relay:${{ github.sha }}
```

**Ожидаемый результат**: SBOM автоматически генерируется и прикрепляется к образам

---

### 6.2.4 Security Scanning интеграция (День 13-14)

**Действия**:
- [ ] Клонировать `anchore/grype` для vulnerability scanning
- [ ] Интегрировать grype с SBOM
- [ ] Создать отчеты по уязвимостям
- [ ] Настроить автоматические алерты

**Grype Integration**:
```bash
# Установка grype
curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh -s -- -b /usr/local/bin

# Сканирование SBOM на уязвимости
grype sbom:sbom-spdx.json -o json > vulnerabilities.json

# Сканирование по severity threshold
grype sbom:sbom-spdx.json --fail-on critical
```

**GitHub Actions добавление**:
```yaml
      - name: Scan SBOM for vulnerabilities
        run: |
          grype sbom:sbom-spdx.json -o json > vulnerabilities.json
          grype sbom:sbom-spdx.json --fail-on high

      - name: Upload vulnerability report
        uses: actions/upload-artifact@v4
        with:
          name: vulnerabilities-${{ github.sha }}
          path: vulnerabilities.json
```

**Ожидаемый результат**: Автоматическое сканирование уязвимостей на основе SBOM

---

## Критерии успеха

### Технические метрики
- [x] Все Docker образы подписаны с помощью cosign
- [x] SBOM генерируется для каждого релиза в форматах SPDX и CycloneDX
- [x] SBOM прикреплен к образам через cosign attestation
- [x] Kubernetes кластер принимает только подписанные образы
- [x] Vulnerability scanning работает автоматически
- [x] CI/CD pipeline проверяет подписи перед deployment

### Безопасность
- [x] Keyless signing через OIDC настроен
- [x] Ключи безопасно хранятся в Kubernetes Secrets
- [x] Admission Controller блокирует непроверенные образы
- [x] Vulnerability reports доступны для каждого релиза

### Compliance
- [x] SBOM доступен для аудита зависимостей
- [x] Подписи можно верифицировать независимо
- [x] Transparency log (Rekor) используется для audit trail
- [x] Документация по Supply Chain Security создана

---

## Документация

### Создать следующие документы:

1. **docs/supply-chain/cosign-setup.md**
   - Установка и настройка cosign
   - Генерация и управление ключами
   - Keyless signing через OIDC
   - Верификация подписей

2. **docs/supply-chain/syft-setup.md**
   - Установка syft
   - Генерация SBOM в разных форматах
   - Интеграция с CI/CD
   - Best practices для SBOM

3. **docs/supply-chain/security-scanning.md**
   - Настройка grype
   - Интерпретация отчетов об уязвимостях
   - Remediation workflow
   - Интеграция с monitoring

4. **docs/supply-chain/kyverno-policies.md**
   - Настройка Kyverno для image verification
   - Policy enforcement
   - Troubleshooting
   - Exemptions и exceptions

---

## Следующие шаги после завершения

После завершения Supply Chain Security:

1. **TURN сервер и NAT fallback** (Этап 1.3 доработка)
   - Настройка coturn TURN сервера
   - Создание fallback стратегий
   - Production тестирование

2. **External Secrets production deployment** (Этап 5.2 доработка)
   - Развертывание в production кластере
   - Тестирование автоматической ротации
   - Настройка мониторинга

3. **MetalLB/kube-vip** (Этап 5.1 доработка)
   - Интеграция MetalLB для bare-metal
   - Настройка kube-vip fallback
   - Документирование миграции

---

## Timeline

**Общее время**: 2-3 недели (14 рабочих дней)

- **Неделя 1 (День 1-7)**: cosign интеграция и подпись образов
- **Неделя 2 (День 8-14)**: syft интеграция и SBOM генерация
- **Неделя 3 (опционально)**: Тестирование и документация

**Текущий статус**: Готовы начать День 1
