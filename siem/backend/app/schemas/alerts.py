"""
Alert schemas for SIEM API
"""
from datetime import datetime
from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field
import uuid


class AlertBase(BaseModel):
    """Base alert schema"""
    alert_id: str = Field(..., description="Unique alert identifier")
    title: str = Field(..., description="Alert title")
    description: Optional[str] = Field(None, description="Alert description")
    severity: str = Field(..., description="Alert severity (high|medium|low|info)")
    status: str = Field(default="open", description="Alert status")
    source: str = Field(..., description="Alert source")
    event_ids: List[uuid.UUID] = Field(default=[], description="Related event IDs")
    correlation_rules: List[str] = Field(default=[], description="Triggered correlation rules")


class AlertCreate(AlertBase):
    """Schema for creating alerts"""
    assigned_to: Optional[uuid.UUID] = Field(None, description="Assigned user ID")


class AlertUpdate(BaseModel):
    """Schema for updating alerts"""
    title: Optional[str] = Field(None, description="Alert title")
    description: Optional[str] = Field(None, description="Alert description")
    severity: Optional[str] = Field(None, description="Alert severity")
    status: Optional[str] = Field(None, description="Alert status")
    event_ids: Optional[List[uuid.UUID]] = Field(None, description="Related event IDs")
    correlation_rules: Optional[List[str]] = Field(None, description="Triggered correlation rules")
    assigned_to: Optional[uuid.UUID] = Field(None, description="Assigned user ID")


class AlertResponse(AlertBase):
    """Schema for alert responses"""
    id: str = Field(..., description="Alert ID")
    assigned_to: Optional[uuid.UUID] = Field(None, description="Assigned user ID")
    created_at: datetime = Field(..., description="Creation timestamp")
    updated_at: datetime = Field(..., description="Last update timestamp")
    resolved_at: Optional[datetime] = Field(None, description="Resolution timestamp")
    
    class Config:
        from_attributes = True


class AlertList(BaseModel):
    """Schema for alert list responses"""
    alerts: List[AlertResponse] = Field(..., description="List of alerts")
    total: int = Field(..., description="Total number of alerts")
    page: int = Field(..., description="Current page")
    size: int = Field(..., description="Page size")
    pages: int = Field(..., description="Total number of pages")


class AlertFilter(BaseModel):
    """Schema for alert filtering"""
    severity: Optional[str] = Field(None, description="Filter by severity")
    status: Optional[str] = Field(None, description="Filter by status")
    source: Optional[str] = Field(None, description="Filter by source")
    assigned_to: Optional[uuid.UUID] = Field(None, description="Filter by assigned user")
    start_time: Optional[datetime] = Field(None, description="Start time filter")
    end_time: Optional[datetime] = Field(None, description="End time filter")
    limit: int = Field(default=100, description="Limit number of results")
    offset: int = Field(default=0, description="Offset for pagination")


class AlertStats(BaseModel):
    """Schema for alert statistics"""
    total_alerts: int = Field(..., description="Total number of alerts")
    open_alerts: int = Field(..., description="Number of open alerts")
    resolved_alerts: int = Field(..., description="Number of resolved alerts")
    high_severity: int = Field(..., description="Number of high severity alerts")
    medium_severity: int = Field(..., description="Number of medium severity alerts")
    low_severity: int = Field(..., description="Number of low severity alerts")
    alerts_by_source: Dict[str, int] = Field(..., description="Alerts grouped by source")
    alerts_by_status: Dict[str, int] = Field(..., description="Alerts grouped by status") 