import jwt
import time
import socket
import json
import struct
import base64

def send_message(host, port, message):
    # Создаем сокет
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    try:
        # Подключаемся к серверу
        print(f"Connecting to {host}:{port}...")
        sock.connect((host, port))
        
        # Преобразуем сообщение в JSON
        json_data = json.dumps(message)
        
        # Получаем длину сообщения
        length = len(json_data)
        print(f"Message length: {length}")
        print(f"Message length (hex): {length:08x}")
        
        # Отправляем длину сообщения (4 байта BigEndian)
        sock.send(struct.pack('>I', length))
        
        # Отправляем само сообщение
        sock.send(json_data.encode())
        
        # Получаем ответ
        response_length_data = sock.recv(4)
        response_length = struct.unpack('>I', response_length_data)[0]
        response_data = sock.recv(response_length)
        response = json.loads(response_data.decode())
        
        print(f"Response: {response}")
        
    except Exception as e:
        print(f"Error: {e}")
    finally:
        sock.close()

if __name__ == '__main__':
    # Секрет для подписи JWT (base64-decode)
    secret = "Yk7+4of0fCmcHY+p0QCZHonEMO0+WUvbKsrjfvJL3zk="
    
    # Создаем JWT токен
    payload = {
        "exp": int(time.time()) + 3600,  # срок действия 1 час
        "iat": int(time.time()),         # время выпуска
        "sub": "test-user"               # идентификатор пользователя
    }
    
    # Генерируем токен
    token = jwt.encode(payload, secret, algorithm="HS256")
    print(f"Generated token: {token}")
    
    # Тестовое сообщение
    test_message = {
        'type': 'register',  # MessageTypeRegister
        'token': token,
        'version': '1.0.0'
    }
    
    print(f"Message to send: {test_message}")
    
    # Параметры подключения
    HOST = '127.0.0.1'
    PORT = 8080
    
    # Отправляем сообщение
    send_message(HOST, PORT, test_message) 