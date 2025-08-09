#!/usr/bin/env python3
"""
Скрипт для отправки тестовых данных на порт 51820/udp
для тестирования P2P mesh соединения
"""

import socket
import time
import json
import sys

def send_udp_data(host='localhost', port=51820, message=None, count=1, interval=1):
    """
    Отправляет UDP данные на указанный хост и порт
    
    Args:
        host (str): IP адрес или hostname
        port (int): UDP порт
        message (str): Сообщение для отправки
        count (int): Количество сообщений
        interval (float): Интервал между сообщениями в секундах
    """
    
    # Создаем UDP сокет
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.settimeout(5)  # Таймаут 5 секунд
    
    server_address = (host, port)
    
    print(f"🔍 Отправка данных на {host}:{port}")
    print(f"📊 Количество сообщений: {count}")
    print(f"⏱️  Интервал: {interval} сек")
    print("-" * 50)
    
    for i in range(count):
        # Формируем сообщение
        if message:
            data = message
        else:
            # Создаем тестовое сообщение с метаданными
            test_data = {
                "type": "p2p_test",
                "timestamp": time.time(),
                "sequence": i + 1,
                "total": count,
                "data": f"Test message {i + 1} for P2P mesh"
            }
            data = json.dumps(test_data, indent=2)
        
        try:
            # Отправляем данные
            print(f"📤 Отправка сообщения {i + 1}/{count}:")
            print(f"   Данные: {data[:100]}{'...' if len(data) > 100 else ''}")
            
            bytes_sent = sock.sendto(data.encode('utf-8'), server_address)
            print(f"   ✅ Отправлено {bytes_sent} байт")
            
            # Пытаемся получить ответ (если сервер отвечает)
            try:
                sock.settimeout(2)
                response, addr = sock.recvfrom(1024)
                print(f"   📥 Ответ от {addr}: {response.decode('utf-8')}")
            except socket.timeout:
                print(f"   ⏰ Таймаут ответа (ожидаемо для UDP)")
            except Exception as e:
                print(f"   ❌ Ошибка получения ответа: {e}")
            
        except Exception as e:
            print(f"   ❌ Ошибка отправки: {e}")
        
        print()
        
        # Пауза между сообщениями
        if i < count - 1:
            time.sleep(interval)
    
    sock.close()
    print("🎉 Отправка завершена!")

def send_wireguard_style_data(host='localhost', port=51820):
    """
    Отправляет данные в стиле WireGuard handshake
    """
    print("🔐 Отправка WireGuard-style данных...")
    
    # Симуляция WireGuard handshake
    handshake_data = {
        "type": "wireguard_handshake",
        "version": "1.0",
        "timestamp": time.time(),
        "peer_id": "test-peer-001",
        "public_key": "test-public-key-12345",
        "endpoint": f"{host}:{port}",
        "capabilities": ["p2p", "mesh", "quic"]
    }
    
    send_udp_data(host, port, json.dumps(handshake_data), 1, 1)

def send_p2p_discovery_data(host='localhost', port=51820):
    """
    Отправляет данные для peer discovery
    """
    print("🔍 Отправка P2P discovery данных...")
    
    discovery_data = {
        "type": "peer_discovery",
        "action": "register",
        "peer_info": {
            "id": "discovery-peer-001",
            "region": "test-region",
            "capabilities": ["mesh", "quic", "wireguard"],
            "endpoints": [f"{host}:{port}"],
            "last_seen": time.time()
        }
    }
    
    send_udp_data(host, port, json.dumps(discovery_data), 1, 1)

def interactive_mode(host='localhost', port=51820):
    """
    Интерактивный режим для отправки данных
    """
    print("🎮 Интерактивный режим отправки UDP данных")
    print("Введите сообщения (пустая строка для выхода):")
    print("-" * 50)
    
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.settimeout(2)
    server_address = (host, port)
    
    try:
        while True:
            message = input("📝 Введите сообщение: ").strip()
            if not message:
                break
            
            try:
                bytes_sent = sock.sendto(message.encode('utf-8'), server_address)
                print(f"   ✅ Отправлено {bytes_sent} байт")
                
                # Пытаемся получить ответ
                try:
                    response, addr = sock.recvfrom(1024)
                    print(f"   📥 Ответ: {response.decode('utf-8')}")
                except socket.timeout:
                    print(f"   ⏰ Нет ответа (ожидаемо)")
                
            except Exception as e:
                print(f"   ❌ Ошибка: {e}")
            
            print()
    
    except KeyboardInterrupt:
        print("\n👋 Завершение интерактивного режима")
    finally:
        sock.close()

if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="Отправка UDP данных для тестирования P2P mesh")
    parser.add_argument("--host", default="localhost", help="Хост (по умолчанию: localhost)")
    parser.add_argument("--port", type=int, default=51820, help="Порт (по умолчанию: 51820)")
    parser.add_argument("--message", help="Сообщение для отправки")
    parser.add_argument("--count", type=int, default=1, help="Количество сообщений (по умолчанию: 1)")
    parser.add_argument("--interval", type=float, default=1, help="Интервал между сообщениями в секундах (по умолчанию: 1)")
    parser.add_argument("--wireguard", action="store_true", help="Отправить WireGuard-style данные")
    parser.add_argument("--discovery", action="store_true", help="Отправить P2P discovery данные")
    parser.add_argument("--interactive", action="store_true", help="Интерактивный режим")
    
    args = parser.parse_args()
    
    if args.interactive:
        interactive_mode(args.host, args.port)
    elif args.wireguard:
        send_wireguard_style_data(args.host, args.port)
    elif args.discovery:
        send_p2p_discovery_data(args.host, args.port)
    else:
        send_udp_data(args.host, args.port, args.message, args.count, args.interval)


