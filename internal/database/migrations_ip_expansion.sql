-- Migration: IP Space Expansion (10.100.0.0/16 → 10.0.0.0/8)
-- Date: 27 сентября 2025
-- Purpose: Expand IPAM from 256 subnets to 65,536 subnets

-- Update existing tenant networks to new IP range
UPDATE tenant_networks 
SET network_cidr = '10.0.77.0/24'
WHERE tenant_id = 'tenant-216420165' AND network_cidr = '10.100.77.0/24';

UPDATE tenant_networks 
SET network_cidr = '10.0.78.0/24'
WHERE tenant_id = 'tenant-123456789' AND network_cidr = '10.100.78.0/24';

UPDATE tenant_networks 
SET network_cidr = '10.0.79.0/24'
WHERE tenant_id = 'tenant-987654321' AND network_cidr = '10.100.79.0/24';

-- Update subnet_counter to reflect new range
-- Reset to 79 to continue from where we left off
UPDATE tenant_networks 
SET subnet_counter = 79
WHERE tenant_id IN ('tenant-216420165', 'tenant-123456789', 'tenant-987654321');

-- Add index for better performance with expanded IP range
CREATE INDEX IF NOT EXISTS idx_tenant_networks_network_cidr 
ON tenant_networks(network_cidr);

-- Add constraint to ensure valid IP ranges (10.0.0.0/8)
ALTER TABLE tenant_networks 
ADD CONSTRAINT tenant_networks_network_cidr_range_check 
CHECK (network_cidr ~ '^10\.\d{1,3}\.\d{1,3}\.0/24$');

-- Log migration completion
INSERT INTO tenant_networks (tenant_id, network_cidr, subnet_counter, is_active, created_at, updated_at)
VALUES ('migration-ip-expansion', '10.0.0.0/8', 0, false, NOW(), NOW())
ON CONFLICT (tenant_id) DO NOTHING;

