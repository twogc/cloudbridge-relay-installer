#!/usr/bin/env python3

import os
import re

def fix_p2p_manager_test():
    """Исправляет manager_test.go, заменяя WireGuard на P2P QUIC"""
    file_path = "internal/p2p/manager_test.go"
    
    print(f"Исправление файла: {file_path}")
    
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Создаем резервную копию
    backup_path = file_path + '.backup'
    with open(backup_path, 'w', encoding='utf-8') as f:
        f.write(content)
    print(f"  ✓ Резервная копия создана: {backup_path}")
    
    # Замены
    replacements = [
        # Заменяем WireGuard на P2P
        (r'wireguard\.', ''),
        (r'WireGuard', 'P2P'),
        (r'wireguard', 'p2p'),
        
        # Заменяем структуры
        (r'WireGuardPeer', 'Peer'),
        (r'WireGuardConfig', 'QuicConfig'),
        (r'DiscoveredPeer', 'DiscoveredPeer'),
        
        # Заменяем поля
        (r'PublicKey', 'ID'),
        (r'AllowedIPs', 'Capabilities'),
        (r'LastHandshake', 'LastSeen'),
        (r'PeerStatus', 'Status'),
        
        # Заменяем значения статусов
        (r'PeerStatusConnecting', '"connecting"'),
        (r'PeerStatusConnected', '"connected"'),
        (r'PeerStatusDisconnected', '"disconnected"'),
        
        # Заменяем TrustLevel
        (r'TrustLevel', 'TrustLevel'),
        (r'TrustBasic', '"basic"'),
        (r'TrustHigh', '"high"'),
        
        # Заменяем интерфейсы
        (r'WireGuardMeshTopologyInterface', 'P2PMeshTopologyInterface'),
    ]
    
    # Применяем замены
    for pattern, replacement in replacements:
        content = re.sub(pattern, replacement, content)
    
    # Сохраняем исправленный файл
    with open(file_path, 'w', encoding='utf-8') as f:
        f.write(content)
    
    print(f"  ✓ Файл исправлен")
    return True

if __name__ == "__main__":
    fix_p2p_manager_test()
