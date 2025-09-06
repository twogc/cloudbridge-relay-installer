-- CloudBridge Relay Domain Initialization for Cassandra
-- This script initializes the CloudBridge Relay domain in Cadence

-- Insert into domains table for CloudBridge Relay
INSERT INTO domains (
    id,
    domain
) VALUES (
    550e8400-e29b-41d4-a716-446655440000,
    textAsBlob('{"id":"550e8400-e29b-41d4-a716-446655440000","name":"cloudbridge","status":0,"description":"CloudBridge Relay P2P Mesh and Workflow Management Domain","data":{"project":"cloudbridge-relay","version":"1.0.0","environment":"production","features":["p2p-mesh","wireguard","quic","workflow-engine"],"owner":"2GC Technologies","contact":"support@2gc.uk"},"owner_email":"support@2gc.uk"}')
) IF NOT EXISTS;

-- Insert into domains_by_name_v2 table for CloudBridge Relay
INSERT INTO domains_by_name_v2 (
    name,
    domain,
    config,
    replication_config,
    is_global_domain,
    config_version,
    failover_version,
    failover_notification_version,
    notification_version
) VALUES (
    'cloudbridge',
    textAsBlob('{"id":"550e8400-e29b-41d4-a716-446655440000","name":"cloudbridge","status":0,"description":"CloudBridge Relay P2P Mesh and Workflow Management Domain","data":{"project":"cloudbridge-relay","version":"1.0.0","environment":"production","features":["p2p-mesh","wireguard","quic","workflow-engine"],"owner":"2GC Technologies","contact":"support@2gc.uk"},"owner_email":"support@2gc.uk"}'),
    textAsBlob('{"retention":30,"emit_metric":true,"history_archival_status":1,"visibility_archival_status":1}'),
    textAsBlob('{"active_cluster_name":"cloudbridge-cluster","clusters":[{"cluster_name":"cloudbridge-cluster"},{"cluster_name":"cloudbridge-backup-cluster"}]}'),
    true,
    1,
    0,
    0,
    0
) IF NOT EXISTS;
