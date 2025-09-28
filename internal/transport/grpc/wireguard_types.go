package grpc

// Временные типы для WireGuard gRPC сервиса
// TODO: Заменить на сгенерированные из protobuf

// CreateWireGuardTunnelRequest запрос на создание туннеля
type CreateWireGuardTunnelRequest struct {
	TenantId  string            `json:"tenant_id"`
	RequestId string            `json:"request_id"`
	Metadata  map[string]string `json:"metadata"`
}

// CreateWireGuardTunnelResponse ответ на создание туннеля
type CreateWireGuardTunnelResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	RequestId    string `json:"request_id"`
	PublicKey    string `json:"public_key"`
	Endpoint     string `json:"endpoint"`
	Port         int32  `json:"port"`
	ClientConfig string `json:"client_config"`
}

// DeleteWireGuardTunnelRequest запрос на удаление туннеля
type DeleteWireGuardTunnelRequest struct {
	TenantId  string `json:"tenant_id"`
	RequestId string `json:"request_id"`
}

// DeleteWireGuardTunnelResponse ответ на удаление туннеля
type DeleteWireGuardTunnelResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
}

// GetWireGuardTunnelRequest запрос на получение информации о туннеле
type GetWireGuardTunnelRequest struct {
	TenantId  string `json:"tenant_id"`
	RequestId string `json:"request_id"`
}

// GetWireGuardTunnelResponse ответ с информацией о туннеле
type GetWireGuardTunnelResponse struct {
	Success   bool                 `json:"success"`
	Message   string               `json:"message"`
	RequestId string               `json:"request_id"`
	Tunnel    *WireGuardTunnelInfo `json:"tunnel"`
}

// ListWireGuardTunnelsRequest запрос на получение списка туннелей
type ListWireGuardTunnelsRequest struct {
	RequestId string `json:"request_id"`
	Limit     int32  `json:"limit"`
	Offset    int32  `json:"offset"`
}

// ListWireGuardTunnelsResponse ответ со списком туннелей
type ListWireGuardTunnelsResponse struct {
	Success   bool                 `json:"success"`
	Message   string               `json:"message"`
	RequestId string               `json:"request_id"`
	Tunnels   []*WireGuardTunnelInfo `json:"tunnels"`
}

// GetWireGuardStatusRequest запрос на получение статуса
type GetWireGuardStatusRequest struct {
	RequestId string `json:"request_id"`
}

// GetWireGuardStatusResponse ответ со статусом
type GetWireGuardStatusResponse struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message"`
	RequestId string            `json:"request_id"`
	Status    map[string]interface{} `json:"status"`
}

// GetWireGuardMetricsRequest запрос на получение метрик
type GetWireGuardMetricsRequest struct {
	RequestId string `json:"request_id"`
}

// GetWireGuardMetricsResponse ответ с метриками
type GetWireGuardMetricsResponse struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message"`
	RequestId string            `json:"request_id"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// UpdateWireGuardConfigRequest запрос на обновление конфигурации
type UpdateWireGuardConfigRequest struct {
	RequestId   string `json:"request_id"`
	ForceReload bool   `json:"force_reload"`
}

// UpdateWireGuardConfigResponse ответ на обновление конфигурации
type UpdateWireGuardConfigResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
}

// WireGuardTunnelInfo информация о туннеле
type WireGuardTunnelInfo struct {
	TenantId      string `json:"tenant_id"`
	PublicKey     string `json:"public_key"`
	Endpoint      string `json:"endpoint"`
	Port          int32  `json:"port"`
	IsActive      bool   `json:"is_active"`
	BytesReceived int64  `json:"bytes_received"`
	BytesSent     int64  `json:"bytes_sent"`
	LastConnected int64  `json:"last_connected"`
}

