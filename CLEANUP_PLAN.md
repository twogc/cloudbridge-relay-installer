# 🧹 План очистки папки k8s

## 📊 Анализ текущего состояния

### ✅ Рабочие файлы (оставить):
- `k8s/deployments/cadence-deployment-simple.yaml` - ✅ РАБОЧИЙ (используется)
- `k8s/cadence/configmap-simplified.yaml` - ✅ РАБОЧИЙ (используется)
- `k8s/deployments/cloudbridge-relay-deployment.yaml` - ✅ РАБОЧИЙ
- `k8s/deployments/cassandra-deployment.yaml` - ✅ РАБОЧИЙ
- `k8s/monitoring/prometheus-deployment.yaml` - ✅ РАБОЧИЙ
- `k8s/monitoring/grafana-deployment.yaml` - ✅ РАБОЧИЙ
- `k8s/services/cloudbridge-services.yaml` - ✅ РАБОЧИЙ
- `k8s/namespaces/cloudbridge-namespace.yaml` - ✅ РАБОЧИЙ
- `k8s/persistent-volumes/cloudbridge-pv.yaml` - ✅ РАБОЧИЙ
- `k8s/configmaps/cloudbridge-config.yaml` - ✅ РАБОЧИЙ
- `k8s/secrets/cloudbridge-secrets.yaml` - ✅ РАБОЧИЙ

### ❌ Ненужные файлы (удалить):

#### Cadence (множество дубликатов):
- `k8s/cadence/configmap-fixed.yaml` - дубликат
- `k8s/cadence/configmap-minimal.yaml` - дубликат
- `k8s/cadence/configmap-template.yaml` - дубликат
- `k8s/cadence/configmap.yaml` - старый
- `k8s/cadence/cadence-dynamicconfig.yaml` - не используется
- `k8s/cadence/frontend-deployment.yaml` - дубликат
- `k8s/cadence/frontend-service.yaml` - дубликат
- `k8s/cadence/namespace.yaml` - дубликат

#### Deployments (множество дубликатов):
- `k8s/deployments/cadence-deployment-env-only.yaml` - дубликат
- `k8s/deployments/cadence-deployment-final.yaml` - дубликат
- `k8s/deployments/cadence-deployment-fixed-final.yaml` - дубликат
- `k8s/deployments/cadence-deployment-fixed.yaml` - дубликат
- `k8s/deployments/cadence-deployment.yaml` - старый
- `k8s/deployments/cassandra-minimal.yaml` - дубликат
- `k8s/deployments/cassandra-simple.yaml` - дубликат

#### Неиспользуемые папки:
- `k8s/ai-service/` - не используется
- `k8s/autoscaling/` - не используется
- `k8s/cadence-service/` - не используется
- `k8s/cassandra/` - дубликат
- `k8s/ingress/` - не используется
- `k8s/production/` - не используется
- `k8s/quantum-service/` - не используется
- `k8s/quic-service/` - не используется
- `k8s/staging/` - не используется
- `k8s/velero/` - не используется
- `k8s/wireguard-service/` - не используется

#### Отдельные файлы:
- `k8s/cloudbridge-relay-wireguard.yaml` - не используется
- `k8s/ingress.yaml` - не используется
- `k8s/namespace.yaml` - дубликат
- `k8s/storage-classes.yaml` - не используется

## 🎯 План действий

### 1. Удалить дубликаты Cadence
```bash
rm k8s/cadence/configmap-fixed.yaml
rm k8s/cadence/configmap-minimal.yaml
rm k8s/cadence/configmap-template.yaml
rm k8s/cadence/configmap.yaml
rm k8s/cadence/cadence-dynamicconfig.yaml
rm k8s/cadence/frontend-deployment.yaml
rm k8s/cadence/frontend-service.yaml
rm k8s/cadence/namespace.yaml
```

### 2. Удалить дубликаты Deployments
```bash
rm k8s/deployments/cadence-deployment-env-only.yaml
rm k8s/deployments/cadence-deployment-final.yaml
rm k8s/deployments/cadence-deployment-fixed-final.yaml
rm k8s/deployments/cadence-deployment-fixed.yaml
rm k8s/deployments/cadence-deployment.yaml
rm k8s/deployments/cassandra-minimal.yaml
rm k8s/deployments/cassandra-simple.yaml
```

### 3. Удалить неиспользуемые папки
```bash
rm -rf k8s/ai-service/
rm -rf k8s/autoscaling/
rm -rf k8s/cadence-service/
rm -rf k8s/cassandra/
rm -rf k8s/ingress/
rm -rf k8s/production/
rm -rf k8s/quantum-service/
rm -rf k8s/quic-service/
rm -rf k8s/staging/
rm -rf k8s/velero/
rm -rf k8s/wireguard-service/
```

### 4. Удалить отдельные файлы
```bash
rm k8s/cloudbridge-relay-wireguard.yaml
rm k8s/ingress.yaml
rm k8s/namespace.yaml
rm k8s/storage-classes.yaml
```

## 📊 Результат очистки

### До очистки:
- **Всего файлов**: ~50+
- **Дубликатов**: ~30+
- **Неиспользуемых**: ~20+

### После очистки:
- **Рабочих файлов**: ~15
- **Дубликатов**: 0
- **Неиспользуемых**: 0

## 🎯 Итоговая структура

```
k8s/
├── deployments/
│   ├── cadence-deployment-simple.yaml      # ✅ РАБОЧИЙ
│   ├── cloudbridge-relay-deployment.yaml  # ✅ РАБОЧИЙ
│   └── cassandra-deployment.yaml          # ✅ РАБОЧИЙ
├── cadence/
│   └── configmap-simplified.yaml          # ✅ РАБОЧИЙ
├── monitoring/
│   ├── prometheus-deployment.yaml         # ✅ РАБОЧИЙ
│   ├── grafana-deployment.yaml            # ✅ РАБОЧИЙ
│   ├── prometheus-config.yaml             # ✅ РАБОЧИЙ
│   └── namespace.yaml                     # ✅ РАБОЧИЙ
├── services/
│   └── cloudbridge-services.yaml          # ✅ РАБОЧИЙ
├── namespaces/
│   └── cloudbridge-namespace.yaml         # ✅ РАБОЧИЙ
├── persistent-volumes/
│   └── cloudbridge-pv.yaml                # ✅ РАБОЧИЙ
├── configmaps/
│   └── cloudbridge-config.yaml            # ✅ РАБОЧИЙ
└── secrets/
    └── cloudbridge-secrets.yaml           # ✅ РАБОЧИЙ
```

## 🚀 Преимущества очистки

1. **Упрощение структуры** - только нужные файлы
2. **Устранение путаницы** - нет дубликатов
3. **Легкость поддержки** - понятная структура
4. **Быстрота развертывания** - меньше файлов для обработки
5. **Чистота репозитория** - убраны неиспользуемые файлы

---

*План создан: 7 сентября 2025*
