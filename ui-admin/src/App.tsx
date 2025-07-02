import React from 'react';
import { AuthProvider } from './components/AuthProvider';
import { Dashboard } from './components/Dashboard';
import { LoginPage } from './components/LoginPage';
import { useAuth } from './components/AuthProvider';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Callback from './pages/Callback';
import './App.css';

// Компонент для условного рендеринга
const AppContent: React.FC = () => {
  const { isAuthenticated, isAdmin, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="app-loading">
        <div className="spinner"></div>
        <p>Инициализация приложения...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <LoginPage />;
  }

  if (!isAdmin) {
    return (
      <div className="access-denied-page">
        <div className="access-denied-card">
          <h2>Доступ запрещен</h2>
          <p>У вас недостаточно прав для доступа к административной панели.</p>
          <p>Обратитесь к администратору системы.</p>
        </div>
      </div>
    );
  }

  return <Dashboard />;
};

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/callback" element={<Callback />} />
          <Route path="/*" element={<AppContent />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
