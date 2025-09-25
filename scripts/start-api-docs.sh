#!/bin/bash

# AI Service API Documentation Server
# Starts a local server to serve API documentation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PORT="8080"
DOCS_DIR="docs"
API_DOCS_FILE="ai-api-docs.html"

echo -e "${BLUE}🚀 AI Service API Documentation Server${NC}"
echo "================================================"

# Function to create API documentation HTML
create_api_docs() {
    echo -e "${YELLOW}📝 Creating API documentation...${NC}"
    
    # Create docs directory if it doesn't exist
    mkdir -p "$DOCS_DIR"
    
    # Create simple API documentation
    cat > "$DOCS_DIR/$API_DOCS_FILE" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CloudBridge AI Service API Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #2c3e50; border-bottom: 3px solid #3498db; padding-bottom: 10px; }
        h2 { color: #34495e; margin-top: 30px; border-left: 4px solid #3498db; padding-left: 15px; }
        .endpoint { background: #f8f9fa; border: 1px solid #e9ecef; border-radius: 5px; padding: 15px; margin: 10px 0; }
        .method { display: inline-block; padding: 4px 8px; border-radius: 3px; font-weight: bold; font-size: 12px; margin-right: 10px; }
        .get { background: #d4edda; color: #155724; }
        .post { background: #cce5ff; color: #004085; }
        .path { font-family: 'Courier New', monospace; font-weight: bold; color: #e74c3c; }
        .response { background: #f1f3f4; border-left: 4px solid #28a745; padding: 10px; margin: 10px 0; font-family: 'Courier New', monospace; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 CloudBridge AI Service API Documentation</h1>
        
        <h2>🔍 Basic Endpoints</h2>
        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/health</span>
            <div>Health check endpoint</div>
        </div>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/ai/status</span>
            <div>Get AI service status</div>
        </div>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/ai/metrics</span>
            <div>Get AI service metrics</div>
        </div>
        
        <h2>🔍 Anomaly Detection</h2>
        <div class="endpoint">
            <span class="method post">POST</span>
            <span class="path">/api/v1/ai/anomaly/detect</span>
            <div>Detect anomalies in network metrics</div>
        </div>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/ai/history/anomalies</span>
            <div>Get anomaly detection history</div>
        </div>
        
        <h2>🔍 Predictive Analytics</h2>
        <div class="endpoint">
            <span class="method post">POST</span>
            <span class="path">/api/v1/ai/predict/load</span>
            <div>Predict system load</div>
        </div>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/ai/history/predictions</span>
            <div>Get prediction history</div>
        </div>
        
        <h2>🔍 Real-time Data</h2>
        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/ai/realtime/connections</span>
            <div>Get real-time connection data</div>
        </div>
        
        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/ai/realtime/performance</span>
            <div>Get real-time performance metrics</div>
        </div>
        
        <h2>🔍 Alerts & Notifications</h2>
        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/ai/alerts</span>
            <div>Get active alerts</div>
        </div>
        
        <div style="text-align: center; margin-top: 40px; padding: 20px; background: #f8f9fa; border-radius: 5px;">
            <p><strong>🚀 CloudBridge AI Service API Documentation</strong></p>
            <p>Generated on: <script>document.write(new Date().toISOString())</script></p>
        </div>
    </div>
</body>
</html>
EOF

    echo -e "${GREEN}API documentation created${NC}"
}

# Function to start documentation server
start_docs_server() {
    echo -e "${YELLOW}🚀 Starting API documentation server...${NC}"
    
    # Check if Python is available
    if command -v python3 >/dev/null 2>&1; then
        PYTHON_CMD="python3"
    elif command -v python >/dev/null 2>&1; then
        PYTHON_CMD="python"
    else
        echo -e "${RED}❌ Python not found. Please install Python to run the documentation server.${NC}"
        exit 1
    fi
    
    # Create API documentation
    create_api_docs
    
    echo -e "${GREEN}Documentation server starting on port $PORT${NC}"
    echo -e "${BLUE}📖 Access the documentation at: http://localhost:$PORT/$API_DOCS_FILE${NC}"
    echo
    echo -e "${YELLOW}Press Ctrl+C to stop the server${NC}"
    echo
    
    # Start the server
    cd "$DOCS_DIR"
    $PYTHON_CMD -m http.server $PORT
}

# Main function
main() {
    echo -e "${YELLOW}AI Service API Documentation Server${NC}"
    echo
    
    # Check if port is available
    if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "${RED}❌ Port $PORT is already in use${NC}"
        echo -e "${YELLOW}   Please stop the service using port $PORT or use a different port${NC}"
        exit 1
    fi
    
    start_docs_server
}

# Run main function
main "$@"