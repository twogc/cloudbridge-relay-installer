#!/usr/bin/env python3
"""
SIEM Web UI
Простой веб-интерфейс для CloudBridge SIEM
"""

from fastapi import FastAPI, Request, HTTPException
from fastapi.responses import HTMLResponse, JSONResponse
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates
import uvicorn
import os
import json
from datetime import datetime, timedelta
from typing import Dict, List, Any
import requests

app = FastAPI(title="CloudBridge SIEM", version="1.0.0")

# Настройка шаблонов
templates = Jinja2Templates(directory="webui/templates")

# Конфигурация
ELASTICSEARCH_URL = os.getenv("ELASTICSEARCH_URL", "http://elasticsearch:9200")
KIBANA_URL = os.getenv("KIBANA_URL", "http://kibana:5601")

@app.get("/", response_class=HTMLResponse)
async def dashboard(request: Request):
    """Главная страница дашборда"""
    try:
        # Получение статистики
        stats = await get_siem_stats()
        return templates.TemplateResponse("dashboard.html", {
            "request": request,
            "stats": stats,
            "kibana_url": KIBANA_URL
        })
    except Exception as e:
        return templates.TemplateResponse("error.html", {
            "request": request,
            "error": str(e)
        })

@app.get("/api/stats")
async def get_stats():
    """API для получения статистики"""
    try:
        stats = await get_siem_stats()
        return JSONResponse(content=stats)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/alerts")
async def get_alerts(limit: int = 10, severity: str = None):
    """API для получения уведомлений"""
    try:
        alerts = await get_recent_alerts(limit, severity)
        return JSONResponse(content=alerts)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/events")
async def get_events(limit: int = 50, source: str = None):
    """API для получения событий"""
    try:
        events = await get_recent_events(limit, source)
        return JSONResponse(content=events)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

async def get_siem_stats() -> Dict[str, Any]:
    """Получение статистики SIEM"""
    try:
        # Статистика событий
        events_query = {
            "size": 0,
            "aggs": {
                "events_by_source": {
                    "terms": {"field": "source"}
                },
                "events_by_severity": {
                    "terms": {"field": "severity"}
                },
                "events_over_time": {
                    "date_histogram": {
                        "field": "timestamp",
                        "interval": "1h"
                    }
                }
            }
        }
        
        events_response = requests.post(
            f"{ELASTICSEARCH_URL}/siem-events/_search",
            json=events_query,
            timeout=10
        )
        
        if events_response.status_code == 200:
            events_data = events_response.json()
            events_count = events_data["hits"]["total"]["value"]
        else:
            events_count = 0
        
        # Статистика уведомлений
        alerts_query = {
            "size": 0,
            "aggs": {
                "alerts_by_severity": {
                    "terms": {"field": "severity"}
                },
                "alerts_by_status": {
                    "terms": {"field": "status"}
                }
            }
        }
        
        alerts_response = requests.post(
            f"{ELASTICSEARCH_URL}/siem-alerts/_search",
            json=alerts_query,
            timeout=10
        )
        
        if alerts_response.status_code == 200:
            alerts_data = alerts_response.json()
            alerts_count = alerts_data["hits"]["total"]["value"]
        else:
            alerts_count = 0
        
        return {
            "total_events": events_count,
            "total_alerts": alerts_count,
            "timestamp": datetime.utcnow().isoformat(),
            "system_status": "healthy"
        }
        
    except Exception as e:
        return {
            "total_events": 0,
            "total_alerts": 0,
            "timestamp": datetime.utcnow().isoformat(),
            "system_status": "error",
            "error": str(e)
        }

async def get_recent_alerts(limit: int = 10, severity: str = None) -> List[Dict[str, Any]]:
    """Получение последних уведомлений"""
    try:
        query = {
            "size": limit,
            "sort": [{"timestamp": {"order": "desc"}}]
        }
        
        if severity:
            query["query"] = {"term": {"severity": severity}}
        
        response = requests.post(
            f"{ELASTICSEARCH_URL}/siem-alerts/_search",
            json=query,
            timeout=10
        )
        
        if response.status_code == 200:
            data = response.json()
            alerts = []
            for hit in data["hits"]["hits"]:
                alert = hit["_source"]
                alert["id"] = hit["_id"]
                alerts.append(alert)
            return alerts
        else:
            return []
            
    except Exception as e:
        return []

async def get_recent_events(limit: int = 50, source: str = None) -> List[Dict[str, Any]]:
    """Получение последних событий"""
    try:
        query = {
            "size": limit,
            "sort": [{"timestamp": {"order": "desc"}}]
        }
        
        if source:
            query["query"] = {"term": {"source": source}}
        
        response = requests.post(
            f"{ELASTICSEARCH_URL}/siem-events/_search",
            json=query,
            timeout=10
        )
        
        if response.status_code == 200:
            data = response.json()
            events = []
            for hit in data["hits"]["hits"]:
                event = hit["_source"]
                event["id"] = hit["_id"]
                events.append(event)
            return events
        else:
            return []
            
    except Exception as e:
        return []

@app.get("/health")
async def health_check():
    """Проверка здоровья системы"""
    try:
        # Проверка Elasticsearch
        es_response = requests.get(f"{ELASTICSEARCH_URL}/_cluster/health", timeout=5)
        es_healthy = es_response.status_code == 200
        
        # Проверка Kibana
        kibana_response = requests.get(f"{KIBANA_URL}/api/status", timeout=5)
        kibana_healthy = kibana_response.status_code == 200
        
        return {
            "status": "healthy" if es_healthy and kibana_healthy else "degraded",
            "timestamp": datetime.utcnow().isoformat(),
            "services": {
                "elasticsearch": "healthy" if es_healthy else "unhealthy",
                "kibana": "healthy" if kibana_healthy else "unhealthy"
            }
        }
    except Exception as e:
        return {
            "status": "unhealthy",
            "timestamp": datetime.utcnow().isoformat(),
            "error": str(e)
        }

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000) 