#!/usr/bin/env python3

import os
import re

def fix_cadence_integration_test():
    """Исправляет cadence_integration_test.go, заменяя WireGuard на P2P QUIC"""
    file_path = "internal/integration/cadence_integration_test.go"
    
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
        # Заменяем PeerLifecycleWorkflowInput на P2PConnectionOrchestrationWorkflowInput
        (r'workflows\.PeerLifecycleWorkflowInput', 'workflows.P2PConnectionOrchestrationWorkflowInput'),
        (r'workflows\.PeerLifecycleWorkflowResult', 'workflows.P2PConnectionOrchestrationWorkflowResult'),
        (r'"PeerLifecycleWorkflow"', '"P2PConnectionOrchestrationWorkflow"'),
        (r'TestPeerLifecycleWorkflow', 'TestP2PConnectionOrchestrationWorkflow'),
        (r'PeerLifecycleWorkflow', 'P2PConnectionOrchestrationWorkflow'),
        
        # Заменяем поля структуры
        (r'PeerID:\s*"([^"]*)"', r'ConnectionID: "\1"'),
        (r'Operation:\s*"([^"]*)"', r'ConnectionType: "direct"'),
        (r'PeerConfig:\s*map\[string\]interface\{\}', 'Configuration: map[string]interface{}'),
        (r'NetworkConfig:\s*map\[string\]interface\{\}', 'Configuration: map[string]interface{}'),
        
        # Заменяем проверки результата
        (r'result\.PeerID', 'result.ConnectionID'),
        (r'result\.Operation', 'result.ConnectionType'),
        (r'"peer_id"', '"connection_id"'),
        (r'"operation"', '"connection_type"'),
        
        # Заменяем комментарии
        (r'WireGuard workflow', 'P2P QUIC workflow'),
        (r'управления пирами', 'управления соединениями'),
        (r'добавления пира', 'установления соединения'),
        (r'удаления пира', 'закрытия соединения'),
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
    fix_cadence_integration_test()
