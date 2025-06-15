import jwt
import time
import socket
import json
import struct
import base64
from contextlib import contextmanager

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

if __name__ == '__main__':
    # Секрет для подписи JWT (должен совпадать с RELAY_JWT_SECRET в Django)
    secret = "Yk7+4of0fCmcHY+p0QCZHonEMO0+WUvbKsrjfvJL3zk="
    
    # Создаем JWT токен
    payload = {
        "id": "test-id",           # Добавляем идентификатор
        "server_id": "test-server",
        "company_id": "test-company",
        "exp": int(time.time()) + 3600,  # срок действия 1 час
        "iat": int(time.time()),         # время выпуска
        "type": "relay"                  # тип токена
    }
    
    # Генерируем токен
    token = jwt.encode(payload, secret, algorithm="HS256")
    print(f"Generated token: {token}")
    
    # Декодируем токен для проверки
    decoded = jwt.decode(token, secret, algorithms=["HS256"])
    print(f"Decoded token: {json.dumps(decoded, indent=2)}")
    
    # Тестовое сообщение
    test_message = {
        'type': 1,  # MessageTypeRegister
        'token': token
    }
    
    print(f"Message to send: {test_message}")
    
    # Отправляем сообщение
    with socket_context('127.0.0.1', 8080) as sock:
        print("Sending message...")
        send_message(sock, test_message)
        
        print("Waiting for response...")
        response = read_message(sock)
        print(f"Response: {json.dumps(response, indent=2)}") 