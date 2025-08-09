# 🚀 Быстрая настройка секретов GitHub

## ⚡ Создайте 3 секрета:

### 1. Откройте ссылку:
```
https://github.com/twogc/cloudbridge-relay-installer/settings/secrets/actions
```

### 2. Создайте секреты:

| Имя | Значение |
|-----|----------|
| `GH_TOKEN` | `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` |
| `DOCKERHUB_USERNAME` | `mlanies` |
| `DOCKERHUB_TOKEN` | `dckr_pat_FS4aVfXofxlgVoxsBT74KNVhTMs` |

## ⚠️ ВАЖНО:
- **НЕ используйте** `GITHUB_TOKEN` - это зарезервировано!
- **Используйте** `GH_TOKEN` вместо этого
- Убедитесь, что токен имеет права `workflow`

## 🧪 Проверка:
После создания секретов у вас должно быть:
- ✅ GH_TOKEN
- ✅ DOCKERHUB_USERNAME  
- ✅ DOCKERHUB_TOKEN

## 🚀 Тестирование:
```bash
git tag v1.0.0
git push origin v1.0.0
# Workflow запустится автоматически!
```
