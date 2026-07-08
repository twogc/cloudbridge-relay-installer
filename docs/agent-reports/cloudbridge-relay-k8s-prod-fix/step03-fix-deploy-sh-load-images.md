# Отчёт агента — cursor: fix deploy.sh load_images

**Дата:** 2026-07-09 02:13  
**Миссия:** cloudbridge-relay-k8s-prod-fix  
**Шаг:** 3

## Выполнено

1. **`scripts/deploy/deploy.sh` → `load_images()`**
   - **ai-service:** alpine-stub (echo на 8080) заменён на сборку из `docker/dockerfiles/Dockerfile.ai-service` (`go build ./cmd/ai-service`, EXPOSE 8083).
   - **cadence:** удалены stub build/tag/push `cadence:latest`; добавлен log: образ `ubercadence/server:v0.22.4` берётся с Docker Hub (как в `05-deployments.yaml`).
   - **relay:** без изменений — сборка из `docker/Dockerfile`.
   - Tag/push только `cloudbridge-relay:latest` и `ai-service:latest` → `localhost:5000/...` (port-forward к in-cluster registry).

2. **`k8s-manifests/05-deployments.yaml`** — согласование image с registry flow:
   - `docker-registry:5000/cloudbridge-relay:latest` + `imagePullPolicy: IfNotPresent`
   - `docker-registry:5000/ai-service:latest` + `imagePullPolicy: IfNotPresent`
   - `ubercadence/server:v0.22.4` + `imagePullPolicy: IfNotPresent` (Docker Hub, без push)

3. **`k8s-manifests/README.md`** — документирована стратегия pull: localhost push ↔ in-cluster `docker-registry:5000/...` ↔ cadence с Docker Hub.

## Изменённые файлы

| Путь | Действие |
|------|----------|
| `scripts/deploy/deploy.sh` | real ai build, cadence stub removed, push только relay/ai |
| `k8s-manifests/05-deployments.yaml` | image paths + imagePullPolicy для relay/ai/cadence |
| `k8s-manifests/README.md` | секция «Стратегия pull (step 3)» |
| `docs/agent-reports/cloudbridge-relay-k8s-prod-fix/step03-fix-deploy-sh-load-images.md` | создан (этот отчёт) |

## Проверки (команды и результат)

```bash
# Bash syntax
bash -n scripts/deploy/deploy.sh
# → bash -n OK

# Alpine stubs removed from deploy.sh
rg 'alpine|cadence:latest|8080|echo .AI Service' scripts/deploy/deploy.sh
# → (no matches)

# Image alignment deploy.sh ↔ 05
rg 'docker build|docker push|docker tag|image:|Dockerfile' \
  scripts/deploy/deploy.sh k8s-manifests/05-deployments.yaml
# → ai: docker/dockerfiles/Dockerfile.ai-service
# → push localhost:5000/cloudbridge-relay:latest, localhost:5000/ai-service:latest
# → 05: docker-registry:5000/cloudbridge-relay:latest, docker-registry:5000/ai-service:latest
# → cadence: ubercadence/server:v0.22.4 (no push in deploy.sh)

# YAML parse + image fields
python3 -c "
import yaml
with open('k8s-manifests/05-deployments.yaml') as f:
    for d in yaml.safe_load_all(f):
        c = d['spec']['template']['spec']['containers'][0]
        print(d['metadata']['name'], c['image'], c.get('imagePullPolicy'))
"
# → cloudbridge-relay docker-registry:5000/cloudbridge-relay:latest IfNotPresent
# → ai-service docker-registry:5000/ai-service:latest IfNotPresent
# → cadence ubercadence/server:v0.22.4 IfNotPresent

# Real ai-service Docker build
docker build -f docker/dockerfiles/Dockerfile.ai-service -t ai-service:latest .
# → exit 0 (~85s)

# Verify ai-service is real binary (not alpine echo)
docker inspect ai-service:latest --format '{{.Config.Cmd}} {{.Config.ExposedPorts}}'
# → [./ai-service] map[8083/tcp:{}]

docker run --rm ai-service:latest ./ai-service --help 2>&1 | head -3
# → AI service initializes (Go logs, not shell echo loop)

# Cadence image from Docker Hub
docker pull ubercadence/server:v0.22.4
# → Status: Downloaded newer image for ubercadence/server:v0.22.4
```

## Блокеры / не сделано

- **End-to-end push в registry** не проверен — нет доступного Kubernetes-кластера (`kubectl` не в PATH); port-forward + `docker push localhost:5000/...` не прогонялись.
- **Cadence runtime env** (Cassandra/Ringpop/SERVICES) не добавлялся — вне scope step 3; pod может потребовать доп. конфигурацию из `k8s/cadence/`.
- **Commit/push** — не выполнялся (по инструкции).
