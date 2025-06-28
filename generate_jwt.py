import jwt
import os
from datetime import datetime, timedelta

def generate_jwt():
    """Генерация JWT токена для тестирования"""
    # Получаем секрет из переменной окружения
    secret = os.getenv("RELAY_JWT_SECRET")
    if not secret:
        print("ОШИБКА: Переменная окружения RELAY_JWT_SECRET не установлена")
        print("Установите её в файле .env или экспортируйте в системе")
        return None
    
    # Полезная нагрузка токена
    payload = {
        'user_id': 123,
        'username': 'test_user',
        'type': 'relay_access',
        'exp': datetime.utcnow() + timedelta(hours=1),
        'iat': datetime.utcnow(),
        'iss': 'cloudbridge-relay',
        'aud': 'relay-client'
    }
    
    # Генерация токена
    token = jwt.encode(payload, secret, algorithm="HS256")
    print(token)
    return token

if __name__ == "__main__":
    generate_jwt()