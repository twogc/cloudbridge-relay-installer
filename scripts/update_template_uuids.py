#!/usr/bin/env python3

import re
import uuid

# Список UUID для замены
uuids = [
    "f066fcfd-2e4d-4e5c-943f-98c6030fbaa1",  # p2p-quic-connections-active-001
    "ade9e116-fd01-4d0c-8e36-5ecc492c35d9",  # p2p-quic-connections-total-002
    "c2cef790-18b2-4793-bcf6-41dcc0cd70e1",  # p2p-quic-peers-active-003
    "45a127b9-11c2-41ab-9715-96a01be7a89e",  # p2p-quic-peers-discovered-004
    "5314d0a2-d365-4d3c-b3dd-5e9d704a3a5d",  # p2p-quic-peers-offline-005
    "c62ca807-1a84-47d2-87c2-b51608579bdb",  # p2p-quic-throughput-inbound-006
    "dfdfdc5b-c22c-4d4a-845d-2e0da0679560",  # p2p-quic-throughput-outbound-007
    "8fbe0bc3-0bda-4cfd-9fef-39e144db8edb",  # p2p-quic-latency-average-008
    "d0293d46-1b5c-432c-89b1-799de28d793b",  # p2p-quic-handshake-duration-009
    "acdf3ab8-ef0f-4aac-81dd-f64805e16c9d",  # p2p-quic-packets-sent-010
    "8cf8a73c-4098-46b7-9184-f33462741e11",  # p2p-quic-packets-received-011
    "9ff191e0-b399-413a-b4f9-663d473c0366",  # p2p-quic-packets-lost-012
    "f803d243-819e-4e09-aaea-853577a1afd1",  # p2p-quic-mesh-routes-active-013
    "252336c7-4080-45b8-b6d4-64ff41390f95",  # p2p-quic-streams-active-014
    "639f1764-8b1f-4ce6-a643-87b300aeb188",  # p2p-quic-security-ddos-blocked-015
    "ec8dc366-0053-48c6-ba3d-9de6908a08ce",  # p2p-quic-security-whitelist-violations-016
    "3540d8e1-5570-482a-8ea0-bba6d9ac629b",  # p2p-quic-security-rate-limit-hits-017
    "123c95a9-02ab-4be2-a777-833bf8d878f5",  # p2p-quic-enhanced-multipath-connections-018
    "10eb6343-e7ba-4729-8404-faf2fc2022ff",  # p2p-quic-enhanced-zero-rtt-connections-019
    "1619cf5c-8822-4bd7-9f27-5715e15ffca6",  # p2p-quic-enhanced-connection-migrations-020
    "acaeb879-b8bf-44fe-b7ab-bf790d0d7948",  # p2p-quic-fallback-events-021
    # Triggers
    "75fa9e0d-c978-43a6-b19a-809e6f1bf6e5",  # p2p-quic-no-active-connections-001
    "030ada5f-ba41-4a75-90c7-5693bb39ef19",  # p2p-quic-high-latency-002
    "636ff3e4-2fa1-43b7-bc7d-275d308a4fc3",  # p2p-quic-critical-latency-003
    "89372ca3-2b9e-42ae-8e1a-12417b48ed5a",  # p2p-quic-low-throughput-004
    "8cdf2989-e21d-4ebb-b674-9d6a6216ca90",  # p2p-quic-high-packet-loss-005
    "c244acb1-4a5c-4ce1-b3d9-393af4838b15",  # p2p-quic-critical-packet-loss-006
    "7ce5e57b-bbc9-43b1-b2db-ad8e4d586e66",  # p2p-quic-ddos-attack-007
    "b3a9589f-790f-4e45-b0da-782d9e706fe8",  # p2p-quic-whitelist-violations-008
    "c003932b-0bda-4e0f-8e45-9d30fba79447",  # p2p-quic-many-fallback-events-009
    "fdd919a6-7d21-4f6c-95b8-8f50dfd06dce",  # p2p-quic-no-mesh-routes-010
]

# Паттерны для замены
patterns = [
    (r'<uuid>p2p-quic-connections-active-001</uuid>', uuids[0]),
    (r'<uuid>p2p-quic-connections-total-002</uuid>', uuids[1]),
    (r'<uuid>p2p-quic-peers-active-003</uuid>', uuids[2]),
    (r'<uuid>p2p-quic-peers-discovered-004</uuid>', uuids[3]),
    (r'<uuid>p2p-quic-peers-offline-005</uuid>', uuids[4]),
    (r'<uuid>p2p-quic-throughput-inbound-006</uuid>', uuids[5]),
    (r'<uuid>p2p-quic-throughput-outbound-007</uuid>', uuids[6]),
    (r'<uuid>p2p-quic-latency-average-008</uuid>', uuids[7]),
    (r'<uuid>p2p-quic-handshake-duration-009</uuid>', uuids[8]),
    (r'<uuid>p2p-quic-packets-sent-010</uuid>', uuids[9]),
    (r'<uuid>p2p-quic-packets-received-011</uuid>', uuids[10]),
    (r'<uuid>p2p-quic-packets-lost-012</uuid>', uuids[11]),
    (r'<uuid>p2p-quic-mesh-routes-active-013</uuid>', uuids[12]),
    (r'<uuid>p2p-quic-streams-active-014</uuid>', uuids[13]),
    (r'<uuid>p2p-quic-security-ddos-blocked-015</uuid>', uuids[14]),
    (r'<uuid>p2p-quic-security-whitelist-violations-016</uuid>', uuids[15]),
    (r'<uuid>p2p-quic-security-rate-limit-hits-017</uuid>', uuids[16]),
    (r'<uuid>p2p-quic-enhanced-multipath-connections-018</uuid>', uuids[17]),
    (r'<uuid>p2p-quic-enhanced-zero-rtt-connections-019</uuid>', uuids[18]),
    (r'<uuid>p2p-quic-enhanced-connection-migrations-020</uuid>', uuids[19]),
    (r'<uuid>p2p-quic-fallback-events-021</uuid>', uuids[20]),
    # Triggers
    (r'<uuid>p2p-quic-no-active-connections-001</uuid>', uuids[21]),
    (r'<uuid>p2p-quic-high-latency-002</uuid>', uuids[22]),
    (r'<uuid>p2p-quic-critical-latency-003</uuid>', uuids[23]),
    (r'<uuid>p2p-quic-low-throughput-004</uuid>', uuids[24]),
    (r'<uuid>p2p-quic-high-packet-loss-005</uuid>', uuids[25]),
    (r'<uuid>p2p-quic-critical-packet-loss-006</uuid>', uuids[26]),
    (r'<uuid>p2p-quic-ddos-attack-007</uuid>', uuids[27]),
    (r'<uuid>p2p-quic-whitelist-violations-008</uuid>', uuids[28]),
    (r'<uuid>p2p-quic-many-fallback-events-009</uuid>', uuids[29]),
    (r'<uuid>p2p-quic-no-mesh-routes-010</uuid>', uuids[30]),
]

def update_template():
    # Читаем файл
    with open('cloudbridge-relay-template-fixed.xml', 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Заменяем все UUID
    for pattern, new_uuid in patterns:
        content = re.sub(pattern, f'<uuid>{new_uuid}</uuid>', content)
    
    # Записываем обновленный файл
    with open('cloudbridge-relay-template-fixed.xml', 'w', encoding='utf-8') as f:
        f.write(content)
    
    print("✅ Все UUID успешно обновлены!")

if __name__ == "__main__":
    update_template()
