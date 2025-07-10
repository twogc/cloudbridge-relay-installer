"""
User schemas for SIEM API
"""
from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field, EmailStr
import uuid


class UserBase(BaseModel):
    """Base user schema"""
    username: str = Field(..., description="Username")
    email: EmailStr = Field(..., description="Email address")
    full_name: Optional[str] = Field(None, description="Full name")
    role: str = Field("user", description="User role")
    is_active: bool = Field(True, description="Whether user is active")


class UserCreate(UserBase):
    """Schema for creating user"""
    password: str = Field(..., description="Password")


class UserUpdate(BaseModel):
    """Schema for updating user"""
    username: Optional[str] = None
    email: Optional[EmailStr] = None
    full_name: Optional[str] = None
    role: Optional[str] = None
    is_active: Optional[bool] = None
    password: Optional[str] = None


class UserResponse(UserBase):
    """Schema for user response"""
    id: str = Field(..., description="User ID")
    created_at: datetime = Field(..., description="Creation timestamp")
    updated_at: datetime = Field(..., description="Last update timestamp")
    last_login: Optional[datetime] = Field(None, description="Last login timestamp")

    class Config:
        from_attributes = True 