package pipeline

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RealFeatureExtractor извлекает реальные признаки из данных
type RealFeatureExtractor struct {
	logger *zap.Logger
}

// NewRealFeatureExtractor создает новый извлекатель признаков
func NewRealFeatureExtractor(logger *zap.Logger) *RealFeatureExtractor {
	return &RealFeatureExtractor{
		logger: logger,
	}
}

// Process обрабатывает данные и извлекает признаки
func (fe *RealFeatureExtractor) Process(data []DataPoint) ([]DataPoint, error) {
	var processedData []DataPoint

	for _, point := range data {
		features := fe.extractFeatures(point)

		processedData = append(processedData, DataPoint{
			Timestamp: point.Timestamp,
			Source:    point.Source,
			Type:      "features",
			Data: map[string]interface{}{
				"original_data": point.Data,
				"features":      features,
				"feature_count": len(features),
			},
		})
	}

	return processedData, nil
}

// extractFeatures извлекает признаки из данных
func (fe *RealFeatureExtractor) extractFeatures(point DataPoint) map[string]interface{} {
	features := make(map[string]interface{})

	// Базовые признаки
	features["timestamp_unix"] = float64(point.Timestamp.Unix())
	features["hour_of_day"] = float64(point.Timestamp.Hour())
	features["day_of_week"] = float64(point.Timestamp.Weekday())
	features["is_weekend"] = fe.isWeekend(point.Timestamp)

	// Извлекаем числовые признаки
	numericFeatures := fe.extractNumericFeatures(point.Data)
	for key, value := range numericFeatures {
		features[key] = value
	}

	// Извлекаем статистические признаки
	statisticalFeatures := fe.extractStatisticalFeatures(point.Data)
	for key, value := range statisticalFeatures {
		features[key] = value
	}

	// Извлекаем категориальные признаки
	categoricalFeatures := fe.extractCategoricalFeatures(point.Data)
	for key, value := range categoricalFeatures {
		features[key] = value
	}

	// Извлекаем временные признаки
	temporalFeatures := fe.extractTemporalFeatures(point.Data)
	for key, value := range temporalFeatures {
		features[key] = value
	}

	return features
}

// extractNumericFeatures извлекает числовые признаки
func (fe *RealFeatureExtractor) extractNumericFeatures(data map[string]interface{}) map[string]interface{} {
	features := make(map[string]interface{})

	for key, value := range data {
		switch v := value.(type) {
		case float64:
			features[key] = v
			// Добавляем логарифмические признаки для больших значений
			if v > 0 {
				features[key+"_log"] = math.Log(v + 1)
			}
		case int:
			features[key] = float64(v)
			if v > 0 {
				features[key+"_log"] = math.Log(float64(v) + 1)
			}
		case int64:
			features[key] = float64(v)
			if v > 0 {
				features[key+"_log"] = math.Log(float64(v) + 1)
			}
		}
	}

	return features
}

// extractStatisticalFeatures извлекает статистические признаки
func (fe *RealFeatureExtractor) extractStatisticalFeatures(data map[string]interface{}) map[string]interface{} {
	features := make(map[string]interface{})

	// Собираем все числовые значения
	var values []float64
	for _, value := range data {
		switch v := value.(type) {
		case float64:
			values = append(values, v)
		case int:
			values = append(values, float64(v))
		case int64:
			values = append(values, float64(v))
		}
	}

	if len(values) > 0 {
		// Базовые статистики
		features["mean"] = fe.mean(values)
		features["median"] = fe.median(values)
		features["std"] = fe.std(values)
		features["min"] = fe.min(values)
		features["max"] = fe.max(values)
		features["range"] = fe.max(values) - fe.min(values)
		features["variance"] = fe.variance(values)
		features["skewness"] = fe.skewness(values)
		features["kurtosis"] = fe.kurtosis(values)

		// Квантили
		features["q25"] = fe.quantile(values, 0.25)
		features["q75"] = fe.quantile(values, 0.75)
		features["iqr"] = fe.quantile(values, 0.75) - fe.quantile(values, 0.25)
	}

	return features
}

// extractCategoricalFeatures извлекает категориальные признаки
func (fe *RealFeatureExtractor) extractCategoricalFeatures(data map[string]interface{}) map[string]interface{} {
	features := make(map[string]interface{})

	for key, value := range data {
		switch v := value.(type) {
		case string:
			features[key+"_length"] = float64(len(v))
			features[key+"_is_empty"] = fe.boolToFloat(len(v) == 0)
			features[key+"_has_numbers"] = fe.boolToFloat(fe.hasNumbers(v))
			features[key+"_has_special_chars"] = fe.boolToFloat(fe.hasSpecialChars(v))
		case bool:
			features[key] = fe.boolToFloat(v)
		}
	}

	return features
}

// extractTemporalFeatures извлекает временные признаки
func (fe *RealFeatureExtractor) extractTemporalFeatures(data map[string]interface{}) map[string]interface{} {
	features := make(map[string]interface{})

	// Ищем временные поля
	for key, value := range data {
		if strings.Contains(strings.ToLower(key), "time") ||
			strings.Contains(strings.ToLower(key), "date") ||
			strings.Contains(strings.ToLower(key), "timestamp") {

			if timeStr, ok := value.(string); ok {
				if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
					features[key+"_hour"] = float64(t.Hour())
					features[key+"_minute"] = float64(t.Minute())
					features[key+"_day"] = float64(t.Day())
					features[key+"_month"] = float64(t.Month())
					features[key+"_year"] = float64(t.Year())
					features[key+"_weekday"] = float64(t.Weekday())
					features[key+"_is_weekend"] = fe.isWeekend(t)
				}
			}
		}
	}

	return features
}

// Вспомогательные функции для статистики
func (fe *RealFeatureExtractor) mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (fe *RealFeatureExtractor) median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func (fe *RealFeatureExtractor) std(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := fe.mean(values)
	sum := 0.0
	for _, v := range values {
		sum += math.Pow(v-mean, 2)
	}
	return math.Sqrt(sum / float64(len(values)))
}

func (fe *RealFeatureExtractor) variance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := fe.mean(values)
	sum := 0.0
	for _, v := range values {
		sum += math.Pow(v-mean, 2)
	}
	return sum / float64(len(values))
}

func (fe *RealFeatureExtractor) min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func (fe *RealFeatureExtractor) max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func (fe *RealFeatureExtractor) quantile(values []float64, q float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	index := q * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func (fe *RealFeatureExtractor) skewness(values []float64) float64 {
	if len(values) < 3 {
		return 0
	}
	mean := fe.mean(values)
	std := fe.std(values)
	if std == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += math.Pow((v-mean)/std, 3)
	}
	return sum / float64(len(values))
}

func (fe *RealFeatureExtractor) kurtosis(values []float64) float64 {
	if len(values) < 4 {
		return 0
	}
	mean := fe.mean(values)
	std := fe.std(values)
	if std == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += math.Pow((v-mean)/std, 4)
	}
	return (sum / float64(len(values))) - 3 // Excess kurtosis
}

// Вспомогательные функции для категориальных признаков
func (fe *RealFeatureExtractor) hasNumbers(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func (fe *RealFeatureExtractor) hasSpecialChars(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ') {
			return true
		}
	}
	return false
}

func (fe *RealFeatureExtractor) isWeekend(t time.Time) float64 {
	weekday := t.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return 1.0
	}
	return 0.0
}

func (fe *RealFeatureExtractor) boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

// RealDataNormalizer нормализует данные с использованием различных методов
type RealDataNormalizer struct {
	logger *zap.Logger
	stats  map[string]*NormalizationStats
	mutex  sync.RWMutex
}

// NormalizationStats содержит статистики для нормализации
type NormalizationStats struct {
	Min        float64
	Max        float64
	Mean       float64
	Std        float64
	Count      int64
	LastUpdate time.Time
}

// NewRealDataNormalizer создает новый нормализатор данных
func NewRealDataNormalizer(logger *zap.Logger) *RealDataNormalizer {
	return &RealDataNormalizer{
		logger: logger,
		stats:  make(map[string]*NormalizationStats),
	}
}

// Process нормализует данные
func (dn *RealDataNormalizer) Process(data []DataPoint) ([]DataPoint, error) {
	// Обновляем статистики
	dn.updateStats(data)

	var normalizedData []DataPoint

	for _, point := range data {
		normalized := DataPoint{
			Timestamp: point.Timestamp,
			Source:    point.Source,
			Type:      "normalized",
			Data:      dn.normalizeData(point.Data),
		}
		normalizedData = append(normalizedData, normalized)
	}

	return normalizedData, nil
}

// updateStats обновляет статистики нормализации
func (dn *RealDataNormalizer) updateStats(data []DataPoint) {
	dn.mutex.Lock()
	defer dn.mutex.Unlock()

	// Собираем все значения по ключам
	keyValues := make(map[string][]float64)

	for _, point := range data {
		for key, value := range point.Data {
			if num, ok := dn.toFloat64(value); ok {
				keyValues[key] = append(keyValues[key], num)
			}
		}
	}

	// Обновляем статистики
	for key, values := range keyValues {
		if len(values) == 0 {
			continue
		}

		stats := &NormalizationStats{
			Count:      int64(len(values)),
			LastUpdate: time.Now(),
		}

		// Вычисляем статистики
		stats.Min = values[0]
		stats.Max = values[0]
		sum := 0.0

		for _, v := range values {
			if v < stats.Min {
				stats.Min = v
			}
			if v > stats.Max {
				stats.Max = v
			}
			sum += v
		}

		stats.Mean = sum / float64(len(values))

		// Вычисляем стандартное отклонение
		variance := 0.0
		for _, v := range values {
			variance += math.Pow(v-stats.Mean, 2)
		}
		stats.Std = math.Sqrt(variance / float64(len(values)))

		dn.stats[key] = stats
	}
}

// normalizeData нормализует данные
func (dn *RealDataNormalizer) normalizeData(data map[string]interface{}) map[string]interface{} {
	dn.mutex.RLock()
	defer dn.mutex.RUnlock()

	normalized := make(map[string]interface{})

	for key, value := range data {
		if num, ok := dn.toFloat64(value); ok {
			if stats, exists := dn.stats[key]; exists {
				// Min-Max нормализация
				if stats.Max != stats.Min {
					normalized[key+"_minmax"] = (num - stats.Min) / (stats.Max - stats.Min)
				} else {
					normalized[key+"_minmax"] = 0.0
				}

				// Z-score нормализация
				if stats.Std > 0 {
					normalized[key+"_zscore"] = (num - stats.Mean) / stats.Std
				} else {
					normalized[key+"_zscore"] = 0.0
				}

				// Robust нормализация (используя медиану и IQR)
				normalized[key+"_robust"] = dn.robustNormalize(num, stats)

				// Логарифмическая нормализация
				if num > 0 {
					normalized[key+"_log"] = math.Log(num + 1)
				} else {
					normalized[key+"_log"] = 0.0
				}
			}

			// Сохраняем оригинальное значение
			normalized[key] = num
		} else {
			normalized[key] = value
		}
	}

	return normalized
}

// robustNormalize выполняет robust нормализацию
func (dn *RealDataNormalizer) robustNormalize(value float64, stats *NormalizationStats) float64 {
	// Используем квантили для robust нормализации
	// Для простоты используем приближение через min/max
	if stats.Max != stats.Min {
		return (value - stats.Min) / (stats.Max - stats.Min)
	}
	return 0.0
}

// toFloat64 конвертирует значение в float64
func (dn *RealDataNormalizer) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num, true
		}
	}
	return 0, false
}

// DataAggregator агрегирует данные по временным окнам
type DataAggregator struct {
	logger *zap.Logger
	window time.Duration
}

// NewDataAggregator создает новый агрегатор данных
func NewDataAggregator(logger *zap.Logger, window time.Duration) *DataAggregator {
	return &DataAggregator{
		logger: logger,
		window: window,
	}
}

// Process агрегирует данные
func (da *DataAggregator) Process(data []DataPoint) ([]DataPoint, error) {
	if len(data) == 0 {
		return data, nil
	}

	// Группируем данные по временным окнам
	windows := make(map[int64][]DataPoint)

	for _, point := range data {
		windowKey := point.Timestamp.Truncate(da.window).Unix()
		windows[windowKey] = append(windows[windowKey], point)
	}

	var aggregatedData []DataPoint

	for windowTime, points := range windows {
		aggregated := da.aggregateWindow(time.Unix(windowTime, 0), points)
		aggregatedData = append(aggregatedData, aggregated)
	}

	return aggregatedData, nil
}

// aggregateWindow агрегирует данные в одном временном окне
func (da *DataAggregator) aggregateWindow(windowTime time.Time, points []DataPoint) DataPoint {
	// Собираем все числовые значения по ключам
	keyValues := make(map[string][]float64)

	for _, point := range points {
		for key, value := range point.Data {
			if num, ok := da.toFloat64(value); ok {
				keyValues[key] = append(keyValues[key], num)
			}
		}
	}

	// Агрегируем значения
	aggregatedData := make(map[string]interface{})

	for key, values := range keyValues {
		if len(values) == 0 {
			continue
		}

		aggregatedData[key+"_count"] = float64(len(values))
		aggregatedData[key+"_sum"] = da.sum(values)
		aggregatedData[key+"_mean"] = da.mean(values)
		aggregatedData[key+"_min"] = da.min(values)
		aggregatedData[key+"_max"] = da.max(values)
		aggregatedData[key+"_std"] = da.std(values)
		aggregatedData[key+"_median"] = da.median(values)
	}

	return DataPoint{
		Timestamp: windowTime,
		Source:    "aggregated",
		Type:      "aggregated",
		Data:      aggregatedData,
	}
}

// Вспомогательные функции для агрегации
func (da *DataAggregator) sum(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum
}

func (da *DataAggregator) mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return da.sum(values) / float64(len(values))
}

func (da *DataAggregator) min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func (da *DataAggregator) max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func (da *DataAggregator) std(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := da.mean(values)
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	return math.Sqrt(variance / float64(len(values)))
}

func (da *DataAggregator) median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func (da *DataAggregator) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num, true
		}
	}
	return 0, false
}
