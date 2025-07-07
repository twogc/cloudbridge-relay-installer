import React, { useEffect, useState } from 'react';
import { useAuth } from './AuthProvider';
import { adminApi, type ApiStatus, type AuditLogs } from '../api/adminApi';
import './Dashboard.css';

export const Dashboard: React.FC = () => {
  const { user, isAuthenticated, isAdmin, logout } = useAuth();
  const [status, setStatus] = useState<ApiStatus | null>(null);
  const [auditLogs, setAuditLogs] = useState<AuditLogs | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'audit' | 'settings'>('overview');

  useEffect(() => {
    if (isAuthenticated && isAdmin) {
      loadDashboardData();
    }
  }, [isAuthenticated, isAdmin]);

  const loadDashboardData = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const [statusData, auditData] = await Promise.all([
        adminApi.getStatus(),
        adminApi.getAuditLogs(),
      ]);
      
      setStatus(statusData);
      setAuditLogs(auditData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка загрузки данных');
    } finally {
      setLoading(false);
    }
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString('ru-RU');
  };

  if (!isAuthenticated) {
    return (
      <div className="dashboard-container">
        <div className="auth-required">
          <h2>Требуется авторизация</h2>
          <p>Для доступа к дашборду необходимо войти в систему.</p>
        </div>
      </div>
    );
  }

  if (!isAdmin) {
    return (
      <div className="dashboard-container">
        <div className="access-denied">
          <h2>Доступ запрещен</h2>
          <p>У вас недостаточно прав для доступа к административной панели.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="dashboard-container">
      {/* Header */}
      <div className="dashboard-header">
        <div className="header-left">
          <h1>Cloudbridge Relay Admin Dashboard</h1>
          <p>Добро пожаловать, {user?.name}</p>
        </div>
        <div className="header-right">
          <button onClick={logout} className="logout-btn">
            Выйти
          </button>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="dashboard-tabs">
        <button
          className={`tab-btn ${activeTab === 'overview' ? 'active' : ''}`}
          onClick={() => setActiveTab('overview')}
        >
          Обзор
        </button>
        <button
          className={`tab-btn ${activeTab === 'audit' ? 'active' : ''}`}
          onClick={() => setActiveTab('audit')}
        >
          Аудит
        </button>
        <button
          className={`tab-btn ${activeTab === 'settings' ? 'active' : ''}`}
          onClick={() => setActiveTab('settings')}
        >
          Настройки
        </button>
      </div>

      {/* Content */}
      <div className="dashboard-content">
        {loading && (
          <div className="loading">
            <div className="spinner"></div>
            <p>Загрузка данных...</p>
          </div>
        )}

        {error && (
          <div className="error-message">
            <h3>Ошибка</h3>
            <p>{error}</p>
            <button onClick={loadDashboardData} className="retry-btn">
              Повторить
            </button>
          </div>
        )}

        {!loading && !error && (
          <>
            {activeTab === 'overview' && (
              <div className="overview-tab">
                <div className="status-cards">
                  <div className="status-card">
                    <h3>Статус API</h3>
                    <div className={`status-indicator ${status?.status === 'ok' ? 'success' : 'error'}`}>
                      {status?.status === 'ok' ? 'Работает' : 'Ошибка'}
                    </div>
                    <p>{status?.message}</p>
                  </div>
                  
                  <div className="status-card">
                    <h3>Пользователь</h3>
                    <p><strong>Имя:</strong> {user?.name}</p>
                    <p><strong>Email:</strong> {user?.email}</p>
                    <p><strong>Роли:</strong> {user?.roles.join(', ')}</p>
                  </div>
                </div>
              </div>
            )}

            {activeTab === 'audit' && (
              <div className="audit-tab">
                <h3>Аудит логи</h3>
                {auditLogs && auditLogs.events.length > 0 ? (
                  <div className="audit-table">
                    <table>
                      <thead>
                        <tr>
                          <th>Время</th>
                          <th>Событие</th>
                          <th>Пользователь</th>
                          <th>Детали</th>
                        </tr>
                      </thead>
                      <tbody>
                        {auditLogs.events.map((event, index) => (
                          <tr key={index}>
                            <td>{formatTimestamp(event.timestamp)}</td>
                            <td>{event.event}</td>
                            <td>{event.user_id}</td>
                            <td>
                              <pre>{JSON.stringify(event.details, null, 2)}</pre>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                ) : (
                  <p>Аудит логи отсутствуют</p>
                )}
              </div>
            )}

            {activeTab === 'settings' && (
              <div className="settings-tab">
                <h3>Настройки</h3>
                <p>Настройки будут добавлены в будущих версиях.</p>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}; 