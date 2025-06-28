#!/bin/bash

# CloudBridge Relay - Production Startup Wrapper
# Запускается от root и переключается на ubuntu

set -e

exec su - ubuntu -c "cd /home/ubuntu/cloudbridge-relay-installer && ./scripts/start_production.sh --env .env.production" 