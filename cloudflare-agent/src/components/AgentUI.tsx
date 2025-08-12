import React, { useState, useEffect, useRef } from 'react';
import { useChat } from 'ai/react';

interface AgentUIProps {
  agentUrl?: string;
  title?: string;
}

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: string;
}

interface AgentStatus {
  isRunning: boolean;
  totalChecks: number;
  anomaliesDetected: number;
  lastCheck: string | null;
}

export const AgentUI: React.FC<AgentUIProps> = ({ 
  agentUrl = 'http://localhost:8787',
  title = 'CloudBridge AI Security Agent'
}) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [agentStatus, setAgentStatus] = useState<AgentStatus | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  // Используем AI SDK useChat hook для обработки чата
  const { messages: aiMessages, input, handleInputChange, handleSubmit, isLoading: aiLoading } = useChat({
    api: `${agentUrl}/chat`,
    onError: (error) => {
      setError(`AI Chat Error: ${error.message}`);
    }
  });

  // Автоскролл к последнему сообщению
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Получение статуса агента
  const fetchAgentStatus = async () => {
    try {
      const response = await fetch(`${agentUrl}/metrics`);
      if (response.ok) {
        const status = await response.json();
        setAgentStatus(status);
      }
    } catch (error) {
      console.error('Ошибка получения статуса:', error);
    }
  };

  // Выполнение проверки безопасности
  const performSecurityCheck = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`${agentUrl}/check`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        }
      });

      if (response.ok) {
        const result = await response.json();
        addMessage('assistant', result.result);
        await fetchAgentStatus();
      } else {
        throw new Error(`HTTP ${response.status}`);
      }
    } catch (error) {
      setError(`Ошибка проверки: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setIsLoading(false);
    }
  };

  // Добавление сообщения в чат
  const addMessage = (role: 'user' | 'assistant', content: string) => {
    const newMessage: Message = {
      id: Date.now().toString(),
      role,
      content,
      timestamp: new Date().toISOString()
    };
    setMessages(prev => [...prev, newMessage]);
  };

  // Обработка отправки сообщения
  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputValue.trim()) return;

    const userMessage = inputValue.trim();
    setInputValue('');
    addMessage('user', userMessage);

    try {
      const response = await fetch(`${agentUrl}/chat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ message: userMessage })
      });

      if (response.ok) {
        const result = await response.json();
        addMessage('assistant', result.response);
      } else {
        throw new Error(`HTTP ${response.status}`);
      }
    } catch (error) {
      setError(`Ошибка отправки сообщения: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  // WebSocket соединение для real-time чата
  useEffect(() => {
    const connectWebSocket = () => {
      try {
        const ws = new WebSocket(`${agentUrl.replace('http', 'ws')}/ws`);
        
        ws.onopen = () => {
          setIsConnected(true);
          setError(null);
        };

        ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            if (data.type === 'response') {
              addMessage('assistant', data.data);
            }
          } catch (error) {
            console.error('Ошибка парсинга WebSocket сообщения:', error);
          }
        };

        ws.onclose = () => {
          setIsConnected(false);
        };

        ws.onerror = (error) => {
          setError(`WebSocket ошибка: ${error}`);
          setIsConnected(false);
        };

        wsRef.current = ws;
      } catch (error) {
        setError(`Ошибка подключения WebSocket: ${error}`);
      }
    };

    connectWebSocket();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [agentUrl]);

  // Периодическое обновление статуса
  useEffect(() => {
    fetchAgentStatus();
    const interval = setInterval(fetchAgentStatus, 30000); // каждые 30 секунд
    
    return () => clearInterval(interval);
  }, [agentUrl]);

  return (
    <div className="agent-ui" style={styles.container}>
      {/* Заголовок */}
      <div style={styles.header}>
        <h1 style={styles.title}>{title}</h1>
        <div style={styles.statusIndicator}>
          <div 
            style={{
              ...styles.statusDot,
              backgroundColor: isConnected ? '#4CAF50' : '#f44336'
            }}
          />
          <span style={styles.statusText}>
            {isConnected ? 'Подключен' : 'Отключен'}
          </span>
        </div>
      </div>

      {/* Статус агента */}
      {agentStatus && (
        <div style={styles.statusPanel}>
          <div style={styles.statusItem}>
            <span style={styles.statusLabel}>Состояние:</span>
            <span style={{
              ...styles.statusValue,
              color: agentStatus.isRunning ? '#4CAF50' : '#f44336'
            }}>
              {agentStatus.isRunning ? '🟢 Активен' : '🔴 Остановлен'}
            </span>
          </div>
          <div style={styles.statusItem}>
            <span style={styles.statusLabel}>Проверок:</span>
            <span style={styles.statusValue}>{agentStatus.totalChecks}</span>
          </div>
          <div style={styles.statusItem}>
            <span style={styles.statusLabel}>Аномалий:</span>
            <span style={styles.statusValue}>{agentStatus.anomaliesDetected}</span>
          </div>
          <div style={styles.statusItem}>
            <span style={styles.statusLabel}>Последняя проверка:</span>
            <span style={styles.statusValue}>
              {agentStatus.lastCheck ? 
                new Date(agentStatus.lastCheck).toLocaleString() : 
                'Нет'
              }
            </span>
          </div>
        </div>
      )}

      {/* Панель управления */}
      <div style={styles.controls}>
        <button 
          onClick={performSecurityCheck}
          disabled={isLoading}
          style={{
            ...styles.button,
            ...styles.primaryButton,
            opacity: isLoading ? 0.6 : 1
          }}
        >
          {isLoading ? 'Проверка...' : '🔍 Проверить безопасность'}
        </button>
        
        <button 
          onClick={() => addMessage('user', '/status')}
          style={styles.button}
        >
          📊 Статус
        </button>
        
        <button 
          onClick={() => addMessage('user', '/history')}
          style={styles.button}
        >
          📝 История
        </button>
      </div>

      {/* Ошибки */}
      {error && (
        <div style={styles.error}>
          <span style={styles.errorText}>⚠️ {error}</span>
          <button 
            onClick={() => setError(null)}
            style={styles.errorClose}
          >
            ✕
          </button>
        </div>
      )}

      {/* Чат */}
      <div style={styles.chatContainer}>
        <div style={styles.messages}>
          {messages.length === 0 ? (
            <div style={styles.welcome}>
              <h3>🤖 Добро пожаловать в AI Security Agent!</h3>
              <p>Этот агент поможет вам мониторить безопасность CloudBridge Relay.</p>
              <p>Доступные команды:</p>
              <ul>
                <li><code>/status</code> - получить статус системы</li>
                <li><code>/check</code> - выполнить проверку безопасности</li>
                <li><code>/start</code> - запустить мониторинг</li>
                <li><code>/stop</code> - остановить мониторинг</li>
                <li><code>/history</code> - показать историю действий</li>
              </ul>
            </div>
          ) : (
            messages.map((message) => (
              <div 
                key={message.id}
                style={{
                  ...styles.message,
                  ...(message.role === 'user' ? styles.userMessage : styles.assistantMessage)
                }}
              >
                <div style={styles.messageHeader}>
                  <span style={styles.messageRole}>
                    {message.role === 'user' ? '👤 Вы' : '🤖 AI Agent'}
                  </span>
                  <span style={styles.messageTime}>
                    {new Date(message.timestamp).toLocaleTimeString()}
                  </span>
                </div>
                <div style={styles.messageContent}>
                  {message.content}
                </div>
              </div>
            ))
          )}
          <div ref={messagesEndRef} />
        </div>

        {/* Форма ввода */}
        <form onSubmit={handleSendMessage} style={styles.inputForm}>
          <input
            type="text"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            placeholder="Введите сообщение или команду..."
            style={styles.input}
            disabled={!isConnected}
          />
          <button 
            type="submit" 
            disabled={!inputValue.trim() || !isConnected}
            style={{
              ...styles.sendButton,
              opacity: (!inputValue.trim() || !isConnected) ? 0.6 : 1
            }}
          >
            📤
          </button>
        </form>
      </div>
    </div>
  );
};

// Стили
const styles = {
  container: {
    maxWidth: '1200px',
    margin: '0 auto',
    padding: '20px',
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    backgroundColor: '#f5f5f5',
    minHeight: '100vh'
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '20px',
    padding: '20px',
    backgroundColor: '#fff',
    borderRadius: '8px',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
  },
  title: {
    margin: 0,
    color: '#333',
    fontSize: '24px',
    fontWeight: '600'
  },
  statusIndicator: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px'
  },
  statusDot: {
    width: '12px',
    height: '12px',
    borderRadius: '50%',
    transition: 'background-color 0.3s'
  },
  statusText: {
    fontSize: '14px',
    color: '#666'
  },
  statusPanel: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
    gap: '15px',
    marginBottom: '20px',
    padding: '15px',
    backgroundColor: '#fff',
    borderRadius: '8px',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
  },
  statusItem: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center'
  },
  statusLabel: {
    fontSize: '14px',
    color: '#666',
    fontWeight: '500'
  },
  statusValue: {
    fontSize: '14px',
    fontWeight: '600',
    color: '#333'
  },
  controls: {
    display: 'flex',
    gap: '10px',
    marginBottom: '20px',
    flexWrap: 'wrap'
  },
  button: {
    padding: '10px 20px',
    border: '1px solid #ddd',
    borderRadius: '6px',
    backgroundColor: '#fff',
    color: '#333',
    cursor: 'pointer',
    fontSize: '14px',
    transition: 'all 0.2s',
    ':hover': {
      backgroundColor: '#f0f0f0'
    }
  },
  primaryButton: {
    backgroundColor: '#007bff',
    color: '#fff',
    borderColor: '#007bff',
    ':hover': {
      backgroundColor: '#0056b3'
    }
  },
  error: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '12px 16px',
    marginBottom: '20px',
    backgroundColor: '#ffebee',
    border: '1px solid #ffcdd2',
    borderRadius: '6px',
    color: '#c62828'
  },
  errorText: {
    fontSize: '14px'
  },
  errorClose: {
    background: 'none',
    border: 'none',
    color: '#c62828',
    cursor: 'pointer',
    fontSize: '16px',
    padding: '0 8px'
  },
  chatContainer: {
    backgroundColor: '#fff',
    borderRadius: '8px',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
    overflow: 'hidden'
  },
  messages: {
    height: '500px',
    overflowY: 'auto',
    padding: '20px'
  },
  welcome: {
    textAlign: 'center',
    color: '#666',
    padding: '40px 20px'
  },
  message: {
    marginBottom: '20px',
    padding: '15px',
    borderRadius: '8px',
    maxWidth: '80%'
  },
  userMessage: {
    backgroundColor: '#e3f2fd',
    marginLeft: 'auto',
    borderBottomRightRadius: '4px'
  },
  assistantMessage: {
    backgroundColor: '#f5f5f5',
    marginRight: 'auto',
    borderBottomLeftRadius: '4px'
  },
  messageHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '8px',
    fontSize: '12px',
    color: '#666'
  },
  messageRole: {
    fontWeight: '600'
  },
  messageTime: {
    opacity: '0.7'
  },
  messageContent: {
    fontSize: '14px',
    lineHeight: '1.5',
    whiteSpace: 'pre-wrap'
  },
  inputForm: {
    display: 'flex',
    padding: '20px',
    borderTop: '1px solid #eee',
    gap: '10px'
  },
  input: {
    flex: 1,
    padding: '12px 16px',
    border: '1px solid #ddd',
    borderRadius: '6px',
    fontSize: '14px',
    outline: 'none',
    ':focus': {
      borderColor: '#007bff'
    }
  },
  sendButton: {
    padding: '12px 16px',
    border: '1px solid #007bff',
    borderRadius: '6px',
    backgroundColor: '#007bff',
    color: '#fff',
    cursor: 'pointer',
    fontSize: '16px',
    transition: 'background-color 0.2s',
    ':hover': {
      backgroundColor: '#0056b3'
    }
  }
} as const; 