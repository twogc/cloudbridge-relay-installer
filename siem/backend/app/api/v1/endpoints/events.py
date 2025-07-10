"""
Events API endpoints
"""
from typing import List, Optional
from fastapi import APIRouter, HTTPException, Query
import uuid

from ...schemas.events import EventCreate, EventUpdate, EventResponse
from ...services.events import EventService

router = APIRouter()


@router.post("/", response_model=EventResponse)
async def create_event(event: EventCreate):
    """Create a new event"""
    try:
        event_service = EventService()
        return await event_service.create_event(event)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/", response_model=List[EventResponse])
async def get_events(
    source: Optional[str] = Query(None, description="Filter by source"),
    severity: Optional[str] = Query(None, description="Filter by severity"),
    category: Optional[str] = Query(None, description="Filter by category"),
    limit: int = Query(100, ge=1, le=1000, description="Limit number of results"),
    offset: int = Query(0, ge=0, description="Offset for pagination"),
):
    """Get events with filtering and pagination"""
    try:
        event_service = EventService()
        events = await event_service.get_events(
            source=source,
            severity=severity,
            category=category,
            limit=limit,
            offset=offset,
        )
        return events
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/{event_id}", response_model=EventResponse)
async def get_event(event_id: str):
    """Get a specific event by ID"""
    try:
        event_service = EventService()
        event = await event_service.get_event_by_id(event_id)
        if not event:
            raise HTTPException(status_code=404, detail="Event not found")
        return event
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.put("/{event_id}", response_model=EventResponse)
async def update_event(event_id: str, event_update: EventUpdate):
    """Update an event"""
    try:
        event_service = EventService()
        event = await event_service.update_event(event_id, event_update)
        if not event:
            raise HTTPException(status_code=404, detail="Event not found")
        return event
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.delete("/{event_id}")
async def delete_event(event_id: str):
    """Delete an event"""
    try:
        event_service = EventService()
        success = await event_service.delete_event(event_id)
        if not success:
            raise HTTPException(status_code=404, detail="Event not found")
        return {"message": "Event deleted successfully"}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/{event_id}/correlated")
async def get_correlated_events(event_id: str):
    """Get events correlated with the specified event"""
    try:
        event_service = EventService()
        correlated_events = await event_service.get_correlated_events(event_id)
        return {"correlated_events": correlated_events}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/bulk")
async def create_events_bulk(events: List[EventCreate]):
    """Create multiple events in bulk"""
    try:
        event_service = EventService()
        created_events = await event_service.create_events_bulk(events)
        return {"created_events": len(created_events), "events": created_events}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e)) 