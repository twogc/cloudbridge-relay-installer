import pandas as pd
from anomaly_detector import AnomalyDetector
import sys

# Использование: python3 train_anomaly_detector.py historical_data.csv
if len(sys.argv) < 2:
    print("Usage: python3 train_anomaly_detector.py <historical_data.csv>")
    sys.exit(1)

csv_path = sys.argv[1]

# Загрузка данных
print(f"Loading data from {csv_path}...")
df = pd.read_csv(csv_path)

# Проверяем наличие нужных полей
required_fields = [
    'timestamp', 'bytes_sent', 'bytes_received', 'packets_sent', 'packets_received',
    'connection_duration', 'requests_per_second', 'error_rate', 'user_id_hash',
    'session_duration', 'requests_in_session', 'unique_endpoints',
    'failed_auth_attempts', 'suspicious_ips', 'rate_limit_violations', 'ssl_errors'
    # 'latency'  # для обучения performance_predictor (делаем не обязательным)
]
missing = [f for f in required_fields if f not in df.columns]
if missing:
    print(f"Missing required fields in CSV: {missing}")
    sys.exit(1)

# Преобразуем в список dict
historical_data = df.to_dict(orient='records')

# Обучаем модель
print("Training anomaly detector models...")
detector = AnomalyDetector()
trained = detector.train_models(historical_data)
if trained:
    detector.save_models('models.joblib')
    print("Models trained and saved to models.joblib")
else:
    print("Not enough data for training (need at least 100 samples)") 