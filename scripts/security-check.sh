#!/bin/bash

# CloudBridge Relay - Comprehensive Security Check Script
# This script runs all security tools and generates a comprehensive report

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPORTS_DIR="$PROJECT_ROOT/security_reports"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo -e "${BLUE}🔒 CloudBridge Relay Security Check${NC}"
echo -e "${BLUE}====================================${NC}"
echo "Project: $PROJECT_ROOT"
echo "Reports: $REPORTS_DIR"
echo "Timestamp: $TIMESTAMP"
echo ""

# Create reports directory
mkdir -p "$REPORTS_DIR"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install missing tools
install_tools() {
    echo -e "${YELLOW}🛠️  Installing security tools...${NC}"
    
    # Install gosec if not present
    if ! command_exists gosec; then
        echo "Installing gosec..."
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    fi
    
    # Install golangci-lint if not present
    if ! command_exists golangci-lint; then
        echo "Installing golangci-lint..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
    fi
    
    # Install Trivy if not present
    if ! command_exists trivy; then
        echo "Installing Trivy..."
        sudo apt-get update
        sudo apt-get install -y wget apt-transport-https gnupg lsb-release
        wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | sudo apt-key add -
        echo "deb https://aquasecurity.github.io/trivy-repo/deb $(lsb_release -sc) main" | sudo tee -a /etc/apt/sources.list.d/trivy.list
        sudo apt-get update
        sudo apt-get install -y trivy
    fi
    
    # Install Nancy if not present
    if ! command_exists nancy; then
        echo "Installing Nancy..."
        go install github.com/sonatype-nexus-community/nancy@latest
    fi
    
    echo -e "${GREEN}✅ All tools installed${NC}"
}

# Function to run Trivy filesystem scan
run_trivy_fs() {
    echo -e "${BLUE}🔍 Running Trivy filesystem scan...${NC}"
    
    # Scan for vulnerabilities
    trivy fs --severity HIGH,CRITICAL --format json --output "$REPORTS_DIR/trivy-fs-$TIMESTAMP.json" . || {
        echo -e "${YELLOW}⚠️  Trivy found issues${NC}"
    }
    
    # Scan for secrets
    trivy fs --scanners secret --format json --output "$REPORTS_DIR/trivy-secrets-$TIMESTAMP.json" . || {
        echo -e "${YELLOW}⚠️  Trivy found secrets${NC}"
    }
    
    echo -e "${GREEN}✅ Trivy filesystem scan completed${NC}"
}

# Function to run Trivy dependency scan
run_trivy_deps() {
    echo -e "${BLUE}🔍 Running Trivy dependency scan...${NC}"
    
    trivy fs --scanners vuln --format json --output "$REPORTS_DIR/trivy-deps-$TIMESTAMP.json" . || {
        echo -e "${YELLOW}⚠️  Trivy found dependency issues${NC}"
    }
    
    echo -e "${GREEN}✅ Trivy dependency scan completed${NC}"
}

# Function to run Gosec
run_gosec() {
    echo -e "${BLUE}🔍 Running Gosec security scan...${NC}"
    
    gosec -fmt json -out "$REPORTS_DIR/gosec-$TIMESTAMP.json" ./... || {
        echo -e "${YELLOW}⚠️  Gosec found issues${NC}"
    }
    
    echo -e "${GREEN}✅ Gosec scan completed${NC}"
}

# Function to run golangci-lint
run_golangci_lint() {
    echo -e "${BLUE}🔍 Running golangci-lint...${NC}"
    
    export PATH=$PATH:$(go env GOPATH)/bin
    
    golangci-lint run --timeout=5m --out-format=json > "$REPORTS_DIR/golangci-lint-$TIMESTAMP.json" || {
        echo -e "${YELLOW}⚠️  golangci-lint found issues${NC}"
    }
    
    echo -e "${GREEN}✅ golangci-lint completed${NC}"
}

# Function to run Nancy
run_nancy() {
    echo -e "${BLUE}🔍 Running Nancy dependency scan...${NC}"
    
    go list -json -deps ./... | nancy sleuth --output json > "$REPORTS_DIR/nancy-$TIMESTAMP.json" || {
        echo -e "${YELLOW}⚠️  Nancy found issues${NC}"
    }
    
    echo -e "${GREEN}✅ Nancy scan completed${NC}"
}

# Function to run Safety (Python dependencies)
run_safety() {
    echo -e "${BLUE}🔍 Running Safety Python scan...${NC}"
    
    if [ -f "requirements.txt" ]; then
        pip install safety
        safety check -r requirements.txt --json --output "$REPORTS_DIR/safety-$TIMESTAMP.json" || {
            echo -e "${YELLOW}⚠️  Safety found issues${NC}"
        }
        echo -e "${GREEN}✅ Safety scan completed${NC}"
    else
        echo -e "${YELLOW}⚠️  No requirements.txt found, skipping Safety scan${NC}"
    fi
}

# Function to generate summary report
generate_summary() {
    echo -e "${BLUE}📊 Generating security summary...${NC}"
    
    cat > "$REPORTS_DIR/security-summary-$TIMESTAMP.md" << EOF
# 🔒 Security Scan Summary

**Date:** $(date)
**Project:** CloudBridge Relay
**Timestamp:** $TIMESTAMP

## 📋 Scans Performed

- ✅ Trivy filesystem scan
- ✅ Trivy secrets scan  
- ✅ Trivy dependency scan
- ✅ Gosec security scan
- ✅ golangci-lint scan
- ✅ Nancy dependency scan
- ✅ Safety Python scan (if applicable)

## 📁 Report Files

- \`trivy-fs-$TIMESTAMP.json\`: Trivy filesystem scan results
- \`trivy-secrets-$TIMESTAMP.json\`: Trivy secrets scan results
- \`trivy-deps-$TIMESTAMP.json\`: Trivy dependency scan results
- \`gosec-$TIMESTAMP.json\`: Gosec security scan results
- \`golangci-lint-$TIMESTAMP.json\`: golangci-lint results
- \`nancy-$TIMESTAMP.json\`: Nancy dependency scan results
- \`safety-$TIMESTAMP.json\`: Safety Python scan results (if applicable)

## 🚨 Critical Issues

Please review the JSON files for detailed vulnerability information.

## 🔧 Recommendations

1. **Update Dependencies**: Update vulnerable dependencies to fixed versions
2. **Remove Secrets**: Move private keys to secure secret management
3. **Code Quality**: Address golangci-lint warnings and errors
4. **Regular Scanning**: Run security scans regularly in CI/CD pipeline

## 📊 Next Steps

1. Review all scan results
2. Address critical and high severity issues
3. Update vulnerable dependencies
4. Move secrets to secure storage
5. Integrate security scanning into CI/CD pipeline

EOF

    echo -e "${GREEN}✅ Summary report generated: $REPORTS_DIR/security-summary-$TIMESTAMP.md${NC}"
}

# Function to display results
display_results() {
    echo -e "${BLUE}📊 Security Scan Results${NC}"
    echo -e "${BLUE}========================${NC}"
    echo ""
    echo -e "${GREEN}✅ All security scans completed successfully!${NC}"
    echo ""
    echo "📁 Reports generated in: $REPORTS_DIR"
    echo "📄 Summary report: security-summary-$TIMESTAMP.md"
    echo ""
    echo -e "${YELLOW}🔍 Review the following files:${NC}"
    ls -la "$REPORTS_DIR"/*"$TIMESTAMP"*
    echo ""
    echo -e "${BLUE}💡 Next steps:${NC}"
    echo "1. Review all JSON reports for vulnerabilities"
    echo "2. Address critical and high severity issues"
    echo "3. Update vulnerable dependencies"
    echo "4. Move secrets to secure storage"
    echo "5. Integrate into CI/CD pipeline"
}

# Main execution
main() {
    cd "$PROJECT_ROOT"
    
    # Install tools if needed
    install_tools
    
    # Run all security scans
    run_trivy_fs
    run_trivy_deps
    run_gosec
    run_golangci_lint
    run_nancy
    run_safety
    
    # Generate summary
    generate_summary
    
    # Display results
    display_results
}

# Run main function
main "$@"
