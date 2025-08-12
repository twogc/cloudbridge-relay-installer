"""
Event model for SIEM system
"""
from datetime import datetime
from typing import Optional, List, Dict, Any
from sqlalchemy import Column, String, DateTime, Text, JSON, Index, UUID
from sqlalchemy.dialects.postgresql import UUID as PostgresUUID, ARRAY
from sqlalchemy.ext.declarative import declarative_base
import uuid

Base = declarative_base()


class Event(Base):
    """Event model for storing security events"""
    
    __tablename__ = "events"
    
    # Primary key
    id = Column(PostgresUUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    
    # Event identification
    event_id = Column(String(255), unique=True, nullable=False, index=True)
    
    # Event metadata
    timestamp = Column(DateTime, nullable=False, index=True)
    source = Column(String(100), nullable=False, index=True)
    severity = Column(String(20), nullable=False, index=True)
    category = Column(String(50), nullable=False, index=True)
    
    # Event data
    data = Column(JSON, nullable=False)
    tags = Column(ARRAY(String), default=[])
    
    # Correlation
    correlation_id = Column(PostgresUUID(as_uuid=True), nullable=True, index=True)
    
    # Timestamps
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)
    
    # Indexes for performance
    __table_args__ = (
        Index('idx_events_timestamp_source', 'timestamp', 'source'),
        Index('idx_events_severity_category', 'severity', 'category'),
        Index('idx_events_correlation_id', 'correlation_id'),
    )
    
    def __repr__(self):
        return f"<Event(id={self.id}, event_id='{self.event_id}', source='{self.source}', severity='{self.severity}')>"
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert event to dictionary"""
        return {
            "id": str(self.id),
            "event_id": self.event_id,
            "timestamp": self.timestamp.isoformat() if self.timestamp else None,
            "source": self.source,
            "severity": self.severity,
            "category": self.category,
            "data": self.data,
            "tags": self.tags or [],
            "correlation_id": str(self.correlation_id) if self.correlation_id else None,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Event":
        """Create event from dictionary"""
        return cls(
            event_id=data.get("event_id"),
            timestamp=datetime.fromisoformat(data["timestamp"]) if data.get("timestamp") else datetime.utcnow(),
            source=data.get("source"),
            severity=data.get("severity"),
            category=data.get("category"),
            data=data.get("data", {}),
            tags=data.get("tags", []),
            correlation_id=uuid.UUID(data["correlation_id"]) if data.get("correlation_id") else None,
        ) 