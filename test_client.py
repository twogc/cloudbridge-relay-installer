import requests
import json
import jwt
import os
import time
from datetime import datetime, timedelta
import socket
import struct
import base64
from contextlib import contextmanager

# URL сервера
BASE_URL = "http://localhost:8080"

@contextmanager
def socket_context(host, port):
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    try:
        sock.connect((host, port))
        yield sock
    finally:
        sock.close()

def send_message(sock, message):
    # Convert message to JSON
    data = json.dumps(message).encode()
    # Add 4-byte length prefix (BigEndian)
    length = struct.pack('>I', len(data))
    # Send length prefix and message
    sock.sendall(length + data)

def read_message(sock):
    # Read 4-byte length prefix
    length_data = sock.recv(4)
    if len(length_data) != 4:
        raise Exception("Invalid message length")
    length = struct.unpack('>I', length_data)[0]
    # Read message data
    data = sock.recv(length)
    return json.loads(data)

def generate_jwt_token():
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
    print(f"Generated token: {token}")
    
    # Декодирование для проверки
    try:
        decoded = jwt.decode(token, secret, algorithms=["HS256"])
        print(f"Decoded token: {json.dumps(decoded, indent=2)}")
    except Exception as e:
        print(f"Error decoding token: {e}")
        return None
    
    return token

def test_relay_api(token):
    """Тестирование API relay"""
    if not token:
        print("Нет токена для тестирования")
        return
    
    headers = {
        'Content-Type': 'application/json',
        'Authorization': f'Bearer {token}'
    }
    
    # Тестовые данные
    data = {
        'token': token
    }
    
    try:
        # Тест валидации токена
        response = requests.post(
            f"{BASE_URL}/api/relay/token/validate/",
            headers=headers,
            json=data,
            timeout=10
        )
        
        print(f"Status Code: {response.status_code}")
        print(f"Response: {response.text}")
        
        if response.status_code == 200:
            print("✅ Тест валидации токена прошел успешно")
        else:
            print("❌ Тест валидации токена не прошел")
            
    except requests.exceptions.RequestException as e:
        print(f"Ошибка запроса: {e}")

def main():
    print("=== CloudBridge Relay Test Client ===")
    print(f"Testing against: {BASE_URL}")
    print()
    
    # Генерация токена
    token = generate_jwt_token()
    
    if token:
        # Тестирование API
        test_relay_api(token)
    else:
        print("Не удалось сгенерировать токен. Проверьте настройки.")

if __name__ == "__main__":
    main() 