package predictive

import (
	"sort"
	"sync"
	"time"
)

// TimeSeriesBuffer буфер для хранения временных рядов
type TimeSeriesBuffer struct {
	data    []TimeSeriesPoint
	maxSize int
	mu      sync.RWMutex
}

// TimeSeriesPoint точка временного ряда
type TimeSeriesPoint struct {
	Timestamp time.Time
	Value     float64
}

// NewTimeSeriesBuffer создает новый буфер временных рядов
func NewTimeSeriesBuffer(maxSize int) *TimeSeriesBuffer {
	return &TimeSeriesBuffer{
		data:    make([]TimeSeriesPoint, 0),
		maxSize: maxSize,
	}
}

// Add добавляет новую точку в буфер
func (tsb *TimeSeriesBuffer) Add(timestamp time.Time, value float64) {
	tsb.mu.Lock()
	defer tsb.mu.Unlock()

	point := TimeSeriesPoint{
		Timestamp: timestamp,
		Value:     value,
	}

	tsb.data = append(tsb.data, point)

	// Сортируем по времени
	sort.Slice(tsb.data, func(i, j int) bool {
		return tsb.data[i].Timestamp.Before(tsb.data[j].Timestamp)
	})

	// Ограничиваем размер буфера
	if len(tsb.data) > tsb.maxSize {
		tsb.data = tsb.data[len(tsb.data)-tsb.maxSize:]
	}
}

// GetLast возвращает последние n точек
func (tsb *TimeSeriesBuffer) GetLast(n int) []float64 {
	tsb.mu.RLock()
	defer tsb.mu.RUnlock()

	if len(tsb.data) == 0 {
		return []float64{}
	}

	start := len(tsb.data) - n
	if start < 0 {
		start = 0
	}

	result := make([]float64, len(tsb.data)-start)
	for i, point := range tsb.data[start:] {
		result[i] = point.Value
	}

	return result
}

// GetRange возвращает точки в заданном временном диапазоне
func (tsb *TimeSeriesBuffer) GetRange(start, end time.Time) []TimeSeriesPoint {
	tsb.mu.RLock()
	defer tsb.mu.RUnlock()

	var result []TimeSeriesPoint
	for _, point := range tsb.data {
		if point.Timestamp.After(start) && point.Timestamp.Before(end) {
			result = append(result, point)
		}
	}

	return result
}

// Size возвращает текущий размер буфера
func (tsb *TimeSeriesBuffer) Size() int {
	tsb.mu.RLock()
	defer tsb.mu.RUnlock()

	return len(tsb.data)
}

// Clear очищает буфер
func (tsb *TimeSeriesBuffer) Clear() {
	tsb.mu.Lock()
	defer tsb.mu.Unlock()

	tsb.data = tsb.data[:0]
}

// GetLatest возвращает последнюю точку
func (tsb *TimeSeriesBuffer) GetLatest() (TimeSeriesPoint, bool) {
	tsb.mu.RLock()
	defer tsb.mu.RUnlock()

	if len(tsb.data) == 0 {
		return TimeSeriesPoint{}, false
	}

	return tsb.data[len(tsb.data)-1], true
}

// GetOldest возвращает самую старую точку
func (tsb *TimeSeriesBuffer) GetOldest() (TimeSeriesPoint, bool) {
	tsb.mu.RLock()
	defer tsb.mu.RUnlock()

	if len(tsb.data) == 0 {
		return TimeSeriesPoint{}, false
	}

	return tsb.data[0], true
}
