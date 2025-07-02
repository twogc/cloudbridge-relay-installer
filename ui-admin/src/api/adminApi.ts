import { getAccessToken } from '../config/keycloak';

// Базовый URL для admin API
const API_BASE_URL = 'http://localhost:8083';

// Типы для API ответов
export interface ApiStatus {
  status: string;
  message: string;
}

export interface AuditEvent {
  timestamp: string;
  event: string;
  user_id: string;
  details: any;
}

export interface AuditLogs {
  events: AuditEvent[];
}

// Функция для создания заголовков с авторизацией
const createAuthHeaders = async (): Promise<HeadersInit> => {
  const token = await getAccessToken();
  return {
    'Content-Type': 'application/json',
    ...(token && { Authorization: `Bearer ${token}` }),
  };
};

// Функция для выполнения API запросов
const apiRequest = async <T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> => {
  const headers = await createAuthHeaders();
  
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      ...headers,
      ...options.headers,
    },
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new Error('Unauthorized - Please login');
    }
    if (response.status === 403) {
      throw new Error('Forbidden - Insufficient permissions');
    }
    throw new Error(`API Error: ${response.status} ${response.statusText}`);
  }

  return response.json();
};

// API функции
export const adminApi = {
  // Получить статус API
  getStatus: (): Promise<ApiStatus> => {
    return apiRequest<ApiStatus>('/api/status');
  },

  // Получить аудит логи
  getAuditLogs: (): Promise<AuditLogs> => {
    return apiRequest<AuditLogs>('/api/audit/logs');
  },

  // Проверить доступность API
  checkHealth: async (): Promise<boolean> => {
    try {
      await apiRequest('/api/status');
      return true;
    } catch (error) {
      console.error('Health check failed:', error);
      return false;
    }
  },
}; 