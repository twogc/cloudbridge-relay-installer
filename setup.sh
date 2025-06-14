#!/bin/bash

# Create necessary directories
mkdir -p prometheus
mkdir -p config

# Set proper permissions
chmod 755 prometheus
chmod 755 config

# Create prometheus directory if it doesn't exist
if [ ! -f prometheus/prometheus.yml ]; then
    echo "Creating prometheus configuration..."
    cp prometheus.yml prometheus/
fi

# Start the services
docker-compose up -d

echo "Setup complete! Services are starting..."
echo "Prometheus is available at: http://localhost:9090"
echo "Relay server metrics are available at: http://localhost:8080/metrics"

sudo docker stop cloudbridge-relay-installer_relay_1 