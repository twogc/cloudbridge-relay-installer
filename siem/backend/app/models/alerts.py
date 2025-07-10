"""
Alert model for SIEM system
"""
from datetime import datetime
from typing import Optional, List, Dict, Any
from sqlalchemy import Column, String, DateTime, Text, JSON, Index, Boolean
from sqlalchemy.dialects.postgresql import UUID as PostgresUUID, ARRAY
from sqlalchemy.ext.declarative import declarative_base
import uuid

Base = declarative_base()


class Alert(Base):
    """Alert model for storing security alerts"""
    
    __tablename__ = "alerts"
    
    # Primary key
    id = Column(PostgresUUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    
    # Alert identification
    alert_id = Column(String(255), unique=True, nullable=False, index=True)
    
    # Alert metadata
    title = Column(String(255), nullable=False)
    description = Column(Text, nullable=True)
    severity = Column(String(20), nullable=False, index=True)
    status = Column(String(20), default="open", nullable=False, index=True)
    source = Column(String(100), nullable=False, index=True)
    
    # Related events
    event_ids = Column(ARRAY(PostgresUUID(as_uuid=True)), default=[])
    
    # Correlation rules
    correlation_rules = Column(ARRAY(String), default=[])
    
    # Assignment
    assigned_to = Column(PostgresUUID(as_uuid=True), nullable=True, index=True)
    
    # Timestamps
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)
    resolved_at = Column(DateTime, nullable=True)
    
    # Indexes for performance
    __table_args__ = (
        Index('idx_alerts_severity_status', 'severity', 'status'),
        Index('idx_alerts_source_status', 'source', 'status'),
        Index('idx_alerts_created_at', 'created_at'),
    )
    
    def __repr__(self):
        return f"<Alert(id={self.id}, alert_id='{self.alert_id}', title='{self.title}', severity='{self.severity}')>"
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert alert to dictionary"""
        return {
            "id": str(self.id),
            "alert_id": self.alert_id,
            "title": self.title,
            "description": self.description,
            "severity": self.severity,
            "status": self.status,
            "source": self.source,
            "event_ids": [str(event_id) for event_id in (self.event_ids or [])],
            "correlation_rules": self.correlation_rules or [],
            "assigned_to": str(self.assigned_to) if self.assigned_to else None,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
            "resolved_at": self.resolved_at.isoformat() if self.resolved_at else None,
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Alert":
        """Create alert from dictionary"""
        return cls(
            alert_id=data.get("alert_id"),
            title=data.get("title"),
            description=data.get("description"),
            severity=data.get("severity"),
            status=data.get("status", "open"),
            source=data.get("source"),
            event_ids=[uuid.UUID(event_id) for event_id in data.get("event_ids", [])],
            correlation_rules=data.get("correlation_rules", []),
            assigned_to=uuid.UUID(data["assigned_to"]) if data.get("assigned_to") else None,
        )
    
    def resolve(self) -> None:
        """Mark alert as resolved"""
        self.status = "resolved"
        self.resolved_at = datetime.utcnow()
    
    def assign(self, user_id: uuid.UUID) -> None:
        """Assign alert to user"""
        self.assigned_to = user_id
        self.updated_at = datetime.utcnow() 