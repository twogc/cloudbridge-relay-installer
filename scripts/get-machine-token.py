#!/usr/bin/env python3
"""
ZITADEL Machine User Token Generator
Генерирует JWT Bearer токен из машинного ключа ZITADEL для API доступа
"""

import json
import jwt
import time
import requests
import sys
import os
from datetime import datetime, timedelta
from cryptography.hazmat.primitives import serialization

def load_machine_key(key_file):
    """Загружает machine user ключ из JSON файла"""
    try:
        with open(key_file, 'r') as f:
            return json.load(f)
    except Exception as e:
        print(f"Error loading machine key: {e}", file=sys.stderr)
        sys.exit(1)

def create_jwt_assertion(machine_key, issuer_url):
    """Создает JWT assertion для получения access token"""
    
    # Извлекаем приватный ключ
    private_key_pem = machine_key['key']
    private_key = serialization.load_pem_private_key(
        private_key_pem.encode('utf-8'),
        password=None
    )
    
    # Создаем claims для JWT assertion
    now = datetime.utcnow()
    claims = {
        'iss': machine_key['userId'],     # Issuer = user ID
        'sub': machine_key['userId'],     # Subject = user ID  
        'aud': issuer_url,                # Audience = ZITADEL URL
        'exp': int((now + timedelta(minutes=5)).timestamp()),
        'iat': int(now.timestamp()),
        'jti': f"jwt-{int(now.timestamp())}"  # Unique ID
    }
    
    # Создаем JWT токен
    token = jwt.encode(
        claims,
        private_key,
        algorithm='RS256',
        headers={'kid': machine_key['keyId']}
    )
    
    return token

def get_access_token(jwt_assertion, token_url):
    """Получает access token используя JWT Bearer assertion"""
    
    data = {
        'grant_type': 'urn:ietf:params:oauth:grant-type:jwt-bearer',
        'assertion': jwt_assertion,
        'scope': 'openid profile email urn:zitadel:iam:org:project:id:zitadel:aud'
    }
    
    headers = {
        'Content-Type': 'application/x-www-form-urlencoded'
    }
    
    try:
        response = requests.post(token_url, data=data, headers=headers, verify=False)
        response.raise_for_status()
        
        token_data = response.json()
        return token_data.get('access_token')
        
    except requests.exceptions.RequestException as e:
        print(f"Error getting access token: {e}", file=sys.stderr)
        if hasattr(e, 'response') and e.response is not None:
            print(f"Response: {e.response.text}", file=sys.stderr)
        return None

def main():
    # Конфигурация
    key_file = os.environ.get('ZITADEL_MACHINE_KEY_FILE', 
                             os.path.expanduser('~/.zitadel/machine-key.json'))
    issuer_url = 'https://edge.2gc.ru'
    token_url = f'{issuer_url}/oauth/v2/token'
    
    if not os.path.exists(key_file):
        print(f"Machine key file not found: {key_file}", file=sys.stderr)
        sys.exit(1)
    
    # Загружаем ключ
    print(f"Loading machine key from: {key_file}", file=sys.stderr)
    machine_key = load_machine_key(key_file)
    
    # Создаем JWT assertion
    print("Creating JWT assertion...", file=sys.stderr)
    jwt_assertion = create_jwt_assertion(machine_key, issuer_url)
    
    # Получаем access token
    print("Getting access token...", file=sys.stderr)
    access_token = get_access_token(jwt_assertion, token_url)
    
    if access_token:
        print("Access token obtained successfully!", file=sys.stderr)
        print(access_token)  # Выводим токен в stdout
    else:
        print("Failed to get access token", file=sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    main()
