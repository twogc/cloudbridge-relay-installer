from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field, validator
from datetime import datetime
from anomaly_detector import AnomalyDetector
import joblib
import os
import structlog
from dotenv import load_dotenv
import csv
from fastapi import APIRouter
from prometheus_client import Counter, Gauge, generate_latest, CONTENT_TYPE_LATEST
from fastapi import Response

load_dotenv(dotenv_path=os.path.join(os.path.dirname(__file__), '.env'), override=True)

app = FastAPI(title="Anomaly Detection API")
logger = structlog.get_logger()

# Метрики Prometheus
anomaly_counter = Counter('anomaly_detected_total', 'Total detected anomalies')
error_rate_gauge = Gauge('error_rate_last', 'Last error rate value')
requests_total = Counter('detect_requests_total', 'Total /detect requests')

# Pydantic модель для входных данных
class AnomalyRequest(BaseModel):
    timestamp: str = Field(..., description="ISO8601 timestamp")
    bytes_sent: int = Field(..., ge=0, le=2_000_000_000)
    bytes_received: int = Field(..., ge=0, le=2_000_000_000)
    packets_sent: int = Field(..., ge=0, le=10_000_000)
    packets_received: int = Field(..., ge=0, le=10_000_000)
    connection_duration: int = Field(..., ge=0, le=86_400, description="Seconds")
    requests_per_second: int = Field(..., ge=0, le=100_000)
    error_rate: float = Field(..., ge=0.0, le=1.0)
    user_id_hash: int = Field(..., ge=0)
    session_duration: int = Field(..., ge=0, le=86_400)
    requests_in_session: int = Field(..., ge=0, le=1_000_000)
    unique_endpoints: int = Field(..., ge=0, le=10_000)
    failed_auth_attempts: int = Field(..., ge=0, le=1_000)
    suspicious_ips: int = Field(..., ge=0, le=1_000)
    rate_limit_violations: int = Field(..., ge=0, le=10_000)
    ssl_errors: int = Field(..., ge=0, le=10_000)

    @validator('timestamp')
    def validate_timestamp(cls, v):
        try:
            datetime.fromisoformat(v)
        except Exception:
            raise ValueError('timestamp must be valid ISO8601 string')
        return v

# Инициализация детектора и загрузка моделей
MODEL_PATH = os.getenv("MODEL_PATH", "models.joblib")
detector = AnomalyDetector()
if os.path.exists(MODEL_PATH):
    detector.load_models(MODEL_PATH)
    logger.info("Loaded trained models.", model_path=MODEL_PATH)
else:
    logger.warning("Trained models not found. API будет работать в режиме без обучения.", model_path=MODEL_PATH)

@app.post("/detect")
def detect_anomaly(request: AnomalyRequest):
    data = request.dict()
    logger.info("Received detect request", data=data)
    # --- Сохраняем входящий запрос в CSV ---
    csv_file = "detect_requests_log.csv"
    file_exists = os.path.isfile(csv_file)
    with open(csv_file, mode="a", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=data.keys())
        if not file_exists:
            writer.writeheader()
        writer.writerow(data)
    # --- Prometheus метрики ---
    requests_total.inc()
    # --- конец блока сохранения ---
    try:
        result = detector.detect_anomalies(data)
        if result.get('is_anomaly'):
            anomaly_counter.inc()
        error_rate_gauge.set(data.get('error_rate', 0))
        logger.info("Detection result", result=result)
        return result
    except Exception as e:
        logger.error("Error in /detect endpoint", error=str(e), data=data)
        raise HTTPException(status_code=500, detail="Internal server error")

@app.post("/reload_model")
def reload_model():
    try:
        detector.load_models(MODEL_PATH)
        logger.info("Model reloaded via /reload_model endpoint", model_path=MODEL_PATH)
        return {"status": "ok", "message": "Model reloaded"}
    except Exception as e:
        logger.error("Error reloading model", error=str(e))
        return {"status": "error", "message": str(e)}

@app.get("/health")
def health():
    logger.info("Health check", models_trained=detector.models_trained)
    return {"status": "ok", "models_trained": detector.models_trained}

@app.get("/metrics")
def metrics():
    return Response(generate_latest(), media_type=CONTENT_TYPE_LATEST) 