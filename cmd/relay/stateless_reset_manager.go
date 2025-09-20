package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// StatelessResetManager управляет StatelessResetKey для корректного завершения QUIC соединений
type StatelessResetManager struct {
	logger    *logger.Logger
	key       quic.StatelessResetKey
	keyFile   string
	transport *quic.Transport

	// Защита от одновременного доступа
	mutex sync.RWMutex
}

// NewStatelessResetManager создает новый менеджер StatelessResetKey
func NewStatelessResetManager(logger *logger.Logger, keyFile string) *StatelessResetManager {
	return &StatelessResetManager{
		logger:  logger,
		keyFile: keyFile,
	}
}

// LoadOrGenerateKey загружает существующий ключ или генерирует новый
func (srm *StatelessResetManager) LoadOrGenerateKey() error {
	srm.mutex.Lock()
	defer srm.mutex.Unlock()

	// Попытаться загрузить существующий ключ
	if err := srm.loadKey(); err == nil {
		srm.logger.Info("StatelessResetKey loaded from file",
			zap.String("key_file", srm.keyFile),
			zap.String("key_hex", hex.EncodeToString(srm.key[:])))
		return nil
	}

	// Если загрузка не удалась, генерируем новый ключ
	srm.logger.Info("Failed to load StatelessResetKey, generating new one",
		zap.String("key_file", srm.keyFile))

	if err := srm.generateKey(); err != nil {
		return fmt.Errorf("failed to generate StatelessResetKey: %v", err)
	}

	// Сохраняем новый ключ
	if err := srm.saveKey(); err != nil {
		srm.logger.Warn("Failed to save StatelessResetKey", zap.Error(err))
		// Не возвращаем ошибку, так как ключ уже сгенерирован
	}

	srm.logger.Info("StatelessResetKey generated and saved",
		zap.String("key_file", srm.keyFile),
		zap.String("key_hex", hex.EncodeToString(srm.key[:])))

	return nil
}

// loadKey загружает ключ из файла
func (srm *StatelessResetManager) loadKey() error {
	if srm.keyFile == "" {
		return fmt.Errorf("key file path is empty")
	}

	// Проверяем существование файла
	if _, err := os.Stat(srm.keyFile); os.IsNotExist(err) {
		return fmt.Errorf("key file does not exist: %s", srm.keyFile)
	}

	// Читаем ключ из файла
	keyData, err := os.ReadFile(srm.keyFile)
	if err != nil {
		return fmt.Errorf("failed to read key file: %v", err)
	}

	// Декодируем hex строку
	keyHex := string(keyData)
	keyHex = keyHex[:len(keyHex)-1] // Убираем символ новой строки

	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		return fmt.Errorf("failed to decode key hex: %v", err)
	}

	if len(keyBytes) != len(srm.key) {
		return fmt.Errorf("invalid key length: expected %d, got %d", len(srm.key), len(keyBytes))
	}

	copy(srm.key[:], keyBytes)
	return nil
}

// generateKey генерирует новый случайный ключ
func (srm *StatelessResetManager) generateKey() error {
	_, err := rand.Read(srm.key[:])
	if err != nil {
		return fmt.Errorf("failed to generate random key: %v", err)
	}
	return nil
}

// saveKey сохраняет ключ в файл
func (srm *StatelessResetManager) saveKey() error {
	if srm.keyFile == "" {
		return fmt.Errorf("key file path is empty")
	}

	// Создаем директорию если не существует
	dir := filepath.Dir(srm.keyFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %v", err)
	}

	// Конвертируем ключ в hex строку
	keyHex := hex.EncodeToString(srm.key[:]) + "\n"

	// Записываем ключ в файл с ограниченными правами доступа
	if err := os.WriteFile(srm.keyFile, []byte(keyHex), 0600); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}

	return nil
}

// GetKey возвращает текущий ключ
func (srm *StatelessResetManager) GetKey() *quic.StatelessResetKey {
	srm.mutex.RLock()
	defer srm.mutex.RUnlock()

	// Возвращаем копию ключа
	key := srm.key
	return &key
}

// SetTransport устанавливает QUIC Transport для использования ключа
func (srm *StatelessResetManager) SetTransport(transport *quic.Transport) {
	srm.mutex.Lock()
	defer srm.mutex.Unlock()

	srm.transport = transport
	if transport != nil {
		transport.StatelessResetKey = srm.GetKey()
		srm.logger.Info("StatelessResetKey set on QUIC Transport")
	}
}

// HandleStatelessReset обрабатывает входящие stateless reset пакеты
func (srm *StatelessResetManager) HandleStatelessReset(data []byte) {
	srm.mutex.RLock()
	defer srm.mutex.RUnlock()

	// Проверяем, является ли пакет stateless reset
	if len(data) < 17 { // Минимальная длина stateless reset пакета
		return
	}

	// Stateless reset пакет начинается с 0x00 и содержит токен
	if data[0] != 0x00 {
		return
	}

	// Извлекаем токен (последние 16 байт)
	if len(data) < 16 {
		return
	}

	token := data[len(data)-16:]

	srm.logger.Debug("Received stateless reset packet",
		zap.String("token_hex", hex.EncodeToString(token)),
		zap.Int("packet_length", len(data)))

	// Здесь можно добавить логику для обработки stateless reset
	// Например, закрытие соответствующих соединений
}

// RotateKey генерирует новый ключ и обновляет Transport
func (srm *StatelessResetManager) RotateKey() error {
	srm.mutex.Lock()
	defer srm.mutex.Unlock()

	srm.logger.Info("Rotating StatelessResetKey")

	// Генерируем новый ключ
	if err := srm.generateKey(); err != nil {
		return fmt.Errorf("failed to generate new key: %v", err)
	}

	// Сохраняем новый ключ
	if err := srm.saveKey(); err != nil {
		srm.logger.Warn("Failed to save new key", zap.Error(err))
	}

	// Обновляем Transport если он установлен
	if srm.transport != nil {
		srm.transport.StatelessResetKey = srm.GetKey()
		srm.logger.Info("StatelessResetKey rotated and updated on Transport")
	}

	return nil
}

// GetKeyInfo возвращает информацию о ключе
func (srm *StatelessResetManager) GetKeyInfo() map[string]interface{} {
	srm.mutex.RLock()
	defer srm.mutex.RUnlock()

	return map[string]interface{}{
		"key_file":      srm.keyFile,
		"key_hex":       hex.EncodeToString(srm.key[:]),
		"key_length":    len(srm.key),
		"has_file":      srm.keyFile != "",
		"transport_set": srm.transport != nil,
	}
}

// ValidateKey проверяет валидность ключа
func (srm *StatelessResetManager) ValidateKey() error {
	srm.mutex.RLock()
	defer srm.mutex.RUnlock()

	// Проверяем, что ключ не пустой
	emptyKey := quic.StatelessResetKey{}
	if srm.key == emptyKey {
		return fmt.Errorf("key is empty")
	}

	// Проверяем, что ключ не состоит из одних нулей
	allZero := true
	for _, b := range srm.key {
		if b != 0 {
			allZero = false
			break
		}
	}

	if allZero {
		return fmt.Errorf("key consists of all zeros")
	}

	return nil
}
