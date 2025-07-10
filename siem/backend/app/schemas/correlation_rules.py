"""
Correlation rules schemas for SIEM API
"""
from datetime import datetime
from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field
import uuid


class CorrelationRuleBase(BaseModel):
    """Base correlation rule schema"""
    name: str = Field(..., description="Rule name")
    description: Optional[str] = Field(None, description="Rule description")
    enabled: bool = Field(True, description="Whether rule is enabled")
    conditions: List[Dict[str, Any]] = Field(..., description="Rule conditions")
    actions: List[Dict[str, Any]] = Field(..., description="Rule actions")
    priority: int = Field(1, ge=1, le=10, description="Rule priority (1-10)")
    tags: List[str] = Field(default_factory=list, description="Rule tags")


class CorrelationRuleCreate(CorrelationRuleBase):
    """Schema for creating correlation rule"""
    pass


class CorrelationRuleUpdate(BaseModel):
    """Schema for updating correlation rule"""
    name: Optional[str] = None
    description: Optional[str] = None
    enabled: Optional[bool] = None
    conditions: Optional[List[Dict[str, Any]]] = None
    actions: Optional[List[Dict[str, Any]]] = None
    priority: Optional[int] = Field(None, ge=1, le=10)
    tags: Optional[List[str]] = None


class CorrelationRuleResponse(CorrelationRuleBase):
    """Schema for correlation rule response"""
    id: str = Field(..., description="Rule ID")
    created_at: datetime = Field(..., description="Creation timestamp")
    updated_at: datetime = Field(..., description="Last update timestamp")
    execution_count: int = Field(0, description="Number of times rule was executed")
    last_execution: Optional[datetime] = Field(None, description="Last execution timestamp")

    class Config:
        from_attributes = True 