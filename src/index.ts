/**
 * CloudBridge Relay - Enterprise Edition
 * 
 * Enterprise-grade high-performance relay server with WireGuard integration
 * 
 * @package @twogc/cloudbridge-relay
 * @version 1.0.0
 * @author 2GC Team <dev@2gc.ru>
 * @license MIT
 */

export * from './types';
export * from './client';
export * from './server';
export * from './wireguard';
export * from './quic';
export * from './metrics';
export * from './utils';

// Основные экспорты
export { CloudBridgeRelay } from './client';
export { RelayServer } from './server';
export { WireGuardManager } from './wireguard';
export { QUICManager } from './quic';
export { MetricsCollector } from './metrics';

// Утилиты
export { 
  generateKeyPair, 
  validateConfig, 
  createLogger,
  healthCheck 
} from './utils';

// Константы
export const VERSION = '1.0.0';
export const DEFAULT_CONFIG = {
  server: {
    port: 8080,
    host: '0.0.0.0'
  },
  wireguard: {
    enabled: true,
    interface: 'wg-mesh0',
    port: 51820
  },
  quic: {
    enabled: true,
    enhanced: true
  },
  metrics: {
    enabled: true,
    port: 9090
  }
};

// Типы конфигурации
export interface CloudBridgeConfig {
  server: {
    port: number;
    host: string;
    timeout?: number;
  };
  wireguard: {
    enabled: boolean;
    interface: string;
    port: number;
    privateKey?: string;
    publicKey?: string;
  };
  quic: {
    enabled: boolean;
    enhanced: boolean;
    maxStreams?: number;
  };
  metrics: {
    enabled: boolean;
    port: number;
    path?: string;
  };
  security?: {
    rateLimit?: number;
    jwtSecret?: string;
  };
}

// Основной класс для быстрого старта
export class CloudBridgeRelayQuickStart {
  private config: CloudBridgeConfig;
  private client?: CloudBridgeRelay;
  private server?: RelayServer;

  constructor(config: Partial<CloudBridgeConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  /**
   * Запуск клиента
   */
  async startClient(): Promise<CloudBridgeRelay> {
    this.client = new CloudBridgeRelay(this.config);
    await this.client.connect();
    return this.client;
  }

  /**
   * Запуск сервера
   */
  async startServer(): Promise<RelayServer> {
    this.server = new RelayServer(this.config);
    await this.server.start();
    return this.server;
  }

  /**
   * Остановка всех компонентов
   */
  async stop(): Promise<void> {
    if (this.client) {
      await this.client.disconnect();
    }
    if (this.server) {
      await this.server.stop();
    }
  }

  /**
   * Получение статуса
   */
  getStatus(): { client: boolean; server: boolean } {
    return {
      client: this.client?.isConnected() ?? false,
      server: this.server?.isRunning() ?? false
    };
  }
}

// Экспорт по умолчанию
export default CloudBridgeRelayQuickStart;
