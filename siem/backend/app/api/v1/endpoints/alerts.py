"""
Alerts API endpoints
"""
from typing import List, Optional
from fastapi import APIRouter, HTTPException, Query
import uuid

from app.schemas.alerts import AlertCreate, AlertUpdate, AlertResponse, AlertFilter, AlertStats
from app.services.alerts import AlertService

router = APIRouter()


@router.post("/", response_model=AlertResponse)
async def create_alert(alert: AlertCreate):
    """Create a new alert"""
    try:
        alert_service = AlertService()
        return await alert_service.create_alert(alert)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/", response_model=List[AlertResponse])
async def get_alerts(
    severity: Optional[str] = Query(None, description="Filter by severity"),
    status: Optional[str] = Query(None, description="Filter by status"),
    source: Optional[str] = Query(None, description="Filter by source"),
    assigned_to: Optional[str] = Query(None, description="Filter by assigned user"),
    limit: int = Query(100, ge=1, le=1000, description="Limit number of results"),
    offset: int = Query(0, ge=0, description="Offset for pagination"),
):
    """Get alerts with filtering and pagination"""
    try:
        alert_service = AlertService()
        alerts = await alert_service.get_alerts(
            skip=offset,
            limit=limit,
            severity=severity,
            status=status,
            source=source,
        )
        return alerts
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/{alert_id}", response_model=AlertResponse)
async def get_alert(alert_id: str):
    """Get a specific alert by ID"""
    try:
        alert_service = AlertService()
        alert = await alert_service.get_alert_by_id(alert_id)
        if not alert:
            raise HTTPException(status_code=404, detail="Alert not found")
        return alert
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.put("/{alert_id}", response_model=AlertResponse)
async def update_alert(alert_id: str, alert_update: AlertUpdate):
    """Update an alert"""
    try:
        alert_service = AlertService()
        alert = await alert_service.update_alert(alert_id, alert_update)
        if not alert:
            raise HTTPException(status_code=404, detail="Alert not found")
        return alert
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.delete("/{alert_id}")
async def delete_alert(alert_id: str):
    """Delete an alert"""
    try:
        alert_service = AlertService()
        success = await alert_service.delete_alert(alert_id)
        if not success:
            raise HTTPException(status_code=404, detail="Alert not found")
        return {"message": "Alert deleted successfully"}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/{alert_id}/resolve")
async def resolve_alert(alert_id: str):
    """Resolve an alert"""
    try:
        alert_service = AlertService()
        success = await alert_service.resolve_alert(alert_id)
        if not success:
            raise HTTPException(status_code=404, detail="Alert not found")
        return {"message": "Alert resolved successfully"}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/{alert_id}/assign")
async def assign_alert(alert_id: str, user_id: str):
    """Assign an alert to a user"""
    try:
        alert_service = AlertService()
        success = await alert_service.assign_alert(alert_id, user_id)
        if not success:
            raise HTTPException(status_code=404, detail="Alert not found")
        return {"message": "Alert assigned successfully"}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/stats/summary", response_model=AlertStats)
async def get_alert_stats():
    """Get alert statistics"""
    try:
        alert_service = AlertService()
        # TODO: Implement statistics calculation
        return AlertStats(
            total_alerts=0,
            open_alerts=0,
            resolved_alerts=0,
            high_severity=0,
            medium_severity=0,
            low_severity=0,
            alerts_by_source={},
            alerts_by_status={},
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e)) 