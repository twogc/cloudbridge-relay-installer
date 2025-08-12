"""
Event schemas for SIEM API
"""
from datetime import datetime
from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field
import uuid


class EventBase(BaseModel):
    """Base event schema"""
    event_id: str = Field(..., description="Unique event identifier")
    timestamp: datetime = Field(..., description="Event timestamp")
    source: str = Field(..., description="Event source")
    severity: str = Field(..., description="Event severity (high|medium|low|info)")
    category: str = Field(..., description="Event category")
    data: Dict[str, Any] = Field(..., description="Event data")
    tags: List[str] = Field(default=[], description="Event tags")


class EventCreate(EventBase):
    """Schema for creating events"""
    tenant_id: Optional[str] = Field(None, description="Tenant ID")
    correlation_id: Optional[uuid.UUID] = Field(None, description="Correlation ID")


class EventUpdate(BaseModel):
    """Schema for updating events"""
    severity: Optional[str] = Field(None, description="Event severity")
    category: Optional[str] = Field(None, description="Event category")
    data: Optional[Dict[str, Any]] = Field(None, description="Event data")
    tags: Optional[List[str]] = Field(None, description="Event tags")
    correlation_id: Optional[uuid.UUID] = Field(None, description="Correlation ID")


class EventResponse(EventBase):
    """Schema for event responses"""
    id: str = Field(..., description="Event ID")
    tenant_id: Optional[str] = Field(None, description="Tenant ID")
    correlation_id: Optional[uuid.UUID] = Field(None, description="Correlation ID")
    created_at: datetime = Field(..., description="Creation timestamp")
    updated_at: datetime = Field(..., description="Last update timestamp")
    
    class Config:
        from_attributes = True


class EventList(BaseModel):
    """Schema for event list responses"""
    events: List[EventResponse] = Field(..., description="List of events")
    total: int = Field(..., description="Total number of events")
    page: int = Field(..., description="Current page")
    size: int = Field(..., description="Page size")
    pages: int = Field(..., description="Total number of pages")


class EventFilter(BaseModel):
    """Schema for event filtering"""
    source: Optional[str] = Field(None, description="Filter by source")
    severity: Optional[str] = Field(None, description="Filter by severity")
    category: Optional[str] = Field(None, description="Filter by category")
    tags: Optional[List[str]] = Field(None, description="Filter by tags")
    start_time: Optional[datetime] = Field(None, description="Start time filter")
    end_time: Optional[datetime] = Field(None, description="End time filter")
    correlation_id: Optional[uuid.UUID] = Field(None, description="Filter by correlation ID")
    limit: int = Field(default=100, description="Limit number of results")
    offset: int = Field(default=0, description="Offset for pagination") 