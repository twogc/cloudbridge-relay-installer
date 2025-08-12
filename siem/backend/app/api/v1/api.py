"""
Main API router for v1
"""
from fastapi import APIRouter

from .endpoints import events, alerts, correlation_rules, users, webhooks, analytics

api_router = APIRouter()

# Include all endpoint routers
api_router.include_router(events.router, prefix="/events", tags=["events"])
api_router.include_router(alerts.router, prefix="/alerts", tags=["alerts"])
api_router.include_router(correlation_rules.router, prefix="/correlation-rules", tags=["correlation-rules"])
api_router.include_router(users.router, prefix="/users", tags=["users"])
api_router.include_router(webhooks.router, prefix="/webhooks", tags=["webhooks"])
api_router.include_router(analytics.router, prefix="/analytics", tags=["analytics"]) 