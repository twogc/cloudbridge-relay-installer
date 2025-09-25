#!/bin/bash
# CloudBridge Relay Build Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build configuration
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse HEAD 2>/dev/null || echo "unknown")}
BUILD_DIR="build"
BIN_DIR="bin"

echo -e "${BLUE}🚀 CloudBridge Relay Build Script${NC}"
echo -e "${BLUE}================================${NC}"
echo -e "Version: ${GREEN}$VERSION${NC}"
echo -e "Build Time: ${GREEN}$BUILD_TIME${NC}"
echo -e "Git Commit: ${GREEN}$GIT_COMMIT${NC}"
echo ""

# Create build directories
echo -e "${YELLOW}📁 Creating build directories...${NC}"
mkdir -p $BUILD_DIR $BIN_DIR

# Clean previous builds
echo -e "${YELLOW}🧹 Cleaning previous builds...${NC}"
rm -rf $BUILD_DIR/* $BIN_DIR/*

# Download dependencies
echo -e "${YELLOW}📦 Downloading Go dependencies...${NC}"
go mod download
go mod verify

# Run tests
echo -e "${YELLOW}🧪 Running tests...${NC}"
CGO_ENABLED=1 go test -v -race -coverprofile=coverage.txt ./...

# Build main relay binary
echo -e "${YELLOW}🔨 Building relay binary...${NC}"
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" \
    -o $BIN_DIR/relay \
    ./cmd/relay

# Build other services if they exist
for service in ai cadence quic wireguard; do
    if [ -d "cmd/$service" ]; then
        echo -e "${YELLOW}🔨 Building $service service...${NC}"
        CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
            -ldflags "-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" \
            -o $BIN_DIR/$service \
            ./cmd/$service
    fi
done

# Copy configuration files
echo -e "${YELLOW}Copying configuration files...${NC}"
cp -r configs $BUILD_DIR/
cp -r scripts $BUILD_DIR/ 2>/dev/null || true
cp docker-compose.yml $BUILD_DIR/ 2>/dev/null || true
cp Dockerfile* $BUILD_DIR/ 2>/dev/null || true

# Copy binaries
echo -e "${YELLOW}📦 Copying binaries...${NC}"
cp -r $BIN_DIR $BUILD_DIR/

# Create installation script
echo -e "${YELLOW}📝 Creating installation script...${NC}"
cat > $BUILD_DIR/install.sh << 'EOF'
#!/bin/bash
# CloudBridge Relay Installation Script

set -e

echo "🚀 CloudBridge Relay Installation"
echo "================================="

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    echo "❌ Please do not run this script as root"
    exit 1
fi

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "📦 Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    echo "Docker installed. Please log out and log back in to use Docker without sudo."
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "📦 Installing Docker Compose..."
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
fi

# Start services
echo "🚀 Starting CloudBridge Relay services..."
docker-compose up -d

echo "CloudBridge Relay installed and started!"
echo "🌐 Web UI: http://localhost:8088"
echo "📊 Metrics: http://localhost:8084/metrics"
echo " API: http://localhost:9090/health"
EOF

chmod +x $BUILD_DIR/install.sh

# Create README for build
echo -e "${YELLOW}📖 Creating build README...${NC}"
cat > $BUILD_DIR/README.md << EOF
# CloudBridge Relay - Build $VERSION

## Installation

1. Extract this archive
2. Run \`./install.sh\` to install and start services

## Manual Installation

1. Start dependencies:
   \`\`\`bash
   docker-compose up -d
   \`\`\`

2. Run relay server:
   \`\`\`bash
   ./bin/relay --config configs/config.json
   \`\`\`

## Services

- **Relay Server**: Main CloudBridge Relay service
- **Cadence**: Workflow engine
- **Grafana**: Monitoring dashboard
- **Prometheus**: Metrics collection

## Ports

- 9090: Relay API
- 8088: Cadence Web UI
- 3000: Grafana
- 9092: Prometheus

## Build Information

- Version: $VERSION
- Build Time: $BUILD_TIME
- Git Commit: $GIT_COMMIT
EOF

# Create archive
echo -e "${YELLOW}📦 Creating distribution archive...${NC}"
tar -czf $BUILD_DIR/relay-installer.tar.gz -C $BUILD_DIR .

# Show build summary
echo ""
echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${GREEN}================================${NC}"
echo -e "Build directory: ${BLUE}$BUILD_DIR${NC}"
echo -e "Archive: ${BLUE}$BUILD_DIR/relay-installer.tar.gz${NC}"
echo -e "Binaries: ${BLUE}$(ls $BIN_DIR/)${NC}"
echo ""
echo -e "${YELLOW}Build Summary:${NC}"
echo -e "  Version: $VERSION"
echo -e "  Build Time: $BUILD_TIME"
echo -e "  Git Commit: $GIT_COMMIT"
echo -e "  Archive Size: $(du -h $BUILD_DIR/relay-installer.tar.gz | cut -f1)"
echo ""
echo -e "${GREEN}🚀 Ready for deployment!${NC}"
