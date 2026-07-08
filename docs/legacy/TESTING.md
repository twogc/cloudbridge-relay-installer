# Тестирование AI сервиса

Этот документ описывает систему тестирования для AI сервиса CloudBridge Relay Installer.

## Обзор

AI сервис включает в себя комплексную систему тестирования, которая покрывает:

- **Unit тесты** - тестирование отдельных компонентов
- **Интеграционные тесты** - тестирование взаимодействия компонентов
- **Тесты производительности** - бенчмарки и нагрузочное тестирование
- **Тесты обработки ошибок** - проверка устойчивости к сбоям
- **Тесты покрытия кода** - анализ покрытия тестами

## Быстрый старт

### Запуск всех тестов
```bash
make -f Makefile.test test-all
```

### Запуск тестов с покрытием кода
```bash
make -f Makefile.test test-coverage
```

### Просмотр покрытия кода
```bash
make -f Makefile.test coverage-func
```

## Структура тестов

```
internal/ai/
├── routing/
│   ├── network_metrics_collector_test.go      # Тесты NetworkMetricsCollector
│   ├── path_finder_real_data_test.go          # Тесты PathFinder с реальными данными
│   └── latency_predictor_real_data_test.go    # Тесты LatencyPredictor с реальными данными
├── cadence/
│   └── cadence_integration_real_test.go       # Тесты реальной интеграции с Cadence
├── predictive/
│   └── real_load_predictor_improved_test.go   # Тесты улучшенного RealLoadPredictor
├── real_service_integration_test.go           # Интеграционные тесты RealService
├── error_handling_test.go                     # Тесты обработки ошибок
└── performance_test.go                        # Тесты производительности
```

## Типы тестов

### 1. Unit тесты

Тестируют отдельные компоненты в изоляции:

```bash
# Все unit тесты
make -f Makefile.test test-unit

# Конкретные компоненты
make -f Makefile.test test-network-metrics
make -f Makefile.test test-path-finder
make -f Makefile.test test-latency-predictor
```

**Покрытие:**
- NetworkMetricsCollector - сбор метрик, обработка ошибок, периодический сбор
- PathFinder - поиск путей, расчет загрузки, обновление топологии
- LatencyPredictor - предсказание латентности, вариации, стабильность
- RealLoadPredictor - прогнозирование нагрузки без синусоидальных заглушек

### 2. Интеграционные тесты

Тестируют взаимодействие компонентов:

```bash
make -f Makefile.test test-integration
```

**Покрытие:**
- RealService - инициализация, запуск, остановка, все компоненты
- Cadence интеграция - API вызовы, fallback, обработка ошибок
- Взаимодействие между компонентами

### 3. Тесты производительности

Бенчмарки и нагрузочное тестирование:

```bash
make -f Makefile.test test-performance
```

**Покрытие:**
- BenchmarkRealServiceAnomalyDetection
- BenchmarkRealServiceLoadPrediction
- BenchmarkRealServiceRouteOptimization
- Тесты латентности под нагрузкой
- Тесты пропускной способности
- Тесты использования памяти

### 4. Тесты обработки ошибок

Проверка устойчивости к сбоям:

```bash
make -f Makefile.test test-error-handling
```

**Покрытие:**
- Недоступность внешних сервисов
- Таймауты сети
- Некорректные данные
- Частичные отказы сервисов
- Отмена контекста
- Некорректная конфигурация

##  Конфигурация тестов

### Переменные окружения

```bash
# URL для тестирования (по умолчанию используются mock серверы)
export PROMETHEUS_URL="http://localhost:9090"
export CLOUDBRIDGE_URL="http://localhost:8080"
export CADENCE_URL="http://localhost:8080"

# Настройки тестирования
export TEST_TIMEOUT="30s"
export TEST_PARALLEL="4"
export TEST_RACE="true"
```

### Mock серверы

Тесты используют встроенные mock серверы для имитации внешних зависимостей:

- **Prometheus mock** - возвращает метрики в формате Prometheus API
- **CloudBridge Relay mock** - возвращает статистику релея
- **Cadence mock** - возвращает данные о workflow

## Анализ покрытия

### Генерация отчета о покрытии

```bash
# HTML отчет
make -f Makefile.test test-coverage
open coverage/coverage.html

# Текстовый отчет
make -f Makefile.test coverage-text

# Отчет по функциям
make -f Makefile.test coverage-func
```

### Цели покрытия

- **Unit тесты**: > 80%
- **Интеграционные тесты**: > 70%
- **Критические компоненты**: > 90%

## Обработка ошибок

### Типы ошибок в тестах

1. **Сетевые ошибки**
   - Таймауты
   - Недоступность сервисов
   - Некорректные ответы

2. **Ошибки данных**
   - Некорректный JSON
   - Отсутствующие поля
   - Неверные типы данных

3. **Ошибки конфигурации**
   - Некорректные параметры
   - Отсутствующие настройки
   - Неверные URL

### Стратегии обработки

- **Fallback данные** - использование резервных данных при сбоях
- **Retry логика** - повторные попытки при временных сбоях
- **Graceful degradation** - постепенное снижение функциональности
- **Circuit breaker** - отключение проблемных компонентов

## Производительность

### Метрики производительности

- **Латентность операций** - время выполнения отдельных операций
- **Пропускная способность** - количество операций в секунду
- **Использование памяти** - потребление RAM
- **Использование CPU** - загрузка процессора

### Бенчмарки

```bash
# Все бенчмарки
make -f Makefile.test test-performance

# Конкретные бенчмарки
go test -bench=BenchmarkRealServiceAnomalyDetection ./internal/ai/
go test -bench=BenchmarkRealServiceLoadPrediction ./internal/ai/
go test -bench=BenchmarkRealServiceRouteOptimization ./internal/ai/
```

### Цели производительности

- **Обнаружение аномалий**: < 100ms
- **Прогнозирование нагрузки**: < 200ms
- **Оптимизация маршрутов**: < 50ms
- **Пропускная способность**: > 100 ops/sec

## Отладка тестов

### Детальный вывод

```bash
# Подробный вывод
make -f Makefile.test test-verbose

# С таймаутом
make -f Makefile.test test-timeout

# В параллельном режиме
make -f Makefile.test test-parallel
```

### Профилирование

```bash
# CPU профилирование
make -f Makefile.test test-profile

# Анализ профилей
go tool pprof cpu.prof
go tool pprof mem.prof
```

### Race detection

```bash
# Детекция гонок данных
make -f Makefile.test test-race

# Детальная детекция
make -f Makefile.test test-race-detailed
```

## 🧹 Очистка

```bash
# Очистка временных файлов
make -f Makefile.test clean

# Очистка кэша тестов
go clean -testcache
```

##  Непрерывная интеграция

### GitHub Actions

```yaml
name: AI Service Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: make -f Makefile.test deps
      - run: make -f Makefile.test test-all
      - run: make -f Makefile.test test-coverage
```

### Локальная проверка

```bash
# Полная проверка (линтер + тесты)
make -f Makefile.test check

# Быстрая проверка (только unit тесты)
make -f Makefile.test quick-check
```

## Лучшие практики

### Написание тестов

1. **Именование тестов**
   - Используйте описательные имена
   - Включайте ожидаемое поведение
   - Пример: `TestPathFinderWithHighLoadNodes`

2. **Структура тестов**
   - Arrange - подготовка данных
   - Act - выполнение действия
   - Assert - проверка результата

3. **Mock объекты**
   - Используйте mock серверы для внешних зависимостей
   - Проверяйте вызовы mock объектов
   - Тестируйте различные сценарии ответов

### Поддержка тестов

1. **Регулярное обновление**
   - Обновляйте тесты при изменении кода
   - Добавляйте тесты для новых функций
   - Удаляйте устаревшие тесты

2. **Мониторинг покрытия**
   - Отслеживайте изменения покрытия
   - Увеличивайте покрытие критических компонентов
   - Анализируйте непокрытые участки

3. **Производительность тестов**
   - Оптимизируйте медленные тесты
   - Используйте параллельное выполнение
   - Кэшируйте результаты где возможно

## Дополнительные ресурсы

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Go Benchmark Documentation](https://golang.org/pkg/testing/#hdr-Benchmarks)
- [Go Coverage Documentation](https://golang.org/cmd/cover/)
- [Go Race Detector](https://golang.org/doc/articles/race_detector.html)

## 🤝 Вклад в тестирование

При добавлении новых функций:

1. Напишите unit тесты для новой функциональности
2. Добавьте интеграционные тесты если необходимо
3. Создайте бенчмарки для критических путей
4. Обновите тесты обработки ошибок
5. Проверьте покрытие кода
6. Запустите все тесты перед коммитом

```bash
# Полная проверка перед коммитом
make -f Makefile.test check
```
