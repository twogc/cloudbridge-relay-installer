package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/twogc/cloudbridge-relay-installer/internal/p2p"
	"go.uber.org/zap"
)

// Используем моки из mocks.go

// TestPeerStore тесты для PeerStore
func TestNewPeerStore(t *testing.T) {
	store := NewPeerStore()

	assert.NotNil(t, store)
	assert.NotNil(t, store.peers)
	assert.Empty(t, store.peers)
}

func TestPeerStore_Store(t *testing.T) {
	store := NewPeerStore()
	peer := &p2p.Peer{
		ID:       "peer1",
		TenantID: "tenant1",
	}

	store.Store("tenant1", "peer1", peer)

	// Проверяем, что пир сохранен
	retrievedPeer, exists := store.Get("tenant1", "peer1")
	assert.True(t, exists)
	assert.Equal(t, peer, retrievedPeer)
}

func TestPeerStore_Get(t *testing.T) {
	store := NewPeerStore()
	peer := &p2p.Peer{
		ID:       "peer1",
		TenantID: "tenant1",
	}

	// Тест получения несуществующего пира
	_, exists := store.Get("tenant1", "nonexistent")
	assert.False(t, exists)

	// Сохраняем пира
	store.Store("tenant1", "peer1", peer)

	// Тест получения существующего пира
	retrievedPeer, exists := store.Get("tenant1", "peer1")
	assert.True(t, exists)
	assert.Equal(t, peer, retrievedPeer)
}

func TestPeerStore_Update(t *testing.T) {
	store := NewPeerStore()
	peer := &p2p.Peer{
		ID:       "peer1",
		TenantID: "tenant1",
	}
	updatedPeer := &p2p.Peer{
		ID:       "peer1",
		TenantID: "tenant1",
		IsOnline: true,
	}

	// Сохраняем пира
	store.Store("tenant1", "peer1", peer)

	// Обновляем пира
	store.Update("tenant1", "peer1", updatedPeer)

	// Проверяем обновление
	retrievedPeer, exists := store.Get("tenant1", "peer1")
	assert.True(t, exists)
	assert.Equal(t, updatedPeer, retrievedPeer)
}

func TestPeerStore_GetAll(t *testing.T) {
	store := NewPeerStore()
	peer1 := &p2p.Peer{ID: "peer1", TenantID: "tenant1"}
	peer2 := &p2p.Peer{ID: "peer2", TenantID: "tenant1"}

	// Тест получения пиров для несуществующего тенанта
	peers := store.GetAll("nonexistent")
	assert.Empty(t, peers)

	// Сохраняем пиров
	store.Store("tenant1", "peer1", peer1)
	store.Store("tenant1", "peer2", peer2)

	// Получаем всех пиров
	peers = store.GetAll("tenant1")
	assert.Len(t, peers, 2)

	// Проверяем, что пиры присутствуют
	peerIDs := make(map[string]bool)
	for _, peer := range peers {
		peerIDs[peer.ID] = true
	}
	assert.True(t, peerIDs["peer1"])
	assert.True(t, peerIDs["peer2"])
}

func TestPeerStore_IsPeerOnline(t *testing.T) {
	store := NewPeerStore()
	peer := &p2p.Peer{
		ID:       "peer1",
		TenantID: "tenant1",
		LastSeen: time.Now().Add(-2 * time.Minute), // 2 минуты назад
	}

	// Тест для несуществующего пира
	assert.False(t, store.IsPeerOnline("nonexistent"))

	// Сохраняем пира
	store.Store("tenant1", "peer1", peer)

	// Тест для онлайн пира (менее 5 минут назад)
	assert.True(t, store.IsPeerOnline("peer1"))

	// Обновляем время последнего контакта на 10 минут назад
	peer.LastSeen = time.Now().Add(-10 * time.Minute)
	store.Update("tenant1", "peer1", peer)

	// Тест для офлайн пира
	assert.False(t, store.IsPeerOnline("peer1"))
}

func TestPeerStore_GetLastSeen(t *testing.T) {
	store := NewPeerStore()
	lastSeen := time.Now()
	peer := &p2p.Peer{
		ID:       "peer1",
		TenantID: "tenant1",
		LastSeen: lastSeen,
	}

	// Тест для несуществующего пира
	lastSeenTime := store.GetLastSeen("nonexistent")
	assert.True(t, lastSeenTime.IsZero())

	// Сохраняем пира
	store.Store("tenant1", "peer1", peer)

	// Тест для существующего пира
	lastSeenTime = store.GetLastSeen("peer1")
	assert.Equal(t, lastSeen, lastSeenTime)
}

// TestNewPeerManager тесты для PeerManager
func TestNewPeerManager(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}

	pm := NewPeerManager(logger, validator)

	assert.NotNil(t, pm)
	assert.NotNil(t, pm.peers)
	assert.NotNil(t, pm.connections)
	assert.NotNil(t, pm.logger)
	assert.NotNil(t, pm.validator)
	assert.NotNil(t, pm.ipValidator)
	assert.NotNil(t, pm.peerStore)
	assert.NotNil(t, pm.allocatedIPs)
}

func TestNewPeerManagerWithConfig(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	allowedIPs := []string{"192.168.1.0/24", "10.0.0.1"}

	pm := NewPeerManagerWithConfig(logger, validator, allowedIPs)

	assert.NotNil(t, pm)
	assert.NotNil(t, pm.peers)
	assert.NotNil(t, pm.connections)
	assert.NotNil(t, pm.logger)
	assert.NotNil(t, pm.validator)
	assert.NotNil(t, pm.ipValidator)
	assert.NotNil(t, pm.peerStore)
}

// TestPeerManager_allocateIPForTenant тесты для выделения IP адресов
func TestPeerManager_allocateIPForTenant(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Тест выделения IP для нового пира
	ip, err := pm.allocateIPForTenant("tenant1", "peer1")
	assert.NoError(t, err)
	assert.NotEmpty(t, ip)
	assert.Contains(t, ip, "10.100.77.")

	// Тест повторного выделения IP для того же пира
	ip2, err := pm.allocateIPForTenant("tenant1", "peer1")
	assert.NoError(t, err)
	assert.Equal(t, ip, ip2)

	// Тест выделения IP для другого пира
	ip3, err := pm.allocateIPForTenant("tenant1", "peer2")
	assert.NoError(t, err)
	assert.NotEqual(t, ip, ip3)
}

func TestPeerManager_deallocateIP(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Выделяем IP
	_, err := pm.allocateIPForTenant("tenant1", "peer1")
	assert.NoError(t, err)

	// Освобождаем IP
	pm.deallocateIP("peer1")

	// Проверяем, что IP освобожден
	_, exists := pm.allocatedIPs["peer1"]
	assert.False(t, exists)
}

// TestPeerManager_RegisterPeer тесты для регистрации пиров
func TestPeerManager_RegisterPeer_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Создаем тестовый запрос
	req := RelayAssistedRegistrationRequest{
		PublicKey:  "test-public-key",
		AllowedIPs: []string{"192.168.1.0/24"},
	}
	reqBody, _ := json.Marshal(req)

	// Создаем тестовый контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/tenants/tenant1/peers", bytes.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer test-token")
	c.Params = gin.Params{{Key: "tenant_id", Value: "tenant1"}}

	// Настраиваем мок валидатора
	claims := &P2PClaims{
		TenantID:       "tenant1",
		ServerID:       "server1",
		ConnectionType: "relay-assisted",
	}
	validator.On("ValidateP2PToken", mock.Anything).Return(claims, nil)

	// Выполняем запрос
	pm.RegisterPeer(c)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)

	var response RelayAssistedRegistrationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotEmpty(t, response.PeerID)
	assert.NotEmpty(t, response.RelaySessionID)
	assert.NotEmpty(t, response.AllocatedIP)

	validator.AssertExpectations(t)
}

func TestPeerManager_RegisterPeer_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Создаем тестовый контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/tenants/tenant1/peers", nil)
	c.Params = gin.Params{{Key: "tenant_id", Value: "tenant1"}}

	// Настраиваем мок валидатора для возврата ошибки
	validator.On("ValidateP2PToken", mock.Anything).Return(nil, assert.AnError)

	// Выполняем запрос
	pm.RegisterPeer(c)

	// Проверяем результат
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid token", response.Error)
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	validator.AssertExpectations(t)
}

func TestPeerManager_RegisterPeer_TenantMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Создаем тестовый контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/tenants/tenant1/peers", nil)
	c.Params = gin.Params{{Key: "tenant_id", Value: "tenant1"}}

	// Настраиваем мок валидатора с другим tenant_id
	claims := &P2PClaims{
		TenantID:       "tenant2", // Другой tenant
		ServerID:       "server1",
		ConnectionType: "relay-assisted",
	}
	validator.On("ValidateP2PToken", mock.Anything).Return(claims, nil)

	// Выполняем запрос
	pm.RegisterPeer(c)

	// Проверяем результат
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Tenant mismatch", response.Error)
	assert.Equal(t, http.StatusForbidden, response.Code)

	validator.AssertExpectations(t)
}

// TestPeerManager_DiscoverPeers тесты для обнаружения пиров
func TestPeerManager_DiscoverPeers_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Создаем тестовый контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/tenants/tenant1/peers", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	c.Request = req
	c.Params = gin.Params{{Key: "tenant_id", Value: "tenant1"}}

	// Настраиваем мок валидатора
	claims := &P2PClaims{
		TenantID: "tenant1",
		ServerID: "server1",
	}
	validator.On("ValidateP2PToken", mock.Anything).Return(claims, nil)

	// Выполняем запрос
	pm.DiscoverPeers(c)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)

	// Отладочная информация
	t.Logf("Response body: %s", w.Body.String())

	var response PeerDiscoveryResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	// Peers может быть nil для пустого списка
	assert.True(t, response.Success)
	assert.Equal(t, "tenant1", response.TenantID)

	validator.AssertExpectations(t)
}

func TestPeerManager_DiscoverPeers_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Создаем тестовый контекст без tenant_id
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/peers", nil)

	// Выполняем запрос
	pm.DiscoverPeers(c)

	// Проверяем результат
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Missing tenant_id", response.Error)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestPeerManager_DiscoverPeers_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Создаем тестовый контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/tenants/tenant1/peers", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	c.Request = req
	c.Params = gin.Params{{Key: "tenant_id", Value: "tenant1"}}

	// Настраиваем мок валидатора для возврата ошибки
	validator.On("ValidateP2PToken", mock.Anything).Return(nil, assert.AnError)

	// Выполняем запрос
	pm.DiscoverPeers(c)

	// Проверяем результат
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid token", response.Error)
	assert.Equal(t, http.StatusUnauthorized, response.Code)

	validator.AssertExpectations(t)
}

// TestPeerManager_UpdatePeerStatus тесты для обновления статуса пира
func TestPeerManager_UpdatePeerStatus_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Создаем тестовый запрос
	updateReq := PeerStatusUpdateRequest{
		IsOnline:   true,
		MeshStatus: "connected",
	}
	reqBody, _ := json.Marshal(updateReq)

	// Создаем тестовый контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	httpReq := httptest.NewRequest("PUT", "/tenants/tenant1/peers/peer1/status", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer test-token")
	c.Request = httpReq
	c.Params = gin.Params{
		{Key: "tenant_id", Value: "tenant1"},
		{Key: "peer_id", Value: "peer1"},
	}

	// Настраиваем мок валидатора
	claims := &P2PClaims{
		TenantID: "tenant1",
		ServerID: "server1",
	}
	validator.On("ValidateP2PToken", mock.Anything).Return(claims, nil)

	// Выполняем запрос
	pm.UpdatePeerStatus(c)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)

	var response PeerStatusUpdateResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	validator.AssertExpectations(t)
}

// TestPeerManager_GetPeerInfo тесты для получения информации о пире
func TestPeerManager_GetPeerInfo_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Создаем тестовый контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/tenants/tenant1/peers/peer1", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	c.Request = req
	c.Params = gin.Params{
		{Key: "tenant_id", Value: "tenant1"},
		{Key: "peer_id", Value: "peer1"},
	}

	// Настраиваем мок валидатора
	claims := &P2PClaims{
		TenantID: "tenant1",
		ServerID: "server1",
	}
	validator.On("ValidateP2PToken", mock.Anything).Return(claims, nil)

	// Выполняем запрос
	pm.GetPeerInfo(c)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)

	var response PeerInfoResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response.Peer)

	validator.AssertExpectations(t)
}

func TestPeerManager_GetPeerInfo_PeerNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Создаем тестовый контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/tenants/tenant1/peers/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	c.Request = req
	c.Params = gin.Params{
		{Key: "tenant_id", Value: "tenant1"},
		{Key: "peer_id", Value: "nonexistent"},
	}

	// Настраиваем мок валидатора
	claims := &P2PClaims{
		TenantID: "tenant1",
		ServerID: "server1",
	}
	validator.On("ValidateP2PToken", mock.Anything).Return(claims, nil)

	// Выполняем запрос
	pm.GetPeerInfo(c)

	// Проверяем результат
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Peer not found", response.Error)
	assert.Equal(t, http.StatusNotFound, response.Code)

	validator.AssertExpectations(t)
}

// TestPeerManager_RegisterPeerDirect тесты для прямой регистрации пира
func TestPeerManager_RegisterPeerDirect_Success(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Создаем тестовый запрос
	req := DirectPeerRegistrationRequest{
		PeerID:   "peer1",
		TenantID: "tenant1",
		DeviceID: "device1",
	}

	// Выполняем регистрацию
	response, err := pm.RegisterPeerDirectWithRequest(req)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Success)
	assert.Equal(t, "peer1", response.PeerID)
	assert.NotEmpty(t, response.AllocatedIP)
}

func TestPeerManager_RegisterPeerDirect_InvalidRequest(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Создаем невалидный запрос
	req := DirectPeerRegistrationRequest{
		PeerID:   "", // Пустой PeerID
		TenantID: "tenant1",
		DeviceID: "device1",
	}

	// Выполняем регистрацию
	response, err := pm.RegisterPeerDirectWithRequest(req)

	// Проверяем результат
	assert.Error(t, err)
	assert.Nil(t, response)
}

// TestPeerManager_ConcurrentAccess тесты для конкурентного доступа
func TestPeerManager_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Запускаем несколько горутин для конкурентного доступа
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			req := DirectPeerRegistrationRequest{
				PeerID:   fmt.Sprintf("peer%d", id),
				TenantID: "tenant1",
				DeviceID: fmt.Sprintf("device%d", id),
			}

			response, err := pm.RegisterPeerDirectWithRequest(req)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.True(t, response.Success)

			done <- true
		}(i)
	}

	// Ждем завершения всех горутин
	for i := 0; i < 10; i++ {
		<-done
	}

	// Проверяем, что все пиры зарегистрированы
	peers := pm.peerStore.GetAll("tenant1")
	assert.Len(t, peers, 10)
}

// TestPeerManager_ErrorHandling тесты для обработки ошибок
func TestPeerManager_ErrorHandling(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Тест с невалидными данными
	req := DirectPeerRegistrationRequest{
		PeerID:   "peer1",
		TenantID: "", // Пустой TenantID
		DeviceID: "device1",
	}

	response, err := pm.RegisterPeerDirectWithRequest(req)
	assert.Error(t, err)
	assert.Nil(t, response)
}

// TestPeerManager_Integration тесты интеграции
func TestPeerManager_Integration(t *testing.T) {
	logger := zap.NewNop()
	validator := &MockTokenValidator{}
	pm := NewPeerManager(logger, validator)

	// Регистрируем пира
	req := DirectPeerRegistrationRequest{
		PeerID:   "peer1",
		TenantID: "tenant1",
		DeviceID: "device1",
	}

	response, err := pm.RegisterPeerDirectWithRequest(req)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	// Проверяем, что пир сохранен
	peer, exists := pm.peerStore.Get("tenant1", "peer1")
	assert.True(t, exists)
	assert.NotNil(t, peer)
	assert.Equal(t, "peer1", peer.ID)
	assert.Equal(t, "tenant1", peer.TenantID)

	// Проверяем, что IP выделен
	assert.NotEmpty(t, peer.AllocatedIP)

	// Проверяем, что пир в списке всех пиров
	peers := pm.peerStore.GetAll("tenant1")
	assert.Len(t, peers, 1)
	assert.Equal(t, "peer1", peers[0].ID)
}
