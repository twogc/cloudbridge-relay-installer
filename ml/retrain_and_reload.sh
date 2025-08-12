#!/bin/bash
# Скрипт: дообучение модели и hot-reload в FastAPI
set -e
cd "$(dirname "$0")"
source venv/bin/activate
python3 train_anomaly_detector.py detect_requests_log.csv
curl -X POST http://127.0.0.1:8000/reload_model 