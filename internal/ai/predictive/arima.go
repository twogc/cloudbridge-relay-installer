package predictive

import (
	"fmt"
	"math"
)

// ARIMAModel реализует ARIMA модель для прогнозирования временных рядов
type ARIMAModel struct {
	p, d, q      int
	coefficients []float64
	residuals    []float64
	trained      bool
}

// NewARIMAModel создает новую ARIMA модель
func NewARIMAModel(p, d, q int) *ARIMAModel {
	return &ARIMAModel{
		p:            p,
		d:            d,
		q:            q,
		coefficients: make([]float64, p+q+1),
		residuals:    make([]float64, 0),
		trained:      false,
	}
}

// Forecast выполняет прогнозирование на заданный горизонт
func (am *ARIMAModel) Forecast(data []float64, horizon int) ([]float64, error) {
	if !am.trained {
		if err := am.Train(data); err != nil {
			return nil, fmt.Errorf("failed to train model: %w", err)
		}
	}

	// Простое прогнозирование на основе тренда
	forecast := make([]float64, horizon)

	// Вычисляем тренд
	trend := am.calculateTrend(data)

	// Прогнозируем на основе тренда
	lastValue := data[len(data)-1]
	for i := 0; i < horizon; i++ {
		forecast[i] = lastValue + trend*float64(i+1)

		// Добавляем небольшую случайную вариацию
		variation := math.Sin(float64(i)) * 0.1
		forecast[i] += variation

		// Ограничиваем значения
		if forecast[i] < 0 {
			forecast[i] = 0
		}
		if forecast[i] > 100 {
			forecast[i] = 100
		}
	}

	return forecast, nil
}

// Train обучает ARIMA модель
func (am *ARIMAModel) Train(data []float64) error {
	if len(data) < 10 {
		return fmt.Errorf("insufficient data for training")
	}

	// Простое обучение - вычисляем коэффициенты на основе данных
	am.computeCoefficients(data)
	am.trained = true

	return nil
}

// Retrain переобучает модель на новых данных
func (am *ARIMAModel) Retrain(data []float64) {
	if err := am.Train(data); err != nil {
		// В реальной реализации здесь была бы обработка ошибки
		// Для демонстрации просто игнорируем ошибку
		_ = err
	}
}

// computeCoefficients вычисляет коэффициенты модели
func (am *ARIMAModel) computeCoefficients(data []float64) {
	// Простое вычисление коэффициентов
	// В реальной реализации здесь был бы более сложный алгоритм

	// AR коэффициенты (авторегрессия)
	for i := 0; i < am.p; i++ {
		am.coefficients[i] = 0.1 * float64(i+1)
	}

	// MA коэффициенты (скользящее среднее)
	for i := 0; i < am.q; i++ {
		am.coefficients[am.p+i] = 0.05 * float64(i+1)
	}

	// Константа
	am.coefficients[am.p+am.q] = 0.5
}

// calculateTrend вычисляет тренд в данных
func (am *ARIMAModel) calculateTrend(data []float64) float64 {
	if len(data) < 2 {
		return 0
	}

	// Простое вычисление тренда как среднее изменение
	var totalChange float64
	for i := 1; i < len(data); i++ {
		totalChange += data[i] - data[i-1]
	}

	return totalChange / float64(len(data)-1)
}

// GetCoefficients возвращает коэффициенты модели
func (am *ARIMAModel) GetCoefficients() []float64 {
	return am.coefficients
}

// IsTrained проверяет, обучена ли модель
func (am *ARIMAModel) IsTrained() bool {
	return am.trained
}
