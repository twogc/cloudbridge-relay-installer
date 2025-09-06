-- Insert into domains table for CloudBridge Relay
INSERT INTO domains (
    id,
    domain
) VALUES (
    550e8400-e29b-41d4-a716-446655440000, -- CloudBridge Relay Domain UUID
    {
        id: 550e8400-e29b-41d4-a716-446655440000, -- CloudBridge Relay Domain UUID
        name: 'cloudbridge',
        status: 0, -- Registered
        description: 'CloudBridge Relay P2P Mesh and Workflow Management Domain',
        data: {
            'project': 'cloudbridge-relay',
            'version': '1.0.0',
            'environment': 'production',
            'features': ['p2p-mesh', 'wireguard', 'quic', 'workflow-engine'],
            'owner': '2GC Technologies',
            'contact': 'support@2gc.eu'
        },
        owner_email: 'support@2gc.eu'
    }
) IF NOT EXISTS;

-- Insert into domains_by_name_v2 table for CloudBridge Relay
INSERT INTO domains_by_name_v2 (
    domains_partition,
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
    0,
    'cloudbridge',
    {
        id: 550e8400-e29b-41d4-a716-446655440000, -- CloudBridge Relay Domain UUID
        name: 'cloudbridge',
        status: 0, -- Registered
        description: 'CloudBridge Relay P2P Mesh and Workflow Management Domain',
        data: {
            'project': 'cloudbridge-relay',
            'version': '1.0.0',
            'environment': 'production',
            'features': ['p2p-mesh', 'wireguard', 'quic', 'workflow-engine'],
            'owner': '2GC Technologies',
            'contact': 'support@2gc.uk'
        },
        owner_email: 'support@2gc.uk'
    },
    {
        retention: 30, -- 30 days retention for production
        emit_metric: True,
        history_archival_status: 1, -- Enabled for production
        visibility_archival_status: 1 -- Enabled for production
    },
    {
        active_cluster_name: 'cloudbridge-cluster',
        clusters: [
            {cluster_name: 'cloudbridge-cluster'},
            {cluster_name: 'cloudbridge-backup-cluster'}
        ]
    },
    True, -- is_global_domain
    1,    -- config_version
    0,    -- failover_version
    0,    -- failover_notification_version
    0     -- notification_version
) IF NOT EXISTS;
