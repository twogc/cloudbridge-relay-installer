"""
Authentication and authorization service for multi-tenant support
"""
from typing import Optional
from fastapi import HTTPException, Depends, Request
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
import jwt
from datetime import datetime, timedelta

from ..config import settings

security = HTTPBearer()


class AuthService:
    """Authentication service for multi-tenant support"""
    
    def __init__(self):
        self.secret_key = settings.jwt_secret_key
        self.algorithm = "HS256"
        self.access_token_expire_minutes = 30
    
    def create_access_token(self, data: dict, expires_delta: Optional[timedelta] = None):
        """Create JWT access token"""
        to_encode = data.copy()
        if expires_delta:
            expire = datetime.utcnow() + expires_delta
        else:
            expire = datetime.utcnow() + timedelta(minutes=self.access_token_expire_minutes)
        
        to_encode.update({"exp": expire})
        encoded_jwt = jwt.encode(to_encode, self.secret_key, algorithm=self.algorithm)
        return encoded_jwt
    
    def verify_token(self, token: str) -> dict:
        """Verify JWT token and return payload"""
        try:
            payload = jwt.decode(token, self.secret_key, algorithms=[self.algorithm])
            return payload
        except jwt.ExpiredSignatureError:
            raise HTTPException(status_code=401, detail="Token has expired")
        except jwt.JWTError:
            raise HTTPException(status_code=401, detail="Could not validate credentials")
    
    def get_tenant_from_token(self, token: str) -> str:
        """Extract tenant ID from JWT token"""
        payload = self.verify_token(token)
        tenant_id = payload.get("tenant_id")
        if not tenant_id:
            raise HTTPException(status_code=401, detail="Invalid token: no tenant_id")
        return tenant_id
    
    def get_user_from_token(self, token: str) -> dict:
        """Extract user information from JWT token"""
        payload = self.verify_token(token)
        user_id = payload.get("user_id")
        if not user_id:
            raise HTTPException(status_code=401, detail="Invalid token: no user_id")
        return {
            "user_id": user_id,
            "tenant_id": payload.get("tenant_id"),
            "roles": payload.get("roles", []),
            "permissions": payload.get("permissions", [])
        }


async def get_current_tenant(request: Request) -> str:
    """Dependency to get current tenant from request"""
    # For development, we'll use a header-based approach
    # In production, this should use proper JWT tokens from Keycloak
    
    # Check for tenant header (for development)
    tenant_id = request.headers.get("X-Tenant-ID")
    if tenant_id:
        return tenant_id
    
    # Check for authorization header
    auth_header = request.headers.get("Authorization")
    if auth_header and auth_header.startswith("Bearer "):
        token = auth_header.split(" ")[1]
        auth_service = AuthService()
        return auth_service.get_tenant_from_token(token)
    
    # Default tenant for development
    if settings.debug:
        return "default"
    
    raise HTTPException(status_code=401, detail="Tenant ID required")


async def get_current_user(request: Request) -> dict:
    """Dependency to get current user from request"""
    auth_header = request.headers.get("Authorization")
    if not auth_header or not auth_header.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Valid authorization header required")
    
    token = auth_header.split(" ")[1]
    auth_service = AuthService()
    return auth_service.get_user_from_token(token)


def require_permission(permission: str):
    """Decorator to require specific permission"""
    def decorator(func):
        async def wrapper(*args, **kwargs):
            # This would check user permissions
            # For now, we'll just pass through
            return await func(*args, **kwargs)
        return wrapper
    return decorator


def require_role(role: str):
    """Decorator to require specific role"""
    def decorator(func):
        async def wrapper(*args, **kwargs):
            # This would check user roles
            # For now, we'll just pass through
            return await func(*args, **kwargs)
        return wrapper
    return decorator 