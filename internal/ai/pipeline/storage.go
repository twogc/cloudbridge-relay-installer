package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RealFileStorage реализует реальное файловое хранилище данных
type RealFileStorage struct {
	config     *Config
	logger     *zap.Logger
	basePath   string
	buffer     []DataPoint
	bufferSize int
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewRealFileStorage создает новое файловое хранилище
func NewRealFileStorage(config *Config, logger *zap.Logger) *RealFileStorage {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Используем второй диск для хранения данных
	basePath := "/data/prometheus/cloudbridge/pipeline"
	if config.StoragePath != "" {
		basePath = config.StoragePath
	}
	
	return &RealFileStorage{
		config:     config,
		logger:     logger,
		basePath:   basePath,
		buffer:     make([]DataPoint, 0),
		bufferSize: config.BatchSize,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Store сохраняет данные
func (fs *RealFileStorage) Store(data []DataPoint) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	
	// Добавляем данные в буфер
	fs.buffer = append(fs.buffer, data...)
	
	// Если буфер заполнен, сохраняем данные
	if len(fs.buffer) >= fs.bufferSize {
		return fs.flushBuffer()
	}
	
	return nil
}

// flushBuffer сохраняет данные из буфера
func (fs *RealFileStorage) flushBuffer() error {
	if len(fs.buffer) == 0 {
		return nil
	}
	
	// Создаем директорию если не существует
	if err := os.MkdirAll(fs.basePath, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Группируем данные по дате и типу
	groupedData := fs.groupDataByDateAndType(fs.buffer)
	
	// Сохраняем каждую группу
	for key, data := range groupedData {
		if err := fs.saveGroup(key, data); err != nil {
			fs.logger.Error("Failed to save data group", 
				zap.String("key", key), 
				zap.Error(err))
			continue
		}
	}
	
	// Очищаем буфер
	fs.buffer = fs.buffer[:0]
	
	fs.logger.Debug("Data buffer flushed", 
		zap.Int("points_saved", len(fs.buffer)))
	
	return nil
}

// groupDataByDateAndType группирует данные по дате и типу
func (fs *RealFileStorage) groupDataByDateAndType(data []DataPoint) map[string][]DataPoint {
	groups := make(map[string][]DataPoint)
	
	for _, point := range data {
		date := point.Timestamp.Format("2006-01-02")
		key := fmt.Sprintf("%s_%s_%s", date, point.Source, point.Type)
		groups[key] = append(groups[key], point)
	}
	
	return groups
}

// saveGroup сохраняет группу данных
func (fs *RealFileStorage) saveGroup(key string, data []DataPoint) error {
	// Создаем имя файла
	fileName := fmt.Sprintf("%s.jsonl", key)
	filePath := filepath.Join(fs.basePath, fileName)
	
	// Проверяем безопасность пути
	if !strings.HasPrefix(filePath, fs.basePath) {
		return fmt.Errorf("unsafe file path: %s", filePath)
	}
	
	// Открываем файл для записи (append mode)
	// #nosec G304 - путь проверен на безопасность выше
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	// Записываем данные в формате JSONL
	encoder := json.NewEncoder(file)
	for _, point := range data {
		if err := encoder.Encode(point); err != nil {
			fs.logger.Error("Failed to encode data point", zap.Error(err))
			continue
		}
	}
	
	return nil
}

// Retrieve извлекает данные за указанный период
func (fs *RealFileStorage) Retrieve(timeRange TimeRange) ([]DataPoint, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	
	var allData []DataPoint
	
	// Сканируем файлы в диапазоне дат
	_ = timeRange.Start.Format("2006-01-02") // startDate
	_ = timeRange.End.Format("2006-01-02")   // endDate
	
	currentDate := timeRange.Start
	for currentDate.Before(timeRange.End) || currentDate.Equal(timeRange.End) {
		dateStr := currentDate.Format("2006-01-02")
		
		// Ищем файлы для этой даты
		pattern := filepath.Join(fs.basePath, dateStr+"_*.jsonl")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fs.logger.Warn("Failed to glob files", 
				zap.String("pattern", pattern), 
				zap.Error(err))
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}
		
		// Читаем каждый файл
		for _, filePath := range matches {
			data, err := fs.readFile(filePath, timeRange)
			if err != nil {
				fs.logger.Warn("Failed to read file", 
					zap.String("file", filePath), 
					zap.Error(err))
				continue
			}
			allData = append(allData, data...)
		}
		
		currentDate = currentDate.AddDate(0, 0, 1)
	}
	
	return allData, nil
}

// readFile читает файл и фильтрует данные по времени
func (fs *RealFileStorage) readFile(filePath string, timeRange TimeRange) ([]DataPoint, error) {
	// Проверяем безопасность пути
	if !strings.HasPrefix(filePath, fs.basePath) {
		return nil, fmt.Errorf("unsafe file path: %s", filePath)
	}
	
	// #nosec G304 - путь проверен на безопасность выше
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var data []DataPoint
	decoder := json.NewDecoder(file)
	
	for decoder.More() {
		var point DataPoint
		if err := decoder.Decode(&point); err != nil {
			fs.logger.Warn("Failed to decode data point", zap.Error(err))
			continue
		}
		
		// Фильтруем по времени
		if point.Timestamp.After(timeRange.Start) && point.Timestamp.Before(timeRange.End) {
			data = append(data, point)
		}
	}
	
	return data, nil
}

// Start запускает периодическое сохранение буфера
func (fs *RealFileStorage) Start() error {
	fs.logger.Info("Real file storage started", zap.String("base_path", fs.basePath))
	
	// Запускаем горутину для периодического сохранения
	go fs.periodicFlush()
	
	return nil
}

// Stop останавливает хранилище
func (fs *RealFileStorage) Stop() error {
	fs.logger.Info("Real file storage stopping...")
	
	// Отменяем контекст
	fs.cancel()
	
	// Сохраняем оставшиеся данные в буфере
	fs.mutex.Lock()
	if len(fs.buffer) > 0 {
		if err := fs.flushBuffer(); err != nil {
			fs.logger.Error("Failed to flush buffer on stop", zap.Error(err))
		}
	}
	fs.mutex.Unlock()
	
	fs.logger.Info("Real file storage stopped")
	return nil
}

// periodicFlush периодически сохраняет буфер
func (fs *RealFileStorage) periodicFlush() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-fs.ctx.Done():
			return
		case <-ticker.C:
			fs.mutex.Lock()
			if len(fs.buffer) > 0 {
				if err := fs.flushBuffer(); err != nil {
					fs.logger.Error("Failed to flush buffer periodically", zap.Error(err))
				}
			}
			fs.mutex.Unlock()
		}
	}
}

// InMemoryStorage реализует хранилище в памяти с TTL
type InMemoryStorage struct {
	logger   *zap.Logger
	data     map[string][]DataPoint
	mutex    sync.RWMutex
	ttl      time.Duration
	cleanup  *time.Ticker
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewInMemoryStorage создает новое хранилище в памяти
func NewInMemoryStorage(logger *zap.Logger, ttl time.Duration) *InMemoryStorage {
	ctx, cancel := context.WithCancel(context.Background())
	
	storage := &InMemoryStorage{
		logger:  logger,
		data:    make(map[string][]DataPoint),
		ttl:     ttl,
		cleanup: time.NewTicker(5 * time.Minute),
		ctx:     ctx,
		cancel:  cancel,
	}
	
	// Запускаем очистку устаревших данных
	go storage.cleanupExpired()
	
	return storage
}

// Store сохраняет данные в памяти
func (ims *InMemoryStorage) Store(data []DataPoint) error {
	ims.mutex.Lock()
	defer ims.mutex.Unlock()
	
	for _, point := range data {
		key := fmt.Sprintf("%s_%s", point.Source, point.Type)
		ims.data[key] = append(ims.data[key], point)
	}
	
	ims.logger.Debug("Data stored in memory", zap.Int("points", len(data)))
	return nil
}

// Retrieve извлекает данные из памяти
func (ims *InMemoryStorage) Retrieve(timeRange TimeRange) ([]DataPoint, error) {
	ims.mutex.RLock()
	defer ims.mutex.RUnlock()
	
	var result []DataPoint
	
	for _, points := range ims.data {
		for _, point := range points {
			if point.Timestamp.After(timeRange.Start) && point.Timestamp.Before(timeRange.End) {
				result = append(result, point)
			}
		}
	}
	
	return result, nil
}

// cleanupExpired очищает устаревшие данные
func (ims *InMemoryStorage) cleanupExpired() {
	for {
		select {
		case <-ims.ctx.Done():
			return
		case <-ims.cleanup.C:
			ims.mutex.Lock()
			
			cutoff := time.Now().Add(-ims.ttl)
			
			for key, points := range ims.data {
				var validPoints []DataPoint
				for _, point := range points {
					if point.Timestamp.After(cutoff) {
						validPoints = append(validPoints, point)
					}
				}
				
				if len(validPoints) == 0 {
					delete(ims.data, key)
				} else {
					ims.data[key] = validPoints
				}
			}
			
			ims.mutex.Unlock()
			ims.logger.Debug("Cleaned up expired data from memory")
		}
	}
}

// Stop останавливает хранилище
func (ims *InMemoryStorage) Stop() error {
	ims.cancel()
	ims.cleanup.Stop()
	ims.logger.Info("In-memory storage stopped")
	return nil
}

// HybridStorage комбинирует файловое и in-memory хранилище
type HybridStorage struct {
	logger        *zap.Logger
	fileStorage   *RealFileStorage
	memoryStorage *InMemoryStorage
	config        *Config
}

// NewHybridStorage создает новое гибридное хранилище
func NewHybridStorage(config *Config, logger *zap.Logger) *HybridStorage {
	return &HybridStorage{
		logger:        logger,
		fileStorage:   NewRealFileStorage(config, logger),
		memoryStorage: NewInMemoryStorage(logger, 1*time.Hour),
		config:        config,
	}
}

// Store сохраняет данные в оба хранилища
func (hs *HybridStorage) Store(data []DataPoint) error {
	// Сохраняем в файловое хранилище (основное)
	if err := hs.fileStorage.Store(data); err != nil {
		hs.logger.Error("Failed to store in file storage", zap.Error(err))
	}
	
	// Сохраняем в память (для быстрого доступа)
	if err := hs.memoryStorage.Store(data); err != nil {
		hs.logger.Error("Failed to store in memory storage", zap.Error(err))
	}
	
	return nil
}

// Retrieve извлекает данные (сначала из памяти, потом из файлов)
func (hs *HybridStorage) Retrieve(timeRange TimeRange) ([]DataPoint, error) {
	// Сначала пробуем получить из памяти
	memoryData, err := hs.memoryStorage.Retrieve(timeRange)
	if err != nil {
		hs.logger.Warn("Failed to retrieve from memory", zap.Error(err))
	}
	
	// Если данных в памяти недостаточно, дополняем из файлов
	if len(memoryData) == 0 {
		fileData, err := hs.fileStorage.Retrieve(timeRange)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve from file storage: %w", err)
		}
		return fileData, nil
	}
	
	return memoryData, nil
}

// Start запускает гибридное хранилище
func (hs *HybridStorage) Start() error {
	if err := hs.fileStorage.Start(); err != nil {
		return fmt.Errorf("failed to start file storage: %w", err)
	}
	
	hs.logger.Info("Hybrid storage started")
	return nil
}

// Stop останавливает гибридное хранилище
func (hs *HybridStorage) Stop() error {
	if err := hs.fileStorage.Stop(); err != nil {
		hs.logger.Error("Failed to stop file storage", zap.Error(err))
	}
	
	if err := hs.memoryStorage.Stop(); err != nil {
		hs.logger.Error("Failed to stop memory storage", zap.Error(err))
	}
	
	hs.logger.Info("Hybrid storage stopped")
	return nil
}
