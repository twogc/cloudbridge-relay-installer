"""
Configuration settings for SIEM Backend API
"""
from typing import List, Optional
from pydantic_settings import BaseSettings
from pydantic.types import SecretStr


class Settings(BaseSettings):
    """Application settings"""
    
    # Application
    app_name: str = "SIEM Backend API"
    app_version: str = "1.0.0"
    debug: bool = False
    
    # Server
    host: str = "0.0.0.0"
    port: int = 8000
    workers: int = 1
    
    # Database
    database_url: str = "postgresql://siem:password@localhost:5432/siem"
    database_pool_size: int = 10
    database_max_overflow: int = 20
    
    # Redis
    redis_url: str = "redis://localhost:6379/0"
    redis_password: Optional[SecretStr] = None
    
    # Elasticsearch
    elasticsearch_url: str = "http://localhost:9200"
    elasticsearch_username: Optional[str] = None
    elasticsearch_password: Optional[SecretStr] = None
    
    # Security
    secret_key: SecretStr = SecretStr("your-secret-key-change-in-production")
    jwt_secret_key: SecretStr = SecretStr("your-jwt-secret-key-change-in-production")
    algorithm: str = "HS256"
    access_token_expire_minutes: int = 30
    
    # OAuth2
    oauth_client_id: Optional[str] = None
    oauth_client_secret: Optional[SecretStr] = None
    oauth_authorization_url: Optional[str] = None
    oauth_token_url: Optional[str] = None
    
    # External APIs
    cloudflare_agent_url: str = "https://cloudbridge-ai-monitoring-agent.2gc.workers.dev"
    ml_api_url: str = "http://localhost:8000/detect"
    relay_metrics_url: str = "http://localhost:9090/metrics"
    
    # Monitoring
    prometheus_url: str = "http://localhost:9090"
    grafana_url: str = "http://localhost:3000"
    
    # Notifications
    email_smtp_server: Optional[str] = None
    email_smtp_port: int = 587
    email_username: Optional[str] = None
    email_password: Optional[SecretStr] = None
    email_from: Optional[str] = None
    email_to: List[str] = []
    
    telegram_bot_token: Optional[SecretStr] = None
    telegram_chat_id: Optional[str] = None
    
    slack_webhook_url: Optional[str] = None
    
    # Performance
    max_events_per_second: int = 10000
    event_processing_timeout: int = 100
    
    # Logging
    log_level: str = "INFO"
    log_format: str = "json"
    
    # CORS
    cors_origins: List[str] = ["http://localhost:3000", "http://localhost:8081"]
    cors_allow_credentials: bool = True
    cors_allow_methods: List[str] = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    cors_allow_headers: List[str] = ["*"]
    
    class Config:
        env_file = ".env"
        case_sensitive = False


# Global settings instance
settings = Settings() 