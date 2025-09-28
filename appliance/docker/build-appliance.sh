#!/bin/bash
# CloudBridge Appliance Build Script

set -e

# Configuration
APP_NAME="cloudbridge-appliance"
VERSION=${1:-"latest"}
REGISTRY=${2:-"localhost:5000"}

echo "🔧 Building CloudBridge Appliance Docker Image"
echo "Version: $VERSION"
echo "Registry: $REGISTRY"

# Build the appliance image
echo "📦 Building Docker image..."
docker build -f Dockerfile.appliance -t $APP_NAME:$VERSION .

# Tag for registry
docker tag $APP_NAME:$VERSION $REGISTRY/$APP_NAME:$VERSION

# Optional: Push to registry
if [ "$3" = "push" ]; then
    echo "🚀 Pushing to registry..."
    docker push $REGISTRY/$APP_NAME:$VERSION
fi

# Create appliance package
echo "📁 Creating appliance package..."
mkdir -p ../packages/appliance-$VERSION
cp docker-compose.appliance.yml ../packages/appliance-$VERSION/
cp -r ../configs ../packages/appliance-$VERSION/
cp -r ../scripts ../packages/appliance-$VERSION/

# Create installation script
cat > ../packages/appliance-$VERSION/install.sh << 'EOF'
#!/bin/bash
# CloudBridge Appliance Installation Script

set -e

echo "🚀 Installing CloudBridge Appliance..."

# Check prerequisites
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Create directories
mkdir -p ./logs ./data ./configs/ssl

# Set permissions
chmod +x ./scripts/*.sh

# Start services
echo "🔄 Starting CloudBridge Appliance..."
docker-compose -f docker-compose.appliance.yml up -d

# Wait for health check
echo "⏳ Waiting for services to be healthy..."
sleep 30

# Check status
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ CloudBridge Appliance is running successfully!"
    echo "🌐 Admin interface: http://localhost:8080"
    echo "🔧 gRPC endpoint: localhost:8444"
else
    echo "❌ Health check failed. Check logs with: docker-compose logs"
    exit 1
fi
EOF

chmod +x ../packages/appliance-$VERSION/install.sh

echo "✅ Appliance package created: ../packages/appliance-$VERSION/"
echo "📋 To install: cd ../packages/appliance-$VERSION && ./install.sh"
