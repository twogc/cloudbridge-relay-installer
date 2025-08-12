"""
Correlation Rule model for SIEM system
"""
from datetime import datetime
from typing import Optional, List, Dict, Any
from sqlalchemy import Column, String, DateTime, Text, JSON, Index, Boolean
from sqlalchemy.dialects.postgresql import UUID as PostgresUUID
from sqlalchemy.ext.declarative import declarative_base
import uuid

Base = declarative_base()


class CorrelationRule(Base):
    """Correlation Rule model for storing event correlation rules"""
    
    __tablename__ = "correlation_rules"
    
    # Primary key
    id = Column(PostgresUUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    
    # Rule identification
    name = Column(String(255), nullable=False, unique=True)
    description = Column(Text, nullable=True)
    
    # Rule configuration
    conditions = Column(JSON, nullable=False)  # Rule conditions
    actions = Column(JSON, nullable=False)     # Rule actions
    severity = Column(String(20), nullable=False, index=True)
    
    # Rule status
    enabled = Column(Boolean, default=True, nullable=False, index=True)
    
    # Timestamps
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)
    
    # Indexes for performance
    __table_args__ = (
        Index('idx_correlation_rules_enabled', 'enabled'),
        Index('idx_correlation_rules_severity', 'severity'),
    )
    
    def __repr__(self):
        return f"<CorrelationRule(id={self.id}, name='{self.name}', severity='{self.severity}', enabled={self.enabled})>"
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert correlation rule to dictionary"""
        return {
            "id": str(self.id),
            "name": self.name,
            "description": self.description,
            "conditions": self.conditions,
            "actions": self.actions,
            "severity": self.severity,
            "enabled": self.enabled,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "CorrelationRule":
        """Create correlation rule from dictionary"""
        return cls(
            name=data.get("name"),
            description=data.get("description"),
            conditions=data.get("conditions", {}),
            actions=data.get("actions", {}),
            severity=data.get("severity"),
            enabled=data.get("enabled", True),
        )
    
    def enable(self) -> None:
        """Enable the rule"""
        self.enabled = True
        self.updated_at = datetime.utcnow()
    
    def disable(self) -> None:
        """Disable the rule"""
        self.enabled = False
        self.updated_at = datetime.utcnow()
    
    def update_conditions(self, conditions: Dict[str, Any]) -> None:
        """Update rule conditions"""
        self.conditions = conditions
        self.updated_at = datetime.utcnow()
    
    def update_actions(self, actions: Dict[str, Any]) -> None:
        """Update rule actions"""
        self.actions = actions
        self.updated_at = datetime.utcnow() 