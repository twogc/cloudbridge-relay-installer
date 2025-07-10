-- SIEM Database Initialization Script
-- This script creates the necessary tables and indexes for the SIEM system

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create events table
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id VARCHAR(255) UNIQUE NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    source VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('high', 'medium', 'low', 'info')),
    category VARCHAR(50) NOT NULL,
    data JSONB NOT NULL,
    tags TEXT[],
    correlation_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for events table
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_events_source ON events(source);
CREATE INDEX IF NOT EXISTS idx_events_severity ON events(severity);
CREATE INDEX IF NOT EXISTS idx_events_category ON events(category);
CREATE INDEX IF NOT EXISTS idx_events_correlation_id ON events(correlation_id);
CREATE INDEX IF NOT EXISTS idx_events_tags ON events USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_events_data ON events USING GIN(data);

-- Create alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('high', 'medium', 'low', 'info')),
    status VARCHAR(20) DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'resolved', 'closed')),
    source VARCHAR(100) NOT NULL,
    event_ids UUID[],
    correlation_rules TEXT[],
    assigned_to UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

-- Create indexes for alerts table
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_source ON alerts(source);
CREATE INDEX IF NOT EXISTS idx_alerts_assigned_to ON alerts(assigned_to);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);

-- Create correlation_rules table
CREATE TABLE IF NOT EXISTS correlation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    conditions JSONB NOT NULL,
    actions JSONB NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('high', 'medium', 'low', 'info')),
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for correlation_rules table
CREATE INDEX IF NOT EXISTS idx_correlation_rules_enabled ON correlation_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_correlation_rules_severity ON correlation_rules(severity);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    hashed_password VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'analyst', 'user')),
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    last_login TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for users table
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- Create audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(255),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for audit_logs table
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('alert', 'system', 'info', 'warning')),
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for notifications table
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);

-- Create threat_intelligence table
CREATE TABLE IF NOT EXISTS threat_intelligence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    indicator VARCHAR(255) NOT NULL,
    indicator_type VARCHAR(50) NOT NULL CHECK (indicator_type IN ('ip', 'domain', 'url', 'hash', 'email')),
    threat_type VARCHAR(100) NOT NULL,
    confidence_score DECIMAL(3,2) CHECK (confidence_score >= 0 AND confidence_score <= 1),
    source VARCHAR(100) NOT NULL,
    first_seen TIMESTAMPTZ DEFAULT NOW(),
    last_seen TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for threat_intelligence table
CREATE INDEX IF NOT EXISTS idx_threat_intelligence_indicator ON threat_intelligence(indicator);
CREATE INDEX IF NOT EXISTS idx_threat_intelligence_type ON threat_intelligence(indicator_type);
CREATE INDEX IF NOT EXISTS idx_threat_intelligence_threat_type ON threat_intelligence(threat_type);
CREATE INDEX IF NOT EXISTS idx_threat_intelligence_is_active ON threat_intelligence(is_active);

-- Create system_metrics table
CREATE TABLE IF NOT EXISTS system_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(15,4) NOT NULL,
    metric_unit VARCHAR(20),
    tags JSONB,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for system_metrics table
CREATE INDEX IF NOT EXISTS idx_system_metrics_name ON system_metrics(metric_name);
CREATE INDEX IF NOT EXISTS idx_system_metrics_timestamp ON system_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_system_metrics_tags ON system_metrics USING GIN(tags);

-- Create default admin user (password: admin123)
INSERT INTO users (username, email, hashed_password, full_name, role, is_active, is_verified)
VALUES (
    'admin',
    'admin@siem.local',
    '$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK.s5u.Ge', -- admin123
    'System Administrator',
    'admin',
    true,
    true
) ON CONFLICT (username) DO NOTHING;

-- Create default correlation rules
INSERT INTO correlation_rules (name, description, conditions, actions, severity, enabled)
VALUES 
(
    'High Anomaly Score + Multiple Failed Auth',
    'Detects high anomaly scores combined with multiple failed authentication attempts',
    '{"conditions": [{"anomaly_score": {"operator": ">", "value": 0.8}}, {"failed_auth_attempts": {"operator": ">", "value": 5}}, {"time_window": "5 minutes"}]}',
    '{"actions": ["create_alert", "send_notification"], "severity": "high"}',
    'high',
    true
),
(
    'Multiple Security Events from Same Source',
    'Detects multiple security events from the same source within a short time period',
    '{"conditions": [{"event_count": {"operator": ">", "value": 10}}, {"source": "same"}, {"time_window": "10 minutes"}]}',
    '{"actions": ["create_alert", "block_source"], "severity": "medium"}',
    'medium',
    true
),
(
    'Performance Degradation',
    'Detects significant performance degradation across multiple metrics',
    '{"conditions": [{"error_rate": {"operator": ">", "value": 0.1}}, {"response_time": {"operator": ">", "value": 1000}}, {"time_window": "5 minutes"}]}',
    '{"actions": ["create_alert", "scale_resources"], "severity": "medium"}',
    'medium',
    true
) ON CONFLICT DO NOTHING;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_events_updated_at BEFORE UPDATE ON events FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_alerts_updated_at BEFORE UPDATE ON alerts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_correlation_rules_updated_at BEFORE UPDATE ON correlation_rules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_threat_intelligence_updated_at BEFORE UPDATE ON threat_intelligence FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO siem;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO siem;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO siem; 