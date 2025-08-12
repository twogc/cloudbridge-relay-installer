"""
Events API endpoints
"""
from typing import List, Optional
from fastapi import APIRouter, HTTPException, Query, Depends, Request
from prometheus_client import Counter, Histogram
import uuid
import time

from app.schemas.events import EventCreate, EventUpdate, EventResponse
from app.services.events import EventService
from app.services.auth import get_current_tenant

# Prometheus metrics
EVENTS_CREATED = Counter('siem_events_created_total', 'Total events created', ['tenant_id', 'source', 'severity'])
EVENTS_RETRIEVED = Counter('siem_events_retrieved_total', 'Total events retrieved', ['tenant_id'])
EVENT_PROCESSING_TIME = Histogram('siem_event_processing_seconds', 'Event processing time', ['operation'])

router = APIRouter()


@router.post("/", response_model=EventResponse)
async def create_event(
    event: EventCreate,
    tenant_id: str = Depends(get_current_tenant)
):
    """Create a new event"""
    start_time = time.time()
    try:
        event_service = EventService()
        # Add tenant_id to event
        event.tenant_id = tenant_id
        created_event = await event_service.create_event(event)
        
        # Record metrics
        EVENTS_CREATED.labels(
            tenant_id=tenant_id,
            source=event.source,
            severity=event.severity
        ).inc()
        
        return created_event
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
    finally:
        EVENT_PROCESSING_TIME.labels(operation="create_event").observe(time.time() - start_time)


@router.get("/", response_model=List[EventResponse])
async def get_events(
    source: Optional[str] = Query(None, description="Filter by source"),
    severity: Optional[str] = Query(None, description="Filter by severity"),
    category: Optional[str] = Query(None, description="Filter by category"),
    limit: int = Query(100, ge=1, le=1000, description="Limit number of results"),
    offset: int = Query(0, ge=0, description="Offset for pagination"),
    tenant_id: str = Depends(get_current_tenant),
):
    """Get events with filtering and pagination"""
    start_time = time.time()
    try:
        event_service = EventService()
        events = await event_service.get_events(
            tenant_id=tenant_id,
            source=source,
            severity=severity,
            category=category,
            limit=limit,
            offset=offset,
        )
        
        # Record metrics
        EVENTS_RETRIEVED.labels(tenant_id=tenant_id).inc()
        
        return events
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
    finally:
        EVENT_PROCESSING_TIME.labels(operation="get_events").observe(time.time() - start_time)


@router.get("/{event_id}", response_model=EventResponse)
async def get_event(
    event_id: str,
    tenant_id: str = Depends(get_current_tenant)
):
    """Get a specific event by ID"""
    try:
        event_service = EventService()
        event = await event_service.get_event_by_id(event_id, tenant_id=tenant_id)
        if not event:
            raise HTTPException(status_code=404, detail="Event not found")
        return event
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.put("/{event_id}", response_model=EventResponse)
async def update_event(
    event_id: str, 
    event_update: EventUpdate,
    tenant_id: str = Depends(get_current_tenant)
):
    """Update an event"""
    try:
        event_service = EventService()
        event = await event_service.update_event(event_id, event_update, tenant_id=tenant_id)
        if not event:
            raise HTTPException(status_code=404, detail="Event not found")
        return event
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.delete("/{event_id}")
async def delete_event(
    event_id: str,
    tenant_id: str = Depends(get_current_tenant)
):
    """Delete an event"""
    try:
        event_service = EventService()
        success = await event_service.delete_event(event_id, tenant_id=tenant_id)
        if not success:
            raise HTTPException(status_code=404, detail="Event not found")
        return {"message": "Event deleted successfully"}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/{event_id}/correlated")
async def get_correlated_events(
    event_id: str,
    tenant_id: str = Depends(get_current_tenant)
):
    """Get events correlated with the specified event"""
    try:
        event_service = EventService()
        correlated_events = await event_service.get_correlated_events(event_id, tenant_id=tenant_id)
        return {"correlated_events": correlated_events}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/bulk")
async def create_events_bulk(
    events: List[EventCreate],
    tenant_id: str = Depends(get_current_tenant)
):
    """Create multiple events in bulk"""
    try:
        event_service = EventService()
        # Add tenant_id to all events
        for event in events:
            event.tenant_id = tenant_id
        created_events = await event_service.create_events_bulk(events)
        return {"created_events": len(created_events), "events": created_events}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e)) 