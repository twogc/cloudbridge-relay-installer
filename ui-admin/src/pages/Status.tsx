import { useEffect, useState } from 'react';

export default function Status() {
  const [status, setStatus] = useState<string>('');
  const [message, setMessage] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetch('/api/status')
      .then(res => res.json())
      .then(data => {
        setStatus(data.status);
        setMessage(data.message);
        setLoading(false);
      })
      .catch(() => {
        setError('Ошибка запроса к API');
        setLoading(false);
      });
  }, []);

  if (loading) return <div>Загрузка...</div>;
  if (error) return <div style={{color: 'red'}}>{error}</div>;

  return (
    <div>
      <h2>Статус API сервера</h2>
      <p>Status: <b>{status}</b></p>
      <p>Message: {message}</p>
    </div>
  );
} 