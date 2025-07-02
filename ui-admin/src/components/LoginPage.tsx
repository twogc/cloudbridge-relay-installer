import React from 'react';
import { useAuth } from './AuthProvider';
import './LoginPage.css';

export const LoginPage: React.FC = () => {
  const { login, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="login-container">
        <div className="loading">
          <div className="spinner"></div>
          <p>Загрузка...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="login-container">
      <div className="login-card">
        <div className="login-header">
          <h1>Cloudbridge Relay</h1>
          <h2>Административная панель</h2>
        </div>
        
        <div className="login-content">
          <p>Для доступа к административной панели необходимо войти в систему.</p>
          
          <button onClick={login} className="login-btn">
            <svg className="login-icon" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
            </svg>
            Войти через Keycloak
          </button>
        </div>
        
        <div className="login-footer">
          <p>Система защищена аутентификацией Keycloak</p>
        </div>
      </div>
    </div>
  );
}; 