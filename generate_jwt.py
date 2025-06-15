import jwt
import time

# Ваш секрет
secret = "Yk7+4of0fCmcHY+p0QCZHonEMO0+WUvbKsrjfvJL3zk="

# Параметры токена
payload = {
    "exp": int(time.time()) + 3600,  # срок действия 1 час
    "iat": int(time.time()),         # время выпуска
    "sub": "test-user"               # любой идентификатор пользователя
}

# Генерация токена
token = jwt.encode(payload, secret, algorithm="HS256")
print(token)