#!/bin/bash
# CloudBridge Relay - Docker Build Script
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
REGISTRY=${REGISTRY:-"ghcr.io/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^/]*\/[^/]*\)\.git.*/\1/')"}
WITH_WIREGUARD_GO=${WITH_WIREGUARD_GO:-false}

echo -e "${BLUE}🐳 CloudBridge Relay - Docker Build Script${NC}"
echo -e "${BLUE}==========================================${NC}"
echo -e "Version: ${GREEN}$VERSION${NC}"
echo -e "Build Time: ${GREEN}$BUILD_TIME${NC}"
echo -e "Git Commit: ${GREEN}$GIT_COMMIT${NC}"
echo -e "Registry: ${GREEN}$REGISTRY${NC}"
echo -e "WireGuard Go: ${GREEN}$WITH_WIREGUARD_GO${NC}"
echo ""

# Function to build Docker image
build_image() {
    local service=$1
    local dockerfile=$2
    local image_name=$3
    
    echo -e "${YELLOW}🔨 Building $service service...${NC}"
    
    # Check if Dockerfile exists
    if [[ ! -f "$dockerfile" ]]; then
        echo -e "${RED}Dockerfile $dockerfile not found, skipping${NC}"
        return 1
    fi
    
    # Build the image
    docker build \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        --build-arg GIT_COMMIT="$GIT_COMMIT" \
        --build-arg WITH_WIREGUARD_GO="$WITH_WIREGUARD_GO" \
        -f "$dockerfile" \
        -t "$image_name:$VERSION" \
        -t "$image_name:latest" \
        .
    
    echo -e "${GREEN}$service service built successfully${NC}"
    echo -e "   Image: $image_name:$VERSION"
    echo -e "   Image: $image_name:latest"
    echo ""
}

# Function to push Docker image
push_image() {
    local image_name=$1
    
    echo -e "${YELLOW}📤 Pushing $image_name...${NC}"
    
    # Push versioned tag
    docker push "$image_name:$VERSION"
    
    # Push latest tag
    docker push "$image_name:latest"
    
    echo -e "${GREEN}$image_name pushed successfully${NC}"
    echo ""
}

# Parse command line arguments
SERVICES=()
PUSH=false
CLEAN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --service)
            SERVICES+=("$2")
            shift 2
            ;;
        --push)
            PUSH=true
            shift
            ;;
        --clean)
            CLEAN=true
            shift
            ;;
        --registry)
            REGISTRY="$2"
            shift 2
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --with-wireguard-go)
            WITH_WIREGUARD_GO=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --service SERVICE    Build specific service (relay, ai, cadence, quic, wireguard, all)"
            echo "  --push              Push images to registry after building"
            echo "  --clean             Clean up old images before building"
            echo "  --registry REGISTRY Set custom registry (default: ghcr.io/owner/repo)"
            echo "  --version VERSION   Set custom version (default: git describe)"
            echo "  --with-wireguard-go Include WireGuard Go userspace binary"
            echo "  --help              Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0 --service relay --push"
            echo "  $0 --service all --clean"
            echo "  $0 --registry myregistry.com/cloudbridge --version v1.0.0"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Default to all services if none specified
if [[ ${#SERVICES[@]} -eq 0 ]]; then
    SERVICES=("all")
fi

# Clean up old images if requested
if [[ "$CLEAN" == "true" ]]; then
    echo -e "${YELLOW}🧹 Cleaning up old images...${NC}"
    docker system prune -f
    echo ""
fi

# Build images based on services
for service in "${SERVICES[@]}"; do
    case $service in
        "all")
            build_image "relay" "Dockerfile.relay.enhanced" "$REGISTRY/cloudbridge-relay"
            build_image "ai" "Dockerfile.ai.enhanced" "$REGISTRY/cloudbridge-ai"
            build_image "cadence" "Dockerfile.cadence.enhanced" "$REGISTRY/cloudbridge-cadence"
            build_image "quic" "Dockerfile.quic.enhanced" "$REGISTRY/cloudbridge-quic"
            build_image "wireguard" "Dockerfile.wireguard.enhanced" "$REGISTRY/cloudbridge-wireguard"
            ;;
        "relay")
            build_image "relay" "Dockerfile.relay.enhanced" "$REGISTRY/cloudbridge-relay"
            ;;
        "ai")
            build_image "ai" "Dockerfile.ai.enhanced" "$REGISTRY/cloudbridge-ai"
            ;;
        "cadence")
            build_image "cadence" "Dockerfile.cadence.enhanced" "$REGISTRY/cloudbridge-cadence"
            ;;
        "quic")
            build_image "quic" "Dockerfile.quic.enhanced" "$REGISTRY/cloudbridge-quic"
            ;;
        "wireguard")
            build_image "wireguard" "Dockerfile.wireguard.enhanced" "$REGISTRY/cloudbridge-wireguard"
            ;;
        *)
            echo -e "${RED}Unknown service: $service${NC}"
            echo "Available services: relay, ai, cadence, quic, wireguard, all"
            exit 1
            ;;
    esac
done

# Push images if requested
if [[ "$PUSH" == "true" ]]; then
    echo -e "${YELLOW}📤 Pushing images to registry...${NC}"
    
    for service in "${SERVICES[@]}"; do
        case $service in
            "all")
                push_image "$REGISTRY/cloudbridge-relay"
                push_image "$REGISTRY/cloudbridge-ai"
                push_image "$REGISTRY/cloudbridge-cadence"
                push_image "$REGISTRY/cloudbridge-quic"
                push_image "$REGISTRY/cloudbridge-wireguard"
                ;;
            "relay")
                push_image "$REGISTRY/cloudbridge-relay"
                ;;
            "ai")
                push_image "$REGISTRY/cloudbridge-ai"
                ;;
            "cadence")
                push_image "$REGISTRY/cloudbridge-cadence"
                ;;
            "quic")
                push_image "$REGISTRY/cloudbridge-quic"
                ;;
            "wireguard")
                push_image "$REGISTRY/cloudbridge-wireguard"
                ;;
        esac
    done
fi

# Show build summary
echo -e "${GREEN}Docker build completed successfully!${NC}"
echo -e "${GREEN}=====================================${NC}"
echo -e "Built services: ${BLUE}${SERVICES[*]}${NC}"
echo -e "Version: ${BLUE}$VERSION${NC}"
echo -e "Registry: ${BLUE}$REGISTRY${NC}"
echo -e "Push: ${BLUE}$PUSH${NC}"
echo ""
echo -e "${YELLOW}Built Images:${NC}"
for service in "${SERVICES[@]}"; do
    case $service in
        "all")
            echo -e "  • $REGISTRY/cloudbridge-relay:$VERSION"
            echo -e "  • $REGISTRY/cloudbridge-ai:$VERSION"
            echo -e "  • $REGISTRY/cloudbridge-cadence:$VERSION"
            echo -e "  • $REGISTRY/cloudbridge-quic:$VERSION"
            echo -e "  • $REGISTRY/cloudbridge-wireguard:$VERSION"
            ;;
        "relay")
            echo -e "  • $REGISTRY/cloudbridge-relay:$VERSION"
            ;;
        "ai")
            echo -e "  • $REGISTRY/cloudbridge-ai:$VERSION"
            ;;
        "cadence")
            echo -e "  • $REGISTRY/cloudbridge-cadence:$VERSION"
            ;;
        "quic")
            echo -e "  • $REGISTRY/cloudbridge-quic:$VERSION"
            ;;
        "wireguard")
            echo -e "  • $REGISTRY/cloudbridge-wireguard:$VERSION"
            ;;
    esac
done
echo ""
echo -e "${GREEN}Ready for deployment!${NC}"

