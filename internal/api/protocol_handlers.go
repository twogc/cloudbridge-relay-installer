package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/keycloak"
	"go.uber.org/zap"
)

// QuickConnectionRequest запрос на создание Quick соединения
type QuickConnectionRequest struct {
	TargetServer   string            `json:"target_server"`
	ConnectionType string            `json:"connection_type"`
	ClientConfig   QuickClientConfig `json:"client_config"`
}

// QuickClientConfig конфигурация клиента для Quick протокола
type QuickClientConfig struct {
	Endpoint  string `json:"endpoint"`
	PublicKey string `json:"public_key"`
}

// QuickConnectionResponse ответ на создание Quick соединения
type QuickConnectionResponse struct {
	Status       string      `json:"status"`
	ConnectionID string      `json:"connection_id"`
	Endpoint     string      `json:"endpoint"`
	Config       interface{} `json:"config"`
}

// P2PMeshRequest запрос на создание P2P Mesh соединения
type P2PMeshRequest struct {
	MeshID       string          `json:"mesh_id"`
	ServerConfig P2PServerConfig `json:"server_config"`
}

// P2PServerConfig конфигурация сервера для P2P протокола
type P2PServerConfig struct {
	Endpoint   string   `json:"endpoint"`
	PublicKey  string   `json:"public_key"`
	AllowedIPs []string `json:"allowed_ips"`
}

// P2PMeshResponse ответ на создание P2P Mesh соединения
type P2PMeshResponse struct {
	Status string         `json:"status"`
	MeshID string         `json:"mesh_id"`
	Peers  []ProtocolPeer `json:"peers"`
	Config interface{}    `json:"config"`
}

// ProtocolPeer информация о пире в mesh сети
type ProtocolPeer struct {
	ID        string `json:"id"`
	Endpoint  string `json:"endpoint"`
	PublicKey string `json:"public_key"`
}

// ProtocolHandler обработчик протоколов
type ProtocolHandler struct {
	keycloakValidator *keycloak.KeycloakValidator
	logger            *zap.Logger
	peerManager       *PeerManager
	relayManager      *RelayManager
}

// NewProtocolHandler создает новый обработчик протоколов
func NewProtocolHandler(
	keycloakValidator *keycloak.KeycloakValidator,
	logger *zap.Logger,
	peerManager *PeerManager,
	relayManager *RelayManager,
) *ProtocolHandler {
	return &ProtocolHandler{
		keycloakValidator: keycloakValidator,
		logger:            logger,
		peerManager:       peerManager,
		relayManager:      relayManager,
	}
}

// RegisterClient обрабатывает регистрацию клиента
func (ph *ProtocolHandler) RegisterClient(w http.ResponseWriter, r *http.Request) {
	// Определяем протокол по URL path
	if strings.Contains(r.URL.Path, "/quick/") {
		ph.handleQuickConnection(w, r)
	} else if strings.Contains(r.URL.Path, "/p2p/") {
		ph.handleP2PMeshConnection(w, r)
	} else {
		ph.logger.Warn("Unknown endpoint", zap.String("path", r.URL.Path))
		http.Error(w, "Unknown endpoint", http.StatusNotFound)
		return
	}
}

// handleQuickConnection обрабатывает Quick соединения
func (ph *ProtocolHandler) handleQuickConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлечение токена из Authorization header
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		ph.logger.Warn("Missing or invalid authorization header")
		http.Error(w, "Missing or invalid authorization header", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Валидация Quick токена
	claims, err := ph.keycloakValidator.ValidateQuickToken(token)
	if err != nil {
		ph.logger.Warn("Quick token validation failed", zap.Error(err))
		http.Error(w, "Invalid quick token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var req QuickConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ph.logger.Warn("Invalid request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация запроса
	if req.TargetServer == "" {
		http.Error(w, "target_server is required", http.StatusBadRequest)
		return
	}

	// Проверка лимитов
	maxConnections, _ := strconv.Atoi(claims.MaxConnections)
	currentConnections := ph.getCurrentConnections(claims.TenantID, claims.ServerID)

	if currentConnections >= maxConnections {
		ph.logger.Warn("Connection limit exceeded",
			zap.String("tenant_id", claims.TenantID),
			zap.String("server_id", claims.ServerID),
			zap.Int("current", currentConnections),
			zap.Int("max", maxConnections),
		)
		http.Error(w, "Connection limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Создание Quick соединения
	connection, err := ph.createQuickConnection(claims, req)
	if err != nil {
		ph.logger.Error("Failed to create quick connection", zap.Error(err))
		http.Error(w, "Failed to create connection", http.StatusInternalServerError)
		return
	}

	// Ответ клиенту
	response := QuickConnectionResponse{
		Status:       "success",
		ConnectionID: connection.ID,
		Endpoint:     connection.Endpoint,
		Config:       connection.Config,
	}

	ph.logger.Info("Quick connection created successfully",
		zap.String("connection_id", connection.ID),
		zap.String("tenant_id", claims.TenantID),
		zap.String("server_id", claims.ServerID),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleP2PMeshConnection обрабатывает P2P Mesh соединения
func (ph *ProtocolHandler) handleP2PMeshConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлечение токена из Authorization header
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		ph.logger.Warn("Missing or invalid authorization header")
		http.Error(w, "Missing or invalid authorization header", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Валидация P2P Mesh токена
	claims, err := ph.keycloakValidator.ValidateP2PMeshToken(token)
	if err != nil {
		ph.logger.Warn("P2P Mesh token validation failed", zap.Error(err))
		http.Error(w, "Invalid P2P mesh token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var req P2PMeshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ph.logger.Warn("Invalid request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация запроса
	if req.MeshID == "" {
		http.Error(w, "mesh_id is required", http.StatusBadRequest)
		return
	}

	// Проверка лимитов
	maxPeers, _ := strconv.Atoi(claims.MaxPeers)
	currentPeers := ph.getCurrentPeers(claims.TenantID, claims.ServerID, req.MeshID)

	if currentPeers >= maxPeers {
		ph.logger.Warn("Peer limit exceeded",
			zap.String("tenant_id", claims.TenantID),
			zap.String("server_id", claims.ServerID),
			zap.String("mesh_id", req.MeshID),
			zap.Int("current", currentPeers),
			zap.Int("max", maxPeers),
		)
		http.Error(w, "Peer limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Создание P2P Mesh соединения
	mesh, err := ph.createP2PMeshConnection(claims, req)
	if err != nil {
		ph.logger.Error("Failed to create P2P mesh connection", zap.Error(err))
		http.Error(w, "Failed to create mesh connection", http.StatusInternalServerError)
		return
	}

	// Ответ клиенту
	response := P2PMeshResponse{
		Status: "success",
		MeshID: mesh.ID,
		Peers:  mesh.Peers,
		Config: mesh.Config,
	}

	ph.logger.Info("P2P Mesh connection created successfully",
		zap.String("mesh_id", mesh.ID),
		zap.String("tenant_id", claims.TenantID),
		zap.String("server_id", claims.ServerID),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getCurrentConnections возвращает текущее количество соединений
func (ph *ProtocolHandler) getCurrentConnections(tenantID, serverID string) int {
	// TODO: Реализовать получение текущего количества соединений из базы данных
	// Пока возвращаем 0 для тестирования
	return 0
}

// getCurrentPeers возвращает текущее количество пиров в mesh
func (ph *ProtocolHandler) getCurrentPeers(tenantID, serverID, meshID string) int {
	// TODO: Реализовать получение текущего количества пиров из базы данных
	// Пока возвращаем 0 для тестирования
	return 0
}

// createQuickConnection создает Quick соединение
func (ph *ProtocolHandler) createQuickConnection(claims *keycloak.QuickTokenClaims, req QuickConnectionRequest) (*QuickConnection, error) {
	// TODO: Реализовать создание Quick соединения
	// Пока возвращаем mock объект
	return &QuickConnection{
		ID:       fmt.Sprintf("quick-%s-%d", claims.TenantID, time.Now().Unix()),
		Endpoint: "192.168.1.100:8080",
		Config: map[string]interface{}{
			"protocol": "quick",
			"target":   req.TargetServer,
		},
	}, nil
}

// createP2PMeshConnection создает P2P Mesh соединение
func (ph *ProtocolHandler) createP2PMeshConnection(claims *keycloak.P2PMeshTokenClaims, req P2PMeshRequest) (*P2PMesh, error) {
	// TODO: Реализовать создание P2P Mesh соединения
	// Пока возвращаем mock объект
	return &P2PMesh{
		ID: req.MeshID,
		Peers: []ProtocolPeer{
			{
				ID:        "peer-1",
				Endpoint:  "192.168.1.101:51820",
				PublicKey: "peer1-public-key",
			},
		},
		Config: map[string]interface{}{
			"protocol": "p2p-mesh",
			"mesh_id":  req.MeshID,
		},
	}, nil
}

// QuickConnection структура Quick соединения
type QuickConnection struct {
	ID       string      `json:"id"`
	Endpoint string      `json:"endpoint"`
	Config   interface{} `json:"config"`
}

// P2PMesh структура P2P Mesh соединения
type P2PMesh struct {
	ID     string         `json:"id"`
	Peers  []ProtocolPeer `json:"peers"`
	Config interface{}    `json:"config"`
}
