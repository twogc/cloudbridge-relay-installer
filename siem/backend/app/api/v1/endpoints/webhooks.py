"""
Webhook endpoints for SIEM API
"""
from typing import Dict, Any
from fastapi import APIRouter, HTTPException, Request
import structlog
from datetime import datetime
import uuid

from app.services.events import EventService
from app.schemas.events import EventCreate

router = APIRouter()
logger = structlog.get_logger()


@router.post("/cloudflare-agent")
async def cloudflare_agent_webhook(request: Request):
    """Webhook endpoint for Cloudflare AI Agent events"""
    try:
        # Get request body
        body = await request.json()
        logger.info("Received Cloudflare Agent webhook", data=body)
        
        # Extract event data
        event_data = {
            "event_id": str(uuid.uuid4()),
            "timestamp": body.get("timestamp", datetime.utcnow().isoformat()),
            "source": "cloudflare_agent",
            "severity": body.get("severity", "medium"),
            "category": "security",
            "data": {
                "anomaly_score": body.get("anomaly_score", 0.0),
                "is_anomaly": body.get("is_anomaly", False),
                "metrics": body.get("metrics", {}),
                "ai_analysis": body.get("ai_analysis", ""),
                "context": body.get("context", {})
            },
            "tags": body.get("tags", []),
            "correlation_id": body.get("correlation_id")
        }
        
        # Create event
        event_service = EventService()
        from ...schemas.events import EventCreate
        event_create = EventCreate(**event_data)
        created_event = await event_service.create_event(event_create)
        
        logger.info("Created event from Cloudflare Agent", event_id=created_event.id)
        
        return {
            "status": "success",
            "event_id": created_event.id,
            "message": "Event created successfully"
        }
        
    except Exception as e:
        logger.error("Error processing Cloudflare Agent webhook", error=str(e))
        raise HTTPException(status_code=500, detail="Failed to process webhook")


@router.post("/ml-api")
async def ml_api_webhook(request: Request):
    """Webhook endpoint for ML API events"""
    try:
        # Get request body
        body = await request.json()
        logger.info("Received ML API webhook", data=body)
        
        # Extract event data
        event_data = {
            "event_id": str(uuid.uuid4()),
            "timestamp": body.get("timestamp", datetime.utcnow().isoformat()),
            "source": "ml_api",
            "severity": "high" if body.get("is_anomaly", False) else "medium",
            "category": "anomaly",
            "data": {
                "anomaly_score": body.get("anomaly_score", 0.0),
                "is_anomaly": body.get("is_anomaly", False),
                "confidence": body.get("confidence", 0.0),
                "anomaly_type": body.get("anomaly_type", "unknown"),
                "recommendations": body.get("recommendations", []),
                "features_used": body.get("features_used", 0)
            },
            "tags": ["ml_analysis", "anomaly_detection"] + (["anomaly"] if body.get("is_anomaly") else []),
            "correlation_id": body.get("correlation_id")
        }
        
        # Create event
        event_service = EventService()
        from ...schemas.events import EventCreate
        event_create = EventCreate(**event_data)
        created_event = await event_service.create_event(event_create)
        
        logger.info("Created event from ML API", event_id=created_event.id)
        
        return {
            "status": "success",
            "event_id": created_event.id,
            "message": "Event created successfully"
        }
        
    except Exception as e:
        logger.error("Error processing ML API webhook", error=str(e))
        raise HTTPException(status_code=500, detail="Failed to process webhook")


@router.post("/relay-metrics")
async def relay_metrics_webhook(request: Request):
    """Webhook endpoint for Relay Metrics events"""
    try:
        # Get request body
        body = await request.json()
        logger.info("Received Relay Metrics webhook", data=body)
        
        # Extract event data
        event_data = {
            "event_id": str(uuid.uuid4()),
            "timestamp": body.get("timestamp", datetime.utcnow().isoformat()),
            "source": "relay_metrics",
            "severity": "info",
            "category": "performance",
            "data": {
                "bytes_sent": body.get("bytes_sent", 0),
                "bytes_received": body.get("bytes_received", 0),
                "packets_sent": body.get("packets_sent", 0),
                "packets_received": body.get("packets_received", 0),
                "connection_duration": body.get("connection_duration", 0),
                "requests_per_second": body.get("requests_per_second", 0),
                "error_rate": body.get("error_rate", 0.0),
                "metrics": body
            },
            "tags": ["relay_metrics", "performance"],
            "correlation_id": body.get("correlation_id")
        }
        
        # Create event
        event_service = EventService()
        from ...schemas.events import EventCreate
        event_create = EventCreate(**event_data)
        created_event = await event_service.create_event(event_create)
        
        logger.info("Created event from Relay Metrics", event_id=created_event.id)
        
        return {
            "status": "success",
            "event_id": created_event.id,
            "message": "Event created successfully"
        }
        
    except Exception as e:
        logger.error("Error processing Relay Metrics webhook", error=str(e))
        raise HTTPException(status_code=500, detail="Failed to process webhook")


@router.post("/system-logs")
async def system_logs_webhook(request: Request):
    """Webhook endpoint for System Logs events"""
    try:
        # Get request body
        body = await request.json()
        logger.info("Received System Logs webhook", data=body)
        
        # Extract event data
        event_data = {
            "event_id": str(uuid.uuid4()),
            "timestamp": body.get("timestamp", datetime.utcnow().isoformat()),
            "source": "system_logs",
            "severity": body.get("severity", "info"),
            "category": "system",
            "data": {
                "log_level": body.get("log_level", "info"),
                "message": body.get("message", ""),
                "component": body.get("component", ""),
                "context": body.get("context", {})
            },
            "tags": ["system_logs"] + body.get("tags", []),
            "correlation_id": body.get("correlation_id")
        }
        
        # Create event
        event_service = EventService()
        from ...schemas.events import EventCreate
        event_create = EventCreate(**event_data)
        created_event = await event_service.create_event(event_create)
        
        logger.info("Created event from System Logs", event_id=created_event.id)
        
        return {
            "status": "success",
            "event_id": created_event.id,
            "message": "Event created successfully"
        }
        
    except Exception as e:
        logger.error("Error processing System Logs webhook", error=str(e))
        raise HTTPException(status_code=500, detail="Failed to process webhook")


@router.get("/health")
async def webhook_health():
    """Health check for webhook endpoints"""
    return {
        "status": "healthy",
        "endpoints": [
            "/cloudflare-agent",
            "/ml-api", 
            "/relay-metrics",
            "/system-logs"
        ],
        "timestamp": datetime.utcnow().isoformat()
    } 