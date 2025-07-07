import { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { userManager } from '../config/keycloak';

const Callback = () => {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [info, setInfo] = useState<string>('Обрабатываем callback...');
  const calledRef = useRef(false);

  useEffect(() => {
    if (calledRef.current) return;
    calledRef.current = true;

    setInfo('Выполняем signinRedirectCallback...');
    userManager.signinRedirectCallback()
      .then((user) => {
        setInfo('Авторизация успешна! Пользователь: ' + JSON.stringify(user));
        setTimeout(() => navigate('/', { replace: true }), 5000);
      })
      .catch((e) => {
        setError('Auth callback error: ' + (e && e.message ? e.message : e));
        setInfo('Ошибка при signinRedirectCallback. См. консоль.');
        console.error('Auth callback error', e);
        setTimeout(() => navigate('/', { replace: true }), 5000);
      });
  }, [navigate]);

  return (
    <div style={{ padding: 32 }}>
      <h2>Завершаем вход...</h2>
      <div style={{ color: 'blue', marginBottom: 16 }}>{info}</div>
      {error && <div style={{ color: 'red', fontWeight: 'bold' }}>{error}</div>}
      <div style={{ marginTop: 24, fontSize: 12, color: '#888' }}>
        Если вы видите ошибку, сделайте скриншот этого экрана и отправьте разработчику.<br/>
        После 5 секунд произойдет возврат на главную страницу.
      </div>
    </div>
  );
};

export default Callback; 