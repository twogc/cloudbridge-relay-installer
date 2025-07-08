from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from anomaly_detector import AnomalyDetector
import joblib
import os

app = FastAPI(title="Anomaly Detection API")

# Pydantic модель для входных данных
class AnomalyRequest(BaseModel):
    timestamp: str
    bytes_sent: int
    bytes_received: int
    packets_sent: int
    packets_received: int
    connection_duration: int
    requests_per_second: int
    error_rate: float
    user_id_hash: int
    session_duration: int
    requests_in_session: int
    unique_endpoints: int
    failed_auth_attempts: int
    suspicious_ips: int
    rate_limit_violations: int
    ssl_errors: int

# Инициализация детектора и загрузка моделей
MODEL_PATH = "models.joblib"
detector = AnomalyDetector()
if os.path.exists(MODEL_PATH):
    detector.load_models(MODEL_PATH)
    print("Loaded trained models.")
else:
    print("Warning: Trained models not found. API будет работать в режиме без обучения.")

@app.post("/detect")
def detect_anomaly(request: AnomalyRequest):
    data = request.dict()
    result = detector.detect_anomalies(data)
    return result

@app.get("/health")
def health():
    return {"status": "ok", "models_trained": detector.models_trained} 