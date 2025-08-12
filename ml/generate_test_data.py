import csv
import random
from datetime import datetime, timedelta

fieldnames = [
    'timestamp', 'bytes_sent', 'bytes_received', 'packets_sent', 'packets_received',
    'connection_duration', 'requests_per_second', 'error_rate', 'user_id_hash',
    'session_duration', 'requests_in_session', 'unique_endpoints',
    'failed_auth_attempts', 'suspicious_ips', 'rate_limit_violations', 'ssl_errors'
]

rows = []
start_time = datetime(2024, 6, 1, 12, 0, 0)
for i in range(150):
    row = {
        'timestamp': (start_time + timedelta(minutes=i)).isoformat(),
        'bytes_sent': random.randint(100, 10000),
        'bytes_received': random.randint(100, 20000),
        'packets_sent': random.randint(10, 1000),
        'packets_received': random.randint(10, 2000),
        'connection_duration': random.randint(1, 3600),
        'requests_per_second': random.randint(1, 100),
        'error_rate': round(random.uniform(0, 0.1), 4),
        'user_id_hash': random.randint(10000, 99999),
        'session_duration': random.randint(10, 3600),
        'requests_in_session': random.randint(1, 100),
        'unique_endpoints': random.randint(1, 20),
        'failed_auth_attempts': random.randint(0, 5),
        'suspicious_ips': random.randint(0, 2),
        'rate_limit_violations': random.randint(0, 3),
        'ssl_errors': random.randint(0, 2),
    }
    rows.append(row)

with open('detect_requests_log.csv', 'w', newline='') as f:
    writer = csv.DictWriter(f, fieldnames=fieldnames)
    writer.writeheader()
    writer.writerows(rows)

print('Generated 150 test rows in detect_requests_log.csv') 