# 🚀 Создание GitHub Workflow вручную

## ❗ Проблема
GitHub блокирует создание workflow файлов через Git push, даже если токен имеет права `workflow`. Это может происходить из-за особенностей работы GitHub API.

## ✅ Решение: Создать workflow через веб-интерфейс

### Шаг 1: Откройте GitHub Actions
1. Перейдите на https://github.com/twogc/cloudbridge-relay-installer
2. Нажмите на вкладку **Actions**
3. Нажмите **New workflow**
4. Выберите **"set up a workflow yourself"**

### Шаг 2: Создайте файл release.yml
1. Назовите файл: `release.yml`
2. Скопируйте содержимое из файла `release_workflow_ready.yml`
3. Нажмите **Start commit**
4. Добавьте сообщение: `feat: добавлен автоматический workflow для релизов`
5. Нажмите **Commit new file**

### Шаг 3: Проверьте секреты
Убедитесь, что созданы секреты:
- ✅ `GH_TOKEN`
- ✅ `DOCKERHUB_USERNAME` 
- ✅ `DOCKERHUB_TOKEN`

### Шаг 4: Тестирование
После создания workflow запустите тест:

```bash
# Создайте тег
git tag v1.0.0

# Отправьте тег
git push origin v1.0.0
```

Workflow запустится автоматически!

## 🔍 Альтернативные варианты

### Вариант A: Локальное создание релиза
```bash
./tools/utils/create_release.sh v1.0.0
```

### Вариант B: Ручной запуск workflow
1. Перейдите в **Actions**
2. Выберите **Release** workflow
3. Нажмите **Run workflow**
4. Введите версию (например, `v1.0.0`)
5. Нажмите **Run workflow**

## 📊 Что делает workflow

- ✅ Сборка для Linux (amd64, arm64)
- ✅ Сборка для Windows (amd64)
- ✅ Сборка для macOS (amd64, arm64)
- ✅ Создание архивов (.tar.gz, .zip)
- ✅ Docker образы для Docker Hub и GitHub Registry
- ✅ Автоматический changelog
- ✅ GitHub Release с файлами

## 🎯 После создания workflow

1. Workflow будет запускаться при создании тегов `v*`
2. Можно запускать вручную через GitHub UI
3. Релизы будут создаваться автоматически
4. Docker образы будут загружаться на Hub
